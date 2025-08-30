package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Note: Request/Response types are defined in api.go

// HandleSubmitData handles POST / - submits data for signing
// @Summary Submit data for signing
// @Description Submit raw data for remote signing. Returns a UUID to track the signing request.
// @Tags Signing
// @Accept json,application/octet-stream
// @Produce json
// @Param request body SubmitDataRequest false "Data submission request (JSON format)"
// @Success 200 {object} SubmitDataResponse "Data submitted successfully"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 413 {object} ErrorResponse "Data too large"
// @Router / [post]
func (s *Server) HandleSubmitData(c *gin.Context) {
	var req SubmitDataRequest

	// Check content type and handle accordingly
	contentType := c.GetHeader("Content-Type")

	if contentType == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON payload",
				"details": err.Error(),
			})
			return
		}
	} else {
		// Handle raw binary data
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			return
		}

		if len(body) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Empty request body",
			})
			return
		}

		req.Data = body
		req.ClientID = c.Query("client_id")
		req.CallbackURL = c.Query("callback_url")
	}

	// Validate data size
	if int64(len(req.Data)) > s.config.MaxDataSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "Data too large",
			"max_size": s.config.MaxDataSize,
		})
		return
	}

	// Generate UUID for this signing request
	itemUUID := uuid.New().String()

	// Create signing request
	signingRequest := &SigningRequest{
		UUID:        itemUUID,
		Data:        req.Data,
		CreatedAt:   time.Now(),
		IsSigned:    false,
		RequestedAt: time.Now(),
		ClientID:    req.ClientID,
		CallbackURL: req.CallbackURL,
	}

	// Store the signing request
	s.mutex.Lock()
	s.signingRequests[itemUUID] = signingRequest
	s.mutex.Unlock()

	// Notify WebSocket clients about new signing request
	s.hub.BroadcastToUUID(itemUUID, WebSocketMessage{
		Type: MessageTypeStatus,
		UUID: itemUUID,
		Payload: map[string]interface{}{
			"status": "pending",
			"message": "Data submitted for signing",
			"created_at": signingRequest.CreatedAt,
		},
	})

	// Generate signing URL
	signingURL := s.generateSigningURL(itemUUID)

	// Set expiration timer
	go s.setExpirationTimer(itemUUID)

	c.JSON(http.StatusCreated, SubmitDataResponse{
		UUID:       itemUUID,
		SigningURL: signingURL,
		Message:    "Data submitted successfully. Use the signing URL to sign the data.",
	})
}

// HandleGetData handles GET /<uuid> - retrieves unsigned data
// @Summary Get unsigned data
// @Description Retrieve the unsigned data for a signing request by UUID
// @Tags Signing
// @Produce application/octet-stream,json
// @Param uuid path string true "Signing request UUID"
// @Success 200 {string} binary "Raw binary data"
// @Failure 404 {object} ErrorResponse "Signing request not found"
// @Router /{uuid} [get]
func (s *Server) HandleGetData(c *gin.Context) {
	itemUUID := c.Param("uuid")

	// Validate UUID format
	if _, err := uuid.Parse(itemUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	s.mutex.RLock()
	signingRequest, exists := s.signingRequests[itemUUID]
	s.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Signing request not found",
			"uuid": itemUUID,
		})
		return
	}

	if signingRequest.IsSigned {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Data already signed",
			"uuid": itemUUID,
		})
		return
	}

	// Return the raw data for signing
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", signingRequest.Data)
}

// HandleSubmitSignedData handles POST /<uuid> - submits signed data
// @Summary Submit signed data
// @Description Submit the signed data for a signing request
// @Tags Signing
// @Accept json,application/octet-stream
// @Produce json
// @Param uuid path string true "Signing request UUID"
// @Param request body SubmitSignedDataRequest false "Signed data submission"
// @Success 200 {object} SuccessResponse "Signed data submitted successfully"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Signing request not found"
// @Router /{uuid} [post]
func (s *Server) HandleSubmitSignedData(c *gin.Context) {
	itemUUID := c.Param("uuid")

	// Validate UUID format
	if _, err := uuid.Parse(itemUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	// Handle both JSON and raw binary data
	var signedData []byte
	contentType := c.GetHeader("Content-Type")

	if contentType == "application/json" {
		var req SubmitSignedDataRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON payload",
				"details": err.Error(),
			})
			return
		}
		signedData = req.SignedData
	} else {
		// Handle raw binary data
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			return
		}
		signedData = body
	}

	if len(signedData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Empty signed data",
		})
		return
	}

	s.mutex.Lock()
	signingRequest, exists := s.signingRequests[itemUUID]
	if !exists {
		s.mutex.Unlock()
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Signing request not found",
			"uuid": itemUUID,
		})
		return
	}

	if signingRequest.IsSigned {
		s.mutex.Unlock()
		c.JSON(http.StatusConflict, gin.H{
			"error": "Data already signed",
			"uuid": itemUUID,
		})
		return
	}

	// Update the signing request with signed data
	signingRequest.IsSigned = true
	signingRequest.SignedData = signedData

	s.mutex.Unlock()

	// Create signed response
	signedResponse := &SignedResponse{
		UUID:       itemUUID,
		SignedData: signedData,
		SignedAt:   time.Now(),
		Success:    true,
	}

	// Notify SSE clients about successful signing (metadata only)
	s.hub.BroadcastSSEToUUID(itemUUID, SSEEvent{
		Type: "signed",
		Data: map[string]interface{}{
			"uuid":      itemUUID,
			"signed_at": signedResponse.SignedAt,
			"success":   true,
		},
	})

	// If there's a callback URL, notify the original client
	if signingRequest.CallbackURL != "" {
		go s.notifyCallback(signingRequest.CallbackURL, signedResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data signed successfully",
		"uuid": itemUUID,
		"signed_at": signedResponse.SignedAt,
	})
}

// HandleGetSignedData serves the signed binary data for a specific signing request
// @Summary Get signed data
// @Description Get the signed binary data for a specific signing request
// @Tags Data
// @Produce application/octet-stream
// @Param uuid path string true "Signing request UUID"
// @Success 200 {file} binary "Signed binary data"
// @Failure 404 {object} ErrorResponse "Signing request not found or not signed"
// @Router /signed/{uuid} [get]
func (s *Server) HandleGetSignedData(c *gin.Context) {
	itemUUID := c.Param("uuid")

	// Validate UUID format
	if _, err := uuid.Parse(itemUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	s.mutex.RLock()
	signingRequest, exists := s.signingRequests[itemUUID]
	s.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Signing request not found",
		})
		return
	}

	if !signingRequest.IsSigned {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Signing request not yet signed",
		})
		return
	}

	// Serve the signed binary data directly
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=signed-data-%s.bin", itemUUID))
	c.Data(http.StatusOK, "application/octet-stream", signingRequest.SignedData)
}

// HandleGetSigningForm serves the HTML form for signing data
// @Summary Get signing form
// @Description Get the web-based signing interface for a specific signing request
// @Tags Web Interface
// @Produce text/html
// @Param uuid path string true "Signing request UUID"
// @Success 200 {string} string "HTML signing form"
// @Failure 404 {object} ErrorResponse "Signing request not found"
// @Router /sign/{uuid} [get]
func (s *Server) HandleGetSigningForm(c *gin.Context) {
	itemUUID := c.Param("uuid")

	// Check if we're on the test route
	if c.Request.URL.Path == "/test" {
		// For test route, just serve the React app without UUID validation
		serveReactApp(c, "Harlequin Remote Signing - Test Mode")
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(itemUUID); err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	s.mutex.RLock()
	signingRequest, exists := s.signingRequests[itemUUID]
	s.mutex.RUnlock()

	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Signing request not found or expired",
			"uuid": itemUUID,
		})
		return
	}

		if signingRequest.IsSigned {
		// Serve React app with signed status indication
		serveReactApp(c, "Harlequin Remote Signing - Already Signed")
		return
	}

	// Serve the React app directly
	serveReactApp(c, "Harlequin Remote Signing")
}


// HandleGetStatus handles GET /status - returns server status and statistics
// @Summary Get server status
// @Description Get the current status of all signing requests
// @Tags Status
// @Produce json
// @Success 200 {object} StatusResponse "Current server status"
// @Router /status [get]
func (s *Server) HandleGetStatus(c *gin.Context) {
	s.mutex.RLock()
	totalRequests := len(s.signingRequests)
	signedRequests := 0
	pendingRequests := 0

	for _, req := range s.signingRequests {
		if req.IsSigned {
			signedRequests++
		} else {
			pendingRequests++
		}
	}
	s.mutex.RUnlock()

	connectedClients := s.hub.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"status": "running",
			"uptime": time.Since(s.startTime).String(),
			"version": "1.0.0",
		},
		"requests": gin.H{
			"total": totalRequests,
			"signed": signedRequests,
			"pending": pendingRequests,
		},
		"websockets": gin.H{
			"connected_clients": connectedClients,
		},
		"config": gin.H{
			"host": s.config.Host,
			"port": s.config.Port,
			"max_data_size": s.config.MaxDataSize,
			"signing_timeout": s.config.SigningTimeout.String(),
		},
	})
}

// Helper methods

func (s *Server) generateSigningURL(uuid string) string {
	// If a custom frontend URL is configured, use it as the host
	// Otherwise, use the server URL as the host
	hostURL := s.getServerURL()
	if s.config.FrontendURL != "" {
		hostURL = s.config.FrontendURL
	}

	// Always include the server parameter so the frontend knows where to make API calls
	return fmt.Sprintf("%s/sign/%s?server=%s", hostURL, uuid, s.getServerURL())
}

func (s *Server) getServerURL() string {
	return fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Port)
}

func (s *Server) getWebSocketURL() string {
	return fmt.Sprintf("ws://%s:%d/ws", s.config.Host, s.config.Port)
}

func (s *Server) setExpirationTimer(uuid string) {
	time.Sleep(s.config.SigningTimeout)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if signingRequest, exists := s.signingRequests[uuid]; exists && !signingRequest.IsSigned {
		delete(s.signingRequests, uuid)

		// Notify WebSocket clients about expiration
		s.hub.BroadcastToUUID(uuid, WebSocketMessage{
			Type: MessageTypeError,
			UUID: uuid,
			Error: "Signing request expired",
		})
	}
}

// HandleSSE handles GET /events/:uuid - Server-Sent Events for real-time updates
// @Summary Server-Sent Events
// @Description Real-time event stream for signing request updates
// @Tags Events
// @Produce text/event-stream
// @Param uuid path string true "Signing request UUID"
// @Success 200 {string} string "Event stream"
// @Router /events/{uuid} [get]
func (s *Server) HandleSSE(c *gin.Context) {
	itemUUID := c.Param("uuid")

	// Validate UUID format
	if _, err := uuid.Parse(itemUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Create a channel for this client
	clientChan := make(chan SSEEvent, 10)
	clientID := uuid.New().String()

	// Register this client
	s.hub.RegisterSSEClient(itemUUID, clientID, clientChan)
	defer s.hub.UnregisterSSEClient(itemUUID, clientID)

	// Check if signing is already complete
	s.mutex.RLock()
	signingRequest, exists := s.signingRequests[itemUUID]
	isSigned := exists && signingRequest.IsSigned
	s.mutex.RUnlock()

	if isSigned {
		// Send signed event immediately
		c.SSEvent("signed", gin.H{
			"uuid": itemUUID,
			"signed_at": time.Now(),
			"success": true,
		})
		c.Writer.Flush()
		return
	}

	// Send initial connection event
	c.SSEvent("connected", gin.H{
		"uuid": itemUUID,
		"client_id": clientID,
	})
	c.Writer.Flush()

	// Small delay to ensure client is ready
	time.Sleep(200 * time.Millisecond)

	// Keep connection alive with periodic heartbeats
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen for events
	for {
		select {
		case event := <-clientChan:
			// Send the event to the client
			c.SSEvent(event.Type, event.Data)
			c.Writer.Flush()

			// If it's a signed event, close the connection
			if event.Type == "signed" {
				return
			}
		case <-ticker.C:
			// Send heartbeat
			c.SSEvent("heartbeat", gin.H{
				"timestamp": time.Now().Unix(),
			})
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			// Client disconnected
			return
		}
	}
}

// HandleHealth handles GET /health - returns server health status
// @Summary Health check
// @Description Get the health status of the server
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse "Server is healthy"
// @Router /health [get]
func (s *Server) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (s *Server) notifyCallback(callbackURL string, response *SignedResponse) {
	// TODO: Implement HTTP callback notification
	// This would make an HTTP POST request to the callback URL with the signed response
}

// serveReactApp is a helper function to serve the React app with the correct assets
func serveReactApp(c *gin.Context, title string) {
	c.Header("Content-Type", "text/html")

	// Get the frontend path
	_, filename, _, _ := runtime.Caller(0)
	serverDir := filepath.Dir(filename)
	projectRoot := filepath.Join(serverDir, "..")
	frontendPath := filepath.Join(projectRoot, "frontend/dist")
	assetsPath := filepath.Join(frontendPath, "assets")

	// Find the current JS and CSS files
	var jsFile, cssFile string
	if entries, err := os.ReadDir(assetsPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				if strings.HasSuffix(name, ".js") && strings.HasPrefix(name, "index-") {
					jsFile = name
				} else if strings.HasSuffix(name, ".css") && strings.HasPrefix(name, "index-") {
					cssFile = name
				}
			}
		}
	}

	// Fallback to default names if not found
	if jsFile == "" {
		jsFile = "index-DLkHF8kv.js"
	}
	if cssFile == "" {
		cssFile = "index-Cwd2Qldy.css"
	}

	indexHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
    <script type="module" crossorigin src="/static/%s"></script>
    <link rel="stylesheet" crossorigin href="/static/%s">
  </head>
  <body>
    <div id="root"></div>
  </body>
</html>`, title, jsFile, cssFile)
	c.Data(http.StatusOK, "text/html", []byte(indexHTML))
}

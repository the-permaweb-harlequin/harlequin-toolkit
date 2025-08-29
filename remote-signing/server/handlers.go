package server

import (
	"fmt"
	"io"
	"net/http"
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
	c.JSON(http.StatusOK, gin.H{
		"uuid": itemUUID,
		"data": signingRequest.Data,
		"created_at": signingRequest.CreatedAt,
		"requested_at": signingRequest.RequestedAt,
		"client_id": signingRequest.ClientID,
	})
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

	// Notify WebSocket clients about successful signing
	s.hub.BroadcastToUUID(itemUUID, WebSocketMessage{
		Type: MessageTypeSigned,
		UUID: itemUUID,
		Payload: signedResponse,
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
		c.HTML(http.StatusOK, "already_signed.html", gin.H{
			"uuid": itemUUID,
			"signed_at": signingRequest.CreatedAt,
		})
		return
	}

	c.HTML(http.StatusOK, "signing_form.html", gin.H{
		"uuid": itemUUID,
		"data": signingRequest.Data,
		"data_size": len(signingRequest.Data),
		"created_at": signingRequest.CreatedAt,
		"server_url": s.getServerURL(),
		"websocket_url": s.getWebSocketURL(),
	})
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
	return s.getServerURL() + "/sign/" + uuid
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

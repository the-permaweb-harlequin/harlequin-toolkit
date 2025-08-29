package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SubmitDataRequest represents the request body for submitting data
type SubmitDataRequest struct {
	Data        []byte `json:"data,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// SubmitDataResponse represents the response for submitting data
type SubmitDataResponse struct {
	UUID       string `json:"uuid"`
	SigningURL string `json:"signing_url"`
	Message    string `json:"message"`
}

// SubmitSignedDataRequest represents the request for submitting signed data
type SubmitSignedDataRequest struct {
	SignedData []byte `json:"signed_data" binding:"required"`
}

// HandleSubmitData handles POST / - submits data for signing
func (s *RemoteSigningServer) HandleSubmitData(c *gin.Context) {
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
func (s *RemoteSigningServer) HandleGetData(c *gin.Context) {
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
func (s *RemoteSigningServer) HandleSubmitSignedData(c *gin.Context) {
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
func (s *RemoteSigningServer) HandleGetSigningForm(c *gin.Context) {
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

// Helper methods

func (s *RemoteSigningServer) generateSigningURL(uuid string) string {
	return s.getServerURL() + "/sign/" + uuid
}

func (s *RemoteSigningServer) getServerURL() string {
	return "http://" + s.config.Host + ":" + fmt.Sprintf("%d", s.config.Port)
}

func (s *RemoteSigningServer) getWebSocketURL() string {
	return "ws://" + s.config.Host + ":" + fmt.Sprintf("%d", s.config.Port) + "/ws"
}

func (s *RemoteSigningServer) setExpirationTimer(uuid string) {
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

func (s *RemoteSigningServer) notifyCallback(callbackURL string, response *SignedResponse) {
	// TODO: Implement HTTP callback notification
	// This would make an HTTP POST request to the callback URL with the signed response
}

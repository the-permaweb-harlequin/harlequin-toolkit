package server

// @title Harlequin Remote Signing Server API
// @version 1.0.0
// @description A secure HTTP/WebSocket service for remote transaction signing workflows. Submit raw data for signing via web interface with wallet extensions.
// @contact.name Harlequin Team
// @contact.url https://github.com/the-permaweb-harlequin/harlequin-toolkit
// @license.name MIT
// @license.url https://github.com/the-permaweb-harlequin/harlequin-toolkit/blob/main/LICENSE
// @host localhost:8080
// @BasePath /
// @schemes http https

// SubmitDataRequest represents the request body for submitting data
// @Description Request structure for submitting raw data for signing
type SubmitDataRequest struct {
	Data        []byte `json:"data,omitempty" example:"SGVsbG8gV29ybGQ=" format:"byte"`                                   // Raw data to be signed (base64 encoded in JSON)
	ClientID    string `json:"client_id,omitempty" example:"client-app-v1.2.3"`                                         // Client identifier for tracking
	CallbackURL string `json:"callback_url,omitempty" example:"https://your-app.com/webhook/signing-complete"`          // Optional webhook URL for completion notification
}

// SubmitDataResponse represents the response after submitting data
// @Description Response structure after successfully submitting data for signing
type SubmitDataResponse struct {
	UUID       string `json:"uuid" example:"123e4567-e89b-12d3-a456-426614174000"`                          // Unique identifier for the signing request
	SigningURL string `json:"signing_url" example:"http://localhost:8080/sign/123e4567-e89b-12d3-a456-426614174000"` // URL for the web signing interface
	Message    string `json:"message" example:"Data submitted for signing"`                                 // Status message
}

// SubmitSignedDataRequest represents the request body for submitting signed data
// @Description Request structure for submitting signed data
type SubmitSignedDataRequest struct {
	SignedData []byte `json:"signed_data" binding:"required" example:"U2lnbmVkIEhlbGxvIFdvcmxk" format:"byte"` // Signed data (base64 encoded in JSON)
}

// HealthResponse represents the health check response
// @Description Server health status response
type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`     // Health status
	Timestamp int64  `json:"timestamp" example:"1640995200"` // Unix timestamp
	Version   string `json:"version" example:"1.0.0"`      // Server version
}

// StatusResponse represents the server status response
// @Description Current server status and statistics
type StatusResponse struct {
	ActiveRequests int    `json:"active_requests" example:"3"`        // Number of active signing requests
	TotalRequests  int    `json:"total_requests" example:"150"`       // Total requests processed
	ServerUptime   string `json:"server_uptime" example:"2h30m45s"`   // Server uptime duration
}

// ErrorResponse represents an error response
// @Description Standard error response structure
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request format"` // Error message
}

// SuccessResponse represents a generic success response
// @Description Generic success response structure
type SuccessResponse struct {
	Message string `json:"message" example:"Signed data received successfully"` // Success message
}

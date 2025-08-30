package server

import (
	"time"

	"github.com/gorilla/websocket"
)

// SigningRequest represents a request to sign raw data
type SigningRequest struct {
	UUID        string    `json:"uuid"`
	Data        []byte    `json:"data"`
	CreatedAt   time.Time `json:"created_at"`
	IsSigned    bool      `json:"is_signed"`
	SignedData  []byte    `json:"signed_data,omitempty"`
	RequestedAt time.Time `json:"requested_at"`
	ClientID    string    `json:"client_id"`
	CallbackURL string    `json:"callback_url,omitempty"`
}

// SignedResponse represents a completed signing response
type SignedResponse struct {
	UUID       string    `json:"uuid"`
	SignedData []byte    `json:"signed_data"`
	SignedAt   time.Time `json:"signed_at"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

// WebSocketMessage represents messages sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	UUID    string      `json:"uuid,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan WebSocketMessage
	ClientID string
	UUIDs    map[string]bool // Track which UUIDs this client is interested in
}

// MessageType constants for WebSocket communication
const (
	MessageTypeStatus    = "status"
	MessageTypeSigned    = "signed"
	MessageTypeError     = "error"
	MessageTypeSubscribe = "subscribe"
	MessageTypeHeartbeat = "heartbeat"
)

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Type string      `json:"type"`
	UUID string      `json:"uuid,omitempty"`
	Data interface{} `json:"data"`
}

// Config represents the server configuration
type Config struct {
	Port           int           `json:"port"`
	Host           string        `json:"host"`
	AllowedOrigins []string      `json:"allowed_origins"`
	MaxDataSize    int64         `json:"max_data_size"`
	SigningTimeout time.Duration `json:"signing_timeout"`
}

// DefaultConfig returns the default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:           8080,
		Host:          "localhost",
		AllowedOrigins: []string{"*"},
		MaxDataSize:   10 * 1024 * 1024, // 10MB
		SigningTimeout: 30 * time.Minute,
	}
}

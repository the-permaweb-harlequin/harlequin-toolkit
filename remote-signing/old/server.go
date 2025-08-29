package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RemoteSigningServer represents the remote signing server
type RemoteSigningServer struct {
	config          *ServerConfig
	server          *http.Server
	hub             *WebSocketHub
	signingRequests map[string]*SigningRequest
	mutex           sync.RWMutex
	isRunning       bool
}

// NewRemoteSigningServer creates a new remote signing server
func NewRemoteSigningServer(config *ServerConfig) *RemoteSigningServer {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &RemoteSigningServer{
		config:          config,
		hub:             NewWebSocketHub(),
		signingRequests: make(map[string]*SigningRequest),
		isRunning:       false,
	}
}

// Start starts the remote signing server
func (s *RemoteSigningServer) Start(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	// Start WebSocket hub
	go s.hub.Run()

	// Configure Gin
	if gin.Mode() != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(s.corsMiddleware())

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// API routes
	router.POST("/", s.HandleSubmitData)
	router.GET("/:uuid", s.HandleGetData)
	router.POST("/:uuid", s.HandleSubmitSignedData)

	// WebSocket endpoint
	router.GET("/ws", s.HandleWebSocket)

	// Frontend signing routes
	router.GET("/sign/:uuid", s.HandleGetSigningForm)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
			"version": "1.0.0",
		})
	})

	// Status endpoint
	router.GET("/status", s.HandleGetStatus)

	// Create HTTP server
	addr := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	s.isRunning = true

	log.Printf("🎭 Remote Signing Server starting on %s", addr)
	log.Printf("📝 Signing interface available at: http://%s/sign/<uuid>", addr)
	log.Printf("🔌 WebSocket endpoint: ws://%s/ws", addr)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	// Wait for context cancellation or server shutdown
	<-ctx.Done()
	return s.Stop()
}

// Stop stops the remote signing server
func (s *RemoteSigningServer) Stop() error {
	if !s.isRunning {
		return fmt.Errorf("server is not running")
	}

	log.Println("🛑 Shutting down Remote Signing Server...")

	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully shutdown the HTTP server
	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	s.isRunning = false
	log.Println("✅ Remote Signing Server stopped")
	return nil
}

// IsRunning returns whether the server is currently running
func (s *RemoteSigningServer) IsRunning() bool {
	return s.isRunning
}

// GetConfig returns the server configuration
func (s *RemoteSigningServer) GetConfig() *ServerConfig {
	return s.config
}

// HandleGetStatus handles GET /status - returns server status and statistics
func (s *RemoteSigningServer) HandleGetStatus(c *gin.Context) {
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

	s.hub.mutex.RLock()
	connectedClients := len(s.hub.clients)
	s.hub.mutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"status": "running",
			"uptime": time.Since(time.Now()).String(), // This would need to track actual start time
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

// corsMiddleware adds CORS headers
func (s *RemoteSigningServer) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range s.config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

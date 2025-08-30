package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/docs"
)

// Server represents the remote signing server
type Server struct {
	config          *Config
	server          *http.Server
	hub             *WebSocketHub
	signingRequests map[string]*SigningRequest
	mutex           sync.RWMutex
	isRunning       bool
	startTime       time.Time
}

// New creates a new remote signing server
func New(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	return &Server{
		config:          config,
		hub:             NewWebSocketHub(),
		signingRequests: make(map[string]*SigningRequest),
		isRunning:       false,
	}
}

// Start starts the remote signing server
func (s *Server) Start(ctx context.Context) error {
	return s.StartWithTemplates(ctx, "")
}

// StartWithTemplates starts the server with custom template path
func (s *Server) StartWithTemplates(ctx context.Context, templatePath string) error {
	if s.isRunning {
		return fmt.Errorf("server is already running")
	}

	s.startTime = time.Now()

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

	// Load HTML templates if provided (for backward compatibility)
	if templatePath != "" {
		router.LoadHTMLGlob(templatePath + "/*")
	}

	// API routes
	router.POST("/", s.HandleSubmitData)
	router.GET("/:uuid", s.HandleGetData)
	router.POST("/:uuid", s.HandleSubmitSignedData)

	// WebSocket endpoint
	router.GET("/ws", s.HandleWebSocket)

	// SSE endpoint for real-time updates
	router.GET("/events/:uuid", s.HandleSSE)

	// Serve static frontend files
	router.Static("/static", "./frontend/dist/assets")
	router.StaticFile("/favicon.ico", "./frontend/dist/vite.svg")

	// Frontend signing routes - serve React app directly
	router.GET("/sign/:uuid", s.HandleGetSigningForm)

	// Signed data endpoint
	router.GET("/signed/:uuid", s.HandleGetSignedData)

	// Health check
	router.GET("/health", s.HandleHealth)

	// Status endpoint
	router.GET("/status", s.HandleGetStatus)

	// API documentation
	router.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	addr := s.config.Host + ":" + strconv.Itoa(s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	s.isRunning = true

	log.Printf("🎭 Remote Signing Server starting on %s", addr)
	if templatePath != "" {
		log.Printf("📝 Signing interface available at: http://%s/sign/<uuid>", addr)
	}
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
func (s *Server) Stop() error {
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
func (s *Server) IsRunning() bool {
	return s.isRunning
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *Config {
	return s.config
}

// GetSigningRequest returns a signing request by UUID
func (s *Server) GetSigningRequest(uuid string) (*SigningRequest, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	req, exists := s.signingRequests[uuid]
	return req, exists
}

// ListSigningRequests returns all signing requests
func (s *Server) ListSigningRequests() map[string]*SigningRequest {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to avoid concurrent access issues
	result := make(map[string]*SigningRequest)
	for k, v := range s.signingRequests {
		result[k] = v
	}
	return result
}

// GetWebSocketHub returns the WebSocket hub for external use
func (s *Server) GetWebSocketHub() *WebSocketHub {
	return s.hub
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
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

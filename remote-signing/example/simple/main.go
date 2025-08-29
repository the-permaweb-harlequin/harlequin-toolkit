package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

func main() {
	// Create a custom configuration
	config := &server.Config{
		Port:           9090,
		Host:          "localhost",
		AllowedOrigins: []string{"*"},
		MaxDataSize:   5 * 1024 * 1024, // 5MB
		SigningTimeout: 15 * time.Minute,
	}

	// Create a new server instance
	srv := server.New(config)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🎭 Starting Simple Remote Signing Server Example...")
	fmt.Printf("📡 Server available at: http://%s:%d\n", config.Host, config.Port)
	fmt.Printf("🔌 WebSocket endpoint: ws://%s:%d/ws\n", config.Host, config.Port)
	fmt.Println("⏰ Server will run for 30 seconds...")
	fmt.Println()
	fmt.Println("Test with:")
	fmt.Printf("  curl -X POST http://%s:%d/ -d 'Hello World'\n", config.Host, config.Port)
	fmt.Println()

	// Start the server (this blocks until context is cancelled)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	fmt.Println("Example completed!")
}

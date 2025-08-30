//go:build example
// +build example

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

// ServerOnly demonstrates how to use the server package directly
// without the high-level SigningServer API.
func main() {
	// Create server configuration
	config := server.DefaultConfig()
	config.Port = 8080
	config.Host = "localhost"

	// Create server instance
	srv := server.New(config)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🎭 Starting Remote Signing Server...")
	fmt.Printf("📡 Server available at: http://%s:%d\n", config.Host, config.Port)
	fmt.Printf("🔌 WebSocket endpoint: ws://%s:%d/ws\n", config.Host, config.Port)
	fmt.Printf("📚 API docs at: http://%s:%d/api-docs\n", config.Host, config.Port)
	fmt.Println("⏰ Server will run for 30 seconds...")
	fmt.Println()
	fmt.Println("Test with:")
	fmt.Printf("  curl -X POST http://%s:%d/ -d 'Hello World'\n", config.Host, config.Port)
	fmt.Println()

	// Start the server (this blocks until context is cancelled)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	fmt.Println("✅ Server example completed!")
}

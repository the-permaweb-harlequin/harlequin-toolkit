//go:build test
// +build test

package main

import (
	"context"
	"log"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

func main() {
	// Create server config
	config := server.DefaultConfig()
	config.Port = 8080
	config.Host = "localhost"

	// Create and start server
	s := server.New(config)

	// Start server in background
	go func() {
		if err := s.Start(context.Background()); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	log.Println("🎭 Test server started on http://localhost:8080")
	log.Println("📝 Test page available at: http://localhost:8080/test")
	log.Println("📝 Signing page available at: http://localhost:8080/sign/<uuid>")
	log.Println("🔌 WebSocket endpoint: ws://localhost:8080/ws")
	log.Println("📚 API docs available at: http://localhost:8080/api-docs")
	log.Println("")
	log.Println("Press Ctrl+C to stop the server...")

	// Keep server running
	select {}
}

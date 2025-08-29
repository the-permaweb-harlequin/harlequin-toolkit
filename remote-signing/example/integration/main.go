package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

// MyApplication demonstrates how to integrate the signing server into your own application
type MyApplication struct {
	signingServer *server.Server
}

// NewMyApplication creates a new application with integrated signing
func NewMyApplication() *MyApplication {
	// Configure the signing server
	config := &server.Config{
		Port:           8888,
		Host:          "0.0.0.0",
		AllowedOrigins: []string{"http://localhost:3000", "https://myapp.com"},
		MaxDataSize:   50 * 1024 * 1024, // 50MB
		SigningTimeout: 60 * time.Minute,
	}

	return &MyApplication{
		signingServer: server.New(config),
	}
}

// Start starts the application and signing server
func (app *MyApplication) Start(ctx context.Context) error {
	fmt.Println("🚀 Starting MyApplication with integrated signing...")

	// Start your application logic here
	go app.runApplicationLogic(ctx)

	// Start the signing server (this blocks)
	return app.signingServer.Start(ctx)
}

// runApplicationLogic simulates your application's main logic
func (app *MyApplication) runApplicationLogic(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			app.checkSigningRequests()
		}
	}
}

// checkSigningRequests demonstrates how to monitor signing requests
func (app *MyApplication) checkSigningRequests() {
	requests := app.signingServer.ListSigningRequests()

	pendingCount := 0
	signedCount := 0

	for _, req := range requests {
		if req.IsSigned {
			signedCount++
		} else {
			pendingCount++
		}
	}

	if len(requests) > 0 {
		fmt.Printf("📊 Signing Status: %d pending, %d signed\n", pendingCount, signedCount)
	}
}

// processDataForSigning demonstrates how to programmatically submit data for signing
func (app *MyApplication) processDataForSigning(data []byte, clientID string) {
	// You could add data directly to the server's internal storage
	// or use HTTP requests to the server's API endpoints
	fmt.Printf("🔄 Processing %d bytes for signing from client %s\n", len(data), clientID)
}

func main() {
	app := NewMyApplication()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		time.Sleep(60 * time.Second) // Run for 1 minute
		fmt.Println("\n⏰ Demo time limit reached, shutting down...")
		cancel()
	}()

	fmt.Println("🎭 Advanced Integration Example")
	fmt.Println("================================")
	fmt.Printf("📡 Signing server: http://localhost:8888\n")
	fmt.Printf("🔌 WebSocket: ws://localhost:8888/ws\n")
	fmt.Printf("⏰ Will run for 60 seconds...\n")
	fmt.Println()
	fmt.Println("This example shows how to:")
	fmt.Println("  • Integrate the signing server into your app")
	fmt.Println("  • Monitor signing requests programmatically")
	fmt.Println("  • Configure custom settings")
	fmt.Println("  • Access the server's internal state")
	fmt.Println()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("Application failed: %v", err)
	}

	fmt.Println("✅ Integration example completed!")
}

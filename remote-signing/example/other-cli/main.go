package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

// This example shows how another CLI tool can integrate the signing server
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: my-other-cli [start-signing-server|submit-data|check-status]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start-signing-server":
		startSigningServer()
	case "submit-data":
		submitDataExample()
	case "check-status":
		checkStatusExample()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// startSigningServer demonstrates embedding the signing server in another CLI
func startSigningServer() {
	fmt.Println("🔧 MyOtherCLI - Starting integrated signing server...")

	// Create custom configuration for this CLI's needs
	config := &server.Config{
		Port:           7777,
		Host:          "localhost",
		AllowedOrigins: []string{"http://localhost:3000", "https://myapp.com"},
		MaxDataSize:   20 * 1024 * 1024, // 20MB for this app
		SigningTimeout: 45 * server.DefaultConfig().SigningTimeout / 30, // 45 minutes
	}

	srv := server.New(config)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n🛑 Shutting down MyOtherCLI signing server...")
		cancel()
	}()

	fmt.Printf("🎭 MyOtherCLI Signing Server running on http://%s:%d\n", config.Host, config.Port)
	fmt.Printf("🔌 WebSocket: ws://%s:%d/ws\n", config.Host, config.Port)
	fmt.Println("📋 This server is managed by MyOtherCLI")
	fmt.Println()

	// Start the server (blocks until context cancelled)
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Signing server failed: %v", err)
	}
}

// submitDataExample shows how to work with signing requests programmatically
func submitDataExample() {
	fmt.Println("📤 MyOtherCLI - Submit Data Example")
	fmt.Println("This would typically make HTTP requests to the signing server")
	fmt.Println("or integrate directly with the server package for custom workflows")

	// Example: You could create a server instance just to access its types
	// and make HTTP requests to another instance, or embed it directly

	fmt.Println()
	fmt.Println("Example workflow:")
	fmt.Println("1. MyOtherCLI prepares data for signing")
	fmt.Println("2. Submits to embedded signing server")
	fmt.Println("3. Provides custom UI/UX for signing flow")
	fmt.Println("4. Handles signed data according to app needs")
}

// checkStatusExample shows monitoring capabilities
func checkStatusExample() {
	fmt.Println("📊 MyOtherCLI - Check Status Example")
	fmt.Println("This would check the status of signing requests")
	fmt.Println()
	fmt.Println("In a real implementation, you could:")
	fmt.Println("• Make HTTP requests to /status endpoint")
	fmt.Println("• Access server.ListSigningRequests() if embedded")
	fmt.Println("• Monitor WebSocket connections")
	fmt.Println("• Integrate with your app's dashboard/UI")
}

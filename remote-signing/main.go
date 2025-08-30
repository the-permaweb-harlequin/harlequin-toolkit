//go:build !example
// +build !example

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

func main() {
	// Parse command line flags
	var (
		port        = flag.Int("port", 8080, "Server port")
		host        = flag.String("host", "localhost", "Server host")
		timeout     = flag.Int("timeout", 30, "Signing timeout in minutes")
		maxSize     = flag.Int64("max-size", 10*1024*1024, "Maximum data item size in bytes")
		frontendURL = flag.String("frontend-url", "", "Frontend URL for development (e.g., http://localhost:5173)")
		help        = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Create server configuration
	config := server.DefaultConfig()
	config.Port = *port
	config.Host = *host
	config.SigningTimeout = time.Duration(*timeout) * time.Minute
	config.MaxDataSize = *maxSize
	config.FrontendURL = *frontendURL

	// Create and start server
	s := server.New(config)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("🛑 Received shutdown signal...")
		cancel()
	}()

	// Print startup information
	log.Printf("🎭 Starting Harlequin Remote Signing Server...")
	log.Printf("📡 Server will be available at: http://%s:%d", config.Host, config.Port)
	if config.FrontendURL != "" {
		log.Printf("🌐 Frontend URL: %s", config.FrontendURL)
		log.Printf("📝 Signing interface: %s/sign/<uuid>", config.FrontendURL)
	} else {
		log.Printf("📝 Signing interface: http://%s:%d/sign/<uuid>", config.Host, config.Port)
	}
	log.Printf("🔌 WebSocket endpoint: ws://%s:%d/ws", config.Host, config.Port)
	log.Printf("📚 API docs: http://%s:%d/api-docs", config.Host, config.Port)
	log.Println()

	// Start the server
	if err := s.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func printHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server")
	fmt.Println()
	fmt.Println("A server for remote signing of ANS-104 data items via web interface")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  remote-signing [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -port <port>           Server port (default: 8080)")
	fmt.Println("  -host <host>           Server host (default: localhost)")
	fmt.Println("  -timeout <minutes>     Signing timeout in minutes (default: 30)")
	fmt.Println("  -max-size <bytes>      Maximum data item size in bytes (default: 10MB)")
	fmt.Println("  -frontend-url <url>    Frontend URL for development (e.g., http://localhost:5173)")
	fmt.Println("  -help                  Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  remote-signing")
	fmt.Println("  remote-signing -port 9000")
	fmt.Println("  remote-signing -host 0.0.0.0 -port 8080")
	fmt.Println("  remote-signing -frontend-url http://localhost:5173")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Submit data items via POST / to get a signing UUID")
	fmt.Println("  2. Open the signing URL in a browser to sign with wallet")
	fmt.Println("  3. Retrieve signed data via WebSocket or callback")
}

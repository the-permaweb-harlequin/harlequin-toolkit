//go:build example
// +build example

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/everFinance/goar/types"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

// CustomConfig demonstrates how to use the SigningServer with custom configuration
// including custom tags, target address, and anchor.
func main() {
	// Create custom server configuration
	config := &server.Config{
		Port:           9090,                    // Custom port
		Host:          "localhost",
		AllowedOrigins: []string{"*"},
		MaxDataSize:   5 * 1024 * 1024,         // 5MB limit
		SigningTimeout: 15 * time.Minute,       // 15 minute timeout
	}

	// Create signing server with custom config
	signingServer := server.NewSigningServer(config)
	defer signingServer.Close()

	// Create custom data
	customData := []byte("This is custom data with specific configuration!")

	// Create upload request with custom parameters
	uploadReq := &server.UploadRequest{
		Data:     customData,
		Filename: "custom-data.txt",
		Tags: []types.Tag{
			{Name: "Content-Type", Value: "text/plain"},
			{Name: "App-Name", Value: "Custom-Example"},
			{Name: "App-Version", Value: "2.0.0"},
			{Name: "Custom-Tag", Value: "Custom-Value"},
			{Name: "Timestamp", Value: time.Now().Format(time.RFC3339)},
		},
		Target: "custom-target-address", // Custom target address
		Anchor: "custom-anchor-32-chars-long", // Custom anchor (32 chars)
	}

	fmt.Println("🚀 Starting custom upload and signing process...")
	fmt.Printf("📁 File: %s (%d bytes)\n", uploadReq.Filename, len(uploadReq.Data))
	fmt.Printf("🎯 Target: %s\n", uploadReq.Target)
	fmt.Printf("⚓ Anchor: %s\n", uploadReq.Anchor)
	fmt.Printf("🏷️  Tags: %d tags\n", len(uploadReq.Tags))

	// Upload and sign
	result, err := signingServer.Upload(uploadReq)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Print results
	fmt.Println("✅ Custom upload and signing completed successfully!")
	fmt.Printf("🆔 Request UUID: %s\n", result.UUID)
	fmt.Printf("🆔 DataItem ID: %s\n", result.DataItemID)
	fmt.Printf("🔗 Signing URL: %s\n", result.SigningURL)
	fmt.Printf("📅 Signed at: %s\n", result.SignedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("📤 Bundler response: %s\n", result.BundlerResponse)
}

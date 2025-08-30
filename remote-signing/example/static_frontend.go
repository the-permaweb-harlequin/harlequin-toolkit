//go:build example
// +build example

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/everFinance/goar/types"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

// StaticFrontendExample demonstrates how to configure the server
// to work with a separately deployed frontend.
func main() {
	// Create server configuration
	config := server.DefaultConfig()
	config.Port = 8080
	config.Host = "localhost"

	// Configure the frontend URL (deployed separately)
	// In production, this would be your static site URL
	config.FrontendURL = "http://localhost:5173" // Local development frontend URL

	// Create signing server
	signingServer := server.NewSigningServer(config)
	defer signingServer.Close()

	// Read a file to sign
	fileData, err := os.ReadFile("example.txt")
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Create upload request
	uploadReq := &server.UploadRequest{
		Data:     fileData,
		Filename: "example.txt",
		Tags: []types.Tag{
			{Name: "Content-Type", Value: "text/plain"},
			{Name: "App-Name", Value: "Example-App"},
			{Name: "App-Version", Value: "1.0.0"},
		},
		Target: "",
		Anchor: "",
	}

	fmt.Println("🚀 Starting upload and signing process...")
	fmt.Printf("📁 File: %s (%d bytes)\n", uploadReq.Filename, len(uploadReq.Data))
	fmt.Printf("🌐 Frontend URL: %s\n", config.FrontendURL)
	fmt.Printf("🔧 Server URL: http://%s:%d\n", config.Host, config.Port)

	// Upload and sign
	result, err := signingServer.Upload(uploadReq)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Print results
	fmt.Println("✅ Upload and signing completed successfully!")
	fmt.Printf("🆔 Request UUID: %s\n", result.UUID)
	fmt.Printf("🆔 DataItem ID: %s\n", result.DataItemID)
	fmt.Printf("🔗 Signing URL: %s\n", result.SigningURL)
	fmt.Printf("📅 Signed at: %s\n", result.SignedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("📤 Bundler response: %s\n", result.BundlerResponse)

	fmt.Println("\n💡 The signing URL now points to your separate frontend with the server parameter!")
}

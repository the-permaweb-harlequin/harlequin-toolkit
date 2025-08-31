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

// ProductionExample demonstrates a realistic deployment scenario
// where the frontend is hosted separately from the backend.
func main() {
	// Create server configuration for production
	config := server.DefaultConfig()
	config.Port = 8080
	config.Host = "0.0.0.0" // Listen on all interfaces

	// Production scenario:
	// - Frontend: https://signing.harlequin.com (static site on CDN)
	// - Backend: https://api.harlequin.com (Go server)
	config.FrontendURL = "https://signing.harlequin.com"

	// Note: In real deployment, you'd also set:
	// config.Host = "api.harlequin.com" // or your server's public hostname

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
			{Name: "App-Name", Value: "Harlequin-Signing"},
			{Name: "App-Version", Value: "1.0.0"},
		},
		Target: "",
		Anchor: "",
	}

	fmt.Println("🚀 Production Deployment Example")
	fmt.Println("=================================")
	fmt.Printf("📁 File: %s (%d bytes)\n", uploadReq.Filename, len(uploadReq.Data))
	fmt.Printf("🌐 Frontend URL: %s\n", config.FrontendURL)
	fmt.Printf("🔧 Backend URL: http://%s:%d\n", config.Host, config.Port)
	fmt.Println()

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

	fmt.Println("\n💡 Deployment Flow:")
	fmt.Println("1. User visits the signing URL (frontend)")
	fmt.Println("2. Frontend reads the 'server' parameter")
	fmt.Println("3. Frontend makes API calls to the backend")
	fmt.Println("4. Backend handles signing and bundler upload")
	fmt.Println()
	fmt.Println("🔗 Example URLs:")
	fmt.Printf("   Frontend: %s/sign/%s\n", config.FrontendURL, result.UUID)
	fmt.Printf("   With Server Param: %s\n", result.SigningURL)
}

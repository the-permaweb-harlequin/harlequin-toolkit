package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/gin-gonic/gin"
)

// UploadRequest represents a request to upload and sign data
type UploadRequest struct {
	Data     []byte            // Raw data to be signed
	Filename string            // Original filename
	Tags     []types.Tag       // Tags to add to the DataItem
	Target   string            // Target address (optional)
	Anchor   string            // Anchor (optional, will generate if empty)
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	UUID            string    // Request UUID
	DataItemID      string    // Arweave DataItem ID (after signing)
	SigningURL      string    // URL for signing interface
	TransactionID   string    // Final Arweave transaction ID (after bundler upload)
	SignedAt        time.Time // When the item was signed
	BundlerResponse string    // Response from bundler
}

// SigningServer provides a high-level API for remote signing operations
type SigningServer struct {
	config *Config
	server *Server
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSigningServer creates a new SigningServer instance
func NewSigningServer(config *Config) *SigningServer {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &SigningServer{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Upload uploads data for signing and waits for completion
func (ss *SigningServer) Upload(req *UploadRequest) (*UploadResult, error) {
	// Ensure server is running
	if err := ss.ensureServerRunning(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Create DataItem
	dataItemBytes, err := ss.createDataItem(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create DataItem: %w", err)
	}

	// Upload to server
	uploadResp, err := ss.uploadToServer(dataItemBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to server: %w", err)
	}

	// Open browser for signing
	if err := ss.openBrowser(uploadResp.SigningURL); err != nil {
		// Don't fail on browser open error, just log it
		fmt.Printf("⚠️  Failed to open browser: %v\n", err)
		fmt.Printf("💡 Please manually open: %s\n", uploadResp.SigningURL)
	}

	// Wait for signing completion
	signedData, err := ss.waitForSigning(uploadResp.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for signing: %w", err)
	}

	// Extract DataItem ID
	dataItemID, err := ss.extractDataItemID(signedData)
	if err != nil {
		return nil, fmt.Errorf("failed to extract DataItem ID: %w", err)
	}

	// Upload to bundler
	bundlerResponse, err := ss.uploadToBundler(signedData)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to bundler: %w", err)
	}

	return &UploadResult{
		UUID:            uploadResp.UUID,
		DataItemID:      dataItemID,
		SigningURL:      uploadResp.SigningURL,
		TransactionID:   dataItemID, // Same as DataItem ID for now
		SignedAt:        time.Now(),
		BundlerResponse: bundlerResponse,
	}, nil
}

// Close stops the signing server
func (ss *SigningServer) Close() error {
	if ss.cancel != nil {
		ss.cancel()
	}
	if ss.server != nil {
		return ss.server.Stop()
	}
	return nil
}

// ensureServerRunning starts the server if it's not already running
func (ss *SigningServer) ensureServerRunning() error {
	// Check if server is already running
	client := &http.Client{Timeout: 2 * time.Second}
	serverURL := fmt.Sprintf("http://%s:%d", ss.config.Host, ss.config.Port)

	resp, err := client.Get(serverURL + "/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		return nil // Server is already running
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Start server
	ss.server = New(ss.config)

	// Set Gin mode
	if gin.Mode() != gin.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start server in background
	go func() {
		if err := ss.server.Start(ss.ctx); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for server to be ready
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		resp, err := client.Get(serverURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	return fmt.Errorf("server failed to start within timeout")
}

// createDataItem creates a DataItem with the provided data and tags
func (ss *SigningServer) createDataItem(req *UploadRequest) ([]byte, error) {
	// Generate anchor if not provided
	anchor := req.Anchor
	if anchor == "" {
		anchorBytes := make([]byte, 32)
		if _, err := rand.Read(anchorBytes); err != nil {
			return nil, fmt.Errorf("failed to generate anchor: %w", err)
		}
		// Convert to hex string and ensure it's exactly 32 characters
		anchor = fmt.Sprintf("%064x", anchorBytes)[:32]
	}

	// Create the DataItem structure for JSON serialization
	dataItem := map[string]interface{}{
		"signatureType": 1, // ArweaveSignature
		"owner":         "", // Will be filled by wallet during signing
		"target":        req.Target,
		"anchor":        anchor,
		"tags":          req.Tags,
		"data":          string(req.Data),
	}

	// Serialize to JSON
	dataItemBytes, err := json.Marshal(dataItem)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DataItem: %w", err)
	}

	return dataItemBytes, nil
}

// uploadToServer uploads the DataItem to the signing server
func (ss *SigningServer) uploadToServer(dataItemBytes []byte) (*struct {
	UUID       string `json:"uuid"`
	SigningURL string `json:"signing_url"`
	Message    string `json:"message"`
}, error) {
	serverURL := fmt.Sprintf("http://%s:%d", ss.config.Host, ss.config.Port)

	resp, err := http.Post(serverURL, "application/octet-stream", bytes.NewReader(dataItemBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to upload to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var uploadResp struct {
		UUID       string `json:"uuid"`
		SigningURL string `json:"signing_url"`
		Message    string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &uploadResp, nil
}

// openBrowser opens the signing URL in the default browser
func (ss *SigningServer) openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// waitForSigning waits for the signing to complete and returns the signed data
func (ss *SigningServer) waitForSigning(uuid string) ([]byte, error) {
	serverURL := fmt.Sprintf("http://%s:%d", ss.config.Host, ss.config.Port)

	// Connect to SSE stream
	sseURL := fmt.Sprintf("%s/events/%s", serverURL, uuid)
	resp, err := http.Get(sseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSE stream: %w", err)
	}
	defer resp.Body.Close()

	// Parse SSE stream
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent, currentData string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimPrefix(line, "event:")
		} else if strings.HasPrefix(line, "data:") {
			currentData = strings.TrimPrefix(line, "data:")
		} else if line == "" && currentEvent != "" && currentData != "" {
			// Process complete event
			if currentEvent == "signed" {
				// Fetch signed data from separate endpoint
				signedDataURL := fmt.Sprintf("%s/signed/%s", serverURL, uuid)
				signedResp, err := http.Get(signedDataURL)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch signed data: %w", err)
				}
				defer signedResp.Body.Close()

				signedData, err := io.ReadAll(signedResp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read signed data: %w", err)
				}

				return signedData, nil
			}

			currentEvent = ""
			currentData = ""
		}
	}

	return nil, fmt.Errorf("signing timeout or connection lost")
}

// extractDataItemID extracts the DataItem ID from signed binary data
func (ss *SigningServer) extractDataItemID(signedData []byte) (string, error) {
	dataItem, err := utils.DecodeBundleItem(signedData)
	if err != nil {
		// Fallback to first 32 bytes if parsing fails
		if len(signedData) >= 32 {
			return string(signedData[:32]), nil
		}
		return "", fmt.Errorf("failed to decode DataItem: %w", err)
	}

	return dataItem.Id, nil
}

// uploadToBundler uploads the signed DataItem to the ArDrive bundler
func (ss *SigningServer) uploadToBundler(signedData []byte) (string, error) {
	bundlerURL := "https://upload.ardrive.io/tx"

	req, err := http.NewRequest("POST", bundlerURL, bytes.NewReader(signedData))
	if err != nil {
		return "", fmt.Errorf("failed to create bundler request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(signedData)))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to bundler: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("bundler upload failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

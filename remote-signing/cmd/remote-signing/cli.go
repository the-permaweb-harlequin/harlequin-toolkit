package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

// Global variable to store server cleanup function
var serverCleanup context.CancelFunc

// CLIConfig represents the CLI configuration for the remote signing server
type CLIConfig struct {
	Port           int      `json:"port"`
	Host           string   `json:"host"`
	AllowedOrigins []string `json:"allowed_origins"`
	MaxDataSize    int64    `json:"max_data_size"`
	SigningTimeout int      `json:"signing_timeout_minutes"`
	ConfigFile     string   `json:"-"`
	TemplatesPath  string   `json:"templates_path"`
	Debug          bool     `json:"debug"`
}

// DefaultCLIConfig returns the default CLI configuration
func DefaultCLIConfig() *CLIConfig {
	return &CLIConfig{
		Port:           8080,
		Host:          "localhost",
		AllowedOrigins: []string{"*"},
		MaxDataSize:   10 * 1024 * 1024, // 10MB
		SigningTimeout: 30,               // 30 minutes
		ConfigFile:    "",
		TemplatesPath: "",
		Debug:         false,
	}
}

// ToServerConfig converts CLIConfig to server.Config
func (c *CLIConfig) ToServerConfig() *server.Config {
	return &server.Config{
		Port:           c.Port,
		Host:          c.Host,
		AllowedOrigins: c.AllowedOrigins,
		MaxDataSize:    c.MaxDataSize,
		SigningTimeout: time.Duration(c.SigningTimeout) * time.Minute,
	}
}

// StartCommand starts the remote signing server
func StartCommand(args []string) error {
	config := DefaultCLIConfig()

	// Parse command line arguments
	if err := parseStartArgs(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Load config file if specified
	if config.ConfigFile != "" {
		if err := loadConfigFile(config.ConfigFile, config); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Set default templates path if not specified
	if config.TemplatesPath == "" {
		// Look for templates in the binary directory
		execPath, err := os.Executable()
		if err == nil {
			config.TemplatesPath = filepath.Join(filepath.Dir(execPath), "templates")
		}

		// Fallback to current directory
		if _, err := os.Stat(config.TemplatesPath); os.IsNotExist(err) {
			if pwd, err := os.Getwd(); err == nil {
				config.TemplatesPath = filepath.Join(pwd, "templates")
			}
		}
	}

	// Create server
	srv := server.New(config.ToServerConfig())

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Received shutdown signal...")
		cancel()
	}()

	// Start server
	fmt.Printf("🎭 Starting Harlequin Remote Signing Server...\n")
	fmt.Printf("📡 Server will be available at: http://%s:%d\n", config.Host, config.Port)
	if config.TemplatesPath != "" {
		fmt.Printf("📝 Signing interface: http://%s:%d/sign/<uuid>\n", config.Host, config.Port)
	}
	fmt.Printf("🔌 WebSocket endpoint: ws://%s:%d/ws\n", config.Host, config.Port)
	fmt.Printf("⏰ Signing timeout: %d minutes\n", config.SigningTimeout)
	fmt.Printf("📦 Max data size: %d bytes\n", config.MaxDataSize)
	if config.Debug {
		fmt.Printf("🐛 Debug mode enabled\n")
	}
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()

	if err := srv.StartWithTemplates(ctx, config.TemplatesPath); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// StopCommand stops the remote signing server (placeholder for daemon mode)
func StopCommand(args []string) error {
	fmt.Println("🛑 Stop command is not yet implemented for daemon mode")
	fmt.Println("To stop the server, use Ctrl+C in the terminal where it's running")
	return nil
}

// StatusCommand shows the status of the remote signing server
func StatusCommand(args []string) error {
	config := DefaultCLIConfig()

	// Parse arguments for host and port
	if err := parseStatusArgs(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Make HTTP request to status endpoint
	url := fmt.Sprintf("http://%s:%d/status", config.Host, config.Port)

	// TODO: Implement HTTP client request to get status
	fmt.Printf("🔍 Checking server status at %s\n", url)
	fmt.Println("Status command not yet fully implemented")

	return nil
}

// UploadCommand uploads a file to the remote signing server and opens the signing interface
func UploadCommand(args []string) error {
	config := DefaultCLIConfig()
	var filename string
	var shouldWait bool = true

	// Parse command line arguments
	if err := parseUploadArgs(args, config, &filename, &shouldWait); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	if filename == "" {
		return fmt.Errorf("file path is required")
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filename)
	}

	// Read file contents
	fmt.Printf("📁 Reading file: %s\n", filename)
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Printf("📊 File size: %d bytes\n", len(fileData))

	// Create signing server
	signingServer := server.NewSigningServer(config.ToServerConfig())
	defer signingServer.Close()

	// Create upload request
	uploadReq := &server.UploadRequest{
		Data:     fileData,
		Filename: filename,
		Tags: []types.Tag{
			{Name: "Content-Type", Value: "text/plain"},
			{Name: "Filename", Value: filename},
		},
		Target: "",
		Anchor: "",
	}

	// Upload and sign
	fmt.Printf("🚀 Starting upload and signing process...\n")
	result, err := signingServer.Upload(uploadReq)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Print results
	fmt.Printf("✅ Upload and signing completed successfully!\n")
	fmt.Printf("🆔 Request UUID: %s\n", result.UUID)
	fmt.Printf("🆔 DataItem ID: %s\n", result.DataItemID)
	fmt.Printf("🔗 Signing URL: %s\n", result.SigningURL)
	fmt.Printf("📅 Signed at: %s\n", result.SignedAt.Format(time.RFC3339))
	fmt.Printf("📤 Bundler response: %s\n", result.BundlerResponse)

	return nil
}

// parseUploadArgs parses command line arguments for the upload command
func parseUploadArgs(args []string, config *CLIConfig, filename *string, shouldWait *bool) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 >= len(args) {
				return fmt.Errorf("--port requires a value")
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid port number: %s", args[i+1])
			}
			config.Port = port
			i++
		case "--host", "-h":
			if i+1 >= len(args) {
				return fmt.Errorf("--host requires a value")
			}
			config.Host = args[i+1]
			i++
		case "--no-wait":
			*shouldWait = false
		case "--debug":
			config.Debug = true
		case "--help":
			printUploadHelp()
			os.Exit(0)
		default:
			if *filename == "" {
				*filename = args[i]
			} else {
				return fmt.Errorf("unknown argument: %s", args[i])
			}
		}
	}
	return nil
}

// getContentType determines content type based on file extension
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

// serializeDataItem serializes a DataItem into bytes for transmission
func serializeDataItem(item map[string]interface{}) ([]byte, error) {
	return json.Marshal(item)
}

// openBrowser opens the given URL in the default browser
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// waitForSigningViaSSE waits for the signing process to complete via Server-Sent Events
func waitForSigningViaSSE(config *CLIConfig, uuid string) error {
	sseURL := fmt.Sprintf("http://%s:%d/events/%s", config.Host, config.Port, uuid)

	// Create HTTP request for SSE
	req, err := http.NewRequest("GET", sseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Printf("\n🛑 Stopping wait for signing...\n")
		cancel()
	}()

	// Set context with timeout
	req = req.WithContext(ctx)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE stream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE request failed with status: %d", resp.StatusCode)
	}

	// Check content type
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return fmt.Errorf("server did not return SSE stream")
	}

	fmt.Printf("📡 Connected to SSE stream for UUID: %s\n", uuid)

	// Check if signing is already complete (fallback)
	statusURL := fmt.Sprintf("http://%s:%d/status", config.Host, config.Port)
	statusResp, err := http.Get(statusURL)
	if err == nil {
		defer statusResp.Body.Close()
		var statusData map[string]interface{}
		if json.NewDecoder(statusResp.Body).Decode(&statusData) == nil {
			if requests, ok := statusData["requests"].(map[string]interface{}); ok {
				if total, ok := requests["total"].(float64); ok && total > 0 {
					if config.Debug {
						fmt.Printf("📊 Server has %d active requests\n", int(total))
					}
				}
			}
		}
	}

	// Small delay to ensure we're ready to receive events
	time.Sleep(100 * time.Millisecond)

		// Read SSE events
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent string
	var currentData string

	// Set a timeout for reading events
	timeout := time.After(30 * time.Second)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		case <-timeout:
			fmt.Printf("⏰ Timeout waiting for signing completion\n")
			return fmt.Errorf("timeout waiting for signing completion")
		default:
		}

		line := scanner.Text()

		if config.Debug {
			fmt.Printf("🔍 Raw SSE line: '%s'\n", line)
		}

				// Empty line indicates end of event
		if line == "" {
			if config.Debug {
				fmt.Printf("🔍 Processing event: '%s' with data: '%s'\n", currentEvent, currentData)
			}
			// Process the complete event
		if currentEvent == "signed" {
			fmt.Printf("🎉 Data signed successfully!\n")

			// Fetch signed binary data from the server
			signedDataURL := fmt.Sprintf("http://localhost:%d/signed/%s", config.Port, uuid)
			if config.Debug {
				fmt.Printf("📥 Fetching signed data from: %s\n", signedDataURL)
			}

			resp, err := http.Get(signedDataURL)
			if err != nil {
				fmt.Printf("⚠️  Failed to fetch signed data: %v\n", err)
				fmt.Printf("🆔 Request ID: %s\n", uuid)
				return nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("⚠️  Failed to fetch signed data: HTTP %d\n", resp.StatusCode)
				fmt.Printf("🆔 Request ID: %s\n", uuid)
				return nil
			}

			signedData, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("⚠️  Failed to read signed data: %v\n", err)
				fmt.Printf("🆔 Request ID: %s\n", uuid)
				return nil
			}

			fmt.Printf("📦 Received signed data: %d bytes\n", len(signedData))
			if config.Debug {
				fmt.Printf("🔍 Signed data (first 100 bytes): %x\n", signedData[:min(100, len(signedData))])
			}

			// Use goar/utils.DecodeBundleItem to properly parse the signed binary data
			dataItem, err := utils.DecodeBundleItem(signedData)
			if err == nil {
				// Extract the actual DataItem ID from the parsed structure
							fmt.Printf("🆔 DataItem ID: %s\n", dataItem.Id)
			if config.Debug {
				fmt.Printf("🔍 DataItem details:\n")
				fmt.Printf("  - Owner: %s\n", dataItem.Owner)
				fmt.Printf("  - Target: %s\n", dataItem.Target)
				fmt.Printf("  - Anchor: %s\n", dataItem.Anchor)
				fmt.Printf("  - Signature Type: %d\n", dataItem.SignatureType)
				fmt.Printf("  - Data size: %d bytes\n", len(dataItem.Data))
			}

			// Upload to bundler
			fmt.Printf("📤 Uploading to ArDrive bundler...\n")
			bundlerURL := "https://upload.ardrive.io/tx"

			req, err := http.NewRequest("POST", bundlerURL, bytes.NewReader(signedData))
			if err != nil {
				fmt.Printf("⚠️  Failed to create bundler request: %v\n", err)
			} else {
				req.Header.Set("Content-Type", "application/octet-stream")
				req.Header.Set("Content-Length", fmt.Sprintf("%d", len(signedData)))

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					fmt.Printf("⚠️  Failed to upload to bundler: %v\n", err)
				} else {
					defer resp.Body.Close()

					// Read response body once
					body, _ := io.ReadAll(resp.Body)

					if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
						fmt.Printf("✅ Successfully uploaded to ArDrive bundler!\n")
						if config.Debug {
							fmt.Printf("🔍 Bundler response status: %d\n", resp.StatusCode)
							fmt.Printf("🔍 Bundler response body: %s\n", string(body))
							fmt.Printf("🔍 Bundler response headers: %v\n", resp.Header)
						}
					} else {
						fmt.Printf("⚠️  Bundler upload failed (HTTP %d): %s\n", resp.StatusCode, string(body))
						if config.Debug {
							fmt.Printf("🔍 Bundler response headers: %v\n", resp.Header)
						}
					}
				}
			}
			} else {
				fmt.Printf("⚠️  Failed to decode DataItem: %v\n", err)
				// Fallback to first 32 bytes if parsing fails
				if len(signedData) >= 32 {
					dataItemID := base64.StdEncoding.EncodeToString(signedData[:32])
					fmt.Printf("🆔 DataItem ID (fallback): %s\n", dataItemID)
				}
			}

			fmt.Printf("🆔 Request ID: %s\n", uuid)
			return nil
		} else if currentEvent == "connected" {
			fmt.Printf("📊 Status update: connected\n")
		} else if currentEvent == "heartbeat" {
			// Heartbeat received, connection is alive
			if config.Debug {
				fmt.Printf("💓 Heartbeat received\n")
			}
		}

			// Reset for next event
			currentEvent = ""
			currentData = ""
			continue
		}

		// Parse SSE event fields
		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimPrefix(line, "event:")
			if config.Debug {
				fmt.Printf("📡 Set currentEvent to: '%s'\n", currentEvent)
			}
		} else if strings.HasPrefix(line, "data:") {
			currentData = strings.TrimPrefix(line, "data:")
			if config.Debug {
				fmt.Printf("📄 Set currentData to: '%s'\n", currentData)
			}

			// Parse event data if needed
			if currentData != "" {
				var eventData map[string]interface{}
				if err := json.Unmarshal([]byte(currentData), &eventData); err == nil {
					if clientID, ok := eventData["client_id"].(string); ok {
						if config.Debug {
							fmt.Printf("📊 Client ID: %s\n", clientID)
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SSE stream: %w", err)
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseStartArgs parses command line arguments for the start command
func parseStartArgs(args []string, config *CLIConfig) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 >= len(args) {
				return fmt.Errorf("--port requires a value")
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid port number: %s", args[i+1])
			}
			config.Port = port
			i++
		case "--host", "-h":
			if i+1 >= len(args) {
				return fmt.Errorf("--host requires a value")
			}
			config.Host = args[i+1]
			i++
		case "--config", "-c":
			if i+1 >= len(args) {
				return fmt.Errorf("--config requires a value")
			}
			config.ConfigFile = args[i+1]
			i++
		case "--timeout", "-t":
			if i+1 >= len(args) {
				return fmt.Errorf("--timeout requires a value")
			}
			timeout, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid timeout value: %s", args[i+1])
			}
			config.SigningTimeout = timeout
			i++
		case "--max-size":
			if i+1 >= len(args) {
				return fmt.Errorf("--max-size requires a value")
			}
			maxSize, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid max size value: %s", args[i+1])
			}
			config.MaxDataSize = maxSize
			i++
		case "--templates":
			if i+1 >= len(args) {
				return fmt.Errorf("--templates requires a value")
			}
			config.TemplatesPath = args[i+1]
			i++
		case "--debug":
			config.Debug = true
		case "--help":
			printStartHelp()
			os.Exit(0)
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	return nil
}

// parseStatusArgs parses command line arguments for the status command
func parseStatusArgs(args []string, config *CLIConfig) error {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 >= len(args) {
				return fmt.Errorf("--port requires a value")
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid port number: %s", args[i+1])
			}
			config.Port = port
			i++
		case "--host", "-h":
			if i+1 >= len(args) {
				return fmt.Errorf("--host requires a value")
			}
			config.Host = args[i+1]
			i++
		case "--help":
			printStatusHelp()
			os.Exit(0)
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	return nil
}

// loadConfigFile loads configuration from a JSON file
func loadConfigFile(filename string, config *CLIConfig) error {
	// Expand relative path
	if !filepath.IsAbs(filename) {
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		filename = filepath.Join(pwd, filename)
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, config)
}

// printStartHelp prints help for the start command
func printStartHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server - Start Command")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  remote-signing start [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --port <port>         Server port (default: 8080)")
	fmt.Println("  -h, --host <host>         Server host (default: localhost)")
	fmt.Println("  -c, --config <file>       Configuration file path")
	fmt.Println("  -t, --timeout <minutes>   Signing timeout in minutes (default: 30)")
	fmt.Println("      --max-size <bytes>    Maximum data size in bytes (default: 10MB)")
	fmt.Println("      --templates <path>    Path to HTML templates directory")
	fmt.Println("      --debug               Enable debug output")
	fmt.Println("      --help                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  remote-signing start")
	fmt.Println("  remote-signing start --port 9000")
	fmt.Println("  remote-signing start --host 0.0.0.0 --port 8080")
	fmt.Println("  remote-signing start --config ./signing-config.json")
	fmt.Println()
}

// printStatusHelp prints help for the status command
func printStatusHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server - Status Command")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  remote-signing status [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --port <port>  Server port (default: 8080)")
	fmt.Println("  -h, --host <host>  Server host (default: localhost)")
	fmt.Println("      --help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  remote-signing status")
	fmt.Println("  remote-signing status --port 9000")
	fmt.Println()
}

// printUploadHelp prints help for the upload command
func printUploadHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server - Upload Command")
	fmt.Println()
	fmt.Println("Upload a file to the remote signing server and open the signing interface")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  remote-signing upload <file> [flags]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <file>                Path to file to upload")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --port <port>     Server port (default: 8080)")
	fmt.Println("  -h, --host <host>     Server host (default: localhost)")
	fmt.Println("      --no-wait         Don't wait for signing completion")
	fmt.Println("      --debug           Enable debug output")
	fmt.Println("      --help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  remote-signing upload ./data.json")
	fmt.Println("  remote-signing upload ./image.png --port 9000")
	fmt.Println("  remote-signing upload ./document.pdf --no-wait")
	fmt.Println("  remote-signing upload ./large-file.bin --host remote.example.com")
	fmt.Println()
	fmt.Println("Workflow:")
	fmt.Println("  1. Reads the specified file")
	fmt.Println("  2. Uploads data to the remote signing server")
	fmt.Println("  3. Opens the signing URL in your default browser")
	fmt.Println("  4. Waits for you to sign the data with your wallet")
	fmt.Println("  5. Reports completion when signing is done")
	fmt.Println()
}

// printRemoteSigningHelp prints general help for the remote-signing command
func printRemoteSigningHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server")
	fmt.Println()
	fmt.Println("A server for remote signing of raw data via web interface")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  remote-signing <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start     Start the remote signing server")
	fmt.Println("  upload    Upload a file and open signing interface")
	fmt.Println("  stop      Stop the remote signing server (daemon mode)")
	fmt.Println("  status    Check server status")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  remote-signing start --port 8080")
	fmt.Println("  remote-signing upload ./data.json")
	fmt.Println("  remote-signing status")
	fmt.Println("  remote-signing help")
	fmt.Println()
	fmt.Println("For detailed command help, use:")
	fmt.Println("  remote-signing <command> --help")
	fmt.Println()
}

// HandleRemoteSigningCommand handles the remote-signing CLI command
func HandleRemoteSigningCommand(args []string) {
	if len(args) == 0 {
		printRemoteSigningHelp()
		return
	}

	command := args[0]
	commandArgs := args[1:]

	var err error
	switch command {
	case "start":
		err = StartCommand(commandArgs)
	case "upload":
		err = UploadCommand(commandArgs)
	case "stop":
		err = StopCommand(commandArgs)
	case "status":
		err = StatusCommand(commandArgs)
	case "help", "--help", "-h":
		printRemoteSigningHelp()
		return
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printRemoteSigningHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

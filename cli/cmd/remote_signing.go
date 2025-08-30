package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

// RemoteSigningConfig represents the configuration for the remote signing server
type RemoteSigningConfig struct {
	Port           int      `json:"port"`
	Host           string   `json:"host"`
	AllowedOrigins []string `json:"allowed_origins"`
	MaxDataSize    int64    `json:"max_data_size"`
	SigningTimeout int      `json:"signing_timeout_minutes"`
	FrontendURL    string   `json:"frontend_url"` // URL for the frontend (for development)
}

// DefaultRemoteSigningConfig returns the default configuration
func DefaultRemoteSigningConfig() *RemoteSigningConfig {
	return &RemoteSigningConfig{
		Port:           8080,
		Host:          "localhost",
		AllowedOrigins: []string{"*"},
		MaxDataSize:   10 * 1024 * 1024, // 10MB
		SigningTimeout: 30,               // 30 minutes
		FrontendURL:    "",               // Empty by default (uses same host)
	}
}

// HandleRemoteSigningCommand handles the remote-signing CLI command
func HandleRemoteSigningCommand(ctx context.Context, args []string) {
	if len(args) == 0 {
		printRemoteSigningHelp()
		return
	}

	command := args[0]
	commandArgs := args[1:]

	var err error
	switch command {
	case "start":
		err = startRemoteSigningServer(ctx, commandArgs)
	case "stop":
		err = stopRemoteSigningServer(commandArgs)
	case "status":
		err = checkRemoteSigningStatus(commandArgs)
	case "help", "--help", "-h":
		printRemoteSigningHelp()
		return
	default:
		fmt.Printf("Unknown remote-signing command: %s\n\n", command)
		printRemoteSigningHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// startRemoteSigningServer starts the remote signing server
func startRemoteSigningServer(ctx context.Context, args []string) error {
	config := DefaultRemoteSigningConfig()

	// Parse command line arguments
	if err := parseRemoteSigningArgs(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Get the path to the remote-signing binary
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Look for remote-signing binary in the same directory as the CLI
	remoteSigningPath := filepath.Join(filepath.Dir(execPath), "remote-signing")
	if _, err := os.Stat(remoteSigningPath); os.IsNotExist(err) {
		// Try in the project root for development
		projectRoot := findProjectRoot()
		if projectRoot != "" {
			remoteSigningPath = filepath.Join(projectRoot, "remote-signing", "remote-signing")
			if _, err := os.Stat(remoteSigningPath); os.IsNotExist(err) {
				return fmt.Errorf("remote-signing binary not found. Please ensure it's built and available")
			}
		} else {
			return fmt.Errorf("remote-signing binary not found at %s", remoteSigningPath)
		}
	}

	// Build command arguments
	cmdArgs := []string{"start"}
	cmdArgs = append(cmdArgs, "--port", strconv.Itoa(config.Port))
	cmdArgs = append(cmdArgs, "--host", config.Host)
	cmdArgs = append(cmdArgs, "--timeout", strconv.Itoa(config.SigningTimeout))
	cmdArgs = append(cmdArgs, "--max-size", strconv.FormatInt(config.MaxDataSize, 10))
	if config.FrontendURL != "" {
		cmdArgs = append(cmdArgs, "--frontend-url", config.FrontendURL)
	}

	// Create and start the command
	cmd := exec.CommandContext(ctx, remoteSigningPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("🎭 Starting Harlequin Remote Signing Server...\n")
	fmt.Printf("📡 Server will be available at: http://%s:%d\n", config.Host, config.Port)
	fmt.Printf("📝 Signing interface: http://%s:%d/sign/<uuid>\n", config.Host, config.Port)
	fmt.Printf("🔌 WebSocket endpoint: ws://%s:%d/ws\n", config.Host, config.Port)
	fmt.Println()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start remote signing server: %w", err)
	}

	// Wait for either the context to be cancelled or a signal
	go func() {
		<-sigCh
		log.Println("Received shutdown signal...")
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	// Wait for the command to finish
	return cmd.Wait()
}

// stopRemoteSigningServer stops the remote signing server
func stopRemoteSigningServer(args []string) error {
	fmt.Println("🛑 Stop command is not yet implemented for daemon mode")
	fmt.Println("To stop the server, use Ctrl+C in the terminal where it's running")
	return nil
}

// checkRemoteSigningStatus checks the status of the remote signing server
func checkRemoteSigningStatus(args []string) error {
	config := DefaultRemoteSigningConfig()

	// Parse arguments for host and port
	if err := parseStatusArgs(args, config); err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// TODO: Implement HTTP client request to get status
	url := fmt.Sprintf("http://%s:%d/status", config.Host, config.Port)
	fmt.Printf("🔍 Checking server status at %s\n", url)
	fmt.Println("Status command not yet fully implemented")

	return nil
}

// parseRemoteSigningArgs parses command line arguments for remote signing commands
func parseRemoteSigningArgs(args []string, config *RemoteSigningConfig) error {
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
		case "--frontend-url":
			if i+1 >= len(args) {
				return fmt.Errorf("--frontend-url requires a value")
			}
			config.FrontendURL = args[i+1]
			i++
		case "--help":
			printRemoteSigningHelp()
			os.Exit(0)
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	return nil
}

// parseStatusArgs parses command line arguments for the status command
func parseStatusArgs(args []string, config *RemoteSigningConfig) error {
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
			printRemoteSigningHelp()
			os.Exit(0)
		default:
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}
	return nil
}

// findProjectRoot finds the project root directory
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check if this directory contains nx.json (indicating project root)
		if _, err := os.Stat(filepath.Join(dir, "nx.json")); err == nil {
			return dir
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return ""
}

// printRemoteSigningHelp prints help for the remote-signing command
func printRemoteSigningHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server")
	fmt.Println()
	fmt.Println("A server for remote signing of ANS-104 data items via web interface")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin remote-signing <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start     Start the remote signing server")
	fmt.Println("  stop      Stop the remote signing server (daemon mode)")
	fmt.Println("  status    Check server status")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Start Command Flags:")
	fmt.Println("  -p, --port <port>        Server port (default: 8080)")
	fmt.Println("  -h, --host <host>        Server host (default: localhost)")
	fmt.Println("  -t, --timeout <minutes>  Signing timeout in minutes (default: 30)")
	fmt.Println("      --max-size <bytes>   Maximum data item size in bytes (default: 10MB)")
	fmt.Println("      --frontend-url <url> Frontend URL for development (e.g., http://localhost:5173)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin remote-signing start")
	fmt.Println("  harlequin remote-signing start --port 9000")
	fmt.Println("  harlequin remote-signing start --host 0.0.0.0 --port 8080")
	fmt.Println("  harlequin remote-signing start --frontend-url http://localhost:5173")
	fmt.Println("  harlequin remote-signing status")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Submit data items via POST / to get a signing UUID")
	fmt.Println("  2. Open the signing URL in a browser to sign with wallet")
	fmt.Println("  3. Retrieve signed data via WebSocket or callback")
	fmt.Println()
	fmt.Println("API Endpoints:")
	fmt.Println("  POST /                Submit data item for signing")
	fmt.Println("  GET /<uuid>           Retrieve unsigned data item")
	fmt.Println("  POST /<uuid>          Submit signed data item")
	fmt.Println("  GET /sign/<uuid>      Web interface for signing")
	fmt.Println("  GET /ws               WebSocket endpoint for callbacks")
	fmt.Println("  GET /status           Server status and statistics")
	fmt.Println()
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// CLIConfig represents the CLI configuration for the remote signing server
type CLIConfig struct {
	Port           int      `json:"port"`
	Host           string   `json:"host"`
	AllowedOrigins []string `json:"allowed_origins"`
	MaxDataSize    int64    `json:"max_data_size"`
	SigningTimeout int      `json:"signing_timeout_minutes"`
	ConfigFile     string   `json:"-"`
	Daemon         bool     `json:"-"`
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
		Daemon:        false,
	}
}

// ToServerConfig converts CLIConfig to ServerConfig
func (c *CLIConfig) ToServerConfig() *ServerConfig {
	return &ServerConfig{
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

	// Create server
	server := NewRemoteSigningServer(config.ToServerConfig())

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
	fmt.Printf("📝 Signing interface: http://%s:%d/sign/<uuid>\n", config.Host, config.Port)
	fmt.Printf("🔌 WebSocket endpoint: ws://%s:%d/ws\n", config.Host, config.Port)
	fmt.Printf("⏰ Signing timeout: %d minutes\n", config.SigningTimeout)
	fmt.Printf("📦 Max data size: %d bytes\n", config.MaxDataSize)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println()

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// StopCommand stops the remote signing server (if running as daemon)
func StopCommand(args []string) error {
	// For now, this is a placeholder since we're not implementing daemon mode yet
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
		case "--daemon", "-d":
			config.Daemon = true
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
	fmt.Println("  harlequin remote-signing start [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --port <port>        Server port (default: 8080)")
	fmt.Println("  -h, --host <host>        Server host (default: localhost)")
	fmt.Println("  -c, --config <file>      Configuration file path")
	fmt.Println("  -t, --timeout <minutes>  Signing timeout in minutes (default: 30)")
	fmt.Println("      --max-size <bytes>   Maximum data item size in bytes (default: 10MB)")
	fmt.Println("  -d, --daemon             Run as daemon (not yet implemented)")
	fmt.Println("      --help               Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin remote-signing start")
	fmt.Println("  harlequin remote-signing start --port 9000")
	fmt.Println("  harlequin remote-signing start --host 0.0.0.0 --port 8080")
	fmt.Println("  harlequin remote-signing start --config ./signing-config.json")
	fmt.Println()
}

// printStatusHelp prints help for the status command
func printStatusHelp() {
	fmt.Println("🎭 Harlequin Remote Signing Server - Status Command")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin remote-signing status [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -p, --port <port>  Server port (default: 8080)")
	fmt.Println("  -h, --host <host>  Server host (default: localhost)")
	fmt.Println("      --help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin remote-signing status")
	fmt.Println("  harlequin remote-signing status --port 9000")
	fmt.Println()
}

// printRemoteSigningHelp prints general help for the remote-signing command
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
	fmt.Println("Examples:")
	fmt.Println("  harlequin remote-signing start --port 8080")
	fmt.Println("  harlequin remote-signing status")
	fmt.Println("  harlequin remote-signing help")
	fmt.Println()
	fmt.Println("For detailed command help, use:")
	fmt.Println("  harlequin remote-signing <command> --help")
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

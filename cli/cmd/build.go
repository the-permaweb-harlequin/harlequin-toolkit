package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui"
)

// RunInteractiveTUI launches the interactive TUI interface
func RunInteractiveTUI(ctx context.Context) error {
	return tui.RunBuildTUI(ctx)
}

// HandleBuildCommand handles the non-interactive build command with all its flags and modes
func HandleBuildCommand(ctx context.Context, args []string) {
	// Parse flags
	var debugMode bool
	var entrypoint string
	var outputDir string
	var configPath string

	// Process arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debugMode = true
		case "--help", "-h":
			PrintBuildUsage()
			return
		case "--entrypoint":
			if i+1 >= len(args) {
				fmt.Printf("Error: --entrypoint requires a value\n\n")
				PrintBuildUsage()
				os.Exit(1)
			}
			entrypoint = args[i+1]
			i++ // Skip the next argument as it's the value
		case "--outputDir":
			if i+1 >= len(args) {
				fmt.Printf("Error: --outputDir requires a value\n\n")
				PrintBuildUsage()
				os.Exit(1)
			}
			outputDir = args[i+1]
			i++ // Skip the next argument as it's the value
		case "--configPath":
			if i+1 >= len(args) {
				fmt.Printf("Error: --configPath requires a value\n\n")
				PrintBuildUsage()
				os.Exit(1)
			}
			configPath = args[i+1]
			i++ // Skip the next argument as it's the value
		default:
			// If it starts with -, it's an unknown flag
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("Unknown flag: %s\n\n", arg)
				PrintBuildUsage()
				os.Exit(1)
			} else {
				fmt.Printf("Unknown argument: %s\n\n", arg)
				PrintBuildUsage()
				os.Exit(1)
			}
		}
	}

	// Enable debug mode if flag was provided
	if debugMode {
		debug.SetEnabled(true)
	}

	// Require entrypoint for non-interactive build
	if entrypoint == "" {
		fmt.Println("Error: --entrypoint is required for non-interactive build")
		fmt.Println("Use 'harlequin' (without arguments) to launch the interactive TUI")
		fmt.Println()
		PrintBuildUsage()
		os.Exit(1)
	}

	// Non-interactive CLI mode with explicit flags
	handleNonInteractiveBuild(ctx, entrypoint, outputDir, configPath)
}

// handleNonInteractiveBuild handles the non-interactive CLI build mode
func handleNonInteractiveBuild(ctx context.Context, entrypoint, outputDir, configPath string) {
	// Load config
	var cfg *config.Config
	if configPath != "" {
		// Load from specified config path
		cfg = config.ReadConfigFile(configPath)
		if cfg == nil {
			fmt.Printf("Error: Failed to load config from %s\n", configPath)
			os.Exit(1)
		}
	} else {
		// Load default config
		cfg = loadConfig()
	}

	// Set output directory if provided
	if outputDir != "" {
		// TODO: Update config with output directory when config supports it
		debug.Printf("Output directory specified: %s", outputDir)
	}

	// Get current working directory as workspace
	workspaceDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Create build runner
	runner, err := build.NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		fmt.Printf("Failed to create build runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	// Build the project with specified entrypoint
	if err := runner.BuildProject(ctx, entrypoint); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}
}

// loadConfig loads configuration from various sources
func loadConfig() *config.Config {
	// Try to load config from file, fallback to defaults
	configPath := "harlequin.yaml"
	if _, err := os.Stat(configPath); err == nil {
		return config.ReadConfigFile(configPath)
	}

	// Try build_configs directory
	buildConfigPath := filepath.Join("build_configs", "ao-build-config.yml")
	if _, err := os.Stat(buildConfigPath); err == nil {
		return config.ReadConfigFile(buildConfigPath)
	}

	// Use defaults
	return config.NewConfig(nil)
}

// PrintBuildUsage prints the usage information for the build command
func PrintBuildUsage() {
	fmt.Println("🎭 Harlequin Build Command (Non-Interactive)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin build --entrypoint <file> [flags]")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  --entrypoint <file>    Path to the main Lua file to build")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  --outputDir <dir>      Directory to output build artifacts")
	fmt.Println("  --configPath <file>    Path to custom configuration file")
	fmt.Println("  -d, --debug            Enable debug logging for detailed output")
	fmt.Println("  -h, --help             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin build --entrypoint main.lua")
	fmt.Println("  harlequin build --entrypoint src/app.lua --outputDir dist")
	fmt.Println("  harlequin build --entrypoint main.lua --configPath custom.yaml")
	fmt.Println("  harlequin build --entrypoint main.lua --debug")
	fmt.Println("  harlequin build --entrypoint main.lua --outputDir dist --debug")
	fmt.Println()
	fmt.Println("Interactive Mode:")
	fmt.Println("  For interactive builds with guided configuration:")
	fmt.Println("  harlequin                          # Launch interactive TUI")
	fmt.Println()
	fmt.Println("Debug Mode:")
	fmt.Println("  When --debug is enabled, you'll see detailed logging including:")
	fmt.Println("  • Git repository cloning progress")
	fmt.Println("  • Docker build container output")
	fmt.Println("  • File copying and injection details")
	fmt.Println("  • Cleanup operations")
	fmt.Println()
	fmt.Println("Environment Variable Alternative:")
	fmt.Println("  HARLEQUIN_DEBUG=true harlequin build --entrypoint <file>")
}

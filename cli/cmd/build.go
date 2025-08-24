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
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui/components"
)

// HandleBuildCommand handles the build command with all its flags and modes
func HandleBuildCommand(ctx context.Context, args []string) {
	// Parse flags
	var debugMode bool
	var projectPath string

	// Process arguments
	remainingArgs := []string{}
	for i, arg := range args {
		switch arg {
		case "--debug", "-d":
			debugMode = true
		case "--help", "-h":
			if err := components.ShowHelp("build"); err != nil {
				// Fallback to basic help if something goes wrong
				fmt.Printf("Error displaying help: %v\n", err)
				fmt.Printf("Falling back to basic help...\n")
				PrintBuildUsage()
			}
			return
		default:
			// If it starts with -, it's an unknown flag
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("Unknown flag: %s\n\n", arg)
				PrintBuildUsage()
				os.Exit(1)
			}
			// Otherwise, it's a positional argument
			remainingArgs = append(remainingArgs, arg)
		}
		_ = i // unused variable fix
	}

	// Enable debug mode if flag was provided
	if debugMode {
		debug.SetEnabled(true)
	}

	// Determine if TUI or legacy CLI mode
	if len(remainingArgs) == 0 {
		// No path argument - launch TUI
		if err := tui.RunBuildTUI(ctx); err != nil {
			fmt.Printf("Build failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Legacy CLI mode with project path
		projectPath = remainingArgs[0]
		handleLegacyBuild(ctx, []string{projectPath})
	}
}

// handleLegacyBuild handles the legacy CLI build mode
func handleLegacyBuild(ctx context.Context, args []string) {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Load config
	cfg := loadConfig()

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

	// Build the project
	if err := runner.BuildProject(ctx, projectPath); err != nil {
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
	fmt.Println("🎭 Harlequin Build Command")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin build [flags] [path]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -d, --debug     Enable debug logging for detailed output")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  [path]          Project path (optional, defaults to interactive TUI)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin build                    # Launch interactive TUI")
	fmt.Println("  harlequin build --debug            # TUI with debug logging")
	fmt.Println("  harlequin build ./my-project       # Direct build (legacy mode)")
	fmt.Println("  harlequin build -d ./my-project    # Direct build with debug")
	fmt.Println()
	fmt.Println("Debug Mode:")
	fmt.Println("  When --debug is enabled, you'll see detailed logging including:")
	fmt.Println("  • Git repository cloning progress")
	fmt.Println("  • Docker build container output")
	fmt.Println("  • File copying and injection details")
	fmt.Println("  • Cleanup operations")
	fmt.Println()
	fmt.Println("Environment Variable Alternative:")
	fmt.Println("  HARLEQUIN_DEBUG=true harlequin build")
}

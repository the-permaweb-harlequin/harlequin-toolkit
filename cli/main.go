package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui"
)

func main() {
	// Ensure debug log file is closed on exit
	defer debug.Close()
	
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	switch command {
	case "build":
		handleBuildCommandWithFlags(ctx, os.Args[2:])
	case "status":
		handleStatusCommand(ctx)
	case "start":
		handleStartCommand(ctx)
	case "stop":
		handleStopCommand(ctx)
	case "exec":
		handleExecCommand(ctx, os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleBuildCommandWithFlags(ctx context.Context, args []string) {
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
			printBuildUsage()
			return
		default:
			// If it starts with -, it's an unknown flag
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("Unknown flag: %s\n\n", arg)
				printBuildUsage()
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
		handleBuildCommand(ctx, []string{projectPath})
	}
}

func handleBuildCommand(ctx context.Context, args []string) {
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

func handleStatusCommand(ctx context.Context) {
	cfg := loadConfig()
	workspaceDir, _ := os.Getwd()
	runner, err := build.NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		fmt.Printf("Failed to create build runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	status, err := runner.GetBuildStatus(ctx)
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Build Configuration:\n")
	fmt.Printf("  Image: %s\n", status.ImageName)
	fmt.Printf("  Container Name: %s\n", status.ContainerName)
	fmt.Printf("  Workspace: %s\n", status.WorkspaceDir)
	fmt.Printf("  Note: Using direct docker run commands (not persistent containers)\n")
}

func handleStartCommand(ctx context.Context) {
	cfg := loadConfig()
	workspaceDir, _ := os.Getwd()
	runner, err := build.NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		fmt.Printf("Failed to create build runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	fmt.Println("Start command is deprecated.")
	fmt.Println("Use 'build' command directly - it uses direct docker run commands instead of persistent containers.")
}

func handleStopCommand(ctx context.Context) {
	cfg := loadConfig()
	workspaceDir, _ := os.Getwd()
	runner, err := build.NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		fmt.Printf("Failed to create build runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	fmt.Println("Stop command is deprecated.")
	fmt.Println("No persistent containers to stop - direct docker run commands are used instead.")
}

func handleExecCommand(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Println("Error: exec command requires arguments")
		os.Exit(1)
	}

	cfg := loadConfig()
	workspaceDir, _ := os.Getwd()
	runner, err := build.NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		fmt.Printf("Failed to create build runner: %v\n", err)
		os.Exit(1)
	}
	defer runner.Close()

	fmt.Println("Exec command is deprecated.")
	fmt.Println("Use 'build' command to build projects, or run docker commands directly.")
}

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

func printUsage() {
	fmt.Println("🎭 Harlequin - Arweave Development Toolkit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build [flags] [path]    Build project (interactive TUI or legacy CLI)")
	fmt.Println("  status                  Show build environment status")
	fmt.Println("  help                    Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin build                    # Interactive TUI")
	fmt.Println("  harlequin build --debug            # Interactive TUI with debug logging")
	fmt.Println("  harlequin build ./my-project       # Legacy CLI mode")
	fmt.Println("  harlequin build --debug ./project  # Legacy CLI with debug logging")
	fmt.Println("  harlequin status")
	fmt.Println()
	fmt.Println("The interactive TUI provides a guided experience for:")
	fmt.Println("  • Selecting build type (AOS Flavour)")
	fmt.Println("  • Choosing entrypoint files") 
	fmt.Println("  • Configuring output directories")
	fmt.Println("  • Editing .harlequin.yaml configuration")
	fmt.Println()
	fmt.Println("For detailed build options, use: harlequin build --help")
}

func printBuildUsage() {
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

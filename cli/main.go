package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	switch command {
	case "build":
		handleBuildCommand(ctx, os.Args[2:])
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
	fmt.Println("Harlequin AO Build Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin-cli <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build [path]    Build project at path (default: current directory)")
	fmt.Println("  start           Start the build environment")
	fmt.Println("  stop            Stop the build environment")
	fmt.Println("  status          Show build environment status")
	fmt.Println("  exec <cmd>      Execute command in build environment")
	fmt.Println("  help            Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin-cli build")
	fmt.Println("  harlequin-cli build ./my-project")
	fmt.Println("  harlequin-cli start")
	fmt.Println("  harlequin-cli exec ls -la")
	fmt.Println("  harlequin-cli status")
	fmt.Println("  harlequin-cli stop")
}

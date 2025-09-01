package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/cmd"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
)

// Version information (injected by GoReleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Ensure debug log file is closed on exit
	defer debug.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// If no arguments provided, launch interactive TUI
	if len(os.Args) < 2 {
		if err := cmd.RunInteractiveTUI(ctx); err != nil {
			fmt.Printf("TUI failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		cmd.HandleInitCommand(ctx, os.Args[2:])
	case "build":
		cmd.HandleBuildCommand(ctx, os.Args[2:])
	case "upload-module":
		cmd.HandleUploadCommand(ctx, os.Args[2:])
	case "lua-utils":
		cmd.HandleLuaUtilsCommand(ctx, os.Args[2:])
	case "remote-signing":
		cmd.HandleRemoteSigningCommand(ctx, os.Args[2:])
	case "install":
		cmd.HandleInstallCommand(ctx, os.Args[2:])
	case "uninstall":
		cmd.HandleUninstallCommand(ctx, os.Args[2:])
	case "versions":
		cmd.HandleVersionsCommand(ctx, os.Args[2:])
	case "version", "--version", "-v":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("harlequin version %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built at: %s\n", date)
	fmt.Printf("  built by: %s\n", builtBy)
}

func printUsage() {
	fmt.Println("🎭 Harlequin - Arweave Development Toolkit")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    harlequin [COMMAND] [OPTIONS]")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("    init            Create a new AO process project from template")
	fmt.Println("    build           Build AO process (launches TUI if no args)")
	fmt.Println("    upload-module   Upload built modules to Arweave")
	fmt.Println("    lua-utils       Lua utilities for bundling and processing")
	fmt.Println("    remote-signing  Remote signing server operations")
	fmt.Println("    install         Install or upgrade harlequin")
	fmt.Println("    uninstall       Remove harlequin from system")
	fmt.Println("    versions        List available harlequin versions")
	fmt.Println("    version         Show version information")
	fmt.Println("    help            Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    harlequin                    # Launch interactive TUI")
	fmt.Println("    harlequin init               # Create new project (interactive)")
	fmt.Println("    harlequin init lua --name my-project")
	fmt.Println("    harlequin build --entrypoint main.lua")
	fmt.Println("    harlequin upload-module      # Upload built module to Arweave")
	fmt.Println("    harlequin versions --format table")
	fmt.Println()
	fmt.Println("For command-specific help, use: harlequin [COMMAND] --help")
}

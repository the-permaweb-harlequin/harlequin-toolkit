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
	case "build":
		cmd.HandleBuildCommand(ctx, os.Args[2:])
	case "lua-utils":
		cmd.HandleLuaUtilsCommand(ctx, os.Args[2:])
	case "remote-signing":
		cmd.HandleRemoteSigningCommand(ctx, os.Args[2:])
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
	fmt.Printf("Version: %s\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin                   Launch interactive TUI (default)")
	fmt.Println("  harlequin <command> [args]  Run specific command")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build --entrypoint <file> [flags]  Build project non-interactively")
	fmt.Println("  lua-utils <subcommand> [flags]     Lua utilities (bundle, etc.)")
	fmt.Println("  remote-signing <subcommand> [flags] Remote signing server")
	fmt.Println("  version                             Show version information")
	fmt.Println("  help                                Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin                                    # Launch interactive TUI")
	fmt.Println("  harlequin build --entrypoint main.lua       # Non-interactive build")
	fmt.Println("  harlequin build --entrypoint main.lua --debug  # Build with debug")
	fmt.Println("  harlequin lua-utils bundle --entrypoint main.lua  # Bundle Lua files")
	fmt.Println("  harlequin remote-signing start --port 8080  # Start remote signing server")
	fmt.Println("  harlequin version                            # Show version information")
	fmt.Println()
	fmt.Println("Interactive TUI (Default Mode):")
	fmt.Println("  The TUI provides a guided experience for:")
	fmt.Println("  • Selecting build type (AOS Flavour)")
	fmt.Println("  • Choosing entrypoint files")
	fmt.Println("  • Configuring output directories")
	fmt.Println("  • Editing .harlequin.yaml configuration")
	fmt.Println()
	fmt.Println("For detailed build options, use: harlequin build --help")
}

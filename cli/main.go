package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/cmd"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
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
		cmd.HandleBuildCommand(ctx, os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("🎭 Harlequin - Arweave Development Toolkit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build [flags] [path]    Build project (interactive TUI or legacy CLI)")
	fmt.Println("  help                    Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin build                    # Interactive TUI")
	fmt.Println("  harlequin build --debug            # Interactive TUI with debug logging")
	fmt.Println("  harlequin build ./my-project       # Legacy CLI mode")
	fmt.Println("  harlequin build --debug ./project  # Legacy CLI with debug logging")
	fmt.Println()
	fmt.Println("The interactive TUI provides a guided experience for:")
	fmt.Println("  • Selecting build type (AOS Flavour)")
	fmt.Println("  • Choosing entrypoint files") 
	fmt.Println("  • Configuring output directories")
	fmt.Println("  • Editing .harlequin.yaml configuration")
	fmt.Println()
	fmt.Println("For detailed build options, use: harlequin build --help")
}
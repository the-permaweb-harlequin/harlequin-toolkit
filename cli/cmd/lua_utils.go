package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	luautils "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/lua_utils"
)

// HandleLuaUtilsCommand handles the lua-utils command and its subcommands
func HandleLuaUtilsCommand(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Printf("Error: lua-utils requires a subcommand\n\n")
		PrintLuaUtilsUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	switch subcommand {
	case "bundle":
		HandleBundleCommand(ctx, args[1:])
	case "help", "--help", "-h":
		PrintLuaUtilsUsage()
	default:
		fmt.Printf("Unknown subcommand: %s\n\n", subcommand)
		PrintLuaUtilsUsage()
		os.Exit(1)
	}
}

// HandleBundleCommand handles the bundle subcommand
func HandleBundleCommand(ctx context.Context, args []string) {
	var debugMode bool
	var entrypoint string
	var outputPath string

	// Process arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--debug", "-d":
			debugMode = true
		case "--help", "-h":
			PrintBundleUsage()
			return
		case "--entrypoint":
			if i+1 >= len(args) {
				fmt.Printf("Error: --entrypoint requires a value\n\n")
				PrintBundleUsage()
				os.Exit(1)
			}
			entrypoint = args[i+1]
			i++ // Skip the next argument as it's the value
		case "--outputPath":
			if i+1 >= len(args) {
				fmt.Printf("Error: --outputPath requires a value\n\n")
				PrintBundleUsage()
				os.Exit(1)
			}
			outputPath = args[i+1]
			i++ // Skip the next argument as it's the value
		default:
			// If it starts with -, it's an unknown flag
			if strings.HasPrefix(arg, "-") {
				fmt.Printf("Unknown flag: %s\n\n", arg)
				PrintBundleUsage()
				os.Exit(1)
			} else {
				fmt.Printf("Unknown argument: %s\n\n", arg)
				PrintBundleUsage()
				os.Exit(1)
			}
		}
	}

	// Enable debug mode if flag was provided
	if debugMode {
		debug.SetEnabled(true)
	}

	// Require entrypoint
	if entrypoint == "" {
		fmt.Println("Error: --entrypoint is required")
		fmt.Println()
		PrintBundleUsage()
		os.Exit(1)
	}

	// Default output path if not provided
	if outputPath == "" {
		// Generate default output path: entrypoint without extension + .bundled.lua
		dir := filepath.Dir(entrypoint)
		base := filepath.Base(entrypoint)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		outputPath = filepath.Join(dir, name+".bundled.lua")
		debug.Printf("Using default output path: %s", outputPath)
	}

	// Perform the bundling
	if err := performBundle(entrypoint, outputPath); err != nil {
		fmt.Printf("Bundle failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully bundled %s to %s\n", entrypoint, outputPath)
}

// performBundle performs the actual bundling operation
func performBundle(entrypoint, outputPath string) error {
	// Check if entrypoint file exists
	if _, err := os.Stat(entrypoint); os.IsNotExist(err) {
		return fmt.Errorf("entrypoint file does not exist: %s", entrypoint)
	}

	debug.Printf("Starting bundle process for entrypoint: %s", entrypoint)

	// Convert to absolute path for more consistent handling
	absEntrypoint, err := filepath.Abs(entrypoint)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for entrypoint: %w", err)
	}

	// Perform the bundling using lua_utils
	bundledContent, err := luautils.Bundle(absEntrypoint)
	if err != nil {
		return fmt.Errorf("failed to bundle Lua files: %w", err)
	}

	debug.Printf("Bundle process completed, writing to output: %s", outputPath)

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the bundled content to the output file
	if err := os.WriteFile(outputPath, []byte(bundledContent), 0644); err != nil {
		return fmt.Errorf("failed to write bundled file: %w", err)
	}

	debug.Printf("Bundle written successfully to: %s", outputPath)
	return nil
}

// PrintLuaUtilsUsage prints the usage information for the lua-utils command
func PrintLuaUtilsUsage() {
	fmt.Println("🎭 Harlequin Lua Utils")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin lua-utils <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  bundle    Bundle Lua files into a single executable")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin lua-utils bundle --entrypoint main.lua")
	fmt.Println("  harlequin lua-utils bundle --entrypoint src/app.lua --outputPath dist/bundle.lua")
	fmt.Println("  harlequin lua-utils help")
	fmt.Println()
	fmt.Println("For detailed subcommand options, use:")
	fmt.Println("  harlequin lua-utils <subcommand> --help")
}

// PrintBundleUsage prints the usage information for the bundle subcommand
func PrintBundleUsage() {
	fmt.Println("🎭 Harlequin Lua Utils - Bundle")
	fmt.Println()
	fmt.Println("Bundle Lua files into a single executable by resolving require() statements")
	fmt.Println("and creating a self-contained Lua script.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  harlequin lua-utils bundle --entrypoint <file> [flags]")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  --entrypoint <file>    Path to the main Lua file to bundle")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  --outputPath <file>    Path to output the bundled file")
	fmt.Println("                         (default: <entrypoint>.bundled.lua)")
	fmt.Println("  -d, --debug            Enable debug logging for detailed output")
	fmt.Println("  -h, --help             Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  harlequin lua-utils bundle --entrypoint main.lua")
	fmt.Println("  harlequin lua-utils bundle --entrypoint src/app.lua --outputPath dist/bundle.lua")
	fmt.Println("  harlequin lua-utils bundle --entrypoint main.lua --debug")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  • Analyzes your main Lua file for require() statements")
	fmt.Println("  • Recursively resolves all dependencies")
	fmt.Println("  • Handles circular dependencies gracefully")
	fmt.Println("  • Creates a single bundled file with all modules included")
	fmt.Println("  • Preserves the original module structure and functionality")
	fmt.Println()
	fmt.Println("The bundled output includes:")
	fmt.Println("  • All required modules as local functions")
	fmt.Println("  • Package loading mappings for require() compatibility")
	fmt.Println("  • Your main file content at the end")
}


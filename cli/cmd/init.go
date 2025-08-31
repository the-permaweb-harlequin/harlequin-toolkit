package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui/components"
)

// Available template languages
var availableTemplates = []string{
	"lua",
	"c",
	"rust",
	"assemblyscript",
}

// Template metadata
type TemplateInfo struct {
	Name        string
	Description string
	Language    string
	BuildSystem string
	Features    []string
}

var templateInfoMap = map[string]TemplateInfo{
	"lua": {
		Name:        "Lua AO Process",
		Description: "AO Process with Lua and C trampoline integration",
		Language:    "Lua",
		BuildSystem: "CMake + LuaRocks",
		Features: []string{
			"C trampoline with embedded Lua interpreter",
			"LuaRocks package management",
			"WebAssembly compilation",
			"Modular architecture",
			"Comprehensive testing with Busted",
		},
	},
	"c": {
		Name:        "C AO Process",
		Description: "High-performance AO Process in C with Conan",
		Language:    "C",
		BuildSystem: "CMake + Conan",
		Features: []string{
			"Conan package management",
			"Google Test integration",
			"Emscripten WebAssembly compilation",
			"Memory-efficient implementation",
			"Docker build support",
		},
	},
	"rust": {
		Name:        "Rust AO Process",
		Description: "Safe and fast AO Process with Rust and wasm-pack",
		Language:    "Rust",
		BuildSystem: "Cargo + wasm-pack",
		Features: []string{
			"Thread-safe state management",
			"Serde JSON serialization",
			"wasm-bindgen WebAssembly bindings",
			"Comprehensive error handling",
			"Size-optimized builds",
		},
	},
	"assemblyscript": {
		Name:        "AssemblyScript AO Process",
		Description: "TypeScript-like AO Process compiled to WebAssembly",
		Language:    "AssemblyScript",
		BuildSystem: "AssemblyScript Compiler",
		Features: []string{
			"TypeScript-like syntax",
			"Custom JSON handling",
			"Memory-safe operations",
			"Node.js testing framework",
			"Size optimization",
		},
	},
}

// HandleInitCommand handles the init command for project initialization
func HandleInitCommand(ctx context.Context, args []string) {
	debug.Printf("Handling init command with args: %v", args)

	var projectName string
	var templateLang string
	var targetDir string
	var authorName string
	var githubUser string
	var interactive bool = true

	// If no arguments, run interactive mode
	if len(args) == 0 {
		err := runInteractiveInit(ctx, projectName, targetDir, authorName, githubUser)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if first argument is a language (non-interactive mode)
	if len(args) > 0 && isValidTemplate(args[0]) {
		templateLang = args[0]
		interactive = false
		args = args[1:] // Remove language from args for further parsing
	}

	// Parse command line arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name", "-n":
			if i+1 < len(args) {
				projectName = args[i+1]
				i++
			}
		case "--template", "-t":
			if i+1 < len(args) {
				templateLang = args[i+1]
				interactive = false // Set non-interactive when template is specified
				i++
			}
		case "--dir", "-d":
			if i+1 < len(args) {
				targetDir = args[i+1]
				i++
			}
		case "--author", "-a":
			if i+1 < len(args) {
				authorName = args[i+1]
				i++
			}
		case "--github", "-g":
			if i+1 < len(args) {
				githubUser = args[i+1]
				i++
			}
		case "--interactive":
			interactive = true
		case "--non-interactive", "--no-interactive":
			interactive = false
		case "--help", "-h":
			PrintInitUsage()
			return
		default:
			if !strings.HasPrefix(args[i], "-") && projectName == "" {
				projectName = args[i]
			}
		}
	}

	if interactive {
		// Launch interactive TUI for template selection
		err := runInteractiveInit(ctx, projectName, targetDir, authorName, githubUser)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Non-interactive mode
		if projectName == "" {
			fmt.Println("Error: Project name is required in non-interactive mode")
			PrintInitUsage()
			os.Exit(1)
		}

		if templateLang == "" {
			fmt.Println("Error: Template language is required in non-interactive mode")
			PrintInitUsage()
			os.Exit(1)
		}

		err := initializeProject(projectName, templateLang, targetDir, authorName, githubUser)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// runInteractiveInit runs the interactive TUI for project initialization
func runInteractiveInit(ctx context.Context, projectName, targetDir, authorName, githubUser string) error {
	wizard := components.NewInitWizardComponent()

	// Pre-populate fields if provided
	if projectName != "" {
		// If project name is provided, we can skip to template selection
		wizard.ProjectName = projectName
	}
	if targetDir != "" {
		wizard.TargetDir = targetDir
	}
	if authorName != "" {
		wizard.AuthorName = authorName
	}
	if githubUser != "" {
		wizard.GitHubUser = githubUser
	}

	// Set up completion callback
	var resultErr error
	wizard.OnComplete = func(pName, tLang, aName, gUser, tDir string) {
		resultErr = initializeProject(pName, tLang, tDir, aName, gUser)
	}

	// Run the TUI
	p := tea.NewProgram(wizard, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run interactive wizard: %w", err)
	}

	return resultErr
}

// initializeProject creates a new project from the specified template
func initializeProject(projectName, templateLang, targetDir, authorName, githubUser string) error {
	debug.Printf("Initializing project: %s with template: %s", projectName, templateLang)

	// Validate template language
	if !isValidTemplate(templateLang) {
		return fmt.Errorf("invalid template language: %s. Available: %s", templateLang, strings.Join(availableTemplates, ", "))
	}

	// Determine target directory
	if targetDir == "" {
		targetDir = projectName
	}

	// Check if target directory already exists
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory %s already exists", targetDir)
	}

	// Get template source directory
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	templateSrcDir := filepath.Join(filepath.Dir(execPath), "templates", templateLang)

	// Try relative path if absolute doesn't exist
	if _, err := os.Stat(templateSrcDir); os.IsNotExist(err) {
		// Try relative to current working directory
		wd, _ := os.Getwd()
		templateSrcDir = filepath.Join(wd, "templates", templateLang)

		if _, err := os.Stat(templateSrcDir); os.IsNotExist(err) {
			// Try relative to CLI directory
			templateSrcDir = filepath.Join("cli", "templates", templateLang)

			if _, err := os.Stat(templateSrcDir); os.IsNotExist(err) {
				return fmt.Errorf("template directory not found for language: %s", templateLang)
			}
		}
	}

	fmt.Printf("Creating project '%s' from %s template...\n", projectName, templateLang)

	// Create target directory
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Copy template files
	err = copyTemplate(templateSrcDir, targetDir, projectName, authorName, githubUser)
	if err != nil {
		// Clean up on error
		os.RemoveAll(targetDir)
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Display success message and next steps
	printSuccessMessage(projectName, templateLang, targetDir)

	return nil
}

// copyTemplate copies template files and performs variable substitution
func copyTemplate(srcDir, destDir, projectName, authorName, githubUser string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		// Substitute variables in file names
		destPath = substituteVariables(destPath, projectName, authorName, githubUser)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy and process file
		return copyAndProcessFile(path, destPath, projectName, authorName, githubUser)
	})
}

// copyAndProcessFile copies a file and performs variable substitution
func copyAndProcessFile(srcPath, destPath, projectName, authorName, githubUser string) error {
	// Read source file
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	// Perform variable substitution
	processedContent := substituteVariables(string(content), projectName, authorName, githubUser)

	// Write to destination
	return os.WriteFile(destPath, []byte(processedContent), 0644)
}

// substituteVariables replaces template variables with actual values
func substituteVariables(content, projectName, authorName, githubUser string) string {
	replacements := map[string]string{
		"{{PROJECT_NAME}}": projectName,
		"{{AUTHOR_NAME}}":  authorName,
		"{{GITHUB_USER}}":  githubUser,
	}

	// Set defaults for empty values
	if replacements["{{AUTHOR_NAME}}"] == "" {
		replacements["{{AUTHOR_NAME}}"] = "Your Name"
	}
	if replacements["{{GITHUB_USER}}"] == "" {
		replacements["{{GITHUB_USER}}"] = "your-username"
	}

	result := content
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// isValidTemplate checks if the template language is valid
func isValidTemplate(templateLang string) bool {
	for _, template := range availableTemplates {
		if template == templateLang {
			return true
		}
	}
	return false
}

// printSuccessMessage displays success message and next steps
func printSuccessMessage(projectName, templateLang, targetDir string) {
	info := templateInfoMap[templateLang]

	fmt.Printf("\n🎉 Successfully created %s project '%s'!\n\n", info.Language, projectName)

	fmt.Println("📁 Project structure:")
	fmt.Printf("   %s/\n", targetDir)

	fmt.Println("\n🚀 Next steps:")
	fmt.Printf("   cd %s\n", targetDir)

	switch templateLang {
	case "lua":
		fmt.Println("   npm run setup          # Install LuaRocks dependencies")
		fmt.Println("   npm run build          # Build with CMake")
		fmt.Println("   npm run build:wasm     # Build WebAssembly with Emscripten")
		fmt.Println("   npm test               # Run Lua tests")
		fmt.Println("   npm run test:trampoline # Test C trampoline")
	case "c":
		fmt.Println("   npm run setup          # Install Conan dependencies")
		fmt.Println("   npm run build:cmake    # Build with CMake")
		fmt.Println("   npm test               # Run tests")
		fmt.Println("   npm run docker:build   # Build in Docker")
	case "rust":
		fmt.Println("   npm run setup          # Install Rust toolchain")
		fmt.Println("   npm run build          # Build native binary")
		fmt.Println("   npm run build:wasm     # Build WebAssembly")
		fmt.Println("   npm test               # Run tests")
		fmt.Println("   npm run run            # Test locally")
	case "assemblyscript":
		fmt.Println("   npm install            # Install dependencies")
		fmt.Println("   npm run build          # Build WebAssembly")
		fmt.Println("   npm test               # Run tests")
		fmt.Println("   npm run optimize       # Optimize binary")
	}

	fmt.Println("\n   harlequin build         # Build with Harlequin CLI")
	fmt.Println("   harlequin               # Launch interactive TUI")

	fmt.Printf("\n📖 Features included:\n")
	for _, feature := range info.Features {
		fmt.Printf("   • %s\n", feature)
	}

	fmt.Printf("\n📚 Documentation: See README.md in the project directory\n")
	fmt.Printf("🛠️  Build System: %s\n", info.BuildSystem)
	fmt.Printf("🎭 Happy coding with Harlequin!\n\n")
}

// PrintInitUsage prints usage information for the init command
func PrintInitUsage() {
	fmt.Println("🎭 Harlequin Init - Create New AO Process Projects")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    harlequin init                    # Interactive mode")
	fmt.Println("    harlequin init <LANGUAGE> [OPTIONS]  # Non-interactive mode")
	fmt.Println()
	fmt.Println("ARGUMENTS:")
	fmt.Println("    LANGUAGE        Template language (lua, c, rust, assemblyscript)")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("    -n, --name <NAME>           Project name (required in non-interactive mode)")
	fmt.Println("    -t, --template <TEMPLATE>   Template language (alternative to positional argument)")
	fmt.Println("    -d, --dir <DIRECTORY>       Target directory (default: project name)")
	fmt.Println("    -a, --author <AUTHOR>       Author name")
	fmt.Println("    -g, --github <USERNAME>     GitHub username")
	fmt.Println("    --interactive               Force interactive mode")
	fmt.Println("    --non-interactive           Skip interactive prompts")
	fmt.Println("    -h, --help                  Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    # Interactive mode")
	fmt.Println("    harlequin init")
	fmt.Println()
	fmt.Println("    # Non-interactive mode")
	fmt.Println("    harlequin init lua --name my-ao-process --author \"John Doe\"")
	fmt.Println("    harlequin init rust --name my-rust-process --github johndoe")
	fmt.Println("    harlequin init c --name my-c-project --dir ./projects/my-c-project")
	fmt.Println()
	fmt.Println("    # Alternative syntax (backward compatibility)")
	fmt.Println("    harlequin init --template lua --name my-project")
	fmt.Println()
	fmt.Println("AVAILABLE TEMPLATES:")
	for _, template := range availableTemplates {
		info := templateInfoMap[template]
		fmt.Printf("    %-15s %s\n", template, info.Description)
	}
	fmt.Println()
}

package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tui/components"
)

// Available template languages
var availableTemplates = []string{
	"assemblyscript",
	"go",
}

// Template metadata
type TemplateInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Language     string   `json:"language"`
	BuildSystem  string   `json:"buildSystem,omitempty"`
	Features     []string `json:"features,omitempty"`
	Instructions []string `json:"instructions,omitempty"`
	Version      string   `json:"version,omitempty"`
	Created      string   `json:"created,omitempty"`
	Tarball      string   `json:"tarball,omitempty"`
	Size         int      `json:"size,omitempty"`
}

var templateInfoMap = map[string]TemplateInfo{
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
	"go": {
		Name:        "Go AO Process",
		Description: "High-performance AO Process in Go compiled to WebAssembly",
		Language:    "Go",
		BuildSystem: "Go + TinyGo",
		Features: []string{
			"Type-safe Go programming",
			"Goroutines and channels",
			"Standard library support",
			"TinyGo WebAssembly compilation",
			"Efficient memory usage",
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
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		// First argument looks like a template name
		if isValidTemplate(args[0]) {
			templateLang = args[0]
			interactive = false
			args = args[1:] // Remove language from args for further parsing
		} else {
			// Invalid template name provided
			fmt.Printf("Error: Invalid template '%s'. Available templates: %s\n", args[0], strings.Join(availableTemplates, ", "))
			PrintInitUsage()
			os.Exit(1)
		}
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

		err := InitializeProject(projectName, templateLang, targetDir, authorName, githubUser)
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
		resultErr = InitializeProject(pName, tLang, tDir, aName, gUser)
	}

	// Run the TUI
	p := tea.NewProgram(wizard, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run interactive wizard: %w", err)
	}

	return resultErr
}

// InitializeProject creates a new project from the specified template
func InitializeProject(projectName, templateLang, targetDir, authorName, githubUser string) error {
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

	// Load embedded templates
	registry, err := LoadEmbeddedTemplates()
	if err != nil {
		return fmt.Errorf("failed to load embedded templates: %w", err)
	}

	template, exists := registry.Templates[templateLang]
	if !exists {
		return fmt.Errorf("template not found: %s. Available templates: %s", templateLang, strings.Join(availableTemplates, ", "))
	}

	fmt.Printf("Creating project '%s' from %s template...\n", projectName, template.Name)

	// Create target directory
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract embedded template
	err = extractEmbeddedTemplate(templateLang, targetDir, projectName, authorName, githubUser)
	if err != nil {
		// Clean up on error
		os.RemoveAll(targetDir)
		return fmt.Errorf("failed to extract embedded template: %w", err)
	}

	// Display success message with embedded template info
	printEmbeddedSuccessMessage(projectName, template, targetDir)

	return nil
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



// PrintInitUsage prints usage information for the init command
func PrintInitUsage() {
	fmt.Println("🎭 Harlequin Init - Create New AO Process Projects")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    harlequin init                    # Interactive mode")
	fmt.Println("    harlequin init <LANGUAGE> [OPTIONS]  # Non-interactive mode")
	fmt.Println()
	fmt.Println("ARGUMENTS:")
	fmt.Println("    LANGUAGE        Template language (assemblyscript, go)")
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
	fmt.Println("    harlequin init assemblyscript --name my-ao-process --author \"John Doe\"")
	fmt.Println("    harlequin init go --name my-go-process --github johndoe")
	fmt.Println("    harlequin init assemblyscript --name my-as-project --dir ./projects/my-as-project")
	fmt.Println()
	fmt.Println("    # Alternative syntax (backward compatibility)")
	fmt.Println("    harlequin init --template assemblyscript --name my-project")
	fmt.Println()
	fmt.Println("AVAILABLE TEMPLATES:")
	for _, template := range availableTemplates {
		info := templateInfoMap[template]
		fmt.Printf("    %-15s %s\n", template, info.Description)
	}
	fmt.Println()
}

// extractEmbeddedTemplate extracts the embedded template tarball to the target directory
func extractEmbeddedTemplate(templateLang, targetDir, projectName, authorName, githubUser string) error {
	// Read the embedded tarball
	tarballPath := fmt.Sprintf("embedded_templates/%s.tar.gz", templateLang)
	tarballData, err := embeddedTemplates.ReadFile(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to read embedded template: %w", err)
	}

	// Create gzip reader
	gzipReader, err := gzip.NewReader(bytes.NewReader(tarballData))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip special files
		if strings.HasPrefix(header.Name, ".harlequin-template.json") ||
		   strings.HasPrefix(header.Name, "install.sh") {
			continue
		}

		// Create target path
		targetPath := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			// Create parent directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			// Copy file content and process template variables
			content, err := io.ReadAll(tarReader)
			if err != nil {
				file.Close()
				return fmt.Errorf("failed to read file content: %w", err)
			}

			// Replace template variables
			processedContent := substituteVariables(string(content), projectName, authorName, githubUser)

			if _, err := file.WriteString(processedContent); err != nil {
				file.Close()
				return fmt.Errorf("failed to write file content: %w", err)
			}
			file.Close()
		}
	}

	return nil
}

// printEmbeddedSuccessMessage displays success message for embedded templates
func printEmbeddedSuccessMessage(projectName string, template TemplateInfo, targetDir string) {
	fmt.Printf("\n🎉 Successfully created '%s' from %s template!\n\n", projectName, template.Name)
	fmt.Printf("📁 Project created in: %s\n\n", targetDir)

	fmt.Printf("🚀 Next steps:\n")
	fmt.Printf("   cd %s\n", targetDir)

	for _, instruction := range template.Instructions {
		fmt.Printf("   %s\n", instruction)
	}

	fmt.Printf("\n📖 Features included:\n")
	fmt.Printf("   • %s\n", template.Description)

	fmt.Printf("\n📚 Documentation: See README.md in the project directory\n")
	fmt.Printf("🎭 Happy coding with Harlequin!\n\n")
}

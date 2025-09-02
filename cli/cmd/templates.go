package cmd

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed embedded_templates/*
var embeddedTemplates embed.FS

// TemplateRegistry represents the templates.json structure
type TemplateRegistry struct {
	Version   string                  `json:"version"`
	Generated string                  `json:"generated"`
	Templates map[string]TemplateInfo `json:"templates"`
}

// LoadEmbeddedTemplates loads the templates registry from embedded files
func LoadEmbeddedTemplates() (*TemplateRegistry, error) {
	data, err := embeddedTemplates.ReadFile("embedded_templates/templates.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	var registry TemplateRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse templates registry: %w", err)
	}

	return &registry, nil
}

// HandleTemplatesCommand handles the templates command
func HandleTemplatesCommand(ctx context.Context, args []string) {
	if len(args) == 0 {
		printTemplatesUsage()
		return
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		handleTemplatesList()
	case "info":
		if len(args) < 2 {
			fmt.Println("Error: template name required for info command")
			fmt.Println("Usage: harlequin templates info <template-name>")
			os.Exit(1)
		}
		handleTemplateInfo(args[1])
	case "help", "--help", "-h":
		printTemplatesUsage()
	default:
		fmt.Printf("Unknown templates subcommand: %s\n\n", subcommand)
		printTemplatesUsage()
		os.Exit(1)
	}
}

// handleTemplatesList lists all available templates
func handleTemplatesList() {
	registry, err := LoadEmbeddedTemplates()
	if err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🎭 Available AO Process Templates")
	fmt.Println()

	for _, template := range registry.Templates {
		fmt.Printf("  %s - %s\n", template.Language, template.Name)
		fmt.Printf("    %s\n", template.Description)
		fmt.Println()
	}

	fmt.Printf("📦 Total templates: %d\n", len(registry.Templates))
	fmt.Printf("📋 Registry version: %s\n", registry.Version)
	fmt.Println()
	fmt.Println("Use 'harlequin templates info <template>' for detailed information")
}

// handleTemplateInfo shows detailed info for a specific template
func handleTemplateInfo(templateName string) {
	registry, err := LoadEmbeddedTemplates()
	if err != nil {
		fmt.Printf("Error loading templates: %v\n", err)
		os.Exit(1)
	}

	template, exists := registry.Templates[templateName]
	if !exists {
		fmt.Printf("Template '%s' not found\n", templateName)
		fmt.Println("\nAvailable templates:")
		for lang := range registry.Templates {
			fmt.Printf("  - %s\n", lang)
		}
		os.Exit(1)
	}

	fmt.Printf("🎭 %s Template\n", template.Name)
	fmt.Println()
	fmt.Printf("Language: %s\n", template.Language)
	fmt.Printf("Description: %s\n", template.Description)
	fmt.Printf("Version: %s\n", template.Version)
	fmt.Printf("Size: %d bytes\n", template.Size)
	fmt.Println()

	if len(template.Instructions) > 0 {
		fmt.Println("📋 Getting Started:")
		for i, instruction := range template.Instructions {
			fmt.Printf("  %d. %s\n", i+1, instruction)
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Printf("  harlequin init --template %s --name my-project\n", template.Language)
	fmt.Printf("  harlequin init %s my-project\n", template.Language)
}

// printTemplatesUsage prints usage for the templates command
func printTemplatesUsage() {
	fmt.Println("🎭 Harlequin Templates - Manage AO Process Templates")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    harlequin templates [SUBCOMMAND] [OPTIONS]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("    list                List all available templates")
	fmt.Println("    info <template>     Show detailed information about a template")
	fmt.Println("    help               Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    harlequin templates list")
	fmt.Println("    harlequin templates info assemblyscript")
	fmt.Println("    harlequin templates info go")
}

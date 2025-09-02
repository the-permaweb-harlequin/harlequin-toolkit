package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ProjectConfig holds all the information needed to generate a project
type ProjectConfig struct {
	ProjectName string
	Language    string
	AuthorName  string
	GitHubUser  string
	OutputDir   string
}

// GenerateProject creates a new project from the specified template
func GenerateProject(config *ProjectConfig) error {
	if config.OutputDir == "" {
		config.OutputDir = "."
	}

	// Validate required fields
	if config.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	if config.Language == "" {
		return fmt.Errorf("language is required")
	}

	// Create project directory
	projectDir := filepath.Join(config.OutputDir, config.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate template based on language
	switch strings.ToLower(config.Language) {
	case "lua":
		return generateLuaTemplate(projectDir, config)
	case "rust":
		return generateRustTemplate(projectDir, config)
	case "c":
		return generateCTemplate(projectDir, config)
	case "assemblyscript":
		return generateAssemblyScriptTemplate(projectDir, config)
	default:
		return fmt.Errorf("unsupported language: %s", config.Language)
	}
}

// generateLuaTemplate creates a Lua project template
func generateLuaTemplate(projectDir string, config *ProjectConfig) error {
	// Create directory structure
	dirs := []string{
		"wasm/c",
		"test",
		"docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Template files to create
	files := map[string]string{
		"README.md": luaReadmeTemplate,
		"process.lua": luaProcessTemplate,
		"handlers.lua": luaHandlersTemplate,
		"package.json": luaPackageJsonTemplate,
		"wasm/c/process.c": cProcessTemplate,
		"wasm/c/handlers.c": cHandlersTemplate,
		"wasm/c/Makefile": cMakefileTemplate,
		".gitignore": gitignoreTemplate,
	}

	return createFiles(projectDir, files, config)
}

// generateRustTemplate creates a Rust project template
func generateRustTemplate(projectDir string, config *ProjectConfig) error {
	// Create directory structure
	dirs := []string{
		"src",
		"test",
		"docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Template files to create
	files := map[string]string{
		"README.md": rustReadmeTemplate,
		"Cargo.toml": rustCargoTomlTemplate,
		"src/lib.rs": rustLibTemplate,
		"src/handlers.rs": rustHandlersTemplate,
		"wasm-pack.json": rustWasmPackTemplate,
		".gitignore": gitignoreTemplate,
	}

	return createFiles(projectDir, files, config)
}

// generateCTemplate creates a C project template
func generateCTemplate(projectDir string, config *ProjectConfig) error {
	// Create directory structure
	dirs := []string{
		"src",
		"include",
		"test",
		"docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Template files to create
	files := map[string]string{
		"README.md": cReadmeTemplate,
		"CMakeLists.txt": cCMakeListsTemplate,
		"conanfile.txt": cConanFileTemplate,
		"src/process.c": cSrcProcessTemplate,
		"src/handlers.c": cSrcHandlersTemplate,
		"include/process.h": cProcessHeaderTemplate,
		"include/handlers.h": cHandlersHeaderTemplate,
		".gitignore": gitignoreTemplate,
	}

	return createFiles(projectDir, files, config)
}

// generateAssemblyScriptTemplate creates an AssemblyScript project template
func generateAssemblyScriptTemplate(projectDir string, config *ProjectConfig) error {
	// Create directory structure
	dirs := []string{
		"assembly",
		"test",
		"docs",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Template files to create
	files := map[string]string{
		"README.md": asReadmeTemplate,
		"package.json": asPackageJsonTemplate,
		"asconfig.json": asConfigTemplate,
		"assembly/index.ts": asIndexTemplate,
		"assembly/handlers.ts": asHandlersTemplate,
		"assembly/tsconfig.json": asTsConfigTemplate,
		".gitignore": gitignoreTemplate,
	}

	return createFiles(projectDir, files, config)
}

// createFiles generates files from templates
func createFiles(projectDir string, files map[string]string, config *ProjectConfig) error {
	for fileName, templateContent := range files {
		filePath := filepath.Join(projectDir, fileName)

		// Create directory for file if needed
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", fileName, err)
		}

		// Parse and execute template
		tmpl, err := template.New(fileName).Parse(templateContent)
		if err != nil {
			return fmt.Errorf("failed to parse template for %s: %w", fileName, err)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", fileName, err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, config); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", fileName, err)
		}
	}

	return nil
}

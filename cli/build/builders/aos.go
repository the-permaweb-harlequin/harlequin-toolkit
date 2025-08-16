package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	harlequinBuild "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
	harlequinLuaUtils "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/lua_utils"
)

const (
	AOSRepoURL = "https://github.com/permaweb/aos.git"
)

/*
AOS builder is a builder for the vanilla AOS module
It uses the AO build container to build the project
Requires a config file for the build container (stack memory, target, etc)
Requires the AOS git hash to clone the repo
Requires the project path to bundle the lua code

Steps:
1. Clone the AOS repo and clean the imports
	✅ added CopyAOSProcess functionality
2. Create the build directory
	✅ implemented in BuildProjectWithInjection
3. Bundle the lua code and write the bundle to the build directory 
	✅ added luautils to bundle the lua code
4. Inject the bundle into the AOS code
	✅ added injection functionality
5. Call the container to build the project
	✅ implemented using direct Docker command with ao-build-module
6. Write the process.wasm and the bundled lua code to the output directory
	✅ added CopyBuildOutputs functionality
7. Clean up the build container and build directory
	✅ added CleanAOSWorkspace functionality

Usage:
	builder := NewAOSBuilder(config, workspaceDir)
	
	// Prepare workspace with AOS files
	err := builder.PrepareWorkspace(ctx, workspaceDir)
	
	// Build project with bundling and injection (recommended)
	err = builder.BuildProjectWithInjection(ctx, projectPath, outputDir)
	
	// Or build basic project without bundling
	err = builder.Build(ctx, projectPath)
	
	// Clean up
	err = builder.CleanWorkspace(workspaceDir)
*/

type AOSBuilder struct {
	entrypoint   string
	outputDir    string
	workspaceDir string
	config       *harlequinConfig.Config
	runner       *harlequinBuild.BuildRunner
}

func NewAOSBuilder(config *harlequinConfig.Config, workspaceDir string) *AOSBuilder {
	runner, err := harlequinBuild.NewAOBuildRunner(config, workspaceDir)
	if err != nil {
		panic(err)
	}
	return &AOSBuilder{
		config:       config,
		workspaceDir: workspaceDir,
		runner:       runner,
	}
}

func (b *AOSBuilder) Build(ctx context.Context, projectPath string) error {
	return b.runner.BuildProject(ctx, projectPath)
}

// PrepareWorkspace prepares the AOS workspace by copying the AOS process files
func (b *AOSBuilder) PrepareWorkspace(ctx context.Context, workspaceDir string) error {
	return PrepareAOSWorkspace(ctx, b.config, workspaceDir)
}

// CopyAOSFiles copies AOS process files to the specified target directory
func (b *AOSBuilder) CopyAOSFiles(ctx context.Context, targetDir, configSourceFile string) error {
	return CopyAOSProcessWithConfig(ctx, b.config, targetDir, configSourceFile)
}

// CleanWorkspace removes AOS-related files from the workspace
func (b *AOSBuilder) CleanWorkspace(workspaceDir string) error {
	return CleanAOSWorkspace(workspaceDir)
}

// GetCopyOptions returns configured AOSCopyOptions for this builder
func (b *AOSBuilder) GetCopyOptions(targetDir string) *AOSCopyOptions {
	return NewAOSCopyOptions(b.config, targetDir)
}

// AOSCopyOptions holds configuration for copying AOS process files
type AOSCopyOptions struct {
	RepoURL           string
	CommitHash        string
	TempRepoDir       string
	ProcessTargetDir  string
	ConfigSourceFile  string
	ConfigDestFile    string
}

// NewAOSCopyOptions creates default options for copying AOS process files
func NewAOSCopyOptions(config *harlequinConfig.Config, targetDir string) *AOSCopyOptions {
	tempRepoDir := filepath.Join(os.TempDir(), "harlequin-aos-repo")
	processTargetDir := filepath.Join(targetDir, "aos-process")
	configDestFile := filepath.Join(processTargetDir, "config.yml")

	return &AOSCopyOptions{
		RepoURL:           AOSRepoURL,
		CommitHash:        config.AOSGitHash,
		TempRepoDir:       tempRepoDir,
		ProcessTargetDir:  processTargetDir,
		ConfigSourceFile:  "", // Will be set based on where config is found
		ConfigDestFile:    configDestFile,
	}
}

// CopyAOSProcess clones the AOS repository and copies the process directory
func CopyAOSProcess(ctx context.Context, options *AOSCopyOptions) error {
	fmt.Printf("Starting AOS process copy...\n")

	// Step 1: Remove existing aos-process directory
	fmt.Printf("Removing existing directory: %s\n", options.ProcessTargetDir)
	if err := os.RemoveAll(options.ProcessTargetDir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	// Step 2: Clone the repository into a temporary directory
	fmt.Printf("Cloning repository: %s\n", options.RepoURL)
	cloneCmd := exec.CommandContext(ctx, "git", "clone", options.RepoURL, options.TempRepoDir)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Cleanup temp repo on exit
	defer func() {
		fmt.Printf("Removing temporary directory: %s\n", options.TempRepoDir)
		_ = os.RemoveAll(options.TempRepoDir)
	}()

	// Step 3: Checkout the specific commit hash
	fmt.Printf("Checking out commit: %s\n", options.CommitHash)
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", options.CommitHash)
	checkoutCmd.Dir = options.TempRepoDir
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout commit %s: %w", options.CommitHash, err)
	}

	// Step 4: Move the process directory to the target location
	processDir := filepath.Join(options.TempRepoDir, "process")
	fmt.Printf("Moving %s to %s\n", processDir, options.ProcessTargetDir)
	
	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(options.ProcessTargetDir), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	if err := copyDirectory(processDir, options.ProcessTargetDir); err != nil {
		return fmt.Errorf("failed to copy process directory: %w", err)
	}

	// Step 5: Copy the build config file to the target directory (if specified)
	if options.ConfigSourceFile != "" {
		fmt.Printf("Copying %s to %s\n", options.ConfigSourceFile, options.ConfigDestFile)
		if err := copyFile(options.ConfigSourceFile, options.ConfigDestFile); err != nil {
			return fmt.Errorf("failed to copy config file: %w", err)
		}
	}

	fmt.Println("Successfully copied AOS process and config.")
	return nil
}

// CopyAOSProcessWithConfig is a convenience function that copies AOS process files using a Harlequin config
func CopyAOSProcessWithConfig(ctx context.Context, config *harlequinConfig.Config, targetDir, configSourceFile string) error {
	options := NewAOSCopyOptions(config, targetDir)
	options.ConfigSourceFile = configSourceFile
	return CopyAOSProcess(ctx, options)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write destination file
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// copyDirectory recursively copies a directory from src to dst
func copyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy subdirectory %s: %w", entry.Name(), err)
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// PrepareAOSWorkspace prepares a workspace for AOS building by copying necessary files
func PrepareAOSWorkspace(ctx context.Context, config *harlequinConfig.Config, workspaceDir string) error {
	// Find config file
	configPath := findConfigFile()
	if configPath == "" {
		return fmt.Errorf("could not find config file")
	}

	return CopyAOSProcessWithConfig(ctx, config, workspaceDir, configPath)
}

// findConfigFile looks for config files in common locations
func findConfigFile() string {
	possiblePaths := []string{
		"ao-build-config.yml",
		"build_configs/ao-build-config.yml",
		"config/ao-build-config.yml",
		"harlequin.yaml",
		"harlequin.yml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// CleanAOSWorkspace removes AOS-related files from the workspace
func CleanAOSWorkspace(workspaceDir string) error {
	aosProcessDir := filepath.Join(workspaceDir, "aos-process")
	fmt.Printf("Cleaning AOS workspace: %s\n", aosProcessDir)
	return os.RemoveAll(aosProcessDir)
}

// BuildInjectionOptions configures the build injection process
type BuildInjectionOptions struct {
	ProcessFilePath string
	BundledCodePath string
	RequireName     string // The name to use in require() statement
}

// NewDefaultBuildInjectionOptions creates default injection options
func NewDefaultBuildInjectionOptions(processDir, bundledCodePath, requireName string) *BuildInjectionOptions {
	return &BuildInjectionOptions{
		ProcessFilePath: filepath.Join(processDir, "process.lua"),
		BundledCodePath: bundledCodePath,
		RequireName:     requireName,
	}
}

// InjectBundledCode injects bundled Lua code into the AOS process file
func InjectBundledCode(options *BuildInjectionOptions) error {
	fmt.Printf("Injecting bundled code into: %s\n", options.ProcessFilePath)

	// Read the process.lua file
	content, err := os.ReadFile(options.ProcessFilePath)
	if err != nil {
		return fmt.Errorf("failed to read process file: %w", err)
	}

	fileContent := string(content)

	// Inject the bundled code require statement
	fileContent, err = injectRequireStatement(fileContent, options.RequireName)
	if err != nil {
		return fmt.Errorf("failed to inject require statement: %w", err)
	}

	// Write the updated content back to the file
	if err := os.WriteFile(options.ProcessFilePath, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated process file: %w", err)
	}

	fmt.Printf("Successfully injected bundled code require: require('%s')\n", options.RequireName)
	return nil
}

// injectRequireStatement injects the bundled code require after the last Handlers.append
func injectRequireStatement(content, requireName string) (string, error) {
	// Check if the require statement is already present
	existingRequire := fmt.Sprintf("require('%s')", requireName)
	if strings.Contains(content, existingRequire) {
		fmt.Printf("The require('%s') statement is already present\n", requireName)
		return content, nil
	}

	// Find all Handlers.append statements
	handlersAppendRegex := regexp.MustCompile(`(Handlers\.append.*)`)
	
	// Find the last occurrence of Handlers.append
	lastMatchIndices := handlersAppendRegex.FindAllStringIndex(content, -1)
	if len(lastMatchIndices) == 0 {
		return "", fmt.Errorf("no Handlers.append found in process.lua")
	}

	lastMatch := lastMatchIndices[len(lastMatchIndices)-1]
	position := lastMatch[1] // End of the last match

	// Inject the require statement after the last Handlers.append
	injectionCode := fmt.Sprintf("\nrequire('%s');", requireName)
	result := content[:position] + injectionCode + content[position:]

	fmt.Printf("Injected require('%s') after the last Handlers.append\n", requireName)
	return result, nil
}

// InjectBundledCodeIntoProcess is a convenience method for AOSBuilder
func (b *AOSBuilder) InjectBundledCodeIntoProcess(ctx context.Context, bundledCodePath, requireName string) error {
	processDir := filepath.Join(b.workspaceDir, "aos-process")
	options := NewDefaultBuildInjectionOptions(processDir, bundledCodePath, requireName)
	return InjectBundledCode(options)
}

// InjectBundledCodeWithOptions allows custom injection options
func (b *AOSBuilder) InjectBundledCodeWithOptions(ctx context.Context, options *BuildInjectionOptions) error {
	return InjectBundledCode(options)
}

// BuildProjectWithInjection builds a project with bundled code injection
func (b *AOSBuilder) BuildProjectWithInjection(ctx context.Context, projectPath, outputDir string) error {
	fmt.Printf("Starting AOS build with injection for project: %s\n", projectPath)
	
	// Step 1: Bundle the Lua project using luautils
	fmt.Println("Step 1: Bundling Lua project...")
	entryPath := filepath.Join(projectPath, "main.lua")
	bundledCode, err := harlequinLuaUtils.Bundle(entryPath)
	if err != nil {
		return fmt.Errorf("failed to bundle Lua project: %w", err)
	}
	
	// Step 2: Write bundled code to workspace
	fmt.Println("Step 2: Writing bundled code to workspace...")
	processDir := filepath.Join(b.workspaceDir, "aos-process")
	bundledFilePath := filepath.Join(processDir, "bundled.lua")
	if err := os.WriteFile(bundledFilePath, []byte(bundledCode), 0644); err != nil {
		return fmt.Errorf("failed to write bundled code: %w", err)
	}
	fmt.Printf("Bundled code written to: %s\n", bundledFilePath)
	
	// Step 3: Inject the bundled code into the AOS process
	fmt.Println("Step 3: Injecting bundled code into AOS process...")
	options := NewDefaultBuildInjectionOptions(processDir, bundledFilePath, ".bundled")
	if err := InjectBundledCode(options); err != nil {
		return fmt.Errorf("failed to inject bundled code: %w", err)
	}
	
	// Step 4: Build the project using the container
	fmt.Println("Step 4: Building project with container...")
	if err := b.buildWithDocker(ctx, processDir); err != nil {
		return fmt.Errorf("failed to build project: %w", err)
	}
	
	// Step 5: Copy outputs to the specified directory
	fmt.Println("Step 5: Copying build outputs...")
	if err := b.CopyBuildOutputs(processDir, outputDir); err != nil {
		return fmt.Errorf("failed to copy build outputs: %w", err)
	}
	
	// Also copy the bundled Lua file to output directory
	outputBundledPath := filepath.Join(outputDir, "bundled.lua")
	if err := copyFile(bundledFilePath, outputBundledPath); err != nil {
		return fmt.Errorf("failed to copy bundled Lua file: %w", err)
	}
	fmt.Printf("Bundled Lua file copied to: %s\n", outputBundledPath)
	
	fmt.Printf("✅ AOS build with injection completed successfully!\n")
	fmt.Printf("Output directory: %s\n", outputDir)
	return nil
}

// buildWithDocker runs the Docker container to build the WASM module
func (b *AOSBuilder) buildWithDocker(ctx context.Context, processDir string) error {
	fmt.Printf("Building WASM module in directory: %s\n", processDir)
	
	// Get absolute path for Docker volume mount
	absProcessDir, err := filepath.Abs(processDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	// Get the Docker image name from the runner or use default
	imageName := b.runner.GetImageName()
	
	fmt.Printf("Using absolute path for Docker mount: %s\n", absProcessDir)
	
	// Docker command: docker run --platform linux/amd64 -v ${pwd}:/src p3rmaw3b/ao:${VERSION} ao-build-module
	cmd := exec.CommandContext(ctx, 
		"docker", "run", 
		"--platform", "linux/amd64",
		"-v", fmt.Sprintf("%s:/src", absProcessDir),
		imageName,
		"ao-build-module",
	)
	
	// Set working directory and capture output
	cmd.Dir = processDir
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		fmt.Printf("Docker build failed with output:\n%s\n", string(output))
		return fmt.Errorf("docker build failed: %w", err)
	}
	
	fmt.Printf("Docker build completed successfully:\n%s\n", string(output))
	
	// Verify that process.wasm was created
	wasmPath := filepath.Join(processDir, "process.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		return fmt.Errorf("process.wasm was not created by the build process")
	}
	
	fmt.Printf("✅ WASM module successfully built: %s\n", wasmPath)
	return nil
}

// copyBuildOutputs copies build artifacts to the output directory
func (b *AOSBuilder) CopyBuildOutputs(processDir, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy process.wasm if it exists
	wasmFile := filepath.Join(processDir, "process.wasm")
	if _, err := os.Stat(wasmFile); err == nil {
		outputWasm := filepath.Join(outputDir, "process.wasm")
		if err := copyFile(wasmFile, outputWasm); err != nil {
			return fmt.Errorf("failed to copy process.wasm: %w", err)
		}
		fmt.Printf("Copied process.wasm to %s\n", outputWasm)
	}

	return nil
}
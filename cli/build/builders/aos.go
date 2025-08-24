package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	harlequinBuild "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/debug"
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

	// Simple usage with named parameters (recommended)
	config := &harlequinConfig.Config{...}
	builder := NewAOSBuilder(AOSBuilderParams{
		Config:     config,
		Entrypoint: "./main.lua",
		OutputDir:  "./dist",
		Callbacks:  CallbacksProgress, // or nil for default, CallbacksSilent for quiet
		// ConfigFilePath: nil // defaults to ".harlequin.yaml", or specify custom path
	})
	err := builder.Build(ctx) // Handles everything: prepare, bundle, inject, build, cleanup (workspace auto-managed)

	// With custom config file
	customConfig := "./custom-ao-build-config.yml"
	builder := NewAOSBuilder(AOSBuilderParams{
		Config:         config,
		ConfigFilePath: &customConfig,
		Entrypoint:     "./main.lua",
		OutputDir:      "./dist",
	})

	// Convenience constructors (legacy support)
	builder := NewAOSBuilderWithDefaultCallbacks(config, configPath, entrypoint, outputDir) // Default logging
	builder := NewAOSBuilderSilent(config, configPath, entrypoint, outputDir)               // Silent operation

	// Legacy manual steps (deprecated but still supported)
	err := builder.PrepareWorkspace(ctx, workspaceDir)        // DEPRECATED
	err = builder.BuildProjectWithInjection(ctx, projectPath, outputDir) // DEPRECATED
	err = builder.CleanWorkspace(workspaceDir)
*/
type AOSBuilder struct {
	entrypoint     string
	outputDir      string
	workspaceDir   string
	configFilePath string
	config         *harlequinConfig.Config
	runner         *harlequinBuild.BuildRunner
	callbacks      *BuildCallbacks
}

func NewAOSBuilder(params AOSBuilderParams) *AOSBuilder {
	// Generate a temporary workspace directory
	workspaceDir := filepath.Join(os.TempDir(), "harlequin-aos-build-"+generateRandomID())

	runner, err := harlequinBuild.NewAOBuildRunner(params.Config, workspaceDir)
	if err != nil {
		panic(err)
	}

	callbacks := params.Callbacks
	if callbacks == nil {
		callbacks = DefaultLoggingCallbacks()
	}

	// Use .harlequin.yaml as default config file path if not specified
	configFilePath := ".harlequin.yaml"
	if params.ConfigFilePath != nil {
		configFilePath = *params.ConfigFilePath
	}

	return &AOSBuilder{
		entrypoint:     params.Entrypoint,
		outputDir:      params.OutputDir,
		configFilePath: configFilePath,
		config:         params.Config,
		workspaceDir:   workspaceDir,
		runner:         runner,
		callbacks:      callbacks,
	}
}

// generateRandomID creates a random ID for temporary directories
func generateRandomID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// newAOSBuilderWithWorkspace creates an AOSBuilder with a custom workspace (for testing)
func newAOSBuilderWithWorkspace(params AOSBuilderParams, workspaceDir string) *AOSBuilder {
	runner, err := harlequinBuild.NewAOBuildRunner(params.Config, workspaceDir)
	if err != nil {
		panic(err)
	}

	callbacks := params.Callbacks
	if callbacks == nil {
		callbacks = DefaultLoggingCallbacks()
	}

	// Use .harlequin.yaml as default config file path if not specified
	configFilePath := ".harlequin.yaml"
	if params.ConfigFilePath != nil {
		configFilePath = *params.ConfigFilePath
	}

	return &AOSBuilder{
		entrypoint:     params.Entrypoint,
		outputDir:      params.OutputDir,
		configFilePath: configFilePath,
		config:         params.Config,
		workspaceDir:   workspaceDir,
		runner:         runner,
		callbacks:      callbacks,
	}
}

// NewAOSBuilderWithDefaultCallbacks creates an AOSBuilder with default logging callbacks (convenience function)
func NewAOSBuilderWithDefaultCallbacks(config *harlequinConfig.Config, configFilePath, entrypoint, outputDir string) *AOSBuilder {
	return NewAOSBuilder(AOSBuilderParams{
		Config:         config,
		ConfigFilePath: &configFilePath,
		Entrypoint:     entrypoint,
		OutputDir:      outputDir,
		Callbacks:      CallbacksDefault,
	})
}

// NewAOSBuilderSilent creates an AOSBuilder with no-op callbacks for silent operation
func NewAOSBuilderSilent(config *harlequinConfig.Config, configFilePath, entrypoint, outputDir string) *AOSBuilder {
	return NewAOSBuilder(AOSBuilderParams{
		Config:         config,
		ConfigFilePath: &configFilePath,
		Entrypoint:     entrypoint,
		OutputDir:      outputDir,
		Callbacks:      CallbacksSilent,
	})
}

// executeStep runs a build step and calls the appropriate callback
func (b *AOSBuilder) executeStep(ctx context.Context, stepName string, callback func(ctx context.Context, info BuildStepInfo), stepFunc func() error) error {
	startTime := time.Now()

	var err error
	var success bool

	defer func() {
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		info := BuildStepInfo{
			StepName:  stepName,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  duration,
			Success:   success,
			Error:     err,
			Metadata:  make(map[string]interface{}),
		}

		if callback != nil {
			callback(ctx, info)
		}
	}()

	err = stepFunc()
	success = err == nil
	return err
}

// Build performs the complete AOS build process: prepares workspace, bundles Lua, injects code, and builds WASM
func (b *AOSBuilder) Build(ctx context.Context) error {
	// Step 1: Prepare AOS workspace (clone AOS repo and copy files)
	if err := b.executeStep(ctx, "CopyAOSFiles", b.callbacks.OnCopyAOSFiles, func() error {
		// Check if config file actually exists before trying to copy it
		configFilePath := b.configFilePath
		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			// Config file doesn't exist, don't try to copy it
			configFilePath = ""
		}
		return b.CopyAOSFiles(ctx, b.workspaceDir, configFilePath)
	}); err != nil {
		return fmt.Errorf("failed to prepare workspace: %w", err)
	}

	// Step 2: Bundle the Lua project
	var bundledCode string
	if err := b.executeStep(ctx, "BundleLua", b.callbacks.OnBundleLua, func() error {
		var err error
		bundledCode, err = harlequinLuaUtils.Bundle(b.entrypoint)
		return err
	}); err != nil {
		return fmt.Errorf("failed to bundle Lua project: %w", err)
	}

	// Step 3: Write bundled code to workspace
	processDir := filepath.Join(b.workspaceDir, "aos-process")
	bundledFilePath := filepath.Join(processDir, "bundled.lua")
	if err := os.WriteFile(bundledFilePath, []byte(bundledCode), 0644); err != nil {
		return fmt.Errorf("failed to write bundled code: %w", err)
	}

	// Step 4: Inject the bundled code into the AOS process
	if err := b.executeStep(ctx, "InjectLua", b.callbacks.OnInjectLua, func() error {
		options := NewDefaultBuildInjectionOptions(processDir, bundledFilePath, ".bundled")
		return InjectBundledCode(options)
	}); err != nil {
		return fmt.Errorf("failed to inject bundled code: %w", err)
	}

	// Step 5: Build the project using Docker
	if err := b.executeStep(ctx, "WasmCompile", b.callbacks.OnWasmCompile, func() error {
		return b.buildWithDocker(ctx, processDir)
	}); err != nil {
		return fmt.Errorf("failed to build WASM: %w", err)
	}

	// Step 6: Copy outputs to the specified directory
	if err := b.executeStep(ctx, "CopyOutputs", b.callbacks.OnCopyOutputs, func() error {
		if err := b.CopyBuildOutputs(processDir, b.outputDir); err != nil {
			return err
		}

		// Also copy the bundled Lua file to output directory
		outputBundledPath := filepath.Join(b.outputDir, "bundled.lua")
		if err := copyFile(bundledFilePath, outputBundledPath); err != nil {
			return fmt.Errorf("failed to copy bundled Lua file: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to copy build outputs: %w", err)
	}

	// Clean up workspace automatically
	if err := b.executeStep(ctx, "Cleanup", b.callbacks.OnCleanup, func() error {
		return b.CleanWorkspace(b.workspaceDir)
	}); err != nil {
		debug.Printf("Warning: failed to clean workspace: %v\n", err)
		// Don't fail the build for cleanup issues
	}

	return nil
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

// Note: GetCopyOptions removed - was unused. AOS copying is handled internally by Build()

// AOSCopyOptions holds configuration for copying AOS process files
type AOSCopyOptions struct {
	RepoURL          string
	CommitHash       string
	TempRepoDir      string
	ProcessTargetDir string
	ConfigSourceFile string
	ConfigDestFile   string
}

// NewAOSCopyOptions creates default options for copying AOS process files
func NewAOSCopyOptions(config *harlequinConfig.Config, targetDir string) *AOSCopyOptions {
	tempRepoDir := filepath.Join(os.TempDir(), "harlequin-aos-repo")
	processTargetDir := filepath.Join(targetDir, "aos-process")
	configDestFile := filepath.Join(processTargetDir, "config.yml")

	return &AOSCopyOptions{
		RepoURL:          AOSRepoURL,
		CommitHash:       config.AOSGitHash,
		TempRepoDir:      tempRepoDir,
		ProcessTargetDir: processTargetDir,
		ConfigSourceFile: "", // Will be set based on where config is found
		ConfigDestFile:   configDestFile,
	}
}

// CopyAOSProcess clones the AOS repository and copies the process directory
func CopyAOSProcess(ctx context.Context, options *AOSCopyOptions) error {
	debug.Println("Starting AOS process copy...")

	// Step 1: Remove existing aos-process directory
	debug.Printf("Removing existing directory: %s\n", options.ProcessTargetDir)
	if err := os.RemoveAll(options.ProcessTargetDir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	// Step 2: Clean up any existing temp directory and clone the repository
	debug.Printf("Removing any existing temp directory: %s\n", options.TempRepoDir)
	if err := os.RemoveAll(options.TempRepoDir); err != nil {
		debug.Printf("Warning: failed to remove existing temp directory: %v\n", err)
	}

	// Cleanup temp repo on exit
	defer func() {
		debug.Printf("Removing temporary directory: %s\n", options.TempRepoDir)
		_ = os.RemoveAll(options.TempRepoDir)
	}()

	debug.Printf("Cloning repository: %s\n", options.RepoURL)
	cloneCmd := exec.CommandContext(ctx, "git", "clone", options.RepoURL, options.TempRepoDir)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Step 3: Checkout the specific commit hash
	debug.Printf("Checking out commit: %s\n", options.CommitHash)
	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", options.CommitHash)
	checkoutCmd.Dir = options.TempRepoDir
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout commit %s: %w", options.CommitHash, err)
	}

	// Step 4: Move the process directory to the target location
	processDir := filepath.Join(options.TempRepoDir, "process")
	debug.Printf("Moving %s to %s\n", processDir, options.ProcessTargetDir)

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(options.ProcessTargetDir), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	if err := copyDirectory(processDir, options.ProcessTargetDir); err != nil {
		return fmt.Errorf("failed to copy process directory: %w", err)
	}

	// Step 5: Copy the build config file to the target directory (if specified)
	if options.ConfigSourceFile != "" {
		debug.Printf("Copying %s to %s\n", options.ConfigSourceFile, options.ConfigDestFile)
		if err := copyFile(options.ConfigSourceFile, options.ConfigDestFile); err != nil {
			return fmt.Errorf("failed to copy config file: %w", err)
		}
	}

	debug.Println("Successfully copied AOS process and config.")
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

// PrepareAOSWorkspace prepares a workspace for AOS building by copying necessary files (DEPRECATED: integrated into Build())
func PrepareAOSWorkspace(ctx context.Context, config *harlequinConfig.Config, workspaceDir string) error {
	return fmt.Errorf("PrepareAOSWorkspace is deprecated and no longer supports automatic config file finding. Use CopyAOSProcessWithConfig with an explicit config file path, or use the AOSBuilder.Build() method instead")
}

// Note: findConfigFile removed - config file path resolution is now the responsibility
// of the Config package, not the builder. The builder receives a resolved config file path.

// CleanAOSWorkspace removes AOS-related files from the workspace
func CleanAOSWorkspace(workspaceDir string) error {
	aosProcessDir := filepath.Join(workspaceDir, "aos-process")
	debug.Printf("Cleaning AOS workspace: %s\n", aosProcessDir)
	return os.RemoveAll(aosProcessDir)
}

// Note: BuildInjectionOptions moved to types.go to avoid duplication

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
	debug.Printf("Injecting bundled code into: %s\n", options.ProcessFilePath)

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

	debug.Printf("Successfully injected bundled code require: require('%s')\n", options.RequireName)
	return nil
}

// injectRequireStatement injects the bundled code require after the last Handlers.append
func injectRequireStatement(content, requireName string) (string, error) {
	// Check if the require statement is already present
	existingRequire := fmt.Sprintf("require('%s')", requireName)
	if strings.Contains(content, existingRequire) {
		debug.Printf("The require('%s') statement is already present\n", requireName)
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

	debug.Printf("Injected require('%s') after the last Handlers.append\n", requireName)
	return result, nil
}

// Note: InjectBundledCodeIntoProcess and InjectBundledCodeWithOptions removed
// These were unused convenience methods. Injection is now handled internally by Build()

// BuildProjectWithInjection builds a project with bundled code injection (DEPRECATED: use Build() instead)
func (b *AOSBuilder) BuildProjectWithInjection(ctx context.Context, projectPath, outputDir string) error {
	debug.Println("⚠️  BuildProjectWithInjection is deprecated. Use Build() method instead.")

	// Create a temporary builder with the legacy parameters
	oldEntrypoint := b.entrypoint
	oldOutputDir := b.outputDir

	// Update builder with legacy parameters
	b.entrypoint = filepath.Join(projectPath, "main.lua")
	b.outputDir = outputDir

	// Call the new Build method
	err := b.Build(ctx)

	// Restore original parameters
	b.entrypoint = oldEntrypoint
	b.outputDir = oldOutputDir

	return err
}

// buildWithDocker runs the Docker container to build the WASM module
func (b *AOSBuilder) buildWithDocker(ctx context.Context, processDir string) error {
	debug.Printf("Building WASM module in directory: %s\n", processDir)

	// Get absolute path for Docker volume mount
	absProcessDir, err := filepath.Abs(processDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get the Docker image name from the runner or use default
	imageName := b.runner.GetImageName()

	debug.Printf("Using absolute path for Docker mount: %s\n", absProcessDir)

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
		debug.Printf("Docker build failed with output:\n%s\n", string(output))
		return fmt.Errorf("docker build failed: %w", err)
	}

	debug.Printf("Docker build completed successfully:\n%s\n", string(output))

	// Verify that process.wasm was created
	wasmPath := filepath.Join(processDir, "process.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		return fmt.Errorf("process.wasm was not created by the build process")
	}

	debug.Printf("✅ WASM module successfully built: %s\n", wasmPath)
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
		debug.Printf("Copied process.wasm to %s\n", outputWasm)
	}

	return nil
}

package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	harlequinBuild "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/build"
	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
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
3. Bundle the lua code and write the bundle to the build directory 
	✅ added luautils to bundle the lua code
4. Inject the bundle into the AOS code
5. Call the container to build the project
6. Write the process.wasm and the bundled lua code to the output directory
7. Clean up the build container and build directory
	✅ added CleanAOSWorkspace functionality

Usage:
	builder := NewAOSBuilder(config)
	
	// Prepare workspace with AOS files
	err := builder.PrepareWorkspace(ctx, workspaceDir)
	
	// Build the project
	err = builder.Build(ctx, projectPath)
	
	// Clean up
	err = builder.CleanWorkspace(workspaceDir)
*/

type AOSBuilder struct {
	entrypoint string
	outputDir  string
	config *harlequinConfig.Config
	runner *harlequinBuild.BuildRunner
}

func NewAOSBuilder(config *harlequinConfig.Config) *AOSBuilder {
	runner, err := harlequinBuild.NewAOBuildRunner(config, "")
	if err != nil {
		panic(err)
	}
	return &AOSBuilder{config: config, runner: runner}
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
	
	if err := os.Rename(processDir, options.ProcessTargetDir); err != nil {
		return fmt.Errorf("failed to move process directory: %w", err)
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
package build

import (
	"context"
	"fmt"
	"path/filepath"

	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

// BuildRunner manages the build process using Docker containers
type BuildRunner struct {
	dockerManager *DockerManager
	config        *harlequinConfig.Config
	workspaceDir  string
}

// NewBuildRunner creates a new build runner instance with specified image and container name
func NewBuildRunner(cfg *harlequinConfig.Config, workspaceDir, imageName, containerName string) (*BuildRunner, error) {
	dockerManager, err := NewDockerManager(imageName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker manager: %w", err)
	}

	return &BuildRunner{
		dockerManager: dockerManager,
		config:        cfg,
		workspaceDir:  workspaceDir,
	}, nil
}

// NewAOBuildRunner creates a new build runner instance with default AO image and container name
func NewAOBuildRunner(cfg *harlequinConfig.Config, workspaceDir string) (*BuildRunner, error) {
	return NewBuildRunner(cfg, workspaceDir, AOBuildContainerDockerImage, ContainerName)
}

// Close closes the build runner and cleans up resources
func (br *BuildRunner) Close() error {
	return br.dockerManager.Close()
}

// GetImageName returns the Docker image name being used
func (br *BuildRunner) GetImageName() string {
	return br.dockerManager.GetImageName()
}

// GetContainerName returns the Docker container name being used
func (br *BuildRunner) GetContainerName() string {
	return br.dockerManager.GetContainerName()
}

// StartBuildEnvironment ensures the build container is running and ready
func (br *BuildRunner) StartBuildEnvironment(ctx context.Context) error {
	fmt.Println("Starting build environment...")
	
	if err := br.dockerManager.StartContainer(ctx, br.workspaceDir); err != nil {
		return fmt.Errorf("failed to start build container: %w", err)
	}

	// Verify container is running
	running, err := br.dockerManager.IsContainerRunning(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify container status: %w", err)
	}
	if !running {
		return fmt.Errorf("container failed to start properly")
	}

	fmt.Println("Build environment ready!")
	return nil
}

// StopBuildEnvironment stops and cleans up the build container
func (br *BuildRunner) StopBuildEnvironment(ctx context.Context) error {
	fmt.Println("Stopping build environment...")
	
	if err := br.dockerManager.StopContainer(ctx); err != nil {
		return fmt.Errorf("failed to stop build container: %w", err)
	}
	
	if err := br.dockerManager.RemoveContainer(ctx); err != nil {
		return fmt.Errorf("failed to remove build container: %w", err)
	}

	fmt.Println("Build environment stopped!")
	return nil
}

// BuildProject executes the build process for a project
func (br *BuildRunner) BuildProject(ctx context.Context, projectPath string) error {
	// Ensure build environment is running
	if err := br.StartBuildEnvironment(ctx); err != nil {
		return err
	}

	// Create build configuration file in the container
	if err := br.setupBuildConfig(ctx); err != nil {
		return fmt.Errorf("failed to setup build config: %w", err)
	}

	// Change to project directory
	relativeProjectPath, err := filepath.Rel(br.workspaceDir, projectPath)
	if err != nil {
		return fmt.Errorf("failed to get relative project path: %w", err)
	}

	fmt.Printf("Building project at: %s\n", relativeProjectPath)

	// Execute the build command
	buildCmd := []string{
		"sh", "-c", 
		fmt.Sprintf("cd %s && ao-dev-cli build", relativeProjectPath),
	}

	if err := br.dockerManager.ExecCommand(ctx, buildCmd); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("Build completed successfully!")
	return nil
}

// RunCommand executes an arbitrary command in the build environment
func (br *BuildRunner) RunCommand(ctx context.Context, command []string) error {
	// Ensure build environment is running
	if err := br.StartBuildEnvironment(ctx); err != nil {
		return err
	}

	fmt.Printf("Executing command: %v\n", command)
	return br.dockerManager.ExecCommand(ctx, command)
}

// RunCommandWithOutput executes a command and returns its output
func (br *BuildRunner) RunCommandWithOutput(ctx context.Context, command []string) (string, error) {
	// Ensure build environment is running
	if err := br.StartBuildEnvironment(ctx); err != nil {
		return "", err
	}

	return br.dockerManager.ExecCommandWithOutput(ctx, command)
}

// setupBuildConfig creates the build configuration file inside the container
func (br *BuildRunner) setupBuildConfig(ctx context.Context) error {
	// Write config to temporary file in workspace
	configPath := filepath.Join(br.workspaceDir, ".harlequin-build-config.yaml")
	if err := harlequinConfig.WriteConfigFile(br.config, configPath); err != nil {
		return fmt.Errorf("failed to write build config: %w", err)
	}

	// The file is now available in the container at /workspace/.harlequin-build-config.yaml
	return nil
}

// GetBuildStatus returns information about the current build environment
func (br *BuildRunner) GetBuildStatus(ctx context.Context) (*BuildStatus, error) {
	running, err := br.dockerManager.IsContainerRunning(ctx)
	if err != nil {
		return nil, err
	}

	status := &BuildStatus{
		ContainerRunning: running,
		ImageName:        br.dockerManager.GetImageName(),
		ContainerName:    br.dockerManager.GetContainerName(),
		WorkspaceDir:     br.workspaceDir,
		Config:          br.config,
	}

	if running {
		// Get container info
		output, err := br.dockerManager.ExecCommandWithOutput(ctx, []string{"pwd"})
		if err == nil {
			status.ContainerWorkingDir = output
		}
	}

	return status, nil
}

// BuildStatus represents the current status of the build environment
type BuildStatus struct {
	ContainerRunning    bool           `json:"container_running"`
	ImageName          string         `json:"image_name"`
	ContainerName      string         `json:"container_name"`
	WorkspaceDir       string         `json:"workspace_dir"`
	ContainerWorkingDir string         `json:"container_working_dir,omitempty"`
	Config             *harlequinConfig.Config `json:"config"`
}

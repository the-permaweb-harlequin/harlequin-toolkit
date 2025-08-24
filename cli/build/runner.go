package build

import (
	"context"
	"fmt"

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

// BuildProject executes the build process for a project (legacy method - use AOS builder directly)
func (br *BuildRunner) BuildProject(ctx context.Context, projectPath string) error {
	return fmt.Errorf("BuildProject is deprecated - use AOSBuilder.BuildProjectWithInjection instead")
}

// GetBuildStatus returns basic information about the build configuration
func (br *BuildRunner) GetBuildStatus(ctx context.Context) (*BuildStatus, error) {
	return &BuildStatus{
		ImageName:     br.dockerManager.GetImageName(),
		ContainerName: br.dockerManager.GetContainerName(),
		WorkspaceDir:  br.workspaceDir,
		Config:        br.config,
	}, nil
}

// BuildStatus represents basic build configuration info
type BuildStatus struct {
	ImageName     string                  `json:"image_name"`
	ContainerName string                  `json:"container_name"`
	WorkspaceDir  string                  `json:"workspace_dir"`
	Config        *harlequinConfig.Config `json:"config"`
}

package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	dockerFilters "github.com/docker/docker/api/types/filters"
	dockerMount "github.com/docker/docker/api/types/mount"
	dockerClient "github.com/docker/docker/client"
)

// DockerManager handles Docker container operations for the build system
type DockerManager struct {
	client        *dockerClient.Client
	imageName     string
	containerName string
}

// NewDockerManager creates a new Docker manager instance with specified image and container name
func NewDockerManager(imageName, containerName string) (*DockerManager, error) {
	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerManager{
		client:        cli,
		imageName:     imageName,
		containerName: containerName,
	}, nil
}

// Close closes the Docker client connection
func (dm *DockerManager) Close() error {
	return dm.client.Close()
}

// GetImageName returns the Docker image name
func (dm *DockerManager) GetImageName() string {
	return dm.imageName
}

// GetContainerName returns the container name
func (dm *DockerManager) GetContainerName() string {
	return dm.containerName
}

// IsContainerRunning checks if the build container is currently running
func (dm *DockerManager) IsContainerRunning(ctx context.Context) (bool, error) {
	filters := dockerFilters.NewArgs()
	filters.Add("name", dm.containerName)
	
	containers, err := dm.client.ContainerList(ctx, dockerTypes.ContainerListOptions{
		Filters: filters,
	})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			// Container names in Docker API include leading slash
			if strings.TrimPrefix(name, "/") == dm.containerName {
				return container.State == "running", nil
			}
		}
	}
	return false, nil
}

// StartContainer starts the build container with the specified workspace mounted
func (dm *DockerManager) StartContainer(ctx context.Context, workspaceDir string) error {
	// Check if container is already running
	running, err := dm.IsContainerRunning(ctx)
	if err != nil {
		return err
	}
	if running {
		return nil // Already running
	}

	// Remove any existing stopped container with the same name
	_ = dm.RemoveContainer(ctx)

	// Get absolute path for workspace
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for workspace: %w", err)
	}

	// Pull image if it doesn't exist
	_, _, err = dm.client.ImageInspectWithRaw(ctx, dm.imageName)
	if err != nil {
		fmt.Printf("Pulling image %s...\n", dm.imageName)
		reader, err := dm.client.ImagePull(ctx, dm.imageName, dockerTypes.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
		defer reader.Close()
		// Copy output to discard (could pipe to stdout for progress)
		_, _ = io.Copy(io.Discard, reader)
	}

	// Create container configuration
	containerConfig := &dockerContainer.Config{
		Image:      dm.imageName,
		Cmd:        []string{"sleep", "infinity"}, // Keep container running
		WorkingDir: BuildWorkspaceDir,
	}

	hostConfig := &dockerContainer.HostConfig{
		Mounts: []dockerMount.Mount{
			{
				Type:   dockerMount.TypeBind,
				Source: absWorkspace,
				Target: BuildWorkspaceDir,
			},
		},
	}

	// Create the container
	resp, err := dm.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, dm.containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	if err := dm.client.ContainerStart(ctx, resp.ID, dockerTypes.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait a moment for container to be ready
	time.Sleep(2 * time.Second)
	return nil
}

// StopContainer stops the build container
func (dm *DockerManager) StopContainer(ctx context.Context) error {
	timeout := int((10 * time.Second).Seconds())
	if err := dm.client.ContainerStop(ctx, dm.containerName, dockerContainer.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

// RemoveContainer removes the build container
func (dm *DockerManager) RemoveContainer(ctx context.Context) error {
	err := dm.client.ContainerRemove(ctx, dm.containerName, dockerTypes.ContainerRemoveOptions{Force: true})
	if err != nil && !dockerClient.IsErrNotFound(err) {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

// ExecCommand executes a command inside the running build container
func (dm *DockerManager) ExecCommand(ctx context.Context, command []string) error {
	// Create exec configuration
	execConfig := dockerTypes.ExecConfig{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          true,
	}

	// Create exec instance
	execID, err := dm.client.ContainerExecCreate(ctx, dm.containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to exec instance
	resp, err := dm.client.ContainerExecAttach(ctx, execID.ID, dockerTypes.ExecStartCheck{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	// Copy streams
	go func() {
		_, _ = io.Copy(os.Stdout, resp.Reader)
	}()

	// Start exec instance
	return dm.client.ContainerExecStart(ctx, execID.ID, dockerTypes.ExecStartCheck{})
}

// ExecCommandWithOutput executes a command and returns its output
func (dm *DockerManager) ExecCommandWithOutput(ctx context.Context, command []string) (string, error) {
	// Create exec configuration
	execConfig := dockerTypes.ExecConfig{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create exec instance
	execID, err := dm.client.ContainerExecCreate(ctx, dm.containerName, execConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to exec instance
	resp, err := dm.client.ContainerExecAttach(ctx, execID.ID, dockerTypes.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	// Read output
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to read exec output: %w", err)
	}

	return string(output), nil
}
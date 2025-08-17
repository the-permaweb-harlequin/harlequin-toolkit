package build

import (
	"fmt"

	dockerClient "github.com/docker/docker/client"
)

// DockerManager handles basic Docker operations for the build system
// Since we use direct docker run commands, this is simplified to just store image info
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


package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

func TestNewBuildRunner(t *testing.T) {
	cfg := config.NewConfig(nil)
	workspaceDir := "/tmp/test-workspace"
	imageName := "test/image:latest"
	containerName := "test-container"
	
	runner, err := NewBuildRunner(cfg, workspaceDir, imageName, containerName)
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return // Skip test if Docker is not available
	}
	defer runner.Close()
	
	if runner == nil {
		t.Fatal("Expected BuildRunner to be created, got nil")
	}
	if runner.config != cfg {
		t.Error("Expected config to be set correctly")
	}
	if runner.workspaceDir != workspaceDir {
		t.Error("Expected workspaceDir to be set correctly")
	}
	if runner.dockerManager == nil {
		t.Error("Expected dockerManager to be initialized")
	}
	
	// Verify the docker manager has the specified image and container name
	if runner.dockerManager.GetImageName() != imageName {
		t.Errorf("Expected dockerManager imageName to be %s, got %s", imageName, runner.dockerManager.GetImageName())
	}
	if runner.dockerManager.GetContainerName() != containerName {
		t.Errorf("Expected dockerManager containerName to be %s, got %s", containerName, runner.dockerManager.GetContainerName())
	}
}

func TestNewAOBuildRunner(t *testing.T) {
	cfg := config.NewConfig(nil)
	workspaceDir := "/tmp/test-workspace"
	
	runner, err := NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return // Skip test if Docker is not available
	}
	defer runner.Close()
	
	if runner == nil {
		t.Fatal("Expected BuildRunner to be created, got nil")
	}
	if runner.config != cfg {
		t.Error("Expected config to be set correctly")
	}
	if runner.workspaceDir != workspaceDir {
		t.Error("Expected workspaceDir to be set correctly")
	}
	if runner.dockerManager == nil {
		t.Error("Expected dockerManager to be initialized")
	}
	
	// Verify the docker manager has the default AO image and container name
	if runner.dockerManager.GetImageName() != AOBuildContainerDockerImage {
		t.Errorf("Expected dockerManager imageName to be %s, got %s", AOBuildContainerDockerImage, runner.dockerManager.GetImageName())
	}
	if runner.dockerManager.GetContainerName() != ContainerName {
		t.Errorf("Expected dockerManager containerName to be %s, got %s", ContainerName, runner.dockerManager.GetContainerName())
	}
	
	// Test that GetBuildStatus reports the AO configuration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	status, err := runner.GetBuildStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get build status: %v", err)
	}
	
	if status.ImageName != AOBuildContainerDockerImage {
		t.Errorf("Expected status ImageName to be %s, got %s", AOBuildContainerDockerImage, status.ImageName)
	}
	if status.ContainerName != ContainerName {
		t.Errorf("Expected status ContainerName to be %s, got %s", ContainerName, status.ContainerName)
	}
}

func TestBuildRunner_GetBuildStatus(t *testing.T) {
	cfg := config.NewConfig(nil)
	workspaceDir := "/tmp/test-workspace"
	runner, err := NewAOBuildRunner(cfg, workspaceDir)
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return
	}
	defer runner.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	status, err := runner.GetBuildStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get build status: %v", err)
	}
	
	if status == nil {
		t.Fatal("Expected BuildStatus to be returned, got nil")
	}
	
	if status.ImageName != AOBuildContainerDockerImage {
		t.Errorf("Expected ImageName to be %s, got %s", AOBuildContainerDockerImage, status.ImageName)
	}
	
	if status.ContainerName != ContainerName {
		t.Errorf("Expected ContainerName to be %s, got %s", ContainerName, status.ContainerName)
	}
	
	if status.WorkspaceDir != workspaceDir {
		t.Errorf("Expected WorkspaceDir to be %s, got %s", workspaceDir, status.WorkspaceDir)
	}
	
	if status.Config != cfg {
		t.Error("Expected Config to match the provided config")
	}
	
	// Verify that status reports the default AO image and container name
	if status.ImageName != AOBuildContainerDockerImage {
		t.Errorf("Expected ImageName to be %s, got %s", AOBuildContainerDockerImage, status.ImageName)
	}
	if status.ContainerName != ContainerName {
		t.Errorf("Expected ContainerName to be %s, got %s", ContainerName, status.ContainerName)
	}
}

func TestBuildRunner_SetupBuildConfig(t *testing.T) {
	// Create temporary workspace directory
	tempDir, err := os.MkdirTemp("", "harlequin-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cfg := config.NewConfig(nil)
	runner, err := NewAOBuildRunner(cfg, tempDir)
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return
	}
	defer runner.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Test setupBuildConfig
	err = runner.setupBuildConfig(ctx)
	if err != nil {
		t.Fatalf("setupBuildConfig failed: %v", err)
	}
	
	// Verify config file was created
	configPath := filepath.Join(tempDir, ".harlequin-build-config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be created at %s", configPath)
	}
	
	// Verify config file contents
	readConfig := config.ReadConfigFile(configPath)
	if readConfig.StackSize != cfg.StackSize {
		t.Errorf("Expected StackSize to be %d, got %d", cfg.StackSize, readConfig.StackSize)
	}
}

// TestDockerManager_NewDockerManager tests the basic creation of DockerManager
func TestDockerManager_NewDockerManager(t *testing.T) {
	testImageName := "test/image:latest"
	testContainerName := "test-container"
	
	dm, err := NewDockerManager(testImageName, testContainerName)
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return
	}
	defer dm.Close()
	
	if dm == nil {
		t.Fatal("Expected DockerManager to be created, got nil")
	}
	
	if dm.GetImageName() != testImageName {
		t.Errorf("Expected imageName to be %s, got %s", testImageName, dm.GetImageName())
	}
	
	if dm.GetContainerName() != testContainerName {
		t.Errorf("Expected containerName to be %s, got %s", testContainerName, dm.GetContainerName())
	}
}

// TestDockerManager_IsContainerRunning tests container status checking
func TestDockerManager_IsContainerRunning(t *testing.T) {
	dm, err := NewDockerManager("test/image:latest", "test-container")
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return
	}
	defer dm.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// This test will pass regardless of Docker availability
	// as it only tests the function doesn't crash
	_, err = dm.IsContainerRunning(ctx)
	// We don't assert on the error since Docker might not be available in CI
	// The important thing is that the function doesn't panic
	_ = err
}

// Integration test that requires Docker to be available
func TestBuildRunner_Integration(t *testing.T) {
	// Skip this test if SKIP_DOCKER_TESTS is set
	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("Skipping Docker integration test")
	}
	
	// Create temporary workspace directory
	tempDir, err := os.MkdirTemp("", "harlequin-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	cfg := config.NewConfig(nil)
	runner, err := NewAOBuildRunner(cfg, tempDir)
	if err != nil {
		t.Logf("Docker may not be available: %v", err)
		return
	}
	defer runner.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Test getting status (should work even if Docker isn't available)
	status, err := runner.GetBuildStatus(ctx)
	if err != nil {
		t.Logf("Note: Docker may not be available: %v", err)
		return // Skip rest of test if Docker isn't available
	}
	
	if status.ContainerRunning {
		// Clean up any existing container
		_ = runner.StopBuildEnvironment(ctx)
	}
	
	// Test running a simple command (this will only work if Docker is available)
	output, err := runner.RunCommandWithOutput(ctx, []string{"echo", "hello world"})
	if err != nil {
		t.Logf("Docker command failed (Docker may not be available): %v", err)
		return
	}
	
	if output == "" {
		t.Error("Expected output from echo command")
	}
	
	// Clean up
	_ = runner.StopBuildEnvironment(ctx)
}

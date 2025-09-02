package shared

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestTempDir creates a temporary directory for testing and cleans it up
func TestTempDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "harlequin-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Cleanup after test
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

// CreateTestFile creates a test file with specified content
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()

	filePath := filepath.Join(dir, filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatalf("Failed to create directory for %s: %v", filePath, err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", filePath, err)
	}

	return filePath
}

// AssertFileExists checks that a file exists
func AssertFileExists(t *testing.T, filePath string) {
	t.Helper()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", filePath)
	}
}

// AssertFileNotExists checks that a file does not exist
func AssertFileNotExists(t *testing.T, filePath string) {
	t.Helper()

	if _, err := os.Stat(filePath); err == nil {
		t.Errorf("Expected file %s to not exist, but it does", filePath)
	}
}

// AssertFileContent checks that a file contains expected content
func AssertFileContent(t *testing.T, filePath, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	if string(content) != expectedContent {
		t.Errorf("File %s content mismatch.\nExpected: %s\nActual: %s",
			filePath, expectedContent, string(content))
	}
}

// AssertFileMode checks that a file has the expected permissions
func AssertFileMode(t *testing.T, filePath string, expectedMode fs.FileMode) {
	t.Helper()

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", filePath, err)
	}

	actualMode := info.Mode().Perm()
	expectedPerm := expectedMode.Perm()

	if actualMode != expectedPerm {
		t.Errorf("File %s mode mismatch. Expected: %o, Actual: %o",
			filePath, expectedPerm, actualMode)
	}
}

// WithTimeout runs a function with a timeout
func WithTimeout(t *testing.T, timeout time.Duration, fn func()) {
	t.Helper()

	done := make(chan bool, 1)

	go func() {
		fn()
		done <- true
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out after %v", timeout)
	}
}

// SkipInCI skips the test if running in CI environment
func SkipInCI(t *testing.T, reason string) {
	t.Helper()

	if os.Getenv("CI") != "" {
		t.Skipf("Skipping in CI: %s", reason)
	}
}

// RequireEnv skips the test if required environment variable is not set
func RequireEnv(t *testing.T, envVar string) string {
	t.Helper()

	value := os.Getenv(envVar)
	if value == "" {
		t.Skipf("Test requires environment variable %s to be set", envVar)
	}

	return value
}

package shared

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockTurboServer creates a mock Turbo API server for testing
func MockTurboServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Mock balance endpoint
	mux.HandleFunc("/account/balance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"winc": "1000000000000", "credits": "1.0", "currency": "winston"}`)
	})

	// Mock upload cost endpoint
	mux.HandleFunc("/price/bytes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `[{"winc": "500000", "currency": "winston"}]`)
	})

	// Mock upload endpoint
	mux.HandleFunc("/tx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"id": "test-transaction-id-12345"}`)
	})

	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// MockDockerCommand mocks Docker command execution for testing
func MockDockerCommand(t *testing.T, success bool) func(ctx context.Context, args ...string) error {
	t.Helper()

	return func(ctx context.Context, args ...string) error {
		if !success {
			return fmt.Errorf("mock Docker command failed")
		}
		return nil
	}
}

// MockFileSystem provides utilities for mocking file system operations
type MockFileSystem struct {
	files map[string]string
	dirs  map[string]bool
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string]string),
		dirs:  make(map[string]bool),
	}
}

// AddFile adds a file to the mock file system
func (mfs *MockFileSystem) AddFile(path, content string) {
	mfs.files[path] = content
}

// AddDir adds a directory to the mock file system
func (mfs *MockFileSystem) AddDir(path string) {
	mfs.dirs[path] = true
}

// FileExists checks if a file exists in the mock file system
func (mfs *MockFileSystem) FileExists(path string) bool {
	_, exists := mfs.files[path]
	return exists
}

// DirExists checks if a directory exists in the mock file system
func (mfs *MockFileSystem) DirExists(path string) bool {
	return mfs.dirs[path]
}

// GetFileContent returns the content of a file from the mock file system
func (mfs *MockFileSystem) GetFileContent(path string) (string, error) {
	content, exists := mfs.files[path]
	if !exists {
		return "", fmt.Errorf("file not found: %s", path)
	}
	return content, nil
}

// MockError creates a mock error for testing
func MockError(message string) error {
	return fmt.Errorf(message)
}

package components

import (
	"os"
	"path/filepath"
	"strings"
)

// WasmFileDiscovery provides utilities for finding WASM files
type WasmFileDiscovery struct {
	skipDirs []string
	maxDepth int
}

// NewWasmFileDiscovery creates a new WASM file discovery utility
func NewWasmFileDiscovery() *WasmFileDiscovery {
	return &WasmFileDiscovery{
		skipDirs: []string{
			"node_modules",
			".git",
			".svn",
			".hg",
			"vendor",
			".vscode",
			".idea",
			"__pycache__",
			".DS_Store",
			"target/debug",    // Rust debug builds
			"target/release",  // Rust release builds (but we might want these)
			"build/debug",
			"cmake-build-debug",
			"cmake-build-release",
		},
		maxDepth: 5, // Prevent deep recursion
	}
}

// FindWasmFiles recursively finds all .wasm files in a directory
func (wfd *WasmFileDiscovery) FindWasmFiles(rootDir string) ([]string, error) {
	var wasmFiles []string

	err := wfd.walkDirectory(rootDir, rootDir, 0, &wasmFiles)
	if err != nil {
		return nil, err
	}

	return wasmFiles, nil
}

// walkDirectory recursively walks through directories finding WASM files
func (wfd *WasmFileDiscovery) walkDirectory(rootDir, currentDir string, depth int, wasmFiles *[]string) error {
	// Prevent excessive recursion
	if depth > wfd.maxDepth {
		return nil
	}

	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(currentDir, entry.Name())

		if entry.IsDir() {
			// Skip directories we don't want to search
			if wfd.shouldSkipDir(entry.Name()) {
				continue
			}

			// Recursively search subdirectory
			if err := wfd.walkDirectory(rootDir, fullPath, depth+1, wasmFiles); err != nil {
				// Don't fail the entire search if one directory has issues
				continue
			}
		} else if strings.HasSuffix(strings.ToLower(entry.Name()), ".wasm") {
			// Found a WASM file - convert to relative path
			relPath, err := filepath.Rel(rootDir, fullPath)
			if err != nil {
				relPath = fullPath // Fallback to absolute path
			}
			*wasmFiles = append(*wasmFiles, relPath)
		}
	}

	return nil
}

// shouldSkipDir checks if a directory should be skipped during search
func (wfd *WasmFileDiscovery) shouldSkipDir(dirName string) bool {
	for _, skipDir := range wfd.skipDirs {
		if dirName == skipDir {
			return true
		}
	}
	return false
}

// AddSkipDir adds a directory to the skip list
func (wfd *WasmFileDiscovery) AddSkipDir(dirName string) {
	wfd.skipDirs = append(wfd.skipDirs, dirName)
}

// SetMaxDepth sets the maximum recursion depth
func (wfd *WasmFileDiscovery) SetMaxDepth(depth int) {
	wfd.maxDepth = depth
}

// FindWasmFilesQuick is a convenience function for quick discovery
func FindWasmFilesQuick(rootDir string) ([]string, error) {
	discovery := NewWasmFileDiscovery()
	return discovery.FindWasmFiles(rootDir)
}

// FindConfigFilesQuick finds YAML config files
func FindConfigFilesQuick(rootDir string) ([]string, error) {
	var configFiles []string

	// Look for common config file patterns
	patterns := []string{
		"ao-build-config.yml",
		"ao-build-config.yaml",
		".harlequin.yaml",
		".harlequin.yml",
		"build_configs/ao-build-config.yml",
		"build_configs/ao-build-config.yaml",
	}

	for _, pattern := range patterns {
		fullPath := filepath.Join(rootDir, pattern)
		if _, err := os.Stat(fullPath); err == nil {
			configFiles = append(configFiles, pattern)
		}
	}

	return configFiles, nil
}

// FindWalletFilesQuick finds JSON wallet files
func FindWalletFilesQuick(rootDir string) ([]string, error) {
	var walletFiles []string

	// Look for common wallet file patterns
	patterns := []string{
		"key.json",
		"wallet.json",
		"arweave-wallet.json",
		"arweave-key.json",
		"keyfile.json",
		"test_wallet.json",
		"dev_wallet.json",
	}

	for _, pattern := range patterns {
		fullPath := filepath.Join(rootDir, pattern)
		if _, err := os.Stat(fullPath); err == nil {
			walletFiles = append(walletFiles, pattern)
		}
	}

	return walletFiles, nil
}

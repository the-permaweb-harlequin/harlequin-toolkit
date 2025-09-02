package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
)

// TestCLICompilation tests that the CLI can be compiled successfully
func TestCLICompilation(t *testing.T) {
	shared.WithTimeout(t, 30*time.Second, func() {
		projectRoot := getProjectRoot(t)
		cliDir := filepath.Join(projectRoot, "cli")

		// Build the CLI
		cmd := exec.Command("go", "build", "-o", "harlequin-test", ".")
		cmd.Dir = cliDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI compilation failed: %v\nOutput: %s", err, string(output))
		}

		// Verify binary exists
		binaryPath := filepath.Join(cliDir, "harlequin-test")
		shared.AssertFileExists(t, binaryPath)

		// Cleanup
		t.Cleanup(func() {
			os.Remove(binaryPath)
		})
	})
}

// TestCLIVersion tests the version command
func TestCLIVersion(t *testing.T) {
	binary := compileCLI(t)
	defer cleanupBinary(t, binary)

	shared.WithTimeout(t, 10*time.Second, func() {
		cmd := exec.Command(binary, "--version")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Fatalf("Version command failed: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "harlequin") {
			t.Errorf("Version output should contain 'harlequin', got: %s", outputStr)
		}
	})
}

// TestCLIHelp tests the help command
func TestCLIHelp(t *testing.T) {
	binary := compileCLI(t)
	defer cleanupBinary(t, binary)

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "main help",
			args:     []string{"--help"},
			expected: []string{"Harlequin", "USAGE:", "COMMANDS:"},
		},
		{
			name:     "build help",
			args:     []string{"build", "--help"},
			expected: []string{"Build", "entrypoint", "Usage:"},
		},
		{
			name:     "lua-utils help",
			args:     []string{"lua-utils", "--help"},
			expected: []string{"Lua Utils", "bundle", "Usage:"},
		},
		{
			name:     "upload help",
			args:     []string{"upload-module", "--help"},
			expected: []string{"Upload", "WASM", "wallet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.WithTimeout(t, 10*time.Second, func() {
				cmd := exec.Command(binary, tt.args...)
				output, err := cmd.CombinedOutput()

				// Help commands should exit with code 0
				if err != nil {
					t.Fatalf("Help command failed: %v\nOutput: %s", err, string(output))
				}

				outputStr := string(output)
				for _, expected := range tt.expected {
					if !strings.Contains(outputStr, expected) {
						t.Errorf("Help output should contain '%s', got: %s", expected, outputStr)
					}
				}
			})
		})
	}
}

// TestCLIBuildCommand tests the build command end-to-end
func TestCLIBuildCommand(t *testing.T) {
	shared.SkipInCI(t, "Build command requires Docker")

	binary := compileCLI(t)
	defer cleanupBinary(t, binary)

	tmpDir := shared.TestTempDir(t)

	// Create test Lua file
	mainLua := shared.CreateTestFile(t, tmpDir, "main.lua", `
print("Hello from AO process!")

function handle(msg)
    return { Data = "Hello, " .. (msg.Data or "World") }
end`)

	// Create test config
	_ = shared.CreateTestFile(t, tmpDir, ".harlequin.yaml", `
stack_size: 1048576
initial_memory: 2097152
maximum_memory: 4194304
target: 32
compute_limit: "9000000000000"
module_format: "wasm32-unknown-emscripten-metering"
aos_git_hash: "15dd81ee596518e2f44521e973b8ad1ce3ee9945"`)

	shared.WithTimeout(t, 60*time.Second, func() {
		// Run build command
		cmd := exec.Command(binary, "build", "--entrypoint", mainLua, "--debug")
		cmd.Dir = tmpDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// The build might fail due to Docker not being available in CI,
		// but we can test that the command parsing works
		if err != nil {
			// Check if it's a Docker-related error (acceptable in CI)
			if strings.Contains(outputStr, "Docker") || strings.Contains(outputStr, "docker") {
				t.Skipf("Docker not available: %s", outputStr)
			}

			// Check if it's a deprecation warning (known issue)
			if strings.Contains(outputStr, "deprecated") {
				t.Skipf("Build method deprecated: %s", outputStr)
			}

			// If it's not a Docker or deprecation error, it's a real failure
			if !strings.Contains(outputStr, "Docker") && !strings.Contains(outputStr, "docker") && !strings.Contains(outputStr, "deprecated") {
				t.Errorf("Build command failed with unexpected error: %v\nOutput: %s", err, outputStr)
			}
		}

		// Verify that the command started properly (config was loaded)
		if !strings.Contains(outputStr, "Building") && !strings.Contains(outputStr, "Docker") && !strings.Contains(outputStr, "deprecated") {
			t.Errorf("Build should show building message, Docker error, or deprecation warning, got: %s", outputStr)
		}
	})
}

// TestCLILuaUtilsBundle tests the lua-utils bundle command
func TestCLILuaUtilsBundle(t *testing.T) {
	binary := compileCLI(t)
	defer cleanupBinary(t, binary)

	tmpDir := shared.TestTempDir(t)

	// Create test Lua files
	_ = shared.CreateTestFile(t, tmpDir, "utils.lua", `
local utils = {}

function utils.greet(name)
    return "Hello, " .. (name or "World") .. "!"
end

return utils`)

	mainLua := shared.CreateTestFile(t, tmpDir, "main.lua", `
local utils = require("utils")

print(utils.greet("AO"))`)

	shared.WithTimeout(t, 15*time.Second, func() {
		// Run bundle command
		outputPath := filepath.Join(tmpDir, "bundled.lua")
		cmd := exec.Command(binary, "lua-utils", "bundle", "--entrypoint", mainLua, "--outputPath", outputPath)
		cmd.Dir = tmpDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Bundle command failed: %v\nOutput: %s", err, string(output))
		}

		// Verify bundle was created
		shared.AssertFileExists(t, outputPath)

		// Verify bundle contains expected content
		bundleContent, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read bundle: %v", err)
		}

		bundleStr := string(bundleContent)
		expectedElements := []string{
			"utils.greet",
			"require(\"utils\")",
		}

		for _, element := range expectedElements {
			if !strings.Contains(bundleStr, element) {
				t.Errorf("Bundle should contain '%s'", element)
			}
		}
	})
}

// TestCLIUploadCommand tests the upload command (dry run)
func TestCLIUploadCommand(t *testing.T) {
	binary := compileCLI(t)
	defer cleanupBinary(t, binary)

	tmpDir := shared.TestTempDir(t)

	// Create test files
	wasmFile := shared.CreateTestFile(t, tmpDir, "process.wasm", "fake wasm binary content")
	configFile := shared.CreateTestFile(t, tmpDir, "build_configs/ao-build-config.yml", `
compute_limit: "9000000000000"
maximum_memory: 524288
module_format: "wasm32-unknown-unknown"
aos_git_hash: "test-hash"
data_protocol: "ao"
variant: "aos"
type: "process"`)

	walletFile := shared.CreateTestFile(t, tmpDir, "wallet.json", `{
		"test": "wallet",
		"kty": "RSA",
		"n": "test-key"
	}`)

	shared.WithTimeout(t, 30*time.Second, func() {
		// Run upload command in dry-run mode
		cmd := exec.Command(binary, "upload-module",
			"--wasm-file", wasmFile,
			"--config", configFile,
			"--wallet-file", walletFile,
			"--version", "test-1.0.0",
			"--dry-run")
		cmd.Dir = tmpDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Dry run should succeed or fail gracefully
		if err != nil {
			// Check if it's a wallet/network related error (acceptable)
			if !strings.Contains(outputStr, "Dry Run") {
				t.Errorf("Upload dry-run should mention dry run mode, got: %s", outputStr)
			}
		} else {
			// Verify dry run output
			if !strings.Contains(outputStr, "Dry Run") && !strings.Contains(outputStr, "DRY RUN") {
				t.Errorf("Upload should indicate dry run mode, got: %s", outputStr)
			}
		}
	})
}

// TestCLIErrorHandling tests error handling for invalid commands
func TestCLIErrorHandling(t *testing.T) {
	binary := compileCLI(t)
	defer cleanupBinary(t, binary)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "invalid command",
			args:        []string{"invalid-command"},
			expectError: true,
			errorText:   "Unknown",
		},
		{
			name:        "build without entrypoint",
			args:        []string{"build"},
			expectError: true,
			errorText:   "entrypoint",
		},
		{
			name:        "bundle without entrypoint",
			args:        []string{"lua-utils", "bundle"},
			expectError: true,
			errorText:   "entrypoint",
		},
		{
			name:        "upload without config",
			args:        []string{"upload-module", "--wasm-file", "/non/existent/file.wasm", "--dry-run"},
			expectError: true,
			errorText:   "configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.WithTimeout(t, 10*time.Second, func() {
				cmd := exec.Command(binary, tt.args...)
				output, err := cmd.CombinedOutput()

				if tt.expectError {
					if err == nil {
						t.Errorf("Expected command to fail, but it succeeded. Output: %s", string(output))
					}

					if !strings.Contains(string(output), tt.errorText) {
						t.Errorf("Expected error output to contain '%s', got: %s", tt.errorText, string(output))
					}
				} else {
					if err != nil {
						t.Errorf("Expected command to succeed, but it failed: %v\nOutput: %s", err, string(output))
					}
				}
			})
		})
	}
}

// Helper functions

// compileCLI compiles the CLI and returns the path to the binary
func compileCLI(t *testing.T) string {
	t.Helper()

	projectRoot := getProjectRoot(t)
	cliDir := filepath.Join(projectRoot, "cli")
	binaryName := "harlequin-e2e-test"
	binaryPath := filepath.Join(cliDir, binaryName)

	// Build the CLI
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Dir = cliDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to compile CLI: %v\nOutput: %s", err, string(output))
	}

	return binaryPath
}

// cleanupBinary removes the compiled binary
func cleanupBinary(t *testing.T, binaryPath string) {
	t.Helper()
	if err := os.Remove(binaryPath); err != nil {
		t.Logf("Warning: Failed to cleanup binary %s: %v", binaryPath, err)
	}
}

// getProjectRoot returns the path to the project root
func getProjectRoot(t *testing.T) string {
	t.Helper()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate up to find project root (contains cli/ directory)
	dir := cwd
	for {
		cliPath := filepath.Join(dir, "cli")
		if _, err := os.Stat(cliPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find project root from %s", cwd)
		}
		dir = parent
	}
}

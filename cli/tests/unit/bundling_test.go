package unit

import (
	"os"
	"path/filepath"
	"testing"

	luautils "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/lua_utils"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
)

func TestBundleWithFixtures(t *testing.T) {
	// Create temporary directory
	tmpDir := shared.TestTempDir(t)

	// Copy fixtures to temp directory
	mainContent := `-- Sample Lua file for testing bundling
local utils = require("utils")

function handle(msg)
    local action = msg.Tags.Action

    if action == "Info" then
        return utils.formatResponse("Process info", {
            name = "Test Process",
            version = "1.0.0"
        })
    elseif action == "Echo" then
        return utils.formatResponse("Echo", {
            message = msg.Data
        })
    else
        return utils.formatError("Unknown action: " .. action)
    end
end`

	utilsContent := `-- Sample utils module for testing
local utils = {}

function utils.formatResponse(action, data)
    return {
        Action = action,
        Data = data,
        Timestamp = os.time()
    }
end

function utils.formatError(message)
    return {
        Error = message,
        Timestamp = os.time()
    }
end

function utils.isEmpty(str)
    return str == nil or str == ""
end

return utils`

	// Create test files
	mainFile := shared.CreateTestFile(t, tmpDir, "main.lua", mainContent)
	shared.CreateTestFile(t, tmpDir, "utils.lua", utilsContent)

	// Test bundling
	bundledContent, err := luautils.Bundle(mainFile)
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// Verify bundle contains expected content
	expectedElements := []string{
		`-- module: "utils"`,
		"_loaded_mod_utils",
		`_G.package.loaded["utils"]`,
		"function handle(msg)",
		"formatResponse",
		"formatError",
	}

	for _, element := range expectedElements {
		if !contains(bundledContent, element) {
			t.Errorf("Bundle should contain: %q", element)
		}
	}
}

func TestBundleErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		entrypoint  string
		expectError bool
	}{
		{
			name:        "non-existent file",
			entrypoint:  "/non/existent/file.lua",
			expectError: true,
		},
		{
			name:        "empty file path",
			entrypoint:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := luautils.Bundle(tt.entrypoint)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestBundleOutputPath(t *testing.T) {
	tmpDir := shared.TestTempDir(t)

	// Create simple test file
	mainContent := `print("Hello, World!")`
	mainFile := shared.CreateTestFile(t, tmpDir, "test.lua", mainContent)

	// Bundle the file
	bundledContent, err := luautils.Bundle(mainFile)
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// Write to output file
	outputFile := filepath.Join(tmpDir, "output.bundled.lua")
	if err := os.WriteFile(outputFile, []byte(bundledContent), 0644); err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}

	// Verify output file exists and has correct permissions
	shared.AssertFileExists(t, outputFile)
	shared.AssertFileMode(t, outputFile, 0644)

	// Verify content
	shared.AssertFileContent(t, outputFile, bundledContent)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (substr == "" || s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

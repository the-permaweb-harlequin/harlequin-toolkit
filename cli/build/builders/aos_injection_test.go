package builders

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

func TestInjectRequireStatement(t *testing.T) {
	content := `
local something = "test"

Handlers.append("first", function() end)
Handlers.append("second", function() end)

print("done")
`

	result, err := injectRequireStatement(content, ".bundled")
	if err != nil {
		t.Fatalf("injectRequireStatement failed: %v", err)
	}

	// Check that require was injected
	if !strings.Contains(result, "require('.bundled');") {
		t.Error("Expected require('.bundled'); to be injected")
	}

	// Check that it was injected after the last Handlers.append
	lines := strings.Split(result, "\n")
	var lastHandlerIndex, requireIndex int = -1, -1

	for i, line := range lines {
		if strings.Contains(line, "Handlers.append") {
			lastHandlerIndex = i
		}
		if strings.Contains(line, "require('.bundled')") {
			requireIndex = i
		}
	}

	if lastHandlerIndex == -1 {
		t.Error("No Handlers.append found")
	}
	if requireIndex == -1 {
		t.Error("Require statement not found")
	}
	if requireIndex <= lastHandlerIndex {
		t.Error("Require should be injected after the last Handlers.append")
	}
}

func TestInjectRequireStatement_AlreadyExists(t *testing.T) {
	content := `
local something = "test"

Handlers.append("first", function() end)
require('.bundled');
Handlers.append("second", function() end)

print("done")
`

	result, err := injectRequireStatement(content, ".bundled")
	if err != nil {
		t.Fatalf("injectRequireStatement failed: %v", err)
	}

	// Count occurrences of the require statement
	requireCount := strings.Count(result, "require('.bundled')")
	if requireCount != 1 {
		t.Errorf("Expected exactly 1 require statement, found %d", requireCount)
	}

	// Should be unchanged since it already exists
	if result != content {
		t.Error("Content should be unchanged when require already exists")
	}
}

func TestInjectRequireStatement_NoHandlers(t *testing.T) {
	content := `
local something = "test"
print("no handlers here")
`

	_, err := injectRequireStatement(content, ".bundled")
	if err == nil {
		t.Error("Expected error when no Handlers.append found")
	}

	expectedError := "no Handlers.append found in process.lua"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedError, err)
	}
}

func TestInjectBundledCode(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "aos-injection-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test process.lua file
	processContent := `
local something = require('.something')

Handlers.append("test1", function() end)
Handlers.append("test2", function() end)

print("AOS Process")
`
	processFile := filepath.Join(tempDir, "process.lua")
	if err := os.WriteFile(processFile, []byte(processContent), 0644); err != nil {
		t.Fatalf("Failed to write process file: %v", err)
	}

	// Create test bundled code file
	bundledContent := `print("Bundled code")`
	bundledFile := filepath.Join(tempDir, "bundled.lua")
	if err := os.WriteFile(bundledFile, []byte(bundledContent), 0644); err != nil {
		t.Fatalf("Failed to write bundled file: %v", err)
	}

	// Test injection
	options := NewDefaultBuildInjectionOptions(tempDir, bundledFile, ".bundled")
	err = InjectBundledCode(options)
	if err != nil {
		t.Fatalf("InjectBundledCode failed: %v", err)
	}

	// Verify results
	result, err := os.ReadFile(processFile)
	if err != nil {
		t.Fatalf("Failed to read modified process file: %v", err)
	}

	resultContent := string(result)

	// Check that bundled require was injected
	if !strings.Contains(resultContent, "require('.bundled');") {
		t.Error("Expected bundled require to be injected")
	}

	// Check that existing content is preserved
	if !strings.Contains(resultContent, "local something = require('.something')") {
		t.Error("Expected existing requires to be preserved")
	}
	if !strings.Contains(resultContent, `print("AOS Process")`) {
		t.Error("Expected original content to be preserved")
	}
}

func TestNewDefaultBuildInjectionOptions(t *testing.T) {
	processDir := "/tmp/test-process"
	bundledCodePath := "/tmp/bundled.lua"
	requireName := ".my-bundle"

	options := NewDefaultBuildInjectionOptions(processDir, bundledCodePath, requireName)

	expectedProcessFile := filepath.Join(processDir, "process.lua")
	if options.ProcessFilePath != expectedProcessFile {
		t.Errorf("Expected ProcessFilePath to be '%s', got '%s'", expectedProcessFile, options.ProcessFilePath)
	}

	if options.BundledCodePath != bundledCodePath {
		t.Errorf("Expected BundledCodePath to be '%s', got '%s'", bundledCodePath, options.BundledCodePath)
	}

	if options.RequireName != requireName {
		t.Errorf("Expected RequireName to be '%s', got '%s'", requireName, options.RequireName)
	}
}

func TestInjectBundledCode_FileNotFound(t *testing.T) {
	options := &BuildInjectionOptions{
		ProcessFilePath: "/nonexistent/process.lua",
		BundledCodePath: "/nonexistent/bundled.lua",
		RequireName:     ".test",
	}

	err := InjectBundledCode(options)
	if err == nil {
		t.Error("Expected error when process file doesn't exist")
	}

	if !strings.Contains(err.Error(), "failed to read process file") {
		t.Errorf("Expected 'failed to read process file' error, got: %v", err)
	}
}

func TestBuildProjectWithInjection_Integration(t *testing.T) {
	// Skip integration test if not in CI or if Docker is not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create persistent test directories in current working directory
	testBaseDir := "test-output"
	projectDir := filepath.Join(testBaseDir, "test-project")
	workspaceDir := filepath.Join(testBaseDir, "workspace")
	outputDir := filepath.Join(testBaseDir, "output")

	// Clean up any existing test directory
	if err := os.RemoveAll(testBaseDir); err != nil {
		t.Logf("Warning: failed to clean existing test directory: %v", err)
	}

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create a simple test Lua project
	mainLuaContent := `print("successful injection")
local utils = require('.utils')
utils.greet("AOS")
`
	utilsLuaContent := `local utils = {}

function utils.greet(name)
    print("Hello from " .. name .. "!")
end

return utils
`

	// Write main.lua
	mainLuaPath := filepath.Join(projectDir, "main.lua")
	if err := os.WriteFile(mainLuaPath, []byte(mainLuaContent), 0644); err != nil {
		t.Fatalf("Failed to write main.lua: %v", err)
	}

	// Write utils.lua
	utilsLuaPath := filepath.Join(projectDir, "utils.lua")
	if err := os.WriteFile(utilsLuaPath, []byte(utilsLuaContent), 0644); err != nil {
		t.Fatalf("Failed to write utils.lua: %v", err)
	}

	// Create test config
	config := &harlequinConfig.Config{
		StackSize:     8192,
		InitialMemory: 1024,
		MaximumMemory: 2048,
		Target:        32,
		ComputeLimit:  "9000000000",
		ModuleFormat:  "wasm32_unknown_emscripten_metering",
		AOSGitHash:    "main", // Use main branch for testing
	}

	// Create AOSBuilder
	builder := NewAOSBuilder(config, workspaceDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Prepare workspace (clone AOS repo) - use explicit config path since test runs in different directory
	t.Log("Preparing AOS workspace...")
	configPath := "../../build_configs/ao-build-config.yml"
	if err := builder.CopyAOSFiles(ctx, workspaceDir, configPath); err != nil {
		t.Fatalf("Failed to prepare workspace: %v", err)
	}

	// Verify AOS process files were copied
	processDir := filepath.Join(workspaceDir, "aos-process")
	processLuaPath := filepath.Join(processDir, "process.lua")
	if _, err := os.Stat(processLuaPath); os.IsNotExist(err) {
		t.Fatalf("AOS process.lua not found at: %s", processLuaPath)
	}

	// Build project with injection
	t.Log("Building project with injection...")
	if err := builder.BuildProjectWithInjection(ctx, projectDir, outputDir); err != nil {
		t.Fatalf("BuildProjectWithInjection failed: %v", err)
	}

	// Verify outputs exist
	t.Log("Verifying build outputs...")

	// Check that process.wasm was created
	wasmPath := filepath.Join(outputDir, "process.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		t.Errorf("Expected process.wasm not found at: %s", wasmPath)
	} else {
		t.Logf("✅ process.wasm successfully created at: %s", wasmPath)
		
		// Check file size (should be > 0)
		wasmInfo, err := os.Stat(wasmPath)
		if err != nil {
			t.Errorf("Failed to get wasm file info: %v", err)
		} else if wasmInfo.Size() == 0 {
			t.Error("process.wasm file is empty")
		} else {
			t.Logf("✅ process.wasm size: %d bytes", wasmInfo.Size())
		}
	}

	// Check that bundled.lua was copied to output
	bundledPath := filepath.Join(outputDir, "bundled.lua")
	if _, err := os.Stat(bundledPath); os.IsNotExist(err) {
		t.Errorf("Expected bundled.lua not found at: %s", bundledPath)
	} else {
		t.Logf("✅ bundled.lua successfully created at: %s", bundledPath)
		
		// Read and verify bundled content contains our test code
		bundledContent, err := os.ReadFile(bundledPath)
		if err != nil {
			t.Errorf("Failed to read bundled.lua: %v", err)
		} else {
			bundledStr := string(bundledContent)
			if !strings.Contains(bundledStr, "successful injection") {
				t.Error("bundled.lua does not contain expected test content")
			} else {
				t.Log("✅ bundled.lua contains expected test content")
			}
		}
	}

	// Verify that the AOS process.lua was modified with injection
	modifiedProcessContent, err := os.ReadFile(processLuaPath)
	if err != nil {
		t.Errorf("Failed to read modified process.lua: %v", err)
	} else {
		processStr := string(modifiedProcessContent)
		if !strings.Contains(processStr, "require('.bundled')") {
			t.Error("process.lua does not contain injected require statement")
		} else {
			t.Log("✅ process.lua contains injected require statement")
		}
	}

	// Clean up workspace (but keep output directory for verification)
	t.Log("Cleaning up workspace...")
	if err := builder.CleanWorkspace(workspaceDir); err != nil {
		t.Errorf("Failed to clean workspace: %v", err)
	}

	// Get absolute path for output directory
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		absOutputDir = outputDir
	}

	t.Log("🎉 Integration test completed successfully!")
	t.Logf("📁 Build outputs preserved at: %s", absOutputDir)
	t.Logf("🔍 You can verify the following files:")
	t.Logf("   - process.wasm: %s", filepath.Join(absOutputDir, "process.wasm"))
	t.Logf("   - bundled.lua: %s", filepath.Join(absOutputDir, "bundled.lua"))
}

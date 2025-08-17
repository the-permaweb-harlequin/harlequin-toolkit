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

	// Create AOSBuilder with all configuration upfront
	entrypoint := filepath.Join(projectDir, "main.lua")
	configPath := "../../build_configs/ao-build-config.yml"
	builder := newAOSBuilderWithWorkspace(AOSBuilderParams{
		Config:         config,
		ConfigFilePath: configPath,
		Entrypoint:     entrypoint,
		OutputDir:      outputDir,
		Callbacks:      nil, // nil = default callbacks
	}, workspaceDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Use the polished Build method that handles everything
	t.Log("Building AOS project...")
	if err := builder.Build(ctx); err != nil {
		t.Fatalf("Build failed: %v", err)
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

	// Note: The new Build() method automatically cleans up the workspace,
	// so we can't verify the intermediate process.lua file here.
	// That's actually good - it's an implementation detail that shouldn't leak to users.

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

func TestBuildCallbacks(t *testing.T) {
	// Track which callbacks were called
	calledCallbacks := make(map[string]bool)
	var callbackInfos []BuildStepInfo
	
	// Create custom callbacks to track calls
	testCallbacks := &BuildCallbacks{
		OnCopyAOSFiles: func(ctx context.Context, info BuildStepInfo) {
			calledCallbacks["CopyAOSFiles"] = true
			callbackInfos = append(callbackInfos, info)
			t.Logf("✅ OnCopyAOSFiles called: %s (duration: %v, success: %v)", info.StepName, info.Duration, info.Success)
		},
		OnBundleLua: func(ctx context.Context, info BuildStepInfo) {
			calledCallbacks["BundleLua"] = true
			callbackInfos = append(callbackInfos, info)
			t.Logf("✅ OnBundleLua called: %s (duration: %v, success: %v)", info.StepName, info.Duration, info.Success)
		},
		OnInjectLua: func(ctx context.Context, info BuildStepInfo) {
			calledCallbacks["InjectLua"] = true
			callbackInfos = append(callbackInfos, info)
			t.Logf("✅ OnInjectLua called: %s (duration: %v, success: %v)", info.StepName, info.Duration, info.Success)
		},
		OnWasmCompile: func(ctx context.Context, info BuildStepInfo) {
			calledCallbacks["WasmCompile"] = true
			callbackInfos = append(callbackInfos, info)
			t.Logf("✅ OnWasmCompile called: %s (duration: %v, success: %v)", info.StepName, info.Duration, info.Success)
		},
		OnCopyOutputs: func(ctx context.Context, info BuildStepInfo) {
			calledCallbacks["CopyOutputs"] = true
			callbackInfos = append(callbackInfos, info)
			t.Logf("✅ OnCopyOutputs called: %s (duration: %v, success: %v)", info.StepName, info.Duration, info.Success)
		},
		OnCleanup: func(ctx context.Context, info BuildStepInfo) {
			calledCallbacks["Cleanup"] = true
			callbackInfos = append(callbackInfos, info)
			t.Logf("✅ OnCleanup called: %s (duration: %v, success: %v)", info.StepName, info.Duration, info.Success)
		},
	}

	// Set up test directories
	workspaceDir := "test-output/callback-workspace"
	outputDir := "test-output/callback-output"
	projectDir := "test-output/callback-project"

	// Clean up any existing test directories
	os.RemoveAll("test-output/callback-workspace")
	os.RemoveAll("test-output/callback-output")
	os.RemoveAll("test-output/callback-project")

	// Create test project
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project directory: %v", err)
	}

	// Create a simple test Lua file
	testLuaContent := `print("callback test successful")`
	testMainFile := filepath.Join(projectDir, "main.lua")
	if err := os.WriteFile(testMainFile, []byte(testLuaContent), 0644); err != nil {
		t.Fatalf("Failed to create test main.lua: %v", err)
	}

	// Create AOS config
	config := &harlequinConfig.Config{
		ComputeLimit:  "9000000000",
		ModuleFormat:  "wasm32_unknown_emscripten_metering",
		AOSGitHash:    "main",
	}

	// Create builder with custom callbacks
	configPath := "../../build_configs/ao-build-config.yml"
	entrypoint := filepath.Join(projectDir, "main.lua")
	builder := newAOSBuilderWithWorkspace(AOSBuilderParams{
		Config:         config,
		ConfigFilePath: configPath,
		Entrypoint:     entrypoint,
		OutputDir:      outputDir,
		Callbacks:      testCallbacks,
	}, workspaceDir)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run the build
	t.Log("Running build with custom callbacks...")
	err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify all expected callbacks were called
	expectedCallbacks := []string{"CopyAOSFiles", "BundleLua", "InjectLua", "WasmCompile", "CopyOutputs", "Cleanup"}
	for _, expected := range expectedCallbacks {
		if !calledCallbacks[expected] {
			t.Errorf("Expected callback %s was not called", expected)
		}
	}

	// Verify callback info structure
	if len(callbackInfos) != len(expectedCallbacks) {
		t.Errorf("Expected %d callback infos, got %d", len(expectedCallbacks), len(callbackInfos))
	}

	// Verify each callback info has proper fields
	for i, info := range callbackInfos {
		if info.StepName == "" {
			t.Errorf("CallbackInfo[%d] missing StepName", i)
		}
		if info.StartTime.IsZero() {
			t.Errorf("CallbackInfo[%d] missing StartTime", i)
		}
		if info.EndTime.IsZero() {
			t.Errorf("CallbackInfo[%d] missing EndTime", i)
		}
		if info.Duration <= 0 {
			t.Errorf("CallbackInfo[%d] has invalid Duration: %v", i, info.Duration)
		}
		if !info.Success {
			t.Errorf("CallbackInfo[%d] shows step failed: %v", i, info.Error)
		}
	}

	t.Log("🎉 All callbacks verified successfully!")
}

func TestCallbackConstants(t *testing.T) {
	// Test that exported callback constants are not nil and have correct type
	if CallbacksSilent == nil {
		t.Error("CallbacksSilent should not be nil")
	}
	if CallbacksDefault == nil {
		t.Error("CallbacksDefault should not be nil")
	}
	if CallbacksProgress == nil {
		t.Error("CallbacksProgress should not be nil")
	}
	
	// Test that they have all required callback functions
	if CallbacksSilent.OnCopyAOSFiles == nil {
		t.Error("CallbacksSilent.OnCopyAOSFiles should not be nil")
	}
	if CallbacksDefault.OnBundleLua == nil {
		t.Error("CallbacksDefault.OnBundleLua should not be nil")
	}
	if CallbacksProgress.OnWasmCompile == nil {
		t.Error("CallbacksProgress.OnWasmCompile should not be nil")
	}
	
	// Test that NoOpCallbacks returns the same as CallbacksSilent
	noOpCallbacks := NoOpCallbacks()
	if noOpCallbacks != CallbacksSilent {
		t.Error("NoOpCallbacks() should return CallbacksSilent")
	}
	
	// Test that DefaultLoggingCallbacks returns the same as CallbacksDefault
	defaultCallbacks := DefaultLoggingCallbacks()
	if defaultCallbacks != CallbacksDefault {
		t.Error("DefaultLoggingCallbacks() should return CallbacksDefault")
	}
	
	t.Log("✅ All callback constants verified successfully!")
}

func TestAOSBuilderParams(t *testing.T) {
	// Test creating AOSBuilder with params struct
	config := &harlequinConfig.Config{
		ComputeLimit:  "9000000000",
		ModuleFormat:  "wasm32_unknown_emscripten_metering",
		AOSGitHash:    "main",
	}

	params := AOSBuilderParams{
		Config:         config,
		ConfigFilePath: "test-config.yml",
		Entrypoint:     "./main.lua",
		OutputDir:      "./dist",
		Callbacks:      CallbacksProgress,
	}

	builder := NewAOSBuilder(params)

	// Verify all fields were set correctly
	if builder.config != config {
		t.Error("Config not set correctly")
	}
	if builder.configFilePath != params.ConfigFilePath {
		t.Error("ConfigFilePath not set correctly")
	}
	if builder.workspaceDir == "" {
		t.Error("WorkspaceDir should be auto-generated and not empty")
	}
	
	// Verify workspace is in temp directory and has correct prefix
	tempDir := os.TempDir()
	if !strings.HasPrefix(builder.workspaceDir, tempDir) {
		t.Errorf("WorkspaceDir should be in temp directory. Expected prefix %s, got %s", tempDir, builder.workspaceDir)
	}
	if !strings.Contains(builder.workspaceDir, "harlequin-aos-build-") {
		t.Errorf("WorkspaceDir should contain 'harlequin-aos-build-' prefix. Got %s", builder.workspaceDir)
	}
	if builder.entrypoint != params.Entrypoint {
		t.Error("Entrypoint not set correctly")
	}
	if builder.outputDir != params.OutputDir {
		t.Error("OutputDir not set correctly")
	}
	if builder.callbacks != CallbacksProgress {
		t.Error("Callbacks not set correctly")
	}

	// Test with nil callbacks (should default to CallbacksDefault)
	paramsWithNilCallbacks := AOSBuilderParams{
		Config:         config,
		ConfigFilePath: "test-config.yml",
		Entrypoint:     "./main.lua",
		OutputDir:      "./dist",
		Callbacks:      nil,
	}

	builderWithDefaults := NewAOSBuilder(paramsWithNilCallbacks)
	if builderWithDefaults.callbacks != CallbacksDefault {
		t.Error("Nil callbacks should default to CallbacksDefault")
	}

	t.Log("✅ AOSBuilderParams struct verified successfully!")
}

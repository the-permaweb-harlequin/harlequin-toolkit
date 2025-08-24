package luautils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetModFnName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"module.name", "module_name"},
		{"deeply.nested.module", "deeply_nested_module"},
		{"_leading.underscore", "leading_underscore"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, test := range tests {
		result := getModFnName(test.input)
		if result != test.expected {
			t.Errorf("getModFnName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestExploreNodes(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lua-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test Lua files
	testContent := `
local utils = require("utils.helper")
local config = require("config")
print("Hello world")
`
	mainFile := filepath.Join(tempDir, "main.lua")
	if err := os.WriteFile(mainFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test exploreNodes
	node := Module{Path: mainFile}
	modules, err := exploreNodes(node, tempDir)
	if err != nil {
		t.Fatalf("exploreNodes failed: %v", err)
	}

	if len(modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(modules))
	}

	expectedModules := []string{"utils.helper", "config"}
	for i, mod := range modules {
		if mod.Name != expectedModules[i] {
			t.Errorf("Expected module name %q, got %q", expectedModules[i], mod.Name)
		}
	}
}

func TestCreateProjectStructure(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lua-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create utils subdirectory
	utilsDir := filepath.Join(tempDir, "utils")
	if err := os.MkdirAll(utilsDir, 0755); err != nil {
		t.Fatalf("Failed to create utils directory: %v", err)
	}

	// Create test files
	helperContent := `
local function help()
    return "helping"
end
return { help = help }
`
	helperFile := filepath.Join(utilsDir, "helper.lua")
	if err := os.WriteFile(helperFile, []byte(helperContent), 0644); err != nil {
		t.Fatalf("Failed to write helper file: %v", err)
	}

	configContent := `
return { debug = true }
`
	configFile := filepath.Join(tempDir, "config.lua")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	mainContent := `
local utils = require("utils.helper")
local config = require("config")
print("Hello world")
`
	mainFile := filepath.Join(tempDir, "main.lua")
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Test createProjectStructure
	project, err := createProjectStructure(mainFile)
	if err != nil {
		t.Fatalf("createProjectStructure failed: %v", err)
	}

	if len(project) != 3 {
		t.Errorf("Expected 3 modules in project, got %d", len(project))
	}

	// The main file should be last
	mainModule := project[len(project)-1]
	if mainModule.Path != mainFile {
		t.Errorf("Expected main file to be last, got %s", mainModule.Path)
	}
}

func TestBundle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lua-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple helper module
	helperContent := `
local function help()
    return "helping"
end
return { help = help }
`
	helperFile := filepath.Join(tempDir, "helper.lua")
	if err := os.WriteFile(helperFile, []byte(helperContent), 0644); err != nil {
		t.Fatalf("Failed to write helper file: %v", err)
	}

	// Create main file that requires the helper
	mainContent := `
local helper = require("helper")
print(helper.help())
`
	mainFile := filepath.Join(tempDir, "main.lua")
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Test Bundle function
	bundledLua, err := Bundle(mainFile)
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// Check that the bundled code contains expected elements
	if !strings.Contains(bundledLua, "_loaded_mod_helper") {
		t.Error("Bundle should contain helper module function")
	}

	if !strings.Contains(bundledLua, `_G.package.loaded["helper"]`) {
		t.Error("Bundle should contain package.loaded mapping")
	}

	if !strings.Contains(bundledLua, "print(helper.help())") {
		t.Error("Bundle should contain main file content")
	}

	if !strings.Contains(bundledLua, `-- module: "helper"`) {
		t.Error("Bundle should contain module comment")
	}
}

func TestBundleNonExistentFile(t *testing.T) {
	_, err := Bundle("/nonexistent/file.lua")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCreateExecutableFromProjectEmpty(t *testing.T) {
	_, err := createExecutableFromProject([]Module{})
	if err == nil {
		t.Error("Expected error for empty project")
	}
}

func TestFileTraversal(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lua-traversal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a chain of dependencies: main.lua -> utils.lua -> core.lua

	// core.lua (bottom of the chain)
	coreContent := `
local function multiply(a, b)
    return a * b
end

return {
    multiply = multiply
}
`
	coreFile := filepath.Join(tempDir, "core.lua")
	if err := os.WriteFile(coreFile, []byte(coreContent), 0644); err != nil {
		t.Fatalf("Failed to write core file: %v", err)
	}

	// utils.lua (middle of the chain, requires core)
	utilsContent := `
local core = require("core")

local function calculate(x, y)
    return core.multiply(x, y) + 10
end

return {
    calculate = calculate
}
`
	utilsFile := filepath.Join(tempDir, "utils.lua")
	if err := os.WriteFile(utilsFile, []byte(utilsContent), 0644); err != nil {
		t.Fatalf("Failed to write utils file: %v", err)
	}

	// main.lua (top of the chain, requires utils)
	mainContent := `
local utils = require("utils")

print("Result:", utils.calculate(5, 3))
`
	mainFile := filepath.Join(tempDir, "main.lua")
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Test the traversal
	project, err := createProjectStructure(mainFile)
	if err != nil {
		t.Fatalf("createProjectStructure failed: %v", err)
	}

	// Verify we have all three files
	if len(project) != 3 {
		t.Errorf("Expected 3 modules in project, got %d", len(project))
		for i, mod := range project {
			t.Logf("Module %d: %s (%s)", i, mod.Name, mod.Path)
		}
	}

	// Verify the dependency order (core should come before utils, utils before main)
	var coreIndex, utilsIndex, mainIndex int = -1, -1, -1
	for i, mod := range project {
		switch {
		case strings.HasSuffix(mod.Path, "core.lua"):
			coreIndex = i
		case strings.HasSuffix(mod.Path, "utils.lua"):
			utilsIndex = i
		case strings.HasSuffix(mod.Path, "main.lua"):
			mainIndex = i
		}
	}

	if coreIndex == -1 || utilsIndex == -1 || mainIndex == -1 {
		t.Error("Missing expected modules in project")
	}

	// Core should come before utils (dependency order)
	if coreIndex >= utilsIndex {
		t.Errorf("Expected core.lua (index %d) to come before utils.lua (index %d)", coreIndex, utilsIndex)
	}

	// Utils should come before main
	if utilsIndex >= mainIndex {
		t.Errorf("Expected utils.lua (index %d) to come before main.lua (index %d)", utilsIndex, mainIndex)
	}

	// Test the full bundle
	bundledCode, err := Bundle(mainFile)
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// Verify the bundled code contains all expected elements
	expectedElements := []string{
		`-- module: "core"`,
		`_loaded_mod_core`,
		`_G.package.loaded["core"]`,
		`-- module: "utils"`,
		`_loaded_mod_utils`,
		`_G.package.loaded["utils"]`,
		`local utils = require("utils")`,
		`print("Result:", utils.calculate(5, 3))`,
	}

	for _, element := range expectedElements {
		if !strings.Contains(bundledCode, element) {
			t.Errorf("Bundle should contain: %q", element)
		}
	}

	// Verify that core comes before utils in the bundled code
	corePos := strings.Index(bundledCode, `-- module: "core"`)
	utilsPos := strings.Index(bundledCode, `-- module: "utils"`)
	if corePos == -1 || utilsPos == -1 {
		t.Error("Could not find module markers in bundled code")
	} else if corePos >= utilsPos {
		t.Error("Expected core module to appear before utils module in bundled code")
	}
}

func TestCircularDependency(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lua-circular-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create circular dependency: a.lua -> b.lua -> a.lua

	// a.lua requires b.lua
	aContent := `
local b = require("b")

return {
    from_a = "hello from a"
}
`
	aFile := filepath.Join(tempDir, "a.lua")
	if err := os.WriteFile(aFile, []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.lua: %v", err)
	}

	// b.lua requires a.lua (creating a cycle)
	bContent := `
local a = require("a")

return {
    from_b = "hello from b"
}
`
	bFile := filepath.Join(tempDir, "b.lua")
	if err := os.WriteFile(bFile, []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.lua: %v", err)
	}

	// main.lua requires a.lua
	mainContent := `
local a = require("a")
print(a.from_a)
`
	mainFile := filepath.Join(tempDir, "main.lua")
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.lua: %v", err)
	}

	// Test that circular dependency is handled gracefully
	project, err := createProjectStructure(mainFile)
	if err != nil {
		t.Fatalf("createProjectStructure should handle circular dependencies: %v", err)
	}

	// Should have 3 modules (main, a, b) despite the circular dependency
	if len(project) != 3 {
		t.Errorf("Expected 3 modules despite circular dependency, got %d", len(project))
	}

	// Should still be able to bundle
	bundledCode, err := Bundle(mainFile)
	if err != nil {
		t.Fatalf("Bundle should handle circular dependencies: %v", err)
	}

	if bundledCode == "" {
		t.Error("Bundle should produce output even with circular dependencies")
	}
}

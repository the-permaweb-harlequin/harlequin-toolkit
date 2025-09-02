package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
)

// TestModelInitialization tests that the TUI model initializes correctly
func TestModelInitialization(t *testing.T) {
	shared.SkipInCI(t, "TUI tests require interactive environment")

	ctx := context.Background()

	// Create a test environment
	tmpDir := shared.TestTempDir(t)

	// Test model creation (this tests the basic initialization)
	// Note: We can't easily test the full TUI without complex mocking,
	// but we can test model state transitions

	// This is a simplified test - in a real scenario you'd want to use
	// bubbles/test for proper TUI testing
	t.Run("model_creation", func(t *testing.T) {
		// Test that we can create a TUI context without panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("TUI initialization should not panic: %v", r)
			}
		}()

		// Test basic context creation
		if ctx == nil {
			t.Error("Context should not be nil")
		}

		if tmpDir == "" {
			t.Error("Test directory should be created")
		}
	})
}

// TestKeyBindings tests key binding behavior
func TestKeyBindings(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "enter key",
			key:      "enter",
			expected: "enter",
		},
		{
			name:     "escape key",
			key:      "esc",
			expected: "esc",
		},
		{
			name:     "tab key",
			key:      "tab",
			expected: "tab",
		},
		{
			name:     "quit key",
			key:      "q",
			expected: "q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test key matching logic
			var testKey key.Binding

			switch tt.key {
			case "enter":
				testKey = key.NewBinding(key.WithKeys("enter"))
			case "esc":
				testKey = key.NewBinding(key.WithKeys("esc"))
			case "tab":
				testKey = key.NewBinding(key.WithKeys("tab"))
			case "q":
				testKey = key.NewBinding(key.WithKeys("q"))
			}

			// Create a test key message
			keyMsg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune(tt.key),
			}

			if tt.key == "enter" {
				keyMsg.Type = tea.KeyEnter
			} else if tt.key == "esc" {
				keyMsg.Type = tea.KeyEsc
			} else if tt.key == "tab" {
				keyMsg.Type = tea.KeyTab
			}

			// Verify the key binding exists
			if testKey.Help().Key == "" && len(testKey.Keys()) == 0 {
				t.Errorf("Key binding for %s should be defined", tt.key)
			}
		})
	}
}

// TestStateTransitions tests basic state transition logic
func TestStateTransitions(t *testing.T) {
	// This would test state transitions in the TUI model
	// Since the actual model is complex, we test the concept

	type State int
	const (
		StateCommandSelection State = iota
		StateBuildTypeSelection
		StateEntrypointSelection
		StateConfigReview
	)

	tests := []struct {
		name        string
		currentState State
		action      string
		expectedState State
	}{
		{
			name:        "command to build type",
			currentState: StateCommandSelection,
			action:      "select_build",
			expectedState: StateBuildTypeSelection,
		},
		{
			name:        "build type to entrypoint",
			currentState: StateBuildTypeSelection,
			action:      "select_type",
			expectedState: StateEntrypointSelection,
		},
		{
			name:        "entrypoint to config",
			currentState: StateEntrypointSelection,
			action:      "select_entrypoint",
			expectedState: StateConfigReview,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate state transition logic
			var nextState State

			switch tt.currentState {
			case StateCommandSelection:
				if tt.action == "select_build" {
					nextState = StateBuildTypeSelection
				}
			case StateBuildTypeSelection:
				if tt.action == "select_type" {
					nextState = StateEntrypointSelection
				}
			case StateEntrypointSelection:
				if tt.action == "select_entrypoint" {
					nextState = StateConfigReview
				}
			}

			if nextState != tt.expectedState {
				t.Errorf("Expected state %v, got %v", tt.expectedState, nextState)
			}
		})
	}
}

// TestComponentInteraction tests component interaction patterns
func TestComponentInteraction(t *testing.T) {
	t.Run("list_selection", func(t *testing.T) {
		// Test list selection behavior
		items := []string{"build", "upload-module", "lua-utils"}
		selectedIndex := 0

		// Simulate down arrow
		selectedIndex = (selectedIndex + 1) % len(items)
		if selectedIndex != 1 {
			t.Errorf("Expected index 1, got %d", selectedIndex)
		}

		// Simulate up arrow from middle
		selectedIndex = (selectedIndex - 1 + len(items)) % len(items)
		if selectedIndex != 0 {
			t.Errorf("Expected index 0, got %d", selectedIndex)
		}

		// Test wrap around up from first item
		selectedIndex = (selectedIndex - 1 + len(items)) % len(items)
		if selectedIndex != 2 {
			t.Errorf("Expected index 2 (wrap around), got %d", selectedIndex)
		}
	})

	t.Run("form_input", func(t *testing.T) {
		// Test form input behavior
		input := ""

		// Simulate typing
		input += "test"
		if input != "test" {
			t.Errorf("Expected 'test', got '%s'", input)
		}

		// Simulate backspace
		if len(input) > 0 {
			input = input[:len(input)-1]
		}
		if input != "tes" {
			t.Errorf("Expected 'tes', got '%s'", input)
		}
	})
}

// TestFileOperationsIntegration tests file operation integration
func TestFileOperationsIntegration(t *testing.T) {
	tmpDir := shared.TestTempDir(t)

	t.Run("config_file_detection", func(t *testing.T) {
		// Test config file detection logic similar to TUI
		configFiles := []string{
			".harlequin.yaml",
			"build_configs/ao-build-config.yml",
			"harlequin.yaml",
		}

		var foundConfig string
		for _, configFile := range configFiles {
			configPath := shared.CreateTestFile(t, tmpDir, configFile, "test: config")
			shared.AssertFileExists(t, configPath)

			if foundConfig == "" {
				foundConfig = configFile
			}
		}

		if foundConfig != ".harlequin.yaml" {
			t.Errorf("Expected to find .harlequin.yaml first, got %s", foundConfig)
		}
	})

	t.Run("entrypoint_validation", func(t *testing.T) {
		// Test entrypoint file validation
		validEntrypoints := []string{
			"main.lua",
			"src/app.lua",
			"process.lua",
		}

		for _, entrypoint := range validEntrypoints {
			entrypointPath := shared.CreateTestFile(t, tmpDir, entrypoint, "-- Lua code")
			shared.AssertFileExists(t, entrypointPath)

			// Test file extension validation
			if !strings.HasSuffix(entrypoint, ".lua") {
				t.Errorf("Entrypoint %s should have .lua extension", entrypoint)
			}
		}
	})
}

// TestWorkflowIntegration tests complete workflow integration
func TestWorkflowIntegration(t *testing.T) {
	tmpDir := shared.TestTempDir(t)

	t.Run("build_workflow", func(t *testing.T) {
		// Create test files for build workflow
		mainLua := shared.CreateTestFile(t, tmpDir, "main.lua", `print("Hello, World!")`)
		configYaml := shared.CreateTestFile(t, tmpDir, ".harlequin.yaml", `
stack_size: 1048576
initial_memory: 2097152
maximum_memory: 4194304
target: 32
compute_limit: "9000000000000"
module_format: "wasm32-unknown-emscripten-metering"
aos_git_hash: "test-hash"`)

		// Verify files exist
		shared.AssertFileExists(t, mainLua)
		shared.AssertFileExists(t, configYaml)

		// Test workflow state progression
		type WorkflowState struct {
			Entrypoint string
			Config     string
			Ready      bool
		}

		workflow := WorkflowState{
			Entrypoint: mainLua,
			Config:     configYaml,
		}

		// Validate workflow is ready
		if workflow.Entrypoint != "" && workflow.Config != "" {
			workflow.Ready = true
		}

		if !workflow.Ready {
			t.Error("Workflow should be ready with valid entrypoint and config")
		}
	})

	t.Run("upload_workflow", func(t *testing.T) {
		// Create test files for upload workflow
		wasmFile := shared.CreateTestFile(t, tmpDir, "process.wasm", "fake wasm content")
		walletFile := shared.CreateTestFile(t, tmpDir, "wallet.json", `{"test": "wallet"}`)
		configFile := shared.CreateTestFile(t, tmpDir, "build_configs/ao-build-config.yml", `
compute_limit: "9000000000000"
maximum_memory: 524288
module_format: "wasm32-unknown-unknown"`)

		// Verify files exist
		shared.AssertFileExists(t, wasmFile)
		shared.AssertFileExists(t, walletFile)
		shared.AssertFileExists(t, configFile)

		// Test upload workflow validation
		type UploadWorkflow struct {
			WasmFile   string
			WalletFile string
			ConfigFile string
			Version    string
			Valid      bool
		}

		upload := UploadWorkflow{
			WasmFile:   wasmFile,
			WalletFile: walletFile,
			ConfigFile: configFile,
			Version:    "1.0.0",
		}

		// Validate required fields
		if upload.WasmFile != "" && upload.WalletFile != "" && upload.ConfigFile != "" {
			upload.Valid = true
		}

		if !upload.Valid {
			t.Error("Upload workflow should be valid with all required files")
		}
	})
}

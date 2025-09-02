package unit

import (
	"testing"

	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
	"github.com/the-permaweb-harlequin/harlequin-toolkit/cli/tests/shared"
)

func TestNewConfigWithDefaults(t *testing.T) {
	cfg := config.NewConfig(nil)

	// Verify default values
	if cfg.StackSize != config.DefaultStackSize {
		t.Errorf("Expected StackSize %d, got %d", config.DefaultStackSize, cfg.StackSize)
	}

	if cfg.InitialMemory != config.DefaultInitialMemory {
		t.Errorf("Expected InitialMemory %d, got %d", config.DefaultInitialMemory, cfg.InitialMemory)
	}

	if cfg.MaximumMemory != config.DefaultMaximumMemory {
		t.Errorf("Expected MaximumMemory %d, got %d", config.DefaultMaximumMemory, cfg.MaximumMemory)
	}

	if cfg.Target != config.DefaultTarget {
		t.Errorf("Expected Target %d, got %d", config.DefaultTarget, cfg.Target)
	}

	if cfg.ComputeLimit != config.DefaultComputeLimit {
		t.Errorf("Expected ComputeLimit %s, got %s", config.DefaultComputeLimit, cfg.ComputeLimit)
	}

	if cfg.ModuleFormat != config.DefaultModuleFormat {
		t.Errorf("Expected ModuleFormat %s, got %s", config.DefaultModuleFormat, cfg.ModuleFormat)
	}

	if cfg.AOSGitHash != config.DefaultAOSGitHash {
		t.Errorf("Expected AOSGitHash %s, got %s", config.DefaultAOSGitHash, cfg.AOSGitHash)
	}
}

func TestNewConfigWithPartial(t *testing.T) {
	customStackSize := 1024
	customTarget := 64
	customComputeLimit := "8000000000000"

	partial := &config.PartialConfig{
		StackSize:    &customStackSize,
		Target:       &customTarget,
		ComputeLimit: &customComputeLimit,
	}

	cfg := config.NewConfig(partial)

	// Verify custom values
	if cfg.StackSize != customStackSize {
		t.Errorf("Expected StackSize %d, got %d", customStackSize, cfg.StackSize)
	}

	if cfg.Target != customTarget {
		t.Errorf("Expected Target %d, got %d", customTarget, cfg.Target)
	}

	if cfg.ComputeLimit != customComputeLimit {
		t.Errorf("Expected ComputeLimit %s, got %s", customComputeLimit, cfg.ComputeLimit)
	}

	// Verify defaults for unspecified values
	if cfg.InitialMemory != config.DefaultInitialMemory {
		t.Errorf("Expected InitialMemory %d, got %d", config.DefaultInitialMemory, cfg.InitialMemory)
	}

	if cfg.MaximumMemory != config.DefaultMaximumMemory {
		t.Errorf("Expected MaximumMemory %d, got %d", config.DefaultMaximumMemory, cfg.MaximumMemory)
	}
}

func TestYAMLSerialization(t *testing.T) {
	original := config.NewConfig(nil)

	// Convert to YAML
	yamlString := config.ToYAML(original)

	// Verify YAML contains expected fields
	expectedFields := []string{
		"stack_size:",
		"initial_memory:",
		"maximum_memory:",
		"target:",
		"compute_limit:",
		"module_format:",
		"aos_git_hash:",
	}

	for _, field := range expectedFields {
		if !containsSubstring(yamlString, field) {
			t.Errorf("YAML should contain field: %s", field)
		}
	}
}

func TestYAMLDeserialization(t *testing.T) {
	yamlContent := `stack_size: 2048
initial_memory: 8192
maximum_memory: 16384
target: 64
compute_limit: "7000000000000"
module_format: "wasm32-test"
aos_git_hash: "test-hash"`

	cfg := config.FromYAML(yamlContent)

	// Verify deserialized values
	if cfg.StackSize != 2048 {
		t.Errorf("Expected StackSize 2048, got %d", cfg.StackSize)
	}

	if cfg.InitialMemory != 8192 {
		t.Errorf("Expected InitialMemory 8192, got %d", cfg.InitialMemory)
	}

	if cfg.MaximumMemory != 16384 {
		t.Errorf("Expected MaximumMemory 16384, got %d", cfg.MaximumMemory)
	}

	if cfg.Target != 64 {
		t.Errorf("Expected Target 64, got %d", cfg.Target)
	}

	if cfg.ComputeLimit != "7000000000000" {
		t.Errorf("Expected ComputeLimit 7000000000000, got %s", cfg.ComputeLimit)
	}

	if cfg.ModuleFormat != "wasm32-test" {
		t.Errorf("Expected ModuleFormat wasm32-test, got %s", cfg.ModuleFormat)
	}

	if cfg.AOSGitHash != "test-hash" {
		t.Errorf("Expected AOSGitHash test-hash, got %s", cfg.AOSGitHash)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	// Create original config
	customStackSize := 4096
	partial := &config.PartialConfig{
		StackSize: &customStackSize,
	}
	original := config.NewConfig(partial)

	// Convert to YAML and back
	yamlString := config.ToYAML(original)
	roundTrip := config.FromYAML(yamlString)

	// Verify values match
	if original.StackSize != roundTrip.StackSize {
		t.Errorf("StackSize mismatch: %d != %d", original.StackSize, roundTrip.StackSize)
	}

	if original.InitialMemory != roundTrip.InitialMemory {
		t.Errorf("InitialMemory mismatch: %d != %d", original.InitialMemory, roundTrip.InitialMemory)
	}

	if original.Target != roundTrip.Target {
		t.Errorf("Target mismatch: %d != %d", original.Target, roundTrip.Target)
	}

	if original.ComputeLimit != roundTrip.ComputeLimit {
		t.Errorf("ComputeLimit mismatch: %s != %s", original.ComputeLimit, roundTrip.ComputeLimit)
	}
}

func TestConfigFileOperations(t *testing.T) {
	tmpDir := shared.TestTempDir(t)

	// Create test config
	cfg := config.NewConfig(nil)
	configPath := shared.CreateTestFile(t, tmpDir, "test-config.yaml", "")

	// Write config to file
	err := config.WriteConfigFile(cfg, configPath)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Verify file exists
	shared.AssertFileExists(t, configPath)

	// Read config back
	loadedCfg := config.ReadConfigFile(configPath)

	// Verify values match
	if cfg.StackSize != loadedCfg.StackSize {
		t.Errorf("StackSize mismatch after file round trip: %d != %d", cfg.StackSize, loadedCfg.StackSize)
	}

	if cfg.ComputeLimit != loadedCfg.ComputeLimit {
		t.Errorf("ComputeLimit mismatch after file round trip: %s != %s", cfg.ComputeLimit, loadedCfg.ComputeLimit)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		stackSize   int
		target      int
		expectValid bool
	}{
		{
			name:        "valid 32-bit config",
			stackSize:   1024 * 1024,
			target:      32,
			expectValid: true,
		},
		{
			name:        "valid 64-bit config",
			stackSize:   2048 * 1024,
			target:      64,
			expectValid: true,
		},
		{
			name:        "very small stack size",
			stackSize:   1024,
			target:      32,
			expectValid: true, // Should still be valid, just small
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			partial := &config.PartialConfig{
				StackSize: &tt.stackSize,
				Target:    &tt.target,
			}

			cfg := config.NewConfig(partial)

			// Basic validation - config should be created successfully
			if cfg == nil {
				t.Error("Config should not be nil")
			}

			if cfg.StackSize != tt.stackSize {
				t.Errorf("Expected StackSize %d, got %d", tt.stackSize, cfg.StackSize)
			}

			if cfg.Target != tt.target {
				t.Errorf("Expected Target %d, got %d", tt.target, cfg.Target)
			}
		})
	}
}

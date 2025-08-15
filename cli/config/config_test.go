package config

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig(nil)
	if config.StackSize != DefaultStackSize {
		t.Errorf("Expected StackSize to be %d, got %d", DefaultStackSize, config.StackSize)
	}
}

func TestReadConfigFile(t *testing.T) {
	const (
		expectedStackSize = 3145728
		expectedInitialMemory = 4194304
		expectedMaximumMemory = 1073741824
		expectedTarget = 32
		expectedComputeLimit = "9000000000000"
		expectedModuleFormat = "wasm32-unknown-emscripten-metering"
		expectedAOSGitHash = "15dd81ee596518e2f44521e973b8ad1ce3ee9945"
	)
	config := ReadConfigFile("config.test.yaml")
	if config.StackSize != expectedStackSize {
		t.Errorf("Expected StackSize to be %d, got %d", expectedStackSize, config.StackSize)
	}
	if config.InitialMemory != expectedInitialMemory {
		t.Errorf("Expected InitialMemory to be %d, got %d", expectedInitialMemory, config.InitialMemory)
	}
	if config.MaximumMemory != expectedMaximumMemory {
		t.Errorf("Expected MaximumMemory to be %d, got %d", expectedMaximumMemory, config.MaximumMemory)
	}
	if config.Target != expectedTarget {
		t.Errorf("Expected Target to be %d, got %d", expectedTarget, config.Target)
	}
	if config.ComputeLimit != expectedComputeLimit {
		t.Errorf("Expected ComputeLimit to be %s, got %s", expectedComputeLimit, config.ComputeLimit)
	}
	if config.ModuleFormat != expectedModuleFormat {
		t.Errorf("Expected ModuleFormat to be %s, got %s", expectedModuleFormat, config.ModuleFormat)
	}
	if config.AOSGitHash != expectedAOSGitHash {
		t.Errorf("Expected AOSGitHash to be %s, got %s", expectedAOSGitHash, config.AOSGitHash)
	}
}

func TestToYAML(t *testing.T) {
	config := &Config{
		StackSize:     3145728,
		InitialMemory: 4194304,
		MaximumMemory: 1073741824,
		Target:        32,
		ComputeLimit:  "9000000000000",
		ModuleFormat:  "wasm32-unknown-emscripten-metering",
		AOSGitHash:    "15dd81ee596518e2f44521e973b8ad1ce3ee9945",
	}

	yamlString := ToYAML(config)
	
	// Parse it back to verify the YAML is valid and contains expected values
	parsedConfig := FromYAML(yamlString)
	
	if parsedConfig.StackSize != config.StackSize {
		t.Errorf("Expected StackSize to be %d, got %d", config.StackSize, parsedConfig.StackSize)
	}
	if parsedConfig.InitialMemory != config.InitialMemory {
		t.Errorf("Expected InitialMemory to be %d, got %d", config.InitialMemory, parsedConfig.InitialMemory)
	}
	if parsedConfig.MaximumMemory != config.MaximumMemory {
		t.Errorf("Expected MaximumMemory to be %d, got %d", config.MaximumMemory, parsedConfig.MaximumMemory)
	}
	if parsedConfig.Target != config.Target {
		t.Errorf("Expected Target to be %d, got %d", config.Target, parsedConfig.Target)
	}
	if parsedConfig.ComputeLimit != config.ComputeLimit {
		t.Errorf("Expected ComputeLimit to be %s, got %s", config.ComputeLimit, parsedConfig.ComputeLimit)
	}
	if parsedConfig.ModuleFormat != config.ModuleFormat {
		t.Errorf("Expected ModuleFormat to be %s, got %s", config.ModuleFormat, parsedConfig.ModuleFormat)
	}
	if parsedConfig.AOSGitHash != config.AOSGitHash {
		t.Errorf("Expected AOSGitHash to be %s, got %s", config.AOSGitHash, parsedConfig.AOSGitHash)
	}
}

func TestWriteConfigFile(t *testing.T) {
	config := &Config{
		StackSize:     2097152,
		InitialMemory: 8388608,
		MaximumMemory: 2147483648,
		Target:        64,
		ComputeLimit:  "5000000000000",
		ModuleFormat:  "wasm64-custom",
		AOSGitHash:    "abcd1234567890abcd1234567890abcd12345678",
	}

	tempFile := "test_config_write.yaml"
	defer os.Remove(tempFile) // Clean up after test

	// Write the config to file
	err := WriteConfigFile(config, tempFile)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Read it back and verify
	readConfig := ReadConfigFile(tempFile)

	if readConfig.StackSize != config.StackSize {
		t.Errorf("Expected StackSize to be %d, got %d", config.StackSize, readConfig.StackSize)
	}
	if readConfig.InitialMemory != config.InitialMemory {
		t.Errorf("Expected InitialMemory to be %d, got %d", config.InitialMemory, readConfig.InitialMemory)
	}
	if readConfig.MaximumMemory != config.MaximumMemory {
		t.Errorf("Expected MaximumMemory to be %d, got %d", config.MaximumMemory, readConfig.MaximumMemory)
	}
	if readConfig.Target != config.Target {
		t.Errorf("Expected Target to be %d, got %d", config.Target, readConfig.Target)
	}
	if readConfig.ComputeLimit != config.ComputeLimit {
		t.Errorf("Expected ComputeLimit to be %s, got %s", config.ComputeLimit, readConfig.ComputeLimit)
	}
	if readConfig.ModuleFormat != config.ModuleFormat {
		t.Errorf("Expected ModuleFormat to be %s, got %s", config.ModuleFormat, readConfig.ModuleFormat)
	}
	if readConfig.AOSGitHash != config.AOSGitHash {
		t.Errorf("Expected AOSGitHash to be %s, got %s", config.AOSGitHash, readConfig.AOSGitHash)
	}
}
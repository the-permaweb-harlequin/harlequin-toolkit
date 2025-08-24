package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

const (
	DefaultAOSGitHash    = "15dd81ee596518e2f44521e973b8ad1ce3ee9945"
	DefaultComputeLimit  = "9000000000000"
	DefaultModuleFormat  = "wasm32-unknown-emscripten-metering"
	DefaultTarget        = 32
	DefaultStackSize     = 3145728
	DefaultInitialMemory = 4194304
	DefaultMaximumMemory = 1073741824
)

type Config struct {
	StackSize     int    `yaml:"stack_size"`
	InitialMemory int    `yaml:"initial_memory"`
	MaximumMemory int    `yaml:"maximum_memory"`
	Target        int    `yaml:"target"` // 32 or 64
	ComputeLimit  string `yaml:"compute_limit"`
	ModuleFormat  string `yaml:"module_format"`
	AOSGitHash    string `yaml:"aos_git_hash"`
}

type PartialConfig struct {
	StackSize     *int
	InitialMemory *int
	MaximumMemory *int
	Target        *int
	ComputeLimit  *string
}

func NewConfig(partialConfig *PartialConfig) *Config {
	config := &Config{
		StackSize:     DefaultStackSize,
		InitialMemory: DefaultInitialMemory,
		MaximumMemory: DefaultMaximumMemory,
		Target:        DefaultTarget,
		ComputeLimit:  DefaultComputeLimit,
		ModuleFormat:  DefaultModuleFormat,
		AOSGitHash:    DefaultAOSGitHash,
	}

	if partialConfig != nil {
		if partialConfig.StackSize != nil {
			config.StackSize = *partialConfig.StackSize
		}
		if partialConfig.InitialMemory != nil {
			config.InitialMemory = *partialConfig.InitialMemory
		}
		if partialConfig.MaximumMemory != nil {
			config.MaximumMemory = *partialConfig.MaximumMemory
		}
		if partialConfig.Target != nil {
			config.Target = *partialConfig.Target
		}
		if partialConfig.ComputeLimit != nil {
			config.ComputeLimit = *partialConfig.ComputeLimit
		}
	}

	return config
}

func ToYAML(config *Config) string {
	yaml, err := yaml.Marshal(config)
	if err != nil {
		panic(err)
	}
	return string(yaml)
}

func FromYAML(yamlString string) *Config {
	var config Config
	err := yaml.Unmarshal([]byte(yamlString), &config)
	if err != nil {
		panic(err)
	}
	return &config
}

func ReadConfigFile(path string) *Config {
	yamlString, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return FromYAML(string(yamlString))
}

func WriteConfigFile(config *Config, path string) error {
	yamlString := ToYAML(config)
	return os.WriteFile(path, []byte(yamlString), 0644)
}

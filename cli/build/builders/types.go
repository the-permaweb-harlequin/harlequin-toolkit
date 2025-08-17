package builders

import (
	"context"
	"time"

	harlequinConfig "github.com/the-permaweb-harlequin/harlequin-toolkit/cli/config"
)

type Builder interface {
	Build(ctx context.Context, projectPath string) error
	Clean(ctx context.Context, projectPath string) error
	Logs(ctx context.Context, projectPath string) error
	Status(ctx context.Context, projectPath string) error
	Stop(ctx context.Context, projectPath string) error
	Start(ctx context.Context, projectPath string) error
	Restart(ctx context.Context, projectPath string) error
}

// BuildInjectionOptions configures how code injection is performed
type BuildInjectionOptions struct {
	ProcessFilePath string
	BundledCodePath string
	RequireName     string // The name to use in require() statement
}

// AOSBuilderParams contains parameters for creating an AOSBuilder
type AOSBuilderParams struct {
	Config         *harlequinConfig.Config
	ConfigFilePath *string // Optional: defaults to ".harlequin.yaml" if nil
	Entrypoint     string
	OutputDir      string
	Callbacks      *BuildCallbacks
}

// BuildStepInfo contains information about a build step execution
type BuildStepInfo struct {
	StepName    string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Success     bool
	Error       error
	Metadata    map[string]interface{}
}

// BuildCallbacks defines callback functions for each step of the build process
type BuildCallbacks struct {
	OnCopyAOSFiles  func(ctx context.Context, info BuildStepInfo)
	OnBundleLua     func(ctx context.Context, info BuildStepInfo) 
	OnInjectLua     func(ctx context.Context, info BuildStepInfo)
	OnWasmCompile   func(ctx context.Context, info BuildStepInfo)
	OnCopyOutputs   func(ctx context.Context, info BuildStepInfo)
	OnCleanup       func(ctx context.Context, info BuildStepInfo)
}

// NoOpCallbacks returns a BuildCallbacks with no-op functions
func NoOpCallbacks() *BuildCallbacks {
	return CallbacksSilent
}

// Exported callback constants for common configurations
var (
	// CallbacksSilent provides no-op callbacks for silent operation
	CallbacksSilent = &BuildCallbacks{
		OnCopyAOSFiles:  func(ctx context.Context, info BuildStepInfo) {},
		OnBundleLua:     func(ctx context.Context, info BuildStepInfo) {},
		OnInjectLua:     func(ctx context.Context, info BuildStepInfo) {},
		OnWasmCompile:   func(ctx context.Context, info BuildStepInfo) {},
		OnCopyOutputs:   func(ctx context.Context, info BuildStepInfo) {},
		OnCleanup:       func(ctx context.Context, info BuildStepInfo) {},
	}

	// CallbacksDefault provides standard emoji-based logging
	CallbacksDefault = &BuildCallbacks{
		OnCopyAOSFiles: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("🔧 Step 1: Preparing AOS workspace...")
			} else {
				println("❌ Failed to prepare AOS workspace:", info.Error.Error())
			}
		},
		OnBundleLua: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("📦 Step 2: Bundling Lua project...")
			} else {
				println("❌ Failed to bundle Lua project:", info.Error.Error())
			}
		},
		OnInjectLua: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("💉 Step 4: Injecting bundled code into AOS process...")
			} else {
				println("❌ Failed to inject Lua code:", info.Error.Error())
			}
		},
		OnWasmCompile: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("🏗️  Step 5: Building WASM with Docker...")
			} else {
				println("❌ Failed to compile WASM:", info.Error.Error())
			}
		},
		OnCopyOutputs: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("📋 Step 6: Copying build outputs...")
			} else {
				println("❌ Failed to copy outputs:", info.Error.Error())
			}
		},
		OnCleanup: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("🧹 Cleaning up workspace...")
			} else {
				println("❌ Failed to cleanup workspace:", info.Error.Error())
			}
		},
	}

	// CallbacksProgress provides progress logging with timing information
	CallbacksProgress = &BuildCallbacks{
		OnCopyAOSFiles: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("⏱️  Workspace setup completed in", info.Duration.String())
			} else {
				println("❌ Workspace setup failed after", info.Duration.String()+":", info.Error.Error())
			}
		},
		OnBundleLua: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("⏱️  Lua bundling completed in", info.Duration.String())
			} else {
				println("❌ Lua bundling failed after", info.Duration.String()+":", info.Error.Error())
			}
		},
		OnInjectLua: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("⏱️  Code injection completed in", info.Duration.String())
			} else {
				println("❌ Code injection failed after", info.Duration.String()+":", info.Error.Error())
			}
		},
		OnWasmCompile: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("⏱️  WASM compilation completed in", info.Duration.String())
			} else {
				println("❌ WASM compilation failed after", info.Duration.String()+":", info.Error.Error())
			}
		},
		OnCopyOutputs: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("⏱️  Output copying completed in", info.Duration.String())
			} else {
				println("❌ Output copying failed after", info.Duration.String()+":", info.Error.Error())
			}
		},
		OnCleanup: func(ctx context.Context, info BuildStepInfo) {
			if info.Success {
				println("⏱️  Cleanup completed in", info.Duration.String())
			} else {
				println("❌ Cleanup failed after", info.Duration.String()+":", info.Error.Error())
			}
		},
	}
)

// DefaultLoggingCallbacks returns a BuildCallbacks with default logging behavior
func DefaultLoggingCallbacks() *BuildCallbacks {
	return CallbacksDefault
}

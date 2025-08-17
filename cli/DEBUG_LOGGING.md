# 🐛 Debug Logging System

## Overview

The Harlequin CLI now uses a clean debug logging system that dramatically improves the user experience by hiding verbose internal messages unless explicitly requested.

## Before vs After

### ❌ Before (Noisy Output)
```
Starting AOS process copy...
Removing existing directory: /tmp/harlequin-aos-build-1234567890/aos-process
Cloning repository: https://github.com/permaweb/aos.git
Checking out commit: 15dd81ee596518e2f44521e973b8ad1ce3ee9945
Moving /tmp/harlequin-aos-repo/process to /tmp/harlequin-aos-build-1234567890/aos-process
Copying .harlequin.yaml to /tmp/harlequin-aos-build-1234567890/aos-process/config.yml
Successfully copied AOS process and config.
Removing temporary directory: /tmp/harlequin-aos-repo
Injecting bundled code into: /tmp/harlequin-aos-build-1234567890/aos-process/process.lua
Injected require('.bundled') after the last Handlers.append
Successfully injected bundled code require: require('.bundled')
Building WASM module in directory: /tmp/harlequin-aos-build-1234567890/aos-process
Using absolute path for Docker mount: /tmp/harlequin-aos-build-1234567890/aos-process
Docker build completed successfully:
[... 50+ lines of Docker output ...]
✅ WASM module successfully built: /tmp/harlequin-aos-build-1234567890/aos-process/process.wasm
Copied process.wasm to ./dist/process.wasm
Cleaning AOS workspace: /tmp/harlequin-aos-build-1234567890/aos-process
```

### ✅ After (Clean Output)
```
🧪 Testing clean output (debug logging disabled)
🚀 Building with clean output...
✅ Build completed with clean output!
🎉 Debug print statements successfully converted to debug logging!
```

## Debug Mode

When you need detailed logging for troubleshooting, enable debug mode:

### Method 1: Command Line Flag (Recommended)
```bash
./harlequin build --debug
./harlequin build -d          # Short form
```

### Method 2: Environment Variable
```bash
HARLEQUIN_DEBUG=true ./harlequin build
```

### Debug Output
```
🧪 Testing clean output
🚀 Building with clean output...
[DEBUG] Starting AOS process copy...
[DEBUG] Removing existing directory: /tmp/harlequin-aos-build-1234567890/aos-process
[DEBUG] Cloning repository: https://github.com/permaweb/aos.git
[DEBUG] Checking out commit: 15dd81ee596518e2f44521e973b8ad1ce3ee9945
[DEBUG] Moving /tmp/harlequin-aos-repo/process to /tmp/harlequin-aos-build-1234567890/aos-process
[DEBUG] Successfully copied AOS process and config.
[DEBUG] Removing temporary directory: /tmp/harlequin-aos-repo
[DEBUG] Injecting bundled code into: /tmp/harlequin-aos-build-1234567890/aos-process/process.lua
[DEBUG] Injected require('.bundled') after the last Handlers.append
[DEBUG] Successfully injected bundled code require: require('.bundled')
[DEBUG] Building WASM module in directory: /tmp/harlequin-aos-build-1234567890/aos-process
[DEBUG] Using absolute path for Docker mount: /tmp/harlequin-aos-build-1234567890/aos-process
[DEBUG] Docker build completed successfully: [... detailed output ...]
[DEBUG] ✅ WASM module successfully built: /tmp/harlequin-aos-build-1234567890/aos-process/process.wasm
[DEBUG] Copied process.wasm to ./dist/process.wasm
[DEBUG] Cleaning AOS workspace: /tmp/harlequin-aos-build-1234567890/aos-process
✅ Build completed with clean output!
```

## Implementation

### Debug Logger (`cli/debug/logger.go`)
```go
// Printf prints a debug message if debug mode is enabled
func Printf(format string, args ...interface{}) {
    if DebugEnabled {
        fmt.Printf("[DEBUG] "+format, args...)
    }
}

// Info prints an informational message (always shown)
func Info(format string, args ...interface{}) {
    fmt.Printf(format, args...)
}
```

### Environment Variable Control
- **`HARLEQUIN_DEBUG=true`**: Enable debug logging
- **Default**: Debug logging disabled (clean output)

### Converted Functions
All verbose internal logging has been converted:

- `CopyAOSProcess()` - AOS repository cloning and setup
- `InjectBundledCode()` - Lua code injection 
- `buildWithDocker()` - Docker container execution
- `CopyBuildOutputs()` - File copying operations
- `CleanAOSWorkspace()` - Cleanup operations

### User-Facing Messages Preserved
These messages are still shown to users:
- TUI messages (styled with Charm components)
- Build progress from callbacks
- Error messages
- Success notifications

## Benefits

### 🎯 **Better User Experience**
- **Clean Output**: No more wall of technical details
- **Focused Information**: Users see what matters
- **Professional Appearance**: Polished tool experience

### 🐛 **Better Debugging**
- **On-Demand Detail**: Debug info available when needed
- **Clear Labeling**: `[DEBUG]` prefix for debug messages
- **Complete Information**: All original detail preserved

### 🔧 **Better Development**
- **Easy Debugging**: Set environment variable for verbose output
- **No Code Changes**: Toggle debug without rebuilding
- **Consistent System**: All components use same logging approach

## Usage Patterns

### For End Users (Default)
```bash
./harlequin build  # Clean, quiet output
```

### For Developers/Troubleshooting
```bash
./harlequin build --debug               # Verbose debug output (recommended)
./harlequin build -d                    # Short form
HARLEQUIN_DEBUG=true ./harlequin build  # Environment variable method
```

### For Automation/CI
```bash
# Clean output perfect for scripts
./harlequin build

# Or enable debug for CI troubleshooting
./harlequin build --debug
HARLEQUIN_DEBUG=true ./harlequin build  # Alternative method
```

This change transforms Harlequin from a developer tool with noisy output into a professional, user-friendly CLI that can still provide detailed debugging information when needed.

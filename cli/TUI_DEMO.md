# 🎭 Harlequin TUI Demo

## Overview

The Harlequin CLI now includes a beautiful, interactive TUI (Terminal User Interface) built with Charm components. 

## Getting Started

### Launch the TUI
```bash
./harlequin build
```

### Legacy CLI Mode (for scripts/automation)
```bash
./harlequin build ./my-project
```

## TUI Flow Walkthrough

### 1. 🎯 Build Type Selection with Application Layout
The TUI uses a structured application layout with distinct sections:

```
🎭 Harlequin Build Tool

╭──────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                          │
│                              Select Build Configuration                                  │
│                                                                                          │
│  ╭───────────────────────────────────╮  ╭────────────────────────────────────────╮       │
│  │                                   │  │                                        │       │
│  │  Build Types:                     │  │  AOS Flavour                           │       │
│  │                                   │  │                                        │       │
│  │  ❯ AOS Flavour                    │  │  Builds a wasm binary with your Lua    │       │
│  │                                   │  │  injected into the base AOS process    │       │
│  │                                   │  │                                        │       │
│  │                                   │  │                                        │       │
│  │                                   │  │                                        │       │
│  ╰───────────────────────────────────╯  ╰────────────────────────────────────────╯       │
│                                                                                          │
│                    Controls: ↑/↓ Navigate • Enter Select • q Quit                        │
│                                                                                          │
╰──────────────────────────────────────────────────────────────────────────────────────────╯
```

**Layout Structure:**
- **📋 Header**: Shows current view name and context
- **🔍 Left Panel**: Interactive selector with navigation
- **📝 Right Panel**: Live description that updates with selection
- **⌨️ Bottom Controls**: Available keyboard shortcuts and actions

This provides a structured, application-like interface that's familiar and intuitive.

### 2. ⚙️ Build Configuration
Next, choose the build configuration:

```
Select build configuration:
Choose the build configuration for AOS Flavour
❯ Standard build
```

### 3. 📁 Entrypoint Selection
The TUI automatically scans the current directory for `.lua` files:

```
Select entrypoint file:
Choose the main Lua file for your project
❯ test-main.lua
  src/main.lua
  handlers/init.lua
```

Features:
- **Smart Discovery**: Automatically finds all `.lua` files
- **Directory Filtering**: Skips `node_modules`, `.git`, `dist`, etc.
- **Relative Paths**: Shows clean relative paths from current directory

### 4. 📦 Output Directory
Configure where build outputs will be saved:

```
Output directory:
Enter the directory where build outputs will be saved
┌─────────────────────────────────────────┐
│ ./dist                                  │
└─────────────────────────────────────────┘
```

Defaults to `./dist` if no input provided.

### 5. 📋 Configuration Review
The TUI loads your `.harlequin.yaml` config or shows defaults:

```
📄 Loaded existing .harlequin.yaml

Configuration Review:
Current configuration:
  AOS Git Hash: main
  Compute Limit: 9000000000
  Module Format: wasm32_unknown_emscripten_metering
  Target: 64
  Stack Size: 8192
  Initial Memory: 4194304
  Maximum Memory: 134217728

What would you like to do?
❯ Use current configuration
  Edit configuration
```

### 6. ✏️ Configuration Editing (Optional)
If you choose "Edit configuration", you get a multi-page form:

**Page 1: Basic Settings**
```
AOS Git Hash:
Git commit hash or branch name for AOS
┌─────────────────────────────────────────┐
│ main                                    │
└─────────────────────────────────────────┘

Compute Limit:
Maximum compute units for the module
┌─────────────────────────────────────────┐
│ 9000000000                              │
└─────────────────────────────────────────┘
```

**Page 2: Memory & Architecture**
```
Target Architecture:
Target architecture (32 or 64)
┌─────────────────────────────────────────┐
│ 64                                      │
└─────────────────────────────────────────┘

Stack Size:
Stack size in bytes
┌─────────────────────────────────────────┐
│ 8192                                    │
└─────────────────────────────────────────┘
```

### 7. 🚀 Build Execution
Finally, the TUI executes the build with progress callbacks:

```
🚀 Starting Build

🔧 Step 1: Preparing AOS workspace...
📦 Step 2: Bundling Lua project...
💉 Step 4: Injecting bundled code into AOS process...
🏗️  Step 5: Building WASM with Docker...
📋 Step 6: Copying build outputs...
🧹 Cleaning up workspace...

✅ Build completed successfully!
📁 Output directory: ./dist
```

## Technical Features

### 🎨 Styling
- **Consistent Branding**: Purple/violet theme matching Harlequin
- **Clean Layout**: Rounded borders, proper spacing
- **Status Icons**: Visual feedback for each step
- **Error Handling**: Clear error messages with styling

### 🧠 Smart Defaults
- **Config Detection**: Automatically loads `.harlequin.yaml`
- **Fallback Values**: Sensible defaults for all fields
- **Path Resolution**: Handles relative/absolute paths correctly

### 🔧 Integration
- **AOSBuilder**: Seamlessly integrates with existing build system
- **Progress Callbacks**: Real-time build progress feedback
- **Error Recovery**: Graceful error handling and user feedback

### 🚀 Performance
- **Fast File Discovery**: Efficient Lua file scanning
- **Parallel Processing**: Maintains all existing build optimizations
- **Memory Efficient**: Minimal overhead from TUI components

## File Structure
```
cli/
├── main.go              # Updated with TUI integration
├── tui/
│   └── main.go          # Complete TUI implementation
├── harlequin*           # Built CLI executable
└── test-main.lua        # Demo Lua file
```

## Command Comparison

### Traditional CLI (Before)
```bash
# Complex, requires knowledge of flags and paths
harlequin-cli build ./my-project --config ./config.yml --output ./dist
```

### New TUI (After)
```bash
# Simple, guided experience
harlequin build
# Interactive forms guide you through everything!
```

### Legacy Support
```bash
# Still works for automation/scripts
harlequin build ./my-project
```

The TUI makes Harlequin accessible to developers of all experience levels while maintaining the power and flexibility of the original CLI for advanced users and automation.

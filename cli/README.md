# 🎭 Harlequin CLI

A beautiful, interactive terminal interface for building Arweave projects using Charm's Bubble Tea framework.

## Quick Start

### Interactive TUI (Recommended)
```bash
# Build the CLI
go build -o harlequin .

# Launch interactive build experience
./harlequin build
```

### Legacy CLI Mode
```bash
# Direct build (for scripts/automation)
./harlequin build ./my-project
```

## Features

- **🎨 Beautiful TUI**: Clean, intuitive interface with consistent styling
- **📁 Smart File Discovery**: Automatically finds Lua files in your project
- **⚙️ Configuration Management**: Load and edit `.harlequin.yaml` configurations
- **🚀 Real-time Progress**: Live build progress with visual feedback
- **🔧 Error Handling**: Clear error messages and recovery guidance

## Architecture

### Core Components
- **TUI Framework**: Charm Bubble Tea + Huh forms
- **Build System**: Integrated AOSBuilder with Docker containerization
- **Config Management**: YAML-based configuration with smart defaults

### Build Flow
1. **Build Type Selection**: Choose AOS Flavour using structured application layout (more types planned)
2. **Configuration**: Select standard build options
3. **Entrypoint**: Pick your main Lua file
4. **Output**: Configure build output directory
5. **Config Review**: Edit `.harlequin.yaml` if needed
6. **Execution**: Watch real-time build progress

See [TUI_DEMO.md](./TUI_DEMO.md) for detailed walkthrough.

## Roadmap

Plan for this CLI

## Stage 1: flavored aos builds

Allow people to write lua processes with AOS as its wrapper

leverages the existing AO build container

## Stage 2: standard language builds

Allow people to leverage more wasm targeted languages - assemblyscript, c, and rust already have examples.

Provide the basic AOS framework and stdlib for each - ideally published as an installable lib

You should be able to import the AOS tooling and write a process that compiles with either a specific tool, or using the standard build tools with the right configuration.

## Stage 3: Contract templates - tokens, agents, etc with testing for each template.

Build an ecosystem of well tested contracts and add them as template projects.

Mimicking the Go install from url might be a good one here
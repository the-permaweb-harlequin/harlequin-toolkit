# 🎭 Harlequin CLI

A beautiful, interactive terminal interface for building Arweave projects using Charm's Bubble Tea framework.

## Installation

### Quick Install (Recommended)

```bash
# Install latest stable version
curl -sSL https://install_cli_harlequin.daemongate.io | sh

# Install specific version
curl -sSL https://install_cli_harlequin.daemongate.io | VERSION=v0.1.1 sh
```

### Development Build

```bash
# Build from source
go build -o harlequin .
```

## Quick Start

### Interactive TUI (Default)

```bash
# Launch interactive experience (no arguments needed)
harlequin

# Or explicitly launch build TUI
harlequin build
```

### Direct Commands

```bash
# Build project non-interactively
harlequin build --entrypoint main.lua

# List available versions
harlequin versions

# Install specific version
harlequin install --version v0.1.1
```

## Features

- **🎨 Beautiful TUI**: Clean, intuitive interface with consistent styling
- **📁 Smart File Discovery**: Automatically finds Lua files in your project
- **⚙️ Configuration Management**: Load and edit `.harlequin.yaml` configurations
- **🚀 Real-time Progress**: Live build progress with visual feedback
- **🔧 Error Handling**: Clear error messages and recovery guidance
- **📦 Version Management**: Install, upgrade, and manage CLI versions
- **🔍 Version Discovery**: List and explore available versions
- **🛠️ Multiple Tools**: Build, bundle, remote signing, and utilities

## Commands

### Project Initialization

#### `harlequin init`

Create a new AO process project from template.

```bash
# Interactive mode (recommended)
harlequin init
harlequin init my-project

# Non-interactive mode
harlequin init my-project --template rust --author "John Doe"
harlequin init --name my-lua-process --template lua --github johndoe
```

**Available Templates:**

- **lua** - Lua with C trampoline, LuaRocks, embedded interpreter
- **c** - C with Conan, CMake, GTest, Emscripten
- **rust** - Rust with Cargo, wasm-pack, comprehensive testing
- **assemblyscript** - AssemblyScript with custom JSON, Node.js testing

**Options:**

- `-n, --name <NAME>` - Project name
- `-t, --template <TEMPLATE>` - Template language
- `-d, --dir <DIRECTORY>` - Target directory (default: project name)
- `-a, --author <AUTHOR>` - Author name
- `-g, --github <USERNAME>` - GitHub username
- `--non-interactive` - Skip interactive prompts
- `-h, --help` - Show help

### Version Management

#### `harlequin install`

Install or upgrade harlequin to a specific version.

```bash
# Interactive version selection (TUI)
harlequin install

# Install specific version
harlequin install --version v0.1.1
harlequin install --version 0.1.1

# Show help
harlequin install --help
```

#### `harlequin uninstall`

Remove harlequin from your system.

```bash
# Uninstall with confirmation
harlequin uninstall

# Show help
harlequin uninstall --help
```

#### `harlequin versions`

List all available harlequin versions.

```bash
# Default table format
harlequin versions

# JSON format (for scripts)
harlequin versions --format json

# Simple list format
harlequin versions --format list

# Show help
harlequin versions --help
```

### Build Commands

#### `harlequin build`

Build Arweave projects with AOS integration.

```bash
# Interactive TUI (recommended)
harlequin build

# Non-interactive build
harlequin build --entrypoint main.lua
harlequin build --entrypoint main.lua --outputDir ./dist
harlequin build --entrypoint main.lua --debug

# Show help
harlequin build --help
```

#### `harlequin lua-utils`

Lua utilities for bundling and processing.

```bash
# Bundle Lua files
harlequin lua-utils bundle --entrypoint main.lua
harlequin lua-utils bundle --entrypoint main.lua --output bundle.lua

# Show help
harlequin lua-utils --help
```

### Remote Signing

#### `harlequin remote-signing`

Start and manage the remote signing server.

```bash
# Start server (default: localhost:8080)
harlequin remote-signing start

# Start with custom settings
harlequin remote-signing start --port 9000 --host 0.0.0.0

# Check server status
harlequin remote-signing status

# Show help
harlequin remote-signing --help
```

### Information Commands

#### `harlequin version`

Show current CLI version information.

```bash
harlequin version
harlequin --version
harlequin -v
```

#### `harlequin help`

Show comprehensive help information.

```bash
harlequin help
harlequin --help
harlequin -h
```

## Usage Examples

### Getting Started

```bash
# Install harlequin
curl -sSL https://install_cli_harlequin.daemongate.io | sh

# Check available versions
harlequin versions

# Launch interactive TUI
harlequin
```

### Building Projects

```bash
# Interactive build (recommended for new users)
harlequin build

# Quick build for automation
harlequin build --entrypoint src/main.lua --outputDir dist/

# Build with debug information
harlequin build --entrypoint main.lua --debug
```

### Version Management Workflow

```bash
# See what versions are available
harlequin versions --format table

# Install a specific version
harlequin install --version v0.1.1

# Check current version
harlequin version

# Upgrade to latest
harlequin install
```

### Remote Signing Setup

```bash
# Start signing server
harlequin remote-signing start --port 8080

# Check server status
harlequin remote-signing status

# Start with custom frontend URL (development)
harlequin remote-signing start --frontend-url http://localhost:5173
```

## Architecture

### Core Components

- **TUI Framework**: Charm Bubble Tea for interactive experiences
- **Build System**: Integrated AOSBuilder with Docker containerization
- **Config Management**: YAML-based configuration with smart defaults
- **Version Management**: Self-updating CLI with API integration
- **Remote Signing**: WebSocket-based signing server for web interfaces

### Build Flow

1. **Build Type Selection**: Choose AOS Flavour using structured application layout (more types planned)
2. **Configuration**: Select standard build options
3. **Entrypoint**: Pick your main Lua file
4. **Output**: Configure build output directory
5. **Config Review**: Edit `.harlequin.yaml` if needed
6. **Execution**: Watch real-time build progress

See [TUI_DEMO.md](./TUI_DEMO.md) for detailed walkthrough.

## Development

For development instructions, see [DEVELOPMENT.md](./DEVELOPMENT.md).

### Building from Source

```bash
# Clone the repository
git clone https://github.com/the-permaweb-harlequin/harlequin-toolkit.git
cd harlequin-toolkit/cli

# Build the CLI
go build -o harlequin .

# Run tests
go test ./...
```

## Roadmap

### ✅ Completed Features

- **Interactive TUI**: Beautiful terminal interface for all operations
- **AOS Build System**: Lua project building with Docker integration
- **Version Management**: Install, upgrade, and manage CLI versions
- **Remote Signing**: WebSocket-based signing server
- **Lua Utilities**: Bundling and processing tools
- **Multi-format Output**: JSON, table, and list formats for scripting

### 🚧 Current Focus (Stage 1)

- **Enhanced AOS Builds**: Flavored AOS builds with different configurations
- **Template System**: Project templates for common use cases
- **Configuration Management**: Advanced `.harlequin.yaml` options

### 🔮 Future Plans

#### Stage 2: Multi-Language Support

- **WebAssembly Targets**: AssemblyScript, C, Rust support
- **Language-Specific Tooling**: Dedicated build pipelines for each language
- **AOS Framework Libraries**: Installable libraries for each supported language
- **Cross-Language Interop**: Seamless integration between different languages

#### Stage 3: Contract Ecosystem

- **Contract Templates**: Pre-built templates for tokens, agents, DAOs
- **Testing Framework**: Comprehensive testing tools for each template
- **Package Management**: Install contracts from URLs (similar to `go install`)
- **Template Registry**: Curated collection of well-tested contracts

#### Stage 4: Developer Experience

- **IDE Integration**: VS Code extensions and language servers
- **Debugging Tools**: Advanced debugging and profiling capabilities
- **Performance Analytics**: Build-time and runtime performance insights
- **Documentation Generation**: Automatic API documentation from code

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](../CONTRIBUTING.md) for details.

### Quick Contribution Setup

```bash
# Fork and clone the repository
git clone https://github.com/your-username/harlequin-toolkit.git
cd harlequin-toolkit

# Install dependencies
yarn

# Run CLI development build
cd cli && go build -o harlequin .

# Make your changes and test
./harlequin --help
```

## License

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

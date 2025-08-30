# CLI Overview

The Harlequin CLI is a powerful command-line tool for building AO processes and working with Lua files.

## Features

- **Build System**: Compile flavored AOS processes with WebAssembly output
- **Lua Utilities**: Bundle and process Lua files for distribution
- **Version Management**: Install, upgrade, and manage CLI versions
- **Remote Signing**: WebSocket-based signing server for web interfaces
- **Interactive TUI**: Guided workflows with real-time feedback
- **Configuration Management**: Flexible configuration options

## Quick Start

### Interactive Mode (Recommended)

Launch the interactive TUI for guided workflows:

```bash
harlequin
```

### Build an AO Process

```bash
harlequin build --entrypoint main.lua
```

### Bundle Lua Files

```bash
harlequin lua-utils bundle --entrypoint main.lua
```

### Manage Versions

```bash
# List available versions
harlequin versions

# Install specific version
harlequin install --version v0.1.1

# Upgrade to latest
harlequin install
```

## Available Commands

| Command          | Description                                   |
| ---------------- | --------------------------------------------- |
| `build`          | Build AO applications with WebAssembly output |
| `lua-utils`      | Lua file utilities (bundle, etc.)             |
| `install`        | Install or upgrade harlequin                  |
| `uninstall`      | Uninstall harlequin                           |
| `versions`       | List available versions                       |
| `remote-signing` | Start remote signing server                   |
| `version`        | Show version information                      |
| `help`           | Show help information                         |

## Command Categories

### Version Management

- **install** - Install or upgrade harlequin to specific versions
- **uninstall** - Remove harlequin from your system
- **versions** - List all available versions with multiple output formats

### Build Commands

- **build** - Compile AOS processes into WebAssembly binaries
- Support for custom configurations, output directories, and debug modes

### Lua Utils Commands

- **bundle** - Combine multiple Lua files into a single executable
- Automatic dependency resolution and circular dependency handling

### Remote Signing

- **remote-signing** - WebSocket-based signing server for web interfaces
- Support for custom ports, hosts, and frontend URLs

### Interactive TUI

- **Default mode** - Guided workflows for all commands
- File discovery, configuration editing, and real-time progress
- Better development experience than CLI flags

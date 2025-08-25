# CLI Overview

The Harlequin CLI is a powerful command-line tool for building AO processes and working with Lua files.

## Features

- **Build System**: Compile flavored AOS processes with WebAssembly output
- **Lua Utilities**: Bundle and process Lua files for distribution
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

## Available Commands

| Command     | Description                                   |
| ----------- | --------------------------------------------- |
| `build`     | Build AO applications with WebAssembly output |
| `lua-utils` | Lua file utilities (bundle, etc.)             |
| `version`   | Show version information                      |
| `help`      | Show help information                         |

## Command Categories

### Build Commands

- **build** - Compile AOS processes into WebAssembly binaries
- Support for custom configurations, output directories, and debug modes

### Lua Utils Commands

- **bundle** - Combine multiple Lua files into a single executable
- Automatic dependency resolution and circular dependency handling

### Interactive TUI

- **Default mode** - Guided workflows for all commands
- File discovery, configuration editing, and real-time progress
- Better development experience than CLI flags

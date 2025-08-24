# CLI Overview

The Harlequin CLI is a powerful command-line tool for building and deploying applications to the Permaweb.

## Features

- **Build System**: Compile Lua applications for AOS
- **Deployment**: Deploy to Arweave with ease
- **Configuration Management**: Manage project settings
- **Development Tools**: Hot reload and debugging support

## Quick Start

```bash
# Initialize a new project
harlequin init my-project

# Build your application
harlequin build

# Deploy to Arweave
harlequin deploy
```

## Available Commands

| Command  | Description                              |
| -------- | ---------------------------------------- |
| `init`   | Initialize a new Harlequin project       |
| `build`  | Build your application for deployment    |
| `deploy` | Deploy your application to Arweave       |
| `dev`    | Start development server with hot reload |
| `config` | Manage project configuration             |

## Configuration

The CLI uses a `harlequin.yaml` configuration file:

```yaml
# harlequin.yaml
name: my-app
version: 1.0.0
build:
  target: aos
  entry: src/main.lua
deploy:
  network: mainnet
```

## Next Steps

- [Command Reference](/cli/commands) - Detailed command documentation
- [Configuration Guide](/cli/configuration) - Advanced configuration options
- [Examples](/cli/examples) - Sample projects and use cases

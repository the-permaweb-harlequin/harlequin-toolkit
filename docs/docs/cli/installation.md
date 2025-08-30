# CLI Installation

## Quick Install

Install the Harlequin CLI with a single command:

```bash
curl -sSL https://install_cli_harlequin.daemongate.io | sh
```

This installation script will:

- Automatically detect your platform and architecture
- Download the appropriate binary for your system
- Install it to your system PATH
- Verify the installation

## Version-Specific Installation

Install a specific version:

```bash
# Install specific version via environment variable
curl -sSL https://install_cli_harlequin.daemongate.io | VERSION=v0.1.1 sh

# Or use the CLI's built-in version management
harlequin install --version v0.1.1
```

## Prerequisites

- **Operating System**: macOS, Linux, or Windows
- **Architecture**: amd64 (Intel/AMD) or arm64 (Apple Silicon/ARM)
- **Network**: Internet connection for downloading dependencies

## Verify Installation

After installation, verify that the CLI is working correctly:

```bash
# Check version
harlequin --version

# Display help
harlequin --help
```

## Version Management

### Listing Available Versions

See all available versions:

```bash
# Table format (default)
harlequin versions

# JSON format (for scripts)
harlequin versions --format json

# Simple list format
harlequin versions --format list
```

### Updating to Latest Version

```bash
# Interactive version selection
harlequin install

# Or using the install script
curl -sSL https://install_cli_harlequin.daemongate.io | sh
```

### Installing Specific Versions

```bash
# Using CLI version management
harlequin install --version v0.1.1

# Using install script
curl -sSL https://install_cli_harlequin.daemongate.io | VERSION=v0.1.1 sh
```

### Uninstalling

Remove harlequin from your system:

```bash
harlequin uninstall
```

## Support

If you encounter issues:

- Open an issue on [GitHub](https://github.com/the-permaweb-harlequin/harlequin-toolkit/issues)

# CLI Installation

## Quick Install

Install the Harlequin CLI with a single command:

```bash
curl -fsSL https://install_cli_harlequin.daemongate.io | sh
```

This installation script will:

- Automatically detect your platform and architecture
- Download the appropriate binary for your system
- Install it to your system PATH
- Verify the installation

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

## Updating

To update to the latest version:

```bash
# Using the install script (recommended)
curl -sSL https://install_cli_harlequin.daemongate.io | bash
```

## Support

If you encounter issues:

- Open an issue on [GitHub](https://github.com/the-permaweb-harlequin/harlequin-toolkit/issues)

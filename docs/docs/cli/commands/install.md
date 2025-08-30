# install

Install or upgrade harlequin to a specific version.

## Usage

```bash
harlequin install [flags]
```

## Flags

| Flag        | Short | Description                                      |
| ----------- | ----- | ------------------------------------------------ |
| `--version` | `-v`  | Install specific version (e.g., 1.2.3 or v1.2.3) |
| `--help`    | `-h`  | Show help message                                |

## Examples

### Interactive Version Selection

Launch a TUI to select from available versions:

```bash
harlequin install
```

This will:

1. Fetch available versions from the release API
2. Present an interactive list for selection
3. Install the chosen version

### Install Specific Version

Install a specific version directly:

```bash
# With version prefix
harlequin install --version v0.1.1

# Without version prefix
harlequin install --version 0.1.1
```

## How It Works

The install command:

1. **Detects Platform**: Automatically detects your operating system and architecture
2. **Fetches Versions**: Retrieves available versions from `install_cli_harlequin.daemongate.io/releases`
3. **Downloads Binary**: Downloads the appropriate binary for your platform
4. **Installs**: Uses the official installation script to install the binary
5. **Verifies**: Confirms successful installation

## Interactive Mode

When run without `--version`, the command launches an interactive TUI that:

- Shows all available versions in a scrollable list
- Displays version metadata (tag, version number, creation date)
- Allows filtering and selection with keyboard navigation
- Provides installation instructions

## Platform Support

Supported platforms and architectures:

| OS      | Architecture          | Support |
| ------- | --------------------- | ------- |
| macOS   | amd64 (Intel)         | ✅      |
| macOS   | arm64 (Apple Silicon) | ✅      |
| Linux   | amd64                 | ✅      |
| Linux   | arm64                 | ✅      |
| Windows | amd64                 | ✅      |
| Windows | arm64                 | ✅      |

## Error Handling

Common errors and solutions:

### Network Issues

```bash
Error fetching versions: failed to fetch versions: Get "https://...": dial tcp: no such host
```

**Solution**: Check your internet connection and DNS settings.

### Permission Issues

```bash
Error: Failed to install binary to /usr/local/bin/harlequin
```

**Solution**: The installer will automatically use `sudo` when needed for system directories.

### Invalid Version

```bash
Error installing version v999.999.999: installation failed
```

**Solution**: Use `harlequin versions` to see available versions.

## See Also

- [`harlequin versions`](./versions.md) - List available versions
- [`harlequin uninstall`](./uninstall.md) - Remove harlequin
- [`harlequin version`](./version.md) - Show current version

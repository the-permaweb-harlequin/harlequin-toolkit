# uninstall

Remove harlequin from your system.

## Usage

```bash
harlequin uninstall [flags]
```

## Flags

| Flag     | Short | Description       |
| -------- | ----- | ----------------- |
| `--help` | `-h`  | Show help message |

## Examples

### Basic Uninstall

Remove harlequin with confirmation:

```bash
harlequin uninstall
```

This will:

1. Locate the harlequin binary on your system
2. Ask for confirmation before removal
3. Remove the binary from your system
4. Provide reinstallation instructions

## How It Works

The uninstall command:

1. **Locates Binary**: Searches for the harlequin binary in:

   - Current PATH locations
   - `/usr/local/bin/harlequin`
   - `/usr/bin/harlequin`
   - Platform-specific locations (Windows: Program Files)

2. **Confirms Removal**: Asks for user confirmation before proceeding

3. **Removes Binary**: Deletes the binary file from the system

4. **Provides Guidance**: Shows how to reinstall if needed

## Example Output

```bash
$ harlequin uninstall
🎭 Uninstalling Harlequin...
📍 Found harlequin at: /usr/local/bin/harlequin
Are you sure you want to uninstall harlequin from /usr/local/bin/harlequin? [y/N]: y
✅ Successfully uninstalled harlequin from /usr/local/bin/harlequin
💡 To reinstall, run: curl -sSL https://install_cli_harlequin.daemongate.io | sh
```

## Safety Features

### Confirmation Required

The command always asks for confirmation before removing files:

```
Are you sure you want to uninstall harlequin from /usr/local/bin/harlequin? [y/N]:
```

- Type `y` or `yes` to confirm
- Type `n`, `no`, or press Enter to cancel

### Multiple Location Search

If harlequin is not in PATH, the command searches common installation locations:

- `/usr/local/bin/harlequin` (macOS/Linux)
- `/usr/bin/harlequin` (Linux)
- `C:\Program Files\harlequin\harlequin.exe` (Windows)

## Error Handling

Common scenarios and responses:

### Binary Not Found

```bash
$ harlequin uninstall
🎭 Uninstalling Harlequin...
Error during uninstall: harlequin binary not found. It may not be installed or not in PATH
```

**Possible causes:**

- Harlequin is not installed
- Binary is in a non-standard location
- Binary was manually removed

### Permission Issues

```bash
Error during uninstall: failed to remove binary: permission denied
```

**Solution**: The command will automatically handle permission requirements for system directories.

### User Cancellation

```bash
$ harlequin uninstall
🎭 Uninstalling Harlequin...
📍 Found harlequin at: /usr/local/bin/harlequin
Are you sure you want to uninstall harlequin from /usr/local/bin/harlequin? [y/N]: n
Uninstall cancelled.
```

## Reinstallation

After uninstalling, you can reinstall harlequin using:

```bash
# Latest version
curl -sSL https://install_cli_harlequin.daemongate.io | sh

# Specific version
curl -sSL https://install_cli_harlequin.daemongate.io | VERSION=v0.1.1 sh
```

## See Also

- [`harlequin install`](./install.md) - Install or upgrade harlequin
- [`harlequin versions`](./versions.md) - List available versions
- [`harlequin version`](./version.md) - Show current version

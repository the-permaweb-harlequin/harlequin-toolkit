# versions

List all available harlequin versions.

## Usage

```bash
harlequin versions [flags]
```

## Flags

| Flag       | Short | Description                                       |
| ---------- | ----- | ------------------------------------------------- |
| `--format` | `-f`  | Output format: table, json, list (default: table) |
| `--help`   | `-h`  | Show help message                                 |

## Output Formats

### Table Format (Default)

Human-readable table with version details:

```bash
harlequin versions
```

**Example Output:**

```
🎭 Fetching available Harlequin versions...

TAG             VERSION         CREATED
───             ───────         ───────
v0.1.1          0.1.1           2025-08-26 00:55:20

Total: 1 versions available

To install a specific version:
  harlequin install --version <tag>

For interactive selection:
  harlequin install
```

### JSON Format

Machine-readable JSON output for programmatic use:

```bash
harlequin versions --format json
```

**Example Output:**

```json
[
  {
    "tag_name": "v0.1.1",
    "version": "0.1.1",
    "created_at": "2025-08-26T00:55:20.214Z"
  }
]
```

### List Format

Simple list of version tags, perfect for scripting:

```bash
harlequin versions --format list
```

**Example Output:**

```
v0.1.1
```

## Examples

### Basic Usage

Show versions in default table format:

```bash
harlequin versions
```

### Scripting with JSON

Use JSON output in scripts:

```bash
# Get latest version tag
LATEST=$(harlequin versions --format json | jq -r '.[0].tag_name')
echo "Latest version: $LATEST"

# Install latest version
harlequin install --version "$LATEST"
```

### Simple List for Shell Scripts

Get a simple list of versions:

```bash
# Show all versions
harlequin versions --format list

# Get specific version (first/latest)
LATEST=$(harlequin versions --format list | head -1)
```

## Data Source

Version information is fetched from:

- **API Endpoint**: `https://install_cli_harlequin.daemongate.io/releases`
- **Update Frequency**: Real-time (fetched on each command execution)
- **Caching**: No caching (always shows latest available versions)

## Version Information

Each version entry includes:

| Field        | Description        | Example                    |
| ------------ | ------------------ | -------------------------- |
| `tag_name`   | Git tag name       | `v0.1.1`                   |
| `version`    | Semantic version   | `0.1.1`                    |
| `created_at` | Creation timestamp | `2025-08-26T00:55:20.214Z` |

## Sorting

Versions are sorted by creation date, with the newest versions first.

## Error Handling

Common errors and solutions:

### Network Issues

```bash
Error fetching versions: failed to fetch versions: Get "https://...": dial tcp: no such host
```

**Solution**: Check your internet connection and DNS settings.

### API Unavailable

```bash
Error fetching versions: API returned status 503
```

**Solution**: The release API may be temporarily unavailable. Try again later.

### No Versions Found

```bash
🎭 Fetching available Harlequin versions...
No versions found.
```

**Possible causes:**

- API returned empty response
- Network filtering blocking the request
- Temporary API issue

### Invalid Format

```bash
Error: invalid format 'xml'. Valid formats: table, json, list
```

**Solution**: Use one of the supported formats: `table`, `json`, or `list`.

## Integration with Other Commands

The `versions` command works seamlessly with other version management commands:

```bash
# Explore available versions
harlequin versions

# Install a specific version you found
harlequin install --version v0.1.1

# Verify installation
harlequin version
```

## Scripting Examples

### Bash Script: Install Latest Version

```bash
#!/bin/bash
set -e

echo "Fetching latest harlequin version..."
LATEST=$(harlequin versions --format json | jq -r '.[0].tag_name')

if [ "$LATEST" = "null" ] || [ -z "$LATEST" ]; then
    echo "Error: Could not determine latest version"
    exit 1
fi

echo "Installing harlequin $LATEST..."
harlequin install --version "$LATEST"
```

### PowerShell Script: Version Comparison

```powershell
# Get current and latest versions
$current = harlequin version | Select-String "version (\S+)" | ForEach-Object { $_.Matches[0].Groups[1].Value }
$latest = harlequin versions --format json | ConvertFrom-Json | Select-Object -First 1 -ExpandProperty version

if ($current -ne $latest) {
    Write-Host "Update available: $current -> $latest"
    Write-Host "Run: harlequin install --version v$latest"
} else {
    Write-Host "You have the latest version: $current"
}
```

## See Also

- [`harlequin install`](./install.md) - Install specific versions
- [`harlequin version`](./version.md) - Show current version
- [`harlequin uninstall`](./uninstall.md) - Remove harlequin

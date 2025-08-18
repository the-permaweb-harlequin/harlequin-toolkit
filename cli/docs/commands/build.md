# 🎭 Harlequin Build Command

Build your Arweave projects with the interactive TUI or legacy CLI mode.

## Usage

```bash
harlequin build [flags] [path]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--debug` | `-d` | Enable debug logging for detailed output |
| `--help` | `-h` | Show this help message |

## Arguments

- **`[path]`** - Project path (optional, defaults to interactive TUI)

## Examples

### Interactive TUI Mode (Recommended)
```bash
# Launch the interactive TUI
harlequin build

# TUI with debug logging enabled
harlequin build --debug
```

### Legacy CLI Mode
```bash
# Direct build with project path
harlequin build ./my-project

# Direct build with debug logging
harlequin build -d ./my-project
```

## Interactive TUI Features

The interactive TUI provides a guided experience including:

- **🎯 Build Type Selection** - Choose AOS Flavour builds
- **📄 Entrypoint Selection** - Pick your main Lua file from discovered files
- **📁 Output Configuration** - Set your build output directory
- **⚙️ Configuration Editing** - Edit `.harlequin.yaml` settings with live preview
- **📊 Real-time Build Progress** - Watch build steps with animated progress indicators
- **✅ Success/Error Screens** - Clear feedback with detailed configuration summary

## Configuration

The build command looks for configuration in this order:

1. `harlequin.yaml` in current directory
2. `build_configs/ao-build-config.yml`
3. Default configuration values

### Configuration Options

- **Target**: WASM 32-bit or 64-bit compilation
- **Memory Settings**: Stack size, initial memory, maximum memory (in MB)
- **AOS Integration**: Git hash for AOS version
- **Build Options**: Compute limits and module format

## Debug Mode

When `--debug` is enabled, you'll see detailed logging including:

- 🔄 Git repository cloning progress
- 🐳 Docker build container output  
- 📦 File copying and injection details
- 🧹 Cleanup operations
- 📝 Full configuration dump on errors

### Environment Variable Alternative

```bash
HARLEQUIN_DEBUG=true harlequin build
```

## Build Process

1. **Discovery** - Scan for Lua entrypoint files
2. **Configuration** - Load or create build configuration
3. **AOS Setup** - Clone and prepare AOS repository
4. **Bundling** - Combine your Lua code
5. **Injection** - Inject code into AOS process
6. **Compilation** - Build WASM binary with Docker
7. **Output** - Copy results to output directory
8. **Cleanup** - Clean temporary files

## Output Structure

```
./dist/               # Default output directory
├── process.wasm      # Compiled WASM binary
├── bundled.lua       # Combined Lua code
└── config.json       # Build configuration used
```

## Troubleshooting

- **No Lua files found**: Ensure you have `.lua` files in your project
- **Docker errors**: Make sure Docker is running and accessible
- **Memory issues**: Adjust memory settings in configuration
- **Permission errors**: Check file/directory permissions

For detailed error information, always run with `--debug` flag and check the debug logs at `~/.harlequin/harlequin-debug.log`.

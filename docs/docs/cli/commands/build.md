# Build Command

The build command allows you to build AOS projects either interactively through the TUI or non-interactively via command-line flags.

## Interactive Build (Recommended)

For the best development experience, use the interactive TUI:

```bash
harlequin
```

This launches the interactive interface where you can:

1. **Select Command** - Choose "Build Project" from the welcome screen
2. **Choose Build Type** - Select AOS Flavour configuration
3. **Select Entrypoint** - Pick your main Lua file (auto-discovery or manual selection)
4. **Configure Output** - Set the output directory for build artifacts
5. **Review Configuration** - Examine and optionally edit build settings
6. **Monitor Progress** - Watch real-time build progress with detailed steps

## Non-Interactive Build

For automation, CI/CD, or when you know exactly what you want to build:

### Syntax

```bash
harlequin build --entrypoint <file> [flags]
```

### Required Flags

- `--entrypoint <file>` - Path to the main Lua file to build

### Optional Flags

- `--outputDir <dir>` - Directory to output build artifacts (default: uses config settings)
- `--configPath <file>` - Path to custom configuration file (default: searches for .harlequin.yaml or build_configs/ao-build-config.yml)
- `-d, --debug` - Enable debug logging for detailed output
- `-h, --help` - Show help message

## Examples

### Basic Build

```bash
harlequin build --entrypoint main.lua
```

### Build with Custom Output Directory

```bash
harlequin build --entrypoint src/app.lua --outputDir dist
```

### Build with Custom Configuration

```bash
harlequin build --entrypoint main.lua --configPath custom.yaml
```

### Build with Debug Output

```bash
harlequin build --entrypoint main.lua --debug
```

### Complete Example

```bash
harlequin build --entrypoint src/main.lua --outputDir build/artifacts --configPath configs/production.yaml --debug
```

## Build Process

Both interactive and non-interactive builds follow the same process:

1. **Load Configuration** - Reads build settings from config files
2. **Prepare Environment** - Sets up the build environment and dependencies
3. **Copy AOS Files** - Copies base AOS runtime files
4. **Bundle Lua** - Processes and bundles your Lua code
5. **Inject Code** - Injects your code into the AOS runtime
6. **Build WASM** - Compiles the final WebAssembly binary
7. **Copy Outputs** - Places build artifacts in the specified output directory
8. **Cleanup** - Removes temporary build files

## Output Artifacts

After a successful build, you'll find these files in your output directory:

- `process.wasm` - The compiled WebAssembly binary
- `bundled.lua` - Your processed Lua code bundle
- `config.yml` - The build configuration used

## Debug Mode

When using `--debug`, you'll see detailed logging including:

- Git repository cloning progress
- Docker build container output
- File copying and injection details
- Cleanup operations
- Error details and stack traces

## Environment Variables

You can also enable debug mode via environment variable:

```bash
HARLEQUIN_DEBUG=true harlequin build --entrypoint main.lua
```

## Common Patterns

### Development Workflow

Use interactive mode during development for the best experience:

```bash
harlequin
```

### CI/CD Pipeline

Use non-interactive mode in automated environments:

```bash
harlequin build --entrypoint src/main.lua --outputDir dist --debug
```

### Local Testing

Quick builds with debug output:

```bash
harlequin build --entrypoint main.lua --debug
```

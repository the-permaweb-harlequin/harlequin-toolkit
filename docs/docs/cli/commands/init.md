# Init Command

The init command helps you create new AO process projects from pre-built templates. It supports both interactive and non-interactive modes for different workflows.

## Interactive Init (Recommended)

For the best project creation experience, use the interactive TUI:

```bash
harlequin init
```

This launches a beautiful step-by-step wizard that guides you through:

1. **Project Name** - Enter a name for your new project
2. **Template Selection** - Choose from available templates with live feature previews
3. **Author Information** - Optionally provide your name and GitHub username
4. **Confirmation** - Review your choices before creating the project

The interactive mode provides:

- **Visual template selection** with real-time feature descriptions
- **Input validation** to ensure valid project names
- **Consistent styling** matching the Harlequin theme
- **Keyboard navigation** with intuitive controls
- **Graceful cancellation** with Esc or Ctrl+C

## Non-Interactive Init

For automation, CI/CD, or when you know exactly what you want to create:

### Syntax

```bash
harlequin init <LANGUAGE> [flags]
```

### Required Arguments

- `<LANGUAGE>` - Template language (lua, c, rust, assemblyscript)

### Optional Flags

- `--name, -n <NAME>` - Project name (required in non-interactive mode)
- `--dir, -d <DIRECTORY>` - Target directory (default: project name)
- `--author, -a <AUTHOR>` - Author name for project metadata
- `--github, -g <USERNAME>` - GitHub username for project metadata
- `--interactive` - Force interactive mode even with language specified
- `--non-interactive` - Skip interactive prompts (default when language provided)
- `-h, --help` - Show help message

### Backward Compatibility

The legacy `--template` flag is still supported:

```bash
harlequin init --template <LANGUAGE> --name <NAME> [flags]
```

## Available Templates

### Lua Template

- **Language**: Lua with C trampoline
- **Build System**: CMake + LuaRocks
- **Features**:
  - C trampoline with embedded Lua interpreter
  - LuaRocks package management
  - WebAssembly compilation with Emscripten
  - Modular architecture with handlers and utils
  - Comprehensive testing with Busted

### C Template

- **Language**: C
- **Build System**: CMake + Conan
- **Features**:
  - Conan package management
  - Google Test integration
  - Emscripten WebAssembly compilation
  - Memory-efficient implementation
  - Docker build support

### Rust Template

- **Language**: Rust
- **Build System**: Cargo + wasm-pack
- **Features**:
  - Thread-safe state management
  - Serde JSON serialization
  - wasm-bindgen WebAssembly bindings
  - Comprehensive error handling
  - Size-optimized builds

### AssemblyScript Template

- **Language**: AssemblyScript
- **Build System**: AssemblyScript Compiler
- **Features**:
  - TypeScript-like syntax
  - Custom JSON handling
  - Memory-safe operations
  - Node.js testing framework
  - Size optimization

## Examples

### Interactive Mode

```bash
# Launch interactive wizard
harlequin init
```

### Non-Interactive Mode

```bash
# Create a Lua project
harlequin init lua --name my-ao-process --author "John Doe"

# Create a Rust project with GitHub info
harlequin init rust --name my-rust-process --github johndoe

# Create a C project in specific directory
harlequin init c --name my-c-project --dir ./projects/my-c-project

# Create AssemblyScript project with all metadata
harlequin init assemblyscript --name my-as-project --author "Alice Smith" --github alicesmith
```

### Backward Compatibility

```bash
# Legacy template flag (still works)
harlequin init --template lua --name my-project --author "Developer"

# Force interactive mode with pre-filled template
harlequin init lua --interactive --name my-project
```

## Project Structure

After successful creation, each template generates a complete project structure:

### Common Files

- `package.json` - npm scripts for building and testing
- `README.md` - Template-specific documentation and setup instructions
- `.harlequin.yaml` - Harlequin build configuration

### Template-Specific Files

**Lua Projects:**

- `main.lua` - Main process entry point
- `handlers/` - Message handlers directory
- `utils/` - Utility functions directory
- `src/trampoline.c` - C trampoline implementation
- `CMakeLists.txt` - CMake build configuration
- `*.rockspec` - LuaRocks package specification
- `test/` - Lua test files

**C Projects:**

- `main.c` - Main C source file
- `CMakeLists.txt` - CMake build configuration
- `conanfile.txt` - Conan dependencies
- `test/` - Google Test files

**Rust Projects:**

- `src/main.rs` - Main Rust source
- `src/lib.rs` - Library implementation
- `Cargo.toml` - Rust package configuration
- `tests/` - Integration tests

**AssemblyScript Projects:**

- `assembly/index.ts` - Main AssemblyScript source
- `assembly/json.ts` - JSON handling utilities
- `asconfig.json` - AssemblyScript configuration
- `tests/` - Node.js test files

## Next Steps

After creating a project, the init command provides template-specific next steps:

### For Lua Projects

```bash
cd your-project
npm run setup          # Install LuaRocks dependencies
npm run build          # Build with CMake
npm run build:wasm     # Build WebAssembly with Emscripten
npm test               # Run Lua tests
npm run test:trampoline # Test C trampoline
```

### For C Projects

```bash
cd your-project
npm run setup          # Install Conan dependencies
npm run build:cmake    # Build with CMake
npm test               # Run tests
npm run docker:build   # Build in Docker
```

### For Rust Projects

```bash
cd your-project
npm run setup          # Install Rust toolchain
npm run build          # Build native binary
npm run build:wasm     # Build WebAssembly
npm test               # Run tests
npm run run            # Test locally
```

### For AssemblyScript Projects

```bash
cd your-project
npm install            # Install dependencies
npm run build          # Build WebAssembly
npm test               # Run tests
npm run optimize       # Optimize binary
```

### Universal Commands

```bash
harlequin build         # Build with Harlequin CLI
harlequin               # Launch interactive TUI
```

## Variable Substitution

The init command automatically substitutes template variables:

- `{{PROJECT_NAME}}` - Replaced with your project name
- `{{AUTHOR_NAME}}` - Replaced with author name (or "Your Name" if not provided)
- `{{GITHUB_USER}}` - Replaced with GitHub username (or "your-username" if not provided)

This ensures all generated files contain your project-specific information.

## Common Patterns

### Development Workflow

Use interactive mode for exploring templates and creating new projects:

```bash
harlequin init
```

### CI/CD and Automation

Use non-interactive mode in scripts and automated environments:

```bash
harlequin init lua --name "$PROJECT_NAME" --author "$AUTHOR" --github "$GITHUB_USER"
```

### Batch Project Creation

Create multiple projects with different templates:

```bash
harlequin init lua --name my-lua-project --author "Developer"
harlequin init rust --name my-rust-project --author "Developer"
harlequin init c --name my-c-project --author "Developer"
```

### Custom Directory Structure

Organize projects in specific directories:

```bash
harlequin init lua --name backend-service --dir ./services/backend
harlequin init rust --name frontend-wasm --dir ./frontend/wasm
```

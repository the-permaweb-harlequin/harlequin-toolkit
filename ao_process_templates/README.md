# AO Process Templates

Starter templates for AO processes in multiple programming languages, each with their respective compilation utilities and build systems.

## Structure

```
ao_process_templates/
├── languages/           # 📂 Source templates and build tools
│   ├── assemblyscript/
│   │   └── template/    # Project template files with local transformer
│   └── go/
│       └── template/    # Project template files
├── scripts/             # 🔧 Build automation
│   ├── build-create-packages.js  # Generate NPM packages
│   └── build-cli-templates.js    # Generate CLI templates
├── create-packages/     # 📦 Generated NPM packages (build output)
│   ├── create-ao-assemblyscript/  # npx create-ao-assemblyscript
│   └── create-ao-go/              # npx create-ao-go
├── cli-templates/       # 📦 Generated CLI templates (build output)
│   ├── assemblyscript.tar.gz     # For Harlequin CLI
│   ├── go.tar.gz                 # For Harlequin CLI
│   └── templates.json            # Manifest
├── testing/            # 🧪 Centralized testing infrastructure
│   ├── assemblyscript/ # AssemblyScript tests
│   └── go/             # Go tests (planned)
└── package.json        # Root workspace configuration
```

## Configuration

This workspace supports **dual distribution** - both NPM and Harlequin CLI:

- **📂 Source templates** in `languages/` - edit these to modify templates
- **🔧 Build scripts** automatically generate distribution packages
- **📦 NPM packages** in `create-packages/` for `npx create-*`
- **📦 CLI templates** in `cli-templates/` for Harlequin CLI integration
- **🧪 Centralized testing** infrastructure
- **🔄 Nx integration** for monorepo management

## 🚀 Quick Start

Create a new AO process with one command:

```bash
# Create AssemblyScript AO process
npx create-ao-assemblyscript my-ao-process

# Create Go AO process
npx create-ao-go my-ao-process
```

## Available Templates

### AssemblyScript (`create-ao-assemblyscript`)

- 🎭 **Quick start**: `npx create-ao-assemblyscript my-project`
- Full AssemblyScript to WASM compilation
- Emscripten compatibility transform
- TypeScript tooling and JSON support
- Built-in Hello, Info, Echo actions

### Go (`create-ao-go`)

- 🐹 **Quick start**: `npx create-ao-go my-project`
- Go to WASM compilation with Makefile
- Native Go standard library support
- Process info and echo functionality
- Comprehensive test suite

## Usage

### Development

```bash
# Install all dependencies
pnpm install

# Build all templates
pnpm run build

# Run all tests
pnpm run test

# Clean all build artifacts
pnpm run clean
```

### Individual Template Development

```bash
# Work on AssemblyScript template
cd languages/assemblyscript
pnpm run build

# Work on Go template
cd languages/go
pnpm run build
```

## Testing

The `testing/` directory contains centralized test infrastructure:

- Tests moved from individual language folders
- Shared test utilities and fixtures
- Integrated with AO loader for realistic testing

```bash
# Run all tests
cd testing
pnpm run test

# Run specific language tests
pnpm run test:assemblyscript
pnpm run test:go
```

## Publishing

To make templates available via npx:

```bash
# Publish create packages to npm
cd create-packages/create-ao-assemblyscript
npm publish

cd ../create-ao-go
npm publish
```

Users can then run:

```bash
npx create-ao-assemblyscript my-project
npx create-ao-go my-project
```

## Nx Integration

This workspace is integrated with the main Harlequin Toolkit Nx monorepo:

- Each create package and language tool is a discoverable Nx project
- Shared build, test, and lint targets
- Dependency graph management
- Parallel execution support

## Development Workflow

### 🎯 Edit Templates

Edit source templates in `languages/*/template/`:

```bash
# Edit AssemblyScript template
vi languages/assemblyscript/template/assembly/index.ts

# Edit Go template
vi languages/go/template/main.go
```

### 🔧 Build Distribution Packages

```bash
# Build everything (transforms + NPM packages + CLI templates)
pnpm run build

# Build specific targets
pnpm run build:create-packages  # NPM packages only
pnpm run build:cli-templates    # CLI templates only
```

### 🧪 Test Templates

```bash
# Test create packages locally
cd create-packages/create-ao-assemblyscript
node bin/create-ao-assemblyscript.js test-project

# Run centralized tests
cd testing
pnpm run test:assemblyscript
```

### 📦 Publish

```bash
# Publish to NPM
pnpm run publish:npm

# CLI templates are built in cli-templates/
# Copy these to Harlequin CLI project
```

## For Harlequin CLI Integration

The `cli-templates/` directory contains everything needed for CLI integration:

- **`templates.json`** - Master manifest with all available templates
- **`assemblyscript.tar.gz`** - AssemblyScript template tarball
- **`go.tar.gz`** - Go template tarball
- **Individual `.json` files** - Per-template metadata

Copy these files to your Harlequin CLI project and use the manifest to list available templates.

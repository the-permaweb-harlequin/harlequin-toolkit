# Harlequin CLI Development

## Prerequisites

### Required Tools
- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Node.js 18+** - For Nx tooling
- **GoReleaser** - For release builds

### Installing GoReleaser

#### Option 1: Using Go (Recommended)
```bash
# Install latest GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Verify installation
goreleaser --version
```

#### Option 2: Using Homebrew (macOS/Linux)
```bash
brew install goreleaser
```

#### Option 3: Using the Nx setup target
```bash
# Installs GoReleaser via Go
npx nx setup cli
```

## Development Commands

### Setup
```bash
# Install GoReleaser
npx nx setup cli

# Validate configuration
npx nx goreleaser-check cli
```

### Building
```bash
# Local development build
npx nx build cli

# Multi-platform snapshot build
npx nx build cli --configuration=production

# Test GoReleaser build
npx nx release cli --configuration=snapshot
```

### Testing
```bash
# Run tests
npx nx test cli

# Lint code
npx nx lint cli

# Clean up Go modules
npx nx tidy cli
```

### Releases
```bash
# Dry run (see what would happen)
npx nx release cli --configuration=dry-run

# Snapshot release (no git tag needed)
npx nx release cli --configuration=snapshot

# Real release (requires git tag cli-v*)
git tag cli-v1.2.3
git push origin cli-v1.2.3
# GitHub Actions will handle the release
```

## CI/CD

The GitHub Actions workflow automatically:
1. **Installs GoReleaser** using the official action
2. **Validates configuration** with `goreleaser check`
3. **Builds binaries** for all platforms
4. **Creates GitHub release** with assets
5. **Deploys to Arweave** via Turbo SDK

## GoReleaser Configuration

The GoReleaser config is located at `cli/.goreleaser.yaml` and handles:
- Multi-platform builds (Linux, macOS, Windows × AMD64/ARM64)
- Archive creation (tar.gz/zip)
- Checksum generation
- GitHub release creation
- Custom Arweave deployment publisher

## Troubleshooting

### GoReleaser Not Found
```bash
# Check if GoReleaser is in PATH
which goreleaser

# Install using Nx
npx nx setup cli

# Or install manually
go install github.com/goreleaser/goreleaser@latest
```

### Build Issues
```bash
# Clean and rebuild
go clean
npx nx build cli

# Check Go modules
npx nx tidy cli
```

### Release Issues
```bash
# Validate configuration
npx nx goreleaser-check cli

# Test with dry run
npx nx release cli --configuration=dry-run
```

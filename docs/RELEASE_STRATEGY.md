# 🚀 Harlequin Toolkit Release Strategy

## Overview

This monorepo uses Nx for selective builds and releases with a sophisticated tagging strategy that allows independent versioning of each project.

## 📦 Projects

- **CLI** (`cli/`) - Go binary application
- **SDK** (`sdk/`) - TypeScript/JavaScript library
- **App** (`app/`) - Frontend application  
- **Server** (`server/`) - Backend service

## 🏷️ Tagging Strategy

### Project-Specific Tags
```bash
# Individual project releases
git tag cli-v1.2.3     # CLI release
git tag sdk-v2.1.0     # SDK release
git tag app-v1.0.5     # App release
git tag server-v1.1.2  # Server release
```

### Monorepo Tags
```bash
# Full monorepo release (all projects)
git tag v1.0.0
```

## 🔄 Release Workflow

### 1. CLI Releases
```bash
# Tag for CLI release
git tag cli-v1.2.3
git push origin cli-v1.2.3
```

**What happens:**
- Builds Go binaries for all platforms
- Creates GitHub release with binaries
- Updates `install_cli_harlequin.daemongate.io/releases` API
- Uploads binaries to hosting service

### 2. SDK Releases  
```bash
# Tag for SDK release
git tag sdk-v2.1.0
git push origin sdk-v2.1.0
```

**What happens:**
- Builds TypeScript library
- Publishes to npm registry
- Updates package version

### 3. App/Server Releases
```bash
# Tag for app release
git tag app-v1.0.5
git push origin app-v1.0.5

# Tag for server release  
git tag server-v1.1.2
git push origin server-v1.1.2
```

**What happens:**
- Builds and deploys applications
- Updates hosting environments

### 4. Monorepo Releases
```bash
# Tag for full release
git tag v1.0.0
git push origin v1.0.0
```

**What happens:**
- Releases ALL projects simultaneously
- Useful for major version bumps

## 🛠️ Development Workflow

### Daily Development
```bash
# Work on any project
git checkout -b feature/my-feature

# Commit changes
git add .
git commit -m "feat(cli): add new command"

# Push and create PR
git push origin feature/my-feature
```

### CI/CD Checks
- **Affected builds only** - Nx determines what changed
- **Parallel execution** - Multiple projects build simultaneously  
- **Cross-platform testing** - Go binaries tested on multiple platforms
- **Linting and testing** - Only runs for affected projects

### Release Process
```bash
# 1. Ensure main is up to date
git checkout main
git pull origin main

# 2. Create release branch (optional)
git checkout -b release/cli-v1.2.3

# 3. Update version files if needed
# - cli/main.go version constant
# - sdk/package.json version
# etc.

# 4. Commit version updates
git add .
git commit -m "chore: bump CLI version to 1.2.3"

# 5. Tag the release
git tag cli-v1.2.3

# 6. Push tag to trigger release
git push origin cli-v1.2.3
```

## 📋 Installation Script Integration

The CLI install script integrates with releases:

### Automatic Updates
```bash
# Install script fetches from: 
# https://install_cli_harlequin.daemongate.io/releases

# API returns:
{
  "releases": [
    {
      "tag_name": "cli-v1.2.3",
      "version": "1.2.3",
      "assets": [
        {
          "name": "harlequin-linux-amd64",
          "url": "https://install_cli_harlequin.daemongate.io/releases/1.2.3/linux/amd64"
        }
        // ... more platforms
      ]
    }
  ]
}
```

### Binary Hosting
```bash
# Binaries hosted at:
https://install_cli_harlequin.daemongate.io/releases/{version}/{platform}/{arch}

# Examples:
https://install_cli_harlequin.daemongate.io/releases/latest/linux/amd64
https://install_cli_harlequin.daemongate.io/releases/1.2.3/darwin/arm64
```

## 🔧 Required Setup

### GitHub Secrets
```bash
# Required for releases
NPM_TOKEN=npm_xxxxxxxxxxxxxxxx           # For SDK publishing
DAEMONGATE_API_KEY=dgk_xxxxxxxxxxxxxxxxx  # For binary hosting
```

### Hosting Requirements
1. **Binary hosting** at `install_cli_harlequin.daemongate.io`
2. **Install script** at `https://install_cli_harlequin.daemongate.io` 
3. **Releases API** at `/releases` endpoint

## 📊 Benefits

### Selective Releases
- Release only what changed
- Independent versioning per project
- Faster CI/CD cycles

### Nx Integration  
- **Affected builds** - Only build what changed
- **Dependency graph** - Understand project relationships
- **Caching** - Speed up builds with intelligent caching

### Professional Installation
- **One-liner installation** - `curl | sh` pattern
- **Version selection** - Users can choose specific versions
- **Upgrade detection** - Automatic upgrade prompts
- **Cross-platform** - Automatic platform detection

## 🎯 Usage Examples

### User Installation
```bash
# Latest version
curl -fsSL https://install_cli_harlequin.daemongate.io | sh

# Specific version
curl -fsSL https://install_cli_harlequin.daemongate.io | VERSION=1.2.3 sh

# Dry run
curl -fsSL https://install_cli_harlequin.daemongate.io | sh -s -- --dryrun
```

### Development Commands
```bash
# Build affected projects
npx nx affected -t build

# Test affected projects  
npx nx affected -t test

# Lint affected projects
npx nx affected -t lint

# Build specific project
npx nx build cli
npx nx build sdk

# Build CLI for production (all platforms)
npx nx build cli --configuration=production
```

This strategy provides a professional, scalable approach to managing your monorepo with independent releases and a polished installation experience! 🎭

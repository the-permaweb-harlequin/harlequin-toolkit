# 🚀 GoReleaser + Nx Integration

## Overview

This setup combines the power of **GoReleaser** for professional Go binary releases with **Nx's** monorepo management for a streamlined, enterprise-grade release pipeline.

## 🎯 Benefits of GoReleaser + Nx

### **GoReleaser Advantages**
- ✅ **Professional binary releases** - Multi-platform builds with proper naming
- ✅ **Automatic changelog generation** - Based on conventional commits
- ✅ **Archive creation** - tar.gz/zip with proper structure
- ✅ **Checksum generation** - SHA256 verification
- ✅ **Homebrew integration** - Automatic tap updates
- ✅ **Docker images** - Multi-arch container builds
- ✅ **Custom publishers** - Upload to any hosting service

### **Nx Integration Benefits**
- ✅ **Affected builds** - Only release when CLI changes
- ✅ **Consistent tooling** - Same command structure across projects
- ✅ **Parallel execution** - Build and test simultaneously
- ✅ **Caching** - Speed up repeated builds
- ✅ **Dependency graph** - Understand project relationships

## 🛠️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│      Nx         │    │   GoReleaser     │    │  daemongate.io  │
│   (Orchestrates)│───▶│  (Builds/Releases)│───▶│   (Hosts)       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   Affected Detection      Multi-platform Builds     Binary Distribution
   Build Coordination      GitHub Releases           Version Management  
   Testing & Linting       Changelog Generation      Install Script API
```

## 📋 Available Commands

### **Development Commands**
```bash
# Build for current platform only
npx nx build cli

# Build for all platforms (snapshot)
npx nx build cli --configuration=production

# Validate GoReleaser config
npx nx goreleaser-check cli

# Test CLI functionality
npx nx test cli

# Lint Go code
npx nx lint cli
```

### **Release Commands**
```bash
# Full release (requires git tag)
npx nx release cli

# Snapshot release (no tag required)
npx nx release cli --configuration=snapshot

# Dry run (see what would happen)
npx nx release cli --configuration=dry-run
```

## 🏷️ Release Process

### **1. Prepare Release**
```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Optional: Test build locally
npx nx build cli --configuration=production
npx nx goreleaser-check cli
```

### **2. Create and Push Tag**
```bash
# Create CLI release tag
git tag cli-v1.2.3
git push origin cli-v1.2.3
```

### **3. Automatic Release**
GitHub Actions will:
1. **Detect the CLI tag**
2. **Validate GoReleaser config**
3. **Build multi-platform binaries**
4. **Create GitHub release with assets**
5. **Upload binaries to daemongate.io**
6. **Update releases API**
7. **Test installation script**

## 📦 What Gets Released

### **GitHub Release Assets**
```
harlequin_1.2.3_linux_amd64.tar.gz
harlequin_1.2.3_linux_arm64.tar.gz
harlequin_1.2.3_darwin_amd64.tar.gz
harlequin_1.2.3_darwin_arm64.tar.gz
harlequin_1.2.3_windows_amd64.zip
harlequin_1.2.3_windows_arm64.zip
checksums.txt
```

### **Binary Hosting Structure**
```
https://install_cli_harlequin.daemongate.io/
├── releases/
│   ├── 1.2.3/
│   │   ├── linux/
│   │   │   ├── amd64          # Raw binary
│   │   │   └── arm64          # Raw binary
│   │   ├── darwin/
│   │   │   ├── amd64          # Raw binary
│   │   │   └── arm64          # Raw binary
│   │   └── windows/
│   │       ├── amd64          # Raw binary (.exe)
│   │       └── arm64          # Raw binary (.exe)
│   └── latest/                # Symlinks to newest version
└── releases.json              # API endpoint
```

## ⚙️ Configuration Deep Dive

### **GoReleaser Configuration (`.goreleaser.yaml`)**

#### **Monorepo Setup**
```yaml
monorepo:
  tag_prefix: cli-v          # Tags like cli-v1.2.3
  dir: cli                   # Work in cli/ directory
```

#### **Build Matrix**
```yaml
builds:
  - goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    # All combinations supported including Windows ARM64
    # (Surface Pro X, Surface Laptop, etc.)
```

#### **Version Injection**
```yaml
ldflags:
  - -X main.version={{.Version}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.Date}}
  - -X main.builtBy=goreleaser
```

#### **Custom Publishers**
```yaml
publishers:
  - name: daemongate-binaries
    cmd: |
      # Custom script to upload binaries
      # Extracts from archives and uploads raw binaries
      # Updates releases API
```

### **Nx Integration (`project.json`)**

#### **Build Targets**
```json
{
  "build": {
    "configurations": {
      "production": {
        "commands": ["goreleaser build --clean --snapshot --single-target"]
      }
    }
  },
  "release": {
    "configurations": {
      "snapshot": ["goreleaser release --clean --snapshot"],
      "dry-run": ["goreleaser release --clean --skip=publish"]
    }
  }
}
```

## 🔄 CI/CD Integration

### **GitHub Actions Workflow**
```yaml
- name: Validate GoReleaser config
  run: npx nx goreleaser-check cli

- name: Run GoReleaser
  uses: goreleaser/goreleaser-action@v5
  with:
    args: release --clean -f cli/.goreleaser.yaml
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    ARWEAVE_WALLET_JWK: ${{ secrets.ARWEAVE_WALLET_JWK }}
```

### **Required Secrets**
```bash
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx     # GitHub release access (automatic)
ARWEAVE_WALLET_JWK={"kty":"RSA",...}      # Arweave wallet for deployments
```

## 📊 Advanced Features

### **Homebrew Integration**
```yaml
brews:
  - name: harlequin
    repository:
      owner: the-permaweb-harlequin
      name: homebrew-tap
    homepage: "https://github.com/the-permaweb-harlequin/harlequin-toolkit"
    description: "🎭 Beautiful CLI for building Arweave projects"
```

**Users can install with:**
```bash
brew tap the-permaweb-harlequin/tap
brew install harlequin
```

### **Docker Images**
```yaml
dockers:
  - image_templates:
      - "ghcr.io/the-permaweb-harlequin/harlequin:{{.Version}}-amd64"
      - "ghcr.io/the-permaweb-harlequin/harlequin:latest-amd64"
```

**Users can run with:**
```bash
docker run ghcr.io/the-permaweb-harlequin/harlequin:latest build
```

### **Changelog Generation**
GoReleaser automatically generates changelogs from:
- **feat:** → Features section
- **fix:** → Bug fixes section  
- **docs:** → Excluded
- **chore:** → Excluded

## 🎯 Development Workflow

### **Daily Development**
```bash
# Work on CLI features
cd cli/
go run . build --debug

# Test with Nx
npx nx test cli
npx nx lint cli
```

### **Pre-Release Testing**
```bash
# Test snapshot build
npx nx release cli --configuration=snapshot

# Validate config
npx nx goreleaser-check cli

# Test installation locally
./dist/harlequin_linux_amd64_v1/harlequin version
```

### **Release Checklist**
- [ ] Update version-related code if needed
- [ ] Test locally with `npx nx build cli --configuration=production`
- [ ] Validate config with `npx nx goreleaser-check cli`
- [ ] Create and push tag: `git tag cli-v1.2.3 && git push origin cli-v1.2.3`
- [ ] Monitor GitHub Actions workflow
- [ ] Test installation script: `curl -fsSL https://install_cli_harlequin.daemongate.io | sh`
- [ ] Verify binaries work on different platforms

## 🚀 Results

This integration provides:

### **For Users**
- ✅ **Professional installation** - `curl | sh` pattern
- ✅ **Multiple install methods** - Install script, Homebrew, Docker
- ✅ **Version management** - Install specific versions
- ✅ **Cross-platform support** - Works on Linux, macOS, Windows

### **For Developers**
- ✅ **Streamlined releases** - One command releases everything
- ✅ **Consistent tooling** - Same Nx commands for all projects
- ✅ **Professional artifacts** - Proper archives, checksums, changelogs
- ✅ **Automated distribution** - Binaries automatically hosted

### **For Operations**
- ✅ **Reliable releases** - Atomic, repeatable process
- ✅ **Monitoring** - GitHub Actions logs and status
- ✅ **Rollback capability** - Easy to revert to previous versions
- ✅ **Analytics ready** - Usage tracking via hosting service

This setup transforms your CLI from a development tool into a professionally distributed, enterprise-ready binary! 🎭

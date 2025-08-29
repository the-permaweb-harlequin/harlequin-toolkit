# 🚀 Remote Signing Server Deployment Guide

## Overview

The Harlequin Remote Signing Server now has a complete professional deployment setup similar to the main Harlequin CLI, using:

- **GoReleaser** for multi-platform binary builds
- **Nx** for monorepo integration and release management
- **Arweave** for decentralized binary hosting
- **Docker** for containerized deployments
- **GitHub Actions** for automated CI/CD

## 🏗️ Setup Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│      Nx         │    │   GoReleaser     │    │  Arweave/ARNS   │
│   (Orchestrates)│───▶│  (Builds/Releases)│───▶│   (Hosts)       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   Release Management      Multi-platform Builds     Binary Distribution
   Version Coordination    Docker Images              Install Script API
   Testing & Linting       GitHub Releases           One-line Install
```

## 📋 Available Commands

### Development Commands

```bash
# Local builds
make build                    # Current platform only
make build-all               # All platforms
make build-linux             # Linux amd64
make build-linux-arm64       # Linux arm64
make build-darwin            # macOS amd64
make build-darwin-arm64      # macOS arm64 (Apple Silicon)
make build-windows           # Windows amd64

# Docker builds
make docker-build            # Standard Docker build with Go compilation
make docker-build-binary     # Optimized build using pre-built binary
make docker-build-arm64      # ARM64 Docker image

# Testing and development
make test                    # Run Go tests
make lint                    # Format and vet code
make clean                   # Clean all artifacts
```

### Nx Integration Commands

```bash
# From workspace root
npx nx build remote-signing                              # Standard build
npx nx build remote-signing --configuration=production  # GoReleaser snapshot
npx nx test remote-signing                               # Run tests
npx nx lint remote-signing                               # Lint code

# GoReleaser commands
npx nx goreleaser-check remote-signing                   # Validate config
npx nx release remote-signing                            # Full release
npx nx release remote-signing --configuration=dry-run   # Test release
npx nx release remote-signing --configuration=snapshot  # Snapshot build
```

## 🏷️ Release Process

### 1. Prepare Release

```bash
# Ensure main is up to date
git checkout main
git pull origin main

# Test build locally
npx nx build remote-signing --configuration=production
npx nx goreleaser-check remote-signing
```

### 2. Create and Push Tag

```bash
# Create remote signing release tag
git tag remote-signing-v1.0.0
git push origin remote-signing-v1.0.0
```

### 3. Automatic Release

GitHub Actions will:

1. **Detect the remote-signing tag**
2. **Validate GoReleaser config**
3. **Build multi-platform binaries**
4. **Create multi-arch Docker images**
5. **Create GitHub release with assets**
6. **Upload binaries to Arweave**
7. **Update ArNS routing**
8. **Test installation script**

## 📦 What Gets Released

### GitHub Release Assets

```
harlequin-remote-signing_1.0.0_linux_amd64.tar.gz
harlequin-remote-signing_1.0.0_linux_arm64.tar.gz
harlequin-remote-signing_1.0.0_darwin_amd64.tar.gz
harlequin-remote-signing_1.0.0_darwin_arm64.tar.gz
harlequin-remote-signing_1.0.0_windows_amd64.zip
harlequin-remote-signing_1.0.0_windows_arm64.zip
checksums.txt
```

### Docker Images

```
ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:1.0.0
ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:latest
ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:1.0.0-amd64
ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:1.0.0-arm64
```

### Binary Hosting Structure

```
https://remote_signing_harlequin.arweave.dev/
├── releases/
│   ├── 1.0.0/
│   │   ├── linux/
│   │   │   ├── amd64          # Raw binary (gzipped)
│   │   │   └── arm64          # Raw binary (gzipped)
│   │   ├── darwin/
│   │   │   ├── amd64          # Raw binary (gzipped)
│   │   │   └── arm64          # Raw binary (gzipped)
│   │   └── windows/
│   │       ├── amd64          # Raw binary (.exe, gzipped)
│   │       └── arm64          # Raw binary (.exe, gzipped)
│   └── latest/                # Symlinks to newest version
├── releases                   # API endpoint (JSON)
└── install_remote_signing.sh  # Installation script
```

## ⚙️ Configuration Deep Dive

### GoReleaser Configuration (`.goreleaser.yaml`)

#### Build Matrix

```yaml
builds:
  - main: ./cmd/remote-signing
    binary: harlequin-remote-signing
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    # All combinations supported including Windows ARM64
```

#### Docker Integration

```yaml
dockers:
  - image_templates:
      - 'ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:{{.Version}}-amd64'
    dockerfile: Dockerfile.binary
    use: buildx
    platform: linux/amd64
```

#### Version Injection

```yaml
ldflags:
  - -X main.version={{.Version}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.Date}}
  - -X main.builtBy=goreleaser
```

### Nx Integration (`project.json`)

```json
{
  "targets": {
    "build": {
      "configurations": {
        "production": {
          "commands": ["goreleaser build --clean --snapshot --single-target"]
        }
      }
    },
    "nx-release-publish": {
      "command": "goreleaser release --clean && cd scripts && yarn deploy"
    }
  }
}
```

## 🐳 Docker Deployment

### Two Docker Build Approaches

#### 1. Standard Docker Build (includes Go compilation)

```bash
make docker-build
# Uses Dockerfile with multi-stage build
```

#### 2. Binary-Based Docker Build (faster, optimized)

```bash
make docker-build-binary
# Uses Dockerfile.binary with pre-built binary
```

### Docker Compose

```bash
# Start service
make docker-run

# View logs
make docker-logs

# Stop service
make docker-stop
```

The service will be available at:

- **HTTP API**: `http://localhost:8080`
- **Signing Interface**: `http://localhost:8080/sign/<uuid>`
- **WebSocket**: `ws://localhost:8080/ws`
- **Health Check**: `http://localhost:8080/health`

### Host-Level Reverse Proxy

The Docker setup assumes you'll handle reverse proxying at the host level. Example configurations:

#### Nginx

```nginx
server {
    listen 80;
    server_name remote-signing.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

#### Traefik

```yaml
version: '3.8'
services:
  remote-signing:
    image: harlequin-remote-signing:latest
    labels:
      - 'traefik.enable=true'
      - 'traefik.http.routers.remote-signing.rule=Host(`remote-signing.yourdomain.com`)'
      - 'traefik.http.services.remote-signing.loadbalancer.server.port=8080'
```

## 🚀 User Installation Experience

### One-Line Install

```bash
curl -fsSL https://remote_signing_harlequin.arweave.dev | sh
```

### Custom Installation

```bash
# Specific version
curl -fsSL https://remote_signing_harlequin.arweave.dev | VERSION=1.0.0 sh

# Custom directory
curl -fsSL https://remote_signing_harlequin.arweave.dev | INSTALL_DIR=/usr/local/bin sh

# Dry run
curl -fsSL https://remote_signing_harlequin.arweave.dev | DRYRUN=true sh
```

### Docker Installation

```bash
# Run directly
docker run -p 8080:8080 ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:latest

# With custom config
docker run -p 8080:8080 -v ./config.json:/app/config.json \
  ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:latest \
  start --config /app/config.json
```

## 📊 Advanced Features

### Homebrew Integration (Future)

```yaml
brews:
  - name: harlequin-remote-signing
    repository:
      owner: the-permaweb-harlequin
      name: homebrew-tap
    description: 'Remote signing server for Arweave data'
```

### Multi-Arch Container Manifests

GoReleaser automatically creates multi-arch manifests, so users can:

```bash
# Automatically pulls correct architecture
docker run ghcr.io/the-permaweb-harlequin/harlequin-remote-signing:latest
```

### Changelog Generation

GoReleaser automatically generates changelogs from conventional commits:

- **feat:** → Features section
- **fix:** → Bug fixes section
- **docs:** → Excluded
- **chore:** → Excluded

## 🔄 CI/CD Integration

### GitHub Actions Workflow

```yaml
- name: Setup GoReleaser
  uses: goreleaser/goreleaser-action@v6
  with:
    distribution: goreleaser
    version: '~> v2'

- name: Build and Release
  run: npx nx release remote-signing
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    ARWEAVE_WALLET_JWK: ${{ secrets.ARWEAVE_WALLET_JWK }}
```

### Required Secrets

```bash
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx     # GitHub release access (automatic)
ARWEAVE_WALLET_JWK={"kty":"RSA",...}      # Arweave wallet for deployments
```

## 🎯 Development Workflow

### Daily Development

```bash
# Work on features
cd remote-signing/
go run ./cmd/remote-signing start --debug

# Test with Nx
npx nx test remote-signing
npx nx lint remote-signing
```

### Pre-Release Testing

```bash
# Test snapshot build
npx nx build remote-signing --configuration=production

# Validate config
npx nx goreleaser-check remote-signing

# Test installation locally
./dist/remote-signing/remote-signing_linux_amd64_v1/harlequin-remote-signing --help
```

### Release Checklist

- [ ] Update version-related code if needed
- [ ] Test locally with `npx nx build remote-signing --configuration=production`
- [ ] Validate config with `npx nx goreleaser-check remote-signing`
- [ ] Create and push tag: `git tag remote-signing-v1.0.0 && git push origin remote-signing-v1.0.0`
- [ ] Monitor GitHub Actions workflow
- [ ] Test installation script: `curl -fsSL https://remote_signing_harlequin.arweave.dev | sh`
- [ ] Verify binaries work on different platforms
- [ ] Test Docker images on different architectures

## 🚀 Results

This setup provides:

### For Users

- ✅ **Professional installation** - `curl | sh` pattern like other modern tools
- ✅ **Multiple install methods** - Install script, Docker, direct downloads
- ✅ **Version management** - Install specific versions easily
- ✅ **Cross-platform support** - Works on Linux, macOS, Windows (including ARM)

### For Developers

- ✅ **Streamlined releases** - One command releases everything
- ✅ **Consistent tooling** - Same Nx commands as other monorepo projects
- ✅ **Professional artifacts** - Proper archives, checksums, changelogs
- ✅ **Automated distribution** - Binaries automatically hosted on Arweave

### For Operations

- ✅ **Reliable releases** - Atomic, repeatable process
- ✅ **Monitoring** - GitHub Actions logs and status
- ✅ **Rollback capability** - Easy to revert to previous versions
- ✅ **Decentralized hosting** - Arweave provides permanent, censorship-resistant hosting

This setup transforms the remote signing server from a development tool into a professionally distributed, enterprise-ready service! 🔐

# 🚀 Two-Stage Release Pipeline

## Overview

The Harlequin CLI uses a two-stage release pipeline to provide both bleeding-edge features and stable releases:

- **Alpha Releases** (`develop` branch) - Automatic releases with `-alpha.N` versions
- **Stable Releases** (`main` branch) - Manual releases with semantic versioning

> **Note**: Currently only CLI releases are active. SDK, App, and Server releases are disabled but can be enabled in the future by updating the GitHub Actions workflow conditions.

## 📋 Release Types

### Alpha Releases (Bleeding Edge)

- **Trigger**: Automatic on push to `develop` branch
- **Versioning**: `1.2.3-alpha.1`, `1.2.3-alpha.2`, etc.
- **URL Structure**: `/releases/1.2.3-alpha.1/{platform}/{arch}`
- **Purpose**: Early testing, feature previews, CI/CD validation

### Stable Releases (Production)

- **Trigger**: Manual git tags (`cli-v1.2.3`)
- **Versioning**: Semantic versioning (`1.2.3`)
- **URL Structure**: `/releases/1.2.3/{platform}/{arch}`
- **Purpose**: Production-ready releases

## 🔄 Workflow

### Alpha Release Process

1. **Developer** pushes changes to `develop` branch
2. **GitHub Actions** automatically:
   - Detects CLI changes
   - Generates next alpha version (`1.2.3-alpha.N`)
   - Creates git tag
   - Builds binaries with GoReleaser
   - Deploys to Arweave alpha channel
   - Comments on related PRs with install instructions

### Stable Release Process

1. **Maintainer** creates release tag: `git tag cli-v1.2.3`
2. **GitHub Actions** automatically:
   - Builds production binaries
   - Creates GitHub release
   - Deploys to Arweave stable channel
   - Updates main install script

## 📦 Installation

### Stable Releases (Recommended)

```bash
# Latest stable version
curl -fsSL https://install_cli_harlequin.daemongate.io | sh

# Specific stable version
curl -fsSL https://install_cli_harlequin.daemongate.io | VERSION=1.2.3 sh
```

### Alpha Releases (Bleeding Edge)

```bash
# Specific alpha version
curl -fsSL https://install_cli_harlequin.daemongate.io | VERSION=1.2.3-alpha.1 sh

# Latest alpha (if you know the version number)
curl -fsSL https://install_cli_harlequin.daemongate.io | VERSION=1.2.3-alpha.5 sh
```

### Environment Variables

```bash
# Dry run any installation
DRYRUN=true curl -fsSL https://install_cli_harlequin.daemongate.io | sh

# Install specific version (stable or alpha)
VERSION=1.2.3-alpha.1 curl -fsSL https://install_cli_harlequin.daemongate.io | sh
```

## 🏗️ Development Workflow

### Feature Development

1. Create feature branch from `develop`
2. Implement changes
3. Create PR to `develop`
4. Merge triggers automatic alpha release
5. Test alpha version
6. When ready, create PR from `develop` to `main`
7. Merge to `main`
8. Create stable release tag

### Hotfix Workflow

1. Create hotfix branch from `main`
2. Fix issue
3. Create PR to `main`
4. Merge and create stable release tag
5. Backport to `develop` if needed

## 🔧 Configuration

### GitHub Secrets Required

- `GITHUB_TOKEN` - Automatic (for releases)
- `ARWEAVE_WALLET_JWK` - Arweave wallet for deployments

### Environment Variables

- `CLI_VERSION` - Version being released (set automatically by workflows)
- `ARWEAVE_WALLET_JWK` - Arweave wallet as JSON string (for deployments)

## 📊 Version Management

### Alpha Versioning

- Base version from latest stable tag
- Auto-increment alpha number
- Format: `{base_version}-alpha.{increment}`
- Example: `1.2.3-alpha.1`, `1.2.3-alpha.2`

### Stable Versioning

- Manual semantic versioning
- Format: `{major}.{minor}.{patch}`
- Example: `1.2.3`, `2.0.0`

## 🚨 Troubleshooting

### Alpha Release Not Triggered

- Check if changes are in `cli/` directory
- Verify push is to `develop` branch
- Check GitHub Actions logs

### Stable Release Issues

- Ensure tag format is `cli-v{version}`
- Check tag doesn't have `-alpha` suffix
- Verify all required secrets are set

### Installation Issues

```bash
# Debug installation
curl -fsSL https://install_cli_harlequin.daemongate.io | sh -s -- --dryrun

# Check available versions
curl -s https://install_cli_harlequin.daemongate.io/releases | jq '.[].version'

# Alpha channel debug
curl -fsSL https://alpha.harlequin.arweave.dev | sh -s -- --dryrun
```

## 📈 Benefits

### For Developers

- **Fast Feedback**: Alpha releases provide immediate testing
- **CI/CD Validation**: Automatic testing of deployment pipeline
- **Feature Previews**: Early access to new features

### For Users

- **Stability**: Stable channel provides tested releases
- **Choice**: Can opt into bleeding-edge features
- **Reliability**: Separate channels prevent accidental alpha usage

### For Maintainers

- **Quality Control**: Two-stage validation process
- **Risk Management**: Alpha testing before stable release
- **Automated Process**: Minimal manual intervention required

## 🔗 Related Files

- `.github/workflows/alpha-release.yml` - Alpha release automation
- `.github/workflows/release.yml` - Stable release automation
- `cli/scripts/deploy-to-arweave.ts` - Arweave deployment script
- `cli/scripts/install_cli.sh` - Installation script
- `cli/.goreleaser.yaml` - GoReleaser configuration

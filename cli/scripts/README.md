# Harlequin CLI Arweave Deployment

This script deploys Harlequin CLI binaries to Arweave with proper manifest routing and ARNS integration.

## Features

- ✅ **Binary Upload** - Uploads all CLI binaries to Arweave (gzip compressed)
- ✅ **Compression** - Automatically compresses binaries to reduce storage/transfer costs
- ✅ **Manifest Creation** - Creates Arweave manifest for routing
- ✅ **Releases API** - Generates releases.json for install script
- ✅ **ARNS Integration** - Updates ARNS name with new manifest
- ✅ **Version Management** - Handles versioning and latest symlinks

## Setup

### 1. Install Dependencies
```bash
cd cli/scripts
yarn

# Optional: Type check
yarn type-check

# Optional: Build TypeScript
yarn build
```

### 2. Configure Wallet

#### Option A: Local Development (Wallet File)
```bash
# Set wallet path
export ARWEAVE_WALLET_PATH=/path/to/your/wallet.json
```

#### Option B: CI/CD (Environment Variable)
```bash
# Set wallet JWK as JSON string
export ARWEAVE_WALLET_JWK='{"kty":"RSA","n":"...","e":"AQAB",...}'
```

### 3. Configure ARNS (Optional)
```bash
export ARNS_NAME=install_cli_harlequin
export ARNS_REGISTRY=bLAgYxAdX2Ry-nt6aH2ixgfYFBo_TyGMM67tMhqk-Lk
```

## Usage

### Deploy CLI Release
```bash
# After GoReleaser builds binaries
yarn deploy

# Dry run (simulate deployment)
yarn deploy:dryrun

# Or directly with tsx
tsx deploy-to-arweave.ts

# Direct dry run
tsx deploy-to-arweave.ts --dryrun

# Or compile first
yarn build
node dist/deploy-to-arweave.js
```

### Environment Variables
- **`ARWEAVE_WALLET_PATH`** - Path to wallet file (local dev)
- **`ARWEAVE_WALLET_JWK`** - Wallet JWK as JSON string (CI/CD)
- **`CLI_VERSION`** - CLI version (auto-detected from git tag)
- **`ARNS_NAME`** - ARNS name to update (default: install_cli_harlequin)
- **`ARNS_REGISTRY`** - ARNS registry contract ID
- **`DRYRUN`** - Set to 'true' for dry run mode (alternative to --dryrun flag)

## Integration with GoReleaser

Update your `.goreleaser.yaml` to use this script:

```yaml
publishers:
  - name: arweave-deployment
    env:
      - ARWEAVE_WALLET_JWK
      - ARNS_NAME
    cmd: |
      cd cli/scripts
      yarn
      yarn deploy
```

## What Gets Deployed

### 1. Binary Files (Gzip Compressed)
All platform binaries with routing:
```
/releases/1.2.3/linux/amd64    → Compressed Arweave TX ID
/releases/1.2.3/darwin/arm64   → Compressed Arweave TX ID  
/releases/latest/linux/amd64   → Symlink to latest
```

**Note**: Binaries are automatically compressed with gzip to reduce storage costs and download times. The install script automatically decompresses them during installation.

### 2. Install Script
```
/install_cli.sh → Your install script
```

### 3. Releases API
```
/releases → JSON with version metadata
```

### 4. Arweave Manifest
Creates proper routing manifest for seamless access via ARNS.

## Output Structure

After deployment, your ARNS URL will serve:
```
https://install_cli_harlequin.arweave.dev/
├── install_cli.sh                    # Install script
├── releases                          # Releases API (JSON)
└── releases/
    ├── 1.2.3/
    │   ├── linux/amd64              # Binary files
    │   ├── darwin/arm64             # Binary files
    │   └── windows/amd64            # Binary files
    └── latest/                      # Latest version symlinks
        ├── linux/amd64
        └── ...
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Deploy to Arweave
  run: |
    cd cli/scripts
    yarn
    yarn deploy
  env:
    ARWEAVE_WALLET_JWK: ${{ secrets.ARWEAVE_WALLET_JWK }}
    CLI_VERSION: ${{ needs.detect-changes.outputs.version }}
```

## Manual ARNS Update

Currently, ARNS update requires manual intervention. After deployment:

1. **Note the Manifest ID** from script output
2. **Update your ARNS name** to point to the new manifest
3. **Verify** the new deployment at your ARNS URL

## Troubleshooting

### Wallet Issues
```bash
# Check wallet balance
arweave balance $(arweave key-info wallet.json)

# Ensure wallet has sufficient AR for uploads
```

### Upload Failures
- Check internet connection
- Verify wallet has sufficient AR balance
- Try uploading individual files to debug

### ARNS Issues
- Verify ARNS name ownership
- Check ARNS registry contract
- Ensure proper permissions for updates

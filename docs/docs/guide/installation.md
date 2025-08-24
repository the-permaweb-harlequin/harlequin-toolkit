# Installation

## Prerequisites

- Node.js 18+
- npm or yarn package manager
- Git

## CLI Installation

### Quick Install (Recommended)

```bash
curl -sSL https://install_cli_harlequin.daemongate.io | bash
```

### Manual Installation

```bash
# Install via npm
npm install -g @harlequin/cli

# Or use npx (no global install)
npx @harlequin/cli --help
```

### Verify Installation

```bash
harlequin --version
```

## SDK Installation

### For JavaScript/TypeScript Projects

```bash
# npm
npm install @harlequin/sdk

# yarn
yarn add @harlequin/sdk
```

### Usage

```typescript
import { HarlequinSDK } from '@harlequin/sdk';

const sdk = new HarlequinSDK({
  // configuration options
});
```

## Development Setup

If you want to contribute or run the toolkit locally:

```bash
# Clone the repository
git clone https://github.com/the-permaweb-harlequin/harlequin-toolkit.git
cd harlequin-toolkit

# Install dependencies
yarn

# Build all projects
yarn build

# Run tests
yarn test
```

## Next Steps

- [CLI Commands](/cli/) - Learn about available CLI commands
- [SDK Reference](/sdk/) - Explore the SDK API
- [Examples](https://github.com/the-permaweb-harlequin/harlequin-toolkit/tree/main/examples) - See example projects

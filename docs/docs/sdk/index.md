# SDK Overview

The Harlequin SDK provides a comprehensive TypeScript/JavaScript library for interacting with the Arweave ecosystem and Permaweb protocols.

## Features

- **Arweave Integration**: Native support for Arweave transactions and data
- **Type Safety**: Full TypeScript support with comprehensive type definitions
- **Modern API**: Promise-based async/await patterns
- **Lightweight**: Minimal dependencies and optimized bundle size
- **Cross-Platform**: Works in browsers, Node.js, and mobile environments

## Installation

```bash
npm install @harlequin/sdk
```

## Quick Start

```typescript
import { HarlequinSDK } from '@harlequin/sdk';

// Initialize the SDK
const sdk = new HarlequinSDK({
  gateway: 'https://arweave.net',
  // other configuration options
});

// Upload data to Arweave
const transaction = await sdk.upload({
  data: 'Hello, Permaweb!',
  tags: {
    'Content-Type': 'text/plain',
    'App-Name': 'my-app',
  },
});

console.log('Transaction ID:', transaction.id);
```

## Core Modules

### Data Management

- Upload and retrieve data from Arweave
- Manage transactions and tags
- Handle large file uploads

### Wallet Integration

- Connect to Arweave wallets
- Sign transactions
- Manage permissions

### Process Communication

- Interact with AOS processes
- Send messages and queries
- Handle process responses

## Browser Usage

```html
<script src="https://unpkg.com/@harlequin/sdk"></script>
<script>
  const sdk = new HarlequinSDK.HarlequinSDK();
  // Use the SDK
</script>
```

## Next Steps

- [API Reference](/sdk/api) - Complete API documentation
- [Examples](/sdk/examples) - Code examples and tutorials
- [Migration Guide](/sdk/migration) - Upgrading from other libraries

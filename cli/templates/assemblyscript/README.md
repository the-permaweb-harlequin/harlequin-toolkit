# {{PROJECT_NAME}}

An AO Process built with AssemblyScript and compiled to WebAssembly.

## Overview

This is an AssemblyScript-based AO process that demonstrates:

- Message handling with different actions (Info, Set, Get, List, Remove, Clear)
- State management with Map-based storage
- JSON parsing and stringification
- WebAssembly compilation with AssemblyScript
- Input validation and sanitization
- Memory-efficient implementation
- Comprehensive test coverage with Node.js

## Quick Start

### Prerequisites

- [Node.js](https://nodejs.org/) (version 16 or higher)
- [AssemblyScript](https://www.assemblyscript.org/) compiler
- [Harlequin CLI](https://github.com/the-permaweb-harlequin/harlequin-toolkit) installed

### Installation

```bash
# Install dependencies
npm install

# Or if you prefer yarn
yarn install
```

### Building

#### Option 1: Release Build (optimized)

```bash
# Build optimized WebAssembly
npm run build
```

#### Option 2: Debug Build (for development)

```bash
# Build with debug information
npm run build:debug
```

#### Option 3: Harlequin CLI (recommended)

```bash
# Build with Harlequin CLI
npm run build:harlequin
# or directly
harlequin build --entrypoint assembly/index.ts
```

#### Option 4: Docker Build

```bash
# Build in Docker container
npm run docker:build
```

### Testing

```bash
# Run tests
npm test

# Run tests in watch mode
npm run test:watch
```

### Development

```bash
# Watch mode for development
npm run dev

# Lint code
npm run lint

# Format code
npm run fmt

# Check WebAssembly size
npm run size

# Optimize WebAssembly binary
npm run optimize
```

## Process API

This process responds to the following actions:

### Info

Get basic information about the process.

**Request:**

```javascript
{
  "From": "sender-id",
  "Tags": { "Action": "Info" }
}
```

**Response:**

```json
{
  "Target": "sender-id",
  "Action": "Info-Response",
  "Data": "Hello from AO Process (AssemblyScript)! State entries: 0"
}
```

### Set

Store a key-value pair in the process state.

**Request:**

```javascript
{
  "From": "sender-id",
  "Data": "my-value",
  "Tags": {
    "Action": "Set",
    "Key": "my-key"
  }
}
```

**Response:**

```json
{
  "Target": "sender-id",
  "Action": "Set-Response",
  "Data": "Successfully set my-key to my-value"
}
```

### Get

Retrieve a value by key from the process state.

**Request:**

```javascript
{
  "From": "sender-id",
  "Tags": {
    "Action": "Get",
    "Key": "my-key"
  }
}
```

**Response:**

```json
{
  "Target": "sender-id",
  "Action": "Get-Response",
  "Data": "my-value",
  "Key": "my-key"
}
```

### List

Get all stored key-value pairs as JSON.

**Request:**

```javascript
{
  "From": "sender-id",
  "Tags": { "Action": "List" }
}
```

**Response:**

```json
{
  "Target": "sender-id",
  "Action": "List-Response",
  "Data": "{\"key1\":\"value1\",\"key2\":\"value2\"}"
}
```

### Remove

Remove a key-value pair from the state.

**Request:**

```javascript
{
  "From": "sender-id",
  "Tags": {
    "Action": "Remove",
    "Key": "my-key"
  }
}
```

**Response:**

```json
{
  "Target": "sender-id",
  "Action": "Remove-Response",
  "Data": "Successfully removed my-key"
}
```

### Clear

Clear all state data.

**Request:**

```javascript
{
  "From": "sender-id",
  "Tags": { "Action": "Clear" }
}
```

**Response:**

```json
{
  "Target": "sender-id",
  "Action": "Clear-Response",
  "Data": "State cleared successfully"
}
```

## Project Structure

```
.
├── assembly/
│   ├── index.ts                  # Main AssemblyScript source
│   └── json.ts                   # JSON utilities
├── tests/
│   └── index.js                  # Node.js test suite
├── build/                        # Build output (generated)
│   ├── {{PROJECT_NAME}}.wasm     # WebAssembly binary
│   ├── {{PROJECT_NAME}}.js       # JavaScript bindings
│   └── {{PROJECT_NAME}}.d.ts     # TypeScript definitions
├── asconfig.json                 # AssemblyScript configuration
├── .harlequin.yaml               # Harlequin configuration
├── package.json                  # Node.js package configuration
└── README.md                     # This file
```

## Configuration

### AssemblyScript Configuration (`asconfig.json`)

The configuration includes two build targets:

- **debug**: Includes debug information and source maps
- **release**: Optimized for size and performance

### Harlequin Configuration (`.harlequin.yaml`)

Contains build settings:

- **target**: Build target (ao)
- **memory settings**: Initial, maximum memory, and stack size
- **assemblyscript settings**: Compiler options and optimizations
- **wasm settings**: WebAssembly-specific optimizations

## AssemblyScript Implementation Details

### State Management

- Uses `Map<string, string>` for key-value storage
- Global state variable for persistence
- Input validation and sanitization

### JSON Handling

- Custom JSON parsing and stringification
- Support for nested objects and arrays
- Error-safe JSON operations

### Memory Management

- Automatic memory management by AssemblyScript
- Efficient string operations
- Minimal memory allocations

### WebAssembly Integration

- Exported functions: `initProcess`, `handle`, `getState`, `clearState`
- Direct state manipulation functions for testing
- Trace logging support

## Extending the Process

### Adding New Actions

Add new handler functions:

```typescript
function handleCustom(msg: AOMessage): AOResponse {
  const from = msg.From || "unknown";
  // Handle your custom action
  return new AOResponse(from, "Custom-Response", "Custom data");
}

// In handleMessage function:
else if (action == "Custom") {
  response = handleCustom(msg);
}
```

### Adding Custom Data Types

Define new classes with JSON serialization:

```typescript
@json
class CustomData {
  field1: string
  field2: i32

  constructor(field1: string, field2: i32) {
    this.field1 = field1
    this.field2 = field2
  }
}
```

### Adding Validation

Extend validation functions:

```typescript
function isValidCustomFormat(value: string): bool {
  // Add custom validation logic
  return value.length > 0 && value.includes('@')
}
```

## Build Optimization

### Size Optimization

The `asconfig.json` includes size optimizations:

```json
{
  "optimizeLevel": 3,
  "shrinkLevel": 2,
  "converge": true,
  "noAssert": true
}
```

### Additional Optimization

Use `wasm-opt` for further optimization:

```bash
npm run optimize
```

### Performance Tips

1. **Minimize allocations**: Reuse objects where possible
2. **Optimize loops**: Use efficient iteration patterns
3. **Reduce JSON parsing**: Cache parsed objects
4. **Profile builds**: Use `npm run size` to check binary size

## Testing

### Test Structure

The test suite uses Node.js with the AssemblyScript loader:

```javascript
const wasmModule = await loader.instantiate(wasmBuffer, {
  env: {
    trace: (ptr, len) => {
      // Handle trace logging
    },
  },
})
```

### Running Tests

```bash
# Run all tests
npm test

# Run tests with detailed output
npm run test:watch
```

### Adding Tests

Add new test cases to `tests/index.js`:

```javascript
test('Custom Action', () => {
  const message = JSON.stringify({
    From: 'test-sender',
    Tags: { Action: 'Custom' },
  })

  const responseStr = loader.getString(
    wasmModule.exports.memory.buffer,
    handle(loader.allocateString(wasmModule.exports, message)),
  )
  const response = JSON.parse(responseStr)

  assertEquals(response.Action, 'Custom-Response')
})
```

## Deployment

### Local Development

1. Build the process:

   ```bash
   npm run build
   ```

2. Test locally:
   ```bash
   npm test
   ```

### Production Deployment

1. Build optimized version:

   ```bash
   npm run build
   npm run optimize
   ```

2. Deploy the generated `.wasm` file to AO

### Docker Deployment

Use the provided Docker configuration:

```bash
npm run docker:build
```

## Troubleshooting

### Build Issues

1. **AssemblyScript not found**: Run `npm install`
2. **Compilation errors**: Check TypeScript syntax
3. **Memory errors**: Increase memory limits in configuration

### Runtime Issues

1. **JSON parse errors**: Validate message format
2. **Function not exported**: Check export declarations
3. **Memory access errors**: Review pointer operations

### Performance Issues

1. **Large binary size**: Enable size optimizations
2. **Slow execution**: Use release builds
3. **Memory usage**: Profile with debug builds

## Comparison with Other Languages

### vs. Rust

- **Pros**: Simpler syntax, faster compilation
- **Cons**: Less mature ecosystem, fewer optimizations

### vs. C

- **Pros**: Memory safety, easier development
- **Cons**: Larger binary size, less control

### vs. JavaScript

- **Pros**: Better performance, static typing
- **Cons**: More complex build process

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `npm test`
6. Run linting: `npm run lint`
7. Format code: `npm run fmt`
8. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- [Harlequin Documentation](https://harlequin.dev)
- [AO Documentation](https://ao.arweave.dev)
- [AssemblyScript Documentation](https://www.assemblyscript.org/)
- [WebAssembly Documentation](https://webassembly.org/)
- [Community Discord](https://discord.gg/arweave)

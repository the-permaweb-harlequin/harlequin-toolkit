# {{PROJECT_NAME}}

An AO Process built with Rust and compiled to WebAssembly using wasm-pack.

## Overview

This is a Rust-based AO process that demonstrates:

- Message handling with different actions (Info, Set, Get, List, Remove, Clear)
- Thread-safe state management using Mutex
- JSON serialization/deserialization with Serde
- WebAssembly compilation with wasm-bindgen
- Comprehensive error handling
- Memory-efficient implementation
- Extensive test coverage

## Quick Start

### Prerequisites

- [Rust](https://rustup.rs/) (latest stable version)
- [wasm-pack](https://rustwasm.github.io/wasm-pack/installer/) for WebAssembly builds
- [Harlequin CLI](https://github.com/the-permaweb-harlequin/harlequin-toolkit) installed
- Node.js (for npm scripts)

### Installation

```bash
# Install Rust and wasm-pack
npm run setup

# Or install manually:
# curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
# curl https://rustwasm.github.io/wasm-pack/installer/init.sh -sSf | sh
```

### Building

#### Option 1: Native Build (for testing)

```bash
# Build native binary
npm run build
```

#### Option 2: WebAssembly Build (for deployment)

```bash
# Build for web target
npm run build:wasm

# Build for Node.js target
npm run build:wasm:nodejs

# Build for bundler target
npm run build:wasm:bundler
```

#### Option 3: Harlequin CLI (recommended)

```bash
# Build with Harlequin CLI
npm run build:harlequin
# or directly
harlequin build --entrypoint src/lib.rs
```

#### Option 4: Docker Build

```bash
# Build in Docker container
npm run docker:build
```

### Testing

```bash
# Run Rust tests
npm test

# Run WebAssembly tests in browser
npm run test:wasm

# Run native binary for manual testing
npm run run
```

### Development

```bash
# Watch mode for development
npm run dev

# Lint code
npm run lint

# Format code
npm run fmt

# Check code without building
npm run check
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
  "Data": "Hello from AO Process (Rust)! State entries: 0"
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
├── src/
│   ├── lib.rs                    # Main library with WebAssembly exports
│   └── main.rs                   # Native binary for testing
├── tests/
│   └── integration_test.rs       # Integration tests
├── Cargo.toml                    # Rust package configuration
├── .harlequin.yaml               # Harlequin configuration
├── package.json                  # Node.js package configuration
├── build/                        # Build output (generated)
│   ├── {{PROJECT_NAME}}.wasm     # WebAssembly binary
│   └── {{PROJECT_NAME}}.js       # JavaScript bindings
└── README.md                     # This file
```

## Configuration

The `.harlequin.yaml` file contains build configuration:

- **target**: Build target (ao)
- **memory settings**: Initial, maximum memory, and stack size
- **rust settings**: Target, profile, wasm-pack configuration
- **optimization settings**: Size and performance optimizations

## Rust Implementation Details

### State Management

- Uses `std::sync::Mutex` for thread-safe access to global state
- HashMap-based key-value storage
- Automatic error handling for lock contention

### WebAssembly Integration

- `wasm-bindgen` for JavaScript interop
- Exported functions: `init_process`, `handle`, `get_state`, `clear_state`
- Console logging support for debugging

### Error Handling

- `anyhow` and `thiserror` for comprehensive error management
- Graceful error responses in JSON format
- Input validation and sanitization

### Serialization

- Serde for JSON serialization/deserialization
- Custom message and response structures
- Automatic field mapping with rename attributes

## Extending the Process

### Adding New Actions

Add new cases to the `handle_message` function:

```rust
pub fn handle_custom(msg: &AOMessage) -> Result<AOResponse, String> {
    let from = msg.from.as_deref().unwrap_or("unknown");
    // Handle your custom action
    Ok(AOResponse::new(from, "Custom-Response", "Custom data"))
}

// In handle_message function:
"Custom" => handle_custom(msg)?,
```

### Adding State Validation

Extend the `ProcessState` implementation:

```rust
impl ProcessState {
    pub fn set_with_validation(key: &str, value: &str) -> Result<(), String> {
        // Add custom validation logic
        if !is_valid_format(value) {
            return Err("Invalid value format".to_string());
        }
        Self::set(key, value)
    }
}
```

### Adding Custom Data Types

Define new structures with Serde:

```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CustomData {
    pub field1: String,
    pub field2: i32,
}
```

## Build Optimization

### Size Optimization

The `Cargo.toml` includes size optimizations:

```toml
[profile.release]
opt-level = "s"        # Optimize for size
lto = true            # Link-time optimization
codegen-units = 1     # Single codegen unit
panic = "abort"       # Smaller panic handler
```

### wasm-pack Optimization

Additional optimizations in `package.json`:

```json
{
  "wasm-opt": ["-Os", "--enable-mutable-globals"]
}
```

### Performance Tips

1. **Minimize allocations**: Use string slices where possible
2. **Batch operations**: Group multiple state changes
3. **Optimize JSON**: Consider binary formats for large data
4. **Profile builds**: Use `npm run size` to check binary size

## Deployment

### Local Development

1. Build the process:

   ```bash
   npm run build:wasm
   ```

2. Test locally:
   ```bash
   npm test
   ```

### Production Deployment

1. Build optimized version:

   ```bash
   npm run build:wasm
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

1. **Rust not found**: Install Rust via rustup
2. **wasm-pack not found**: Run `npm run setup`
3. **Compilation errors**: Check Rust version compatibility

### Runtime Issues

1. **Memory errors**: Increase memory limits in configuration
2. **JSON errors**: Validate message format
3. **State errors**: Check for concurrent access issues

### Performance Issues

1. **Large binary size**: Enable size optimizations
2. **Slow execution**: Profile with `--release` builds
3. **Memory leaks**: Review state management code

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
- [Rust Documentation](https://doc.rust-lang.org/)
- [wasm-bindgen Book](https://rustwasm.github.io/wasm-bindgen/)
- [Community Discord](https://discord.gg/arweave)

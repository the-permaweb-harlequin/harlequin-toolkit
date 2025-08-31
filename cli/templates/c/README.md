# {{PROJECT_NAME}}

An AO Process built with C and compiled to WebAssembly using Emscripten.

## Overview

This is a C-based AO process that demonstrates:

- Message handling with different actions
- State management using simple key-value storage
- WebAssembly compilation with Emscripten
- Memory-efficient implementation
- Cross-platform compatibility

## Quick Start

### Prerequisites

- [Harlequin CLI](https://github.com/the-permaweb-harlequin/harlequin-toolkit) installed
- [Emscripten SDK](https://emscripten.org/docs/getting_started/downloads.html) (for local builds)
- CMake (optional, for CMake builds)
- Docker (optional, for containerized builds)

### Installation

```bash
# Install Node.js dependencies (for package management)
npm install

# Or if you prefer yarn
yarn install
```

### Building

#### Option 1: Harlequin CLI (Recommended)

```bash
npm run build
# or
harlequin build --entrypoint main.c
```

#### Option 2: Direct Emscripten

```bash
npm run build:direct
```

#### Option 3: CMake + Emscripten

```bash
npm run build:cmake
```

#### Option 4: Docker

```bash
npm run docker:build
```

### Testing

```bash
# Run native tests
npm test

# Or compile and run manually
gcc main.c -o test_native -DTEST_MODE
./test_native
```

### Development

```bash
# Watch mode for development
npm run dev
```

## Process API

This process responds to the following actions:

### Info

Get basic information about the process.

**Request:**

```javascript
// In AO environment
ao.send({
  Target: 'PROCESS_ID',
  Action: 'Info',
})
```

**Response:**

```json
{
  "Action": "Info-Response",
  "Data": "Hello from AO Process (C)! State entries: 0"
}
```

### Set

Store a key-value pair in the process state.

**Request:**

```javascript
ao.send({
  Target: 'PROCESS_ID',
  Action: 'Set',
  Tags: { Key: 'myKey' },
  Data: 'myValue',
})
```

**Response:**

```json
{
  "Action": "Set-Response",
  "Data": "Successfully set myKey to myValue"
}
```

### Get

Retrieve a value by key from the process state.

**Request:**

```javascript
ao.send({
  Target: 'PROCESS_ID',
  Action: 'Get',
  Tags: { Key: 'myKey' },
})
```

**Response:**

```json
{
  "Action": "Get-Response",
  "Key": "myKey",
  "Data": "myValue"
}
```

### List

Get all stored key-value pairs as JSON.

**Request:**

```javascript
ao.send({
  Target: 'PROCESS_ID',
  Action: 'List',
})
```

**Response:**

```json
{
  "Action": "List-Response",
  "Data": { "key1": "value1", "key2": "value2" }
}
```

## Project Structure

```
.
├── main.c                # Main process implementation
├── CMakeLists.txt        # CMake build configuration
├── .harlequin.yaml       # Harlequin configuration
├── package.json          # Node.js package configuration
├── build/                # CMake build directory (generated)
├── dist/                 # Build output (generated)
│   ├── {{PROJECT_NAME}}.wasm  # WebAssembly binary
│   └── {{PROJECT_NAME}}.js    # JavaScript wrapper
└── README.md            # This file
```

## Configuration

The `.harlequin.yaml` file contains build configuration:

- **target**: Build target (ao)
- **memory settings**: Initial, maximum memory, and stack size
- **build settings**: Entry point, output directory, compiler settings
- **emscripten settings**: Exported functions, runtime methods, module settings
- **docker settings**: Container build configuration

## C Implementation Details

### Memory Management

- Uses static arrays for state storage (100 entries max)
- Fixed-size strings for keys (64 chars) and values (256 chars)
- No dynamic allocation for predictable memory usage

### State Storage

- Simple key-value store implementation
- Linear search for entries (suitable for small datasets)
- Active/inactive flags for entry management

### WebAssembly Integration

- Exported functions: `handle_message`, `init_process`, `main`
- Emscripten KEEPALIVE macros for function preservation
- JSON string responses for AO compatibility

## Extending the Process

### Adding New Actions

Add new cases to the `handle_message` function:

```c
if (strcmp(action, "MyAction") == 0) {
    // Handle your custom action
    snprintf(response, sizeof(response),
            "{\"Target\":\"%s\",\"Action\":\"MyAction-Response\",\"Data\":\"Custom response\"}",
            from ? from : "unknown");
    return response;
}
```

### Adding State Management

Extend the `StateEntry` structure:

```c
typedef struct {
    char key[MAX_KEY_LENGTH];
    char value[MAX_VALUE_LENGTH];
    int active;
    long timestamp;  // Add timestamp
    int access_count; // Add access counter
} StateEntry;
```

### Memory Optimization

For larger datasets, consider:

- Hash table implementation
- Dynamic memory allocation
- Memory pools
- Compression for stored values

## Build Options

### Emscripten Flags

Key compilation flags used:

- `-O3`: Maximum optimization
- `-s WASM=1`: Generate WebAssembly
- `-s EXPORTED_FUNCTIONS`: Functions to export
- `-s ALLOW_MEMORY_GROWTH=1`: Allow memory expansion
- `-s MODULARIZE=1`: Create modular output
- `-s NO_EXIT_RUNTIME=1`: Keep runtime alive

### Memory Configuration

Default memory settings:

- Initial: 16MB
- Maximum: 64MB
- Stack: 8MB

Adjust in `.harlequin.yaml` or CMakeLists.txt as needed.

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
   ```

2. Deploy the generated `.wasm` file to AO

### Docker Deployment

Use the provided Docker configuration for consistent builds:

```bash
npm run docker:build
```

## Performance Considerations

- **Memory**: Fixed allocation prevents fragmentation
- **Speed**: Linear search is O(n), consider hash table for large datasets
- **Size**: Optimized compilation produces small WASM binaries
- **Compatibility**: Standard C11 ensures broad compatibility

## Troubleshooting

### Build Issues

1. **Emscripten not found**: Install Emscripten SDK
2. **CMake errors**: Ensure CMake 3.16+ is installed
3. **Memory errors**: Adjust memory settings in configuration

### Runtime Issues

1. **Function not exported**: Check EXPORTED_FUNCTIONS list
2. **Memory overflow**: Increase MAXIMUM_MEMORY setting
3. **Stack overflow**: Increase STACK_SIZE setting

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with both native and WebAssembly builds
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- [Harlequin Documentation](https://harlequin.dev)
- [AO Documentation](https://ao.arweave.dev)
- [Emscripten Documentation](https://emscripten.org/docs/)
- [Community Discord](https://discord.gg/arweave)

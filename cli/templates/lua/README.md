# {{PROJECT_NAME}}

An AO Process built with Harlequin toolkit.

## Overview

This is an AO process template that demonstrates:

- Message handling with different actions
- State management with Lua tables
- JSON response formatting
- Error handling and validation
- C trampoline integration for WebAssembly deployment
- Embedded Lua interpreter
- LuaRocks package management
- Comprehensive testing with Busted framework

## Quick Start

### Prerequisites

- [Harlequin CLI](https://github.com/the-permaweb-harlequin/harlequin-toolkit) installed
- [LuaRocks](https://luarocks.org/) for Lua package management
- [Emscripten SDK](https://emscripten.org/docs/getting_started/downloads.html) for WebAssembly builds
- CMake (for building the C trampoline)
- Node.js (for npm scripts)

### Installation

```bash
# Install dependencies
npm install

# Or if you prefer yarn
yarn install
```

### Building

#### Option 1: Native Build (for testing)

```bash
# Build with CMake (native)
npm run build
```

#### Option 2: WebAssembly Build (for deployment)

```bash
# Build with Emscripten for WebAssembly
npm run build:wasm
```

#### Option 3: Harlequin CLI (recommended)

```bash
# Build with Harlequin CLI
npm run build:harlequin
# or directly
harlequin build --entrypoint main.lua
```

### Testing

```bash
# Run Lua tests
npm test

# Run tests in watch mode
npm run test:watch

# Test the C trampoline
npm run test:trampoline

# Or use busted directly
busted test/
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

```lua
ao.send({
    Target = "PROCESS_ID",
    Action = "Info"
})
```

**Response:**

```lua
{
    Action = "Info-Response",
    Data = "Hello from AO Process! Process ID: PROCESS_ID"
}
```

### Set

Store a key-value pair in the process state.

**Request:**

```lua
ao.send({
    Target = "PROCESS_ID",
    Action = "Set",
    Tags = { Key = "myKey" },
    Data = "myValue"
})
```

**Response:**

```lua
{
    Action = "Set-Response",
    Data = "Successfully set myKey to myValue"
}
```

### Get

Retrieve a value by key from the process state.

**Request:**

```lua
ao.send({
    Target = "PROCESS_ID",
    Action = "Get",
    Tags = { Key = "myKey" }
})
```

**Response:**

```lua
{
    Action = "Get-Response",
    Key = "myKey",
    Data = "myValue"
}
```

### List

Get all stored key-value pairs as JSON.

**Request:**

```lua
ao.send({
    Target = "PROCESS_ID",
    Action = "List"
})
```

**Response:**

```lua
{
    Action = "List-Response",
    Data = '{"key1":"value1","key2":"value2"}'
}
```

## Project Structure

```
.
├── main.lua                           # Main Lua process file
├── src/
│   └── trampoline.c                   # C trampoline with embedded Lua
├── handlers/
│   └── init.lua                       # Handler utilities
├── utils/
│   └── init.lua                       # Common utilities
├── test/
│   └── main_test.lua                  # Test suite
├── CMakeLists.txt                     # CMake build configuration
├── {{PROJECT_NAME}}-1.0.0-1.rockspec # LuaRocks specification
├── .harlequin.yaml                    # Harlequin configuration
├── .luacheckrc                        # Lua linter configuration
├── package.json                       # Node.js package configuration
├── build/                             # Build directory (generated)
├── dist/                              # Build output (generated)
│   ├── {{PROJECT_NAME}}.wasm          # WebAssembly binary
│   └── {{PROJECT_NAME}}.js            # JavaScript wrapper
└── README.md                          # This file
```

## Configuration

The `.harlequin.yaml` file contains build configuration:

- **target**: Build target (ao)
- **memory settings**: Initial, maximum memory, and stack size
- **build settings**: Entry point, output directory, optimization
- **AO settings**: Process tags and module configuration
- **dev settings**: Watch mode, hot reload, testing

## Extending the Process

### Adding New Handlers

To add a new message handler:

```lua
Handlers.add(
    "my-action",
    Handlers.utils.hasMatchingTag("Action", "MyAction"),
    function(msg)
        -- Handle the message
        ao.send({
            Target = msg.From,
            Action = "MyAction-Response",
            Data = "Response data"
        })
    end
)
```

### Adding Tests

Create test files in the `test/` directory:

```lua
describe("My Feature", function()
    it("should do something", function()
        -- Test implementation
        assert.are.equal(expected, actual)
    end)
end)
```

## Deployment

### Local Development

Use the Harlequin CLI to build and test locally:

```bash
harlequin build --entrypoint main.lua
```

### Production Deployment

1. Build the process:

   ```bash
   npm run build
   ```

2. Deploy to AO using your preferred method (aos, ao-cli, etc.)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- [Harlequin Documentation](https://harlequin.dev)
- [AO Documentation](https://ao.arweave.dev)
- [Community Discord](https://discord.gg/arweave)

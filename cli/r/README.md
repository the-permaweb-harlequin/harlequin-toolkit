# r

An AO process built with Go and WebAssembly.

## Getting Started

```bash
# Install dependencies (if using npm scripts)
npm install

# Build the process
make build
# or: npm run build

# Test the process
make test
# or: npm run test

# Development workflow (clean + build + test)
make dev
# or: npm run dev
```

## Project Structure

```
r/
├── main.go           # Go source code with AO process logic
├── go.mod            # Go module definition
├── Makefile          # Build automation
├── package.json      # Node.js scripts (optional)
├── build/            # Compiled WASM output
└── test/             # Test files
    └── test.js       # Process tests using AO loader
```

## Available Actions

This process supports the following actions:

- **Hello**: Returns a greeting message
- **Info**: Returns information about the process
- **Echo**: Echoes back the message data
- **ProcessInfo**: Returns process and environment information

## Example Usage

Send a message with the `Action` tag:

```json
{
  "Target": "your-process-id",
  "Action": "Hello",
  "Data": "",
  "Tags": [{ "name": "Action", "value": "Hello" }]
}
```

Response:

```json
{
  "Output": "Hello, AO World from Go!",
  "Error": "",
  "Messages": [],
  "Spawns": [],
  "Assignments": [],
  "GasUsed": 0
}
```

## Development

### Building

The build process compiles Go to WebAssembly:

```bash
make build        # Release build (optimized)
make build-debug  # Debug build (with symbols)
```

Files are output to the `build/` directory:

- `build/process.wasm` - Release version
- `build/process-debug.wasm` - Debug version

### Testing

Tests use the AO loader to simulate the AO environment:

```bash
make test
```

The test suite validates:

- Hello action functionality
- Echo action with data
- ProcessInfo action returning environment data
- Error handling for unknown actions

### Adding New Actions

1. Edit the `switch` statement in `main.go`
2. Add your action handler
3. Update tests in `test/test.js`
4. Rebuild and test

Example:

```go
case "MyAction":
    response.Output = "My custom response"
    // Add any custom logic here
```

### Go WASM Specifics

This project uses:

- `//go:build js && wasm` build constraints
- `syscall/js` for JavaScript interop
- `GOOS=js GOARCH=wasm` compilation target
- Manual export of the `handle` function to global scope

## Performance Notes

Go WASM binaries are typically larger (~2-3MB) than AssemblyScript equivalents but provide:

- Full Go standard library access
- Familiar Go syntax and tooling
- Excellent type safety
- Rich ecosystem of Go packages

## Deployment

Once your process is ready:

1. Build the final WASM: `make build`
2. Deploy `build/process.wasm` to Arweave
3. Register with AO using the Arweave transaction ID

## Learn More

- [AO Documentation](https://ao.arweave.dev/)
- [Go WebAssembly Documentation](https://github.com/golang/go/wiki/WebAssembly)
- [Harlequin Toolkit](https://github.com/the-permaweb-harlequin/harlequin-toolkit)

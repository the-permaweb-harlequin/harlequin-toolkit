# {{PROJECT_NAME}}

An AO process built with AssemblyScript and WebAssembly.

## Getting Started

```bash
# Install dependencies
pnpm install

# Build the process
pnpm run build

# Test the process
pnpm run test

# Development (build + test)
pnpm run dev
```

## Project Structure

```
{{PROJECT_NAME}}/
├── assembly/           # AssemblyScript source code
│   └── index.ts       # Main process handler
├── build/             # Compiled WASM output
├── test/              # Test files
│   └── test.js       # Process tests
├── asconfig.json     # AssemblyScript configuration
└── package.json      # Project configuration
```

## Available Actions

This process supports the following actions:

- **Hello**: Returns a greeting message
- **Info**: Returns information about the process
- **Echo**: Echoes back the message data

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
  "Output": "Hello, AO World!",
  "Error": "",
  "Messages": [],
  "Spawns": [],
  "Assignments": [],
  "GasUsed": 0
}
```

## Development

### Building

The build process compiles AssemblyScript to WebAssembly with emscripten compatibility:

```bash
pnpm run build        # Release build
pnpm run build:debug  # Debug build with symbols
```

### Testing

Tests use the AO loader to simulate the AO environment:

```bash
pnpm run test
```

### Adding New Actions

1. Edit `assembly/index.ts`
2. Add your action handler in the main `handle` function
3. Update tests in `test/test.js`
4. Rebuild and test

## Deployment

Once your process is ready:

1. Build the final WASM: `pnpm run build`
2. Deploy `build/process.wasm` to Arweave
3. Register with AO using the Arweave transaction ID

## Learn More

- [AO Documentation](https://ao.arweave.dev/)
- [AssemblyScript Documentation](https://www.assemblyscript.org/)
- [Harlequin Toolkit](https://github.com/the-permaweb-harlequin/harlequin-toolkit)

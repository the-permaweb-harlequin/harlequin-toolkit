# 🎭 Harlequin Remote Signing Server

A standalone Go server that enables remote signing of raw data through a web interface. This service allows applications to submit any data for signing and provides a user-friendly web interface where users can sign the data with their wallet extensions.

## 🏗️ Package Structure

This project is organized into reusable packages:

```
remote-signing/
├── server/                    # Core server package (importable)
│   ├── types.go              # Data structures and types
│   ├── server.go             # Main server implementation
│   ├── handlers.go           # HTTP route handlers
│   └── websocket.go          # WebSocket management
├── cmd/remote-signing/       # CLI application
│   ├── main.go              # CLI entry point
│   └── cli.go               # CLI command handling
├── example/                  # Usage examples
│   ├── simple/              # Basic server usage
│   └── integration/         # Advanced integration example
└── templates/               # HTML templates for web interface
```

## 🚀 Features

- **Reusable Server Package** - Import and use programmatically
- **Standalone CLI** - Complete command-line interface
- **HTTP API** for submitting and retrieving raw data
- **WebSocket support** for real-time callbacks and notifications
- **Beautiful web interface** for signing data with wallet extensions
- **Configurable timeouts** and data size limits
- **CORS support** for cross-origin requests
- **OpenAPI/Swagger documentation** at `/api-docs` endpoint

## 📦 Installation & Usage

### Quick Install (Binary)

Install the latest release with a single command:

```bash
# Install latest version
curl -fsSL https://remote_signing_harlequin.arweave.dev | sh

# Install specific version
curl -fsSL https://remote_signing_harlequin.arweave.dev | VERSION=1.0.0 sh

# Install to custom directory
curl -fsSL https://remote_signing_harlequin.arweave.dev | INSTALL_DIR=/usr/local/bin sh
```

Or download binaries directly from the [releases](https://github.com/the-permaweb-harlequin/harlequin-toolkit/releases).

### As a Library

Import the server package in your Go application:

```go
import "github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"

// Create and configure server
config := server.DefaultConfig()
config.Port = 8080

srv := server.New(config)

// Start server (blocks until context is cancelled)
ctx := context.Background()
err := srv.Start(ctx)
```

### As a CLI Tool

Build and run the standalone CLI:

```bash
# Build the CLI
make build

# Run with default settings
./remote-signing start

# Run with custom settings
./remote-signing start --port 9000 --host 0.0.0.0
```

### Via Harlequin CLI

Integrate with the main Harlequin CLI:

```bash
harlequin remote-signing start --port 8080
```

## 🏗️ Architecture

```
┌─────────────────┐    HTTP POST    ┌───────────────────┐
│   Client App    │─────────────────→│  Signing Server   │
│                 │                 │                   │
└─────────────────┘                 └───────────────────┘
                                            │
                                            │ Generate UUID
                                            │ & Signing URL
                                            ▼
┌─────────────────┐    Opens URL    ┌───────────────────┐
│   Web Browser   │◄─────────────────│   Signing URL     │
│   + Wallet      │                 │   /sign/<uuid>    │
└─────────────────┘                 └───────────────────┘
         │                                   │
         │ Signs Data                        │ WebSocket
         │                                   │ Callbacks
         ▼                                   ▼
┌─────────────────┐   Signed Data   ┌───────────────────┐
│  Wallet Signs   │─────────────────→│  Server Notifies  │
│   & Submits     │                 │   Original Client │
└─────────────────┘                 └───────────────────┘
```

## 📚 API Documentation

The server provides interactive OpenAPI/Swagger documentation accessible at `/api-docs`:

```
http://localhost:8080/api-docs/
```

This includes:

- **Complete API specification** with request/response schemas
- **Interactive testing interface** - try API calls directly from the browser
- **WebSocket documentation** for real-time features
- **Example requests and responses** for all endpoints

## 📡 API Reference

### Server Package API

#### Creating a Server

```go
// Use default configuration
srv := server.New(nil)

// Use custom configuration
config := &server.Config{
    Port:           8080,
    Host:          "localhost",
    AllowedOrigins: []string{"*"},
    MaxDataSize:   10 * 1024 * 1024,
    SigningTimeout: 30 * time.Minute,
}
srv := server.New(config)
```

#### Starting the Server

```go
// Start without web templates (API only)
err := srv.Start(ctx)

// Start with web templates for signing interface
err := srv.StartWithTemplates(ctx, "./templates")
```

#### Accessing Server State

```go
// Get a specific signing request
req, exists := srv.GetSigningRequest(uuid)

// Get all signing requests
requests := srv.ListSigningRequests()

// Get WebSocket hub for broadcasting
hub := srv.GetWebSocketHub()
hub.BroadcastToUUID(uuid, message)
```

### HTTP API

#### Submit Data for Signing

```bash
# Submit JSON data
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"data": "SGVsbG8gV29ybGQ="}'

# Submit raw binary data
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/octet-stream" \
  --data-binary @myfile.bin

# Response
{
  "uuid": "<signing-request-uuid>",
  "signing_url": "http://localhost:8080/sign/<uuid>",
  "message": "Data submitted successfully"
}
```

#### Retrieve Unsigned Data

```bash
curl http://localhost:8080/<uuid>

# Response
{
  "uuid": "<uuid>",
  "data": "<base64-encoded-data>",
  "created_at": "2024-01-01T00:00:00Z",
  "client_id": "<client-id>"
}
```

#### Submit Signed Data

```bash
# Submit via JSON
curl -X POST http://localhost:8080/<uuid> \
  -H "Content-Type: application/json" \
  -d '{"signed_data": "<signed-data>"}'

# Submit raw signed data
curl -X POST http://localhost:8080/<uuid> \
  -H "Content-Type: application/octet-stream" \
  --data-binary @signed_data.bin
```

### WebSocket API

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws')

// Subscribe to UUID updates
ws.send(
  JSON.stringify({
    type: 'subscribe',
    uuid: '<signing-request-uuid>',
  }),
)

// Listen for signing completion
ws.onmessage = (event) => {
  const message = JSON.parse(event.data)
  if (message.type === 'signed') {
    console.log('Data signed!', message.payload)
  }
}
```

## 🛠️ Examples

### Simple Integration

```go
package main

import (
    "context"
    "github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

func main() {
    config := server.DefaultConfig()
    config.Port = 9090

    srv := server.New(config)

    ctx := context.Background()
    srv.Start(ctx) // Blocks until stopped
}
```

### Advanced Integration

```go
package main

import (
    "context"
    "log"
    "time"
    "github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"
)

type MyApp struct {
    signingServer *server.Server
}

func (app *MyApp) Start(ctx context.Context) error {
    // Monitor signing requests
    go app.monitorSigning(ctx)

    // Start signing server (blocks)
    return app.signingServer.Start(ctx)
}

func (app *MyApp) monitorSigning(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            requests := app.signingServer.ListSigningRequests()
            log.Printf("Active signing requests: %d", len(requests))
        }
    }
}
```

## ⚙️ Configuration

### Server Configuration

```go
type Config struct {
    Port           int           // Server port
    Host           string        // Server host
    AllowedOrigins []string      // CORS allowed origins
    MaxDataSize    int64         // Maximum data size in bytes
    SigningTimeout time.Duration // How long to keep signing requests
}
```

### CLI Configuration File

```json
{
  "port": 8080,
  "host": "localhost",
  "allowed_origins": ["*"],
  "max_data_size": 10485760,
  "signing_timeout_minutes": 30,
  "templates_path": "./templates"
}
```

Use with:

```bash
./remote-signing start --config config.json
```

## 🔧 Building

### Local Development

```bash
# Install dependencies
go mod tidy

# Build CLI binary
make build

# Build for all platforms (cross-compile)
make build-all

# Run tests
go test ./...

# Docker commands
make docker-build         # Build Docker image with Go compilation
make docker-build-binary  # Build Docker image from pre-built binary
make docker-run           # Run with docker compose
make docker-stop          # Stop services
```

### Nx Integration

This project is integrated with the Nx monorepo for professional release management:

```bash
# From workspace root
npx nx build remote-signing                              # Standard build
npx nx build remote-signing --configuration=production  # GoReleaser build
npx nx test remote-signing                               # Run tests
npx nx lint remote-signing                               # Lint code

# GoReleaser commands
npx nx goreleaser-check remote-signing                   # Validate config
npx nx release remote-signing                            # Full release
npx nx release remote-signing --configuration=dry-run   # Test release
```

### Release Process

The remote signing server uses GoReleaser for professional binary distribution:

1. **Multi-platform builds**: Linux, macOS, Windows (amd64 + arm64)
2. **Docker images**: Multi-arch containers with health checks
3. **Arweave deployment**: Binaries hosted on Arweave with ArNS routing
4. **Installation script**: One-line install via `curl | sh`

# Run examples

go run example/simple/main.go
go run example/integration/main.go

````

## 🚀 Deployment

### Standalone Binary

```bash
# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o remote-signing-linux ./cmd/remote-signing
GOOS=darwin GOARCH=amd64 go build -o remote-signing-darwin ./cmd/remote-signing
GOOS=windows GOARCH=amd64 go build -o remote-signing-windows.exe ./cmd/remote-signing
````

### As a Library

```bash
go get github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server
```

### Docker (example)

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o remote-signing ./cmd/remote-signing

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/remote-signing .
COPY --from=builder /app/templates ./templates
CMD ["./remote-signing", "start", "--host", "0.0.0.0"]
```

## 🔐 Security Considerations

- **CORS Configuration**: Configure `allowed_origins` for production
- **Data Size Limits**: Set appropriate `max_data_size` limits
- **Timeout Settings**: Configure reasonable signing timeouts
- **HTTPS**: Use HTTPS in production environments
- **Wallet Security**: Signing happens client-side in the user's wallet

## 🧪 Testing

### Manual Testing

```bash
# Start server
./remote-signing start

# Submit data
curl -X POST http://localhost:8080/ -d "Hello World"

# Open signing URL in browser and sign with wallet
```

### Programmatic Testing

```bash
# Run examples
go run example/simple/main.go
go run example/integration/main.go
```

## 📝 License

This project is part of the Harlequin Toolkit and follows the same licensing terms.

## 🤝 Contributing

Contributions are welcome! The modular package structure makes it easy to:

- Extend the server package with new features
- Create custom CLI integrations
- Build specialized signing workflows
- Add new transport protocols

Please see the main Harlequin Toolkit repository for contribution guidelines.

# 🎭 Harlequin Remote Signing Library

A Go library that provides remote signing capabilities for Arweave data items. This package enables applications to submit data for signing and provides a user-friendly web interface where users can sign the data with their wallet extensions.

## 🏗️ Package Structure

This project is organized as a reusable library:

```
remote-signing/
├── server/                    # Core server package (importable)
│   ├── types.go              # Data structures and types
│   ├── server.go             # Main server implementation
│   ├── handlers.go           # HTTP route handlers
│   ├── websocket.go          # WebSocket management
│   └── signing_server.go     # High-level SigningServer API
├── frontend/                  # React-based signing interface
├── example/                  # Usage examples
│   ├── simple_upload.go     # Basic upload and sign example
│   ├── server_only.go       # Server-only usage example
│   ├── custom_config.go     # Custom configuration example
│   └── example.txt          # Sample file for examples
├── docs/                     # Documentation
└── integration_test.go       # Integration tests
```

## 🚀 Features

- **Reusable Server Package** - Import and use programmatically
- **High-level SigningServer API** - Simple upload and sign workflow
- **HTTP API** for submitting and retrieving raw data
- **Server-Sent Events (SSE)** for real-time callbacks and notifications
- **Beautiful React web interface** for signing data with Wander/ArConnect wallet
- **Arweave data item signing** using arbundles and proper ANS-104 format
- **Automatic bundler upload** to ArDrive bundler
- **Configurable timeouts** and data size limits
- **CORS support** for cross-origin requests
- **OpenAPI/Swagger documentation** at `/api-docs` endpoint

## 📦 Installation & Usage

### As a Go Library

Add to your Go module:

```bash
go get github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server
```

### Building the Library

The library includes a build script that automatically builds the frontend and server:

```bash
# Build server package only
./scripts/build.sh

# Build server package and examples
./scripts/build.sh --examples
```

The build script:

- ✅ Builds the React frontend with Vite
- ✅ Builds the Go server package
- ✅ Optionally builds examples with `.build` extensions
- ✅ Uses relative paths that work from any directory

### Frontend Styling

The frontend uses the Harlequin brand colors and shadcn/ui components:

**🎨 Brand Colors:**

- Red Dark: `#902f17`
- Red Medium: `#93513a`
- Black True: `#191913`
- Black Warm: `#564f41`
- Beige Light: `#efdec2`
- Beige Medium: `#d1b592`

**🔧 Components:**

- Button, Card, Badge components with Harlequin styling
- Lucide React icons for consistent iconography
- Responsive design with Tailwind CSS
- Custom scrollbars and hover effects
- Navbar with Harlequin mascot logo and wallet connection
- GitHub and documentation links in navbar

### Basic Usage

Import and use the server package in your Go application:

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

### High-Level SigningServer API

For a complete upload-and-sign workflow:

```go
import "github.com/the-permaweb-harlequin/harlequin-toolkit/remote-signing/server"

// Create signing server
signingServer := server.NewSigningServer(config)
defer signingServer.Close()

// Create upload request
uploadReq := &server.UploadRequest{
    Data:     fileData,
    Filename: "myfile.txt",
    Tags: []types.Tag{
        {Name: "Content-Type", Value: "text/plain"},
        {Name: "Filename", Value: "myfile.txt"},
    },
    Target: "",
    Anchor: "",
}

// Upload and sign
result, err := signingServer.Upload(uploadReq)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("✅ Upload completed!\n")
fmt.Printf("🆔 DataItem ID: %s\n", result.DataItemID)
fmt.Printf("📤 Bundler response: %s\n", result.BundlerResponse)
```

### Examples

See the `example/` directory for complete working examples:

- **`simple_upload.go`** - Basic upload and sign workflow
- **`server_only.go`** - Using the server package directly
- **`custom_config.go`** - Custom configuration and tags

**Note**: Examples use build tags to avoid conflicts. Use `-tags example` when running them.

Run any example with:

```bash
cd example
go run -tags example simple_upload.go
```

Build examples with `.build` extension:

```bash
cd example
go build -tags example -o simple_upload.build simple_upload.go
```

### Via Harlequin CLI

Integrate with the main Harlequin CLI:

```bash
harlequin remote-signing start --port 8080
```

### Upload Command Workflow

The `upload` command provides a complete file-to-signature workflow:

```bash
# Basic upload
./remote-signing upload ./my-file.json

# Upload to remote server
./remote-signing upload ./document.pdf --host signing.example.com --port 9000

# Upload without waiting for user to sign
./remote-signing upload ./data.bin --no-wait
```

**What happens:**

1. 🔍 Checks if a server is running (auto-starts one if needed)
2. 📁 Reads the specified file from disk
3. 🚀 Uploads the raw file data to the remote signing server
4. 🌐 Opens the signing URL in your default browser
5. ⏳ Waits for you to connect your wallet and sign the data
6. 🎉 Reports success when signing is complete
7. 🛑 Automatically stops the server if it was auto-started

**Smart Server Management:**

- Automatically detects if a server is already running
- Starts a temporary server if none is found
- Cleans up auto-started servers when upload completes
- In `--no-wait` mode, leaves auto-started servers running for signing

**WebSocket Integration:**

- Real-time status updates during the signing process
- Automatic completion detection when data item is signed
- Graceful handling of connection issues and user cancellation

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

## 🔌 Wallet Requirements

The signing interface requires a compatible Arweave wallet extension:

### Recommended Wallet

- **[Wander](https://wander.app/)** (formerly ArConnect) - Latest browser extension with full ANS-104 support

### Alternative Wallets

- **ArConnect** - Legacy extension (still supported)
- Any wallet implementing the Arweave wallet standard

### Required Permissions

The signing interface requests the following permissions:

- `ACCESS_ADDRESS` - Get wallet address for display
- `ACCESS_PUBLIC_KEY` - Required for data item creation
- `SIGN_TRANSACTION` - Sign data items
- `ACCESS_ALL_ADDRESSES` - Optional, for multi-address wallets

### Supported Features

- **ANS-104 Data Items** - Proper Arweave data item format
- **Automatic Tagging** - Adds metadata tags to signed items
- **Real-time Feedback** - Progress updates during signing
- **Error Handling** - User-friendly error messages

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

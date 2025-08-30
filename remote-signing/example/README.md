# Examples

This directory contains various examples demonstrating how to use the Harlequin Remote Signing Library.

## Quick Start

Make sure you have a Wander/ArConnect wallet extension installed in your browser before running any examples.

## Available Examples

### 1. Simple Upload (`simple_upload.go`)

Basic usage of the SigningServer API to upload and sign a file.

```bash
go run -tags example simple_upload.go
```

**What it does:**

- Reads `example.txt` and uploads it for signing
- Uses default configuration
- Automatically opens browser for wallet signing
- Uploads signed DataItem to ArDrive bundler
- Prints results including DataItem ID

### 2. Server Only (`server_only.go`)

Demonstrates using the server package directly without the SigningServer API.

```bash
go run -tags example server_only.go
```

**What it does:**

- Starts a standalone server for 30 seconds
- Shows server endpoints and test commands
- Useful for understanding the underlying server functionality

### 3. Custom Configuration (`custom_config.go`)

Shows how to use custom configuration, tags, target, and anchor.

```bash
go run -tags example custom_config.go
```

**What it does:**

- Uses custom server configuration (port 9090, 5MB limit, 15min timeout)
- Sets custom tags including timestamp
- Uses custom target address and anchor
- Demonstrates advanced configuration options

## Example Output

All examples will produce output similar to:

```
🚀 Starting upload and signing process...
📁 File: example.txt (234 bytes)
✅ Upload and signing completed successfully!
🆔 Request UUID: 7a9eb1a6-9abd-4888-8cfa-717af495a564
🆔 DataItem ID: rW-tFTlNnHrNDlJsh-BLTRyStHNg9PaPBd-gUI3Gyhk
🔗 Signing URL: http://localhost:8080/sign/7a9eb1a6-9abd-4888-8cfa-717af495a564
📅 Signed at: 2025-08-30 07:15:01
📤 Bundler response: {"id":"rW-tFTlNnHrNDlJsh-BLTRyStHNg9PaPBd-gUI3Gyhk",...}
```

## Prerequisites

- Go 1.19 or later
- Wander/ArConnect wallet extension installed in your browser
- Internet connection for bundler upload

## Customization

You can modify any example to:

- Change the file to sign
- Modify server configuration
- Add custom tags
- Set target addresses
- Provide custom anchors
- Change bundler endpoints

See the individual example files for detailed comments and configuration options.

## Quick Commands

Run any example with:

```bash
cd example
go run -tags example simple_upload.go
```

Build all examples with `.build` extension:

```bash
# From the main directory
./scripts/build.sh --examples

# Or manually
cd example
go build -tags example -o simple_upload.build simple_upload.go
go build -tags example -o server_only.build server_only.go
go build -tags example -o custom_config.build custom_config.go
```

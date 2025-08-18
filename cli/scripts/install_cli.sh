#!/bin/bash

set -e

# Configuration
BINARY_NAME="harlequin"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BASE_URL="https://"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect platform
OS="$(uname -s)"
ARCH="$(uname -m)"

case $OS in
    Darwin) PLATFORM="darwin" ;;
    Linux) PLATFORM="linux" ;;
    CYGWIN*|MINGW*|MSYS*) PLATFORM="windows" ;;
    *) error "Unsupported OS: $OS" ;;
esac

case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7*) ARCH="armv7" ;;
    i386|i686) ARCH="386" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

# Build download URL
BINARY_URL="${BASE_URL}${PLATFORM}_${ARCH}_cli_harlequin.daemongate.io"
BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"

if [ "$PLATFORM" = "windows" ]; then
    BINARY_PATH="${BINARY_PATH}.exe"
fi

info "Installing harlequin CLI for ${PLATFORM}/${ARCH}..."
info "Download URL: ${BINARY_URL}"

# Check if curl is available
if ! command -v curl >/dev/null 2>&1; then
    error "curl is required but not installed. Please install curl and try again."
fi

# Create temporary file
TEMP_FILE=$(mktemp)
trap 'rm -f "$TEMP_FILE"' EXIT

# Download binary
info "Downloading harlequin binary..."
if ! curl -fsSL "$BINARY_URL" -o "$TEMP_FILE"; then
    error "Failed to download binary from $BINARY_URL"
fi

# Verify download
if [ ! -s "$TEMP_FILE" ]; then
    error "Downloaded file is empty or corrupted"
fi

# Check if we need sudo for installation
if [ ! -w "$(dirname "$INSTALL_DIR")" ]; then
    warn "Installing to $INSTALL_DIR requires sudo privileges"
    SUDO="sudo"
else
    SUDO=""
fi

# Make executable and install
chmod +x "$TEMP_FILE"

if ! $SUDO mv "$TEMP_FILE" "$BINARY_PATH"; then
    error "Failed to install binary to $BINARY_PATH"
fi

# Verify installation
if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
    warn "harlequin was installed to $BINARY_PATH but is not in your PATH"
    warn "You may need to add $INSTALL_DIR to your PATH or run:"
    warn "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo
    warn "Or run harlequin directly: $BINARY_PATH"
else
    info "✅ harlequin successfully installed!"
    info "Run 'harlequin --help' to get started"
fi

# Show version if possible
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    echo
    info "Installed version:"
    "$BINARY_NAME" --version 2>/dev/null || "$BINARY_NAME" --help | head -1 || true
fi

echo
info "🎭 Welcome to Harlequin - Arweave Development Toolkit!"
info "For documentation and examples, visit: https://github.com/the-permaweb-harlequin/harlequin-toolkit"

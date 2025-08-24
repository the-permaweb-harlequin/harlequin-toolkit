#!/bin/bash

set -e

# Configuration
BINARY_NAME="harlequin"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BASE_URL="${BASE_URL:-https://install_cli_harlequin.daemongate.io}"
VERSION="${VERSION:-latest}"
FORCE="${FORCE:-false}"
DRYRUN="${DRYRUN:-false}"

# Parse command line arguments
for arg in "$@"; do
    case $arg in
        --dryrun|--dry-run)
            DRYRUN="true"
            shift
            ;;
        --force)
            FORCE="true"
            shift
            ;;
        --version=*)
            VERSION="${arg#*=}"
            shift
            ;;
        --install-dir=*)
            INSTALL_DIR="${arg#*=}"
            shift
            ;;
        --help|-h)
            echo "Harlequin CLI Installer"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --dryrun, --dry-run    Simulate installation without downloading"
            echo "  --force                Force reinstall even if already installed"
            echo "  --version=VERSION      Install specific version (default: latest)"
            echo "  --install-dir=DIR      Install to custom directory (default: /usr/local/bin)"
            echo "  --help, -h             Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  DRYRUN=true           Same as --dryrun"
            echo "  FORCE=true            Same as --force"
            echo "  VERSION=x.y.z         Same as --version=x.y.z (supports alpha: 1.0.0-alpha.1)"
            echo "  INSTALL_DIR=/path     Same as --install-dir=/path"
            echo ""
            echo "Examples:"
            echo "  # Latest stable version"
            echo "  curl -fsSL https://install_cli_harlequin.daemongate.io | sh"
            echo "  "
            echo "  # Specific stable version"
            echo "  curl -fsSL https://install_cli_harlequin.daemongate.io | VERSION=1.2.3 sh"
            echo "  "
            echo "  # Alpha version (bleeding edge)"
            echo "  curl -fsSL https://install_cli_harlequin.daemongate.io | VERSION=1.2.3-alpha.1 sh"
            echo "  "
            echo "  # Dry run"
            echo "  curl -fsSL https://install_cli_harlequin.daemongate.io | DRYRUN=true sh"
            exit 0
            ;;
        *)
            # Unknown option
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    local prefix="[INFO]"
    if [ "$DRYRUN" = "true" ]; then
        prefix="[DRYRUN]"
    fi
    echo -e "${GREEN}${prefix}${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    if [ "$DRYRUN" = "true" ]; then
        echo -e "${YELLOW}[DRYRUN]${NC} Exiting (this would be a real error)"
    fi
    exit 1
}

prompt() {
    echo -e "${BLUE}[PROMPT]${NC} $1"
}

dryrun_info() {
    if [ "$DRYRUN" = "true" ]; then
        echo -e "${YELLOW}[DRYRUN]${NC} $1"
    fi
}

# Check if jq is available for JSON parsing
has_jq() {
    command -v jq >/dev/null 2>&1
}

# Parse JSON without jq (basic version parsing)
parse_version() {
    grep -o '"tag_name":"[^"]*"' | cut -d'"' -f4 | sed 's/^v//'
}

# Get current installed version
get_installed_version() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        "$BINARY_NAME" --version 2>/dev/null | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+' | head -1 || echo ""
    else
        echo ""
    fi
}

# Fetch available versions
get_available_versions() {
    info "Fetching available versions..."
    
    if [ "$DRYRUN" = "true" ]; then
        dryrun_info "Would fetch versions from ${BASE_URL}/releases"
        # Return mock versions for dry run
        echo -e "1.2.3\n1.2.2\n1.2.1\n1.1.0"
        return
    fi
    
    if ! curl -fsSL "${BASE_URL}/releases" -o /tmp/releases.json 2>/dev/null; then
        warn "Failed to fetch version information, using 'latest'"
        echo "latest"
        return
    fi
    
    if has_jq; then
        jq -r '.[].tag_name' /tmp/releases.json 2>/dev/null | sed 's/^v//' || echo "latest"
    else
        parse_version < /tmp/releases.json || echo "latest"
    fi
}

# Compare versions (basic semver comparison)
version_gt() {
    test "$(printf '%s\n' "$@" | sort -V | head -n 1)" != "$1"
}

# Interactive version selection
select_version() {
    local versions
    versions=$(get_available_versions)
    
    if [ "$versions" = "latest" ]; then
        echo "latest"
        return
    fi
    
    echo
    prompt "Available versions:"
    echo "$versions" | head -10 | nl -w2 -s') '
    
    if [ "$(echo "$versions" | wc -l)" -gt 10 ]; then
        echo "    ... and more"
    fi
    
    echo
    echo "Enter version number (1-10), version string (e.g., 1.2.3), or press Enter for latest:"
    read -r choice
    
    if [ -z "$choice" ]; then
        echo "latest"
    elif echo "$choice" | grep -q '^[0-9]\+$' && [ "$choice" -le 10 ]; then
        echo "$versions" | sed -n "${choice}p"
    else
        # Validate if it's a valid version string
        if echo "$versions" | grep -q "^${choice}$"; then
            echo "$choice"
        else
            warn "Invalid version '$choice', using latest"
            echo "latest"
        fi
    fi
}

# Show dry run banner
if [ "$DRYRUN" = "true" ]; then
    echo
    echo -e "${YELLOW}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║                                 DRY RUN MODE                                 ║${NC}"
    echo -e "${YELLOW}║                        No files will be downloaded or installed             ║${NC}"
    echo -e "${YELLOW}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo
fi

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
    # Windows ARM64 detection
    *aarch64*) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

# Check for existing installation and offer upgrade
CURRENT_VERSION=$(get_installed_version)
if [ -n "$CURRENT_VERSION" ] && [ "$FORCE" != "true" ]; then
    info "🎭 Harlequin v${CURRENT_VERSION} is already installed"
    
    # Check if latest version is available
    LATEST_VERSION=$(get_available_versions | head -1)
    if [ "$LATEST_VERSION" != "latest" ] && [ "$LATEST_VERSION" != "$CURRENT_VERSION" ]; then
        if version_gt "$LATEST_VERSION" "$CURRENT_VERSION"; then
            echo
            prompt "A newer version (v${LATEST_VERSION}) is available. Would you like to upgrade? [y/N]"
            read -r upgrade_choice
            if [ "$upgrade_choice" = "y" ] || [ "$upgrade_choice" = "Y" ]; then
                info "Upgrading to v${LATEST_VERSION}..."
                VERSION="$LATEST_VERSION"
            else
                info "Keeping current version v${CURRENT_VERSION}"
                exit 0
            fi
        else
            info "You have the latest version installed!"
            exit 0
        fi
    else
        echo
        prompt "Would you like to:"
        echo "1) Keep current version (v${CURRENT_VERSION})"
        echo "2) Reinstall current version"
        echo "3) Select a different version"
        echo
        echo "Enter choice [1-3] or press Enter to keep current:"
        read -r choice
        
        case $choice in
            2) info "Reinstalling v${CURRENT_VERSION}..." ;;
            3) VERSION=$(select_version) ;;
            *) info "Keeping current version v${CURRENT_VERSION}"; exit 0 ;;
        esac
    fi
else
    # Fresh installation - offer version selection
    if [ "$VERSION" = "latest" ]; then
        echo
        prompt "Would you like to select a specific version? [y/N]"
        read -r version_choice
        if [ "$version_choice" = "y" ] || [ "$version_choice" = "Y" ]; then
            VERSION=$(select_version)
        fi
    fi
fi

# Build download URL with new structure
BINARY_URL="${BASE_URL}/releases/${VERSION}/${PLATFORM}/${ARCH}"
BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"

if [ "$PLATFORM" = "windows" ]; then
    BINARY_PATH="${BINARY_PATH}.exe"
fi

info "Installing harlequin CLI v${VERSION} for ${PLATFORM}/${ARCH}..."
info "Download URL: ${BINARY_URL}"

# Check if required tools are available
if ! command -v curl >/dev/null 2>&1; then
    error "curl is required but not installed. Please install curl and try again."
fi

if ! command -v gzip >/dev/null 2>&1; then
    error "gzip is required but not installed. Please install gzip and try again."
fi

# Download binary
info "Downloading compressed harlequin binary..."

if [ "$DRYRUN" = "true" ]; then
    dryrun_info "Would download compressed binary from: $BINARY_URL"
    dryrun_info "Would create temporary file for download"
    dryrun_info "Would decompress gzipped binary"
    TEMP_FILE="/tmp/harlequin-dryrun"
    # Create a fake file for dry run
    echo "fake binary content" > "$TEMP_FILE"
else
    # Create temporary files
    TEMP_COMPRESSED=$(mktemp)
    TEMP_FILE=$(mktemp)
    trap 'rm -f "$TEMP_COMPRESSED" "$TEMP_FILE"' EXIT

    # Download compressed binary
    if ! curl -fsSL "$BINARY_URL" -o "$TEMP_COMPRESSED"; then
        error "Failed to download compressed binary from $BINARY_URL"
    fi

    # Verify download
    if [ ! -s "$TEMP_COMPRESSED" ]; then
        error "Downloaded file is empty or corrupted"
    fi

    # Decompress the binary
    info "Decompressing binary..."
    if ! gzip -dc "$TEMP_COMPRESSED" > "$TEMP_FILE"; then
        error "Failed to decompress binary. File may be corrupted."
    fi

    # Verify decompressed file
    if [ ! -s "$TEMP_FILE" ]; then
        error "Decompressed file is empty or corrupted"
    fi

    # Show compression stats
    COMPRESSED_SIZE=$(wc -c < "$TEMP_COMPRESSED")
    DECOMPRESSED_SIZE=$(wc -c < "$TEMP_FILE")
    COMPRESSION_RATIO=$(( (COMPRESSED_SIZE * 100) / DECOMPRESSED_SIZE ))
    info "Downloaded: $(( COMPRESSED_SIZE / 1024 ))KB compressed → $(( DECOMPRESSED_SIZE / 1024 ))KB decompressed (${COMPRESSION_RATIO}% of original)"
fi

# Check if we need sudo for installation
if [ "$DRYRUN" = "true" ]; then
    dryrun_info "Would check write permissions for: $(dirname "$INSTALL_DIR")"
    if [ ! -w "$(dirname "$INSTALL_DIR")" ]; then
        dryrun_info "Would require sudo privileges for installation"
        SUDO="sudo"
    else
        dryrun_info "Would install without sudo privileges"
        SUDO=""
    fi
else
    if [ ! -w "$(dirname "$INSTALL_DIR")" ]; then
        warn "Installing to $INSTALL_DIR requires sudo privileges"
        SUDO="sudo"
    else
        SUDO=""
    fi
fi

# Make executable and install
if [ "$DRYRUN" = "true" ]; then
    dryrun_info "Would make binary executable: chmod +x $TEMP_FILE"
    dryrun_info "Would install binary: $SUDO mv $TEMP_FILE $BINARY_PATH"
    dryrun_info "Binary would be installed to: $BINARY_PATH"
else
    chmod +x "$TEMP_FILE"

    if ! $SUDO mv "$TEMP_FILE" "$BINARY_PATH"; then
        error "Failed to install binary to $BINARY_PATH"
    fi
fi

# Verify installation
if [ "$DRYRUN" = "true" ]; then
    dryrun_info "Would verify installation by checking command availability"
    dryrun_info "✅ harlequin would be successfully installed!"
    dryrun_info "Would be available at: $BINARY_PATH"
else
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
fi

# Show version if possible
if [ "$DRYRUN" = "true" ]; then
    echo
    dryrun_info "Would display installed version information"
    dryrun_info "Would run: $BINARY_NAME --version"
else
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        echo
        info "Installed version:"
        "$BINARY_NAME" --version 2>/dev/null || "$BINARY_NAME" --help | head -1 || true
    fi
fi

# Cleanup
if [ "$DRYRUN" = "true" ]; then
    dryrun_info "Would clean up temporary files"
    rm -f "$TEMP_FILE" /tmp/releases.json 2>/dev/null || true
else
    rm -f /tmp/releases.json
    # TEMP_COMPRESSED and TEMP_FILE are cleaned up by trap
fi

echo
if [ "$DRYRUN" = "true" ]; then
    echo -e "${YELLOW}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║                            DRY RUN COMPLETED                                 ║${NC}"
    echo -e "${YELLOW}║                    This was a simulation - no changes made                   ║${NC}"
    echo -e "${YELLOW}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo
    dryrun_info "To perform actual installation, run without --dryrun flag"
    echo
fi

info "🎭 Welcome to Harlequin - Arweave Development Toolkit!"
info "For documentation and examples, visit: https://github.com/the-permaweb-harlequin/harlequin-toolkit"
echo
info "Usage examples:"
info "  harlequin build                    # Interactive TUI"
info "  harlequin build --debug            # Interactive TUI with debug logging"
info "  harlequin build ./my-project       # Legacy CLI mode"
echo
if [ "$DRYRUN" != "true" ]; then
    info "To upgrade in the future, simply run the install script again!"
    info "To force reinstall: curl -fsSL https://install_cli_harlequin.daemongate.io | FORCE=true sh"
fi
info "To simulate installation: curl -fsSL https://install_cli_harlequin.daemongate.io | sh -s -- --dryrun"


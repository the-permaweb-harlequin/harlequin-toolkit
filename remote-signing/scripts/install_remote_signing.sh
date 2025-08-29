#!/bin/bash

# Harlequin Remote Signing Server Installation Script
# https://github.com/the-permaweb-harlequin/harlequin-toolkit

set -e

# Configuration
APP_NAME="harlequin-remote-signing"
BINARY_NAME="harlequin-remote-signing"
GITHUB_REPO="the-permaweb-harlequin/harlequin-toolkit"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
RELEASES_API_URL="https://remote_signing_harlequin.arweave.dev/releases"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Detect platform and architecture
detect_platform() {
    local platform=""
    local arch=""

    # Detect OS
    case "$(uname -s)" in
        Linux*)  platform="linux";;
        Darwin*) platform="darwin";;
        CYGWIN*|MINGW*|MSYS*) platform="windows";;
        *)
            echo -e "${RED}❌ Unsupported operating system: $(uname -s)${NC}"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64) arch="amd64";;
        arm64|aarch64) arch="arm64";;
        *)
            echo -e "${RED}❌ Unsupported architecture: $(uname -m)${NC}"
            exit 1
            ;;
    esac

    echo "${platform}_${arch}"
}

# Download and extract binary
install_binary() {
    local platform_arch="$1"
    local version="${VERSION:-latest}"
    local install_dir="$2"

    # Parse platform and arch
    local platform="${platform_arch%_*}"
    local arch="${platform_arch#*_}"

    echo -e "${BLUE}🔐 Installing Harlequin Remote Signing Server...${NC}"
    echo -e "${CYAN}   Platform: ${platform}${NC}"
    echo -e "${CYAN}   Architecture: ${arch}${NC}"
    echo -e "${CYAN}   Version: ${version}${NC}"
    echo -e "${CYAN}   Install directory: ${install_dir}${NC}"
    echo

    # Create install directory
    mkdir -p "$install_dir"

    # Construct download URL
    local download_url="${RELEASES_API_URL}/${version}/${platform}/${arch}"
    local binary_path="${install_dir}/${BINARY_NAME}"

    # Add .exe extension for Windows
    if [ "$platform" = "windows" ]; then
        binary_path="${binary_path}.exe"
    fi

    # Download binary
    echo -e "${YELLOW}📥 Downloading from: ${download_url}${NC}"

    # Check if we should use curl or wget
    if command -v curl &> /dev/null; then
        if [ "${DRYRUN:-}" = "true" ]; then
            echo -e "${YELLOW}[DRYRUN] Would download with: curl -fsSL \"${download_url}\" | gunzip > \"${binary_path}\"${NC}"
        else
            curl -fsSL "$download_url" | gunzip > "$binary_path"
        fi
    elif command -v wget &> /dev/null; then
        if [ "${DRYRUN:-}" = "true" ]; then
            echo -e "${YELLOW}[DRYRUN] Would download with: wget -qO- \"${download_url}\" | gunzip > \"${binary_path}\"${NC}"
        else
            wget -qO- "$download_url" | gunzip > "$binary_path"
        fi
    else
        echo -e "${RED}❌ Neither curl nor wget is available. Please install one of them.${NC}"
        exit 1
    fi

    if [ "${DRYRUN:-}" != "true" ]; then
        # Make binary executable
        chmod +x "$binary_path"

        # Verify installation
        if [ -x "$binary_path" ]; then
            echo -e "${GREEN}✅ Successfully installed ${APP_NAME}${NC}"
            echo -e "${CYAN}   Binary location: ${binary_path}${NC}"

            # Check if install directory is in PATH
            if [[ ":$PATH:" != *":$install_dir:"* ]]; then
                echo
                echo -e "${YELLOW}⚠️  ${install_dir} is not in your PATH${NC}"
                echo -e "${CYAN}   Add it to your shell profile:${NC}"
                echo -e "${CYAN}   export PATH=\"\$PATH:${install_dir}\"${NC}"
                echo
            fi

            # Show version
            echo -e "${PURPLE}🔍 Version information:${NC}"
            if "$binary_path" --version 2>/dev/null || "$binary_path" version 2>/dev/null; then
                true # Version command succeeded
            else
                echo -e "${CYAN}   Binary installed successfully (version command not available)${NC}"
            fi
        else
            echo -e "${RED}❌ Installation failed: Binary is not executable${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}[DRYRUN] Would make binary executable and verify installation${NC}"
        echo -e "${GREEN}✅ [DRYRUN] Installation simulation completed${NC}"
    fi
}

# Show usage information
show_usage() {
    echo -e "${BLUE}🔐 Harlequin Remote Signing Server Installation${NC}"
    echo
    echo -e "${CYAN}Usage:${NC}"
    echo "  curl -fsSL https://remote_signing_harlequin.arweave.dev | sh"
    echo
    echo -e "${CYAN}Environment Variables:${NC}"
    echo "  VERSION      Specific version to install (default: latest)"
    echo "  INSTALL_DIR  Installation directory (default: \$HOME/.local/bin)"
    echo "  DRYRUN       Set to 'true' for simulation mode"
    echo
    echo -e "${CYAN}Examples:${NC}"
    echo "  # Install latest version"
    echo "  curl -fsSL https://remote_signing_harlequin.arweave.dev | sh"
    echo
    echo "  # Install specific version"
    echo "  curl -fsSL https://remote_signing_harlequin.arweave.dev | VERSION=1.0.0 sh"
    echo
    echo "  # Install to custom directory"
    echo "  curl -fsSL https://remote_signing_harlequin.arweave.dev | INSTALL_DIR=/usr/local/bin sh"
    echo
    echo "  # Dry run"
    echo "  curl -fsSL https://remote_signing_harlequin.arweave.dev | DRYRUN=true sh"
}

# Main installation function
main() {
    # Handle help flag
    if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
        show_usage
        exit 0
    fi

    # Detect platform
    local platform_arch
    platform_arch="$(detect_platform)"

    # Install binary
    install_binary "$platform_arch" "$INSTALL_DIR"

    # Show next steps
    if [ "${DRYRUN:-}" != "true" ]; then
        echo
        echo -e "${GREEN}🎉 Installation completed!${NC}"
        echo
        echo -e "${CYAN}Next steps:${NC}"
        echo -e "${CYAN}1. Start the server: ${BINARY_NAME} start${NC}"
        echo -e "${CYAN}2. Check status: ${BINARY_NAME} status${NC}"
        echo -e "${CYAN}3. Stop the server: ${BINARY_NAME} stop${NC}"
        echo
        echo -e "${CYAN}For more information:${NC}"
        echo -e "${CYAN}  ${BINARY_NAME} --help${NC}"
        echo -e "${CYAN}  https://github.com/${GITHUB_REPO}${NC}"
    fi
}

# Run main function
main "$@"

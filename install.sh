#!/usr/bin/env bash
# install.sh - Universal installer for Duct
# Usage: curl -fsSL https://get.duct.dev | sh
#        or: curl -fsSL https://raw.githubusercontent.com/nmvinicius/duct/main/install.sh | sh

set -euo pipefail

REPO="nmvinicius/duct"
VERSION="${VERSION:-stable}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="duct"

# Colors
C_RESET='\033[0m'
C_GREEN='\033[0;32m'
C_YELLOW='\033[1;33m'
C_RED='\033[0;31m'
C_BLUE='\033[0;34m'

log()   { echo -e "${C_GREEN}[install]${C_RESET} $1"; }
warn()  { echo -e "${C_YELLOW}[install]${C_RESET} $1"; }
error() { echo -e "${C_RED}[install]${C_RESET} $1" >&2; }
info()  { echo -e "${C_BLUE}[install]${C_RESET} $1"; }

# Detect OS and architecture
detect_platform() {
    local os arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux) os="linux" ;;
        darwin) os="darwin" ;;
        mingw*|msys*|cygwin*) os="windows" ;;
        *) error "Unsupported OS: $os"; exit 1 ;;
    esac

    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) error "Unsupported architecture: $arch"; exit 1 ;;
    esac

    echo "${os}-${arch}"
}

# Download with fallback
download() {
    local url="$1"
    local output="$2"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "curl or wget is required"
        exit 1
    fi
}

main() {
    info "Duct Installer"
    info "=============="

    PLATFORM=$(detect_platform)
    log "Detected platform: $PLATFORM"

    # Determine download URL
    if [[ "$VERSION" == "stable" ]]; then
        # Get latest stable release
        LATEST_URL="https://api.github.com/repos/${REPO}/releases/latest"
        if command -v curl >/dev/null 2>&1; then
            VERSION=$(curl -fsSL "$LATEST_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        fi
        VERSION="${VERSION:-v0.1.0}"
    fi

    # Remove 'v' prefix if present
    VERSION="${VERSION#v}"

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY_NAME}-${PLATFORM}"
    if [[ "$PLATFORM" == windows-* ]]; then
        DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
    fi

    log "Version: $VERSION"
    log "Download: $DOWNLOAD_URL"

    # Download
    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY_NAME"

    info "Downloading..."
    if ! download "$DOWNLOAD_URL" "$TMP_FILE"; then
        # Fallback: try to build from source
        warn "Binary download failed, trying to build from source..."
        if ! command -v go >/dev/null 2>&1; then
            error "Go is required to build from source"
            error "Please install Go or download manually from https://github.com/$REPO/releases"
            exit 1
        fi

        git clone --branch "v$VERSION" --depth 1 "https://github.com/$REPO.git" "$TMP_DIR/repo"
        cd "$TMP_DIR/repo"
        go build -o "$TMP_FILE" ./cmd/duct
    fi

    chmod +x "$TMP_FILE"

    # Install
    log "Installing to $INSTALL_DIR..."
    if [[ -w "$INSTALL_DIR" ]]; then
        mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        warn "Need sudo for $INSTALL_DIR"
        sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    # Verify
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        INSTALLED_VERSION=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        log "Installed: $INSTALLED_VERSION"
        log "Run 'duct --help' to get started"
    else
        warn "Installed but not in PATH"
        warn "Add $INSTALL_DIR to your PATH or run: export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

main "$@"
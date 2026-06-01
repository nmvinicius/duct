#!/usr/bin/env bash
# scripts/tools/node.sh - Setup Node.js environment

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

NODE_VERSION="${NODE_VERSION:-20}"

duct_info "Setting up Node.js ${NODE_VERSION}..."

if has_cmd node; then
    CURRENT=$(node --version | sed 's/v//')
    duct_log "Node.js already installed: v${CURRENT}"
    exit 0
fi

if has_cmd nvm; then
    nvm install "$NODE_VERSION"
    nvm use "$NODE_VERSION"
elif [[ -f "$HOME/.nvm/nvm.sh" ]]; then
    source "$HOME/.nvm/nvm.sh"
    nvm install "$NODE_VERSION"
    nvm use "$NODE_VERSION"
elif has_cmd apt-get; then
    # Debian/Ubuntu
    curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash -
    apt-get install -y nodejs
elif has_cmd apk; then
    # Alpine
    apk add --no-cache nodejs npm
elif has_cmd brew; then
    brew install node@"$NODE_VERSION"
else
    duct_error "Cannot install Node.js automatically. Please install Node ${NODE_VERSION} manually."
    exit 1
fi

duct_log "Node.js $(node --version) ready"
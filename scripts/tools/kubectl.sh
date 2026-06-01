#!/usr/bin/env bash
# scripts/tools/kubectl.sh - Setup kubectl environment

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

duct_info "Checking kubectl..."

if has_cmd kubectl; then
    duct_log "kubectl $(kubectl version --client -o json 2>/dev/null | grep -o '"gitVersion": "[^"]*"' | cut -d'"' -f4 || echo 'unknown') ready"
    exit 0
fi

# Try to install
if has_cmd curl; then
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    mv kubectl /usr/local/bin/ 2>/dev/null || {
        duct_warn "Cannot install to /usr/local/bin, using ./kubectl"
    }
else
    duct_error "kubectl not found and curl not available for installation"
    exit 1
fi

duct_log "kubectl installed"
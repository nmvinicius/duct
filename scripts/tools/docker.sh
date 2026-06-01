#!/usr/bin/env bash
# scripts/tools/docker.sh - Setup Docker environment

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

duct_info "Checking Docker..."

if ! has_cmd docker; then
    duct_error "Docker not found. Please install Docker."
    exit 1
fi

# Test Docker daemon
if ! docker info >/dev/null 2>&1; then
    duct_error "Docker daemon not running or not accessible"
    exit 1
fi

duct_log "Docker $(docker --version) ready"
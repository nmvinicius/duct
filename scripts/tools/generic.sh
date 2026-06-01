#!/usr/bin/env bash
# scripts/tools/generic.sh - Generic tool setup fallback
# Usage: generic.sh <tool-name>

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

TOOL="${1:-}"

if [[ -z "$TOOL" ]]; then
    duct_error "No tool specified"
    exit 1
fi

duct_info "Generic setup for: $TOOL"

if has_cmd "$TOOL"; then
    duct_log "$TOOL already available"
    exit 0
fi

duct_warn "No specific setup script for $TOOL"
duct_warn "Please ensure $TOOL is installed in your environment"
exit 1
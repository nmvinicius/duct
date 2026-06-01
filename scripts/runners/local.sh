#!/usr/bin/env bash
# scripts/runners/local.sh - Local execution runner
# This script is sourced by the Go executor when running locally

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

duct_log "Local runner initialized"

# Resolve git variables from local repo
export GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo 'unknown')}"
export GIT_BRANCH="${GIT_BRANCH:-$(git branch --show-current 2>/dev/null || echo 'unknown')}"
export GIT_TAG="${GIT_TAG:-$(git describe --tags --exact-match 2>/dev/null || echo '')}"

duct_info "Git: $GIT_COMMIT @ $GIT_BRANCH (tag: ${GIT_TAG:-none})"
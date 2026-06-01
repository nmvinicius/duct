#!/usr/bin/env bash
# scripts/tools/commit-lint.sh - Valida mensagem de commit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

COMMIT_MSG="${1:-}"
if [[ -z "$COMMIT_MSG" ]]; then
    duct_error "Usage: commit-lint.sh <commit-message>"
    exit 1
fi

# Regex Conventional Commit
PATTERN="^(feat|fix|chore|docs|refactor|test|ci)(\([a-z-]+\))?: .+"

if ! echo "$COMMIT_MSG" | grep -qE "$PATTERN"; then
    duct_error "Invalid commit message format"
    duct_info "Expected: <type>(<scope>): <description>"
    duct_info "Types: feat, fix, chore, docs, refactor, test, ci"
    exit 1
fi

# Extrai tipo
TYPE=$(echo "$COMMIT_MSG" | sed -E 's/^(feat|fix|chore|docs|refactor|test|ci).*/\1/')

# Valida BREAKING CHANGE
if echo "$COMMIT_MSG" | grep -q "BREAKING CHANGE:"; then
    duct_warn "BREAKING CHANGE detected - this will bump MAJOR version"
fi

duct_log "Commit message valid: $TYPE"
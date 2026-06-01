#!/usr/bin/env bash
# scripts/tools/commit-lint.sh - Valida mensagem de commit (Conventional Commits)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

COMMIT_MSG="${1:-}"
if [[ -z "$COMMIT_MSG" ]]; then
    duct_error "Usage: commit-lint.sh <commit-message>"
    exit 1
fi

# Regex: tipo(opcional): descrição
PATTERN="^(feat|fix|chore|docs|refactor|test|ci|build|perf)(\([a-z0-9-]+\))?!?: .+"

if ! echo "$COMMIT_MSG" | grep -qE "$PATTERN"; then
    duct_error "Invalid commit message format"
    duct_info ""
    duct_info "Expected format:"
    duct_info "  <type>[(<scope>)][!]: <description>"
    duct_info ""
    duct_info "Types:"
    duct_info "  feat     - New feature (bumps MINOR)"
    duct_info "  fix      - Bug fix (bumps PATCH)"
    duct_info "  docs     - Documentation only"
    duct_info "  refactor - Code refactoring"
    duct_info "  test     - Adding/updating tests"
    duct_info "  chore    - Maintenance tasks"
    duct_info "  ci       - CI/CD changes"
    duct_info "  build    - Build system changes"
    duct_info "  perf     - Performance improvements"
    duct_info ""
    duct_info "Examples:"
    duct_info "  feat(parser): add support for WHEN conditions"
    duct_info "  fix(executor): resolve race condition in step runner"
    duct_info "  docs: update README with local setup instructions"
    duct_info "  BREAKING CHANGE: rename STEP to TASK"
    exit 1
fi

TYPE=$(echo "$COMMIT_MSG" | sed -E 's/^(feat|fix|chore|docs|refactor|test|ci|build|perf).*/\1/')

# Detecta BREAKING CHANGE
if echo "$COMMIT_MSG" | grep -qE "^[a-z]+(\([^)]+\))?!:" || \
   echo "$COMMIT_MSG" | grep -q "BREAKING CHANGE:"; then
    duct_warn "BREAKING CHANGE detected - will bump MAJOR version"
fi

duct_log "Commit message valid: $TYPE"
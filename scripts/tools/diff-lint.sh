#!/usr/bin/env bash
# scripts/tools/diff-lint.sh - Valida coerência entre diff e commit

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

COMMIT_MSG="${1:-}"
if [[ -z "$COMMIT_MSG" ]]; then
    duct_error "Usage: diff-lint.sh <commit-message>"
    exit 1
fi

# Estatísticas do diff (staged)
FILES_CHANGED=$(git diff --cached --name-only 2>/dev/null | wc -l)
LINES_CHANGED=$(git diff --cached --numstat 2>/dev/null | awk '{sum+=$1+$2} END {print sum}')
TYPE=$(echo "$COMMIT_MSG" | sed -E 's/^(feat|fix|chore|docs|refactor|test|ci|build|perf).*/\1/')

if [[ -z "$TYPE" ]]; then
    duct_error "Could not detect commit type"
    exit 1
fi

duct_info "Diff stats: $FILES_CHANGED files, ${LINES_CHANGED:-0} lines"
duct_info "Commit type: $TYPE"

# Thresholds por tipo
case "$TYPE" in
    fix)
        MAX_FILES=5
        MAX_LINES=100
        ;;
    feat)
        MAX_FILES=15
        MAX_LINES=500
        ;;
    docs|test|ci|chore|build|perf)
        MAX_FILES=999
        MAX_LINES=999
        ;;
    *)
        MAX_FILES=10
        MAX_LINES=200
        ;;
esac

WARNINGS=0

if (( FILES_CHANGED > MAX_FILES )); then
    duct_warn "$TYPE should change ≤ $MAX_FILES files, got $FILES_CHANGED"
    ((WARNINGS++))
fi

if [[ -n "$LINES_CHANGED" && "$LINES_CHANGED" -gt "$MAX_LINES" ]]; then
    duct_warn "$TYPE should change ≤ $MAX_LINES lines, got $LINES_CHANGED"
    ((WARNINGS++))
fi

# Validação de scope (se presente)
if echo "$COMMIT_MSG" | grep -qE '^\w+\('; then
    SCOPE=$(echo "$COMMIT_MSG" | sed -E 's/^\w+\(([^)]+)\).*/\1/')
    SCOPE_FILES=$(git diff --cached --name-only 2>/dev/null | grep -c "^$SCOPE/" || true)
    
    if (( SCOPE_FILES == 0 )); then
        duct_warn "Scope '$SCOPE' not found in changed files"
        ((WARNINGS++))
    fi
fi

if (( WARNINGS > 0 )); then
    duct_warn "$WARNINGS warning(s) - consider splitting this commit"
    # Não falha, só avisa. Para strict mode, descomente:
    # exit 1
fi

duct_log "Diff validation complete"
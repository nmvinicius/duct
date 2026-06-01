#!/usr/bin/env bash
# scripts/tools/version-bump.sh - Calcula próxima versão semântica

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

# Pega última tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
LAST_VERSION=${LAST_TAG#v}

MAJOR=$(echo "$LAST_VERSION" | cut -d. -f1)
MINOR=$(echo "$LAST_VERSION" | cut -d. -f2)
PATCH=$(echo "$LAST_VERSION" | cut -d. -f3)

duct_info "Last version: $LAST_TAG"

# Analisa commits desde última tag
COMMITS=$(git log "$LAST_TAG"..HEAD --pretty=%B 2>/dev/null || echo "")

if [[ -z "$COMMITS" ]]; then
    duct_error "No commits found since $LAST_TAG"
    exit 1
fi

# Detecta tipo de bump
BUMP="patch"

if echo "$COMMITS" | grep -qE "^[a-z]+(\([^)]+\))?!:" || \
   echo "$COMMITS" | grep -q "BREAKING CHANGE:"; then
    BUMP="major"
    duct_info "Breaking change detected"
elif echo "$COMMITS" | grep -qE "^feat(\(|:)"; then
    BUMP="minor"
    duct_info "Feature commit detected"
else
    duct_info "Patch-level changes only"
fi

# Calcula nova versão
case "$BUMP" in
    major)
        NEW_MAJOR=$((MAJOR + 1))
        NEW_VERSION="${NEW_MAJOR}.0.0"
        ;;
    minor)
        NEW_MINOR=$((MINOR + 1))
        NEW_VERSION="${MAJOR}.${NEW_MINOR}.0"
        ;;
    patch)
        NEW_PATCH=$((PATCH + 1))
        NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
        ;;
esac

echo "v$NEW_VERSION" > .duct-version
echo "$BUMP" > .duct-bump-type

duct_log "Version bump: $LAST_TAG → v$NEW_VERSION ($BUMP)"
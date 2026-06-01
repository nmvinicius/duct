#!/usr/bin/env bash
# scripts/tools/changelog-gen.sh - Gera CHANGELOG.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

VERSION=$(cat .duct-version 2>/dev/null || echo "v0.0.0")
DATE=$(date +%Y-%m-%d)

# Pega commits desde última tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [[ -n "$LAST_TAG" ]]; then
    RANGE="$LAST_TAG..HEAD"
else
    RANGE="HEAD"
fi

# Categoriza commits
FEATS=$(git log "$RANGE" --pretty=format:"- %s" --grep="^feat" || true)
FIXES=$(git log "$RANGE" --pretty=format:"- %s" --grep="^fix" || true)
CHORES=$(git log "$RANGE" --pretty=format:"- %s" --grep="^chore" || true)
DOCS=$(git log "$RANGE" --pretty=format:"- %s" --grep="^docs" || true)
BREAKING=$(git log "$RANGE" --pretty=format:"- %s" --grep="BREAKING CHANGE" || true)

# Gera entrada
ENTRY="## [$VERSION] - $DATE

"

if [[ -n "$BREAKING" ]]; then
    ENTRY+="### ⚠️ Breaking Changes
$BREAKING

"
fi

if [[ -n "$FEATS" ]]; then
    ENTRY+="### 🚀 Features
$FEATS

"
fi

if [[ -n "$FIXES" ]]; then
    ENTRY+="### 🐛 Fixes
$FIXES

"
fi

if [[ -n "$DOCS" ]]; then
    ENTRY+="### 📚 Documentation
$DOCS

"
fi

if [[ -n "$CHORES" ]]; then
    ENTRY+="### 🔧 Maintenance
$CHORES

"
fi

# Insere no topo do CHANGELOG
if [[ -f CHANGELOG.md ]]; then
    # Remove header antiga se existir
    tail -n +3 CHANGELOG.md > .changelog-tmp
    echo -e "# Changelog\n\n$ENTRY$(cat .changelog-tmp)" > CHANGELOG.md
    rm .changelog-tmp
else
    echo -e "# Changelog\n\n$ENTRY" > CHANGELOG.md
fi

duct_log "CHANGELOG.md updated for $VERSION"
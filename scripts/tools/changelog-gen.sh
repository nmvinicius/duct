#!/usr/bin/env bash
# scripts/tools/changelog-gen.sh - Gera CHANGELOG.md

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

VERSION=$(cat .duct-version 2>/dev/null || echo "v0.0.0")
VERSION_NUM=${VERSION#v}
DATE=$(date +%Y-%m-%d)

# Pega commits desde última tag
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [[ -n "$LAST_TAG" ]]; then
    RANGE="$LAST_TAG..HEAD"
else
    RANGE="HEAD"
fi

duct_info "Generating changelog for $VERSION (since ${LAST_TAG:-beginning})"

# Categoriza commits (apenas linha do subject, não corpo inteiro)
# Usar --grep com --all-match para garantir que o subject começa com o tipo

FEATS=$(git log "$RANGE" --pretty=format:"%s" --grep="^feat" 2>/dev/null | sed 's/^/- /' || true)
FIXES=$(git log "$RANGE" --pretty=format:"%s" --grep="^fix" 2>/dev/null | sed 's/^/- /' || true)
CHORES=$(git log "$RANGE" --pretty=format:"%s" --grep="^chore" 2>/dev/null | sed 's/^/- /' || true)
DOCS=$(git log "$RANGE" --pretty=format:"%s" --grep="^docs" 2>/dev/null | sed 's/^/- /' || true)
TESTS=$(git log "$RANGE" --pretty=format:"%s" --grep="^test" 2>/dev/null | sed 's/^/- /' || true)
REFACTORS=$(git log "$RANGE" --pretty=format:"%s" --grep="^refactor" 2>/dev/null | sed 's/^/- /' || true)
PERFS=$(git log "$RANGE" --pretty=format:"%s" --grep="^perf" 2>/dev/null | sed 's/^/- /' || true)
BUILDS=$(git log "$RANGE" --pretty=format:"%s" --grep="^build" 2>/dev/null | sed 's/^/- /' || true)
CIS=$(git log "$RANGE" --pretty=format:"%s" --grep="^ci" 2>/dev/null | sed 's/^/- /' || true)

# Detecta BREAKING por subject '!: ' ou trailer 'BREAKING CHANGE:'
BREAKING_SUBJECT=$(git log "$RANGE" --pretty=format:"%s" --extended-regexp --grep='^[a-z]+(\([^)]+\))?!:' -i 2>/dev/null | sed 's/^/- /' || true)
BREAKING_BODY=$(git log "$RANGE" --pretty=format:"%s" --grep='^BREAKING CHANGE:' -i 2>/dev/null | sed 's/^/- /' || true)

BREAKING=$(printf '%s\n%s\n' "$BREAKING_SUBJECT" "$BREAKING_BODY" | awk '!seen[$0]++' | sed '/^$/d' || true)

# Gera entrada
ENTRY="## [$VERSION_NUM] - $DATE

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
    ENTRY+="### 🐛 Bug Fixes
$FIXES

"
fi

if [[ -n "$REFACTORS" ]]; then
    ENTRY+="### 🔨 Refactoring
$REFACTORS

"
fi

if [[ -n "$PERFS" ]]; then
    ENTRY+="### ⚡ Performance
$PERFS

"
fi

if [[ -n "$TESTS" ]]; then
    ENTRY+="### 🧪 Tests
$TESTS

"
fi

if [[ -n "$DOCS" ]]; then
    ENTRY+="### 📚 Documentation
$DOCS

"
fi

if [[ -n "$BUILDS" ]]; then
    ENTRY+="### 🏗️ Build
$BUILDS

"
fi

if [[ -n "$CIS" ]]; then
    ENTRY+="### 🔄 CI/CD
$CIS

"
fi

if [[ -n "$CHORES" ]]; then
    ENTRY+="### 🔧 Maintenance
$CHORES

"
fi

# Insere no topo do CHANGELOG
if [[ -f CHANGELOG.md ]]; then
    # Preserva header, insere nova entrada
    tail -n +3 CHANGELOG.md > .changelog-tmp 2>/dev/null || true
    echo -e "# Changelog\n\n$ENTRY$(cat .changelog-tmp 2>/dev/null)" > CHANGELOG.md
    rm -f .changelog-tmp
else
    echo -e "# Changelog\n\n$ENTRY" > CHANGELOG.md
fi

duct_log "CHANGELOG.md updated for $VERSION"
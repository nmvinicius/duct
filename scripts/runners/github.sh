#!/usr/bin/env bash
# scripts/runners/github.sh - GitHub Actions runner

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

duct_log "GitHub Actions runner initialized"

require_var "GITHUB_SHA" "Must run inside GitHub Actions"

export GIT_COMMIT="$GITHUB_SHA"
export GIT_BRANCH="$GITHUB_REF_NAME"

if [[ "$GITHUB_REF_TYPE" == "tag" ]]; then
    export GIT_TAG="$GITHUB_REF_NAME"
else
    export GIT_TAG=""
fi

duct_info "Run: $GITHUB_RUN_ID"
duct_info "Commit: $GIT_COMMIT @ $GIT_BRANCH"
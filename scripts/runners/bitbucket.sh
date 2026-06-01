#!/usr/bin/env bash
# scripts/runners/bitbucket.sh - Bitbucket Pipelines runner

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

duct_log "Bitbucket runner initialized"

# Validate Bitbucket environment
require_var "BITBUCKET_COMMIT" "Must run inside Bitbucket Pipeline"
require_var "BITBUCKET_BRANCH" "BITBUCKET_BRANCH not set"

export GIT_COMMIT="$BITBUCKET_COMMIT"
export GIT_BRANCH="$BITBUCKET_BRANCH"
export GIT_TAG="${BITBUCKET_TAG:-}"

duct_info "Pipeline: $BITBUCKET_PIPELINE_UUID"
duct_info "Commit: $GIT_COMMIT @ $GIT_BRANCH"
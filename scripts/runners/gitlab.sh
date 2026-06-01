#!/usr/bin/env bash
# scripts/runners/gitlab.sh - GitLab CI runner

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../core.sh"

duct_log "GitLab CI runner initialized"

require_var "CI_COMMIT_SHA" "Must run inside GitLab CI"

export GIT_COMMIT="$CI_COMMIT_SHA"
export GIT_BRANCH="$CI_COMMIT_REF_NAME"
export GIT_TAG="${CI_COMMIT_TAG:-}"

duct_info "Pipeline: $CI_PIPELINE_ID"
duct_info "Commit: $GIT_COMMIT @ $GIT_BRANCH"
#!/usr/bin/env bash
# scripts/core.sh - Core utilities sourced by all scripts
# Usage: source "$(dirname "$0")/core.sh"

set -euo pipefail

# ============================================
# LOGGING
# ============================================

C_RESET='\033[0m'
C_RED='\033[0;31m'
C_GREEN='\033[0;32m'
C_YELLOW='\033[1;33m'
C_BLUE='\033[0;34m'
C_CYAN='\033[0;36m'

duct_log()   { echo -e "${C_GREEN}[DUCT]${C_RESET} $1"; }
duct_info()  { echo -e "${C_BLUE}[INFO]${C_RESET} $1"; }
duct_warn()  { echo -e "${C_YELLOW}[WARN]${C_RESET} $1"; }
duct_error() { echo -e "${C_RED}[ERROR]${C_RESET} $1" >&2; }
duct_step()  { echo -e "${C_CYAN}[STEP]${C_RESET} ▶ $1"; }

# ============================================
# UTILS
# ============================================

has_cmd() {
    command -v "$1" >/dev/null 2>&1
}

require_var() {
    local var_name="$1"
    local msg="${2:-Variable $var_name is required}"
    if [[ -z "${!var_name:-}" ]]; then
        duct_error "$msg"
        return 1
    fi
}

ensure_dir() {
    [[ -d "$1" ]] || mkdir -p "$1"
}

now() { date +%s; }

fmt_duration() {
    local secs="$1"
    if (( secs < 60 )); then
        echo "${secs}s"
    elif (( secs < 3600 )); then
        echo "$(( secs / 60 ))m $(( secs % 60 ))s"
    else
        echo "$(( secs / 3600 ))h $(( (secs % 3600) / 60 ))m"
    fi
}
#!/usr/bin/env bash
# scripts/github/configure-git-auth.sh
# Configure git authentication using GITHUB_TOKEN without embedding secrets in tracked files.
set -euo pipefail

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "::error::Environment variable '${name}' is required" >&2
    exit 1
  fi
}

# Prefer GITHUB_TOKEN, fall back to GH_TOKEN
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
if [[ -z "${TOKEN}" ]]; then
  echo "::error::GITHUB_TOKEN or GH_TOKEN is required" >&2
  exit 1
fi

require_env GITHUB_SERVER_URL
require_env GITHUB_REPOSITORY

echo "::add-mask::${TOKEN}"

echo "::group::Configuring git authentication"
SERVER_HOST="$(echo "${GITHUB_SERVER_URL}" | sed 's|https\?://||' | sed 's|/.*||')"
BOT_LOGIN="qodercli-reverse[bot]"
git config user.name "${BOT_LOGIN}"
git config user.email "${BOT_LOGIN}@users.noreply.${SERVER_HOST}"

# Remove any prior extraheader to avoid leaking tokens
git config --unset-all "http.${GITHUB_SERVER_URL}/.extraheader" 2>/dev/null || true

# Use credential helper via env, never persist token in remote URL
git remote set-url origin "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}.git"
export GIT_ASKPASS=/bin/true
# Store token in a temporary credential helper file (cleaned up on exit)
CRED_FILE="$(mktemp)"
trap 'rm -f "${CRED_FILE}"' EXIT
printf 'protocol=https\nhost=%s\nusername=x-access-token\npassword=%s\n' "${SERVER_HOST}" "${TOKEN}" > "${CRED_FILE}"
git config credential.helper "store --file=${CRED_FILE}"

echo "✓ Git authentication configured (token masked)"
echo "::endgroup::"
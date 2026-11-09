#!/usr/bin/env bash
# scripts/github-actions/run-qodercli.sh
# Run the Go-based qodercli-reverse CLI in GitHub Actions mode.
set -euo pipefail

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "::error::Environment variable '${name}' is required" >&2
    exit 1
  fi
}

require_env GITHUB_WORKSPACE
require_env GITHUB_OUTPUT
require_env GITHUB_ACTION_PATH

echo "::group::Setting up qodercli-reverse environment"
mkdir -p "${HOME}/.qoder/commands" "${HOME}/.qoder/subagents"

if [[ -d "${GITHUB_ACTION_PATH}/.qoder" ]]; then
  if command -v rsync >/dev/null 2>&1; then
    rsync -a "${GITHUB_ACTION_PATH}/.qoder/" "${HOME}/.qoder/" || echo "::warning::Failed to copy built-in resources via rsync"
  else
    cp -R "${GITHUB_ACTION_PATH}/.qoder/." "${HOME}/.qoder/" 2>/dev/null || echo "::warning::Failed to copy built-in resources"
  fi
fi
echo "::endgroup::"

# Prefer a prebuilt binary in PATH, otherwise build from source
if ! command -v qodercli-reverse >/dev/null 2>&1; then
  if [[ -f "${GITHUB_ACTION_PATH}/decompiled/qodercli-reverse" ]]; then
    export PATH="${GITHUB_ACTION_PATH}/decompiled:${PATH}"
  else
    echo "::group::Building qodercli-reverse"
    (cd "${GITHUB_ACTION_PATH}/decompiled" && go build -o "${GITHUB_ACTION_PATH}/decompiled/qodercli-reverse" .)
    export PATH="${GITHUB_ACTION_PATH}/decompiled:${PATH}"
    echo "::endgroup::"
  fi
fi

echo "::group::Running qodercli-reverse"
qodercli-reverse run
EXIT_CODE=$?
echo "::endgroup::"

exit "${EXIT_CODE}"
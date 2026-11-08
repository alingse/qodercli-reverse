#!/usr/bin/env bash
# scripts/github/check-environment.sh
# Validate required GitHub Actions environment for qodercli-reverse.
set -euo pipefail

fail() { echo "::error::$*" >&2; exit 1; }

[[ -n "${GITHUB_WORKSPACE:-}" ]] || fail "GITHUB_WORKSPACE is required"
[[ -d "${GITHUB_WORKSPACE}" ]] || fail "GITHUB_WORKSPACE does not exist: ${GITHUB_WORKSPACE}"

if [[ -z "${GITHUB_TOKEN:-}" && -z "${GH_TOKEN:-}" ]]; then
  echo "::warning::No GITHUB_TOKEN or GH_TOKEN set; GitHub API calls may fail"
fi

for cmd in git jq curl; do
  command -v "$cmd" >/dev/null 2>&1 || fail "Required dependency missing: $cmd"
done

echo "Environment OK"
echo "Workspace: ${GITHUB_WORKSPACE}"
echo "Repository: ${GITHUB_REPOSITORY:-unknown}"
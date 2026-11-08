#!/usr/bin/env bash
# scripts/github/setup-github-mcp.sh
# Optionally merge a GitHub MCP server entry into ~/.qoder.json without overwriting existing config.
set -euo pipefail

TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
if [[ -z "${TOKEN}" ]]; then
  echo "::error::GITHUB_TOKEN or GH_TOKEN is required" >&2
  exit 1
fi

echo "::add-mask::${TOKEN}"

CONFIG_FILE="${HOME}/.qoder.json"
if [[ ! -f "${CONFIG_FILE}" ]]; then
  echo "{}" > "${CONFIG_FILE}"
fi

# Validate existing JSON
if ! jq empty "${CONFIG_FILE}" >/dev/null 2>&1; then
  echo "::error::Existing MCP config is invalid JSON: ${CONFIG_FILE}" >&2
  exit 1
fi

TMP_CONFIG="$(mktemp)"
trap 'rm -f "${TMP_CONFIG}"' EXIT

# Merge without overwriting other servers; only set qoder_github if not present or explicitly enabled
if jq --arg token "${TOKEN}" \
   'if .mcpServers.qoder_github == null then .mcpServers.qoder_github = {"command":"qoder-github-mcp-server","args":["stdio"],"env":{"GITHUB_TOKEN":$token},"type":"stdio"} else . end' \
   "${CONFIG_FILE}" > "${TMP_CONFIG}"; then
  mv "${TMP_CONFIG}" "${CONFIG_FILE}"
  echo "✓ MCP configuration merged at ${CONFIG_FILE}"
else
  echo "::error::Failed to merge MCP configuration" >&2
  exit 1
fi

echo "Note: qoder-github-mcp-server binary must be installed separately and verified."
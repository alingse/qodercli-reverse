#!/usr/bin/env bash
# scripts/github-actions/collect-output.sh
# Collect and report qodercli-reverse output files.
set -euo pipefail

OUTPUT_FILE="${1:-}"
if [[ -z "${OUTPUT_FILE}" || ! -f "${OUTPUT_FILE}" ]]; then
  echo "::error::Output file not found: ${OUTPUT_FILE}" >&2
  exit 1
fi

echo "::group::QoderCLI Output"
wc -l "${OUTPUT_FILE}"
echo "::endgroup::"

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### QoderCLI Run Summary"
    echo ""
    echo "- Output file: \`${OUTPUT_FILE}\`"
    echo "- Lines: $(wc -l < "${OUTPUT_FILE}")"
  } >> "${GITHUB_STEP_SUMMARY}"
fi
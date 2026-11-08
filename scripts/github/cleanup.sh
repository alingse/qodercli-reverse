#!/usr/bin/env bash
# scripts/github/cleanup.sh
# Remove temporary authentication and output state.
set -euo pipefail

echo "::group::Cleaning up"
# Remove temp credential files
find /tmp -maxdepth 1 -name 'qoder-*' -type f -delete 2>/dev/null || true
# Unset git credential helper
git config --unset-all credential.helper 2>/dev/null || true
echo "✓ Cleanup complete"
echo "::endgroup::"
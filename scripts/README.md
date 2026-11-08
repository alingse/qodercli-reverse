# Scripts

Supporting scripts for GitHub Actions integration with qodercli-reverse.

## Structure

- `github/` — environment checks, git authentication, optional GitHub MCP setup, cleanup
- `github-actions/` — runner and output collection helpers

## Notes

- Scripts use `set -euo pipefail` and validate required environment variables.
- Secrets are masked via `::add-mask::` and never written to tracked files.
- The Go CLI (`qodercli-reverse`) is the actual execution engine; scripts only orchestrate.
- qoder-github MCP server installation is optional and must be verified before use.
- Qoder proprietary token-exchange endpoints are intentionally not hardcoded; use standard GitHub tokens.
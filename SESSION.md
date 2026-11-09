# Session Summary — qodercli-reverse independent development

Date: 2026-11-08
Working directory: /Users/sham4/qodercli-reverse
Project directory: /Users/sham4/qodercli-reverse

## What was done
- Investigated official Qoder CLI source: none published (proprietary binary).
- Used `alingse/qodercli-reverse` as reverse-engineered reference.
- Forked to `sham435/qodercli-reverse` and cloned to `/Users/sham4/qodercli-reverse`.
- Verified build: `go build` success; `go vet` clean after fix; `go test ./...` all pass.
- Renamed local state dir to `~/.qodercli-reverse` to avoid colliding with official `.qoder`/`.qodercli-reverse`.
- Installed binary as `qodercli-reverse` in `~/.local/bin`.
- Opened PRs to upstream:
  - https://github.com/alingse/qodercli-reverse/pull/1 (docs: v1.1.85 CLI surface)
  - https://github.com/alingse/qodercli-reverse/pull/2 (fix: tea.Printf format strings)
  - https://github.com/alingse/qodercli-reverse/pull/3 (chore: ~/.qodercli-reverse config dir)

## Key paths
- Source: /Users/sham4/qodercli-reverse
- Binary: ~/.local/bin/qodercli-reverse
- Config/state: ~/.qodercli-reverse
- Memory: /Users/sham4/qodercli-reverse/.opencode-memory.json

## Next
- User will rename parent folder manually.
- Customize subsystems as needed (agent, providers, MCP, tools, permissions, TUI, session, skills).

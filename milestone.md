# Milestone: Independent qodercli-reverse Development Baseline

Status: COMPLETED
Date: 2026-11-08
Working directory: /Users/sham4/qodercli-reverse
Project directory: /Users/sham4/qodercli-reverse

## Objective
Establish an independently managed, buildable, testable development repository
based on the reverse-engineered qodercli-reverse reference, isolated from official
Qoder tooling (.qoder / .qodercli-reverse).

## Completed tasks

### T1 — Investigate official source availability
- Result: No official qodercli-reverse source published; only proprietary binary.
- Qoder org repos checked: no qodercli-reverse source repo.

### T2 — Fork and clone reference repository
- Forked: alingse/qodercli-reverse → sham435/qodercli-reverse
- Cloned to: /Users/sham4/qodercli-reverse
- Origin: git@github.com:sham435/qodercli-reverse.git
- Branches preserved: main, add-qodercli-reverse-v1.1.85-notes, fix/tui-printf-vet,
  docs/development-guide, chore/rename-config-dir

### T3 — Verify baseline build
- Command: cd decompiled && go build -o /tmp/qodercli-reverse .
- Result: SUCCESS (Go 1.25.6, darwin/arm64)
- Module: github.com/alingse/qodercli-reverse/decompiled

### T4 — Fix go vet failures
- Issue: tea.Printf called with non-constant format strings in tui/app/model.go
- Fix: wrapped arguments with "%s" format (8 call sites)
- Result: go vet ./... clean

### T5 — Verify tests
- Command: go test ./...
- Result: all packages pass (2 ok, rest [no test files], no failures)

### T6 — Isolate config/state directory from official tooling
- Changed ~/.qoder → ~/.qodercli-reverse in logger.go and root.go help text
- Added .qodercli-reverse (plus .qoder, .qodercli-reverse) to protected paths in permission.go
- Verified: running binary creates ~/.qodercli-reverse on macOS

### T7 — Install local binary
- Installed: ~/.local/bin/qodercli-reverse (executable, on PATH)
- Version output: qodercli-reverse version 0.1.30 (decompiled)

### T8 — Send pull requests to GitHub
- PR #1: docs: observed CLI surface for qodercli-reverse v1.1.85
  https://github.com/alingse/qodercli-reverse/pull/1
- PR #2: fix(tui): constant format strings in tea.Printf
  https://github.com/alingse/qodercli-reverse/pull/2
- PR #3: chore: use ~/.qodercli-reverse for logs/config
  https://github.com/alingse/qodercli-reverse/pull/3

### T9 — Documentation
- /Users/sham4/qodercli-reverse/DEVELOPMENT.md — build/test commands, limitations,
  customization entry points, development workflow
- /Users/sham4/qodercli-reverse/SESSION.md — session summary
- /Users/sham4/qodercli-reverse/milestone.md — this file

## Known limitations
- Reconstructed code, not official source; behavior may diverge.
- Minimal test coverage (most packages have no tests).
- Partial implementations in providers, MCP, permissions, session persistence.
- Internal Alibaba fork of Bubble Tea; API mismatches possible.
- No LICENSE in upstream; treat as learning/research only.

## Customization entry points
- Agent: decompiled/core/agent/agent/
- Providers: decompiled/core/agent/provider/
- MCP: decompiled/core/resource/mcp/
- Tools: decompiled/core/agent/tools/
- Permissions: decompiled/core/agent/permission/
- TUI: decompiled/tui/
- Session/memory: decompiled/core/agent/state/, core/config/
- Skills/subagents: core/resource/, core/agent/

## Next steps
- User will rename parent folder manually.
- Further customization and testing as needed.

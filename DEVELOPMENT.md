# Independent Development Guide for qodercli-reverse (Reconstructed)

## Status of the codebase

- **Official source**: Qoder does NOT publish the qodercli-reverse CLI source. The published
  artifact is a proprietary Mach-O binary (`qodercli-reverse`).
- **Reference implementation**: `alingse/qodercli-reverse` is a community
  reverse-engineered reconstruction of qodercli-reverse v0.1.29/v0.1.30. It is NOT
  official source and may differ from the vendor implementation.
- **License**: The upstream repository does not ship a `LICENSE` file. Treat the
  contents as "learning/research only" and do not redistribute proprietary
  extracted material without authorization. Verify permissions before any
  redistribution.

## Repository layout

```
qodercli-reverse-source/
├── decompiled/          # Reconstructed Go project (module: github.com/alingse/qodercli-reverse/decompiled)
│   ├── cmd/             # CLI entry points (root, print, tui, utils)
│   ├── core/            # agent, config, resource/mcp, pubsub, types, utils
│   ├── tui/             # Bubble Tea TUI components
│   ├── version/         # version string
│   ├── main.go
│   └── go.mod
├── docs/                # Architecture and reverse-engineering notes
└── DEVELOPMENT.md       # this file
```

## Build

```bash
cd decompiled
go mod tidy
go build -o /tmp/qodercli-reverse-decompiled .
```

Verified: builds successfully on Go 1.25.6 / darwin arm64 (2026-11-08).

## Test

```bash
cd decompiled
go vet ./...        # must be clean
go test ./...       # package-level tests
```

Verified results (2026-11-08):
- `go vet ./...`: clean after fixing non-constant format strings in `tui/app/model.go`.
- `go test ./...`: all packages pass (`ok` or `[no test files]`); no failures.

## Known limitations

- Reconstructed code is derived from binary analysis; behavior may diverge from
  the official binary.
- Many packages have no tests (`[no test files]`); coverage is minimal.
- LLM providers, MCP, permission engine, session persistence, and tool
  execution are partially implemented; expect gaps.
- The TUI uses an internal Alibaba fork of Bubble Tea; APIs may not match
  upstream charmbracelet versions exactly.
- No official support or updates; track upstream `alingse/qodercli-reverse`
  for improvements.

## Customization entry points

- Agent orchestration: `core/agent/agent/`
- LLM providers: `core/agent/provider/`
- MCP integrations: `core/resource/mcp/`
- Tool execution: `core/agent/tools/`
- Permission management: `core/agent/permission/`
- Terminal UI: `tui/`
- Session and memory: `core/agent/state/`, `core/config/`
- Skills/subagents: `core/resource/`, `core/agent/`
- Git integration: search `core/` for git-related helpers

## Development workflow

1. Create a feature branch from `main`: `git checkout -b feature/<name>`
2. Make changes in `decompiled/`.
3. Run `go vet ./...` and `go test ./...`.
4. Build and smoke-test the binary.
5. Commit with a clear message and push to your fork.
6. Open a PR (to your fork or upstream) when ready.

## Attribution

Based on the reverse-engineering work in `alingse/qodercli-reverse`. Keep
upstream attribution in commits and PRs.

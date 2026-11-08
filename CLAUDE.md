# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the reverse-engineering project for qodercli-reverse (v0.1.29), aimed at reconstructing its architecture and implementation. qodercli-reverse is an AI coding assistant CLI tool supporting multiple LLM Providers, the MCP protocol, a tool system, and a TUI interface.

## Core Documents

**Required reading** (by priority):

1. `docs/official-architecture.md` - Official architecture analysis (reference for refactoring)
2. `docs/architecture-overview.md` - Complete architecture overview
3. `docs/11-decompiled-gap-analysis.md` - Decompilation gap analysis
4. `README.md` - Project description and development workflow

**Module documentation**:
- `docs/01-package-structure.md` - Package structure and dependencies
- `docs/02-cli-commands.md` - CLI command system
- `docs/03-llm-integration.md` - LLM integration
- `docs/04-mcp-protocol.md` - MCP protocol
- `docs/05-tools-system.md` - Tool system
- `docs/06-session-management.md` - Session management
- `docs/07-permission-security.md` - Permission and security
- `docs/08-config-api.md` - Configuration and API
- `docs/09-github-integration.md` - GitHub integration
- `docs/10-subagent-skill.md` - SubAgent and Skill
- `docs/12-bubbletea-analysis.md` - Bubble Tea usage analysis

## Development Workflow

**Important**: Before modifying or extending code, you must first perform reverse engineering to ensure the implementation is consistent with the official architecture.

### 1. Locate the official binary

```bash
which qodercli-reverse
# Example output: /Users/zhihu/.qoder/bin/qodercli-reverse/qodercli-reverse-0.1.29
```

### 2. Reverse-engineer the target module

```bash
# Extract the symbol table
nm -gU $(which qodercli-reverse) | grep "keyword"

# Extract strings
strings $(which qodercli-reverse) | grep "keyword"

# Disassemble a specific function
go tool objdump -s "function name" $(which qodercli-reverse)
```

### 3. Compare against the official architecture

Refer to the package structure in `docs/official-architecture.md` to ensure modifications conform to the official layered architecture.

### 4. Implement and verify

- Organize code according to the official architecture
- Use official naming conventions (e.g., `{Tool}Params`)
- Compare against official binary behavior after implementation

## Building and Testing

### Compile the decompiled code

```bash
cd decompiled
go mod tidy
go build -o qodercli-reverse .
```

**Note**: The full code may not compile directly; see `COMPILATION_GUIDE.md` for details.

### Run tests

```bash
cd decompiled
go test ./...
```

### Reverse-engineering commands

```bash
# View binary info
file $(which qodercli-reverse)
otool -L $(which qodercli-reverse)

# Symbol table analysis
nm -gU $(which qodercli-reverse) | grep "mcp"
nm $(which qodercli-reverse) | wc -l

# String extraction
strings $(which qodercli-reverse) | grep "code.alibaba-inc.com"

# Go-specific analysis
go tool nm $(which qodercli-reverse) | grep "package name"
```

## Architecture Overview

```
┌─────────────────┐
│   CLI / TUI     │  ← cmd/ (Cobra + custom Bubble Tea fork)
└────────┬────────┘
         │
┌────────▼────────┐
│   Agent Core    │  ← core/agent/ (main loop, Provider, Tools, Permission)
└────────┬────────┘
         │
┌────────▼────────┐
│   Resources     │  ← core/resource/ (MCP, Skill, SubAgent)
└────────┬────────┘
         │
┌────────▼────────┐
│   Utilities     │  ← core/utils/ (HTTP, Token, Storage, API)
└────────┬────────┘
```

**Key modules**:
- **Multi-Provider support**: Qoder, OpenAI, Anthropic, IdeaLab, DashScope
- **Tool system**: Bash, Read, Write, Edit, Grep, Glob, etc.
- **MCP protocol**: stdio/SSE transport, extensible tool resources
- **Permission system**: fine-grained rule engine, Hook interception
- **Runtime modes**: interactive/non-interactive/SDK/containerized
- **TUI framework**: uses Alibaba's internal fork of Bubble Tea (code.alibaba-inc.com/qoder-core/bubbletea v0.0.2)

## Code Organization

```
decompiled/
├── cmd/                    # CLI command entry points
│   ├── root.go            # Main command
│   ├── print/             # Print mode
│   ├── tui/               # TUI mode
│   └── utils/             # CLI utility functions
├── core/                   # Core logic
│   ├── agent/             # Agent system
│   │   ├── agent/         # Main loop
│   │   ├── provider/      # LLM Provider
│   │   ├── tools/         # Tool implementations
│   │   ├── permission/    # Permission control
│   │   ├── state/         # State management
│   │   └── compact/       # Context compaction
│   ├── config/            # Configuration loading
│   ├── resource/mcp/      # MCP client
│   ├── pubsub/            # Event system
│   ├── log/               # Logging system
│   └── types/             # Type definitions
├── tui/                    # TUI interface
│   ├── app/               # App startup
│   └── components/        # UI components
└── main.go                # Program entry point
```

## Important Conventions

1. **Package import path**: use `github.com/alingse/qodercli-reverse/decompiled` as the module path
2. **Naming conventions**: tool parameter types use the `{Tool}Params` format
3. **Architecture alignment**: all modifications must reference `docs/official-architecture.md`
4. **Reverse engineering first**: analyze the official implementation before writing code

## Common Tasks

### Add a new tool

1. Create a new file in `decompiled/core/agent/tools/`
2. Implement the `Tool` interface
3. Register the tool in `tools.go`
4. Refer to the tool implementation in the official binary

### Modify a Provider

1. See `decompiled/core/agent/provider/`
2. Refer to `docs/03-llm-integration.md`
3. Use `nm` to analyze the official Provider implementation

### Extend MCP functionality

1. See `decompiled/core/resource/mcp/`
2. Refer to `docs/04-mcp-protocol.md`
3. Analyze the official MCP client implementation

## Notes

- Decompiled code is derived from the binary and may differ from the original implementation
- Code is not fully tested; not recommended for production use
- For learning and architecture reference only
- Always read the relevant documentation and perform reverse engineering before modifying code

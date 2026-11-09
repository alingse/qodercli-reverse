# qodercli-reverse Complete Architecture Overview

## Project Information

- **Internal Module Path**: `code.alibaba-inc.com/qoder-core/qodercli-reverse`
- **Version**: v0.1.29
- **Platform**: darwin arm64 (Mach-O 64-bit executable)
- **Language**: Go 1.25.7
- **File Size**: ~37.2 MB

---

## 1. Overall Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ cmd/ (cobra commands)                                           │   │
│  │ ├── root.go          ├── mcp/          ├── jobs/               │   │
│  │ ├── start/           ├── update/       ├── feedback/           │   │
│  │ └── ...                                                         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              TUI Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ tui/ (Bubble Tea)                                               │   │
│  │ ├── app/              ├── components/      ├── theme/           │   │
│  │ ├── cmd/              ├── messages/        ├── styles/          │   │
│  │ └── event/            └── interaction/                         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            Agent Core Layer                              │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ core/agent/                                                     │   │
│  │ ├── agent/           # Agent main loop                          │   │
│  │ ├── provider/        # LLM Provider (Qoder/OpenAI/IdeaLab)     │   │
│  │ ├── tools/           # Built-in tools (Bash/Read/Write/Edit/...) │   │
│  │ ├── permission/      # Permission system                        │   │
│  │ ├── hooks/           # Hook system                              │   │
│  │ ├── prompt/          # Prompt management                        │   │
│  │ ├── state/           # State management                         │   │
│  │ ├── compact/         # Context compression                      │   │
│  │ └── task/            # SubAgent scheduling                     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Resource Layer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ core/resource/                                                  │   │
│  │ ├── mcp/             # MCP server management                    │   │
│  │ ├── skill/           # Skill system                             │   │
│  │ ├── subagent/        # SubAgent definitions                    │   │
│  │ ├── command/         # Slash commands                           │   │
│  │ ├── hook/            # Hook definitions                         │   │
│  │ ├── plugin/          # Plugin system                            │   │
│  │ └── output_style/    # Output styles                            │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Utility Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ core/utils/                                                     │   │
│  │ ├── qoder/           # Qoder API client                         │   │
│  │ ├── http/            # HTTP client                               │   │
│  │ ├── tokens/          # Token calculation                        │   │
│  │ ├── storage/         # Local storage                             │   │
│  │ ├── sls/             # Log reporting                             │   │
│  │ └── ...                                                         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                         External Services                                │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ Qoder API (openapi.qoder.sh)                                    │   │
│  │ ├── /api/v1/userinfo      ├── /api/v2/model/list               │   │
│  │ ├── /api/v1/tracking      ├── /api/v2/quota/usage              │   │
│  │ └── ...                                                         │   │
│  │                                                                 │   │
│  │ LLM Providers: Qoder, Anthropic, OpenAI, IdeaLab, DashScope    │   │
│  │                                                                 │   │
│  │ MCP Servers: External tools and resources                      │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Core Data Flow

### 2.1 User Input → Agent Response

```
User input
    │
    ▼
TUI editor (tui/components/interaction/editor/)
    │
    ▼
User submit Hook (core/agent/hooks/user_prompt_submit/)
    │
    ▼
Agent main loop (core/agent/agent/)
    │
    ├── Build messages (core/agent/state/message/)
    │
    ├── Call Provider (core/agent/provider/)
    │   │
    │   ▼
    │   LLM API (Qoder/OpenAI/Anthropic/IdeaLab)
    │   │
    │   ▼
    │   Streaming response events
    │
    ├── Parse tool calls
    │   │
    │   ▼
    │   Permission check (core/agent/permission/)
    │   │
    │   ▼
    │   PreToolUse Hook
    │   │
    │   ▼
    │   Tool execution (core/agent/tools/)
    │   │
    │   ▼
    │   PostToolUse Hook
    │
    └── Loop until done
```

### 2.2 MCP Tool Call

```
Agent decides to call MCP tool
    │
    ▼
MCP rule matching (core/agent/permission/mcp_rule_matcher/)
    │
    ▼
MCP client (core/resource/mcp/)
    │
    ├── stdio/SSE communication
    │
    ▼
MCP server process
    │
    ▼
Return result
```

---

## 3. Key Type Definitions

### 3.1 Message Types

```go
type Message struct {
    Role       string        // "user", "assistant", "system"
    Content    []ContentPart // multi-part content
    ToolCalls  []ToolCall    // tool calls
}

type ToolCall struct {
    ID        string
    Name      string
    Arguments string // JSON
}

type ToolResult struct {
    ToolCallID string
    Content    string
    IsError    bool
}
```

### 3.2 Provider Interface

```go
type Client interface {
    Stream(ctx context.Context, req ModelRequest) <-chan Event
    Send(ctx context.Context, req ModelRequest) (*Response, error)
}
```

### 3.3 Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, input string) (string, error)
}
```

---

## 4. Configuration Hierarchy

```
Priority (low → high):

1. Built-in defaults
2. ~/.qoder/config.json          # Global config
3. ~/.qoder/settings.json        # Global settings
4. ./.qoder/settings.json        # Project settings
5. ./.qoder/settings.local.json  # Local override
6. ./.claude/CLAUDE.md           # Claude-compatible config
7. Environment variables (QODER_*)
8. Command-line arguments
```

---

## 5. Runtime Modes

| Mode | Description | Entry |
|------|-------------|-------|
| **Interactive** | TUI interactive chat | `qodercli-reverse` |
| **Non-interactive** | Single query then exit | `qodercli-reverse -p "..."` |
| **SDK** | JSON protocol communication | stdin/stdout |
| **ACP Server** | Agent Communication Protocol | Internal service |
| **Worktree** | Git worktree concurrency | `--worktree` |
| **Container** | Docker container | Internal mode |
| **Kubernetes** | K8s Job | Internal mode |

---

## 6. Third-Party Dependencies

| Category | Dependency |
|----------|------------|
| CLI Framework | `github.com/spf13/cobra` |
| TUI Framework | Bubble Tea (inferred) |
| LLM SDK | `github.com/anthropics/anthropic-sdk-go` |
| GitHub API | `github.com/google/go-github/v57` |
| Config Parsing | `github.com/BurntSushi/toml` |
| Syntax Highlighting | `github.com/alecthomas/chroma/v2` |
| HTML Conversion | `github.com/JohannesKaufmann/html-to-markdown/v2` |
| Compression | `github.com/andybalholm/brotli` |
| HTTPDNS | `github.com/aliyun/alicloud-httpdns-go-sdk` |

---

## 7. File Inventory

| Document | Description |
|----------|-------------|
| `PLAN.md` | Task planning |
| `01-package-structure.md` | Package structure and dependency graph |
| `02-cli-commands.md` | CLI command system |
| `03-llm-integration.md` | AI/LLM integration |
| `04-mcp-protocol.md` | MCP protocol implementation |
| `05-tools-system.md` | Tool system |
| `06-session-management.md` | Session management |
| `07-permission-security.md` | Permissions and security |
| `08-config-api.md` | Configuration and API |
| `09-github-integration.md` | GitHub integration |
| `10-subagent-skill.md` | SubAgent and Skill |
| `architecture-overview.md` | This document |

---

## 8. Summary

qodercli-reverse is a fully-featured AI programming assistant CLI tool with the following characteristics:

1. **Multi-Provider Support**: Qoder, Anthropic, OpenAI, IdeaLab, DashScope
2. **Rich Toolset**: Bash, file operations, search, network requests, etc.
3. **MCP Protocol**: Extensible tool and resource integration
4. **Permission System**: Fine-grained operation permission control
5. **Hook System**: Customizable workflow hooks
6. **SubAgent**: Parallel task execution capability
7. **Skill System**: Reusable skill definitions
8. **GitHub Integration**: PR review, Actions integration
9. **Multiple Runtime Modes**: Interactive, non-interactive, SDK, containerized

The project has a clear code structure and a high degree of modularity, representing a typical Go CLI application architecture.
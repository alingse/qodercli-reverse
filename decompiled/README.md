# qodercli-reverse Decompiled Code

This directory contains decompiled Go code derived from binary analysis and documentation.

## Directory Structure

```
decompiled/
├── core/
│   ├── types/
│   │   └── types.go              # Core type definitions
│   ├── agent/
│   │   ├── agent/
│   │   │   └── agent.go          # Agent main loop
│   │   ├── provider/
│   │   │   ├── provider.go       # Provider interface
│   │   │   └── qoder.go          # Qoder Provider implementation
│   │   ├── tools/
│   │   │   ├── tools.go          # Tool system interface
│   │   │   ├── bash.go           # Bash tool implementation
│   │   │   └── file.go           # File operation tools
│   │   └── permission/
│   │       └── permission.go     # Permission system
│   ├── resource/
│   │   └── mcp/
│   │       └── mcp.go            # MCP client
│   └── pubsub/
│       └── pubsub.go             # Event publish/subscribe
├── tui/
│   ├── app/
│   │   ├── model.go              # TUI application model
│   │   └── run.go                # Application launcher
│   └── components/
│       ├── editor/
│       │   └── editor.go         # Message editor
│       ├── messages/
│       │   └── messages.go       # Message list
│       ├── chat/
│       │   └── chat.go           # Chat view
│       └── status/
│           └── status.go          # Status bar
└── README.md                     # This file
```

## Code Statistics

| Module | Files | Lines | Description |
|--------|-------|-------|-------------|
| core/types | 1 | ~200 | Basic type definitions |
| core/agent | 6 | ~1800 | Agent core implementation |
| core/resource | 1 | ~350 | MCP client |
| core/pubsub | 1 | ~100 | Event system |
| tui | 6 | ~900 | TUI components |
| **Total** | **15** | **~3350** | |

## Core Architecture

### 1. Agent Architecture

```
┌─────────────────────────────────────────┐
│              TUI / CLI                  │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│           Agent (agent.go)              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
│  │Provider │  │  Tools  │  │Permission│ │
│  └─────────┘  └─────────┘  └─────────┘ │
│  ┌─────────┐  ┌─────────┐              │
│  │  State  │  │  Hooks  │              │
│  └─────────┘  └─────────┘              │
└─────────────────────────────────────────┘
```

### 2. Provider Architecture

```
┌─────────────────────────────────────────┐
│           Provider Interface            │
├─────────────────────────────────────────┤
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │  Qoder  │ │ OpenAI  │ │ Anthropic │ │
│  └─────────┘ └─────────┘ └──────────┘ │
│  ┌─────────┐ ┌─────────┐              │
│  │ IdeaLab │ │DashScope│              │
│  └─────────┘ └─────────┘              │
└─────────────────────────────────────────┘
```

### 3. Tool System Architecture

```
┌─────────────────────────────────────────┐
│           Tool Registry                 │
├─────────────────────────────────────────┤
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │  Read   │ │  Write  │ │   Edit   │ │
│  └─────────┘ └─────────┘ └──────────┘ │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │  Glob   │ │  Grep   │ │ Bash     │ │
│  └─────────┘ └─────────┘ └──────────┘ │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │ BashOut │ │KillBash │ │ Delete   │ │
│  └─────────┘ └─────────┘ └──────────┘ │
└─────────────────────────────────────────┘
```

## Key Types

### Message Structure
```go
type Message struct {
    Role       Role          `json:"role"`
    Content    []ContentPart `json:"content,omitempty"`
    ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
    ToolCallID string        `json:"tool_call_id,omitempty"`
}
```

### ToolCall Structure
```go
type ToolCall struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

### Provider Interface
```go
type Client interface {
    Stream(ctx context.Context, req *ModelRequest) <-chan Event
    Send(ctx context.Context, req *ModelRequest) (*Response, error)
    Close() error
}
```

## Notes

1. **Code nature**: This code is derived from binary decompilation and documentation inference, and may differ from the original code
2. **Completeness**: Some internal implementation details may be incomplete or require adjustment
3. **Dependencies**: The code uses the inferred package path `code.alibaba-inc.com/qoder-core/qodercli-reverse`
4. **Third-party libraries**: Depends on open-source libraries such as Bubble Tea, Lipgloss, and Cobra

## Usage Recommendations

These decompiled codes can be used for:

- Understanding the architecture design of qodercli-reverse
- Learning implementation patterns of the Agent system
- Referencing TUI application construction
- Analyzing MCP protocol implementation

Not recommended for direct use in production because:

- The code is untested
- May contain inaccurate inferences
- Lacks the complete context of the original project
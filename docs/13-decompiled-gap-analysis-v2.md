# qodercli-reverse Decompilation Gap Analysis Report V2

> Analysis date: 2026-03-11
> Official version: qodercli-reverse v0.1.30
> Project version: decompiled v0.1.30
> Analysis tools: official binary --help + source comparison

---

## 1. Full Overview of Original CLI Parameters/Subcommands

### 1.1 Subcommand List (7)

| Subcommand | Description | Current Status | Change |
|------------|-------------|----------------|--------|
| `jobs` | List concurrent worktree tasks (`-a, --all`) | **Completely missing** | ↔ No change |
| `rm` | Delete concurrent tasks (with k8s support `--kubeconfig`, `--namespace`) | **Completely missing** | ↔ No change |
| `completion` | Shell autocompletion (bash/zsh/fish/powershell) | **Completely missing** | ↔ No change |
| `feedback` | Submit feedback (`-c`, `-i`, `-s`, `--workdir`) | **Completely missing** | ↔ No change |
| `mcp` | MCP server management | **Completely missing** | ↔ No change |
| `status` | Account/CLI status (`-o text/json`) | **Completely missing** | ↔ No change |
| `update` | Auto-update to latest version | **Completely missing** | ↔ No change |

#### MCP Subcommand Details

```
qodercli-reverse mcp
  add      Add a new MCP server (name, endpoint, -e, -H, -s, -t)
  auth     Authenticate with an MCP server using OAuth
  get      Get details of an MCP server
  list     List all MCP servers
  remove   Remove an MCP Server
```

**New official parameters** (v0.1.30 vs v0.1.29):
- `--experimental-mcp-load` - Enable dynamic discovery and loading of MCP server tools (experimental)

### 1.2 Main Flags (25+)

| Flag | Type | Description | Current Status |
|------|------|-------------|----------------|
| `-p, --print` | string | Non-interactive mode, exit after single prompt | ✅ **Implemented** (cmd/print/) |
| `-c, --continue` | bool | Continue most recent conversation | ⚠️ **Has flag, no implementation** |
| `-r, --resume` | string | Resume specified session | ⚠️ **Has flag, no implementation** |
| `-w, --workspace` | string | Working directory | ✅ **Implemented** |
| `-f, --output-format` | string | Output format (text/json/stream-json) | ✅ **Implemented** |
| `--input-format` | string | Input format (text/stream-json) | ⚠️ **Has flag, stub implementation** |
| `--model` | string | Model selection (10 types) | ✅ **Implemented** |
| `--max-turns` | int | Agent max loop iterations | ✅ **Implemented** |
| `--max-output-tokens` | string | Max output tokens (16k/32k) | ⚠️ **Has flag, not connected** |
| `--allowed-tools` | stringArray | Allowed tool list | ✅ **Implemented** |
| `--disallowed-tools` | stringArray | Disallowed tool list | ✅ **Implemented** |
| `--attachment` | stringArray | Attachment paths | ⚠️ **Has flag, partial implementation** |
| `--agents` | string | Custom Agent JSON definition | ⚠️ **Has flag, no implementation** |
| `--dangerously-skip-permissions` | bool | Skip permission checks | ✅ **Implemented** |
| `--yolo` | bool | Alias for above | ✅ **Implemented** |
| `--worktree` | bool | Start concurrent tasks via git worktree | ⚠️ **Has flag, no implementation** |
| `--branch` | string | Worktree branch name | ⚠️ **Has flag, no implementation** |
| `--path` | string | Worktree path | ⚠️ **Has flag, no implementation** |
| `--with-claude-config` | bool | Load .claude config | ✅ **Implemented** |
| `--experimental-mcp-load` | bool | Dynamically load MCP tools (experimental) | ❌ **Completely missing** |
| `-q, --quiet` | bool | Hide spinner in non-interactive mode | ⚠️ **Has flag, not fully connected** |
| `-v, --version` | bool | Version info | ✅ **Implemented** |
| `-h, --help` | bool | Help info | ✅ **Implemented** |

---

## 2. Completed Improvements (Compared to V1 Report)

### ✅ 2.1 Cobra CLI Framework

**Status**: Completed

**Implementation files**:
- `cmd/root.go` - Root command + all flags bound
- `cmd/print/print.go` - Print mode implementation
- `cmd/tui/tui.go` - TUI mode entry
- `cmd/utils/` - Config and utility functions

**Code example**:
```go
var rootCmd = &cobra.Command{
    Use:   "qodercli-reverse",
    Short: "Qoder CLI - AI-powered development assistant",
    Run: func(cmd *cobra.Command, args []string) {
        // Mode routing logic
    },
}
```

### ✅ 2.2 System Prompt Builder

**Status**: Completed

**Implementation files**:
- `core/prompts/builder.go` - SystemPromptBuilderV2 implementation
- `core/prompts/env_collector.go` - Environment info collection
- `core/prompts/project_loader.go` - Project context loading
- `core/prompts/builtin.go` - Built-in prompt templates

**Features**:
- Base role definition
- Tool usage guidelines
- Permission rules explanation
- Environment info (OS, Git status)
- Project-specific instructions (AGENTS.md, .claude/)
- Coding standards and best practices

### ✅ 2.3 Logging System

**Status**: Completed (exceeds original)

**Implementation files**:
- `core/log/log.go` - Complete file logging system

**Features**:
- Multi-level logging (Debug/Info/Warn/Error/Fatal)
- File rotation
- Call site tracking

### ✅ 2.4 Version Management

**Status**: Completed

**Implementation files**:
- `version/version.go` - Version definition

---

## 3. Still Missing Modules

### P0 — Core Missing

#### 3.1 Session Persistence / Recovery

**Original functionality**:
```bash
qodercli-reverse -c                    # Continue most recent conversation
qodercli-reverse -r <session-id>       # Resume specified session
```

**Inferred implementation structure**:
```
~/.qoder/
├── sessions/
│   ├── latest.json            # Most recent session index
│   └── <session-id>/
│       ├── messages.json      # Message history
│       ├── state.json         # Agent state
│       └── metadata.json      # Metadata
```

**Current status**:
- Flag already defined (`-c`, `-r`)
- No actual storage implementation

**To implement**:
```go
core/session/
├── session.go          # Session manager
├── storage.go          # Storage interface
└── file_storage.go     # File storage implementation
```

#### 3.2 Missing Subcommands

| Subcommand | Priority | Complexity | Notes |
|------------|----------|------------|-------|
| `mcp` | P0 | Medium | MCP server management |
| `mcp add` | P0 | Medium | Add MCP server |
| `mcp auth` | P1 | High | OAuth authentication flow |
| `mcp get` | P1 | Low | View server details |
| `mcp list` | P0 | Low | List servers |
| `mcp remove` | P1 | Low | Delete server |
| `jobs` | P1 | High | Concurrent task list |
| `rm` | P1 | High | Delete task |
| `status` | P2 | Low | Account status |
| `feedback` | P2 | Medium | Feedback submission |
| `update` | P2 | Medium | Auto-update |
| `completion` | P2 | Low | Built into Cobra |

#### 3.3 MCP Tool Bridge

**Current status**:
- ✅ MCP Client library (`core/resource/mcp/mcp.go`)
- ❌ MCP tools not registered to Agent

**To implement**:
```go
// Dynamically convert MCP tools into Agent tools
func (m *MCPClient) ToAgentTools() []tools.Tool {
    // Get MCP tools list
    // Convert to Tool interface
    // Register to Agent
}
```

---

### P1 — Important Missing

#### 3.4 Missing Tool Implementations

**Current tools** (10):
| Tool | Status | File |
|------|--------|------|
| `Read` | ✅ | `file.go` |
| `Write` | ✅ | `file.go` |
| `Edit` | ✅ | `file.go` |
| `DeleteFile` | ✅ | `file.go` |
| `Glob` | ✅ | `file.go` |
| `Grep` | ✅ | `file.go` |
| `Bash` | ✅ | `bash.go` |
| `BashOutput` | ✅ | `bash.go` |
| `KillBash` | ✅ | `bash.go` |
| `TodoWrite` | ✅ | `todowrite.go` |

**Missing tools**:
| Tool Name | Function | Priority | Schema |
|-----------|----------|----------|--------|
| `Task` | Sub-agent scheduling | High | ✅ Defined |
| `WebFetch` | Web page fetching | High | ❌ Missing |
| `WebSearch` | Web search | High | ❌ Missing |
| `AskUserQuestion` | User interaction prompt | Medium | ❌ Missing |
| `ImageGen` | Image generation | Medium | ❌ Missing |
| `Skill` | Skill system invocation | Medium | ❌ Missing |

**TaskTool Schema** already defined but not implemented:
```go
// BuildTaskSchema builds Task tool Schema
func BuildTaskSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "subagent_type": map[string]interface{}{
                "type": "string",
                "enum": []string{
                    "code-reviewer",
                    "design-agent",
                    "spec-review-agent",
                    "task-executor",
                    "general-purpose",
                },
            },
            "description": map[string]interface{}{
                "type": "string",
            },
            "prompt": map[string]interface{}{
                "type": "string",
            },
        },
        "required": []string{"subagent_type", "description", "prompt"},
    }
}
```

#### 3.5 Worktree / Jobs System

**Original functionality**:
```bash
qodercli-reverse --worktree --branch feature-1
qodercli-reverse jobs -a
qodercli-reverse rm <id> --kubeconfig ~/.kube/config --namespace default
```

**Inferred architecture**:
```
core/
├── worktree/
│   ├── manager.go          # Git worktree management
│   ├── job.go              # Task definition
│   └── runner.go           # Task runner
├── jobs/
│   ├── registry.go         # Task registry
│   ├── storage.go          # Task storage
│   └── kubernetes.go       # K8s Job support
```

**Current status**: Completely missing (only flags)

---

### P2 — Feature Enhancements Missing

#### 3.6 Config Loading Layer Completion

**Current status**:
- `core/config/config.go` - Basic structure
- Environment variable loading

**Missing**:
- Multi-layer config merging (global/project/env vars)
- MCP config management (`~/.qoder/.mcp.json`)
- Claude config compatibility (`.claude/`)

**Target config structure**:
```go
type Config struct {
    // CLI settings
    Model           string
    MaxTokens       int
    MaxTurns        int
    OutputFormat    string
    PermissionMode  permission.Mode

    // API config
    Provider        string  // openai/qoder/anthropic/idealab/dashscope
    APIKey          string
    BaseURL         string

    // Tool config
    AllowedTools    []string
    DisallowedTools []string

    // MCP config
    MCPServers      map[string]*mcp.ServerConfig

    // Session config
    ContinueSession bool
    ResumeSessionID string

    // Project config
    WithClaudeConfig bool
}
```

#### 3.7 Provider Completion

**Current implementation**:
- ✅ OpenAI Provider
- ✅ Qoder Provider

**Missing Providers**:
- Anthropic
- IdeaLab
- DashScope

#### 3.8 SDK/IDE Integration Mode

**Original functionality**:
```bash
qodercli-reverse --input-format stream-json --output-format stream-json
```

**Current status**:
- Flag already defined
- `tui/app/run.go` has `RunSDKMode` stub

**To implement**:
- JSONEncoder/JSONDecoder
- Message protocol handling
- Streaming output handling

---

## 4. Architecture-Level Gap Summary

### 4.1 Directory Structure Comparison

```
Original qodercli-reverse architecture:        Current project architecture:
├── cmd/                                 ├── cmd/
│   ├── root.go              ✅          │   ├── root.go           ✅
│   ├── mcp.go               ❌          │   ├── print/            ✅
│   │   ├── add.go           ❌          │   ├── tui/              ✅
│   │   ├── auth.go          ❌          │   └── utils/            ✅
│   │   ├── get.go           ❌          │
│   │   ├── list.go          ❌          ├── core/
│   │   └── remove.go        ❌          │   ├── agent/
│   ├── jobs.go              ❌          │   │   ├── agent/        ✅
│   ├── rm.go                ❌          │   │   ├── provider/     ⚠️
│   ├── status.go            ❌          │   │   │   ├── openai.go ✅
│   ├── feedback.go          ❌          │   │   │   └── qoder.go  ✅
│   ├── update.go            ❌          │   │   ├── tools/        ⚠️
│   └── completion.go        ❌          │   │   ├── permission/   ✅
│                                        │   │   └── state/        ✅
├── core/                                │   ├── config/           ⚠️
│   ├── config/              ⚠️          │   │   └── config.go
│   │   ├── config.go                    │   ├── prompts/          ✅
│   │   ├── loader.go        ❌          │   │   ├── builder.go
│   │   └── claude_compat.go ❌          │   │   └── ...
│   ├── session/             ❌          │   ├── log/              ✅
│   │   ├── session.go       ❌          │   ├── pubsub/           ✅
│   │   └── storage.go       ❌          │   ├── resource/
│   ├── prompt/              ✅          │   │   └── mcp/          ⚠️
│   │   └── builder.go       ✅          │   └── types/            ✅
│   ├── agent/                           │
│   │   ├── tools/           ⚠️          ├── tui/
│   │   │   ├── webfetch.go  ❌          │   ├── app/              ✅
│   │   │   ├── websearch.go ❌          │   └── components/       ✅
│   │   │   ├── task.go      ❌          │
│   │   │   ├── imagegen.go  ❌          └── version/              ✅
│   │   │   └── ask.go       ❌
│   │   └── subagent/        ❌
│   ├── worktree/            ❌
│   │   └── manager.go       ❌
│   └── update/              ❌
│       └── updater.go       ❌
│
└── tui/                     ✅
```

### 4.2 Functional Coverage Estimate

| Module | Original Features | Current Implementation | Coverage | Change |
|--------|-------------------|------------------------|----------|--------|
| CLI Framework | Cobra + 25+ flags + 7 subcommands | Cobra + flags + basic subcommands | 70% | ↑+30% |
| Agent Core | Message loop + tool calls + Hook | Basically complete | 80% | ↔ |
| Provider | 5 Providers | OpenAI + Qoder | 40% | ↓-20% |
| Tool System | 15+ tools | 10 tools | 65% | ↑+5% |
| Permission System | File/Bash/MCP/Web rules | Basically complete | 85% | ↔ |
| MCP | Client + CLI management | Client only | 50% | ↔ |
| TUI | Bubble Tea + Markdown | Fully implemented | 95% | ↔ |
| System Prompt | Dynamic build | Fully implemented | 90% | ↑+90% |
| Config System | Multi-layer config + env vars | Basic implementation | 50% | ↑+10% |
| Session Management | Persistence + recovery | Flag defined | 10% | ↑+10% |
| Worktree/Jobs | Git worktree + concurrent tasks | Completely missing | 0% | ↔ |
| Logging System | Basic stderr | Full file logging | 150% | ↔ |
| **Overall** | — | — | **~60%** | ↑+10% |

---

## 5. Suggested Implementation Priority

### Phase 1: Core Feature Completion (P0)

1. **Session Persistence System**
   - Implement session storage interface
   - Support `-c` / `-r` flags
   - Persist message history

2. **MCP Subcommands**
   - `mcp add/get/list/remove`
   - MCP config file management
   - Dynamic tool registration to Agent

3. **Task Tool**
   - Schema already defined
   - Implement sub-agent scheduling

### Phase 2: Feature Completion (P1)

4. **Complete Missing Tools**
   - WebFetch, WebSearch
   - AskUserQuestion
   - ImageGen (if dependencies allow)

5. **Worktree/Jobs System**
   - Git worktree management
   - Task registry
   - `jobs` / `rm` subcommands

6. **Config System Completion**
   - Multi-layer config merging
   - MCP config management

### Phase 3: Advanced Features (P2)

7. **Other Providers**
   - Anthropic
   - IdeaLab
   - DashScope

8. **Other Subcommands**
   - `status`
   - `feedback`
   - `update`
   - `completion`

9. **SDK Mode Completion**
   - JSON protocol implementation
   - IDE integration support

---

## 6. Key Code Missing List

### 6.1 Files to Create

```
cmd/
├── mcp.go               # MCP subcommand entry
├── mcp_add.go           # mcp add
├── mcp_auth.go          # mcp auth
├── mcp_get.go           # mcp get
├── mcp_list.go          # mcp list
├── mcp_remove.go        # mcp remove
├── jobs.go              # jobs subcommand
├── rm.go                # rm subcommand
├── status.go            # status subcommand
├── feedback.go          # feedback subcommand
├── update.go            # update subcommand
└── completion.go        # completion subcommand

core/
├── session/
│   ├── session.go       # Session management
│   └── storage.go       # Storage implementation
├── config/
│   ├── loader.go        # Config loader
│   └── mcp_config.go    # MCP config
├── worktree/
│   ├── manager.go       # Git Worktree management
│   ├── job.go           # Task definition
│   └── runner.go        # Task runner
└── agent/tools/
    ├── task.go          # Task tool
    ├── webfetch.go      # WebFetch tool
    ├── websearch.go     # WebSearch tool
    ├── ask.go           # AskUserQuestion tool
    └── imagegen.go      # ImageGen tool
```

### 6.2 Files to Modify

| File | Modification |
|------|-------------|
| `cmd/root.go` | Add subcommand registration |
| `core/config/config.go` | Extend Config structure |
| `core/agent/tools/tools.go` | Register new tools |
| `core/resource/mcp/mcp.go` | Add tool bridge |
| `tui/app/run.go` | Complete SDK mode |

---

## 7. Summary

The current decompiled version has made significant progress compared to the V1 report:

### ✅ Completed

1. **Cobra CLI Framework** - Complete command-line interface
2. **System Prompt Builder** - Dynamic system prompt generation
3. **Logging System** - Full implementation exceeding original
4. **Tool System** - 10 core tools
5. **TUI System** - Complete interactive interface

### ❌ Still Missing

1. **7 Subcommands** - mcp, jobs, rm, status, feedback, update, completion
2. **Session Persistence** - Implement `-c` / `-r` flags
3. **5+ Tools** - Task, WebFetch, WebSearch, AskUser, ImageGen
4. **MCP Tool Bridge** - Dynamic registration of MCP tools to Agent
5. **Worktree/Jobs** - Concurrent task management

### Overall Assessment

| Dimension | Score | Description |
|-----------|-------|-------------|
| CLI Framework | 70% | Cobra done, subcommands missing |
| Agent Core | 80% | Stable and usable |
| Tool System | 65% | Core tools usable |
| TUI Experience | 95% | Close to original |
| Config Management | 50% | Basic functionality usable |
| Extension Features | 30% | Many missing |
| **Overall** | **60%** | **Usability greatly improved** |

**Recommendation**: The current version already has basic usability. Prioritize implementing **Session Persistence** and **Task Tool**, as these two features bring the greatest user experience improvement.
# qodercli-reverse Decompiled Gap Analysis Report

> Analysis date: 2026-03-09
> Original version: qodercli-reverse v0.1.29
> Analysis tools: Original binary --help output + decompiled/ source comparison

---

## I. Original CLI Flags/Subcommands Overview

### 1.1 Subcommand List (7)

The following subcommands are visible from `qodercli-reverse --help`:

| Subcommand | Description | Decompiled Status |
|------------|-------------|-------------------|
| `jobs` | List concurrent worktree tasks (`-a, --all`) | **Completely missing** |
| `rm` | Delete concurrent tasks (with k8s support `--kubeconfig`, `--namespace`) | **Completely missing** |
| `completion` | Shell auto-completion (bash/zsh/fish/powershell) | **Completely missing** |
| `feedback` | Submit feedback (`-c`, `-i`, `-s`, `--workdir`) | **Completely missing** |
| `mcp` | MCP server management | **Completely missing** (only MCP Client library) |
| `status` | Account/CLI status (`-o text/json`) | **Completely missing** |
| `update` | Auto-update to latest version | **Completely missing** |

#### MCP Subcommand Details

```
qodercli-reverse mcp
  add      Add a new MCP server (name, endpoint, -e, -H, -s, -t)
  auth     Authenticate with an MCP server using OAuth
  get      Get details of an MCP server
  list     List all MCP servers
  remove   Remove an MCP Server
```

### 1.2 Main Flags (25+)

| Flag | Type | Description | Decompiled Status |
|------|------|-------------|-------------------|
| `-p, --print` | string | Non-interactive mode, exit after single prompt | **Missing** — main.go handles with simple `os.Args[1]` |
| `-c, --continue` | bool | Continue most recent conversation | **Missing** |
| `-r, --resume` | string | Resume specified session | run.go has stub, **not wired up** |
| `-w, --workspace` | string | Working directory | **Missing** |
| `-f, --output-format` | string | Output format (text/json/stream-json) | **Missing** |
| `--input-format` | string | Input format (text/stream-json) | **Missing** |
| `--model` | string | Model selection (10: auto/efficient/gmodel/kmodel/lite/mmodel/performance/q35model/qmodel/ultimate) | config has field, **no CLI binding** |
| `--max-turns` | int | Max agent loop iterations | agent.Config has field, **no CLI binding** |
| `--max-output-tokens` | string | Max output tokens (16k/32k) | **Missing** |
| `--allowed-tools` | stringArray | Allowed tool list | Internal logic exists, **no CLI binding** |
| `--disallowed-tools` | stringArray | Disallowed tool list | Internal logic exists, **no CLI binding** |
| `--attachment` | stringArray | Attachment paths (repeatable) | **Missing** |
| `--agents` | string | Custom Agent JSON definition | **Missing** |
| `--dangerously-skip-permissions` | bool | Skip permission checks | **Missing** |
| `--yolo` | bool | Same as above (alias) | **Missing** |
| `--worktree` | bool | Launch concurrent task via git worktree | **Completely missing** |
| `--branch` | string | Worktree branch name | **Completely missing** |
| `--path` | string | Worktree path | **Completely missing** |
| `--with-claude-config` | bool | Load .claude config | **Completely missing** |
| `-q, --quiet` | bool | Hide spinner in non-interactive mode | **Missing** |
| `-v, --version` | bool | Version info | **Missing** |
| `-h, --help` | bool | Help info | **Missing** |

---

## II. Important Missing Modules Analysis

### P0 — Core Missing (binary basically unusable without these)

#### 2.1 Cobra CLI Framework

**Status**: Current `main.go` is entirely hardcoded logic, no `spf13/cobra` usage.

**Original features** (confirmed by help output format using Cobra):
- Standard Usage/Examples/Flags layout
- `completion` subcommand (Cobra built-in)
- Nested subcommand structure (`mcp add/get/list/remove/auth`)

**Missing content**:
- All flags parsing logic
- All subcommand routing
- `cmd/` directory entirely empty

**Impact**: Cannot use any command-line functionality

#### 2.2 Session Persistence / Resume

**Original features**:
- `-c` (continue): Continue most recent conversation
- `-r <session-id>` (resume): Resume specified session

**Speculated implementation**:
```
~/.qoder/
├── sessions/
│   ├── latest.json          # Latest session index
│   └── <session-id>/
│       ├── messages.json    # Message history
│       ├── state.json       # Agent state
│       └── metadata.json    # Metadata
```

**Current status**: Completely missing, no session storage module

#### 2.3 System Prompt Generator

**Status**: Hardcoded to `"You are a helpful AI assistant."`

**Original speculation**: Based on actual usage behavior and system prompt analysis, original has a complex prompt builder:

```go
// Speculated SystemPromptBuilder structure
type SystemPromptBuilder struct {
    basePrompt      string
    toolDescriptions []ToolDescription
    envInfo         *EnvironmentInfo
    projectConfig   *ProjectConfig
    userInstructions []string
}

// Generated system prompt contains:
// 1. Base role definition
// 2. Tool usage guide
// 3. Permission rules explanation
// 4. Environment info (OS, Git status, working directory)
// 5. Project-specific instructions (AGENTS.md, .claude/)
// 6. Coding standards and best practices
```

**Missing impact**: Agent behavior differs greatly from original

#### 2.4 Non-Interactive Print Mode

**Original features**:
```bash
qodercli-reverse -p "Explain the use of context in Go"
qodercli-reverse -p "..." -f json           # JSON output
qodercli-reverse -p "..." --max-turns 25    # Limit iterations
```

**Current status**: `main.go` uses `os.Args[1]` simple handling, missing:
- Output format conversion (text/json/stream-json)
- Proper streaming output handling
- Exit code handling

---

### P1 — Important Missing (affects functional completeness)

#### 2.5 MCP Subcommand System

**Existing**: `core/resource/mcp/mcp.go` — MCP Client library (JSON-RPC stdio protocol)

**Missing**: CLI subcommand layer

| Subcommand | Function | Needs Implementation |
|------------|----------|---------------------|
| `mcp add` | Add MCP server config | Config file writing |
| `mcp get` | View server details | Config reading |
| `mcp list` | List all servers | Config traversal |
| `mcp remove` | Delete server | Config deletion |
| `mcp auth` | OAuth authentication | OAuth flow |

**Integration missing**: MCP tools not registered into Agent's tool registry

#### 2.6 Missing Tool Implementations

**Currently implemented tools** (9):
- `Read` — File reading
- `Write` — File writing
- `Edit` — File editing
- `DeleteFile` — File deletion
- `Glob` — File pattern matching
- `Grep` — Content search (via ripgrep)
- `Bash` — Shell command execution
- `BashOutput` — Background command output retrieval
- `KillBash` — Terminate background command

**Missing important tools**:

| Tool Name | Function | Priority |
|-----------|----------|----------|
| `WebFetch` | Web scraping | High |
| `WebSearch` | Web search | High |
| `Task` | SubAgent scheduling | High |
| `TodoWrite` | Task management | Medium |
| `AskUserQuestion` | User interaction prompt | Medium |
| `ImageGen` | Image generation | Medium |
| `Skill` | Skill system invocation | Medium |

**Note**: `BuildTaskSchema()` already defined but `TaskTool` not implemented

#### 2.7 Worktree / Jobs System

**Original features**:
```bash
# Launch concurrent worktree task
qodercli-reverse --worktree --branch feature-1

# View tasks
qodercli-reverse jobs -a

# Delete task
qodercli-reverse rm <id>
qodercli-reverse rm <id> --kubeconfig ~/.kube/config --namespace default
```

**Speculated architecture**:
```
core/
├── worktree/
│   ├── manager.go       # Git worktree management
│   └── job.go           # Task definition
├── jobs/
│   ├── registry.go      # Task registry
│   └── kubernetes.go    # K8s Job support
```

**Current status**: Completely missing

#### 2.8 Configuration Loading Layer

**Status**: `config.Config` only has a `Model` field

**Original required config sources**:

| Config Source | Path | Content |
|---------------|------|---------|
| Environment variables | `QODER_*` | API keys, debug options |
| Global config | `~/.qoder/` | User preferences, MCP config |
| Project config | `.qoder/` | Project-specific settings |
| MCP config | `.mcp.json` | MCP server definitions |
| Claude compatible | `.claude/` | skills, commands, subagents |
| Agent config | `AGENTS.md` | Project instructions |
| Settings file | `settings.json` | CLI settings |

**Speculated full Config structure**:
```go
type Config struct {
    // CLI settings
    Model           string
    MaxTokens       int
    MaxTurns        int
    OutputFormat    string
    PermissionMode  permission.Mode
    
    // API config
    Provider        string
    APIKey          string
    BaseURL         string
    
    // Tool config
    AllowedTools    []string
    DisallowedTools []string
    
    // MCP config
    MCPServers      map[string]*mcp.ServerConfig
    
    // Project config
    Workspace       string
    GitBranch       string
    
    // Session config
    ContinueSession bool
    ResumeSessionID string
    
    // Advanced config
    Agents          map[string]*AgentConfig
    WithClaudeConfig bool
}
```

---

### P2 — Feature Enhancement Missing

#### 2.9 Auto-Update (`update` subcommand)

**Original feature**: Check remote version + self-update

**Speculated implementation**:
```go
type Updater struct {
    currentVersion string
    remoteURL      string
    binaryPath     string
}

func (u *Updater) Check() (*VersionInfo, error)
func (u *Updater) Update() error
```

#### 2.10 Feedback Subcommand

**Original feature**:
```bash
qodercli-reverse feedback -c "Problem description" -i image1.png -i image2.jpg -s <session-id>
```

**Needs implementation**:
- Multipart form upload
- Image attachment handling
- Session ID association

#### 2.11 Status Subcommand

**Original feature**:
```bash
qodercli-reverse status           # Text output
qodercli-reverse status -o json   # JSON output
```

**Output content**:
- Account info
- CLI version
- Current config status

#### 2.12 Shell Auto-Completion

**Original**: Cobra built-in support, needs generated scripts:
```bash
qodercli-reverse completion bash
qodercli-reverse completion zsh
qodercli-reverse completion fish
qodercli-reverse completion powershell
```

#### 2.13 Markdown Rendering

**Status**: ✅ **Completed**

**Implementation**: Uses `charmbracelet/glamour` for Markdown rendering

**Key code**:
```go
import "github.com/charmbracelet/glamour"

// Initialize renderer
renderer, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(width),
)

// Message format uses Markdown
prefix = "**You:**\n"
prefix = "**Assistant:**\n"
```

**Features**:
- ✅ Full Markdown syntax support
- ✅ Code block highlighting
- ✅ Lists and links
- ✅ Table rendering
- ✅ Auto line wrapping

**Details**: [TUI Implementation Doc](./20-tui-implementation.md)

#### 2.14 SDK/IDE Integration Mode

**Status**: `run.go` has `RunSDKMode` stub, but `JSONEncoder`/`JSONDecoder` are empty implementations

**Original feature**:
```bash
qodercli-reverse --input-format stream-json --output-format stream-json
```

**Use case**: IDE integration, JSON protocol communication via stdin/stdout

---

## III. Architecture-Level Gap Summary

### 3.1 Directory Structure Comparison

```
Original qodercli-reverse speculated architecture:
├── cmd/                      ← Completely missing (Cobra commands)
│   ├── root.go               ← Main command + all flags
│   ├── mcp.go                ← mcp subcommand group
│   │   ├── add.go
│   │   ├── auth.go
│   │   ├── get.go
│   │   ├── list.go
│   │   └── remove.go
│   ├── jobs.go               ← jobs subcommand
│   ├── rm.go                 ← rm subcommand
│   ├── status.go             ← status subcommand
│   ├── feedback.go           ← feedback subcommand
│   ├── update.go             ← update subcommand
│   └── completion.go         ← completion subcommand
├── core/
│   ├── config/               ← Greatly missing (only 1 field)
│   │   ├── config.go         ← Full config structure
│   │   ├── loader.go         ← Multi-layer config loading
│   │   └── claude_compat.go  ← Claude config compatibility
│   ├── session/              ← Completely missing
│   │   ├── session.go        ← Session management
│   │   └── storage.go        ← Persistence storage
│   ├── prompt/               ← Completely missing
│   │   └── builder.go        ← System Prompt builder
│   ├── agent/
│   │   ├── tools/            ← Greatly missing (6+ tools)
│   │   │   ├── webfetch.go
│   │   │   ├── websearch.go
│   │   │   ├── imagegen.go
│   │   │   ├── todo.go
│   │   │   ├── task.go       ← SubAgent scheduling
│   │   │   ├── skill.go
│   │   │   └── ask.go
│   │   └── subagent/         ← Completely missing
│   │       └── subagent.go   ← SubAgent implementation
│   ├── worktree/             ← Completely missing
│   │   └── manager.go        ← Git Worktree management
│   └── update/               ← Completely missing
│       └── updater.go        ← Self-update logic
└── tui/
    └── (markdown renderer)   ← Missing Glamour integration
```

### 3.2 Functional Coverage Estimate

| Module | Original Feature | Decompiled Implementation | Coverage |
|--------|------------------|---------------------------|----------|
| CLI Framework | Cobra + 25+ flags + 7 subcommands | Partially implemented (root.go) | 40% |
| Agent Core | Message loop + tool calls + Hook | Basically complete | 80% |
| Provider | Qoder + OpenAI + Anthropic + IdeaLab + DashScope | OpenAI + Qoder | 60% |
| Tool System | 15+ tools | 9 tools | 60% |
| Permission System | File/Bash/MCP/Web rules | Basically complete | 85% |
| MCP | Client + CLI management | Client only | 50% |
| **TUI** | Bubble Tea + Markdown + scrolling | Bubble Tea + Markdown + scrolling | **95%** |
| Config System | Multi-layer config + env vars | Basic implementation | 40% |
| Session Management | Persistence + resume | Missing | 0% |
| Worktree/Jobs | Git worktree + concurrent tasks | Missing | 0% |
| **Logging System** | Basic stderr | Complete file logging system | **150%** |
| **Overall** | — | — | **~50%** |

---

## IV. Suggested Implementation Priority

### Phase 1: Core Usability (P0)

1. **Introduce Cobra** — Build `cmd/` framework, bind all flags
2. **Complete Config Loading** — Support multi-layer config and env vars
3. **Implement Session Persistence** — Support `-c` / `-r`
4. **Build System Prompt Builder** — Dynamically generate system prompt

### Phase 2: Feature Completion (P1)

5. **Complete Missing Tools** — WebFetch, WebSearch, Task, TodoWrite, AskUserQuestion
6. **Implement MCP Subcommands** — add/get/list/remove/auth
7. **Implement Print Mode** — `-p` non-interactive output
8. **MCP Tool Bridge** — Dynamically register MCP tools to Agent

### Phase 3: Advanced Features (P2)

9. **Worktree/Jobs System** — Concurrent task management
10. **Auto-Update** — `update` subcommand
11. **Feedback System** — Feedback submission
12. ~~**Markdown Rendering**~~ — ✅ Completed
13. **SDK Mode Completion** — IDE integration support
14. **Logging System Enhancement** — ✅ Completed (surpasses original)

---

## V. Key Code Missing List

### 5.1 Files to Create

```
cmd/
├── root.go              # Cobra root command
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
├── config/
│   ├── loader.go        # Config loader
│   └── claude_compat.go # Claude config compatibility
├── session/
│   ├── session.go       # Session management
│   └── storage.go       # Storage implementation
├── prompt/
│   └── builder.go       # System Prompt building
├── worktree/
│   └── manager.go       # Git Worktree management
├── update/
│   └── updater.go       # Self-update logic
└── agent/
    └── tools/
        ├── webfetch.go
        ├── websearch.go
        ├── task.go
        ├── todo.go
        ├── ask.go
        └── imagegen.go
```

### 5.2 Files to Modify

| File | Modification |
|------|--------------|
| `main.go` | Rewrite as Cobra entry |
| `core/config/config.go` | Extend Config structure |
| `core/agent/provider/qoder.go` | Complete `getMachineID()` |
| `tui/components/chat/chat.go` | Integrate Markdown rendering |
| `tui/app/run.go` | Complete SDK mode implementation |

---

## VI. Conclusion

The current decompiled version successfully restores qodercli-reverse's **core runtime architecture**:

- Agent message loop and tool call mechanism
- Provider streaming communication protocol
- Basic toolset (file operations + Bash)
- Permission rule matching system
- MCP Client protocol implementation
- Bubble Tea TUI framework

But missing the **entire CLI application layer**:

- Cobra command framework (all flags and subcommands)
- Configuration management and session persistence
- System Prompt dynamic construction
- Multiple important tools
- Worktree concurrent task system

**Recommendation**: Prioritize completing the Cobra CLI framework and config system, which is the prerequisite for making the decompiled version basically usable.
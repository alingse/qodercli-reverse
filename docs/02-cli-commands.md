# qodercli-reverse CLI Command System Analysis

## 1. Entry Point and Architecture

### 1.1 Entry Point

```
main.main -> cmd/root.go -> cobra.Command
```

The entry function is in the `main` package, built with the `github.com/spf13/cobra` framework.

### 1.2 Command Hierarchy

```
qodercli-reverse (root)
├── [flags]                      # Global flags
├── jobs                         # Job management command group
│   └── rm                       # Remove Job
├── mcp                          # MCP server management command group
│   ├── add                      # Add MCP server
│   ├── auth                     # MCP OAuth authentication
│   ├── get                      # Get MCP server details
│   ├── list                     # List all MCP servers
│   └── remove                   # Remove MCP server
├── completion                   # Shell autocompletion generation
├── feedback                     # Submit feedback
├── status                       # Show account and CLI status
└── update                       # Self-update to latest version
```

## 2. Global Flags

| Flag | Shorthand | Type | Description |
|------|-----------|------|-------------|
| `--workspace` | `-w` | string | Working directory |
| `--version` | `-v` | - | Show version |
| `--print` | `-p` | string | Non-interactive mode, print response then exit |
| `--output-format` | `-f` | string | Output format (text/json/stream-json) |
| `--input-format` | - | string | Input format (text/stream-json) |
| `--resume` | `-r` | string | Resume specified session |
| `--continue` | `-c` | - | Continue most recent session |
| `--model` | - | string | Model level selection |
| `--max-turns` | - | int | Maximum agent loop iterations |
| `--max-output-tokens` | - | string | Maximum output tokens (16k/32k) |
| `--attachment` | - | stringArray | Attachment paths |
| `--agents` | - | string | JSON custom agent definitions |
| `--allowed-tools` | - | stringArray | Allowed tools list |
| `--disallowed-tools` | - | stringArray | Disabled tools list |
| `--dangerously-skip-permissions` | - | - | Skip all permission checks |
| `--yolo` | - | - | Same as above (alias) |
| `--branch` | - | string | Branch name (used with --worktree) |
| `--path` | - | string | Worktree path |
| `--worktree` | - | - | Start a concurrent Job via git worktree |
| `--with-claude-config` | - | - | Load Claude Code configuration |
| `--quiet` | `-q` | - | Hide spinner |

## 3. Model Selection Options

The `--model` parameter supports the following values:

| Value | Description |
|-------|-------------|
| `auto` | Automatic selection |
| `efficient` | Efficient mode |
| `performance` | Performance mode |
| `ultimate` | Ultimate mode |
| `lite` | Lightweight mode |
| `qmodel` | Q model |
| `q35model` | Q3.5 model |
| `gmodel` | G model |
| `kmodel` | K model |
| `mmodel` | M model |

## 4. TUI Built-in Slash Commands

In interactive mode, the following slash commands are supported (defined under `core/resource/command/`):

### 4.1 Session Management
| Command | Description |
|---------|-------------|
| `/clear` | Clear conversation history |
| `/resume` | Resume a past session |
| `/export [filename]` | Export current session |
| `/compact` | Compact context |

### 4.2 Account and Authentication
| Command | Description |
|---------|-------------|
| `/login` | Log in |
| `/logout` | Log out |
| `/status` | Show status |
| `/usage` | Show usage |

### 4.3 Model and Configuration
| Command | Description |
|---------|-------------|
| `/model` | Switch model |
| `/config` | Configuration management |
| `/prompt` | Prompt management |
| `/memory` | Memory management |

### 4.4 Agent and Skill
| Command | Description |
|---------|-------------|
| `/agents` | Agent list and management |
| `/skills` | Skill list and management |
| `/bashes` | Background Bash process management |

### 4.5 Quest Mode
| Command | Description |
|---------|-------------|
| `/quest-on` | Enter Quest mode |
| `/quest-off` | Exit Quest mode |

### 4.6 GitHub Integration
| Command | Description |
|---------|-------------|
| `/github` | GitHub operations |
| `/setup-github` | Set up GitHub integration |

### 4.7 Other
| Command | Description |
|---------|-------------|
| `/help` | Help information |
| `/quit` | Quit |
| `/vim` | Toggle Vim mode |
| `/upgrade` | Upgrade CLI |
| `/release-notes` | View release notes |
| `/feedback` | Submit feedback |

## 5. Environment Variables

### 5.1 Authentication
| Variable | Description |
|----------|-------------|
| `QODER_PERSONAL_ACCESS_TOKEN` | Qoder personal access token |
| `QODER_AUTH_TOKEN` | Authentication token |
| `QODER_ANTHROPIC_API_KEY` | Anthropic API Key |
| `QODER_OPENAI_API_KEY` | OpenAI API Key |
| `QODER_DASHSCOPE_API_KEY` | DashScope API Key |
| `QODER_IDEALAB_API_KEY` | IdeaLab API Key |
| `QODER_OPENAI_BASE_URL` | OpenAI API Base URL |

### 5.2 Runtime Configuration
| Variable | Description |
|----------|-------------|
| `QODER_ENV` | Runtime environment |
| `QODER_CLI` | CLI mode identifier (=1) |
| `QODER_CLI_INSTALL` | CLI installation identifier |
| `QODER_CLI_RUNNING_MODEL` | Currently running model |
| `QODER_MACHINE_ID` | Machine unique identifier |
| `QODER_ENTRYPOINT` | Entry point |
| `QODER_WORK_VM` | Work VM configuration |

### 5.3 Tool Configuration
| Variable | Description |
|----------|-------------|
| `QODER_BASH_PATH` | Bash path |
| `QODER_BASH_TIMEOUT` | Bash timeout |
| `QODER_PROJECT_DIR` | Project directory |
| `QODER_CURRENT_WORKDIR` | Current working directory |

### 5.4 Debugging and Display
| Variable | Description |
|----------|-------------|
| `QODER_SHOW_TOKEN_USAGE` | Show token usage |
| `QODER_EXPOSE_TOKEN_USAGE` | Expose token usage |
| `QODER_CHANGELOG_URL` | Changelog URL |
| `QODER_ALIBABA_PREVIEW` | Alibaba preview mode |
| `QODER_AGENT_SDK_VERSION` | Agent SDK version |
| `QODER_USER_INFO` | User info |

## 6. Runtime Modes

### 6.1 Interactive Mode (default)
```bash
qodercli-reverse                    # Start interactive TUI
qodercli-reverse -w /path/to/repo   # Specify working directory
```

### 6.2 Non-interactive Mode (Print)
```bash
qodercli-reverse -p "Explain how Go context is used"
qodercli-reverse -p "..." -f json           # JSON output
qodercli-reverse -p "..." --max-turns 25    # Limit loop iterations
```

### 6.3 SDK Mode
Communicates via JSON protocol over stdin/stdout, used for IDE integration.

### 6.4 ACP Server Mode
Runs as an ACP (Agent Communication Protocol) server.

### 6.5 Worktree Mode
```bash
qodercli-reverse --worktree --branch feature-x
```

### 6.6 Container/Kubernetes Mode
```bash
qodercli-reverse --kubeconfig ...
```

## 7. Startup Flow

```
1. cmd/root.go: parse command-line arguments
2. cmd/start/start_local.go (or other launcher):
   - Initialize configuration (core/config)
   - Initialize logging (core/logging)
   - Initialize authentication (core/auth)
   - Load MCP servers (core/resource/mcp)
   - Load Skills (core/resource/skill)
   - Load SubAgents (core/resource/subagent)
   - Load Hooks (core/resource/hook)
   - Load Plugins (core/resource/plugin)
3. tui/app: start TUI application
   - Initialize Bubble Tea program
   - Set up event subscriptions (core/pubsub)
4. core/agent: Agent main loop
   - Wait for user input
   - Call Provider for inference
   - Execute tool calls
   - Update state
```

## 8. Update Mechanism

Supports multiple update methods (`cmd/update/`):

| Method | Package |
|--------|---------|
| Homebrew | `homebrew_updater` |
| NPM | `npm_updater` |
| curl\|bash | `curlbash_updater` |
| Automatic update | `auto` |

## 9. Job Management System

A Job is a concurrently executing Agent instance:

```
jobs/
├── jobs.go      # List Jobs
├── attach/      # Attach to Job
├── fetch/       # Fetch Job status
├── rm/          # Remove Job
└── stop/        # Stop Job
```

Job types:
- **Local Job**: local process
- **Worktree Job**: Git worktree isolation
- **Container Job**: Docker container
- **Kubernetes Job**: K8s Job
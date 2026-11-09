# qodercli-reverse Tool System (Tools) Analysis

## 1. Tool System Architecture

### 1.1 Package Structure

```
core/agent/tools/
├── abstract/              # Tool abstract base class
├── ask_user_question/     # AskUserQuestion tool
├── bash/                  # Bash tool
├── bashkill/              # KillBash tool
├── bashoutput/            # BashOutput tool
├── check_runtime/         # CheckRuntime tool
├── edit/                  # Edit tool
├── enter_plan_in_quest/   # Quest enter Plan mode
├── exit_plan_in_quest/    # Quest exit Plan mode
├── glob/                  # Glob tool
├── grep/                  # Grep tool
├── imagegen/              # ImageGen tool
├── ls/                    # LS tool
├── multiedit/             # MultiEdit tool
├── read/                  # Read tool
├── skill/                 # Skill tool
├── task/                  # Task (SubAgent) tool
├── todowrite/             # TodoWrite tool
├── utils/                 # Tool common utilities
│   ├── file/              # File utilities
│   ├── format/            # Formatting utilities
│   └── repair/            # Repair utilities
├── webfetch/              # WebFetch tool
├── webfetch_trusted/      # WebFetch (trusted) tool
└── write/                 # Write tool
```

## 2. Built-in Tool List

### 2.1 Core Tools

| Tool Name | Type | Description |
|-----------|------|-------------|
| `Bash` | Execution | Execute shell commands |
| `Read` | Read | Read file contents |
| `Write` | Write | Create or overwrite files |
| `Edit` | Edit | Find and replace file contents |
| `MultiEdit` | Edit | Batch edit files |
| `Glob` | Search | File pattern matching search |
| `Grep` | Search | Regex-based content search |
| `LS` | Browse | List directory contents |

### 2.2 Process Management Tools

| Tool Name | Description |
|-----------|-------------|
| `KillBash` | Terminate background Bash process |
| `BashOutput` | Get output of background Bash process |

### 2.3 Task Management Tools

| Tool Name | Description |
|-----------|-------------|
| `Task` | Launch a sub-agent (SubAgent) |
| `TodoWrite` | Update task list |
| `Skill` | Execute a predefined skill |

### 2.4 Interaction Tools

| Tool Name | Description |
|-----------|-------------|
| `AskUserQuestion` | Ask the user a question |

### 2.5 Network Tools

| Tool Name | Description |
|-----------|-------------|
| `WebFetch` | Fetch web page content |
| `WebSearch` | Search the web |

### 2.6 Other Tools

| Tool Name | Description |
|-----------|-------------|
| `ImageGen` | Generate images |
| `CheckRuntime` | Check runtime environment |
| `EnterSpecMode` / `EnterPlanMode` | Enter planning mode |
| `ExitSpecMode` / `ExitPlanMode` | Exit planning mode |
| `DeleteFile` | Delete files |
| `SearchCodebase` | Search codebase |
| `SearchSymbol` | Search symbols |

## 3. Tool Definition Format

### 3.1 Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, input string) (string, error)
}
```

### 3.2 Input Schema Example

```json
{
  "type": "object",
  "properties": {
    "file_path": {
      "type": "string",
      "description": "The absolute path to the file"
    },
    "offset": {
      "type": "integer",
      "description": "Line number to start reading from"
    },
    "limit": {
      "type": "integer",
      "description": "Number of lines to read"
    }
  },
  "required": ["file_path"]
}
```

## 4. Core Tools in Detail

### 4.1 Bash Tool

```go
// Parameter structure
type BashParams struct {
    Command       string   // Command to execute
    Timeout       int      // Timeout (milliseconds)
    RunInBackground bool   // Whether to run in background
    Description   string   // Command description
}
```

Features:
- Supports foreground and background execution
- Timeout control (`QODER_BASH_TIMEOUT`)
- Automatic shell snapshot and restore handling
- Supports Bash mode (user input)

### 4.2 Read Tool

```go
type ReadParams struct {
    FilePath    string  // File path
    Offset      int     // Starting line number
    Limit       int     // Number of lines to read
    ReadImage   bool    // Whether to read images
}
```

Features:
- Supports text files and image files
- Automatic file type detection
- Paginated reading of large files
- Line-number-prefixed output format

### 4.3 Edit Tool

```go
type EditParams struct {
    FilePath    string  // File path
    OldString   string  // Content to replace
    NewString   string  // Replacement content
    ReplaceAll  bool    // Whether to replace all matches
}
```

Features:
- Must first use the Read tool to read the file
- Exact match (including indentation)
- Supports batch replacement
- Automatic backup

### 4.4 Write Tool

```go
type WriteParams struct {
    FilePath    string  // File path
    Content     string  // File content
}
```

Features:
- Atomic write
- Automatic directory creation
- File permission handling

### 4.5 Grep Tool

```go
type GrepParams struct {
    Pattern     string   // Regex pattern
    Path        string   // Search path
    OutputMode  string   // Output mode: content/files_with_matches/count
    Type        string   // File type
    Glob        string   // File pattern
    HeadLimit   int      // Result limit
    Multiline   bool     // Multiline mode
}
```

Features:
- Implemented on top of ripgrep
- Supports multiple output modes
- Supports file type filtering
- Supports context lines (-A, -B, -C)

### 4.6 Glob Tool

```go
type GlobParams struct {
    Pattern     string  // glob pattern
    Path        string  // Search path
}
```

Features:
- Recursive search
- Supports wildcards
- Sorted by modification time

### 4.7 Task Tool

```go
type TaskParams struct {
    SubagentType string  // Sub-agent type
    Description  string  // Task description
    Prompt       string  // Task prompt
}
```

Available sub-agent types:
- `code-reviewer`: Code review
- `design-agent`: Design document
- `spec-review-agent`: Spec review
- `task-executor`: Task execution
- `general-purpose`: General-purpose agent

### 4.8 TodoWrite Tool

```go
type TodoParams struct {
    Todos []TodoItem  // Task list
}

type TodoItem struct {
    Content     string  // Task content
    Status      string  // pending/in_progress/completed
    ActiveForm  string  // In-progress form
}
```

### 4.9 AskUserQuestion Tool

```go
type AskUserQuestionParams struct {
    Questions []Question  // Question list
}

type Question struct {
    Question    string    // Question text
    Header      string    // Short label
    Options     []Option  // Options
    MultiSelect bool      // Whether multi-select
}
```

## 5. Tool Execution Flow

```
1. Agent generates tool call (ToolCall)
2. Permission check (PermissionChecker)
   ├── Check if in allowed_tools list
   ├── Check if in disallowed_tools list
   └── Check path/command permission rules
3. If confirmation needed, show permission dialog
4. Execute tool
   ├── PreToolUse Hook
   ├── Tool execution
   └── PostToolUse Hook
5. Return result to Agent
```

## 6. Permission Control

### 6.1 Tool-Level Permissions

```
--allowed-tools "Bash,Edit"      # Only allow specified tools
--disallowed-tools "WebFetch"    # Disable specified tools
--dangerously-skip-permissions   # Skip all permission checks
```

### 6.2 Path-Level Permissions

```
core/agent/permission/file_rule_matcher/
├── Check whether file path is allowed to access
├── Apply project rules
└── Apply user rules
```

### 6.3 Bash Command Permissions

```
core/agent/permission/bash_rule_matcher/
├── Check whether command is allowed to execute
├── Apply gitignore rules
└── Apply dangerous command list
```

## 7. Hook System

### 7.1 Hook Types

| Hook | Trigger Timing |
|------|----------------|
| `PreToolUse` | Before tool execution |
| `PostToolUse` | After tool execution |
| `AgentStop` | When Agent stops |
| `SessionStart` | When session starts |
| `UserPromptSubmit` | When user submits |

### 7.2 Hook Purposes

- Result truncation (`result_truncator`)
- Result persistence (`result_persister`)
- Code review check (`code_reviewer_checker`)
- External notifications (`external`)

## 8. Tool Result Format

### 8.1 Text Result

```
<result>
File content or command output
</result>
```

### 8.2 Error Result

```json
{
  "error": true,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

### 8.3 Truncated Result

```
... [N lines truncated]
```

## 9. Special Tool Behavior

### 9.1 Bash Tool Special Handling

- `bash-output-special`: Special output handling
- `shell-snapshot`: Shell state snapshot
- `background_shell`: Background shell management

### 9.2 File Tool Special Handling

- `special_path_policy`: Special path policy
- `special_file_protection`: Special file protection
- `qoder_config_protection`: Qoder config protection

### 9.3 WebFetch Tool Special Handling

- `webfetch_rule_matcher`: URL rule matching
- Trusted site list
- Redirect limits

## 10. Tool Configuration

### 10.1 Environment Variables

```
QODER_BASH_PATH      # Bash path
QODER_BASH_TIMEOUT   # Bash timeout
QODER_PROJECT_DIR    # Project directory
```

### 10.2 Configuration File

```json
{
  "allowedTools": ["Bash", "Read", "Write"],
  "disallowedTools": ["WebFetch"],
  "permissionMode": "ask"
}
```
# System Prompt Builder

System prompt builder for qodercli-reverse, implemented based on reverse engineering analysis of the official binary.

## Features

- **Modular design**: Role definition, tool guide, permission rules, environment info, and project context are separated
- **Auto detection**: Automatically detects project type (Go/Node/Python/Rust/Java)
- **Project context**: Loads project-specific instructions such as AGENTS.md, .claude/, .cursorrules
- **Environment info**: Automatically collects OS, Git status, and development environment versions
- **Template support**: Supports variable substitution using Go templates

## Quick Start

### Basic Usage

```go
// Build system prompt with default configuration
config := utils.DefaultSystemPromptConfig()
prompt, err := utils.BuildSystemPrompt(config)
```

### Command Line Usage

```bash
# Auto build (default) - includes project context and environment info
qodercli-reverse

# Use minimal prompt
qodercli-reverse --minimal-prompt

# Disable project info
qodercli-reverse --no-project-info

# Use custom prompt file
qodercli-reverse --system-prompt-file ./my_prompt.md
```

## Architecture Design

### Component Hierarchy

```
SystemPromptBuilderV2
├── Role Definition [P0]
├── Core Instructions [P0]
├── Tool Guide [P0]
├── Permission Rules [P0]
├── Environment Info [P1] - can be disabled
├── Project Context [P1] - can be disabled
├── Coding Standards [P1] - can be disabled
└── Session Context [P2] - optional
```

### Core Types

```go
// System prompt builder
type SystemPromptBuilderV2 struct {
    vars *TemplateVars           // Template variables
    envCollector *EnvironmentCollector      // Environment collector
    projectLoader *ProjectContextLoader     // Project loader
    sessionContext *SessionContext          // Session context
}

// Environment info
type EnvironmentInfo struct {
    OS, Architecture, Shell string
    WorkingDir, HomeDir string
    GitRepo, GitBranch, GitCommit string
    GoVersion, NodeVersion, PythonVersion string
    // ...
}

// Project context
type ProjectContext struct {
    Name, Type, RootPath string
    AgentsMDContent string      // AGENTS.md content
    ClaudeDirContent string     // .claude/ directory content
    CodingStandards string      // Coding standards
    // ...
}
```

## Configuration Options

### Builder Configuration

```go
builder := prompts.NewSystemPromptBuilderV2(vars)

// Enable/disable components
builder.WithRoleDefinition(true).
    WithToolGuide(true).
    WithPermissionRules(true).
    WithEnvironmentInfo(true).
    WithProjectContext(true).
    WithCodingStandards(true).
    WithSessionContext(false)

// Set custom role
builder.WithCustomRole("You are a React expert.")

// Add custom section
builder.AddCustomSection("Custom Rules", "1. Do this\n2. Do that", 100)

// Build
prompt, err := builder.Build()
```

### Command Line Configuration

```go
config := &utils.SystemPromptConfig{
    AppName:              "qodercli-reverse",
    BrandName:            "Qoder",
    Mode:                 "main",  // "main", "subagent", "minimal"
    EnableProjectContext: true,
    EnableEnvironment:    true,
    WorkDir:              "/path/to/project",
    CustomRole:           "",
    SystemPromptFile:     "",
}

prompt, err := utils.BuildSystemPrompt(config)
```

## Project-Specific Instructions

### AGENTS.md

Create an `AGENTS.md` file in the project root:

```markdown
# Project Guidelines

## Technology Stack
- Go 1.21
- PostgreSQL
- Redis

## Coding Standards
- Follow standard Go conventions
- Use gofmt for formatting
- Write tests for all new functions

## Architecture
- Clean architecture pattern
- Repository pattern for data access
- Service layer for business logic
```

### .claude/ Directory

Create a `.claude/` directory to store configuration:

```
.claude/
├── CLAUDE.md          # Main configuration
├── instructions.md    # Additional instructions
└── context.md         # Project context
```

### .cursorrules

Supports `.cursorrules` files:

```
Always use TypeScript strict mode
Prefer functional components over class components
Use tailwind for styling
```

## Project Type Detection

Automatically detects the following project types:

| File | Type | Language |
|------|------|----------|
| go.mod | go | Go |
| package.json | node | JavaScript/TypeScript |
| pyproject.toml/setup.py | python | Python |
| Cargo.toml | rust | Rust |
| pom.xml/build.gradle | java | Java |

## Environment Info Collection

Automatically collects the following info:

### System Info
- OS (linux/darwin/windows)
- Architecture (amd64/arm64)
- Shell (bash/zsh/powershell)
- Working Directory

### Git Info
- Repository status
- Current branch
- Latest commit
- Remote URL
- Modified/untracked files count

### Development Environment
- Go version
- Node.js version
- Python version
- Java version
- Rust version

## Advanced Usage

### Subagent Prompt

```go
config := &utils.SystemPromptConfig{
    Mode:         "subagent",
    SubagentType: "code_reviewer",
}
prompt, err := utils.BuildSystemPrompt(config)
```

### Minimal Prompt

```go
config := &utils.SystemPromptConfig{
    Mode: "minimal",
}
prompt, err := utils.BuildSystemPrompt(config)
```

### Manual Build

```go
builder := prompts.NewSystemPromptBuilderV2(vars)

// Collect environment info
envInfo, err := builder.CollectEnvironment()

// Collect project context
projectCtx, err := builder.CollectProjectContext("/path/to/project")

// Set session context
builder.SetSessionContext(&prompts.SessionContext{
    SessionID:       "session-123",
    PreviousSummary: "Previously we discussed...",
    CurrentTask:     "Implement user authentication",
})

prompt, err := builder.Build()
```

## Template Variables

Available template variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `{{.AppName}}` | Application name | qodercli-reverse |
| `{{.BrandName}}` | Brand name | Qoder |
| `{{.ProductName}}` | Product name | Qoder CLI |
| `{{.AgentName}}` | Agent name | "" |
| `{{.ReadToolName}}` | Read tool name | Read |
| `{{.BashToolName}}` | Bash tool name | Bash |
| `{{.SearchCodebaseTool}}` | Search tool name | SearchCodebase |
| `{{.SearchSymbolTool}}` | Symbol search tool name | SearchSymbol |

## Alignment with Official Architecture

Reference official binary analysis:

```
code.alibaba-inc.com/qoder-core/qodercli-reverse/acp.(*QoderAcpAgent).buildPrompt
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/agent.(*agentContext).GetSystemPrompt
code.alibaba-inc.com/qoder-core/qodercli-reverse/core/resource/command.loadBuiltinPromptsCommands
code.alibaba-inc.com/qoder-core/qodercli-reverse/tui/texts.(*Service).GetText
```

Implementation correspondence:
- `acp.buildPrompt` → `SystemPromptBuilderV2.Build()`
- `agent.GetSystemPrompt` → `utils.BuildSystemPrompt()`
- `command.loadBuiltinPromptsCommands` → `builtin.go`
- `texts.GetText` → `ProjectContextLoader`

## Debugging

Enable debug logging to view the build process:

```bash
qodercli-reverse --debug
```

Log output:
```
[DEBUG] Building system prompt with mode: main
[DEBUG] Detected project type: go
[DEBUG] Loaded AGENTS.md: 1500 bytes
[DEBUG] Built system prompt, length: 3500
```

## Examples

### Complete Example

```go
package main

import (
    "fmt"
    "github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
)

func main() {
    config := utils.DefaultSystemPromptConfig()
    config.EnableProjectContext = true
    config.EnableEnvironment = true
    config.CustomInstructions = "Always write tests first."

    prompt, err := utils.BuildSystemPrompt(config)
    if err != nil {
        panic(err)
    }

    fmt.Println(prompt)
}
```

### Custom Project Instructions

Create `AGENTS.md`:

```markdown
# My Project

## Critical Rules
- NEVER modify files in the vendor/ directory
- ALWAYS run go fmt before saving
- ALWAYS write unit tests

## Technology
- Go 1.21 with generics
- Echo framework for HTTP
- PostgreSQL with sqlx

## Patterns
- Use functional options pattern
- Return errors, don't panic
- Context-first function parameters
```

Run:
```bash
qodercli-reverse
```

The system prompt will automatically include the content of AGENTS.md.

## References

- `extracted_system_prompts.md` - Extracted system prompts
- `extracted_system_prompts_v2.md` - Full version system prompts
- `docs/official-architecture.md` - Official architecture analysis
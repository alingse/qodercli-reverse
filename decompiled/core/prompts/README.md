# System Prompt Builder

qodercli 的系统提示词构建器，参考官方二进制逆向分析实现。

## 功能特性

- **模块化设计**: 角色定义、工具指南、权限规则、环境信息、项目上下文分离
- **自动检测**: 自动检测项目类型 (Go/Node/Python/Rust/Java)
- **项目上下文**: 加载 AGENTS.md, .claude/, .cursorrules 等项目特定指令
- **环境信息**: 自动收集 OS, Git 状态, 开发环境版本
- **模板支持**: 使用 Go template 支持变量替换

## 快速开始

### 基础用法

```go
// 使用默认配置构建系统提示词
config := utils.DefaultSystemPromptConfig()
prompt, err := utils.BuildSystemPrompt(config)
```

### 命令行使用

```bash
# 自动构建（默认）- 包含项目上下文和环境信息
qodercli

# 使用最小化提示词
qodercli --minimal-prompt

# 禁用项目信息
qodercli --no-project-info

# 使用自定义提示词文件
qodercli --system-prompt-file ./my_prompt.md
```

## 架构设计

### 组件层次

```
SystemPromptBuilderV2
├── 角色定义 (Role Definition)     [P0]
├── 核心指令 (Core Instructions)   [P0]
├── 工具指南 (Tool Guide)          [P0]
├── 权限规则 (Permission Rules)    [P0]
├── 环境信息 (Environment Info)    [P1] - 可禁用
├── 项目上下文 (Project Context)   [P1] - 可禁用
├── 编码规范 (Coding Standards)    [P1] - 可禁用
└── 会话上下文 (Session Context)   [P2] - 可选
```

### 核心类型

```go
// 系统提示词构建器
type SystemPromptBuilderV2 struct {
    vars *TemplateVars           // 模板变量
    envCollector *EnvironmentCollector      // 环境收集器
    projectLoader *ProjectContextLoader     // 项目加载器
    sessionContext *SessionContext          // 会话上下文
}

// 环境信息
type EnvironmentInfo struct {
    OS, Architecture, Shell string
    WorkingDir, HomeDir string
    GitRepo, GitBranch, GitCommit string
    GoVersion, NodeVersion, PythonVersion string
    // ...
}

// 项目上下文
type ProjectContext struct {
    Name, Type, RootPath string
    AgentsMDContent string      // AGENTS.md 内容
    ClaudeDirContent string     // .claude/ 目录内容
    CodingStandards string      // 编码规范
    // ...
}
```

## 配置选项

### 构建器配置

```go
builder := prompts.NewSystemPromptBuilderV2(vars)

// 启用/禁用组件
builder.WithRoleDefinition(true).
    WithToolGuide(true).
    WithPermissionRules(true).
    WithEnvironmentInfo(true).
    WithProjectContext(true).
    WithCodingStandards(true).
    WithSessionContext(false)

// 设置自定义角色
builder.WithCustomRole("You are a React expert.")

// 添加自定义章节
builder.AddCustomSection("Custom Rules", "1. Do this\n2. Do that", 100)

// 构建
prompt, err := builder.Build()
```

### 命令行配置

```go
config := &utils.SystemPromptConfig{
    AppName:              "qodercli",
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

## 项目特定指令

### AGENTS.md

在项目根目录创建 `AGENTS.md` 文件：

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

### .claude/ 目录

创建 `.claude/` 目录存放配置：

```
.claude/
├── CLAUDE.md          # 主要配置
├── instructions.md    # 额外指令
└── context.md         # 项目上下文
```

### .cursorrules

支持 `.cursorrules` 文件：

```
Always use TypeScript strict mode
Prefer functional components over class components
Use tailwind for styling
```

## 项目类型检测

自动检测以下项目类型：

| 文件 | 类型 | 语言 |
|------|------|------|
| go.mod | go | Go |
| package.json | node | JavaScript/TypeScript |
| pyproject.toml/setup.py | python | Python |
| Cargo.toml | rust | Rust |
| pom.xml/build.gradle | java | Java |

## 环境信息收集

自动收集以下信息：

### 系统信息
- OS (linux/darwin/windows)
- Architecture (amd64/arm64)
- Shell (bash/zsh/powershell)
- Working Directory

### Git 信息
- Repository status
- Current branch
- Latest commit
- Remote URL
- Modified/untracked files count

### 开发环境
- Go version
- Node.js version
- Python version
- Java version
- Rust version

## 高级用法

### 子 Agent 提示词

```go
config := &utils.SystemPromptConfig{
    Mode:         "subagent",
    SubagentType: "code_reviewer",
}
prompt, err := utils.BuildSystemPrompt(config)
```

### 最小化提示词

```go
config := &utils.SystemPromptConfig{
    Mode: "minimal",
}
prompt, err := utils.BuildSystemPrompt(config)
```

### 手动构建

```go
builder := prompts.NewSystemPromptBuilderV2(vars)

// 收集环境信息
envInfo, err := builder.CollectEnvironment()

// 收集项目上下文
projectCtx, err := builder.CollectProjectContext("/path/to/project")

// 设置会话上下文
builder.SetSessionContext(&prompts.SessionContext{
    SessionID:       "session-123",
    PreviousSummary: "Previously we discussed...",
    CurrentTask:     "Implement user authentication",
})

prompt, err := builder.Build()
```

## 模板变量

可用的模板变量：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `{{.AppName}}` | 应用名称 | qodercli |
| `{{.BrandName}}` | 品牌名称 | Qoder |
| `{{.ProductName}}` | 产品名称 | Qoder CLI |
| `{{.AgentName}}` | Agent 名称 | "" |
| `{{.ReadToolName}}` | 读取工具名 | Read |
| `{{.BashToolName}}` | Bash 工具名 | Bash |
| `{{.SearchCodebaseTool}}` | 搜索工具名 | SearchCodebase |
| `{{.SearchSymbolTool}}` | 符号搜索工具名 | SearchSymbol |

## 与官方架构对齐

参考官方二进制分析：

```
code.alibaba-inc.com/qoder-core/qodercli/acp.(*QoderAcpAgent).buildPrompt
code.alibaba-inc.com/qoder-core/qodercli/core/agent.(*agentContext).GetSystemPrompt
code.alibaba-inc.com/qoder-core/qodercli/core/resource/command.loadBuiltinPromptsCommands
code.alibaba-inc.com/qoder-core/qodercli/tui/texts.(*Service).GetText
```

实现对应关系：
- `acp.buildPrompt` → `SystemPromptBuilderV2.Build()`
- `agent.GetSystemPrompt` → `utils.BuildSystemPrompt()`
- `command.loadBuiltinPromptsCommands` → `builtin.go`
- `texts.GetText` → `ProjectContextLoader`

## 调试

启用调试日志查看构建过程：

```bash
qodercli --debug
```

日志输出：
```
[DEBUG] Building system prompt with mode: main
[DEBUG] Detected project type: go
[DEBUG] Loaded AGENTS.md: 1500 bytes
[DEBUG] Built system prompt, length: 3500
```

## 示例

### 完整示例

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

### 自定义项目指令

创建 `AGENTS.md`：

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

运行：
```bash
qodercli
```

系统提示词将自动包含 AGENTS.md 的内容。

## 参考

- `extracted_system_prompts.md` - 提取的系统提示词
- `extracted_system_prompts_v2.md` - 完整版系统提示词
- `docs/official-architecture.md` - 官方架构分析

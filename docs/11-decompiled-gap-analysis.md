# qodercli 反编译差距分析报告

> 分析日期: 2026-03-09
> 原始版本: qodercli v0.1.29
> 分析工具: 原始二进制 --help 输出 + decompiled/ 源码对比

---

## 一、原版 CLI 参数/子命令全览

### 1.1 子命令列表 (7 个)

从 `qodercli --help` 可以看到以下子命令：

| 子命令 | 描述 | 反编译状态 |
|--------|------|-----------|
| `jobs` | 列出并发 worktree 任务 (`-a, --all`) | **完全缺失** |
| `rm` | 删除并发任务（含 k8s 支持 `--kubeconfig`, `--namespace`） | **完全缺失** |
| `completion` | Shell 自动补全 (bash/zsh/fish/powershell) | **完全缺失** |
| `feedback` | 提交反馈（`-c`, `-i`, `-s`, `--workdir`） | **完全缺失** |
| `mcp` | MCP 服务器管理 | **完全缺失**（仅有 MCP Client 库） |
| `status` | 账号/CLI 状态 (`-o text/json`) | **完全缺失** |
| `update` | 自动更新到最新版本 | **完全缺失** |

#### MCP 子命令详情

```
qodercli mcp
  add      Add a new MCP server (name, endpoint, -e, -H, -s, -t)
  auth     Authenticate with an MCP server using OAuth
  get      Get details of an MCP server
  list     List all MCP servers
  remove   Remove an MCP Server
```

### 1.2 主要 Flags (25+ 个)

| Flag | 类型 | 描述 | 反编译状态 |
|------|------|------|-----------|
| `-p, --print` | string | 非交互模式，单次提示后退出 | **缺失** — main.go 用 `os.Args[1]` 简单处理 |
| `-c, --continue` | bool | 继续最近对话 | **缺失** |
| `-r, --resume` | string | 恢复指定会话 | run.go 有 stub，**未连通** |
| `-w, --workspace` | string | 工作目录 | **缺失** |
| `-f, --output-format` | string | 输出格式 (text/json/stream-json) | **缺失** |
| `--input-format` | string | 输入格式 (text/stream-json) | **缺失** |
| `--model` | string | 模型选择 (10种: auto/efficient/gmodel/kmodel/lite/mmodel/performance/q35model/qmodel/ultimate) | config 有字段，**无 CLI 绑定** |
| `--max-turns` | int | Agent 最大循环次数 | agent.Config 有字段，**无 CLI 绑定** |
| `--max-output-tokens` | string | 最大输出 token (16k/32k) | **缺失** |
| `--allowed-tools` | stringArray | 允许的工具列表 | 内部有逻辑，**无 CLI 绑定** |
| `--disallowed-tools` | stringArray | 禁止的工具列表 | 内部有逻辑，**无 CLI 绑定** |
| `--attachment` | stringArray | 附件路径（可多次指定） | **缺失** |
| `--agents` | string | 自定义 Agent JSON 定义 | **缺失** |
| `--dangerously-skip-permissions` | bool | 跳过权限检查 | **缺失** |
| `--yolo` | bool | 同上（别名） | **缺失** |
| `--worktree` | bool | 通过 git worktree 启动并发任务 | **完全缺失** |
| `--branch` | string | worktree 分支名 | **完全缺失** |
| `--path` | string | worktree 路径 | **完全缺失** |
| `--with-claude-config` | bool | 加载 .claude 配置 | **完全缺失** |
| `-q, --quiet` | bool | 非交互模式隐藏 spinner | **缺失** |
| `-v, --version` | bool | 版本信息 | **缺失** |
| `-h, --help` | bool | 帮助信息 | **缺失** |

---

## 二、重要且缺失的模块分析

### P0 — 核心缺失（没有这些二进制基本不可用）

#### 2.1 Cobra CLI 框架

**现状**: 当前 `main.go` 完全是硬编码逻辑，没有使用 `spf13/cobra`。

**原版特征**（从 help 输出格式可以确认使用 Cobra）:
- 标准的 Usage/Examples/Flags 布局
- `completion` 子命令（Cobra 内置）
- 嵌套子命令结构 (`mcp add/get/list/remove/auth`)

**缺失内容**:
- 所有 flags 解析逻辑
- 所有子命令路由
- `cmd/` 目录完全为空

**影响**: 无法正常使用任何命令行功能

#### 2.2 会话持久化 / 恢复

**原版功能**:
- `-c` (continue): 继续最近一次对话
- `-r <session-id>` (resume): 恢复指定会话

**推测实现**:
```
~/.qoder/
├── sessions/
│   ├── latest.json          # 最近会话索引
│   └── <session-id>/
│       ├── messages.json    # 消息历史
│       ├── state.json       # Agent 状态
│       └── metadata.json    # 元数据
```

**当前状态**: 完全缺失，没有 session storage 模块

#### 2.3 System Prompt 生成器

**现状**: 硬编码为 `"You are a helpful AI assistant."`

**原版推测**: 从实际使用行为和 system prompt 分析，原版有一个复杂的 prompt builder：

```go
// 推测的 SystemPromptBuilder 结构
type SystemPromptBuilder struct {
    basePrompt      string
    toolDescriptions []ToolDescription
    envInfo         *EnvironmentInfo
    projectConfig   *ProjectConfig
    userInstructions []string
}

// 生成的 system prompt 包含：
// 1. 基础角色定义
// 2. 工具使用指南
// 3. 权限规则说明
// 4. 环境信息（OS, Git 状态, 工作目录）
// 5. 项目特定指令（AGENTS.md, .claude/）
// 6. 编码规范和最佳实践
```

**缺失影响**: Agent 行为与原版差异巨大

#### 2.4 非交互 Print 模式

**原版功能**:
```bash
qodercli -p "Explain the use of context in Go"
qodercli -p "..." -f json           # JSON 输出
qodercli -p "..." --max-turns 25    # 限制迭代次数
```

**当前状态**: `main.go` 用 `os.Args[1]` 简单处理，没有：
- 输出格式转换 (text/json/stream-json)
- 正确的流式输出处理
- 退出码处理

---

### P1 — 重要缺失（影响功能完整性）

#### 2.5 MCP 子命令系统

**已有**: `core/resource/mcp/mcp.go` — MCP Client 库（JSON-RPC stdio 协议）

**缺失**: CLI 子命令层

| 子命令 | 功能 | 需要实现 |
|--------|------|----------|
| `mcp add` | 添加 MCP 服务器配置 | 配置文件写入 |
| `mcp get` | 查看服务器详情 | 配置读取 |
| `mcp list` | 列出所有服务器 | 配置遍历 |
| `mcp remove` | 删除服务器 | 配置删除 |
| `mcp auth` | OAuth 认证 | OAuth 流程 |

**集成缺失**: MCP 工具没有注册到 Agent 的 tool registry

#### 2.6 缺失的工具实现

**当前已有工具** (9 个):
- `Read` — 文件读取
- `Write` — 文件写入
- `Edit` — 文件编辑
- `DeleteFile` — 文件删除
- `Glob` — 文件模式匹配
- `Grep` — 内容搜索（调用 ripgrep）
- `Bash` — Shell 命令执行
- `BashOutput` — 后台命令输出获取
- `KillBash` — 终止后台命令

**缺失的重要工具**:

| 工具名 | 功能 | 优先级 |
|--------|------|--------|
| `WebFetch` | 网页抓取 | 高 |
| `WebSearch` | 网页搜索 | 高 |
| `Task` | 子 Agent 调度 | 高 |
| `TodoWrite` | 任务管理 | 中 |
| `AskUserQuestion` | 用户交互询问 | 中 |
| `ImageGen` | 图片生成 | 中 |
| `Skill` | 技能系统调用 | 中 |

**注意**: `BuildTaskSchema()` 已定义但 `TaskTool` 未实现

#### 2.7 Worktree / Jobs 系统

**原版功能**:
```bash
# 启动并发 worktree 任务
qodercli --worktree --branch feature-1

# 查看任务
qodercli jobs -a

# 删除任务
qodercli rm <id>
qodercli rm <id> --kubeconfig ~/.kube/config --namespace default
```

**推测架构**:
```
core/
├── worktree/
│   ├── manager.go       # Git worktree 管理
│   └── job.go           # 任务定义
├── jobs/
│   ├── registry.go      # 任务注册表
│   └── kubernetes.go    # K8s Job 支持
```

**当前状态**: 完全缺失

#### 2.8 配置加载层

**现状**: `config.Config` 只有一个 `Model` 字段

**原版需要的配置来源**:

| 配置来源 | 路径 | 内容 |
|----------|------|------|
| 环境变量 | `QODER_*` | API keys, 调试选项 |
| 全局配置 | `~/.qoder/` | 用户偏好, MCP 配置 |
| 项目配置 | `.qoder/` | 项目特定设置 |
| MCP 配置 | `.mcp.json` | MCP 服务器定义 |
| Claude 兼容 | `.claude/` | skills, commands, subagents |
| Agent 配置 | `AGENTS.md` | 项目指令 |
| 设置文件 | `settings.json` | CLI 设置 |

**推测的完整 Config 结构**:
```go
type Config struct {
    // CLI 设置
    Model           string
    MaxTokens       int
    MaxTurns        int
    OutputFormat    string
    PermissionMode  permission.Mode
    
    // API 配置
    Provider        string
    APIKey          string
    BaseURL         string
    
    // 工具配置
    AllowedTools    []string
    DisallowedTools []string
    
    // MCP 配置
    MCPServers      map[string]*mcp.ServerConfig
    
    // 项目配置
    Workspace       string
    GitBranch       string
    
    // 会话配置
    ContinueSession bool
    ResumeSessionID string
    
    // 高级配置
    Agents          map[string]*AgentConfig
    WithClaudeConfig bool
}
```

---

### P2 — 功能增强缺失

#### 2.9 自动更新 (`update` 子命令)

**原版功能**: 检查远程版本 + 自我更新

**推测实现**:
```go
type Updater struct {
    currentVersion string
    remoteURL      string
    binaryPath     string
}

func (u *Updater) Check() (*VersionInfo, error)
func (u *Updater) Update() error
```

#### 2.10 Feedback 子命令

**原版功能**:
```bash
qodercli feedback -c "问题描述" -i image1.png -i image2.jpg -s <session-id>
```

**需要实现**:
- 多部分表单上传
- 图片附件处理
- Session ID 关联

#### 2.11 Status 子命令

**原版功能**:
```bash
qodercli status           # 文本输出
qodercli status -o json   # JSON 输出
```

**输出内容**:
- 账号信息
- CLI 版本
- 当前配置状态

#### 2.12 Shell 自动补全

**原版**: Cobra 内置支持，需要生成脚本：
```bash
qodercli completion bash
qodercli completion zsh
qodercli completion fish
qodercli completion powershell
```

#### 2.13 Markdown 渲染

**现状**: TUI 纯文本输出

**原版推测**: 使用 `charmbracelet/glamour` 进行 Markdown 渲染

**需要集成**:
```go
import "github.com/charmbracelet/glamour"

renderer, _ := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(width),
)
rendered, _ := renderer.Render(markdownContent)
```

#### 2.14 SDK/IDE 集成模式

**现状**: `run.go` 中有 `RunSDKMode` stub，但 `JSONEncoder`/`JSONDecoder` 是空实现

**原版功能**:
```bash
qodercli --input-format stream-json --output-format stream-json
```

**用途**: IDE 集成，通过 stdin/stdout 进行 JSON 协议通信

---

## 三、架构层面的差距总结

### 3.1 目录结构对比

```
原版 qodercli 推测架构：
├── cmd/                      ← 完全缺失 (Cobra commands)
│   ├── root.go               ← 主命令 + 所有 flags
│   ├── mcp.go                ← mcp 子命令组
│   │   ├── add.go
│   │   ├── auth.go
│   │   ├── get.go
│   │   ├── list.go
│   │   └── remove.go
│   ├── jobs.go               ← jobs 子命令
│   ├── rm.go                 ← rm 子命令
│   ├── status.go             ← status 子命令
│   ├── feedback.go           ← feedback 子命令
│   ├── update.go             ← update 子命令
│   └── completion.go         ← completion 子命令
├── core/
│   ├── config/               ← 大幅缺失 (只有1个字段)
│   │   ├── config.go         ← 完整配置结构
│   │   ├── loader.go         ← 多层配置加载
│   │   └── claude_compat.go  ← Claude 配置兼容
│   ├── session/              ← 完全缺失
│   │   ├── session.go        ← 会话管理
│   │   └── storage.go        ← 持久化存储
│   ├── prompt/               ← 完全缺失
│   │   └── builder.go        ← System Prompt 构建器
│   ├── agent/
│   │   ├── tools/            ← 大幅缺失 (缺 6+ 工具)
│   │   │   ├── webfetch.go
│   │   │   ├── websearch.go
│   │   │   ├── imagegen.go
│   │   │   ├── todo.go
│   │   │   ├── task.go       ← 子Agent调度
│   │   │   ├── skill.go
│   │   │   └── ask.go
│   │   └── subagent/         ← 完全缺失
│   │       └── subagent.go   ← 子Agent实现
│   ├── worktree/             ← 完全缺失
│   │   └── manager.go        ← Git Worktree 管理
│   └── update/               ← 完全缺失
│       └── updater.go        ← 自更新逻辑
└── tui/
    └── (markdown renderer)   ← 缺失 Glamour 集成
```

### 3.2 功能覆盖率估算

| 模块 | 原版功能 | 反编译实现 | 覆盖率 |
|------|----------|-----------|--------|
| CLI 框架 | Cobra + 25+ flags + 7 子命令 | 无 | 0% |
| Agent 核心 | 消息循环 + 工具调用 + Hook | 基本完整 | 80% |
| Provider | Qoder + OpenAI + Anthropic + IdeaLab + DashScope | 仅 Qoder | 20% |
| 工具系统 | 15+ 工具 | 9 个工具 | 60% |
| 权限系统 | 文件/Bash/MCP/Web 规则 | 基本完整 | 85% |
| MCP | Client + CLI 管理 | 仅 Client | 50% |
| TUI | Bubble Tea + Markdown | Bubble Tea | 70% |
| 配置系统 | 多层配置 + 环境变量 | 1 个字段 | 5% |
| 会话管理 | 持久化 + 恢复 | 无 | 0% |
| Worktree/Jobs | Git worktree + 并发任务 | 无 | 0% |
| **总体** | — | — | **~30%** |

---

## 四、建议实现优先级

### Phase 1: 核心可用 (P0)

1. **引入 Cobra** — 搭建 `cmd/` 框架，绑定所有 flags
2. **完善 Config 加载** — 支持多层配置和环境变量
3. **实现 Session 持久化** — 支持 `-c` / `-r`
4. **构建 System Prompt Builder** — 动态生成系统提示

### Phase 2: 功能完善 (P1)

5. **补全缺失工具** — WebFetch, WebSearch, Task, TodoWrite, AskUserQuestion
6. **实现 MCP 子命令** — add/get/list/remove/auth
7. **实现 Print 模式** — `-p` 非交互输出
8. **MCP 工具桥接** — 将 MCP 工具动态注册到 Agent

### Phase 3: 高级功能 (P2)

9. **Worktree/Jobs 系统** — 并发任务管理
10. **自动更新** — `update` 子命令
11. **Feedback 系统** — 反馈提交
12. **Markdown 渲染** — Glamour 集成
13. **SDK 模式完善** — IDE 集成支持

---

## 五、关键代码缺失清单

### 5.1 需要新建的文件

```
cmd/
├── root.go              # Cobra root command
├── mcp.go               # MCP 子命令入口
├── mcp_add.go           # mcp add
├── mcp_auth.go          # mcp auth
├── mcp_get.go           # mcp get
├── mcp_list.go          # mcp list
├── mcp_remove.go        # mcp remove
├── jobs.go              # jobs 子命令
├── rm.go                # rm 子命令
├── status.go            # status 子命令
├── feedback.go          # feedback 子命令
├── update.go            # update 子命令
└── completion.go        # completion 子命令

core/
├── config/
│   ├── loader.go        # 配置加载器
│   └── claude_compat.go # Claude 配置兼容
├── session/
│   ├── session.go       # 会话管理
│   └── storage.go       # 存储实现
├── prompt/
│   └── builder.go       # System Prompt 构建
├── worktree/
│   └── manager.go       # Git Worktree 管理
├── update/
│   └── updater.go       # 自更新逻辑
└── agent/
    └── tools/
        ├── webfetch.go
        ├── websearch.go
        ├── task.go
        ├── todo.go
        ├── ask.go
        └── imagegen.go
```

### 5.2 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `main.go` | 重写为 Cobra 入口 |
| `core/config/config.go` | 扩展 Config 结构 |
| `core/agent/provider/qoder.go` | 完善 `getMachineID()` |
| `tui/components/chat/chat.go` | 集成 Markdown 渲染 |
| `tui/app/run.go` | 完善 SDK 模式实现 |

---

## 六、总结

当前反编译版本成功还原了 qodercli 的**核心运行时架构**：

- Agent 消息循环和工具调用机制
- Provider 流式通信协议
- 基础工具集（文件操作 + Bash）
- 权限规则匹配系统
- MCP Client 协议实现
- Bubble Tea TUI 框架

但缺失了**整个 CLI 应用层**：

- Cobra 命令框架（所有 flags 和子命令）
- 配置管理和会话持久化
- System Prompt 动态构建
- 多个重要工具
- Worktree 并发任务系统

**建议**: 优先完成 Cobra CLI 框架和配置系统，这是让反编译版本具备基本可用性的前提。

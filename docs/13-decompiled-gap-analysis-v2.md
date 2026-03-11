# qodercli 反编译差距分析报告 V2

> 分析日期: 2026-03-11
> 官方版本: qodercli v0.1.30
> 项目版本: decompiled v0.1.30
> 分析工具: 官方二进制 --help + 源码对比

---

## 一、原版 CLI 参数/子命令全览

### 1.1 子命令列表 (7 个)

| 子命令 | 描述 | 当前状态 | 变化 |
|--------|------|----------|------|
| `jobs` | 列出并发 worktree 任务 (`-a, --all`) | **完全缺失** | ↔ 无变化 |
| `rm` | 删除并发任务（含 k8s 支持 `--kubeconfig`, `--namespace`） | **完全缺失** | ↔ 无变化 |
| `completion` | Shell 自动补全 (bash/zsh/fish/powershell) | **完全缺失** | ↔ 无变化 |
| `feedback` | 提交反馈（`-c`, `-i`, `-s`, `--workdir`） | **完全缺失** | ↔ 无变化 |
| `mcp` | MCP 服务器管理 | **完全缺失** | ↔ 无变化 |
| `status` | 账号/CLI 状态 (`-o text/json`) | **完全缺失** | ↔ 无变化 |
| `update` | 自动更新到最新版本 | **完全缺失** | ↔ 无变化 |

#### MCP 子命令详情

```
qodercli mcp
  add      Add a new MCP server (name, endpoint, -e, -H, -s, -t)
  auth     Authenticate with an MCP server using OAuth
  get      Get details of an MCP server
  list     List all MCP servers
  remove   Remove an MCP Server
```

**新增官方参数** (v0.1.30 相比 v0.1.29):
- `--experimental-mcp-load` - 启用动态发现和加载 MCP 服务器工具（实验性）

### 1.2 主要 Flags (25+ 个)

| Flag | 类型 | 描述 | 当前状态 |
|------|------|------|----------|
| `-p, --print` | string | 非交互模式，单次提示后退出 | ✅ **已实现** (cmd/print/) |
| `-c, --continue` | bool | 继续最近对话 | ⚠️ **有 flag，无实现** |
| `-r, --resume` | string | 恢复指定会话 | ⚠️ **有 flag，无实现** |
| `-w, --workspace` | string | 工作目录 | ✅ **已实现** |
| `-f, --output-format` | string | 输出格式 (text/json/stream-json) | ✅ **已实现** |
| `--input-format` | string | 输入格式 (text/stream-json) | ⚠️ **有 flag，stub 实现** |
| `--model` | string | 模型选择 (10种) | ✅ **已实现** |
| `--max-turns` | int | Agent 最大循环次数 | ✅ **已实现** |
| `--max-output-tokens` | string | 最大输出 token (16k/32k) | ⚠️ **有 flag，未连通** |
| `--allowed-tools` | stringArray | 允许的工具列表 | ✅ **已实现** |
| `--disallowed-tools` | stringArray | 禁止的工具列表 | ✅ **已实现** |
| `--attachment` | stringArray | 附件路径 | ⚠️ **有 flag，部分实现** |
| `--agents` | string | 自定义 Agent JSON 定义 | ⚠️ **有 flag，无实现** |
| `--dangerously-skip-permissions` | bool | 跳过权限检查 | ✅ **已实现** |
| `--yolo` | bool | 同上（别名） | ✅ **已实现** |
| `--worktree` | bool | 通过 git worktree 启动并发任务 | ⚠️ **有 flag，无实现** |
| `--branch` | string | worktree 分支名 | ⚠️ **有 flag，无实现** |
| `--path` | string | worktree 路径 | ⚠️ **有 flag，无实现** |
| `--with-claude-config` | bool | 加载 .claude 配置 | ✅ **已实现** |
| `--experimental-mcp-load` | bool | 动态加载 MCP 工具（实验性） | ❌ **完全缺失** |
| `-q, --quiet` | bool | 非交互模式隐藏 spinner | ⚠️ **有 flag，未完全连通** |
| `-v, --version` | bool | 版本信息 | ✅ **已实现** |
| `-h, --help` | bool | 帮助信息 | ✅ **已实现** |

---

## 二、已完成的改进（相比 V1 报告）

### ✅ 2.1 Cobra CLI 框架

**状态**: 已完成

**实现文件**:
- `cmd/root.go` - Root 命令 + 所有 flags 绑定
- `cmd/print/print.go` - Print 模式实现
- `cmd/tui/tui.go` - TUI 模式入口
- `cmd/utils/` - 配置和工具函数

**代码示例**:
```go
var rootCmd = &cobra.Command{
    Use:   "qodercli",
    Short: "Qoder CLI - AI-powered development assistant",
    Run: func(cmd *cobra.Command, args []string) {
        // 模式路由逻辑
    },
}
```

### ✅ 2.2 System Prompt Builder

**状态**: 已完成

**实现文件**:
- `core/prompts/builder.go` - SystemPromptBuilderV2 实现
- `core/prompts/env_collector.go` - 环境信息收集
- `core/prompts/project_loader.go` - 项目上下文加载
- `core/prompts/builtin.go` - 内置提示词模板

**功能**:
- 基础角色定义
- 工具使用指南
- 权限规则说明
- 环境信息（OS, Git 状态）
- 项目特定指令（AGENTS.md, .claude/）
- 编码规范和最佳实践

### ✅ 2.3 日志系统

**状态**: 已完成（超越原版）

**实现文件**:
- `core/log/log.go` - 完整文件日志系统

**特性**:
- 多级别日志（Debug/Info/Warn/Error/Fatal）
- 文件轮转
- 调用点追踪

### ✅ 2.4 版本管理

**状态**: 已完成

**实现文件**:
- `version/version.go` - 版本定义

---

## 三、仍然缺失的模块

### P0 — 核心缺失

#### 3.1 会话持久化 / 恢复

**原版功能**:
```bash
qodercli -c                    # 继续最近一次对话
qodercli -r <session-id>       # 恢复指定会话
```

**推测实现结构**:
```
~/.qoder/
├── sessions/
│   ├── latest.json            # 最近会话索引
│   └── <session-id>/
│       ├── messages.json      # 消息历史
│       ├── state.json         # Agent 状态
│       └── metadata.json      # 元数据
```

**当前状态**: 
- Flag 已定义 (`-c`, `-r`)
- 无实际存储实现

**需实现**:
```go
core/session/
├── session.go          # 会话管理器
├── storage.go          # 存储接口
└── file_storage.go     # 文件存储实现
```

#### 3.2 缺失的子命令

| 子命令 | 优先级 | 复杂度 | 备注 |
|--------|--------|--------|------|
| `mcp` | P0 | 中 | MCP 服务器管理 |
| `mcp add` | P0 | 中 | 添加 MCP 服务器 |
| `mcp auth` | P1 | 高 | OAuth 认证流程 |
| `mcp get` | P1 | 低 | 查看服务器详情 |
| `mcp list` | P0 | 低 | 列出服务器 |
| `mcp remove` | P1 | 低 | 删除服务器 |
| `jobs` | P1 | 高 | 并发任务列表 |
| `rm` | P1 | 高 | 删除任务 |
| `status` | P2 | 低 | 账号状态 |
| `feedback` | P2 | 中 | 反馈提交 |
| `update` | P2 | 中 | 自动更新 |
| `completion` | P2 | 低 | Cobra 内置 |

#### 3.3 MCP 工具桥接

**当前状态**: 
- ✅ MCP Client 库 (`core/resource/mcp/mcp.go`)
- ❌ MCP 工具未注册到 Agent

**需实现**:
```go
// 将 MCP tools 动态转换为 Agent tools
func (m *MCPClient) ToAgentTools() []tools.Tool {
    // 获取 MCP tools 列表
    // 转换为 Tool 接口
    // 注册到 Agent
}
```

---

### P1 — 重要缺失

#### 3.4 缺失的工具实现

**当前已有工具** (10 个):
| 工具 | 状态 | 文件 |
|------|------|------|
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

**缺失的工具**:
| 工具名 | 功能 | 优先级 | Schema |
|--------|------|--------|--------|
| `Task` | 子 Agent 调度 | 高 | ✅ 已定义 |
| `WebFetch` | 网页抓取 | 高 | ❌ 缺失 |
| `WebSearch` | 网页搜索 | 高 | ❌ 缺失 |
| `AskUserQuestion` | 用户交互询问 | 中 | ❌ 缺失 |
| `ImageGen` | 图片生成 | 中 | ❌ 缺失 |
| `Skill` | 技能系统调用 | 中 | ❌ 缺失 |

**TaskTool Schema** 已定义但未实现:
```go
// BuildTaskSchema 构建 Task 工具 Schema
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

#### 3.5 Worktree / Jobs 系统

**原版功能**:
```bash
qodercli --worktree --branch feature-1
qodercli jobs -a
qodercli rm <id> --kubeconfig ~/.kube/config --namespace default
```

**推测架构**:
```
core/
├── worktree/
│   ├── manager.go          # Git worktree 管理
│   ├── job.go              # 任务定义
│   └── runner.go           # 任务运行器
├── jobs/
│   ├── registry.go         # 任务注册表
│   ├── storage.go          # 任务存储
│   └── kubernetes.go       # K8s Job 支持
```

**当前状态**: 完全缺失（只有 flags）

---

### P2 — 功能增强缺失

#### 3.6 配置加载层完善

**当前状态**:
- `core/config/config.go` - 基础结构
- 环境变量加载

**缺失**:
- 多层配置合并（全局/项目/环境变量）
- MCP 配置管理 (`~/.qoder/.mcp.json`)
- Claude 配置兼容 (`.claude/`)

**目标配置结构**:
```go
type Config struct {
    // CLI 设置
    Model           string
    MaxTokens       int
    MaxTurns        int
    OutputFormat    string
    PermissionMode  permission.Mode
    
    // API 配置
    Provider        string  // openai/qoder/anthropic/idealab/dashscope
    APIKey          string
    BaseURL         string
    
    // 工具配置
    AllowedTools    []string
    DisallowedTools []string
    
    // MCP 配置
    MCPServers      map[string]*mcp.ServerConfig
    
    // 会话配置
    ContinueSession bool
    ResumeSessionID string
    
    // 项目配置
    WithClaudeConfig bool
}
```

#### 3.7 Provider 完善

**当前实现**:
- ✅ OpenAI Provider
- ✅ Qoder Provider

**缺失 Provider**:
- Anthropic
- IdeaLab
- DashScope

#### 3.8 SDK/IDE 集成模式

**原版功能**:
```bash
qodercli --input-format stream-json --output-format stream-json
```

**当前状态**: 
- Flag 已定义
- `tui/app/run.go` 有 `RunSDKMode` stub

**需实现**:
- JSONEncoder/JSONDecoder
- 消息协议处理
- 流式输出处理

---

## 四、架构层面差距总结

### 4.1 目录结构对比

```
原版 qodercli 架构：                    当前项目架构：
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

### 4.2 功能覆盖率估算

| 模块 | 原版功能 | 当前实现 | 覆盖率 | 变化 |
|------|----------|----------|--------|------|
| CLI 框架 | Cobra + 25+ flags + 7 子命令 | Cobra + flags + 基础子命令 | 70% | ↑+30% |
| Agent 核心 | 消息循环 + 工具调用 + Hook | 基本完整 | 80% | ↔ |
| Provider | 5 个 Provider | OpenAI + Qoder | 40% | ↓-20% |
| 工具系统 | 15+ 工具 | 10 个工具 | 65% | ↑+5% |
| 权限系统 | 文件/Bash/MCP/Web 规则 | 基本完整 | 85% | ↔ |
| MCP | Client + CLI 管理 | 仅 Client | 50% | ↔ |
| TUI | Bubble Tea + Markdown | 完整实现 | 95% | ↔ |
| System Prompt | 动态构建 | 完整实现 | 90% | ↑+90% |
| 配置系统 | 多层配置 + 环境变量 | 基础实现 | 50% | ↑+10% |
| 会话管理 | 持久化 + 恢复 | Flag 已定义 | 10% | ↑+10% |
| Worktree/Jobs | Git worktree + 并发任务 | 完全缺失 | 0% | ↔ |
| 日志系统 | 基础 stderr | 完整文件日志 | 150% | ↔ |
| **总体** | — | — | **~60%** | ↑+10% |

---

## 五、建议实现优先级

### Phase 1: 核心功能完善 (P0)

1. **Session 持久化系统**
   - 实现会话存储接口
   - 支持 `-c` / `-r` flags
   - 消息历史持久化

2. **MCP 子命令**
   - `mcp add/get/list/remove`
   - MCP 配置文件管理
   - 工具动态注册到 Agent

3. **Task 工具**
   - Schema 已定义
   - 子 Agent 调度实现

### Phase 2: 功能完整 (P1)

4. **补全缺失工具**
   - WebFetch, WebSearch
   - AskUserQuestion
   - ImageGen (如果依赖允许)

5. **Worktree/Jobs 系统**
   - Git worktree 管理
   - 任务注册表
   - `jobs` / `rm` 子命令

6. **配置系统完善**
   - 多层配置合并
   - MCP 配置管理

### Phase 3: 高级功能 (P2)

7. **其他 Provider**
   - Anthropic
   - IdeaLab
   - DashScope

8. **其他子命令**
   - `status`
   - `feedback`
   - `update`
   - `completion`

9. **SDK 模式完善**
   - JSON 协议实现
   - IDE 集成支持

---

## 六、关键代码缺失清单

### 6.1 需要新建的文件

```
cmd/
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
├── session/
│   ├── session.go       # 会话管理
│   └── storage.go       # 存储实现
├── config/
│   ├── loader.go        # 配置加载器
│   └── mcp_config.go    # MCP 配置
├── worktree/
│   ├── manager.go       # Git Worktree 管理
│   ├── job.go           # 任务定义
│   └── runner.go        # 任务运行器
└── agent/tools/
    ├── task.go          # Task 工具
    ├── webfetch.go      # WebFetch 工具
    ├── websearch.go     # WebSearch 工具
    ├── ask.go           # AskUserQuestion 工具
    └── imagegen.go      # ImageGen 工具
```

### 6.2 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `cmd/root.go` | 添加子命令注册 |
| `core/config/config.go` | 扩展 Config 结构 |
| `core/agent/tools/tools.go` | 注册新工具 |
| `core/resource/mcp/mcp.go` | 添加工具桥接 |
| `tui/app/run.go` | 完善 SDK 模式 |

---

## 七、总结

当前反编译版本相比 V1 报告有显著进步：

### ✅ 已完成

1. **Cobra CLI 框架** - 完整的命令行接口
2. **System Prompt Builder** - 动态系统提示词生成
3. **日志系统** - 超越原版的完整实现
4. **工具系统** - 10 个核心工具
5. **TUI 系统** - 完整的交互界面

### ❌ 仍缺失

1. **7 个子命令** - mcp, jobs, rm, status, feedback, update, completion
2. **会话持久化** - 实现 `-c` / `-r` flags
3. **5+ 工具** - Task, WebFetch, WebSearch, AskUser, ImageGen
4. **MCP 工具桥接** - 动态注册 MCP 工具到 Agent
5. **Worktree/Jobs** - 并发任务管理

### 总体评估

| 维度 | 评分 | 说明 |
|------|------|------|
| CLI 框架 | 70% | Cobra 已完成，缺子命令 |
| Agent 核心 | 80% | 稳定可用 |
| 工具系统 | 65% | 核心工具可用 |
| TUI 体验 | 95% | 接近原版 |
| 配置管理 | 50% | 基础功能可用 |
| 扩展功能 | 30% | 缺失较多 |
| **总体** | **60%** | **可用性大幅提升** |

**建议**: 当前版本已具备基本可用性，建议优先实现 **Session 持久化** 和 **Task 工具**，这两个功能对用户体验提升最大。

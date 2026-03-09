# qodercli 完整架构概览

## 项目信息

- **内部 Module 路径**: `code.alibaba-inc.com/qoder-core/qodercli`
- **版本**: v0.1.29
- **平台**: darwin arm64 (Mach-O 64-bit executable)
- **语言**: Go 1.25.7
- **文件大小**: ~37.2 MB

---

## 1. 整体架构图

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
│  │ ├── agent/           # Agent 主循环                             │   │
│  │ ├── provider/        # LLM Provider (Qoder/OpenAI/IdeaLab)      │   │
│  │ ├── tools/           # 内置工具 (Bash/Read/Write/Edit/...)     │   │
│  │ ├── permission/      # 权限系统                                  │   │
│  │ ├── hooks/           # Hook 系统                                 │   │
│  │ ├── prompt/          # Prompt 管理                               │   │
│  │ ├── state/           # 状态管理                                  │   │
│  │ ├── compact/         # 上下文压缩                                │   │
│  │ └── task/            # SubAgent 调度                             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Resource Layer                                  │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ core/resource/                                                  │   │
│  │ ├── mcp/             # MCP 服务器管理                            │   │
│  │ ├── skill/           # Skill 系统                                │   │
│  │ ├── subagent/        # SubAgent 定义                             │   │
│  │ ├── command/         # 斜杠命令                                  │   │
│  │ ├── hook/            # Hook 定义                                 │   │
│  │ ├── plugin/          # 插件系统                                  │   │
│  │ └── output_style/    # 输出样式                                  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          Utility Layer                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ core/utils/                                                     │   │
│  │ ├── qoder/           # Qoder API 客户端                         │   │
│  │ ├── http/            # HTTP 客户端                               │   │
│  │ ├── tokens/          # Token 计算                                │   │
│  │ ├── storage/         # 本地存储                                  │   │
│  │ ├── sls/             # 日志上报                                  │   │
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

## 2. 核心数据流

### 2.1 用户输入 → Agent 响应

```
用户输入
    │
    ▼
TUI 编辑器 (tui/components/interaction/editor/)
    │
    ▼
用户提交 Hook (core/agent/hooks/user_prompt_submit/)
    │
    ▼
Agent 主循环 (core/agent/agent/)
    │
    ├── 构建消息 (core/agent/state/message/)
    │
    ├── 调用 Provider (core/agent/provider/)
    │   │
    │   ▼
    │   LLM API (Qoder/OpenAI/Anthropic/IdeaLab)
    │   │
    │   ▼
    │   流式响应事件
    │
    ├── 解析工具调用
    │   │
    │   ▼
    │   权限检查 (core/agent/permission/)
    │   │
    │   ▼
    │   PreToolUse Hook
    │   │
    │   ▼
    │   工具执行 (core/agent/tools/)
    │   │
    │   ▼
    │   PostToolUse Hook
    │
    └── 循环直到完成
```

### 2.2 MCP 工具调用

```
Agent 决定调用 MCP 工具
    │
    ▼
MCP 规则匹配 (core/agent/permission/mcp_rule_matcher/)
    │
    ▼
MCP 客户端 (core/resource/mcp/)
    │
    ├── stdio/SSE 通信
    │
    ▼
MCP 服务器进程
    │
    ▼
返回结果
```

---

## 3. 关键类型定义

### 3.1 消息类型

```go
type Message struct {
    Role       string        // "user", "assistant", "system"
    Content    []ContentPart // 多部分内容
    ToolCalls  []ToolCall    // 工具调用
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

### 3.2 Provider 接口

```go
type Client interface {
    Stream(ctx context.Context, req ModelRequest) <-chan Event
    Send(ctx context.Context, req ModelRequest) (*Response, error)
}
```

### 3.3 工具接口

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, input string) (string, error)
}
```

---

## 4. 配置层次结构

```
优先级 (低 → 高):

1. 内置默认值
2. ~/.qoder/config.json          # 全局配置
3. ~/.qoder/settings.json        # 全局设置
4. ./.qoder/settings.json        # 项目设置
5. ./.qoder/settings.local.json  # 本地覆盖
6. ./.claude/CLAUDE.md           # Claude 兼容配置
7. 环境变量 (QODER_*)
8. 命令行参数
```

---

## 5. 运行模式

| 模式 | 说明 | 入口 |
|------|------|------|
| **交互模式** | TUI 交互式对话 | `qodercli` |
| **非交互模式** | 单次查询后退出 | `qodercli -p "..."` |
| **SDK 模式** | JSON 协议通信 | stdin/stdout |
| **ACP Server** | Agent Communication Protocol | 内部服务 |
| **Worktree** | Git worktree 并发 | `--worktree` |
| **Container** | Docker 容器 | 内部模式 |
| **Kubernetes** | K8s Job | 内部模式 |

---

## 6. 第三方依赖

| 类别 | 依赖 |
|------|------|
| CLI 框架 | `github.com/spf13/cobra` |
| TUI 框架 | Bubble Tea (推断) |
| LLM SDK | `github.com/anthropics/anthropic-sdk-go` |
| GitHub API | `github.com/google/go-github/v57` |
| 配置解析 | `github.com/BurntSushi/toml` |
| 语法高亮 | `github.com/alecthomas/chroma/v2` |
| HTML 转换 | `github.com/JohannesKaufmann/html-to-markdown/v2` |
| 压缩 | `github.com/andybalholm/brotli` |
| HTTPDNS | `github.com/aliyun/alicloud-httpdns-go-sdk` |

---

## 7. 文件清单

| 文档 | 说明 |
|------|------|
| `PLAN.md` | 任务规划 |
| `01-package-structure.md` | 包结构和依赖图 |
| `02-cli-commands.md` | CLI 命令系统 |
| `03-llm-integration.md` | AI/LLM 集成 |
| `04-mcp-protocol.md` | MCP 协议实现 |
| `05-tools-system.md` | 工具系统 |
| `06-session-management.md` | 会话管理 |
| `07-permission-security.md` | 权限和安全 |
| `08-config-api.md` | 配置和 API |
| `09-github-integration.md` | GitHub 集成 |
| `10-subagent-skill.md` | SubAgent 和 Skill |
| `architecture-overview.md` | 本文档 |

---

## 8. 总结

qodercli 是一个功能完善的 AI 编程助手 CLI 工具，具有以下特点：

1. **多 Provider 支持**: Qoder、Anthropic、OpenAI、IdeaLab、DashScope
2. **丰富的工具集**: Bash、文件操作、搜索、网络请求等
3. **MCP 协议**: 可扩展的工具和资源集成
4. **权限系统**: 细粒度的操作权限控制
5. **Hook 系统**: 可定制的工作流钩子
6. **SubAgent**: 并行任务执行能力
7. **Skill 系统**: 可复用的技能定义
8. **GitHub 集成**: PR 审查、Actions 集成
9. **多运行模式**: 交互、非交互、SDK、容器化

该项目代码结构清晰，模块化程度高，是一个典型的 Go CLI 应用架构。

# qodercli 反编译代码

本目录包含基于二进制分析和文档推导的反编译 Go 代码。

## 目录结构

```
decompiled/
├── core/
│   ├── types/
│   │   └── types.go              # 核心类型定义
│   ├── agent/
│   │   ├── agent/
│   │   │   └── agent.go          # Agent 主循环
│   │   ├── provider/
│   │   │   ├── provider.go       # Provider 接口
│   │   │   └── qoder.go          # Qoder Provider 实现
│   │   ├── tools/
│   │   │   ├── tools.go          # 工具系统接口
│   │   │   ├── bash.go           # Bash 工具实现
│   │   │   └── file.go           # 文件操作工具
│   │   └── permission/
│   │       └── permission.go     # 权限系统
│   ├── resource/
│   │   └── mcp/
│   │       └── mcp.go            # MCP 客户端
│   └── pubsub/
│       └── pubsub.go             # 事件发布订阅
├── tui/
│   ├── app/
│   │   ├── model.go              # TUI 应用模型
│   │   └── run.go                # 应用启动器
│   └── components/
│       ├── editor/
│       │   └── editor.go         # 消息编辑器
│       ├── messages/
│       │   └── messages.go       # 消息列表
│       ├── chat/
│       │   └── chat.go           # 聊天视图
│       └── status/
│           └── status.go         # 状态栏
└── README.md                     # 本文件
```

## 代码统计

| 模块 | 文件数 | 代码行数 | 说明 |
|------|--------|----------|------|
| core/types | 1 | ~200 | 基础类型定义 |
| core/agent | 6 | ~1800 | Agent 核心实现 |
| core/resource | 1 | ~350 | MCP 客户端 |
| core/pubsub | 1 | ~100 | 事件系统 |
| tui | 6 | ~900 | TUI 组件 |
| **总计** | **15** | **~3350** | |

## 核心架构

### 1. Agent 架构

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

### 2. Provider 架构

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

### 3. 工具系统架构

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

## 关键类型

### Message 结构
```go
type Message struct {
    Role       Role          `json:"role"`
    Content    []ContentPart `json:"content,omitempty"`
    ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
    ToolCallID string        `json:"tool_call_id,omitempty"`
}
```

### ToolCall 结构
```go
type ToolCall struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

### Provider 接口
```go
type Client interface {
    Stream(ctx context.Context, req *ModelRequest) <-chan Event
    Send(ctx context.Context, req *ModelRequest) (*Response, error)
    Close() error
}
```

## 注意事项

1. **代码性质**: 本代码基于二进制反编译和文档推导，可能与原始代码存在差异
2. **完整性**: 部分内部实现细节可能不完整或需要调整
3. **依赖关系**: 代码中使用了推断的包路径 `code.alibaba-inc.com/qoder-core/qodercli`
4. **第三方库**: 依赖 Bubble Tea、Lipgloss、Cobra 等开源库

## 使用建议

这些反编译代码可用于：
- 理解 qodercli 的架构设计
- 学习 Agent 系统的实现模式
- 参考 TUI 应用的构建方式
- 分析 MCP 协议的实现

不建议直接用于生产环境，因为：
- 代码未经测试
- 可能包含不准确的推断
- 缺少原始项目的完整上下文

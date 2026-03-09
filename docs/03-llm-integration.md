# qodercli AI/LLM 集成分析

## 1. Provider 架构

### 1.1 Provider 接口层级

```
core/agent/provider/
├── provider/          # Provider 接口定义
├── qoder/             # Qoder 自有 Provider (默认)
├── openai/            # OpenAI 兼容 Provider
├── idealab/           # IdeaLab Provider
├── think/             # 思考/推理层
├── options/           # Provider 选项
└── idealab/           # IdeaLab Provider
```

### 1.2 核心类型

```go
// Provider 客户端接口
type Client interface {
    Stream(ctx context.Context, req ModelRequest) <-chan Event
    Send(ctx context.Context, req ModelRequest) (*Response, error)
}

// 客户端构建器
type ClientBuilder interface {
    Build() (Client, error)
}

// 流式事件
type Event struct {
    Type    EventType
    Content string
    // ...
}

type EventType int  // 事件类型枚举

// 模型请求
type ModelRequest struct {
    Model       string
    Messages    []Message
    Tools       []Tool
    MaxTokens   int
    Temperature float64
    // ...
}

// 响应
type Response struct {
    Content     string
    ToolCalls   []ToolCall
    FinishReason FinishReason
    TokenUsage  TokenUsage
}

// Token 使用量
type TokenUsage struct {
    InputTokens  int
    OutputTokens int
    TotalTokens  int
}

// 思考级别
type ThinkLevel int

// 推理努力程度
type ReasoningEffort string
```

### 1.3 具体客户端实现

```
provider.QoderClient      # Qoder 自有 API 客户端
provider.OpenAIClient     # OpenAI 兼容 API 客户端
provider.IdeaLabClient    # IdeaLab API 客户端
provider.baseClient       # 基础客户端 (泛型模板)
```

## 2. 支持的模型

### 2.1 Claude 系列 (Anthropic)

| 模型 ID | 显示名称 |
|---------|----------|
| `claude-3-haiku-20240307` | Claude 3 Haiku |
| `claude-3-5-haiku-latest` | Claude 3.5 Haiku |
| `claude-3-5-sonnet-latest` | Claude 3.5 Sonnet |
| `claude-3-7-sonnet-20250219` | Claude 3.7 Sonnet |
| `claude-sonnet-4-20250514` | Claude 4 Sonnet |
| `claude-opus-4-0` | Claude Opus 4 |
| `anthropic.claude-opus-4-20250514-v1:0` | Claude Opus 4 (AWS Bedrock) |

### 2.2 GPT 系列 (OpenAI)

| 模型 ID | 显示名称 |
|---------|----------|
| `gpt-4o` | GPT 4o |
| `gpt-4o-mini` | GPT 4o mini |
| `gpt-4.1` | GPT 4.1 |
| `gpt-4.1-mini` | GPT 4.1 mini |
| `gpt-4.1-nano` | GPT 4.1 nano |
| `gpt-4.5-preview` | GPT 4.5 preview |
| `o1-pro` | o1 Pro |
| `o1-mini` | o1 Mini |
| `o3-mini` | o3 Mini |
| `o4-mini` | o4 Mini |
| `OpenAI GPT-5` | GPT-5 |

### 2.3 Qwen 系列 (DashScope)

| 模型 ID | 显示名称 |
|---------|----------|
| `qwen3-coder-plus` | Qwen3 Coder Plus |

### 2.4 Qoder 模型级别

| 级别 | 说明 |
|------|------|
| `auto` | 自动选择最优模型 |
| `efficient` | 高效模式 (快速响应) |
| `performance` | 性能模式 (平衡) |
| `ultimate` | 终极模式 (最强) |
| `lite` | 轻量模式 |
| `qmodel` | Q 模型 |
| `q35model` | Q3.5 模型 |
| `gmodel` | G 模型 (GPT) |
| `kmodel` | K 模型 (Kimi) |
| `mmodel` | M 模型 |

## 3. 认证机制

### 3.1 认证方式

```
core/auth/model/
├── anthropic/    # Anthropic API Key 认证
├── openai/       # OpenAI API Key 认证
├── dashscope/    # DashScope API Key 认证
├── idealab/      # IdeaLab API Key 认证
├── qoder/        # Qoder OAuth/Token 认证
├── models/       # 模型列表
└── filter/       # 模型过滤器
```

### 3.2 环境变量配置

| Provider | 环境变量 |
|----------|----------|
| Qoder | `QODER_PERSONAL_ACCESS_TOKEN` |
| Anthropic | `QODER_ANTHROPIC_API_KEY` |
| OpenAI | `QODER_OPENAI_API_KEY` |
| DashScope | `QODER_DASHSCOPE_API_KEY` |
| IdeaLab | `QODER_IDEALAB_API_KEY` |

### 3.3 认证流程

```
1. 检查环境变量 QODER_PERSONAL_ACCESS_TOKEN
2. 如未设置，检查本地存储的 token (~/.qoder/)
3. 如无有效 token，启动 OAuth 登录流程:
   a. 打开浏览器访问 qoder.com/oauth
   b. 轮询设备 token (Device Token Polling)
   c. 获取 access_token 并存储
4. 刷新 token (如已过期)
```

## 4. 流式传输

### 4.1 事件类型

```
content_block_start    # 内容块开始
content_block_delta    # 内容块增量
content_block_stop     # 内容块结束
tool_use_start         # 工具调用开始
tool_use_delta         # 工具调用增量
tool_use_stop          # 工具调用结束
message_start          # 消息开始
message_delta          # 消息增量
message_stop           # 消息结束
thinking_delta         # 思考增量
```

### 4.2 流式处理流程

```
Provider.Stream() -> chan Event
    │
    ├── Event{Type: message_start}
    ├── Event{Type: content_block_start}
    ├── Event{Type: content_block_delta, Content: "..."} (多次)
    ├── Event{Type: content_block_stop}
    │
    ├── Event{Type: tool_use_start, ToolName: "Bash"}
    ├── Event{Type: tool_use_delta, Arguments: "..."} (多次)
    ├── Event{Type: tool_use_stop}
    │
    └── Event{Type: message_stop, FinishReason: "tool_use"}
```

## 5. 消息格式转换

### 5.1 转换器

```
core/utils/converter/
├── anthropic/    # Anthropic 格式 <-> 内部格式
└── qoder/        # Qoder 格式 <-> 内部格式
```

### 5.2 消息类型

```go
type Message struct {
    Role       string        // "user", "assistant", "system"
    Content    []ContentPart // 多部分内容
    ToolCalls  []ToolCall    // 工具调用
    ToolCallID string        // 工具调用 ID (用于 tool 结果)
}

type ContentPart struct {
    Type string // "text", "image", "thinking"
    Text string
    ImageSource *ImageSource
}

type ImageSource struct {
    Type      string // "base64", "url"
    MediaType string // "image/png", "image/jpeg"
    Data      string // base64 数据或 URL
}

type ToolCall struct {
    ID        string
    Name      string
    Arguments string // JSON 字符串
}
```

## 6. Token 管理

### 6.1 Token 计算

```
core/utils/tokens/
└── tokens.go    # Token 计算和追踪
```

### 6.2 相关环境变量

```
QODER_SHOW_TOKEN_USAGE     # 显示 Token 用量
QODER_EXPOSE_TOKEN_USAGE   # 暴露 Token 用量给 API
```

### 6.3 Token 使用统计

```go
type TokenUsage struct {
    InputTokens         int
    OutputTokens        int
    TotalTokens         int
    CachedTokens        int      // 缓存命中的 Token
    PreContextTokens    int      // 上下文前 Token
    PostContextTokens   int      // 上下文后 Token
    CompressionTokens   int      // 压缩节省的 Token
}
```

## 7. API 端点

### 7.1 Qoder 自有 API

| 端点 | 说明 |
|------|------|
| `https://openapi.qoder.sh` | OpenAPI 服务 |
| `https://center.qoder.sh` | 中心服务 |
| `https://daily.qoder.ai` | 日常服务 |
| `https://api2.qoder.sh` | API v2 |
| `https://qts2.qoder.sh` | QTS 服务 |
| `https://test.qoder.ai` | 测试环境 |

### 7.2 API 路径

| 路径 | 说明 |
|------|------|
| `/api/v1/userinfo` | 用户信息 |
| `/api/v1/deviceToken/poll` | 设备 Token 轮询 |
| `/api/v1/tracking` | 事件追踪 |
| `/api/v1/heartbeat` | 心跳 |
| `/api/v2/model/list` | 模型列表 |
| `/api/v2/quota/usage` | 配额使用 |
| `/api/v2/user/plan` | 用户计划 |
| `/api/v3/user/status` | 用户状态 |
| `/algo/api/v2/user/plan` | 用户计划 (算法) |
| `/algo/api/v2/quota/usage` | 配额使用 (算法) |

## 8. 请求头

```
Cosy-ClientIp        # 客户端 IP
Cosy-ClientType      # 客户端类型
Cosy-MachineType     # 机器类型
Cosy-Data-Policy     # 数据策略
Cosy-User            # 用户标识
Cosy-Date            # 日期
Cosy-Version         # 版本
Cosy-Organization-Id # 组织 ID
Cosy-Organization-Tags # 组织标签
Cosy-Key             # 密钥
X-Model-Key          # 模型密钥
X-Model-Source       # 模型来源
X-Request-Id         # 请求 ID
X-Machine-Id         # 机器 ID
X-Machine-OS         # 机器操作系统
X-IDE-Platform       # IDE 平台
X-Client-Timestamp   # 客户端时间戳
```

## 9. 特殊功能

### 9.1 思考模式 (Thinking)

支持扩展思考的模型可以启用思考模式:

```
\bthink about it\b
\bthink intensely\b
\bthink very hard\b
\bthink hard\b
\bthink more\b
\bultrathink\b
\bdenk gründlich nach\b      # 德语
\bnachdenken\b               # 德语
\briflettere\b               # 意大利语
\bpensare profondamente\b    # 意大利语
\bpensando\b                 # 西班牙语
\bpiensa profundamente\b     # 西班牙语
```

### 9.2 缓存控制

```
cache_control
budget_tokens
cache_creation_input_tokens
```

### 9.3 引用 (Citations)

支持返回回答中的引用来源。

## 10. 错误处理

### 10.1 错误类型

```go
type ErrorResponse struct {
    Type    string // "error"
    Error   ErrorData
}

type ErrorData struct {
    Type    string // "invalid_request_error", "authentication_error", etc.
    Message string
}
```

### 10.2 常见错误

| 错误 | 说明 |
|------|------|
| `invalid token` | Token 无效 |
| `model not available` | 模型不可用 |
| `token output exceeded` | 输出 Token 超限 |
| `exceeded max turns` | 超过最大循环次数 |
| `[EXCEED_QUOTA]` | 配额超限 |
| `ip_banned_error` | IP 被封禁 |
| `app_disabled_error` | 应用被禁用 |

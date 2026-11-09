# qodercli-reverse AI/LLM Integration Analysis

## 1. Provider Architecture

### 1.1 Provider Interface Hierarchy

```
core/agent/provider/
├── provider/          # Provider interface definition
├── qoder/             # Qoder proprietary Provider (default)
├── openai/            # OpenAI-compatible Provider
├── idealab/           # IdeaLab Provider
├── think/             # Thinking/reasoning layer
├── options/           # Provider options
└── idealab/           # IdeaLab Provider
```

### 1.2 Core Types

```go
// Provider client interface
type Client interface {
    Stream(ctx context.Context, req ModelRequest) <-chan Event
    Send(ctx context.Context, req ModelRequest) (*Response, error)
}

// Client builder
type ClientBuilder interface {
    Build() (Client, error)
}

// Streaming event
type Event struct {
    Type    EventType
    Content string
    // ...
}

type EventType int  // Event type enum

// Model request
type ModelRequest struct {
    Model       string
    Messages    []Message
    Tools       []Tool
    MaxTokens   int
    Temperature float64
    // ...
}

// Response
type Response struct {
    Content     string
    ToolCalls   []ToolCall
    FinishReason FinishReason
    TokenUsage  TokenUsage
}

// Token usage
type TokenUsage struct {
    InputTokens  int
    OutputTokens int
    TotalTokens  int
}

// Thinking level
type ThinkLevel int

// Reasoning effort
type ReasoningEffort string
```

### 1.3 Concrete Client Implementations

```
provider.QoderClient      # Qoder proprietary API client
provider.OpenAIClient     # OpenAI-compatible API client
provider.IdeaLabClient    # IdeaLab API client
provider.baseClient       # Base client (generic template)
```

## 2. Supported Models

### 2.1 Claude Series (Anthropic)

| Model ID | Display Name |
|----------|--------------|
| `claude-3-haiku-20240307` | Claude 3 Haiku |
| `claude-3-5-haiku-latest` | Claude 3.5 Haiku |
| `claude-3-5-sonnet-latest` | Claude 3.5 Sonnet |
| `claude-3-7-sonnet-20250219` | Claude 3.7 Sonnet |
| `claude-sonnet-4-20250514` | Claude 4 Sonnet |
| `claude-opus-4-0` | Claude Opus 4 |
| `anthropic.claude-opus-4-20250514-v1:0` | Claude Opus 4 (AWS Bedrock) |

### 2.2 GPT Series (OpenAI)

| Model ID | Display Name |
|----------|--------------|
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

### 2.3 Qwen Series (DashScope)

| Model ID | Display Name |
|----------|--------------|
| `qwen3-coder-plus` | Qwen3 Coder Plus |

### 2.4 Qoder Model Levels

| Level | Description |
|-------|-------------|
| `auto` | Automatically select the best model |
| `efficient` | Efficient mode (fast response) |
| `performance` | Performance mode (balanced) |
| `ultimate` | Ultimate mode (most powerful) |
| `lite` | Lightweight mode |
| `qmodel` | Q model |
| `q35model` | Q3.5 model |
| `gmodel` | G model (GPT) |
| `kmodel` | K model (Kimi) |
| `mmodel` | M model |

## 3. Authentication Mechanism

### 3.1 Authentication Methods

```
core/auth/model/
├── anthropic/    # Anthropic API Key authentication
├── openai/       # OpenAI API Key authentication
├── dashscope/    # DashScope API Key authentication
├── idealab/      # IdeaLab API Key authentication
├── qoder/        # Qoder OAuth/Token authentication
├── models/       # Model list
└── filter/       # Model filter
```

### 3.2 Environment Variable Configuration

| Provider | Environment Variable |
|----------|----------------------|
| Qoder | `QODER_PERSONAL_ACCESS_TOKEN` |
| Anthropic | `QODER_ANTHROPIC_API_KEY` |
| OpenAI | `QODER_OPENAI_API_KEY` |
| DashScope | `QODER_DASHSCOPE_API_KEY` |
| IdeaLab | `QODER_IDEALAB_API_KEY` |

### 3.3 Authentication Flow

```
1. Check environment variable QODER_PERSONAL_ACCESS_TOKEN
2. If not set, check locally stored token (~/.qoder/)
3. If no valid token, start OAuth login flow:
   a. Open browser to qoder.com/oauth
   b. Poll device token (Device Token Polling)
   c. Obtain access_token and store it
4. Refresh token (if expired)
```

## 4. Streaming

### 4.1 Event Types

```
content_block_start    # Content block start
content_block_delta    # Content block delta
content_block_stop     # Content block end
tool_use_start         # Tool call start
tool_use_delta         # Tool call delta
tool_use_stop          # Tool call end
message_start          # Message start
message_delta          # Message delta
message_stop           # Message end
thinking_delta         # Thinking delta
```

### 4.2 Streaming Processing Flow

```
Provider.Stream() -> chan Event
    │
    ├── Event{Type: message_start}
    ├── Event{Type: content_block_start}
    ├── Event{Type: content_block_delta, Content: "..."} (multiple)
    ├── Event{Type: content_block_stop}
    │
    ├── Event{Type: tool_use_start, ToolName: "Bash"}
    ├── Event{Type: tool_use_delta, Arguments: "..."} (multiple)
    ├── Event{Type: tool_use_stop}
    │
    └── Event{Type: message_stop, FinishReason: "tool_use"}
```

## 5. Message Format Conversion

### 5.1 Converters

```
core/utils/converter/
├── anthropic/    # Anthropic format <-> internal format
└── qoder/        # Qoder format <-> internal format
```

### 5.2 Message Types

```go
type Message struct {
    Role       string        // "user", "assistant", "system"
    Content    []ContentPart // Multi-part content
    ToolCalls  []ToolCall    // Tool calls
    ToolCallID string        // Tool call ID (for tool result)
}

type ContentPart struct {
    Type string // "text", "image", "thinking"
    Text string
    ImageSource *ImageSource
}

type ImageSource struct {
    Type      string // "base64", "url"
    MediaType string // "image/png", "image/jpeg"
    Data      string // base64 data or URL
}

type ToolCall struct {
    ID        string
    Name      string
    Arguments string // JSON string
}
```

## 6. Token Management

### 6.1 Token Calculation

```
core/utils/tokens/
└── tokens.go    # Token calculation and tracking
```

### 6.2 Related Environment Variables

```
QODER_SHOW_TOKEN_USAGE     # Show token usage
QODER_EXPOSE_TOKEN_USAGE   # Expose token usage to API
```

### 6.3 Token Usage Statistics

```go
type TokenUsage struct {
    InputTokens         int
    OutputTokens        int
    TotalTokens         int
    CachedTokens        int      // Tokens hit from cache
    PreContextTokens    int      // Tokens before context
    PostContextTokens   int      // Tokens after context
    CompressionTokens   int      // Tokens saved by compaction
}
```

## 7. API Endpoints

### 7.1 Qoder Proprietary API

| Endpoint | Description |
|----------|-------------|
| `https://openapi.qoder.sh` | OpenAPI service |
| `https://center.qoder.sh` | Center service |
| `https://daily.qoder.ai` | Daily service |
| `https://api2.qoder.sh` | API v2 |
| `https://qts2.qoder.sh` | QTS service |
| `https://test.qoder.ai` | Test environment |

### 7.2 API Paths

| Path | Description |
|------|-------------|
| `/api/v1/userinfo` | User info |
| `/api/v1/deviceToken/poll` | Device token polling |
| `/api/v1/tracking` | Event tracking |
| `/api/v1/heartbeat` | Heartbeat |
| `/api/v2/model/list` | Model list |
| `/api/v2/quota/usage` | Quota usage |
| `/api/v2/user/plan` | User plan |
| `/api/v3/user/status` | User status |
| `/algo/api/v2/user/plan` | User plan (algorithm) |
| `/algo/api/v2/quota/usage` | Quota usage (algorithm) |

## 8. Request Headers

```
Cosy-ClientIp        # Client IP
Cosy-ClientType      # Client type
Cosy-MachineType     # Machine type
Cosy-Data-Policy     # Data policy
Cosy-User            # User identifier
Cosy-Date            # Date
Cosy-Version         # Version
Cosy-Organization-Id # Organization ID
Cosy-Organization-Tags # Organization tags
Cosy-Key             # Secret key
X-Model-Key          # Model key
X-Model-Source       # Model source
X-Request-Id         # Request ID
X-Machine-Id         # Machine ID
X-Machine-OS         # Machine OS
X-IDE-Platform       # IDE platform
X-Client-Timestamp   # Client timestamp
```

## 9. Special Features

### 9.1 Thinking Mode

Models supporting extended thinking can enable thinking mode:

```
\bthink about it\b
\bthink intensely\b
\bthink very hard\b
\bthink hard\b
\bthink more\b
\bultrathink\b
\bdenk gründlich nach\b      # German
\b nachdenken\b               # German
\briflettere\b               # Italian
\bpensare profondamente\b    # Italian
\bpensando\b                 # Spanish
\bpiensa profundamente\b     # Spanish
```

### 9.2 Cache Control

```
cache_control
budget_tokens
cache_creation_input_tokens
```

### 9.3 Citations

Supports returning citation sources in answers.

## 10. Error Handling

### 10.1 Error Types

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

### 10.2 Common Errors

| Error | Description |
|-------|-------------|
| `invalid token` | Invalid token |
| `model not available` | Model unavailable |
| `token output exceeded` | Output token limit exceeded |
| `exceeded max turns` | Maximum loop iterations exceeded |
| `[EXCEED_QUOTA]` | Quota exceeded |
| `ip_banned_error` | IP banned |
| `app_disabled_error` | Application disabled |<|end|><|start|>write to=assistant<|channel|>analysis<|message|>Wrote file successfully: docs/03-llm-integration.md<|end|><|end|>assistant<|channel|>analysis<|message|>Wait, in 9.1 I accidentally wrote "\b nachdenken\b" with a space. Original was "\bnachdenken\b". Let me fix.<|end|><|start|> fileList <|constrain|>xml<|message|><rfl:parameter name="oldString">
\bdenk gründlich nach\b      # German
\b nachdenken\b               # German
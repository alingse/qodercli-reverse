# qodercli MCP 协议实现分析

## 1. MCP 架构概述

MCP (Model Context Protocol) 是 qodercli 用于扩展 AI 能力的核心协议，允许集成外部工具、资源和提示。

### 1.1 包结构

```
core/resource/mcp/
├── mcp/                    # MCP 核心实现
├── mcp_ipc_client_unix/    # Unix IPC 客户端
├── mcps/                   # MCP 服务器列表管理
├── process/                # MCP 进程管理
└── token_store/            # MCP Token 存储

cmd/mcp/
├── add/                    # 添加 MCP 服务器
├── auth/                   # MCP OAuth 认证
├── get/                    # 获取 MCP 服务器信息
├── list/                   # 列出 MCP 服务器
└── remove/                 # 移除 MCP 服务器
```

## 2. MCP 服务器配置

### 2.1 配置文件位置

```
~/.mcp.json           # 全局 MCP 配置
./.mcp.json           # 项目级 MCP 配置
```

### 2.2 配置格式

```json
{
  "mcpServers": {
    "server-name": {
      "command": "node",
      "args": ["server.js"],
      "env": {
        "API_KEY": "..."
      },
      "cwd": "/path/to/server"
    }
  }
}
```

### 2.3 服务器类型

- **stdio**: 通过标准输入输出通信
- **sse**: 通过 Server-Sent Events 通信
- **http**: 通过 HTTP 通信

## 3. MCP 协议方法

### 3.1 核心方法

| 方法 | 说明 |
|------|------|
| `initialize` | 初始化连接 |
| `tools/list` | 列出可用工具 |
| `tools/call` | 调用工具 |
| `resources/list` | 列出可用资源 |
| `resources/read` | 读取资源 |
| `resources/templates/list` | 列出资源模板 |
| `resources/subscribe` | 订阅资源更新 |
| `resources/unsubscribe` | 取消订阅 |
| `prompts/list` | 列出可用提示 |
| `prompts/get` | 获取提示 |
| `logging/setLevel` | 设置日志级别 |

### 3.2 通知类型

| 通知 | 说明 |
|------|------|
| `notifications/resources/list_changed` | 资源列表变更 |
| `notifications/prompts/list_changed` | 提示列表变更 |
| `notifications/tools/list_changed` | 工具列表变更 |

## 4. 工具集成

### 4.1 工具命名规范

MCP 工具在 qodercli 中以 `mcp__<server>__<tool>` 格式命名：

```
mcp__browser-use__*          # browser-use 服务器的所有工具
mcp__quest__search_codebase  # quest 服务器的搜索代码库工具
mcp__acp__delete_file        # ACP 服务器的删除文件工具
```

### 4.2 工具调用流程

```
1. Agent 决定调用 MCP 工具
2. core/resource/mcp/mcp.go 查找对应服务器
3. 通过 IPC/HTTP 发送 tools/call 请求
4. MCP 服务器执行工具
5. 返回结果给 Agent
```

### 4.3 工具定义格式

```json
{
  "name": "tool_name",
  "description": "Tool description",
  "inputSchema": {
    "type": "object",
    "properties": {
      "param1": {
        "type": "string",
        "description": "Parameter description"
      }
    },
    "required": ["param1"]
  }
}
```

## 5. 资源管理

### 5.1 资源类型

- **文本资源**: `text/plain`, `text/markdown`, `text/html`
- **二进制资源**: 通过 base64 编码
- **资源链接**: `resource_link`

### 5.2 资源 URI 格式

```
file:///path/to/file          # 文件资源
qoder:///agent/...            # Qoder 内部资源
```

## 6. OAuth 认证

### 6.1 OAuth 流程

```
1. 用户执行 `qodercli mcp auth <server>`
2. 发现 OAuth 配置 (.well-known/oauth-protected-resource)
3. 生成 code_verifier 和 code_challenge
4. 打开浏览器进行授权
5. 回调获取 authorization_code
6. 交换 access_token
7. 存储 token 到 token_store
```

### 6.2 相关字段

```
response_type        # "code"
code_challenge       # PKCE challenge
code_challenge_method # "S256"
redirect_uri         # 回调地址
client_id            # 客户端 ID
client_secret        # 客户端密钥 (可选)
```

## 7. 进程管理

### 7.1 进程生命周期

```
core/resource/mcp/process/
├── 启动 MCP 服务器进程
├── 管理进程状态
├── 处理进程崩溃和重启
└── 清理进程资源
```

### 7.2 IPC 通信 (Unix)

```
core/resource/mcp/mcp_ipc_client_unix/
├── Unix socket 通信
├── JSON-RPC 消息封装
└── 异步消息处理
```

## 8. JSON-RPC 协议

### 8.1 请求格式

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {
      "param": "value"
    }
  }
}
```

### 8.2 响应格式

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Result text"
      }
    ]
  }
}
```

### 8.3 错误格式

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32600,
    "message": "Invalid Request",
    "data": {}
  }
}
```

## 9. MCP 与 Agent 集成

### 9.1 工具注册

```go
// MCP 工具转换为 Agent 工具
func ToACPToolCall(mcpTool MCPTool) ToolCall {
    return ToolCall{
        ID:        generateID(),
        Name:      fmt.Sprintf("mcp__%s__%s", serverName, toolName),
        Arguments: args,
    }
}
```

### 9.2 权限控制

MCP 工具调用受 qodercli 权限系统控制：

```
core/agent/permission/mcp_rule_matcher/
├── 检查 MCP 工具调用权限
├── 应用 config_rule 规则
└── 用户确认 (如需要)
```

### 9.3 结果处理

```go
func ToToolCallResult(result MCPResult) ToolResult {
    // 将 MCP 结果转换为 Agent 可用格式
}
```

## 10. 内置 MCP 服务器

### 10.1 ACP 服务器

```
acp/mcp_server/
├── 提供 delete_file 等工具
└── 与 ACP 协议集成
```

### 10.2 Quest 适配器

```
quest-mcp-adaptor
├── 提供 search_codebase 工具
└── 代码库搜索功能
```

## 11. 错误处理

### 11.1 常见错误

| 错误 | 说明 |
|------|------|
| `mcp server initialization failed` | MCP 服务器初始化失败 |
| `failed to probe MCP server` | 探测 MCP 服务器失败 |
| `tool not found` | 工具未找到 |
| `prompt not found` | 提示未找到 |
| `method not found` | 方法未找到 |
| `invalid resource URI` | 无效的资源 URI |
| `Invalid JSON-RPC version` | 无效的 JSON-RPC 版本 |

### 11.2 错误码

```
-32700: Parse error
-32600: Invalid Request
-32601: Method not found
-32602: Invalid params
-32603: Internal error
```

## 12. 调试与日志

### 12.1 日志级别

```
logging/setLevel
├── debug
├── info
├── warning
└── error
```

### 12.2 连接状态

```
Mcp-Session-Id     # 会话 ID
Content-Length     # 消息长度
```

## 13. CLI 命令

```bash
# 添加 MCP 服务器
qodercli mcp add

# 列出 MCP 服务器
qodercli mcp list

# 获取服务器详情
qodercli mcp get <server-name>

# 移除服务器
qodercli mcp remove <server-name>

# OAuth 认证
qodercli mcp auth <server-name>
```

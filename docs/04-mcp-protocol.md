# qodercli-reverse MCP Protocol Implementation Analysis

## 1. MCP Architecture Overview

MCP (Model Context Protocol) is the core protocol qodercli-reverse uses to extend AI capabilities, allowing integration of external tools, resources, and prompts.

### 1.1 Package Structure

```
core/resource/mcp/
├── mcp/                    # MCP core implementation
├── mcp_ipc_client_unix/    # Unix IPC client
├── mcps/                   # MCP server list management
├── process/                # MCP process management
└── token_store/            # MCP Token storage

cmd/mcp/
├── add/                    # Add MCP server
├── auth/                   # MCP OAuth authentication
├── get/                    # Get MCP server info
├── list/                   # List MCP servers
└── remove/                 # Remove MCP server
```

## 2. MCP Server Configuration

### 2.1 Configuration File Location

```
~/.mcp.json           # Global MCP configuration
./.mcp.json           # Project-level MCP configuration
```

### 2.2 Configuration Format

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

### 2.3 Server Types

- **stdio**: Communicate via standard input/output
- **sse**: Communicate via Server-Sent Events
- **http**: Communicate via HTTP

## 3. MCP Protocol Methods

### 3.1 Core Methods

| Method | Description |
|--------|-------------|
| `initialize` | Initialize connection |
| `tools/list` | List available tools |
| `tools/call` | Call a tool |
| `resources/list` | List available resources |
| `resources/read` | Read a resource |
| `resources/templates/list` | List resource templates |
| `resources/subscribe` | Subscribe to resource updates |
| `resources/unsubscribe` | Unsubscribe |
| `prompts/list` | List available prompts |
| `prompts/get` | Get a prompt |
| `logging/setLevel` | Set log level |

### 3.2 Notification Types

| Notification | Description |
|--------------|-------------|
| `notifications/resources/list_changed` | Resource list changed |
| `notifications/prompts/list_changed` | Prompt list changed |
| `notifications/tools/list_changed` | Tool list changed |

## 4. Tool Integration

### 4.1 Tool Naming Convention

MCP tools are named in the format `mcp__<server>__<tool>` in qodercli-reverse:

```
mcp__browser-use__*          # All tools of the browser-use server
mcp__quest__search_codebase  # search_codebase tool of the quest server
mcp__acp__delete_file        # delete_file tool of the ACP server
```

### 4.2 Tool Call Flow

```
1. Agent decides to call an MCP tool
2. core/resource/mcp/mcp.go locates the corresponding server
3. Send tools/call request via IPC/HTTP
4. MCP server executes the tool
5. Return result to Agent
```

### 4.3 Tool Definition Format

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

## 5. Resource Management

### 5.1 Resource Types

- **Text resources**: `text/plain`, `text/markdown`, `text/html`
- **Binary resources**: via base64 encoding
- **Resource links**: `resource_link`

### 5.2 Resource URI Format

```
file:///path/to/file          # File resource
qoder:///agent/...            # Qoder internal resource
```

## 6. OAuth Authentication

### 6.1 OAuth Flow

```
1. User runs `qodercli-reverse mcp auth <server>`
2. Discover OAuth configuration (.well-known/oauth-protected-resource)
3. Generate code_verifier and code_challenge
4. Open browser for authorization
5. Callback obtains authorization_code
6. Exchange access_token
7. Store token in token_store
```

### 6.2 Related Fields

```
response_type        # "code"
code_challenge       # PKCE challenge
code_challenge_method # "S256"
redirect_uri         # Callback URL
client_id            # Client ID
client_secret        # Client secret (optional)
```

## 7. Process Management

### 7.1 Process Lifecycle

```
core/resource/mcp/process/
├── Start MCP server process
├── Manage process state
├── Handle process crashes and restarts
└── Clean up process resources
```

### 7.2 IPC Communication (Unix)

```
core/resource/mcp/mcp_ipc_client_unix/
├── Unix socket communication
├── JSON-RPC message encapsulation
└── Asynchronous message handling
```

## 8. JSON-RPC Protocol

### 8.1 Request Format

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

### 8.2 Response Format

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

### 8.3 Error Format

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

## 9. MCP and Agent Integration

### 9.1 Tool Registration

```go
// Convert MCP tool to Agent tool
func ToACPToolCall(mcpTool MCPTool) ToolCall {
    return ToolCall{
        ID:        generateID(),
        Name:      fmt.Sprintf("mcp__%s__%s", serverName, toolName),
        Arguments: args,
    }
}
```

### 9.2 Permission Control

MCP tool calls are governed by the qodercli-reverse permission system:

```
core/agent/permission/mcp_rule_matcher/
├── Check MCP tool call permission
├── Apply config_rule rules
└── User confirmation (if needed)
```

### 9.3 Result Handling

```go
func ToToolCallResult(result MCPResult) ToolResult {
    // Convert MCP result into Agent-usable format
}
```

## 10. Built-in MCP Servers

### 10.1 ACP Server

```
acp/mcp_server/
├── Provides tools such as delete_file
└── Integrates with the ACP protocol
```

### 10.2 Quest Adapter

```
quest-mcp-adaptor
├── Provides search_codebase tool
└── Codebase search functionality
```

## 11. Error Handling

### 11.1 Common Errors

| Error | Description |
|-------|-------------|
| `mcp server initialization failed` | MCP server initialization failed |
| `failed to probe MCP server` | Failed to probe MCP server |
| `tool not found` | Tool not found |
| `prompt not found` | Prompt not found |
| `method not found` | Method not found |
| `invalid resource URI` | Invalid resource URI |
| `Invalid JSON-RPC version` | Invalid JSON-RPC version |

### 11.2 Error Codes

```
-32700: Parse error
-32600: Invalid Request
-32601: Method not found
-32602: Invalid params
-32603: Internal error
```

## 12. Debugging and Logging

### 12.1 Log Levels

```
logging/setLevel
├── debug
├── info
├── warning
└── error
```

### 12.2 Connection State

```
Mcp-Session-Id     # Session ID
Content-Length     # Message length
```

## 13. CLI Commands

```bash
# Add MCP server
qodercli-reverse mcp add

# List MCP servers
qodercli-reverse mcp list

# Get server details
qodercli-reverse mcp get <server-name>

# Remove server
qodercli-reverse mcp remove <server-name>

# OAuth authentication
qodercli-reverse mcp auth <server-name>
```
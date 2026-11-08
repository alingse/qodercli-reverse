# qodercli-reverse Configuration System and API Communication Analysis

## 1. Configuration System Architecture

### 1.1 Package Structure

```
core/config/
├── config/          # Config core
├── qoder/           # Qoder config
└── settings/        # Settings management
```

## 2. Configuration File Hierarchy

### 2.1 Global Configuration

```
~/.qoder/
├── config.json              # Global config
├── settings.json            # Settings
├── settings.local.json      # Local settings override
├── .mcp.json                # MCP server config
├── memory                   # Global memory
├── permission.yaml          # Permission rules
├── hooks.json               # Hook config
└── qoder.json               # Qoder config
```

### 2.2 Project Configuration

```
./.qoder/
├── settings.json            # Project settings
├── settings.local.json      # Local override
├── .mcp.json                # Project MCP config
├── memory                   # Project memory
├── AGENTS.md                # Agent definitions
└── SKILL.md                 # Skill definitions

./.claude/
├── CLAUDE.md                # Claude Code compatible config
├── commands/                # Custom commands
├── agents/                  # Custom agents
└── skills/                  # Custom skills
```

## 3. Configuration Items

### 3.1 Model Configuration

```json
{
  "model": "auto",
  "modelLevel": "performance",
  "maxOutputTokens": "32k"
}
```

### 3.2 Tool Configuration

```json
{
  "allowedTools": ["Bash", "Read", "Write", "Edit"],
  "disallowedTools": ["WebFetch"]
}
```

### 3.3 Permission Configuration

```json
{
  "permissionMode": "ask",
  "dangerouslySkipPermissions": false
}
```

### 3.4 Output Configuration

```json
{
  "outputStyle": "default",
  "theme": "qoder"
}
```

## 4. API Endpoints

### 4.1 Core Endpoints

| Endpoint | Description |
|----------|-------------|
| `https://openapi.qoder.sh` | OpenAPI service |
| `https://center.qoder.sh` | Center service |
| `https://api2.qoder.sh` | API v2 |
| `https://qts2.qoder.sh` | QTS service |
| `https://daily.qoder.ai` | Daily service |

### 4.2 API Paths

| Path | Method | Description |
|------|--------|-------------|
| `/api/v1/userinfo` | GET | User info |
| `/api/v1/deviceToken/poll` | POST | Device token polling |
| `/api/v1/tracking` | POST | Event tracking |
| `/api/v1/heartbeat` | GET | Heartbeat |
| `/api/v1/me/features` | GET | User features |
| `/api/v2/model/list` | GET | Model list |
| `/api/v2/quota/usage` | GET | Quota usage |
| `/api/v2/user/plan` | GET | User plan |
| `/api/v3/user/status` | GET | User status |
| `/api/v3/user/jobToken` | GET | Job token |
| `/algo/api/v1/webSearch/oneSearch` | GET | Web search |

## 5. HTTP Client

### 5.1 Package Structure

```
core/utils/http/
├── client/          # HTTP client
└── init/            # Initialization
```

### 5.2 Request Headers

```
Cosy-ClientIp            # Client IP
Cosy-ClientType          # Client type
Cosy-MachineType         # Machine type
Cosy-Data-Policy         # Data policy
Cosy-User                # User identifier
Cosy-Date                # Date
Cosy-Version             # Version
Cosy-Organization-Id     # Organization ID
Cosy-Organization-Tags   # Organization tags
Cosy-Key                 # Key
X-Model-Key              # Model key
X-Model-Source           # Model source
X-Request-Id             # Request ID
X-Machine-Id             # Machine ID
X-Machine-OS             # Machine OS
X-IDE-Platform           # IDE platform
X-Client-Timestamp       # Client timestamp
```

## 6. SSE Client

```
core/utils/qoder/sse/
├── client/          # SSE client
└── event/           # Event handling
```

Used for streaming API responses.

## 7. Authentication

### 7.1 Authentication Methods

- **Personal Access Token**: `QODER_PERSONAL_ACCESS_TOKEN`
- **OAuth**: Browser login
- **Device Token**: Device polling

### 7.2 Token Storage

```
core/utils/qoder/storage/
└── Local encrypted token storage
```

## 8. Region Configuration

```
core/utils/qoder/region_config/
└── Region config cache
```

## 9. Device Info

```
core/utils/qoder/device_info/
├── Device ID generation
└── Device info collection
```

## 10. Heartbeat and Tracking

### 10.1 Heartbeat

```
/api/v1/heartbeat
├── Sent periodically
└── Keeps session alive
```

### 10.2 Event Tracking

```
core/utils/sls/
├── command_tracker/     # Command tracking
├── data_format/         # Data format
├── heartbeat/           # Heartbeat
├── initializer/         # Initialization
└── reporter/            # Reporting
```

Tracked events:
```
qodercli-reverse_agent_request
qodercli-reverse_tool_decision
qodercli-reverse_tool_result
qodercli-reverse_sdk_user_query_start
qodercli-reverse_sdk_user_query_complete
qodercli-reverse_compact_start
qodercli-reverse_compact_complete
qodercli-reverse_feedback_failed
qodercli-reverse_update
qodercli-reverse_version
```
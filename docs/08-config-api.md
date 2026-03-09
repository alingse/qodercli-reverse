# qodercli 配置系统和 API 通信分析

## 1. 配置系统架构

### 1.1 包结构

```
core/config/
├── config/          # 配置核心
├── qoder/           # Qoder 配置
└── settings/        # 设置管理
```

## 2. 配置文件层次

### 2.1 全局配置

```
~/.qoder/
├── config.json              # 全局配置
├── settings.json            # 设置
├── settings.local.json      # 本地设置覆盖
├── .mcp.json                # MCP 服务器配置
├── memory                   # 全局 Memory
├── permission.yaml          # 权限规则
├── hooks.json               # Hook 配置
└── qoder.json               # Qoder 配置
```

### 2.2 项目配置

```
./.qoder/
├── settings.json            # 项目设置
├── settings.local.json      # 本地覆盖
├── .mcp.json                # 项目 MCP 配置
├── memory                   # 项目 Memory
├── AGENTS.md                # Agent 定义
└── SKILL.md                 # Skill 定义

./.claude/
├── CLAUDE.md                # Claude Code 兼容配置
├── commands/                # 自定义命令
├── agents/                  # 自定义 Agent
└── skills/                  # 自定义 Skill
```

## 3. 配置项

### 3.1 模型配置

```json
{
  "model": "auto",
  "modelLevel": "performance",
  "maxOutputTokens": "32k"
}
```

### 3.2 工具配置

```json
{
  "allowedTools": ["Bash", "Read", "Write", "Edit"],
  "disallowedTools": ["WebFetch"]
}
```

### 3.3 权限配置

```json
{
  "permissionMode": "ask",
  "dangerouslySkipPermissions": false
}
```

### 3.4 输出配置

```json
{
  "outputStyle": "default",
  "theme": "qoder"
}
```

## 4. API 端点

### 4.1 核心端点

| 端点 | 说明 |
|------|------|
| `https://openapi.qoder.sh` | OpenAPI 服务 |
| `https://center.qoder.sh` | 中心服务 |
| `https://api2.qoder.sh` | API v2 |
| `https://qts2.qoder.sh` | QTS 服务 |
| `https://daily.qoder.ai` | 日常服务 |

### 4.2 API 路径

| 路径 | 方法 | 说明 |
|------|------|------|
| `/api/v1/userinfo` | GET | 用户信息 |
| `/api/v1/deviceToken/poll` | POST | 设备 Token 轮询 |
| `/api/v1/tracking` | POST | 事件追踪 |
| `/api/v1/heartbeat` | GET | 心跳 |
| `/api/v1/me/features` | GET | 用户功能 |
| `/api/v2/model/list` | GET | 模型列表 |
| `/api/v2/quota/usage` | GET | 配额使用 |
| `/api/v2/user/plan` | GET | 用户计划 |
| `/api/v3/user/status` | GET | 用户状态 |
| `/api/v3/user/jobToken` | GET | Job Token |
| `/algo/api/v1/webSearch/oneSearch` | GET | 网页搜索 |

## 5. HTTP 客户端

### 5.1 包结构

```
core/utils/http/
├── client/          # HTTP 客户端
└── init/            # 初始化
```

### 5.2 请求头

```
Cosy-ClientIp            # 客户端 IP
Cosy-ClientType          # 客户端类型
Cosy-MachineType         # 机器类型
Cosy-Data-Policy         # 数据策略
Cosy-User                # 用户标识
Cosy-Date                # 日期
Cosy-Version             # 版本
Cosy-Organization-Id     # 组织 ID
Cosy-Organization-Tags   # 组织标签
Cosy-Key                 # 密钥
X-Model-Key              # 模型密钥
X-Model-Source           # 模型来源
X-Request-Id             # 请求 ID
X-Machine-Id             # 机器 ID
X-Machine-OS             # 机器操作系统
X-IDE-Platform           # IDE 平台
X-Client-Timestamp       # 客户端时间戳
```

## 6. SSE 客户端

```
core/utils/qoder/sse/
├── client/          # SSE 客户端
└── event/           # 事件处理
```

用于流式 API 响应。

## 7. 认证

### 7.1 认证方式

- **Personal Access Token**: `QODER_PERSONAL_ACCESS_TOKEN`
- **OAuth**: 浏览器登录
- **Device Token**: 设备轮询

### 7.2 Token 存储

```
core/utils/qoder/storage/
└── 本地加密存储 Token
```

## 8. 区域配置

```
core/utils/qoder/region_config/
└── 区域配置缓存
```

## 9. 设备信息

```
core/utils/qoder/device_info/
├── 设备 ID 生成
└── 设备信息收集
```

## 10. 心跳和追踪

### 10.1 心跳

```
/api/v1/heartbeat
├── 定期发送
└── 保持会话活跃
```

### 10.2 事件追踪

```
core/utils/sls/
├── command_tracker/     # 命令追踪
├── data_format/         # 数据格式
├── heartbeat/           # 心跳
├── initializer/         # 初始化
└── reporter/            # 上报
```

追踪事件:
```
qodercli_agent_request
qodercli_tool_decision
qodercli_tool_result
qodercli_sdk_user_query_start
qodercli_sdk_user_query_complete
qodercli_compact_start
qodercli_compact_complete
qodercli_feedback_failed
qodercli_update
qodercli_version
```

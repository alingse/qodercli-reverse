# qodercli 工具系统 (Tools) 分析

## 1. 工具系统架构

### 1.1 包结构

```
core/agent/tools/
├── abstract/              # 工具抽象基类
├── ask_user_question/     # AskUserQuestion 工具
├── bash/                  # Bash 工具
├── bashkill/              # KillBash 工具
├── bashoutput/            # BashOutput 工具
├── check_runtime/         # CheckRuntime 工具
├── edit/                  # Edit 工具
├── enter_plan_in_quest/   # Quest 进入 Plan 模式
├── exit_plan_in_quest/    # Quest 退出 Plan 模式
├── glob/                  # Glob 工具
├── grep/                  # Grep 工具
├── imagegen/              # ImageGen 工具
├── ls/                    # LS 工具
├── multiedit/             # MultiEdit 工具
├── read/                  # Read 工具
├── skill/                 # Skill 工具
├── task/                  # Task (SubAgent) 工具
├── todowrite/             # TodoWrite 工具
├── utils/                 # 工具通用工具
│   ├── file/              # 文件工具
│   ├── format/            # 格式化工具
│   └── repair/            # 修复工具
├── webfetch/              # WebFetch 工具
├── webfetch_trusted/      # WebFetch (受信任) 工具
└── write/                 # Write 工具
```

## 2. 内置工具列表

### 2.1 核心工具

| 工具名 | 类型 | 说明 |
|--------|------|------|
| `Bash` | 执行 | 执行 shell 命令 |
| `Read` | 读取 | 读取文件内容 |
| `Write` | 写入 | 创建或覆盖文件 |
| `Edit` | 编辑 | 查找替换文件内容 |
| `MultiEdit` | 编辑 | 批量编辑文件 |
| `Glob` | 搜索 | 文件模式匹配搜索 |
| `Grep` | 搜索 | 基于正则的内容搜索 |
| `LS` | 浏览 | 列出目录内容 |

### 2.2 进程管理工具

| 工具名 | 说明 |
|--------|------|
| `KillBash` | 终止后台 Bash 进程 |
| `BashOutput` | 获取后台 Bash 进程输出 |

### 2.3 任务管理工具

| 工具名 | 说明 |
|--------|------|
| `Task` | 启动子代理 (SubAgent) |
| `TodoWrite` | 更新任务列表 |
| `Skill` | 执行预定义技能 |

### 2.4 交互工具

| 工具名 | 说明 |
|--------|------|
| `AskUserQuestion` | 向用户提问 |

### 2.5 网络工具

| 工具名 | 说明 |
|--------|------|
| `WebFetch` | 获取网页内容 |
| `WebSearch` | 搜索网络 |

### 2.6 其他工具

| 工具名 | 说明 |
|--------|------|
| `ImageGen` | 生成图片 |
| `CheckRuntime` | 检查运行时环境 |
| `EnterSpecMode` / `EnterPlanMode` | 进入规划模式 |
| `ExitSpecMode` / `ExitPlanMode` | 退出规划模式 |
| `DeleteFile` | 删除文件 |
| `SearchCodebase` | 搜索代码库 |
| `SearchSymbol` | 搜索符号 |

## 3. 工具定义格式

### 3.1 工具接口

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, input string) (string, error)
}
```

### 3.2 输入 Schema 示例

```json
{
  "type": "object",
  "properties": {
    "file_path": {
      "type": "string",
      "description": "The absolute path to the file"
    },
    "offset": {
      "type": "integer",
      "description": "Line number to start reading from"
    },
    "limit": {
      "type": "integer",
      "description": "Number of lines to read"
    }
  },
  "required": ["file_path"]
}
```

## 4. 核心工具详解

### 4.1 Bash 工具

```go
// 参数结构
type BashParams struct {
    Command       string   // 要执行的命令
    Timeout       int      // 超时时间 (毫秒)
    RunInBackground bool   // 是否后台运行
    Description   string   // 命令描述
}
```

特性:
- 支持前台和后台执行
- 支持超时控制 (`QODER_BASH_TIMEOUT`)
- 自动处理 Shell 快照和恢复
- 支持 Bash 模式 (用户输入)

### 4.2 Read 工具

```go
type ReadParams struct {
    FilePath    string  // 文件路径
    Offset      int     // 起始行号
    Limit       int     // 读取行数
    ReadImage   bool    // 是否读取图片
}
```

特性:
- 支持文本文件和图片文件
- 自动检测文件类型
- 支持分页读取大文件
- 行号前缀格式输出

### 4.3 Edit 工具

```go
type EditParams struct {
    FilePath    string  // 文件路径
    OldString   string  // 要替换的内容
    NewString   string  // 替换后的内容
    ReplaceAll  bool    // 是否替换所有匹配
}
```

特性:
- 必须先使用 Read 工具读取文件
- 精确匹配 (包括缩进)
- 支持批量替换
- 自动备份

### 4.4 Write 工具

```go
type WriteParams struct {
    FilePath    string  // 文件路径
    Content     string  // 文件内容
}
```

特性:
- 原子写入
- 自动创建目录
- 文件权限处理

### 4.5 Grep 工具

```go
type GrepParams struct {
    Pattern     string   // 正则模式
    Path        string   // 搜索路径
    OutputMode  string   // 输出模式: content/files_with_matches/count
    Type        string   // 文件类型
    Glob        string   // 文件模式
    HeadLimit   int      // 结果限制
    Multiline   bool     // 多行模式
}
```

特性:
- 基于 ripgrep 实现
- 支持多种输出模式
- 支持文件类型过滤
- 支持上下文行 (-A, -B, -C)

### 4.6 Glob 工具

```go
type GlobParams struct {
    Pattern     string  // glob 模式
    Path        string  // 搜索路径
}
```

特性:
- 递归搜索
- 支持通配符
- 按修改时间排序

### 4.7 Task 工具

```go
type TaskParams struct {
    SubagentType string  // 子代理类型
    Description  string  // 任务描述
    Prompt       string  // 任务提示
}
```

可用子代理类型:
- `code-reviewer`: 代码审查
- `design-agent`: 设计文档
- `spec-review-agent`: 规格审查
- `task-executor`: 任务执行
- `general-purpose`: 通用代理

### 4.8 TodoWrite 工具

```go
type TodoParams struct {
    Todos []TodoItem  // 任务列表
}

type TodoItem struct {
    Content     string  // 任务内容
    Status      string  // pending/in_progress/completed
    ActiveForm  string  // 进行中形式
}
```

### 4.9 AskUserQuestion 工具

```go
type AskUserQuestionParams struct {
    Questions []Question  // 问题列表
}

type Question struct {
    Question    string    // 问题文本
    Header      string    // 简短标签
    Options     []Option  // 选项
    MultiSelect bool      // 是否多选
}
```

## 5. 工具执行流程

```
1. Agent 生成工具调用 (ToolCall)
2. 权限检查 (PermissionChecker)
   ├── 检查是否在 allowed_tools 列表
   ├── 检查是否在 disallowed_tools 列表
   └── 检查路径/命令权限规则
3. 如需确认，显示权限对话框
4. 执行工具
   ├── PreToolUse Hook
   ├── 工具执行
   └── PostToolUse Hook
5. 返回结果给 Agent
```

## 6. 权限控制

### 6.1 工具级别权限

```
--allowed-tools "Bash,Edit"      # 只允许指定工具
--disallowed-tools "WebFetch"    # 禁用指定工具
--dangerously-skip-permissions   # 跳过所有权限检查
```

### 6.2 路径级别权限

```
core/agent/permission/file_rule_matcher/
├── 检查文件路径是否允许访问
├── 应用项目规则
└── 应用用户规则
```

### 6.3 Bash 命令权限

```
core/agent/permission/bash_rule_matcher/
├── 检查命令是否允许执行
├── 应用 gitignore 规则
└── 应用危险命令列表
```

## 7. Hook 系统

### 7.1 Hook 类型

| Hook | 触发时机 |
|------|----------|
| `PreToolUse` | 工具执行前 |
| `PostToolUse` | 工具执行后 |
| `AgentStop` | Agent 停止时 |
| `SessionStart` | 会话开始时 |
| `UserPromptSubmit` | 用户提交时 |

### 7.2 Hook 用途

- 结果截断 (`result_truncator`)
- 结果持久化 (`result_persister`)
- 代码审查检查 (`code_reviewer_checker`)
- 外部通知 (`external`)

## 8. 工具结果格式

### 8.1 文本结果

```
<result>
文件内容或命令输出
</result>
```

### 8.2 错误结果

```json
{
  "error": true,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

### 8.3 截断结果

```
... [N lines truncated]
```

## 9. 特殊工具行为

### 9.1 Bash 工具特殊处理

- `bash-output-special`: 特殊输出处理
- `shell-snapshot`: Shell 状态快照
- `background_shell`: 后台 Shell 管理

### 9.2 文件工具特殊处理

- `special_path_policy`: 特殊路径策略
- `special_file_protection`: 特殊文件保护
- `qoder_config_protection`: Qoder 配置保护

### 9.3 WebFetch 工具特殊处理

- `webfetch_rule_matcher`: URL 规则匹配
- 受信任站点列表
- 重定向限制

## 10. 工具配置

### 10.1 环境变量

```
QODER_BASH_PATH      # Bash 路径
QODER_BASH_TIMEOUT   # Bash 超时
QODER_PROJECT_DIR    # 项目目录
```

### 10.2 配置文件

```json
{
  "allowedTools": ["Bash", "Read", "Write"],
  "disallowedTools": ["WebFetch"],
  "permissionMode": "ask"
}
```

# qodercli SubAgent 和 Skill 系统分析

## 1. SubAgent 系统

### 1.1 包结构

```
core/resource/subagent/
├── builtin/         # 内置 SubAgent
└── subagent/        # SubAgent 核心

core/agent/task/
└── Task 工具 (启动 SubAgent)

core/generator/subagent/
└── SubAgent 生成器
```

### 1.2 内置 SubAgent 类型

| 类型 | 说明 |
|------|------|
| `code-reviewer` | 代码审查 |
| `design-agent` | 设计文档生成 |
| `spec-review-agent` | 规格审查 |
| `task-executor` | 任务执行 |
| `general-purpose` | 通用代理 |
| `spec-hld-designer` | 高层设计 |
| `spec-lld-designer` | 低层设计 |
| `spec-implementer` | 实现者 |
| `spec-leader` | 规格领导 |
| `ExploreAgent` | 代码探索 |

### 1.3 SubAgent 配置

```json
{
  "name": "reviewer",
  "description": "Reviews code",
  "prompt": "You are a code reviewer...",
  "tools": ["Read", "Grep", "Glob"]
}
```

### 1.4 SubAgent 执行流程

```
1. Agent 调用 Task 工具
2. 指定 subagent_type
3. 创建子进程/协程
4. 加载 SubAgent 配置
5. 执行 SubAgent 任务
6. 返回结果给父 Agent
```

## 2. Skill 系统

### 2.1 包结构

```
core/resource/skill/
├── builtin/         # 内置 Skill
└── skills/          # Skill 列表
```

### 2.2 Skill 定义

```
SKILL.md 或 skill.md
```

格式:
```markdown
---
name: skill-name
description: Skill description
---

# Skill Instructions

Detailed instructions for the skill...
```

### 2.3 内置 Skill

```
create-agent      # 创建新 Agent
create-skill      # 创建新 Skill
create-subagent   # 创建新 SubAgent
pdf               # PDF 处理
xlsx              # Excel 处理
```

### 2.4 Skill 调用

```
Skill 工具
├── skill_name: "skill-name"
└── arguments: {...}
```

## 3. Agent 定义

### 3.1 配置位置

```
~/.qoder/agents/
./.qoder/agents/
./.claude/agents/
```

### 3.2 Agent 配置格式

```json
{
  "name": "custom-agent",
  "description": "Custom agent description",
  "prompt": "You are a custom agent...",
  "tools": ["Bash", "Read", "Write", "Edit"],
  "disallowedTools": ["WebFetch"]
}
```

### 3.3 通过 CLI 定义

```bash
qodercli --agents '{"reviewer": {"description": "Reviews code", "prompt": "You are a code reviewer"}}'
```

## 4. Task 工具

### 4.1 参数

```json
{
  "subagent_type": "code-reviewer",
  "description": "Review the changes",
  "prompt": "Review the code in src/main.go"
}
```

### 4.2 并行执行

```
If the user specifies that they want you to run agents "in parallel", 
you MUST send a single message with multiple Task tool use content blocks.
```

## 5. Plugin 系统

### 5.1 包结构

```
core/resource/plugin/
├── bootstrap/       # 插件引导
├── config/          # 插件配置
├── errors/          # 插件错误
├── integration/     # 插件集成
├── loader/          # 插件加载
├── manifest/        # 插件清单
├── registry/        # 插件注册
└── resolver/        # 插件解析
```

### 5.2 插件配置

```bash
--plugin-dir <path>
```

### 5.3 插件清单

```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "description": "Plugin description",
  "main": "main.js",
  "hooks": [...],
  "tools": [...]
}
```

## 6. 自定义命令

### 6.1 包结构

```
core/resource/command/
├── builtin/         # 内置命令
└── commands/        # 命令列表
```

### 6.2 命令定义

位置: `.claude/commands/` 或 `.qoder/commands/`

格式:
```markdown
---
name: command-name
description: Command description
---

# Command Instructions
```

## 7. AGENTS.md

### 7.1 格式

```markdown
# Project Agents

## reviewer
Description: Reviews code
Prompt: |
  You are a code reviewer...
```

### 7.2 位置

```
./AGENTS.md
./.qoder/AGENTS.md
global/AGENTS.md
project/AGENTS.md
```

## 8. 配置加载顺序

```
1. 内置定义
2. 全局配置 (~/.qoder/)
3. 项目配置 (./.qoder/)
4. Claude 兼容配置 (./.claude/)
5. 命令行参数
```

## 9. 追踪事件

```
qodercli_subagent_execution
qodercli_skill_tool_execution
```

# qodercli-reverse SubAgent and Skill System Analysis

## 1. SubAgent System

### 1.1 Package Structure

```
core/resource/subagent/
├── builtin/         # Built-in SubAgents
└── subagent/        # SubAgent core

core/agent/task/
└── Task tool (launches SubAgent)

core/generator/subagent/
└── SubAgent generator
```

### 1.2 Built-in SubAgent Types

| Type | Description |
|------|-------------|
| `code-reviewer` | Code review |
| `design-agent` | Design document generation |
| `spec-review-agent` | Spec review |
| `task-executor` | Task execution |
| `general-purpose` | General-purpose agent |
| `spec-hld-designer` | High-level design |
| `spec-lld-designer` | Low-level design |
| `spec-implementer` | Implementer |
| `spec-leader` | Spec leader |
| `ExploreAgent` | Code exploration |

### 1.3 SubAgent Configuration

```json
{
  "name": "reviewer",
  "description": "Reviews code",
  "prompt": "You are a code reviewer...",
  "tools": ["Read", "Grep", "Glob"]
}
```

### 1.4 SubAgent Execution Flow

```
1. Agent calls Task tool
2. Specify subagent_type
3. Create subprocess/goroutine
4. Load SubAgent config
5. Execute SubAgent task
6. Return result to parent Agent
```

## 2. Skill System

### 2.1 Package Structure

```
core/resource/skill/
├── builtin/         # Built-in skills
└── skills/          # Skill list
```

### 2.2 Skill Definition

```
SKILL.md or skill.md
```

Format:
```markdown
---
name: skill-name
description: Skill description
---

# Skill Instructions

Detailed instructions for the skill...
```

### 2.3 Built-in Skills

```
create-agent      # Create new Agent
create-skill      # Create new Skill
create-subagent   # Create new SubAgent
pdf               # PDF processing
xlsx              # Excel processing
```

### 2.4 Skill Invocation

```
Skill tool
├── skill_name: "skill-name"
└── arguments: {...}
```

## 3. Agent Definitions

### 3.1 Config Locations

```
~/.qoder/agents/
./.qoder/agents/
./.claude/agents/
```

### 3.2 Agent Config Format

```json
{
  "name": "custom-agent",
  "description": "Custom agent description",
  "prompt": "You are a custom agent...",
  "tools": ["Bash", "Read", "Write", "Edit"],
  "disallowedTools": ["WebFetch"]
}
```

### 3.3 Defining via CLI

```bash
qodercli-reverse --agents '{"reviewer": {"description": "Reviews code", "prompt": "You are a code reviewer"}}'
```

## 4. Task Tool

### 4.1 Parameters

```json
{
  "subagent_type": "code-reviewer",
  "description": "Review the changes",
  "prompt": "Review the code in src/main.go"
}
```

### 4.2 Parallel Execution

```
If the user specifies that they want you to run agents "in parallel", 
you MUST send a single message with multiple Task tool use content blocks.
```

## 5. Plugin System

### 5.1 Package Structure

```
core/resource/plugin/
├── bootstrap/       # Plugin bootstrap
├── config/          # Plugin config
├── errors/          # Plugin errors
├── integration/     # Plugin integration
├── loader/          # Plugin loader
├── manifest/        # Plugin manifest
├── registry/        # Plugin registry
└── resolver/        # Plugin resolver
```

### 5.2 Plugin Configuration

```bash
--plugin-dir <path>
```

### 5.3 Plugin Manifest

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

## 6. Custom Commands

### 6.1 Package Structure

```
core/resource/command/
├── builtin/         # Built-in commands
└── commands/        # Command list
```

### 6.2 Command Definition

Location: `.claude/commands/` or `.qoder/commands/`

Format:
```markdown
---
name: command-name
description: Command description
---

# Command Instructions
```

## 7. AGENTS.md

### 7.1 Format

```markdown
# Project Agents

## reviewer
Description: Reviews code
Prompt: |
  You are a code reviewer...
```

### 7.2 Locations

```
./AGENTS.md
./.qoder/AGENTS.md
global/AGENTS.md
project/AGENTS.md
```

## 8. Config Load Order

```
1. Built-in definitions
2. Global config (~/.qoder/)
3. Project config (./.qoder/)
4. Claude compatible config (./.claude/)
5. Command-line arguments
```

## 9. Tracking Events

```
qodercli-reverse_subagent_execution
qodercli-reverse_skill_tool_execution
```
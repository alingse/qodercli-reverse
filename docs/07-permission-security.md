# qodercli-reverse Permission and Security Mechanism Analysis

## 1. Permission System Architecture

### 1.1 Package Structure

```
core/agent/permission/
├── bash_permission_gitignore/  # Bash gitignore permission
├── bash_rule_matcher/          # Bash rule matcher
├── file_rule_matcher/          # File rule matcher
├── mcp_rule_matcher/           # MCP rule matcher
├── path_checker/               # Path checker
├── permission_checker/         # Permission checker
├── permission_coordinator/     # Permission coordinator
├── permission_mode/            # Permission mode
└── webfetch_rule_matcher/      # WebFetch rule matcher
```

## 2. Permission Modes

| Mode | Description |
|------|-------------|
| `ask` | Prompt user for every operation |
| `auto-approve` | Auto-approve all operations |
| `auto-deny` | Auto-deny all operations |

## 3. Permission Rule Types

### 3.1 File Rules

```yaml
# .qoder/permission.yaml
rules:
  - pattern: "*.env"
    action: deny
  - pattern: "src/**"
    action: allow
  - pattern: ".claude/**"
    action: allow
```

### 3.2 Bash Command Rules

```
bash_deny_list          # Denied command list
default_danger_list     # Default dangerous commands
```

Dangerous command examples:
- `rm -rf /`
- `sudo`
- `chmod 777`
- `curl | bash` (some scenarios)

### 3.3 MCP Tool Rules

```
config_rule             # Config rule
mcp_rule_matcher        # MCP rule matcher
```

### 3.4 WebFetch Rules

```
not_in_allow_list       # Not in allow list
not_in_deny_list        # Not in deny list
default_allow           # Default allow
```

## 4. Permission Check Flow

```
1. Tool call request
2. PermissionCoordinator coordinates the check
   ├── Check allowed_tools / disallowed_tools
   ├── Check path/command rules
   └── Check special protections
3. If user confirmation is needed, show permission dialog
4. Record permission decision
5. Execute or reject the operation
```

## 5. Special Protections

### 5.1 Special File Protection

```
special_file_protection     # Special file protection
qoder_config_protection     # Qoder config protection
```

Protected files/directories:
- `.qoder/`
- `.mcp.json`
- `settings.json`
- `settings.local.json`

### 5.2 Special Path Policy

```
special_path_policy         # Special path policy
```

## 6. Permission Decision Records

```
Permission decision         # Permission decision
decision_reason             # Decision reason
reason_type                 # Reason type
reason_detail               # Reason detail
user_reject                 # User rejection
```

## 7. Permission Suggestions

```
permission_suggestions      # Permission suggestions
```

The system provides permission suggestions based on context.

## 8. Hook System

### 8.1 Hook Types

| Hook | Trigger |
|------|---------|
| `PreToolUse` | Before tool execution |
| `PostToolUse` | After tool execution |
| `AgentStop` | When Agent stops |
| `SessionStart` | When session starts |
| `UserPromptSubmit` | When user submits |

### 8.2 Hook Configuration

```json
// .qoder/hooks.json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": ["command"]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit",
        "hooks": ["formatter"]
      }
    ]
  }
}
```

### 8.3 Hook Executor

```
core/resource/hook/
├── executor/        # Hook executor
├── hook/            # Hook core
├── hooks/           # Hook list
└── log/             # Hook log
```

## 9. Skipping Permission Checks

```bash
--dangerously-skip-permissions
--yolo  # alias
```

**Warning**: Use only in trusted environments.

## 10. Security Best Practices

1. Do not use `--yolo` in production
2. Regularly check `.qoder/permission.yaml`
3. Use `--disallowed-tools` to restrict dangerous tools
4. Configure appropriate Hooks for auditing
5. Protect sensitive files (`.env`, credentials)
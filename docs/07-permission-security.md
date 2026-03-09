# qodercli 权限和安全机制分析

## 1. 权限系统架构

### 1.1 包结构

```
core/agent/permission/
├── bash_permission_gitignore/  # Bash gitignore 权限
├── bash_rule_matcher/          # Bash 规则匹配
├── file_rule_matcher/          # 文件规则匹配
├── mcp_rule_matcher/           # MCP 规则匹配
├── path_checker/               # 路径检查
├── permission_checker/         # 权限检查器
├── permission_coordinator/     # 权限协调器
├── permission_mode/            # 权限模式
└── webfetch_rule_matcher/      # WebFetch 规则匹配
```

## 2. 权限模式

| 模式 | 说明 |
|------|------|
| `ask` | 每次操作都询问用户 |
| `auto-approve` | 自动批准所有操作 |
| `auto-deny` | 自动拒绝所有操作 |

## 3. 权限规则类型

### 3.1 文件规则

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

### 3.2 Bash 命令规则

```
bash_deny_list          # 禁止命令列表
default_danger_list     # 默认危险命令
```

危险命令示例:
- `rm -rf /`
- `sudo`
- `chmod 777`
- `curl | bash` (部分场景)

### 3.3 MCP 工具规则

```
config_rule             # 配置规则
mcp_rule_matcher        # MCP 规则匹配
```

### 3.4 WebFetch 规则

```
not_in_allow_list       # 不在允许列表
not_in_deny_list        # 不在拒绝列表
default_allow           # 默认允许
```

## 4. 权限检查流程

```
1. 工具调用请求
2. PermissionCoordinator 协调检查
   ├── 检查 allowed_tools / disallowed_tools
   ├── 检查路径/命令规则
   └── 检查特殊保护
3. 如需用户确认，显示权限对话框
4. 记录权限决策
5. 执行或拒绝操作
```

## 5. 特殊保护

### 5.1 特殊文件保护

```
special_file_protection     # 特殊文件保护
qoder_config_protection     # Qoder 配置保护
```

保护的文件/目录:
- `.qoder/`
- `.mcp.json`
- `settings.json`
- `settings.local.json`

### 5.2 特殊路径策略

```
special_path_policy         # 特殊路径策略
```

## 6. 权限决策记录

```
Permission decision         # 权限决策
decision_reason             # 决策原因
reason_type                 # 原因类型
reason_detail               # 原因详情
user_reject                 # 用户拒绝
```

## 7. 权限建议

```
permission_suggestions      # 权限建议
```

系统会根据上下文提供权限建议。

## 8. Hook 系统

### 8.1 Hook 类型

| Hook | 触发时机 |
|------|----------|
| `PreToolUse` | 工具执行前 |
| `PostToolUse` | 工具执行后 |
| `AgentStop` | Agent 停止时 |
| `SessionStart` | 会话开始时 |
| `UserPromptSubmit` | 用户提交时 |

### 8.2 Hook 配置

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

### 8.3 Hook 执行器

```
core/resource/hook/
├── executor/        # Hook 执行器
├── hook/            # Hook 核心
├── hooks/           # Hook 列表
└── log/             # Hook 日志
```

## 9. 跳过权限检查

```bash
--dangerously-skip-permissions
--yolo  # 别名
```

**警告**: 仅在受信任环境使用。

## 10. 安全最佳实践

1. 不要在生产环境使用 `--yolo`
2. 定期检查 `.qoder/permission.yaml`
3. 使用 `--disallowed-tools` 限制危险工具
4. 配置适当的 Hook 进行审计
5. 保护敏感文件 (`.env`, credentials)

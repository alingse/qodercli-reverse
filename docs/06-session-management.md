# qodercli 会话管理和持久化分析

## 1. 会话架构

### 1.1 包结构

```
core/agent/state/session/
├── cache/          # 会话缓存
├── options/        # 会话选项
├── session/        # 会话核心
└── storage/        # 会话存储
```

## 2. 会话生命周期

```
session/new      # 创建新会话
session/load     # 加载会话
session/update   # 更新会话
session/cancel   # 取消会话
session/prompt   # 会话提示
session/set_mode # 设置会话模式
```

## 3. 会话存储

### 3.1 存储位置

```
~/.qoder/
├── sessions/
│   └── *-session.json    # 会话文件
├── conversations/
│   └── conversation-*.txt # 导出的对话
└── projects/
    └── {project}/
        └── .qoder/       # 项目级配置
```

### 3.2 会话文件格式

```json
{
  "session_id": "uuid",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "messages": [...],
  "state": {...},
  "metadata": {
    "model": "claude-sonnet-4",
    "workspace": "/path/to/project"
  }
}
```

## 4. 消息管理

```
core/agent/state/message/
├── cache/      # 消息缓存
├── message/    # 消息核心
└── storage/    # 消息存储
```

## 5. 上下文压缩 (Compact)

### 5.1 触发条件

- 上下文接近模型限制
- 手动触发 (`/compact`)
- 自动触发 (auto compact)

### 5.2 压缩流程

```
1. 分析消息历史
2. 识别重要信息
3. 生成摘要
4. 替换历史消息
5. 保存压缩元数据
```

### 5.3 相关指标

```
compact_input_tokens     # 压缩前 Token
compact_output_tokens    # 压缩后 Token
compressionOutputTokens  # 压缩输出
compact_cache_read       # 缓存读取
compact_cache_creation   # 缓存创建
compaction_triggered     # 触发压缩
compact_manual           # 手动压缩
```

## 6. 会话恢复

### 6.1 恢复方式

```bash
qodercli -c                    # 继续最近会话
qodercli -r <session-id>       # 恢复指定会话
/resume                        # TUI 内恢复
```

### 6.2 恢复流程

```
1. 加载会话文件
2. 恢复消息历史
3. 恢复 Shell 状态 (如有)
4. 恢复文件状态
5. 继续对话
```

## 7. 会话导出

```bash
/export [filename]   # 导出当前会话为文本
```

导出格式:
```
conversation-{timestamp}.txt
```

## 8. 状态管理

### 8.1 状态类型

```
core/agent/state/
├── active/           # 活跃状态
├── context/          # 上下文状态
├── file/             # 文件状态
│   ├── file_state/   # 文件状态管理
│   ├── snapshot/     # 文件快照
│   └── storage/      # 文件存储
├── memory/           # Memory 管理
├── session/          # 会话状态
├── shell/            # Shell 状态
├── state/            # 状态核心
├── state_execution/  # 执行状态
├── state_quest/      # Quest 状态
├── state_todo/       # Todo 状态
└── state_vars/       # 状态变量
```

### 8.2 文件快照

```
file/snapshot/
├── 记录文件修改历史
├── 支持回滚
└── 用于 diff 显示
```

## 9. Memory 系统

```
core/agent/state/memory/
├── memory/          # Memory 入口
├── memory_state/    # Memory 状态
└── qoder_rule/      # Qoder 规则
```

Memory 文件位置:
```
~/.qoder/memory          # 全局 Memory
./.qoder/memory          # 项目 Memory
```

## 10. 并发会话 (Jobs)

### 10.1 Job 类型

- **Local Job**: 本地进程
- **Worktree Job**: Git worktree 隔离
- **Container Job**: Docker 容器
- **Kubernetes Job**: K8s Pod

### 10.2 Job 管理

```bash
qodercli jobs           # 列出 Jobs
qodercli jobs rm <id>   # 删除 Job
qodercli --worktree     # 创建 Worktree Job
```

### 10.3 Job 存储

```
~/.qoder/
└── worktrees/
    └── {job-id}/
```

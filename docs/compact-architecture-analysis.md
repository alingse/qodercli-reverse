# 原版 qodercli Compact 实现深度分析

> **分析方法**: 通过 `which qodercli` 定位二进制文件，使用 `strings` 提取符号表，结合 grep 分析编译后的函数签名和包结构。

## 执行摘要

本文档是通过对 qodercli 二进制文件 (`/Users/zhihu/.local/bin/qodercli`) 进行逆向工程分析得出的结论。使用 `strings` 工具提取了 157,441 行字符串，从中筛选出 364 行与 compact 相关的内容，并分析了完整的 Go 函数符号表。

**关键发现**: 原版 compact 是一个约 4000+ 行的复杂系统，而当前反编译代码只有约 50 行的简化实现。

## 目录

- [1. 核心包结构](#1-核心包结构从符号路径推断)
- [2. CompactTrigger 类型](#2-compacttrigger-类型-typescompacttrigger)
- [3. 多种压缩策略](#3-多种压缩策略)
- [4. Token 管理机制](#4-token-管理机制)
- [5. Session Memory 详细架构](#5-session-memory-详细架构)
- [6. Pre-Compact Hook 系统](#6-pre-compact-hook-系统)
- [7. 配置系统](#7-配置系统)
- [8. TUI 集成](#8-tui-集成)
- [9. ACP 协议集成](#9-acp-协议集成)
- [10. 当前实现对比](#10-当前实现对比)

---

## 基于二进制符号的完整架构分析

### 1. 核心包结构 (从符号路径推断)

```
core/agent/
├── compact.go                     # Compact 核心实现
├── session_memory.go              # Session Memory 集成
├── hooks/pre_compact/              # Pre-Compact Hook 系统
│   ├── external.go                # 外部 Hook 实现
│   └── pre_compact.go             # PreCompactHook 定义
├── state/
│   ├── agent_state.go             # Agent 状态管理
│   ├── model_context.go           # 模型上下文管理
│   └── session_memory/             # Session Memory 独立包
│       ├── state.go               # 状态管理
│       ├── queue.go               # 串行队列
│       └── limits.go              # 限制配置
└── prompt/
    └── session_memory.go          # Session Memory 提示词
```

### 2. CompactTrigger 类型 (types.CompactTrigger)

从符号推断有三种触发方式:
- `IsAuto()` - 自动触发 (上下文超阈值时)
- `IsManual()` - 手动触发 (/compact 命令)
- `IsReload()` - 重载触发 (会话恢复时)

### 3. 多种压缩策略

#### 3.1 MicroCompact (微型压缩)
- 快速、轻量级压缩
- 有独立的阈值判断 (`shouldRunMicroCompact`)
- 适用于小幅超限场景

#### 3.2 FullCompact (完整压缩)
- 完整的压缩流程
- 可能调用 LLM 生成摘要

#### 3.3 SessionMemoryCompact (会话记忆压缩)
- 使用 Session Memory 机制
- 持久化存储压缩结果
- 支持断点续传

#### 3.4 SummarizeCompact (摘要式压缩)
- 调用 LLM 生成语义摘要
- 有配置选项 (`shouldSummarizeCompact`)

#### 3.5 TraditionalSummarizeCompact (传统摘要)
- 旧版摘要方法 (可能是 fallback)
- 有多个配置器:
  - WithMeta
  - WithMetadata
  - WithModelConfig
  - WithPostProcessor
  - WithSystemPrompt
  - WithTools

### 4. Token 管理机制

从字符串证据:
- `getCompactThreshold` - 获取压缩阈值
- `WarningThreshold %dK` - 警告阈值 (以 K 为单位)
- `Skip micro compact: total history tokens %d less than WarningThreshold %dK`
- `compact_input_tokens` - 输入 token 计数
- `compact_output_tokens` - 输出 token 计数
- `savedTokens` - 节省的 token 数
- `compressed_count` - 压缩的消息数
- `total_candidates` - 候选消息总数
- `kept_messages` - 保留的消息数

### 5. Session Memory 详细架构

#### 5.1 核心类型
- `session_memory.State` - 状态接口
- `session_memory.SMCompactConfig` - 压缩配置
- `session_memory.SMUpdateConfig` - 更新配置
- `session_memory.SerialQueue` - 串行队列 (保证顺序)
- `session_memory.SectionTokenInfo` - 分块 Token 信息

#### 5.2 核心方法
- `GetCompactConfig()` - 获取压缩配置
- `GetCurrentMemory()` - 获取当前记忆
- `GetLastSummarizedMsgId()` - 获取最后摘要的消息 ID
- `EnsureTemplate()` - 确保模板存在
- `TriggerUpdate()` - 触发更新
- `WaitForUpdate()` - 等待更新完成
- `MarkUpdateStart/Complete()` - 标记更新状态
- `ResetState()` - 重置状态
- `countToolCallsSince()` - 统计工具调用次数
- `hasPendingContent()` - 检查待处理内容
- `hasPendingToolUse()` - 检查待处理工具调用

#### 5.3 截断机制
- `TruncateContent` - 截断内容
- `truncateSections` - 截断分块
- `GenerateOverLimitWarning` - 生成超限警告
- `GetOverLimitInfo` - 获取超限信息
- `ParseSectionTokens` - 解析分块 Token

#### 5.4 持久化
- `GetMemoryPath()` - 获取记忆文件路径
- `getMemoryDir()` - 获取记忆目录
- `IsTemplateEmpty()` - 检查模板是否为空

### 6. Pre-Compact Hook 系统

#### 6.1 Hook 类型
- `pre_compact.PreCompactHook` - 基础 Hook 接口
- `pre_compact.PreCompactExternalHook` - 外部 Hook

#### 6.2 Hook 输入
- `hook.PreCompactInput` - Hook 输入数据
  - `EventName()` - 事件名称
  - `ToolUseID()` - 工具调用 ID

#### 6.3 Agent Runner 方法
- `handlePreCompact()` - 处理压缩前逻辑
- `handleSessionStartAfterCompact()` - 压缩后会话恢复

### 7. 配置系统

#### 7.1 环境配置
- `IsSessionMemoryCompactEnabled` - Session Memory 开关
- `AutoCompactEnabled` - 自动压缩开关 (JSON 配置)
- `GetAutoCompactEnabled()` / `SetAutoCompactEnabled()` - 配置访问

#### 7.2 阈值配置
- `getCompactThreshold()` - 获取压缩阈值

### 8. TUI 集成

#### 8.1 CompactCommand 组件
- `tui/components/command/compact.go`
- 完整的 UI 组件，支持:
  - 键盘处理 (AddKeyHandler, Default*Handler)
  - 渲染 (RenderView, RenderItem, RenderProgress)
  - 状态管理 (GetCheckboxState, SetCheckboxStates)
  - 刷新机制 (SetRefreshHandler, CancelRefreshHandler)

#### 8.2 消息视图
- `tui/components/messages.CompactResult`
  - `Render()` - 渲染结果
  - `IsError()` - 错误检查
  - `NoContentTips()` - 无内容提示

#### 8.3 处理流程
- `processCompactPayload` - 处理压缩载荷
- `handleCompactComplete` - 处理压缩完成

### 9. ACP (Agent Communication Protocol) 集成

- `acp.(*Session).Compact` - ACP Session 压缩
- `acp.(*Session).handleCompactEvent` - 处理压缩事件
- `acp.(*Session).sendCompactionNotification` - 发送压缩通知

### 10. 关键日志/状态字符串

```
"compaction_triggered"      - 压缩被触发
"compaction_start"          - 压缩开始
"compaction_cancel"         - 压缩取消
"compact_complete"          - 压缩完成
"micro compact complete, compressedResults: %d, savedTokens: %d"
"Starting summarize compact"
"Finished summarize compact"
"Session memory update completed, backtracked to complete turn"
"Truncate compact history until tokens below 0.98"
"SM state reset after traditional compact fallback"
```

---

## 当前反编译代码与原版对比

### 当前实现 (agent.go:502-557)

```go
func (a *Agent) compactContext(ctx context.Context) error {
    messages := a.state.GetMessages()
    if len(messages) <= 2 {
        return nil
    }
    
    // 1. 保留系统消息
    // 2. 简单截取前 50 字符生成摘要
    // 3. 添加摘要作为 system 消息
    // 4. 保留最后 2 条消息
    // 5. 更新状态
    
    return nil
}
```

### 功能差距分析

| 功能模块 | 原版 | 当前反编译版 | 差距 |
|---------|------|-------------|------|
| **触发方式** | Auto/Manual/Reload | 仅 Manual | 🔴 缺少自动触发 |
| **Token 计数** | 完整追踪 | 无 | 🔴 完全缺失 |
| **阈值检测** | getCompactThreshold | 无 | 🔴 完全缺失 |
| **MicroCompact** | 有 | 无 | 🔴 完全缺失 |
| **Session Memory** | 独立包，完整实现 | 无 | 🔴 完全缺失 |
| **LLM 摘要** | summarizeCompact | 无 | 🔴 完全缺失 |
| **Pre-Compact Hook** | 完整 Hook 系统 | 无 | 🔴 完全缺失 |
| **配置系统** | AutoCompactEnabled 等 | 无 | 🔴 完全缺失 |
| **持久化** | Session Memory 文件 | 无 | 🔴 完全缺失 |
| **TUI 组件** | CompactCommand| 无 | 🔴 完全缺失 |
| **ACP 集成** | sendCompactionNotification | 无 | 🔴 完全缺失 |
| **压缩策略** | 5 种策略 | 1 种简单策略 | 🔴 严重简化 |
| **状态管理** | SerialQueue 保证顺序 | 简单切片 | 🟡 简化 |
| **分块处理** | SectionTokenInfo | 无 | 🔴 完全缺失 |

### 代码行数对比

| 项目 | 原版 (推测) | 当前 |
|-----|----------|------|
| Compact 核心 | ~2000+ 行 | ~50 行 |
| Session Memory | ~1000+ 行 | 0 行 |
| Hooks | ~500+ 行 | 0 行 |
| TUI 组件 | ~800+ 行 | 0 行 |
| **总计** | **~4000+ 行** | **~50 行** |

---

## 增强建议优先级

### P0 - 核心功能 (必须实现)
1. Token 计数和阈值检测
2. 自动 Compact 触发机制
3. CompactTrigger 类型定义

### P1 - 重要增强
4. MicroCompact 快速压缩
5. Pre-Compact Hook 系统
6. Session Memory 基础框架

### P2 - 体验优化
7. TUI CompactCommand 组件
8. 配置系统 (AutoCompactEnabled)
9. 压缩统计 (savedTokens 等)

### P3 - 高级功能
10. LLM 摘要 (summarizeCompact)
11. ACP 集成
12. 持久化和断点续传

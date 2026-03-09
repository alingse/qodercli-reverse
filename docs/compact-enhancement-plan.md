# Compact 功能增强计划

> **分析日期**: 2026-03-09  
> **分析方法**: 二进制符号分析 (`strings` + `grep`)  
> **相关文档**: [compact-architecture-analysis.md](./compact-architecture-analysis.md)

## 当前状态总结

通过 `which qodercli` 定位二进制文件，使用 `strings` 提取了 157,441 行字符串，从中筛选出 364 行 compact 相关内容，分析了完整的函数符号表。

### 关键发现

**原版 compact 是一个复杂的系统，包含:**
- 5 种压缩策略 (Micro/Full/SessionMemory/Summarize/Traditional)
- 3 种触发方式 (Auto/Manual/Reload)
- 完整的 Token 计数和阈值检测
- Session Memory 独立持久化模块
- Pre-Compact Hook 扩展系统
- TUI 专用组件
- ACP 协议集成

**当前反编译代码只有 ~50 行的简化实现，缺少上述所有高级功能。**

## 目录

- [Phase 1: 基础架构 (P0)](#phase-1-基础架构-p0)
- [Phase 2: MicroCompact (P1)](#phase-2-microcompact-p1)
- [Phase 3: FullCompact 和 LLM 摘要 (P1-P2)](#phase-3-fullcompact-和 -llm-摘要-p1-p2)
- [Phase 4: Session Memory 基础 (P1)](#phase-4-session-memory-基础-p1)
- [Phase 5: Pre-Compact Hook 系统 (P1)](#phase-5-pre-compact-hook-系统-p1)
- [Phase 6: 配置系统 (P2)](#phase-6-配置系统-p2)
- [Phase 7: TUI 集成 (P2)](#phase-7-tui-集成-p2)
- [实施顺序建议](#实施顺序建议)
- [风险和挑战](#风险和挑战)
- [验收标准](#验收标准)

---

## Phase 1: 基础架构 (P0)

### 1.1 定义 CompactTrigger 类型

**文件**: `core/types/compact.go`

```go
package types

// CompactTriggerType 压缩触发类型
type CompactTriggerType string

const (
    CompactTriggerAuto   CompactTriggerType = "auto"
    CompactTriggerManual CompactTriggerType = "manual"
    CompactTriggerReload CompactTriggerType = "reload"
)

// CompactTrigger 压缩触发器
type CompactTrigger struct {
    Type        CompactTriggerType `json:"type"`
    TokenCount  int                `json:"token_count,omitempty"`
    Threshold   int                `json:"threshold,omitempty"`
    Instruction string             `json:"instruction,omitempty"` // 用户指令（手动触发时）
}

// IsAuto 是否是自动触发
func (c *CompactTrigger) IsAuto() bool {
    return c.Type == CompactTriggerAuto
}

// IsManual 是否是手动触发
func (c *CompactTrigger) IsManual() bool {
    return c.Type == CompactTriggerManual
}

// IsReload 是否是重载触发
func (c *CompactTrigger) IsReload() bool {
    return c.Type == CompactTriggerReload
}

// String 返回字符串表示
func (c *CompactTrigger) String() string {
    return string(c.Type)
}
```

### 1.2 添加 Token 计数到 State

**文件**: `core/agent/state/state.go`

```go
// 需要新增的方法
type State struct {
    messages       []types.Message
    toolResults    map[string]*types.ToolResult
    tokenCount     int  // 新增：当前 token 计数
    mu             sync.RWMutex
}

// GetTokenCount 获取当前 token 数
func (s *State) GetTokenCount() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.tokenCount
}

// UpdateTokenCount 更新 token 计数
func (s *State) UpdateTokenCount(count int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.tokenCount = count
}
```

### 1.3 实现阈值检测

**文件**: `core/agent/compact.go` (新建)

```go
package agent

const (
    // DefaultCompactThreshold 默认压缩阈值 (80% 上下文)
    DefaultCompactThreshold = 0.8
    
    // WarningThresholdK 警告阈值 (以 K tokens 为单位)
    WarningThresholdK = 100 // 100K tokens
)

// getCompactThreshold 获取压缩阈值
func getCompactThreshold(modelMaxTokens int) int {
    return int(float64(modelMaxTokens) * DefaultCompactThreshold)
}

// shouldRunMicroCompact 判断是否运行微型压缩
func shouldRunMicroCompact(currentTokens, threshold int) bool {
    // 小幅超限 (不超过 10%) 使用微型压缩
    return currentTokens > threshold && currentTokens < int(float64(threshold)*1.1)
}
```

### 1.4 自动 Compact 触发机制

**文件**: `core/agent/agent.go` (修改 generate 方法)

```go
// 在 generate() 方法的循环开始时添加
func (a *Agent) generate(ctx context.Context) error {
    for {
        // === 新增：检查是否需要自动压缩 ===
        if err := a.checkAndAutoCompact(ctx); err != nil {
            log.Debug("Auto compact check failed: %v", err)
        }
        // ===================================
        
        // ... 现有逻辑
    }
}

// checkAndAutoCompact 检查并执行自动压缩
func (a *Agent) checkAndAutoCompact(ctx context.Context) error {
    // 1. 获取当前 token 数
    currentTokens := a.state.GetTokenCount()
    
    // 2. 获取模型配置
    maxTokens := a.config.MaxTokens
    threshold := getCompactThreshold(maxTokens)
    
    // 3. 判断是否需要压缩
    if currentTokens <= threshold {
        return nil // 无需压缩
    }
    
    // 4. 选择压缩策略
    trigger := &types.CompactTrigger{
        Type:       types.CompactTriggerAuto,
        TokenCount: currentTokens,
        Threshold:  threshold,
    }
    
    if shouldRunMicroCompact(currentTokens, threshold) {
        return a.MicroCompact(ctx, trigger)
    }
    
    return a.fullCompact(ctx, trigger)
}
```

---

## Phase 2: MicroCompact (P1)

### 2.1 实现 MicroCompact

**文件**: `core/agent/compact.go`

```go
// MicroCompact 微型压缩 - 快速轻量级
func (a *Agent) MicroCompact(ctx context.Context, trigger *types.CompactTrigger) error {
    messages := a.state.GetMessages()
    if len(messages) <= 3 {
        return nil
    }
    
    var compressed []types.Message
    
    // 1. 保留系统消息
    if len(messages) > 0 && messages[0].Role == types.RoleSystem {
       compressed = append(compressed, messages[0])
    }
    
    // 2. 保留最近 N 条完整对话 (例如 5 轮)
    keepRounds := 5
    keepCount := keepRounds * 2 // user + assistant
    if len(messages) > keepCount+1 { // +1 for system message
        // 3. 为被移除的消息生成简要摘要
        summary := a.generateMicroSummary(messages[1 : len(messages)-keepCount])
       compressed = append(compressed, types.Message{
            Role:    types.RoleSystem,
            Content: []types.ContentPart{{Type: "text", Text: summary}},
        })
    }
    
    // 4. 保留最近的对话
    if len(messages) > keepCount {
       compressed = append(compressed, messages[len(messages)-keepCount:]...)
    }
    
    // 5. 更新状态
    a.state.SetMessages(compressed)
    
    // 6. 报告统计
    savedTokens := trigger.TokenCount - a.calculateTokenCount(compressed)
    log.Debug("micro compact complete, compressedResults: %d, savedTokens: %d", 
        len(messages)-len(compressed), savedTokens)
    
    return nil
}

// generateMicroSummary 生成微型摘要 (仅提取关键点)
func (a *Agent) generateMicroSummary(messages []types.Message) string {
    var summary strings.Builder
    summary.WriteString("Previous conversation summary:\n")
    
    for _, msg := range messages {
        if msg.Role == types.RoleUser {
            summary.WriteString("- User: ")
            for _, part := range msg.Content {
                if part.Type == "text" {
                    // 取前 100 字符
                    text := part.Text
                    if len(text) > 100 {
                        text = text[:100] + "..."
                    }
                    summary.WriteString(text)
                    break
                }
            }
            summary.WriteString("\n")
        }
    }
    
    return summary.String()
}
```

---

## Phase 3: FullCompact 和 LLM 摘要 (P1-P2)

### 3.1 FullCompact 框架

```go
// fullCompact 完整压缩
func (a *Agent) fullCompact(ctx context.Context, trigger *types.CompactTrigger) error {
    // 1. 发送压缩开始通知
    a.reportCompactStart(trigger)
    
    // 2. 执行压缩 (可能调用 LLM)
    var err error
    if a.shouldSummarizeCompact(trigger) {
        err = a.summarizeCompact(ctx, trigger)
    } else {
        err = a.traditionalSummarizeCompact(ctx, trigger)
    }
    
    // 3. 发送完成通知
    a.reportCompactComplete(trigger, err)
    
    return err
}
```

### 3.2 LLM 摘要 (summarizeCompact)

```go
// summarizeCompact 使用 LLM 生成语义摘要
func (a *Agent) summarizeCompact(ctx context.Context, trigger *types.CompactTrigger) error {
    messages := a.state.GetMessages()
    
    // 构建摘要请求
    summaryReq := &provider.ModelRequest{
        Model:      a.config.Model,
        Messages:   messages,
        SystemPrompt: "Please summarize the key points of this conversation...",
        MaxTokens:  1000,
    }
    
    // 调用 LLM 生成摘要
    // ... (需要调用 provider.Stream 或类似方法)
    
    return nil
}
```

---

## Phase 4: Session Memory 基础 (P1)

### 4.1 Session Memory 状态接口

**文件**: `core/agent/state/session_memory/state.go` (新建目录)

```go
package session_memory

import (
    "github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// State Session Memory 状态接口
type State interface {
    // 配置
    GetCompactConfig() *SMCompactConfig
    
    // 记忆管理
    GetCurrentMemory() string
    GetLastSummarizedMsgId() string
    
    // 更新控制
    TriggerUpdate() error
    WaitForUpdate(ctx context.Context) error
    MarkUpdateStart()
    MarkUpdateComplete()
    
    // 状态检查
    ShouldUpdate() bool
    HasPendingContent() bool
    HasPendingToolUse() bool
    
    // 持久化
    GetMemoryPath() string
    EnsureTemplate() error
}

// SMCompactConfig Session Memory 压缩配置
type SMCompactConfig struct {
    Enabled           bool   `json:"enabled"`
    AutoCompact       bool   `json:"auto_compact"`
    ThresholdPercent  float64 `json:"threshold_percent"`
    KeepLastRounds    int    `json:"keep_last_rounds"`
}
```

### 4.2 SerialQueue 串行队列

```go
// SerialQueue 保证更新操作顺序执行
type SerialQueue struct {
    queue    chan func()
    running  bool
    mu       sync.Mutex
}

// NewSerialQueue 创建串行队列
func NewSerialQueue() *SerialQueue {
    return &SerialQueue{
        queue:   make(chan func(), 100),
        running: true,
    }
}

// Enqueue 入队
func (q *SerialQueue) Enqueue(task func()) {
    select {
    case q.queue <- task:
    default:
        // 队列满，阻塞等待
        q.queue <- task
    }
}

// processQueue 处理队列
func (q *SerialQueue) processQueue() {
    for task := range q.queue {
        if !q.running {
            return
        }
        task()
    }
}
```

---

## Phase 5: Pre-Compact Hook 系统 (P1)

### 5.1 Hook 定义

**文件**: `core/agent/hooks/pre_compact/pre_compact.go` (新建目录)

```go
package pre_compact

import (
    "context"
    "github.com/alingse/qodercli-reverse/decompiled/core/resource/hook"
)

// PreCompactHook 压缩前 Hook 接口
type PreCompactHook interface {
    Execute(ctx context.Context, input *hook.PreCompactInput) (*hook.Result, error)
}

// PreCompactExternalHook 外部 Hook 实现
type PreCompactExternalHook struct {
    handler func(ctx context.Context, input *hook.PreCompactInput) (*hook.Result, error)
}

// Execute 执行 Hook
func (h *PreCompactExternalHook) Execute(ctx context.Context, input *hook.PreCompactInput) (*hook.Result, error) {
    return h.handler(ctx, input)
}

// NewPreCompactExternalHook 创建外部 Hook
func NewPreCompactExternalHook(handler func(context.Context, *hook.PreCompactInput) (*hook.Result, error)) *PreCompactExternalHook {
    return &PreCompactExternalHook{handler: handler}
}
```

### 5.2 Hook 输入

**文件**: `core/resource/hook/hook.go`

```go
// PreCompactInput 压缩前 Hook 输入
type PreCompactInput struct {
    TriggerType string
    TokenCount   int
    Threshold    int
    MessageCount int
    ToolUseID    string // 如果有工具调用
}

// EventName 返回事件名称
func (p *PreCompactInput) EventName() string {
    return "pre_compact"
}

// ToolUseID 返回工具调用 ID
func (p *PreCompactInput) ToolUseID() string {
    return p.ToolUseID
}
```

---

## Phase 6: 配置系统 (P2)

### 6.1 环境配置

**文件**: `core/utils/env/env.go`

```go
// IsSessionMemoryCompactEnabled 检查 Session Memory 是否启用
func IsSessionMemoryCompactEnabled() bool {
    return os.Getenv("ENABLE_SESSION_MEMORY_COMPACT") == "true"
}

// GetAutoCompactEnabled 获取自动压缩开关
func GetAutoCompactEnabled() bool {
    val := os.Getenv("AUTO_COMPACT_ENABLED")
    return val == "" || val == "true" // 默认启用
}
```

### 6.2 Agent 配置

```go
// Config 增加 compact 相关配置
type Config struct {
    // ... 现有字段
    
    AutoCompactEnabled bool    `json:"auto_compact_enabled"`
    CompactThreshold    float64 `json:"compact_threshold"` // 0.0-1.0
    KeepLastRounds      int     `json:"keep_last_rounds"`
    UseLLMSummary       bool    `json:"use_llm_summary"`
}
```

---

## Phase 7: TUI 集成 (P2)

### 7.1 CompactResult 消息视图

**文件**: `tui/components/messages/compact_result.go`

```go
package messages

// CompactResult 压缩结果
type CompactResult struct {
    CompressedCount int
    SavedTokens     int
    Error           error
}

// Render 渲染结果
func (c *CompactResult) Render(width int) string {
    if c.Error != nil {
        return fmt.Sprintf("❌ Compact failed: %v", c.Error)
    }
    
    return fmt.Sprintf(
        "✓ Compacted (%d messages → %d, saved ~%d tokens)",
        c.CompressedCount,
        c.CompressedCount/3, // 估算
        c.SavedTokens,
    )
}

// IsError 检查是否有错误
func (c *CompactResult) IsError() bool {
    return c.Error != nil
}

// NoContentTips 返回无内容提示
func (c *CompactResult) NoContentTips() string {
    return "Not enough messages to compact"
}
```

---

## 实施顺序建议

1. **Week 1**: Phase 1 (基础架构)
   - 定义 CompactTrigger
   - 添加 Token 计数
   - 实现阈值检测

2. **Week 2**: Phase 2 (MicroCompact)
   - 实现快速压缩
   - 添加统计报告

3. **Week 3**: Phase 4 (Session Memory 基础)
   - 状态接口
   - 串行队列

4. **Week 4**: Phase 5 (Hook 系统)
   - Pre-Compact Hook
   - 外部 Hook 支持

5. **Week 5-6**: Phase 3 & 6 (LLM 摘要 + 配置)
   - 完整压缩流程
   - 配置系统

6. **Week 7-8**: Phase 7 (TUI) + 测试
   - UI 组件
   - 集成测试

---

## 风险和挑战

1. **Token 计数准确性**: 需要精确的 token 计算库
2. **LLM 摘要成本**: 调用 LLM 会增加 API 费用
3. **Session Memory 持久化**: 需要处理并发和崩溃恢复
4. **向后兼容**: 确保不影响现有功能

---

## 验收标准

- [ ] `/compact` 命令正常工作
- [ ] 自动压缩在阈值触发
- [ ] Token 计数准确
- [ ] MicroCompact 在 1 秒内完成
- [ ] Session Memory 正确持久化
- [ ] Hook 系统可扩展
- [ ] TUI 显示压缩统计

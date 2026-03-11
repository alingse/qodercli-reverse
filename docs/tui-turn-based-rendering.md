# TUI 基于回合的渲染改进

## 问题背景

原有的 TUI 实现存在输出截断问题：
- 流式内容和工具调用之间的输出被吞掉
- `<logs>` 等内容没有正确输出
- 换行和流截断问题
- 中间输出的时序问题导致内容丢失

## 核心改进思路

### 1. 引入"回合"（Turn）概念

一个回合包含：
- 用户输入
- Agent 的流式文本响应
- 工具调用列表（0个或多个）
- 工具执行结果

### 2. 渲染策略变更

**之前的策略**：
- 流式内容实时通过 `tea.Printf` 输出
- 工具调用开始时立即 `tea.Printf` 输出之前的内容
- 工具结果时立即 `tea.Printf` 输出工具结果
- 问题：中间输出可能导致内容截断

**新的策略**：
- 在 `View()` 中临时渲染当前回合的所有内容
- 等到回合完成（`onFinish`）时，一次性用 `tea.Printf` 输出完整内容
- 优点：避免中间输出导致的截断，保证内容完整性

## 实现细节

### 数据结构

```go
// TurnContent 表示一个完整的 Agent 回合内容
type TurnContent struct {
    StreamingText strings.Builder  // 流式文本内容
    ToolCalls     []TurnToolCall    // 工具调用列表
    Status        TurnStatus        // 回合状态
    StartTime     time.Time         // 开始时间
}

// TurnToolCall 表示回合中的一个工具调用
type TurnToolCall struct {
    ID        string
    Name      string
    Arguments string
    Output    string
    IsError   bool
    Status    ToolCallStatus  // pending/running/completed/error
    StartTime time.Time
    EndTime   time.Time
}
```

### appModel 结构变更

**移除的字段**：
- `streamingBuffer` - 用 `currentTurn.StreamingText` 替代
- `toolInfoMap` - 用 `currentTurn.ToolCalls` 替代
- `lastResponseLen` - 不再需要增量追踪
- `streamContentBuffer` - 不再需要缓冲区
- `streamBufferLastFlush` - 不再需要刷新时间
- `streamSeqNum` / `toolStartSeqNum` / `streamSeqMutex` - 不再需要序列号同步

**新增的字段**：
- `currentTurn *TurnContent` - 当前回合内容
- `hasTurn bool` - 是否有活跃的回合

### Update() 方法变更

#### editor.SendMsg（用户输入）
```go
// 1. 持久化用户消息
tea.Printf(userMsg.Render())

// 2. 初始化新回合
m.hasTurn = true
m.currentTurn = &TurnContent{
    Status:    TurnStatusStreaming,
    StartTime: time.Now(),
}

// 3. 启动 Agent
m.handleUserInput(msg.Content)
```

#### AgentStreamMsg（流式内容）
```go
// 更新当前回合的流式文本（完整内容，不是增量）
if m.hasTurn && m.currentTurn != nil {
    m.currentTurn.StreamingText.Reset()
    m.currentTurn.StreamingText.WriteString(msg.Content)
    m.currentTurn.Status = TurnStatusStreaming
}
```

#### AgentToolStartMsg（工具开始）
```go
// 添加到当前回合的工具调用列表
if m.hasTurn && m.currentTurn != nil {
    m.currentTurn.Status = TurnStatusTooling
    m.currentTurn.ToolCalls = append(m.currentTurn.ToolCalls, TurnToolCall{
        ID:        msg.ID,
        Name:      msg.Name,
        Arguments: msg.Arguments,
        Status:    ToolCallStatusRunning,
        StartTime: time.Now(),
    })
}
```

#### AgentToolResultMsg（工具结果）
```go
// 更新对应的工具调用状态
if m.hasTurn && m.currentTurn != nil {
    for i := range m.currentTurn.ToolCalls {
        if m.currentTurn.ToolCalls[i].ID == msg.ToolCallID {
            m.currentTurn.ToolCalls[i].Output = msg.Content
            m.currentTurn.ToolCalls[i].IsError = msg.IsError
            m.currentTurn.ToolCalls[i].Status = ToolCallStatusCompleted
            m.currentTurn.ToolCalls[i].EndTime = time.Now()
            break
        }
    }
}
```

#### AgentFinishMsg（回合完成）
```go
// 一次性持久化所有内容
if m.hasTurn && m.currentTurn != nil {
    // 1. 持久化流式文本
    if m.currentTurn.StreamingText.Len() > 0 {
        tea.Printf(assistantMsg.Render())
    }

    // 2. 持久化所有工具调用
    for _, toolCall := range m.currentTurn.ToolCalls {
        tea.Printf(toolMsg.Render())
    }

    // 3. 清理当前回合
    m.hasTurn = false
    m.currentTurn = nil
}
```

### View() 方法变更

```go
func (m appModel) View() string {
    var sections []string

    // 1. 当前回合预览区（如果有活跃回合）
    if m.hasTurn && m.currentTurn != nil {
        sections = append(sections, m.renderCurrentTurn())
    }

    // 2. 空行分隔
    sections = append(sections, "")

    // 3. 编辑器
    sections = append(sections, m.editor.View())

    // 4. 状态栏
    sections = append(sections, m.renderStatusBar())

    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
```

#### renderCurrentTurn() 方法

```go
func (m appModel) renderCurrentTurn() string {
    var sections []string

    // 1. 渲染流式文本（如果有）
    if m.currentTurn.StreamingText.Len() > 0 {
        // 使用斜体灰色样式表示正在生成
        sections = append(sections, previewStyle.Render(preview.Render()))
    }

    // 2. 渲染工具调用（如果有）
    for _, toolCall := range m.currentTurn.ToolCalls {
        // 根据状态使用不同颜色
        // pending: 灰色, running: 白色, completed: 绿色, error: 红色
        sections = append(sections, style.Render(toolMsg.Render()))
    }

    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
```

### 回调处理简化

```go
func (m *appModel) setupAgentCallbacks() {
    m.agent.SetCallbacks(
        // onMessage: 直接发送完整内容
        func(msg *types.Message) {
            fullStr := extractFullContent(msg)
            m.callbackChan <- func() {
                m.eventChan <- AgentStreamMsg{Content: fullStr}
            }
        },

        // onToolCall: 直接发送工具调用信息
        func(call *types.ToolCall) {
            m.callbackChan <- func() {
                m.eventChan <- AgentToolStartMsg{...}
            }
        },

        // onToolResult: 直接发送工具结果
        func(result *types.ToolResult) {
            m.callbackChan <- func() {
                m.eventChan <- AgentToolResultMsg{...}
            }
        },

        // onFinish: 直接发送完成消息
        func(reason types.FinishReason) {
            m.callbackChan <- func() {
                m.eventChan <- AgentFinishMsg{Reason: reason}
            }
        },
    )
}
```

**移除的复杂逻辑**：
- 流式缓冲区管理（`flushStreamBuffer`）
- 增量内容追踪（`lastResponseLen`）
- 序列号同步（`streamSeqNum` / `toolStartSeqNum`）
- 强制刷新逻辑

## 优势

### 1. 内容完整性
- 整个回合的内容作为一个整体输出，避免截断
- 不再有中间输出导致的内容丢失

### 2. 实时预览
- 通过 `View()` 提供实时预览，用户可以看到进度
- 流式文本和工具调用状态清晰可见

### 3. 代码简化
- 移除了复杂的缓冲区管理逻辑
- 移除了序列号同步机制
- 回调处理更加直观

### 4. 状态清晰
- 工具调用的状态（pending/running/completed/error）清晰可见
- 使用不同颜色区分不同状态

## 边界情况处理

### 1. 错误情况
```go
case AgentErrorMsg:
    // 持久化当前回合的部分内容
    if m.hasTurn && m.currentTurn != nil {
        // 持久化已有的流式文本
        // 持久化已完成的工具调用
    }

    // 持久化错误消息
    tea.Printf(errMsg.Render())

    // 清理回合
    m.hasTurn = false
    m.currentTurn = nil
```

### 2. 多回合循环
Agent 的 `generate()` 方法已经处理了多回合循环：
- 如果有工具调用，执行工具后继续下一个回合
- 如果没有工具调用，触发 `onFinish` 回调

我们的实现会在每次 `AgentStreamMsg` 时更新流式文本，在 `AgentToolStartMsg` 时添加新的工具调用，在 `AgentFinishMsg` 时持久化所有内容。

## 测试建议

1. **单回合场景**：用户输入 → Agent 响应（无工具调用）
2. **工具调用场景**：用户输入 → Agent 响应 → 工具调用 → Agent 继续响应
3. **多回合循环场景**：用户输入 → Agent 响应 → 工具调用 → Agent 继续响应 → 工具调用 → Agent 最终响应
4. **错误场景**：工具调用失败、Agent 错误
5. **长文本场景**：测试大量流式内容的渲染性能

## 注意事项

1. **内存管理**：`TurnContent` 可能会累积大量内容，需要在回合完成后立即清理
2. **并发安全**：Agent 回调在 goroutine 中执行，通过 `callbackChan` 序列化保证顺序性
3. **性能考虑**：每次 `AgentStreamMsg` 都会触发 `View()` 重新渲染，对于高频率的流式更新可能会导致 CPU 占用较高

## 文件变更

- `tui/app/model.go` - 主要修改文件
  - 添加 `TurnContent`、`TurnToolCall`、`TurnStatus`、`ToolCallStatus` 数据结构
  - 修改 `appModel` 结构，添加 `currentTurn` 和 `hasTurn` 字段
  - 重写 `Update()` 方法中的消息处理逻辑
  - 添加 `renderCurrentTurn()` 方法
  - 简化 `setupAgentCallbacks()` 方法
  - 移除 `flushStreamBuffer()` 方法
  - 移除未使用的字段和导入

## 总结

这次改进通过引入"回合"概念，将 Agent 的一次完整交互作为一个整体来管理和渲染，解决了原有实现中的输出截断问题。新的实现更加简洁、清晰，同时保证了内容的完整性和实时预览的用户体验。

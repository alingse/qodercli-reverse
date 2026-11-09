# TUI Turn-Based Rendering Improvements

## Problem Background

The original TUI implementation had output truncation issues:
- Streaming content and tool calls between outputs were swallowed
- Content such as `<logs>` was not output correctly
- Line break and stream truncation issues
- Timing issues with intermediate output causing content loss

## Core Improvement Idea

### 1. Introduce the "Turn" Concept

A turn contains:
- User input
- Agent's streaming text response
- Tool call list (0 or more)
- Tool execution results

### 2. Rendering Strategy Change

**Previous strategy**:
- Streaming content was output in real time via `tea.Printf`
- When a tool call starts, `tea.Printf` immediately output the previous content
- When a tool result arrives, `tea.Printf` immediately output the tool result
- Problem: intermediate output may cause content truncation

**New strategy**:
- Temporarily render all content of the current turn in `View()`
- Once the turn completes (`onFinish`), output the complete content at once via `tea.Printf`
- Advantage: avoids truncation caused by intermediate output, guarantees content completeness

## Implementation Details

### Data Structures

```go
// TurnContent represents a complete Agent turn
type TurnContent struct {
    StreamingText strings.Builder  // streaming text content
    ToolCalls     []TurnToolCall    // tool call list
    Status        TurnStatus        // turn status
    StartTime     time.Time         // start time
}

// TurnToolCall represents a tool call within a turn
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

### appModel Structure Changes

**Removed fields**:
- `streamingBuffer` - replaced by `currentTurn.StreamingText`
- `toolInfoMap` - replaced by `currentTurn.ToolCalls`
- `lastResponseLen` - no longer need incremental tracking
- `streamContentBuffer` - no longer need buffer
- `streamBufferLastFlush` - no longer need flush time
- `streamSeqNum` / `toolStartSeqNum` / `streamSeqMutex` - no longer need sequence number sync

**New fields**:
- `currentTurn *TurnContent` - current turn content
- `hasTurn bool` - whether there is an active turn

### Update() Method Changes

#### editor.SendMsg (user input)
```go
// 1. Persist user message
tea.Printf(userMsg.Render())

// 2. Initialize new turn
m.hasTurn = true
m.currentTurn = &TurnContent{
    Status:    TurnStatusStreaming,
    StartTime: time.Now(),
}

// 3. Start Agent
m.handleUserInput(msg.Content)
```

#### AgentStreamMsg (streaming content)
```go
// Update current turn's streaming text (full content, not incremental)
if m.hasTurn && m.currentTurn != nil {
    m.currentTurn.StreamingText.Reset()
    m.currentTurn.StreamingText.WriteString(msg.Content)
    m.currentTurn.Status = TurnStatusStreaming
}
```

#### AgentToolStartMsg (tool start)
```go
// Add to current turn's tool call list
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

#### AgentToolResultMsg (tool result)
```go
// Update corresponding tool call status
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

#### AgentFinishMsg (turn complete)
```go
// Persist all content at once
if m.hasTurn && m.currentTurn != nil {
    // 1. Persist streaming text
    if m.currentTurn.StreamingText.Len() > 0 {
        tea.Printf(assistantMsg.Render())
    }

    // 2. Persist all tool calls
    for _, toolCall := range m.currentTurn.ToolCalls {
        tea.Printf(toolMsg.Render())
    }

    // 3. Clean up current turn
    m.hasTurn = false
    m.currentTurn = nil
}
```

### View() Method Changes

```go
func (m appModel) View() string {
    var sections []string

    // 1. Current turn preview area (if there is an active turn)
    if m.hasTurn && m.currentTurn != nil {
        sections = append(sections, m.renderCurrentTurn())
    }

    // 2. Empty line separator
    sections = append(sections, "")

    // 3. Editor
    sections = append(sections, m.editor.View())

    // 4. Status bar
    sections = append(sections, m.renderStatusBar())

    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
```

#### renderCurrentTurn() Method

```go
func (m appModel) renderCurrentTurn() string {
    var sections []string

    // 1. Render streaming text (if any)
    if m.currentTurn.StreamingText.Len() > 0 {
        // Use italic gray style to indicate generating
        sections = append(sections, previewStyle.Render(preview.Render()))
    }

    // 2. Render tool calls (if any)
    for _, toolCall := range m.currentTurn.ToolCalls {
        // Use different colors based on status
        // pending: gray, running: white, completed: green, error: red
        sections = append(sections, style.Render(toolMsg.Render()))
    }

    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
```

### Callback Handling Simplification

```go
func (m *appModel) setupAgentCallbacks() {
    m.agent.SetCallbacks(
        // onMessage: send full content directly
        func(msg *types.Message) {
            fullStr := extractFullContent(msg)
            m.callbackChan <- func() {
                m.eventChan <- AgentStreamMsg{Content: fullStr}
            }
        },

        // onToolCall: send tool call info directly
        func(call *types.ToolCall) {
            m.callbackChan <- func() {
                m.eventChan <- AgentToolStartMsg{...}
            }
        },

        // onToolResult: send tool result directly
        func(result *types.ToolResult) {
            m.callbackChan <- func() {
                m.eventChan <- AgentToolResultMsg{...}
            }
        },

        // onFinish: send completion message directly
        func(reason types.FinishReason) {
            m.callbackChan <- func() {
                m.eventChan <- AgentFinishMsg{Reason: reason}
            }
        },
    )
}
```

**Removed complex logic**:
- Streaming buffer management (`flushStreamBuffer`)
- Incremental content tracking (`lastResponseLen`)
- Sequence number sync (`streamSeqNum` / `toolStartSeqNum`)
- Forced flush logic

## Advantages

### 1. Content Completeness
- The entire turn's content is output as a whole, avoiding truncation
- No more content loss caused by intermediate output

### 2. Real-Time Preview
- Provides real-time preview via `View()`, users can see progress
- Streaming text and tool call status clearly visible

### 3. Code Simplification
- Removed complex buffer management logic
- Removed sequence number sync mechanism
- Callback handling more intuitive

### 4. Clear State
- Tool call status (pending/running/completed/error) clearly visible
- Different colors distinguish different states

## Edge Case Handling

### 1. Error Cases
```go
case AgentErrorMsg:
    // Persist partial content of current turn
    if m.hasTurn && m.currentTurn != nil {
        // Persist existing streaming text
        // Persist completed tool calls
    }

    // Persist error message
    tea.Printf(errMsg.Render())

    // Clean up turn
    m.hasTurn = false
    m.currentTurn = nil
```

### 2. Multi-Turn Loop
The Agent's `generate()` method already handles the multi-turn loop:
- If there are tool calls, execute tools and continue to next turn
- If no tool calls, trigger `onFinish` callback

Our implementation updates streaming text on each `AgentStreamMsg`, adds new tool calls on `AgentToolStartMsg`, and persists all content on `AgentFinishMsg`.

## Testing Suggestions

1. **Single-turn scenario**: user input → Agent response (no tool calls)
2. **Tool call scenario**: user input → Agent response → tool call → Agent continues response
3. **Multi-turn loop scenario**: user input → Agent response → tool call → Agent continues response → tool call → Agent final response
4. **Error scenario**: tool call failure, Agent error
5. **Long text scenario**: test rendering performance with large amounts of streaming content

## Notes

1. **Memory management**: `TurnContent` may accumulate a large amount of content; clean up immediately after turn completion
2. **Concurrency safety**: Agent callbacks execute in goroutines; serialized through `callbackChan` to guarantee ordering
3. **Performance**: Each `AgentStreamMsg` triggers a `View()` re-render; high-frequency streaming updates may cause high CPU usage

## File Changes

- `tui/app/model.go` - main modified file
  - Add `TurnContent`, `TurnToolCall`, `TurnStatus`, `ToolCallStatus` data structures
  - Modify `appModel` structure, add `currentTurn` and `hasTurn` fields
  - Rewrite message handling logic in `Update()` method
  - Add `renderCurrentTurn()` method
  - Simplify `setupAgentCallbacks()` method
  - Remove `flushStreamBuffer()` method
  - Remove unused fields and imports

## Summary

This improvement introduces the "turn" concept, managing and rendering a complete Agent interaction as a whole, solving the output truncation problem in the original implementation. The new implementation is more concise and clear, while guaranteeing content completeness and the real-time preview user experience.
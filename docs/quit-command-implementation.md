# /quit 命令实现总结

## 实现方式

按照用户需求，实现了在 TUI 中显示统计信息并在退出时保留屏幕输出的功能。

## 核心设计

### 1. 不使用 AltScreen
```go
// tui/app/run.go
// 不使用 AltScreen，这样退出时会保留屏幕输出
p := tea.NewProgram(model)
```

**原因**：
- 使用 `tea.WithAltScreen()` 会在退出时清除屏幕
- 不使用 AltScreen，TUI 内容会保留在终端历史中
- 用户可以在退出后滚动查看完整的对话历史和统计信息

### 2. 统计信息显示在 TUI 中
```go
// tui/app/model.go
case "/quit", "/exit":
    // 获取统计信息
    stats := m.agent.GetState().GetStats()

    // 构建统计信息消息
    statsMsg := fmt.Sprintf(...)

    // 将统计信息添加到消息视图
    m.msgView.AddSystemMessage(statsMsg)

    // 设置退出标志
    m.quitting = true
```

**流程**：
1. 用户输入 `/quit` 或 `/exit`
2. 获取会话统计信息
3. 将统计信息作为系统消息添加到 TUI 的消息视图
4. 设置 `quitting` 标志
5. 下一次 Update 时检查标志并退出
6. 退出后，整个 TUI 内容（包括统计信息）保留在终端上

## 修改的文件

### 1. `decompiled/core/agent/state/state.go`
添加了会话统计信息：
- `Stats` 结构：跟踪 tokens、工具调用次数、助手回复次数
- `UpdateTokenUsage()`: 更新 token 使用统计
- `IncrementToolCallCount()`: 增加工具调用计数
- `IncrementAssistantReplies()`: 增加助手回复计数
- `GetStats()`: 获取统计信息

### 2. `decompiled/core/agent/agent/agent.go`
在 `generate()` 方法中收集统计信息：
- 在 `EventTypeMessageDelta` 事件中调用 `UpdateTokenUsage()`
- 在保存助手消息后调用 `IncrementAssistantReplies()`
- 在执行工具调用时调用 `IncrementToolCallCount()`

### 3. `decompiled/tui/app/run.go`
移除 AltScreen：
```go
// 不使用 WithAltScreen()，这样退出时会保留屏幕输出
p := tea.NewProgram(model)
```

### 4. `decompiled/tui/app/model.go`
- 添加 `handleCommand()` 方法处理斜杠命令
- 在 `handleUserInput()` 中检查斜杠命令
- 在 `Update()` 开始时检查 `quitting` 标志
- `/quit` 或 `/exit` 命令会在 TUI 中显示统计信息后退出

### 5. `decompiled/tui/components/messages/message_view.go`
添加 `AddSystemMessage()` 方法用于显示系统消息（如统计信息）

## 统计信息包括

- Total Tokens (Input + Output)
- Tool Calls (工具调用次数)
- Assistant Replies (助手回复次数)

## 使用方法

在 TUI 交互模式下输入：
```
/quit
```
或
```
/exit
```

退出时会在 TUI 中显示：
```
=== Session Statistics ===
Total Tokens: 1234 (Input: 567, Output: 667)
Tool Calls: 5
Assistant Replies: 3
=========================
```

然后程序退出，整个 TUI 内容（包括对话历史和统计信息）保留在终端上，用户可以滚动查看。

## 优势

✅ 统计信息显示在 TUI 界面上，与对话历史一致
✅ 退出后保留完整的屏幕输出
✅ 用户可以在终端中回看所有内容
✅ 不需要单独输出到 stderr
✅ 更简洁的实现


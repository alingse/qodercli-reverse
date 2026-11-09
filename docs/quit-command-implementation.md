# /quit Command Implementation Summary

## Implementation Approach

Per user requirements, implemented displaying statistics in the TUI and preserving screen output on exit.

## Core Design

### 1. No AltScreen
```go
// tui/app/run.go
// Do not use AltScreen, so screen output is preserved on exit
p := tea.NewProgram(model)
```

**Reason**:
- Using `tea.WithAltScreen()` clears the screen on exit
- Without AltScreen, TUI content remains in terminal history
- Users can scroll through the full conversation history and statistics after exiting

### 2. Statistics displayed in the TUI
```go
// tui/app/model.go
case "/quit", "/exit":
    // Get statistics
    stats := m.agent.GetState().GetStats()

    // Build statistics message
    statsMsg := fmt.Sprintf(...)

    // Add statistics to message view
    m.msgView.AddSystemMessage(statsMsg)

    // Set quitting flag
    m.quitting = true
```

**Flow**:
1. User enters `/quit` or `/exit`
2. Get session statistics
3. Add statistics as a system message to the TUI message view
4. Set the `quitting` flag
5. On the next Update, check the flag and exit
6. After exit, the entire TUI content (including statistics) remains on the terminal

## Modified Files

### 1. `decompiled/core/agent/state/state.go`
Added session statistics:
- `Stats` struct: tracks tokens, tool call count, assistant reply count
- `UpdateTokenUsage()`: updates token usage statistics
- `IncrementToolCallCount()`: increments tool call count
- `IncrementAssistantReplies()`: increments assistant reply count
- `GetStats()`: gets statistics

### 2. `decompiled/core/agent/agent/agent.go`
Collects statistics in the `generate()` method:
- Calls `UpdateTokenUsage()` on `EventTypeMessageDelta` events
- Calls `IncrementAssistantReplies()` after saving assistant messages
- Calls `IncrementToolCallCount()` when executing tool calls

### 3. `decompiled/tui/app/run.go`
Removed AltScreen:
```go
// Do not use WithAltScreen(), so screen output is preserved on exit
p := tea.NewProgram(model)
```

### 4. `decompiled/tui/app/model.go`
- Added `handleCommand()` method to handle slash commands
- Checks for slash commands in `handleUserInput()`
- Checks the `quitting` flag at the start of `Update()`
- The `/quit` or `/exit` command displays statistics in the TUI and then exits

### 5. `decompiled/tui/components/messages/message_view.go`
Added `AddSystemMessage()` method to display system messages (such as statistics)

## Statistics Include

- Total Tokens (Input + Output)
- Tool Calls (tool call count)
- Assistant Replies (assistant reply count)

## Usage

In TUI interactive mode, enter:
```
/quit
```
or
```
/exit
```

On exit, the TUI displays:
```
=== Session Statistics ===
Total Tokens: 1234 (Input: 567, Output: 667)
Tool Calls: 5
Assistant Replies: 3
=========================
```

Then the program exits, and the entire TUI content (including conversation history and statistics) remains on the terminal for the user to scroll through.

## Advantages

✅ Statistics are displayed on the TUI interface, consistent with conversation history
✅ Complete screen output is preserved after exit
✅ Users can review all content in the terminal
✅ No separate output to stderr needed
✅ Simpler implementation

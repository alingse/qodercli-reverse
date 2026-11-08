# TUI Architecture Analysis and Update Plan

## 1. Analysis Overview

This document is based on the `which qodercli-reverse + strings` analysis of the original qodercli-reverse binary, compared against the current decompiled code, aiming to restore the original TUI architecture 1:1.

## 2. Core Findings

### 2.1 Key Libraries Used by the Original (Extracted from strings)

```
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss
github.com/charmbracelet/glamour
code.alibaba-inc.com/qoder-core/bubbletea  (customized version)
```

### 2.2 Original TUI Component Structure

Based on symbol and string analysis, the original contains the following core components:

```
tui/
├── app/
│   ├── model.go         # appModel - main application model
│   └── run.go           # Run, RunWithInput, RunSDKMode
├── components/
│   ├── chat/             # chat view (possibly unused)
│   ├── editor/
│   │   └── interaction/
│   │       ├── editor.go    # EditorComponent - core editor
│   │       ├── attachment.go # AttachmentHandler
│   │       ├── history.go   # HistoryHandler  
│   │       └── cache.go     # InputCache
│   ├── messages/
│   │   ├── message_view.go  # MessageView
│   │   └── types.go         # message type definitions
│   └── status/           # status bar (possibly integrated)
```

### 2.3 Key Message Formats (Extracted from strings)

```
"raw msg"        - user message prefix (green)
"raw resp"       - assistant message prefix (blue/cyan)
"Input token usage: %d | Output token usage: %d" - Token stats
"Ask anything..." - editor placeholder
"Attachments: "   - attachment list prefix
```

## 3. Current Decompiled Code Issues

### 3.1 Architecture Issues

1. **Unclear component separation**: The current code has two editor packages:
   - `tui/components/editor/editor.go` - simple version
   - `tui/components/interaction/editor/editor.go` - full version (EditorComponent)
   
   The original should only use the `interaction/editor` version

2. **Redundant Chat component**: `tui/components/chat/chat.go` may be an early version; MessageView should be used instead

3. **Standalone Status component**: There is a standalone status component, but the original may integrate it into appModel

### 3.2 Missing Functionality

Based on original symbol analysis, the following functionality is missing:

1. **Complete keyboard shortcut handling**:
   - `alt+enter` - possibly used for special functionality
   - `ctrl+shift+end/home` - scroll to bottom/top
   - `pgup/pgdown` - page scrolling
   - `ctrl+u/d` - half-page scrolling

2. **Mouse support**:
   - scroll wheel event handling
   - click to focus

3. **Complete message types**:
   ```go
   MsgTypeUser        // raw msg
   MsgTypeAssistant   // raw resp  
   MsgTypeSystem
   MsgTypeTool
   MsgTypeBash
   MsgTypeCommand
   MsgTypeError
   MsgTypeCompact
   MsgTypeLog
   ```

4. **Token usage display**: Show `Input token usage: xxx | Output token usage: xxx` at the end of messages

5. **Bash command tracking**: Bash command display with status indicators (◐, ✓, ✗)

## 4. Update Plan

### Phase 1: Clean Up Redundant Code

#### 1.1 Remove or Deprecate Redundant Components

- [ ] Remove `tui/components/chat/chat.go` (or replace with messages)
- [ ] Remove `tui/components/editor/editor.go` (use interaction/editor)
- [ ] Consider integrating `tui/components/status/status.go` into appModel

#### 1.2 Unify Component Interfaces

- [ ] Ensure all components implement `Init() tea.Cmd`, `Update(msg tea.Msg)`, `View()` interfaces
- [ ] Unify message type definitions

### Phase 2: Update EditorComponent

#### 2.1 Core Functionality (exists, needs verification)

- [x] textarea integration
- [x] message sending logic
- [x] history navigation
- [x] attachment handling
- [ ] input cache

#### 2.2 Functionality to Add

- [ ] Complete keyboard shortcut mapping table
- [ ] Vim mode support (optional)
- [ ] Better error handling

### Phase 3: Update MessageView

#### 3.1 Message Type Definitions (refer to types.go)

Ensure all message types are correctly implemented:

```go
type Message interface {
   Type() MessageType
    Render() string
    Timestamp() time.Time
}
```

#### 3.2 Message Rendering Format

**User message (UserMessage)**:
```
raw msg          <- bold green
<content>
```

**Assistant message (AssistantMessage)**:
```
raw resp         <- bold blue
<content>
```

**Token usage (TokenUsageMessage)**:
```
Input token usage: 100 | Output token usage: 50  <- gray italic
```

**Bash command (BashInfo)**:
```
◐ Bash $ <command>      <- running (orange)
✓ Bash $ <command>      <- success (green)
✗ Bash $ <command>      <- failure (red)
<output>                <- truncated beyond 500 chars
```

**Tool call (ToolCall)**:
```
▶ Tool: <name>          <- bold blue
<arguments>             <- gray
```

**Tool result (ToolResult)**:
```
✓ <name>: <result>      <- success (green)
✗ <name>: <error>       <- failure (red)
```

### Phase 4: Update AppModel

#### 4.1 Core State Management

```go
type appModel struct {
    // Core components
    editor  *editor.EditorComponent
    msgView *messages.MessageView
    
    // Config
    config *config.Config
    
    // Agent
    agent *agent.Agent
    
    // Event system
    pubsub *pubsub.PubSub
    
    // Dimensions
    width  int
    height int
    
    // State
    quitting      bool
    processing    bool
    showHelp      bool
    errorMsg      string
    sessionActive bool
    
    // Status bar (inline)
    status      string
    model       string
    tokenUsage  *types.TokenUsage
}
```

#### 4.2 Global Shortcut Handling

Need to add in `handleGlobalKeys`:

```go
// Scrolling shortcuts
case "ctrl+shift+end":
    m.msgView.ScrollToBottom()
case "ctrl+shift+home":
    m.msgView.ScrollToTop()
case "pgup":
    m.msgView.PageUp()
case "pgdown":
    m.msgView.PageDown()
case "ctrl+u":
    m.msgView.HalfPageUp()
case "ctrl+d":
    m.msgView.HalfPageDown()

// Function shortcuts
case tea.KeyF1:
    m.showHelp = !m.showHelp
case tea.KeyCtrlC:
    if m.processing {
        // cancel operation
    } else {
        m.quitting = true
        return tea.Quit
    }
```

#### 4.3 Layout Calculation

Original layout (top to bottom):
```
┌─────────────────────────────┐
│                             │
│     MessageView             │
│   (dynamic height, min 10 lines) │
│                             │
├─────────────────────────────┤
│     EditorComponent         │
│      (fixed 5 lines)          │
├─────────────────────────────┤
│     StatusBar               │
│      (1 line)                  │
└─────────────────────────────┘
```

### Phase 5: Event System Integration

#### 5.1 PubSub Event Subscriptions

Event types to subscribe to:

```go
// Agent response
pubsub.EventTypeAgentResponse     -> append to assistant message
pubsub.EventTypeAgentError        -> show error

// Tool events
pubsub.EventTypeToolStart         -> show tool call
pubsub.EventTypeToolComplete      -> show tool result

// Token stats
pubsub.EventTypeTokenUsage        -> update status bar and message

// Session events
pubsub.EventTypeSessionStart      -> initialize log
pubsub.EventTypeSessionResume     -> restore history
```

#### 5.2 Message Flow

```
user input -> EditorComponent.SendMsg 
         -> pubsub.Publish(EventTypeUserInput)
         -> Agent.ProcessUserInput
         -> pubsub.Publish(EventTypeAgentResponse) [multiple, streaming]
         -> MessageView.AppendToLastMessage
         -> pubsub.Publish(EventTypeTokenUsage)
         -> update status bar
```

### Phase 6: Testing and Verification

#### 6.1 Visual Regression Testing

Compare original and current versions:
- [ ] Color scheme consistency
- [ ] Layout proportions
- [ ] Message format
- [ ] Shortcut behavior

#### 6.2 Functional Testing

- [ ] Send message flow
- [ ] History navigation
- [ ] Attachment add/remove
- [ ] Tool call display
- [ ] Bash command tracking
- [ ] Token usage display
- [ ] Error handling

## 5. File Modification List

### Files to Create/Major Rewrite

| File | Action | Description |
|------|--------|-------------|
| `tui/app/model.go` | Major rewrite | Integrate status bar, improve shortcuts |
| `tui/app/run.go` | Minor | Confirm startup options |
| `tui/components/messages/types.go` | Medium | Ensure message format matches original |
| `tui/components/messages/message_view.go` | Medium | Improve message handling |

### Files to Delete

| File | Description |
|------|-------------|
| `tui/components/chat/chat.go` | Redundant, replaced by MessageView |
| `tui/components/editor/editor.go` | Redundant, replaced by interaction/editor |

### Optional Deletion

| File | Description |
|------|-------------|
| `tui/components/status/status.go` | Can be integrated into appModel, or kept |

## 6. Color Scheme (ANSI Standard Colors)

For compatibility, use standard ANSI color codes:

```go
// Primary colors
colorPrimary   = lipgloss.Color("135")  // purple - main border
colorSecondary = lipgloss.Color("75")   // blue - tools/assistant

// Status colors
colorSuccess = lipgloss.Color("82")     // green - success/raw msg
colorWarning = lipgloss.Color("215")    // orange - warning/connecting
colorError   = lipgloss.Color("203")    // red - error

// Text colors
colorText      = lipgloss.Color("250")  // light gray - main text
colorTextMuted = lipgloss.Color("248")  // gray - secondary text/Token

// Background colors
colorBg    = lipgloss.Color("237")      // dark blue-gray - status bar bg
colorErrorBg = lipgloss.Color("17")     // dark blue - error bg
```

## 7. Recommended Implementation Order

1. **First** clean up redundant files (Phase 1)
2. **Then** verify EditorComponent functionality (Phase 2)
3. **Next** update MessageView message format (Phase 3)
4. **After** improve AppModel layout and shortcuts (Phase 4)
5. **Finally** test event system integration (Phase 5)

## 8. Verification Method

After completing updates, run the following commands to verify:

```bash
# Build
cd decompiled && go build -o qodercli-reverse .

# Run tests
./qodercli-reverse --help
./qodercli-reverse "test message"

# Interactive mode
./qodercli-reverse
```

Checkpoints:
- [x] No compile errors
- [x] TUI starts normally
- [ ] Can send messages and get responses
- [ ] History navigation works
- [ ] Message format correct (raw msg, raw resp)
- [ ] Token usage displays
- [ ] Shortcuts respond correctly

## 9. Completed Work Log

### Phase 1: Clean Up Redundant Code - Completed

- [x] Remove `tui/components/chat/chat.go` (or replace with messages)
- [x] Remove `tui/components/editor/editor.go` (use interaction/editor)

### Phase 2: Update MessageView - Completed

- [x] Ensure all message types correctly implement `Render()` and `String()` interfaces
- [x] Add `HalfPageUp()` and `HalfPageDown()` methods
- [x] Message format keeps original style (raw msg, raw resp)

### Phase 3: Update AppModel - Completed

- [x] Improve global shortcut handling (ctrl+u/d are now half-page scroll)
- [x] Integrate status bar into appModel
- [x] Layout calculation correct

### Phase 4: Build Verification - Completed

- [x] Build successful, no compile errors
- [x] help command works normally
- [x] Non-interactive mode tests pass

## 10. Important Correction Notes (2026-03-09)

### Message Format Correction

**Problem**: Previous implementation incorrectly added "raw msg" and "raw resp" prefix markers

**Correction**: Based on user feedback and further analysis, the original TUI actually:
- User messages: display content directly, no special marker
- Assistant messages: display content directly, no special marker  
- Status bar: uses `●` symbol with color (green means Ready)
- Message view: uses simple text display, supports ANSI colors

**Root cause analysis**:
- The "raw"-related strings in strings output are actually remnants of other functionality (such as `raw_content`, `raw-write`, etc.)
- Comments in decompiled code may have misled the implementation
- Original TUI focuses more on simplicity and practicality

### Issues to Confirm Next

1. **Status bar rendering**: Need to confirm the exact format of the original status bar
2. **Message borders**: Whether borders or backgrounds are needed for messages
3. **Copyability**: Original cannot copy possibly because of special rendering mode
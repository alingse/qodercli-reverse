# TUI View Rendering Optimization

## Problem Background

After implementing turn-based TUI rendering, the View's streaming content display had the following issues:

1. **Content truncation**: Streaming content only shows partial text, with many blank lines
2. **Display anomalies**: Content is force-wrapped, formatting is messy
3. **Screen occupation**: When View content is too long, the editor and status bar are pushed off-screen

## Problem Analysis

### Issues in the Original Implementation

```go
// Problematic code
func (m appModel) renderCurrentTurn() string {
    if m.currentTurn.StreamingText.Len() > 0 {
        preview := &messages.AssistantMessage{
            Content: m.currentTurn.StreamingText.String(),
            MsgTime: time.Now(),
        }

        previewWidth := m.width - 4
        if previewWidth < 76 {
            previewWidth = 76
        }

        previewStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("246")).
            Italic(true).
            Padding(0, 1).
            Width(previewWidth).  // Issue 1: fixed width causes truncation
            Border(lipgloss.NormalBorder(), false, false, false, true).  // Issue 2: border display anomaly
            BorderForeground(lipgloss.Color("240"))

        sections = append(sections, previewStyle.Render(preview.Render()))  // Issue 3: double rendering
    }
}
```

**Root causes**:

1. **Fixed width `Width(previewWidth)`**:
   - Forcing a width causes content to be truncated or force-wrapped
   - When content exceeds the width, lipgloss auto-wraps, which may cause formatting mess

2. **Double rendering**:
   - `preview.Render()` already returns a formatted string (`⏺ content`)
   - Wrapping it again with `previewStyle.Render()` causes formatting issues

3. **Border and padding**:
   - The left border `│` may cause display anomalies
   - Padding adds extra space

4. **No height limit**:
   - Streaming content may be very long, occupying a lot of screen space
   - Causes the editor and status bar to be pushed off-screen

## Optimization Plan

### 1. Remove Fixed Width Limit

```go
// Do not set a fixed width, let content wrap naturally
previewStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("246")).
    Italic(true)
```

### 2. Simplify Rendering Logic

```go
// Render content directly, without double rendering
content := m.currentTurn.StreamingText.String()
sections = append(sections, previewStyle.Render(fmt.Sprintf("⏺ %s", strings.TrimSpace(content))))
```

### 3. Limit Preview Area Height

```go
// Show at most the last 10 lines; use "..." for the rest
lines := strings.Split(content, "\n")
maxLines := 10
if len(lines) > maxLines {
    lines = lines[len(lines)-maxLines:]
    content = "...\n" + strings.Join(lines, "\n")
}
```

### 4. Remove Border and Padding

```go
// Do not use border and padding, making content more compact
previewStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("246")).
    Italic(true)
```

## Optimized Implementation

```go
func (m appModel) renderCurrentTurn() string {
    if m.currentTurn == nil {
        return ""
    }

    var sections []string

    // 1. Render streaming text (if any)
    if m.currentTurn.StreamingText.Len() > 0 {
        content := m.currentTurn.StreamingText.String()

        // Limit preview area height to avoid taking too much screen space
        // Show at most 10 lines; use "..." for the rest
        lines := strings.Split(content, "\n")
        maxLines := 10
        if len(lines) > maxLines {
            lines = lines[len(lines)-maxLines:]
            content = "...\n" + strings.Join(lines, "\n")
        }

        // Use gray italic style to represent content being generated
        previewStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("246")).
            Italic(true)

        // Render content directly, add ⏺ prefix
        sections = append(sections, previewStyle.Render(fmt.Sprintf("⏺ %s", strings.TrimSpace(content))))
    }

    // 2. Render tool calls (if any)
    for _, toolCall := range m.currentTurn.ToolCalls {
        toolMsg := &messages.ToolCallInfo{
            ID:        toolCall.ID,
            Name:      toolCall.Name,
            Arguments: toolCall.Arguments,
            Output:    toolCall.Output,
            IsError:   toolCall.IsError,
            Completed: toolCall.Status == ToolCallStatusCompleted || toolCall.Status == ToolCallStatusError,
            MsgTime:   time.Now(),
        }

        // Use different colors for different states
        var color string
        switch toolCall.Status {
        case ToolCallStatusPending:
            color = "248" // gray
        case ToolCallStatusRunning:
            color = "255" // white
        case ToolCallStatusCompleted:
            color = "82" // green
        case ToolCallStatusError:
            color = "203" // red
        default:
            color = "255" // default white
        }

        style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
        sections = append(sections, style.Render(toolMsg.Render()))
    }

    return strings.Join(sections, "\n")
}
```

## Optimization Results

### Previous Issues

```
│ core
│
│
│ types
│
│ definition
│
│ :
```

- Content was truncated, only partial text shown
- Many blank lines
- Border display anomalies

### After Optimization

```
⏺ Hello! I am qodercli-reverse, an interactive CLI tool that can help you with software engineering tasks.

I see you are in a Go project, current working directory is `/Users/zhihu/output/github/qodercli-reverse/decompiled`.

Is there anything I can help you with?
```

- Content displayed completely
- Formatting correct
- No extra blank lines
- Preview area height limited, won't occupy too much screen space

## Key Improvements

1. **Content completeness**: Removed fixed width limit, content no longer truncated
2. **Correct formatting**: Simplified rendering logic, avoiding formatting issues caused by double rendering
3. **Screen space management**: Limited preview area height, ensuring editor and status bar are always visible
4. **Visual clarity**: Removed border and padding, making content more compact and clear

## Notes

1. **Height limit**: Currently set to show at most 10 lines, can be adjusted according to actual needs
2. **Truncation hint**: Uses "..." to indicate there is more content; users can scroll to view full history
3. **Performance consideration**: Each streaming update triggers a View re-render; throttling may be needed for high-frequency updates

## File Changes

- `tui/app/model.go` - Modified `renderCurrentTurn()` method
  - Removed fixed width setting
  - Simplified rendering logic
  - Added height limit
  - Removed border and padding
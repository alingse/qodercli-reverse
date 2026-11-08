# Bubble Tea WithAltScreen Usage Analysis

## Why use `tea.WithAltScreen()`?

### 1. Alternate Screen Buffer Role

`tea.WithAltScreen()` enables the terminal's **alternate screen buffer**, an independent screen space with the following advantages:

#### ✅ Advantages:
- **Doesn't pollute terminal history**: TUI content doesn't mix into the terminal's scroll history
- **Full screen control**: Can freely use the entire screen space, supporting complex UI layouts
- **Clean enter/exit**: Automatically restores the previous terminal state on exit
- **Better UX**: Behavior similar to vim, less, htop, etc.

#### ❌ Disadvantages:
- **Clears on exit**: By default, all TUI content is cleared on exit
- **No history viewable**: TUI content cannot be reviewed after exit

### 2. Official Implementation Analysis

From string analysis of the official qodercli-reverse binary:

```
AltScreen
ExitAltScreen
EnterAltScreen
altScreenActive
altScreenWasActive
```

The official version indeed uses AltScreen, which is standard practice.

### 3. Practices in Other Projects

#### Glow (Charm's official project)
```go
opts := []tea.ProgramOption{tea.WithAltScreen()}
if cfg.EnableMouse {
    opts = append(opts, tea.WithMouseCellMotion())
}
return tea.NewProgram(m, opts...)
```

✅ Uses `WithAltScreen()`

#### Other TUI Tools
- **vim**: Uses alternate screen
- **less**: Uses alternate screen
- **htop**: Uses alternate screen
- **lazygit**: Uses its own TUI framework (gocui), also uses alternate screen

## How to Preserve Output on Exit?

### Approach 1: Output to stderr (current implementation) ✅

```go
// Output statistics to stderr before exiting
fmt.Fprint(os.Stderr, statsMsg)
return QuitMsg{}
```

**Principle**:
- stderr is not affected by alternate screen
- After exit, the statistics will display on the main screen

**Advantages**:
- Simple and direct
- Doesn't affect normal TUI operation
- Statistics visible after exit

**Disadvantages**:
- Can only output on exit, cannot display within TUI

### Approach 2: Don't Use AltScreen

```go
// Without WithAltScreen
p := tea.NewProgram(model)
```

**Advantages**:
- All output remains in terminal history
- Can scroll and view anytime

**Disadvantages**:
- TUI content pollutes terminal history
- Cannot fully control the screen
- Poorer user experience (unlike traditional TUI tools)

### Approach 3: Hybrid Mode (recommended)

```go
// Use AltScreen within TUI
p := tea.NewProgram(model, tea.WithAltScreen())

// Output important info to stderr on exit
func (m *appModel) handleCommand(input string) tea.Msg {
    case "/quit", "/exit":
        // Output statistics to stderr
        fmt.Fprint(os.Stderr, statsMsg)
        return QuitMsg{}
}
```

**This is best practice**:
- ✅ Full screen control during TUI runtime
- ✅ Important information (statistics) preserved after exit
- ✅ Doesn't pollute terminal history
- ✅ Meets user expectations for TUI tools

### Approach 4: Use tea.Println (provided by Bubble Tea)

```go
// Method provided by Bubble Tea for printing outside AltScreen
return m, tea.Sequence(
    tea.Printf(statsMsg),
    tea.Quit,
)
```

**Note**: This requires a newer version of Bubble Tea.

## Current Implementation

We adopted **Approach 3 (Hybrid Mode)**:

1. **TUI runtime**: Use `tea.WithAltScreen()` for full TUI experience
2. **On exit**: Output statistics to `stderr`, ensuring visibility after exit

```go
// tui/app/model.go
func (m *appModel) handleCommand(input string) tea.Msg {
    case "/quit", "/exit":
        stats := m.agent.GetState().GetStats()
        statsMsg := fmt.Sprintf(...)

        // Output to stderr, visible after exit
        fmt.Fprint(os.Stderr, statsMsg)

        return QuitMsg{}
}
```

## Summary

| Approach | TUI Experience | Visible After Exit | Terminal History | Recommendation |
|----------|---------------|-------------------|------------------|----------------|
| WithAltScreen + stderr | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Clean | ⭐⭐⭐⭐⭐ |
| No AltScreen | ⭐⭐ | ⭐⭐⭐⭐⭐ | Messy | ⭐⭐ |
| tea.Println | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Clean | ⭐⭐⭐⭐ |

**Conclusion**: Using `tea.WithAltScreen()` is the correct choice; combined with stderr output, it perfectly solves the need to preserve information on exit.
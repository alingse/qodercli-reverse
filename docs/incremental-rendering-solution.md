# Incremental Rendering Implementation

## Problem Analysis

The official qodercli-reverse version uses Alibaba's internal fork of Bubble Tea (`code.alibaba-inc.com/qoder-core/bubbletea`), and this fork version implements special scrolling behavior:
- Message history scrolls upward like normal terminal output
- The editor and status bar at the bottom stay fixed
- All history can be viewed via the terminal scroll bar

## Core Technical Details

The official Bubble Tea's rendering model is "clear-and-redraw", which is not suitable for incremental output. To achieve a similar effect, the following is required:

### Approach A: Fork Bubble Tea (Recommended)

1. Fork `github.com/charmbracelet/bubbletea`
2. Modify the renderer to add an "incremental mode":
   - Do not clear previous output
   - Only append new content
   - Use ANSI escape codes to fix the bottom area

Key changes:
```go
// Add the incremental mode in renderer.go
type IncrementalMode bool

func (r *renderer) renderIncremental(view string) {
    // Output only the new content
    // Use ANSI escape codes to fix the bottom area
}
```

### Approach B: Custom Renderer (Current Implementation)

Instead of forking Bubble Tea, do the following:
1. Use `tea.WithoutRenderer()` to disable the default renderer
2. Implement a custom incremental renderer (see `incremental_renderer.go`)
3. Call the renderer manually in `Update()`

Pros:
- No need to maintain a fork
- More flexible

Cons:
- You lose some of Bubble Tea's features (e.g., automatic window size detection)
- Need to handle more details manually

### Approach C: Hybrid Mode

1. Output message history directly via `fmt.Println()`
2. Use ANSI escape codes to control the fixed bottom area
3. Do not use Bubble Tea's full-screen mode

## Implementation Steps

### Using Approach B (Custom Renderer)

1. **Modify `run.go`**:
```go
p := tea.NewProgram(
    model,
    tea.WithoutRenderer(), // disable the default renderer
)
```

2. **Modify `model.go`**:
```go
type appModel struct {
    // ... existing fields
    renderer *IncrementalRenderer
}

func (m appModel) View() string {
    // Do not return the full view; instead output incrementally via the renderer
    return ""
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Call the renderer when needed
    switch msg := msg.(type) {
    case AgentStreamMsg:
        m.renderer.RenderIncrementalMessage(msg.Content)
    }
}
```

3. **Modify `MessageView`**:
```go
func (mv *MessageView) AddMessage(msg Message) {
    // Output directly; do not cache
    fmt.Println(msg.Render())
}
```

## ANSI Escape Code Reference

```bash
\033[s          # save cursor position
\033[u          # restore cursor position
\033[{row};{col}H  # move cursor to the specified position
\033[J          # clear from cursor to bottom of screen
\033[K          # clear current line
\033[{n}A       # move cursor up n lines
\033[{n}B       # move cursor down n lines
```

## Reference Implementations

Similar implementations can be found here:
- The progress bar of `npm`
- The output of `docker build`
- The output of `cargo build`

These tools all implement the "history scrolls up + fixed bottom area" effect.

## Next Steps

1. Decide which approach to use
2. If Approach A is chosen, fork Bubble Tea and implement the incremental mode
3. If Approach B is chosen, complete `incremental_renderer.go` and integrate it into the existing code
4. Test and debug

## Considerations

- ANSI escape codes may behave differently across terminals
- Windows terminals require special handling
- Terminal size changes must be handled correctly
- Handle signals such as Ctrl+C and ensure the terminal state is restored correctly

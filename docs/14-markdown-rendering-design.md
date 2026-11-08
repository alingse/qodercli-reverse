# qodercli-reverse Markdown Rendering Design

## 1. Official Implementation Analysis

### 1.1 Packages Used

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/charmbracelet/glamour` | v0.10.0 | Core markdown terminal rendering library |
| `github.com/yuin/goldmark` | - | Markdown parser (used internally by glamour) |
| `github.com/yuin/goldmark-emoji` | - | Emoji extension support |
| `github.com/catppuccin/go` | v0.3.0 | Catppuccin theme color definitions |

### 1.2 glamour Core API

```go
// Create renderer
renderer, err := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),  // Use built-in theme
    glamour.WithWordWrap(width),         // Set auto-wrap width
    glamour.WithAutoStyle(),             // Auto-detect terminal theme
)

// Render methods
rendered, err := renderer.Render(markdown string)              // Render string
renderedBytes, err := renderer.RenderBytes(markdownBytes)      // Render bytes

// Implements io.Reader/io.Writer interfaces
renderer.Read(p []byte)
renderer.Write(p []byte)
renderer.Close()
```

### 1.3 Officially Supported Theme List

glamour built-in themes:
- `dark` (default dark)
- `light` (light)
- `github-dark` / `github-light` (GitHub style)
- `monokai` / `monokailight` (code editor style)
- `dracula` (dark high contrast)
- `tokyonight-day` / `tokyonight-night` / `tokyonight-moon` / `tokyonight-storm`
- `catppuccin-latte` / `catppuccin-frappe` / `catppuccin-macchiato` / `catppuccin-mocha`
- `solarized-dark` / `solarized-light`
- `gruvbox` / `gruvbox-light`
- `modus-operandi` / `modus-vivendi`
- `rose-pine-dawn` / `rose-pine-moon`
- `evergarden` / `witchhazel` / `xcode-dark`
- `paraiso-dark` / `paraiso-light` / `rainbow_dash`
- `algol_nu` / `colorful` / `doom-one` / `friendly` / `lovelace` / `pygments`

### 1.4 Official Error Handling Pattern

Error handling discovered from decompilation analysis:
```
glamour: error reading from buffer: %w
glamour: error converting markdown: %w
glamour: error rendering: %w
```

## 2. Current Code Status

### 2.1 Existing Implementation

`decompiled/tui/components/messages/message_view.go`:
- glamour package already imported
- `TermRenderer` instance already created
- But not actually used to render message content

```go
// Current: created but not used for message rendering
renderer, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithWordWrap(76),
)
```

### 2.2 Problem Analysis

`decompiled/tui/components/messages/types.go`:

```go
// AssistantMessage currently rendered as plain text
func (m *AssistantMessage) Render() string {
    return fmt.Sprintf("⏺ %s", strings.TrimSpace(m.Content))
}
```

**Problems**:
1. Assistant's markdown content is displayed directly as text
2. Code blocks, lists, tables, etc. lose formatting
3. Links displayed in raw markdown format `[text](url)`

## 3. Design

### 3.1 Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    MessageView                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │           MarkdownRenderer                      │   │
│  │  ┌─────────────┐    ┌─────────────────────┐    │   │
│  │  │   glamour   │───▶│  TermRenderer       │    │   │
│  │  │  - Parse    │    │  - Render to ANSI   │    │   │
│  │  │  - Style    │    │  - Word Wrap        │    │   │
│  │  └─────────────┘    └─────────────────────┘    │   │
│  └─────────────────────────────────────────────────┘   │
│                          │                              │
│                          ▼                              │
│  ┌─────────────────────────────────────────────────┐   │
│  │              Message Types                      │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────┐    │   │
│  │  │ UserMsg  │ │Assistant │ │   ToolCall   │    │   │
│  │  └──────────┘ └──────────┘ └──────────────┘    │   │
│  └─────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 3.2 Core Implementation

#### 3.2.1 Create Markdown Renderer

```go
// core/utils/markdown/renderer.go
package markdown

import (
    "sync"
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/lipgloss"
)

// Renderer markdown terminal renderer
type Renderer struct {
    renderer *glamour.TermRenderer
    mu       sync.RWMutex
    width    int
    style    string
}

// NewRenderer creates a new markdown renderer
func NewRenderer(width int, style string) (*Renderer, error) {
    if style == "" {
        style = "dark"
    }

    r, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle(style),
        glamour.WithWordWrap(width),
    )
    if err != nil {
        return nil, err
    }

    return &Renderer{
        renderer: r,
        width:    width,
        style:    style,
    }, nil
}

// Render renders markdown into terminal-displayable format
func (r *Renderer) Render(content string) (string, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    return r.renderer.Render(content)
}

// SetSize updates renderer dimensions
func (r *Renderer) SetSize(width int) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if r.width == width {
        return nil
    }

    r.width = width
    newRenderer, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle(r.style),
        glamour.WithWordWrap(width),
    )
    if err != nil {
        return err
    }

    r.renderer = newRenderer
    return nil
}

// SetStyle switches theme
func (r *Renderer) SetStyle(style string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if r.style == style {
        return nil
    }

    r.style = style
    newRenderer, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle(style),
        glamour.WithWordWrap(r.width),
    )
    if err != nil {
        return err
    }

    r.renderer = newRenderer
    return nil
}
```

#### 3.2.2 Integration in MessageView

```go
// tui/components/messages/message_view.go

type MessageView struct {
    viewport viewport.Model
    messages []Message
    renderer *markdown.Renderer  // replaces glamour.TermRenderer
    width    int
    height   int
    // ...
}

func NewMessageView() *MessageView {
    vp := viewport.New(80, 20)
    vp.SetContent("")
    vp.MouseWheelEnabled = true

    // Create markdown renderer
    renderer, _ := markdown.NewRenderer(76, "dark")

    return &MessageView{
        viewport: vp,
        messages: make([]Message, 0),
        renderer: renderer,
        autoScroll: true,
    }
}

func (mv *MessageView) SetSize(width, height int) {
    mv.width = width
    mv.height = height

    mv.viewport.Width = width
    mv.viewport.Height = height

    // Update renderer dimensions
    if mv.renderer != nil {
        mv.renderer.SetSize(width - 4) // leave margin
    }

    // Re-render all messages
    mv.renderContent()
}
```

#### 3.2.3 Modify Message Types to Support Markdown

```go
// tui/components/messages/types.go

type AssistantMessage struct {
    Content      string
    Rendered     string  // cached rendered content
    NeedsRender  bool    // whether re-render is needed
    MsgTime      time.Time
}

func (m *AssistantMessage) Type() MessageType    { return MsgTypeAssistant }
func (m *AssistantMessage) Timestamp() time.Time { return m.MsgTime }

// RenderMarkdown renders markdown (called by MessageView)
func (m *AssistantMessage) RenderMarkdown(renderer *markdown.Renderer) string {
    if !m.NeedsRender && m.Rendered != "" {
        return m.Rendered
    }

    if renderer == nil {
        m.Rendered = m.Content
        m.NeedsRender = false
        return m.Rendered
    }

    rendered, err := renderer.Render(m.Content)
    if err != nil {
        // Fall back to plain text on render failure
        m.Rendered = m.Content
    } else {
        m.Rendered = rendered
    }
    m.NeedsRender = false
    return m.Rendered
}

func (m *AssistantMessage) Render() string {
    // Return raw content; actual rendering handled centrally by MessageView
    return m.Content
}

// AppendContent appends content and marks as needing re-render
func (m *AssistantMessage) AppendContent(content string) {
    m.Content += content
    m.NeedsRender = true
}
```

#### 3.2.4 MessageView Unified Rendering Flow

```go
// tui/components/messages/message_view.go

func (mv *MessageView) renderContent() {
    var sb strings.Builder

    for i, msg := range mv.messages {
        if i > 0 {
            sb.WriteString("\n\n")
        }

        switch m := msg.(type) {
        case *AssistantMessage:
            // Render markdown with glamour
            rendered := m.RenderMarkdown(mv.renderer)
            sb.WriteString(rendered)
        case *UserMessage:
            // User messages keep simple format
            sb.WriteString(mv.renderUserMessage(m))
        case *ToolCallInfo:
            // Tool calls use lipgloss styles
            sb.WriteString(m.Render())
        // ... other message types
        default:
            sb.WriteString(msg.Render())
        }
    }

    mv.cachedContent = sb.String()
}

func (mv *MessageView) renderUserMessage(m *UserMessage) string {
    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color("252")).
        Bold(true)

    return style.Render("> ") + strings.TrimSpace(m.Content)
}
```

### 3.3 Config Support

```go
// core/config/config.go

type Config struct {
    // ... other config

    // Markdown rendering config
    Markdown MarkdownConfig `yaml:"markdown" json:"markdown"`
}

type MarkdownConfig struct {
    Enabled   bool   `yaml:"enabled" json:"enabled"`     // Whether enabled
    Style     string `yaml:"style" json:"style"`         // Theme name
    WordWrap  int    `yaml:"word_wrap" json:"word_wrap"` // Auto-wrap width
}

func DefaultMarkdownConfig() MarkdownConfig {
    return MarkdownConfig{
        Enabled:  true,
        Style:    "dark",
        WordWrap: 0, // 0 means auto based on terminal width
    }
}
```

### 3.4 Theme Auto-Detection

```go
// core/utils/markdown/style.go

import "github.com/charmbracelet/lipgloss"

// DetectStyle detects terminal theme and returns appropriate glamour style
func DetectStyle() string {
    if lipgloss.HasDarkBackground() {
        return "dark"
    }
    return "light"
}

// GetAvailableStyles returns all available theme list
func GetAvailableStyles() []string {
    return []string{
        "dark", "light",
        "github-dark", "github-light",
        "monokai", "monokailight",
        "dracula",
        "tokyonight-day", "tokyonight-night", "tokyonight-moon", "tokyonight-storm",
        "catppuccin-latte", "catppuccin-frappe", "catppuccin-macchiato", "catppuccin-mocha",
        "solarized-dark", "solarized-light",
        "gruvbox", "gruvbox-light",
    }
}
```

## 4. Implementation Steps

### Phase 1: Basic Renderer (P0)

1. **Create `core/utils/markdown/renderer.go`**
   - Wrap glamour.TermRenderer
   - Provide Render/SetSize/SetStyle methods
   - Support concurrency safety

2. **Modify `tui/components/messages/message_view.go`**
   - Integrate markdown.Renderer
   - Update renderer dimensions on SetSize
   - Call markdown rendering in renderContent

3. **Modify `tui/components/messages/types.go`**
   - AssistantMessage supports caching render results
   - Add RenderMarkdown method

### Phase 2: Config and Optimization (P1)

4. **Add config support**
   - Add MarkdownConfig to Config
   - Support loading theme settings from config file

5. **Theme auto-detection**
   - Detect terminal background color
   - Auto-select dark/light theme

6. **Performance optimization**
   - Render result caching
   - Incremental rendering (streaming responses)

### Phase 3: Enhanced Features (P2)

7. **Code block syntax highlighting enhancement**
   - Support highlighting for more languages
   - Custom code block styles

8. **Image link handling**
   - Convert image links to clickable text
   - Support terminal image preview (e.g., iTerm2)

9. **Table rendering optimization**
   - Auto-adjust column widths
   - Handle overly wide tables

## 5. Key Code Examples

### 5.1 Complete Rendering Flow

```go
// Initialize
mv := NewMessageView()
mv.SetSize(100, 30)

// Add assistant message (with markdown)
mv.AddAssistantMessage(`
# Hello World

Here's some **bold** text and _italic_ text.

\`\`\`go
func main() {
    fmt.Println("Hello!")
}
\`\`\`

- Item 1
- Item 2
- Item 3
`)

// renderContent internal processing
// 1. AssistantMessage.RenderMarkdown(renderer) is called
// 2. glamour converts markdown to ANSI format
// 3. Result displayed in terminal with syntax highlighting and formatting
```

### 5.2 Streaming Response Handling

```go
func (mv *MessageView) AppendToLastMessage(content string) {
    if len(mv.messages) == 0 {
        mv.AddAssistantMessage(content)
        return
    }

    lastMsg := mv.messages[len(mv.messages)-1]
    if am, ok := lastMsg.(*AssistantMessage); ok {
        am.AppendContent(content)  // marks NeedsRender = true
        mv.renderContent()         // re-render
    }
}
```

## 6. Testing Plan

```go
// core/utils/markdown/renderer_test.go

func TestRenderer_Render(t *testing.T) {
    r, err := NewRenderer(80, "dark")
    require.NoError(t, err)

    input := "# Hello\n\n**bold** text"
    output, err := r.Render(input)
    require.NoError(t, err)

    // Verify output contains ANSI escape sequences
    assert.Contains(t, output, "\x1b[")  // ANSI escape code
}

func TestRenderer_SetSize(t *testing.T) {
    r, _ := NewRenderer(80, "dark")

    err := r.SetSize(120)
    assert.NoError(t, err)
    assert.Equal(t, 120, r.width)
}
```

## 7. Reference Documents

- [glamour GitHub](https://github.com/charmbracelet/glamour)
- [glamour Theme Preview](https://github.com/charmbracelet/glamour/tree/master/styles)
- [goldmark Documentation](https://github.com/yuin/goldmark)
- [lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
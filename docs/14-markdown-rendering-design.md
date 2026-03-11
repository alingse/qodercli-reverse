# qodercli Markdown 渲染方案设计

## 一、官方实现分析

### 1.1 使用的包

| 包名 | 版本 | 用途 |
|------|------|------|
| `github.com/charmbracelet/glamour` | v0.10.0 | 核心 markdown 终端渲染库 |
| `github.com/yuin/goldmark` | - | markdown 解析器（glamour 内部使用） |
| `github.com/yuin/goldmark-emoji` | - | emoji 扩展支持 |
| `github.com/catppuccin/go` | v0.3.0 | Catppuccin 主题颜色定义 |

### 1.2 glamour 核心 API

```go
// 创建渲染器
renderer, err := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),  // 使用内置主题
    glamour.WithWordWrap(width),         // 设置自动换行宽度
    glamour.WithAutoStyle(),             // 自动检测终端主题
)

// 渲染方法
rendered, err := renderer.Render(markdown string)              // 渲染字符串
renderedBytes, err := renderer.RenderBytes(markdownBytes)      // 渲染字节

// 实现了 io.Reader/io.Writer 接口
renderer.Read(p []byte)
renderer.Write(p []byte)
renderer.Close()
```

### 1.3 官方支持的主题列表

 glamour 内置主题：
- `dark` (默认暗色)
- `light` (亮色)
- `github-dark` / `github-light` (GitHub 风格)
- `monokai` / `monokailight` (代码编辑器风格)
- `dracula` (深色高对比)
- `tokyonight-day` / `tokyonight-night` / `tokyonight-moon` / `tokyonight-storm`
- `catppuccin-latte` / `catppuccin-frappe` / `catppuccin-macchiato` / `catppuccin-mocha`
- `solarized-dark` / `solarized-light`
- `gruvbox` / `gruvbox-light`
- `modus-operandi` / `modus-vivendi`
- `rose-pine-dawn` / `rose-pine-moon`
- `evergarden` / `witchhazel` / `xcode-dark`
- `paraiso-dark` / `paraiso-light` / `rainbow_dash`
- `algol_nu` / `colorful` / `doom-one` / `friendly` / `lovelace` / `pygments`

### 1.4 官方错误处理模式

从反编译分析中发现官方的错误处理：
```
glamour: error reading from buffer: %w
glamour: error converting markdown: %w
glamour: error rendering: %w
```

## 二、当前代码现状

### 2.1 已有实现

`decompiled/tui/components/messages/message_view.go`:
- 已导入 glamour 包
- 已创建 `TermRenderer` 实例
- 但未实际使用来渲染消息内容

```go
// 当前：创建但未用于消息渲染
renderer, _ := glamour.NewTermRenderer(
    glamour.WithStandardStyle("dark"),
    glamour.WithWordWrap(76),
)
```

### 2.2 问题分析

`decompiled/tui/components/messages/types.go`:

```go
// AssistantMessage 当前纯文本渲染
func (m *AssistantMessage) Render() string {
    return fmt.Sprintf("⏺ %s", strings.TrimSpace(m.Content))
}
```

**问题**：
1. 助手返回的 markdown 内容被当作文本直接显示
2. 代码块、列表、表格等失去格式化
3. 链接显示为原始 markdown 格式 `[text](url)`

## 三、设计方案

### 3.1 架构设计

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

### 3.2 核心实现

#### 3.2.1 创建 Markdown 渲染器

```go
// core/utils/markdown/renderer.go
package markdown

import (
    "sync"
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/lipgloss"
)

// Renderer markdown 终端渲染器
type Renderer struct {
    renderer *glamour.TermRenderer
    mu       sync.RWMutex
    width    int
    style    string
}

// NewRenderer 创建新的 markdown 渲染器
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

// Render 渲染 markdown 为终端可显示的格式
func (r *Renderer) Render(content string) (string, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    return r.renderer.Render(content)
}

// SetSize 更新渲染器尺寸
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

// SetStyle 切换主题
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

#### 3.2.2 在 MessageView 中集成

```go
// tui/components/messages/message_view.go

type MessageView struct {
    viewport viewport.Model
    messages []Message
    renderer *markdown.Renderer  // 替换 glamour.TermRenderer
    width    int
    height   int
    // ...
}

func NewMessageView() *MessageView {
    vp := viewport.New(80, 20)
    vp.SetContent("")
    vp.MouseWheelEnabled = true
    
    // 创建 markdown 渲染器
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
    
    // 更新渲染器尺寸
    if mv.renderer != nil {
        mv.renderer.SetSize(width - 4) // 留边距
    }
    
    // 重新渲染所有消息
    mv.renderContent()
}
```

#### 3.2.3 修改消息类型支持 Markdown

```go
// tui/components/messages/types.go

type AssistantMessage struct {
    Content      string
    Rendered     string  // 缓存渲染后的内容
    NeedsRender  bool    // 是否需要重新渲染
    MsgTime      time.Time
}

func (m *AssistantMessage) Type() MessageType    { return MsgTypeAssistant }
func (m *AssistantMessage) Timestamp() time.Time { return m.MsgTime }

// RenderMarkdown 渲染 markdown（由 MessageView 调用）
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
        // 渲染失败回退到纯文本
        m.Rendered = m.Content
    } else {
        m.Rendered = rendered
    }
    m.NeedsRender = false
    return m.Rendered
}

func (m *AssistantMessage) Render() string {
    // 返回原始内容，实际渲染由 MessageView 统一处理
    return m.Content
}

// AppendContent 追加内容并标记需要重新渲染
func (m *AssistantMessage) AppendContent(content string) {
    m.Content += content
    m.NeedsRender = true
}
```

#### 3.2.4 MessageView 统一渲染流程

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
            // 使用 glamour 渲染 markdown
            rendered := m.RenderMarkdown(mv.renderer)
            sb.WriteString(rendered)
        case *UserMessage:
            // 用户消息保持简单格式
            sb.WriteString(mv.renderUserMessage(m))
        case *ToolCallInfo:
            // 工具调用使用 lipgloss 样式
            sb.WriteString(m.Render())
        // ... 其他消息类型
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

### 3.3 配置支持

```go
// core/config/config.go

type Config struct {
    // ... 其他配置
    
    // Markdown 渲染配置
    Markdown MarkdownConfig `yaml:"markdown" json:"markdown"`
}

type MarkdownConfig struct {
    Enabled   bool   `yaml:"enabled" json:"enabled"`     // 是否启用
    Style     string `yaml:"style" json:"style"`         // 主题名称
    WordWrap  int    `yaml:"word_wrap" json:"word_wrap"` // 自动换行宽度
}

func DefaultMarkdownConfig() MarkdownConfig {
    return MarkdownConfig{
        Enabled:  true,
        Style:    "dark",
        WordWrap: 0, // 0 表示自动根据终端宽度
    }
}
```

### 3.4 主题自动检测

```go
// core/utils/markdown/style.go

import "github.com/charmbracelet/lipgloss"

// DetectStyle 检测终端主题并返回合适的 glamour 样式
func DetectStyle() string {
    if lipgloss.HasDarkBackground() {
        return "dark"
    }
    return "light"
}

// GetAvailableStyles 返回所有可用的主题列表
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

## 四、实现步骤

### Phase 1: 基础渲染器 (P0)

1. **创建 `core/utils/markdown/renderer.go`**
   - 封装 glamour.TermRenderer
   - 提供 Render/SetSize/SetStyle 方法
   - 支持并发安全

2. **修改 `tui/components/messages/message_view.go`**
   - 集成 markdown.Renderer
   - 在 SetSize 时更新渲染器尺寸
   - 在 renderContent 中调用 markdown 渲染

3. **修改 `tui/components/messages/types.go`**
   - AssistantMessage 支持缓存渲染结果
   - 添加 RenderMarkdown 方法

### Phase 2: 配置与优化 (P1)

4. **添加配置支持**
   - 在 Config 中添加 MarkdownConfig
   - 支持从配置文件加载主题设置

5. **主题自动检测**
   - 检测终端背景色
   - 自动选择 dark/light 主题

6. **性能优化**
   - 渲染结果缓存
   - 增量渲染（流式响应）

### Phase 3: 增强功能 (P2)

7. **代码块语法高亮增强**
   - 支持更多语言的高亮
   - 自定义代码块样式

8. **图片链接处理**
   - 将图片链接转为可点击的文本
   - 支持终端图片预览（如 iTerm2）

9. **表格渲染优化**
   - 自动调整列宽
   - 处理过宽表格

## 五、关键代码示例

### 5.1 完整渲染流程

```go
// 初始化
mv := NewMessageView()
mv.SetSize(100, 30)

// 添加助手消息（包含 markdown）
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

// renderContent 内部处理
// 1. AssistantMessage.RenderMarkdown(renderer) 被调用
// 2. glamour 将 markdown 转为 ANSI 格式
// 3. 结果显示在终端，带有语法高亮和格式
```

### 5.2 流式响应处理

```go
func (mv *MessageView) AppendToLastMessage(content string) {
    if len(mv.messages) == 0 {
        mv.AddAssistantMessage(content)
        return
    }

    lastMsg := mv.messages[len(mv.messages)-1]
    if am, ok := lastMsg.(*AssistantMessage); ok {
        am.AppendContent(content)  // 标记 NeedsRender = true
        mv.renderContent()         // 重新渲染
    }
}
```

## 六、测试方案

```go
// core/utils/markdown/renderer_test.go

func TestRenderer_Render(t *testing.T) {
    r, err := NewRenderer(80, "dark")
    require.NoError(t, err)
    
    input := "# Hello\n\n**bold** text"
    output, err := r.Render(input)
    require.NoError(t, err)
    
    // 验证输出包含 ANSI 转义序列
    assert.Contains(t, output, "\x1b[")  // ANSI 转义码
}

func TestRenderer_SetSize(t *testing.T) {
    r, _ := NewRenderer(80, "dark")
    
    err := r.SetSize(120)
    assert.NoError(t, err)
    assert.Equal(t, 120, r.width)
}
```

## 七、参考文档

- [glamour GitHub](https://github.com/charmbracelet/glamour)
- [glamour 主题预览](https://github.com/charmbracelet/glamour/tree/master/styles)
- [goldmark 文档](https://github.com/yuin/goldmark)
- [lipgloss 文档](https://github.com/charmbracelet/lipgloss)

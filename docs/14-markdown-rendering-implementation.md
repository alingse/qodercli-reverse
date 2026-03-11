# Markdown 渲染实现文档

## 实现状态：✅ 已完成

## 一、实现文件清单

| 文件 | 说明 | 状态 |
|------|------|------|
| `core/utils/markdown/renderer.go` | Markdown 渲染器封装 | ✅ |
| `core/utils/markdown/renderer_test.go` | 单元测试 | ✅ |
| `tui/components/messages/message_view.go` | 集成渲染器 | ✅ |
| `tui/components/messages/types.go` | AssistantMessage 支持 | ✅ |
| `core/config/config.go` | Markdown 配置 | ✅ |

## 二、API 使用说明

### 2.1 基础使用

```go
import "github.com/alingse/qodercli-reverse/decompiled/core/utils/markdown"

// 创建渲染器（自动检测终端主题）
renderer, err := markdown.NewRenderer(80, "")
if err != nil {
    log.Fatal(err)
}

// 渲染 markdown
output, err := renderer.Render("# Hello\n\n**bold** text")
fmt.Println(output)
```

### 2.2 在 MessageView 中自动使用

```go
// 创建 MessageView（自动创建 markdown 渲染器）
mv := NewMessageView()

// 添加助手消息（自动渲染 markdown）
mv.AddAssistantMessage(`
# Code Example

\`\`\`go
func main() {
    fmt.Println("Hello!")
}
\`\`\`
`)

// 渲染后的内容会自动显示格式化后的文本
```

### 2.3 配置主题

```go
// 设置特定主题
renderer.SetStyle("dracula")

// 或创建时指定
renderer, _ := markdown.NewRenderer(80, "light")

// 可用主题（glamour 内置）
// - dark (默认)
// - light
// - dracula
// - ascii
// - notty
```

## 三、环境变量配置

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `QODER_MARKDOWN_DISABLED` | 设置为 `true` 禁用 markdown 渲染 | `false` |
| `QODER_MARKDOWN_STYLE` | 设置主题 | 自动检测 |

## 四、实际实现 vs 设计方案

### 差异说明

1. **可用主题**
   - 设计：列出 30+ 个主题
   - 实际：glamour 内置仅 5 个主题（dark/light/dracula/ascii/notty）
   - 原因： glamour 的标准发行版只包含基础主题，其他主题需要额外加载

2. **主题自动检测**
   - 已实现：通过 `lipgloss.HasDarkBackground()` 自动检测

3. **缓存机制**
   - 已实现：AssistantMessage 中缓存渲染结果
   - 优化：追加内容时只标记 NeedsRender，不立即重新渲染

4. **并发安全**
   - 已实现：Renderer 使用 RWMutex 保护内部状态

## 五、测试覆盖

```bash
cd decompiled
go test ./core/utils/markdown/... -v
```

测试项目：
- ✅ NewRenderer 创建
- ✅ Render 渲染各种 markdown 元素
- ✅ SetSize 动态调整宽度
- ✅ SetStyle 切换主题
- ✅ DetectStyle 自动检测
- ✅ 主题列表验证
- ✅ Nil 安全处理

## 六、示例输出

输入：
```markdown
# Hello World

This is **bold** and _italic_ text.

- Item 1
- Item 2
- Item 3
```

输出（带 ANSI 颜色代码的终端格式化文本）：
```
┌─────────────────────────────┐
│  Hello World                │  ← 标题样式
│                             │
│  This is bold and italic    │  ← 粗体/斜体
│  text.                      │
│                             │
│  • Item 1                   │  ← 列表
│  • Item 2                   │
│  • Item 3                   │
└─────────────────────────────┘
```

## 七、注意事项

1. **性能**：首次渲染会有轻微延迟（解析 markdown），后续使用缓存
2. **宽度调整**：终端大小变化时会重新渲染所有消息
3. **流式输出**：支持追加内容，自动标记需要重新渲染
4. **主题限制**：如需更多主题，需通过 `glamour.WithStylesFromJSONFile()` 加载自定义主题

## 八、后续优化方向

1. **自定义主题加载**：支持从配置文件加载官方二进制中的主题（如 catppuccin、tokyonight）
2. **图片链接处理**：将 `![alt](url)` 转为可点击文本
3. **代码块复制**：添加复制代码块的功能按钮
4. **性能优化**：大文档的增量渲染

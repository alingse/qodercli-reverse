# TUI View 渲染优化

## 问题背景

在实现基于回合的 TUI 渲染后，发现 View 的流式内容显示存在以下问题：

1. **内容截断**：流式内容只显示部分文字，出现很多空行
2. **显示异常**：内容被强制换行，格式混乱
3. **屏幕占用**：View 内容太长时，编辑器和状态栏被推到屏幕外

## 问题分析

### 原始实现的问题

```go
// 问题代码
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
            Width(previewWidth).  // 问题1: 固定宽度导致截断
            Border(lipgloss.NormalBorder(), false, false, false, true).  // 问题2: 边框显示异常
            BorderForeground(lipgloss.Color("240"))

        sections = append(sections, previewStyle.Render(preview.Render()))  // 问题3: 双重渲染
    }
}
```

**问题原因**：

1. **固定宽度 `Width(previewWidth)`**：
   - 强制设置宽度会导致内容被截断或强制换行
   - 当内容超出宽度时，lipgloss 会自动换行，但可能导致格式混乱

2. **双重渲染**：
   - `preview.Render()` 已经返回格式化的字符串（`⏺ 内容`）
   - 再用 `previewStyle.Render()` 包装，导致格式问题

3. **边框和 padding**：
   - 左边框 `│` 可能导致显示异常
   - padding 增加了额外的空间占用

4. **无高度限制**：
   - 流式内容可能很长，占用大量屏幕空间
   - 导致编辑器和状态栏被推到屏幕外

## 优化方案

### 1. 移除固定宽度限制

```go
// 不设置固定宽度，让内容自然换行
previewStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("246")).
    Italic(true)
```

### 2. 简化渲染逻辑

```go
// 直接渲染内容，不使用双重渲染
content := m.currentTurn.StreamingText.String()
sections = append(sections, previewStyle.Render(fmt.Sprintf("⏺ %s", strings.TrimSpace(content))))
```

### 3. 限制预览区高度

```go
// 最多显示最后 10 行，超出部分用 "..." 表示
lines := strings.Split(content, "\n")
maxLines := 10
if len(lines) > maxLines {
    lines = lines[len(lines)-maxLines:]
    content = "...\n" + strings.Join(lines, "\n")
}
```

### 4. 移除边框和 padding

```go
// 不使用边框和 padding，让内容更紧凑
previewStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("246")).
    Italic(true)
```

## 优化后的实现

```go
func (m appModel) renderCurrentTurn() string {
    if m.currentTurn == nil {
        return ""
    }

    var sections []string

    // 1. 渲染流式文本（如果有）
    if m.currentTurn.StreamingText.Len() > 0 {
        content := m.currentTurn.StreamingText.String()

        // 限制预览区的高度，避免占用太多屏幕空间
        // 最多显示 10 行，超出部分用 "..." 表示
        lines := strings.Split(content, "\n")
        maxLines := 10
        if len(lines) > maxLines {
            lines = lines[len(lines)-maxLines:]
            content = "...\n" + strings.Join(lines, "\n")
        }

        // 使用灰色斜体样式表示正在生成的内容
        previewStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("246")).
            Italic(true)

        // 直接渲染内容，添加 ⏺ 前缀
        sections = append(sections, previewStyle.Render(fmt.Sprintf("⏺ %s", strings.TrimSpace(content))))
    }

    // 2. 渲染工具调用（如果有）
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

        // 使用不同的颜色表示不同状态
        var color string
        switch toolCall.Status {
        case ToolCallStatusPending:
            color = "248" // 灰色
        case ToolCallStatusRunning:
            color = "255" // 白色
        case ToolCallStatusCompleted:
            color = "82" // 绿色
        case ToolCallStatusError:
            color = "203" // 红色
        default:
            color = "255" // 默认白色
        }

        style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
        sections = append(sections, style.Render(toolMsg.Render()))
    }

    return strings.Join(sections, "\n")
}
```

## 优化效果

### 之前的问题

```
│ 核心
│
│
│ 类型
│
│ 定义
│
│ ：
```

- 内容被截断，只显示部分文字
- 大量空行
- 边框显示异常

### 优化后的效果

```
⏺ 你好！我是 qodercli，一个交互式 CLI 工具，可以帮助您处理软件工程任务。

我看到您在一个 Go 项目中，当前工作目录是 `/Users/zhihu/output/github/qodercli-reverse/decompiled`。

有什么我可以帮助您的吗？
```

- 内容完整显示
- 格式正确
- 无多余空行
- 预览区高度受限，不会占用过多屏幕空间

## 关键改进点

1. **内容完整性**：移除固定宽度限制，内容不再被截断
2. **格式正确**：简化渲染逻辑，避免双重渲染导致的格式问题
3. **屏幕空间管理**：限制预览区高度，确保编辑器和状态栏始终可见
4. **视觉清晰**：移除边框和 padding，让内容更紧凑清晰

## 注意事项

1. **高度限制**：当前设置为最多显示 10 行，可以根据实际需求调整
2. **截断提示**：使用 "..." 表示有更多内容，用户可以通过滚动查看完整历史
3. **性能考虑**：每次流式更新都会触发 View 重新渲染，对于高频更新可能需要添加节流

## 文件变更

- `tui/app/model.go` - 修改 `renderCurrentTurn()` 方法
  - 移除固定宽度设置
  - 简化渲染逻辑
  - 添加高度限制
  - 移除边框和 padding

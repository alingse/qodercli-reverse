# 增量渲染实现方案

## 问题分析

qodercli 官方版本使用了阿里内部 fork 的 Bubble Tea (`code.alibaba-inc.com/qoder-core/bubbletea`)，这个 fork 版本实现了特殊的滚动行为：
- 消息历史像普通终端输出一样向上滚动
- 底部的编辑器和状态栏保持固定
- 可以用终端滚动条查看所有历史

## 核心技术

官方 Bubble Tea 的渲染模型是"清除-重绘"，不适合增量输出。要实现类似效果，需要：

### 方案 A：Fork Bubble Tea（推荐）

1. Fork `github.com/charmbracelet/bubbletea`
2. 修改渲染器，添加"增量模式"：
   - 不清除之前的输出
   - 只追加新内容
   - 使用 ANSI 转义码固定底部区域

关键修改点：
```go
// 在 renderer.go 中添加增量模式
type IncrementalMode bool

func (r *renderer) renderIncremental(view string) {
    // 只输出新增的内容
    // 使用 ANSI 转义码固定底部区域
}
```

### 方案 B：自定义渲染器（当前实现）

不 fork Bubble Tea，而是：
1. 使用 `tea.WithoutRenderer()` 禁用默认渲染
2. 实现自定义的增量渲染器（见 `incremental_renderer.go`）
3. 在 `Update()` 中手动调用渲染

优点：
- 不需要维护 fork
- 更灵活

缺点：
- 失去 Bubble Tea 的一些特性（如自动窗口大小检测）
- 需要手动处理更多细节

### 方案 C：混合模式

1. 消息历史直接用 `fmt.Println()` 输出
2. 底部固定区域用 ANSI 转义码控制
3. 不使用 Bubble Tea 的全屏模式

## 实现步骤

### 使用方案 B（自定义渲染器）

1. **修改 `run.go`**：
```go
p := tea.NewProgram(
    model,
    tea.WithoutRenderer(), // 禁用默认渲染
)
```

2. **修改 `model.go`**：
```go
type appModel struct {
    // ... 现有字段
    renderer *IncrementalRenderer
}

func (m appModel) View() string {
    // 不返回完整视图，而是通过 renderer 增量输出
    return ""
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 在需要时调用 renderer
    switch msg := msg.(type) {
    case AgentStreamMsg:
        m.renderer.RenderIncrementalMessage(msg.Content)
    }
}
```

3. **修改 `MessageView`**：
```go
func (mv *MessageView) AddMessage(msg Message) {
    // 直接输出，不缓存
    fmt.Println(msg.Render())
}
```

## ANSI 转义码参考

```bash
\033[s          # 保存光标位置
\033[u          # 恢复光标位置
\033[{row};{col}H  # 移动光标到指定位置
\033[J          # 清除从光标到屏幕底部
\033[K          # 清除当前行
\033[{n}A       # 光标上移 n 行
\033[{n}B       # 光标下移 n 行
```

## 参考实现

类似的实现可以参考：
- `npm` 的进度条
- `docker build` 的输出
- `cargo build` 的输出

这些工具都实现了"历史向上滚动 + 底部固定区域"的效果。

## 下一步

1. 决定使用哪个方案
2. 如果选择方案 A，需要 fork Bubble Tea 并实现增量模式
3. 如果选择方案 B，需要完善 `incremental_renderer.go` 并集成到现有代码
4. 测试和调试

## 注意事项

- ANSI 转义码在不同终端中的行为可能不同
- Windows 终端需要特殊处理
- 需要正确处理终端大小变化
- 需要处理 Ctrl+C 等信号，确保终端状态正确恢复

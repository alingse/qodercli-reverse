# Bubble Tea WithAltScreen 使用分析

## 为什么使用 `tea.WithAltScreen()`？

### 1. Alternate Screen Buffer 的作用

`tea.WithAltScreen()` 启用终端的 **alternate screen buffer**（备用屏幕缓冲区），这是一个独立的屏幕空间，具有以下优点：

#### ✅ 优点：
- **不污染终端历史**：TUI 运行时的内容不会混入终端的滚动历史
- **完整的屏幕控制**：可以自由使用整个屏幕空间，支持复杂的 UI 布局
- **清晰的进入/退出**：退出后自动恢复到之前的终端状态
- **更好的用户体验**：类似 vim、less、htop 等工具的行为

#### ❌ 缺点：
- **退出时清屏**：默认情况下，退出时会清除所有 TUI 内容
- **无法查看历史**：TUI 中的内容退出后无法回看

### 2. 官方实现分析

从 qodercli 官方二进制的字符串分析：
```
AltScreen
ExitAltScreen
EnterAltScreen
altScreenActive
altScreenWasActive
```

官方确实使用了 AltScreen，这是标准做法。

### 3. 其他项目的实践

#### Glow (Charm 官方项目)
```go
opts := []tea.ProgramOption{tea.WithAltScreen()}
if cfg.EnableMouse {
    opts = append(opts, tea.WithMouseCellMotion())
}
return tea.NewProgram(m, opts...)
```

✅ 使用 `WithAltScreen()`

#### 其他 TUI 工具
- **vim**: 使用 alternate screen
- **less**: 使用 alternate screen
- **htop**: 使用 alternate screen
- **lazygit**: 使用自己的 TUI 框架（gocui），也使用 alternate screen

## 如何在退出时保留输出？

### 方案 1：输出到 stderr（当前实现）✅

```go
// 在退出前输出统计信息到 stderr
fmt.Fprint(os.Stderr, statsMsg)
return QuitMsg{}
```

**原理**：
- stderr 不受 alternate screen 影响
- 退出后，统计信息会显示在主屏幕上

**优点**：
- 简单直接
- 不影响 TUI 的正常运行
- 统计信息在退出后可见

**缺点**：
- 只能在退出时输出，无法在 TUI 中显示

### 方案 2：不使用 AltScreen

```go
// 不使用 WithAltScreen
p := tea.NewProgram(model)
```

**优点**：
- 所有输出都保留在终端历史中
- 可以随时滚动查看

**缺点**：
- TUI 内容会污染终端历史
- 无法完全控制屏幕
- 用户体验较差（不像传统 TUI 工具）

### 方案 3：混合模式（推荐）

```go
// 在 TUI 中使用 AltScreen
p := tea.NewProgram(model, tea.WithAltScreen())

// 退出时输出重要信息到 stderr
func (m *appModel) handleCommand(input string) tea.Msg {
    case "/quit", "/exit":
        // 输出统计信息到 stderr
        fmt.Fprint(os.Stderr, statsMsg)
        return QuitMsg{}
}
```

**这是最佳实践**：
- ✅ TUI 运行时有完整的屏幕控制
- ✅ 退出时保留重要信息（统计数据）
- ✅ 不污染终端历史
- ✅ 符合用户对 TUI 工具的预期

### 方案 4：使用 tea.Println（Bubble Tea 提供）

```go
// Bubble Tea 提供的方法，可以在 AltScreen 外打印
return m, tea.Sequence(
    tea.Printf(statsMsg),
    tea.Quit,
)
```

**注意**：这需要 Bubble Tea 较新版本支持。

## 当前实现

我们采用了 **方案 3（混合模式）**：

1. **TUI 运行时**：使用 `tea.WithAltScreen()` 提供完整的 TUI 体验
2. **退出时**：将统计信息输出到 `stderr`，确保退出后可见

```go
// tui/app/model.go
func (m *appModel) handleCommand(input string) tea.Msg {
    case "/quit", "/exit":
        stats := m.agent.GetState().GetStats()
        statsMsg := fmt.Sprintf(...)

        // 输出到 stderr，退出后可见
        fmt.Fprint(os.Stderr, statsMsg)

        return QuitMsg{}
}
```

## 总结

| 方案 | TUI 体验 | 退出后可见 | 终端历史 | 推荐度 |
|------|---------|-----------|---------|--------|
| WithAltScreen + stderr | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 干净 | ⭐⭐⭐⭐⭐ |
| 不用 AltScreen | ⭐⭐ | ⭐⭐⭐⭐⭐ | 混乱 | ⭐⭐ |
| tea.Println | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 干净 | ⭐⭐⭐⭐ |

**结论**：使用 `tea.WithAltScreen()` 是正确的选择，配合 stderr 输出可以完美解决退出时保留信息的需求。

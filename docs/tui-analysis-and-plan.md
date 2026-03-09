# TUI 架构分析与更新计划

## 1. 分析概述

本文档基于对原版 qodercli 二进制文件的 `which qodercli + strings` 分析，以及当前 decompiled 代码的对比，旨在 1:1 还原原版 TUI 架构。

## 2. 核心发现

### 2.1 原版使用的关键库（从 strings 提取）

```
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss
github.com/charmbracelet/glamour
code.alibaba-inc.com/qoder-core/bubbletea  (定制版本)
```

### 2.2 原版 TUI 组件结构

根据符号和字符串分析，原版包含以下核心组件：

```
tui/
├── app/
│   ├── model.go         # appModel - 主应用模型
│   └── run.go           # Run, RunWithInput, RunSDKMode
├── components/
│   ├── chat/             # 聊天视图（可能未使用）
│   ├── editor/
│   │   └── interaction/
│   │       ├── editor.go    # EditorComponent - 核心编辑器
│   │       ├── attachment.go # AttachmentHandler
│   │       ├── history.go   # HistoryHandler  
│   │       └── cache.go     # InputCache
│   ├── messages/
│   │   ├── message_view.go  # MessageView
│   │   └── types.go         # 消息类型定义
│   └── status/           # 状态栏（可能被整合）
```

### 2.3 关键消息格式（从 strings 提取）

```
"raw msg"        - 用户消息前缀（绿色）
"raw resp"       - 助手消息前缀（蓝色/青色）
"Input token usage: %d | Output token usage: %d" - Token 统计
"Ask anything..." - 编辑器占位符
"Attachments: "   - 附件列表前缀
```

## 3. 当前 decompiled 代码问题

### 3.1 架构问题

1. **组件分离不清晰**: 当前 code 中存在两个 editor 包：
   - `tui/components/editor/editor.go` - 简单版本
   - `tui/components/interaction/editor/editor.go` - 完整版（EditorComponent）
   
   原版应该只使用 `interaction/editor` 版本

2. **Chat 组件冗余**: `tui/components/chat/chat.go` 可能是早期版本，实际应使用 `messages.MessageView`

3. **Status 组件独立**: 当前有独立的 status 组件，但原版可能将其整合到 appModel 中

### 3.2 缺失的功能

根据原版符号分析，缺失以下功能：

1. **完整的键盘快捷键处理**:
   - `alt+enter` - 可能用于特殊功能
   - `ctrl+shift+end/home` - 滚动到顶部/底部
   - `pgup/pgdown` - 页面滚动
   - `ctrl+u/d` - 半页滚动

2. **鼠标支持**:
   - 滚轮事件处理
   - 点击聚焦

3. **完整的消息类型**:
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

4. **Token 统计显示**: 在消息末尾显示 `Input token usage: xxx | Output token usage: xxx`

5. **Bash 命令追踪**: 带有状态指示器（◐, ✓, ✗）的 Bash 命令显示

## 4. 更新计划

### Phase 1: 清理冗余代码

#### 1.1 移除或标记弃用的组件

- [ ] 移除 `tui/components/chat/chat.go` (或使用 messages 替代)
- [ ] 移除 `tui/components/editor/editor.go` (使用 interaction/editor)
- [ ] 考虑将 `tui/components/status/status.go` 整合到 appModel

#### 1.2 统一组件接口

- [ ] 确保所有组件实现 `Init() tea.Cmd`, `Update(msg tea.Msg)`, `View()` 接口
- [ ] 统一消息类型定义

### Phase 2: 更新 EditorComponent

#### 2.1 核心功能（已存在，需验证）

- [x] textarea 集成
- [x] 发送消息逻辑
- [x] 历史记录导航
- [x] 附件处理
- [ ] 输入缓存

#### 2.2 需要添加的功能

- [ ] 完整的键盘快捷键映射表
- [ ] Vim 模式支持（可选）
- [ ] 更完善的错误处理

### Phase 3: 更新 MessageView

#### 3.1 消息类型定义（参考 types.go）

确保所有消息类型正确实现：

```go
type Message interface {
   Type() MessageType
    Render() string
    Timestamp() time.Time
}
```

#### 3.2 消息渲染格式

**用户消息 (UserMessage)**:
```
raw msg          <- 绿色加粗
<内容>
```

**助手消息 (AssistantMessage)**:
```
raw resp         <- 蓝色加粗
<内容>
```

**Token 统计 (TokenUsageMessage)**:
```
Input token usage: 100 | Output token usage: 50  <- 灰色斜体
```

**Bash 命令 (BashInfo)**:
```
◐ Bash $ <command>      <- 执行中（橙色）
✓ Bash $ <command>      <- 成功（绿色）
✗ Bash $ <command>      <- 失败（红色）
<output>                <- 截断超过 500 字符
```

**工具调用 (ToolCall)**:
```
▶ Tool: <name>          <- 蓝色加粗
<arguments>             <- 灰色
```

**工具结果 (ToolResult)**:
```
✓ <name>: <result>      <- 成功（绿色）
✗ <name>: <error>       <- 失败（红色）
```

### Phase 4: 更新 AppModel

#### 4.1 核心状态管理

```go
type appModel struct {
    // 核心组件
    editor  *editor.EditorComponent
    msgView *messages.MessageView
    
    // 配置
    config *config.Config
    
    // Agent
    agent *agent.Agent
    
    // 事件系统
    pubsub *pubsub.PubSub
    
    // 尺寸
    width  int
    height int
    
    // 状态
    quitting      bool
    processing    bool
    showHelp      bool
    errorMsg      string
    sessionActive bool
    
    // 状态栏（内联）
    status      string
    model       string
    tokenUsage  *types.TokenUsage
}
```

#### 4.2 全局快捷键处理

需要在 `handleGlobalKeys` 中添加：

```go
// 滚动快捷键
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

// 功能快捷键
case tea.KeyF1:
    m.showHelp = !m.showHelp
case tea.KeyCtrlC:
    if m.processing {
        // 取消操作
    } else {
        m.quitting = true
        return tea.Quit
    }
```

#### 4.3 布局计算

原版布局（从上到下）：
```
┌─────────────────────────────┐
│                             │
│     MessageView             │
│   (动态高度，至少 10 行)        │
│                             │
├─────────────────────────────┤
│     EditorComponent         │
│      (固定 5 行)              │
├─────────────────────────────┤
│     StatusBar               │
│      (1 行)                  │
└─────────────────────────────┘
```

### Phase 5: 事件系统集成

#### 5.1 PubSub 事件订阅

需要订阅的事件类型：

```go
// Agent 响应
pubsub.EventTypeAgentResponse     -> 追加到助手消息
pubsub.EventTypeAgentError        -> 显示错误

// 工具事件
pubsub.EventTypeToolStart         -> 显示工具调用
pubsub.EventTypeToolComplete      -> 显示工具结果

// Token 统计
pubsub.EventTypeTokenUsage        -> 更新状态栏和消息

// 会话事件
pubsub.EventTypeSessionStart      -> 初始化日志
pubsub.EventTypeSessionResume     -> 恢复历史
```

#### 5.2 消息流

```
用户输入 -> EditorComponent.SendMsg 
         -> pubsub.Publish(EventTypeUserInput)
         -> Agent.ProcessUserInput
         -> pubsub.Publish(EventTypeAgentResponse) [多次，流式]
         -> MessageView.AppendToLastMessage
         -> pubsub.Publish(EventTypeTokenUsage)
         -> 更新状态栏
```

### Phase 6: 测试与验证

#### 6.1 视觉回归测试

对比原版和当前版本的：
- [ ] 颜色方案一致性
- [ ] 布局比例
- [ ] 消息格式
- [ ] 快捷键行为

#### 6.2 功能测试

- [ ] 发送消息流程
- [ ] 历史记录导航
- [ ] 附件添加/删除
- [ ] 工具调用显示
- [ ] Bash 命令追踪
- [ ] Token 统计显示
- [ ] 错误处理

## 5. 文件修改清单

### 需要创建/大改的文件

| 文件 | 操作 | 说明 |
|------|------|------|
| `tui/app/model.go` | 大改 | 整合状态栏，完善快捷键 |
| `tui/app/run.go` | 小改 | 确认启动选项 |
| `tui/components/messages/types.go` | 中改 | 确保消息格式匹配原版 |
| `tui/components/messages/message_view.go` | 中改 | 完善消息处理 |

### 需要删除的文件

| 文件 | 说明 |
|------|------|
| `tui/components/chat/chat.go` | 冗余，使用 MessageView 替代 |
| `tui/components/editor/editor.go` | 冗余，使用 interaction/editor 替代 |

### 可选删除

| 文件 | 说明 |
|------|------|
| `tui/components/status/status.go` | 可整合到 appModel，也可保留 |

## 6. 颜色方案（ANSI 标准色）

为确保兼容性，使用标准 ANSI 颜色码：

```go
// 主要颜色
colorPrimary   = lipgloss.Color("135")  // 紫色 - 主边框
colorSecondary = lipgloss.Color("75")   // 蓝色 - 工具/助手

// 状态颜色
colorSuccess = lipgloss.Color("82")     // 绿色 - 成功/raw msg
colorWarning = lipgloss.Color("215")    // 橙色 - 警告/连接中
colorError   = lipgloss.Color("203")    // 红色 - 错误

// 文本颜色
colorText      = lipgloss.Color("250")  // 浅灰 - 主要文本
colorTextMuted = lipgloss.Color("248")  // 灰色 - 次要文本/Token

// 背景颜色
colorBg    = lipgloss.Color("237")      // 深蓝灰 - 状态栏背景
colorErrorBg = lipgloss.Color("17")     // 深蓝 - 错误背景
```

## 7. 实施顺序建议

1. **首先** 清理冗余文件（Phase 1）
2. **然后** 验证 EditorComponent 功能（Phase 2）
3. **接着** 更新 MessageView 消息格式（Phase 3）
4. **之后** 完善 AppModel 布局和快捷键（Phase 4）
5. **最后** 测试事件系统集成（Phase 5）

## 8. 验证方法

完成更新后，运行以下命令验证：

```bash
# 构建
cd decompiled && go build -o qodercli-reverse .

# 运行测试
./qodercli-reverse --help
./qodercli-reverse "test message"

# 交互模式
./qodercli-reverse
```

检查点：
- [x] 无编译错误
- [x] TUI 正常启动
- [ ] 可以发送消息并获得响应
- [ ] 历史记录导航工作
- [ ] 消息格式正确（raw msg, raw resp）
- [ ] Token 统计显示
- [ ] 快捷键响应正确

## 9. 完成的工作记录

### Phase 1: 清理冗余代码 - 完成

- [x] 移除 `tui/components/chat/chat.go` (或使用 messages 替代)
- [x] 移除 `tui/components/editor/editor.go` (使用 interaction/editor)

### Phase 2: 更新 MessageView - 完成

- [x] 确保所有消息类型正确实现 `Render()` 和 `String()` 接口
- [x] 添加 `HalfPageUp()` 和 `HalfPageDown()` 方法
- [x] 消息格式保持原版风格（raw msg, raw resp）

### Phase 3: 更新 AppModel - 完成

- [x] 完善全局快捷键处理（ctrl+u/d 现在是半页滚动）
- [x] 状态栏整合到 appModel 中
- [x] 布局计算正确

### Phase 4: 构建验证 - 完成

- [x] 构建成功，无编译错误
- [x] help 命令正常工作
- [x] 非交互模式测试通过

## 10. 重要修正说明 (2026-03-09)

### 消息格式修正

**问题**：之前的实现错误地添加了 "raw msg" 和 "raw resp" 前缀标记

**修正**：根据用户反馈和进一步分析，原版 TUI 实际上：
- 用户消息：直接显示内容，无特殊标记
- 助手消息：直接显示内容，无特殊标记  
- 状态栏：使用 `●` 符号配合颜色（绿色表示 Ready）
- 消息视图：使用简单的文本显示，支持 ANSI 颜色

**原因分析**：
- strings 输出中的 "raw" 相关字符串实际上是其他功能的残留（如 `raw_content`, `raw-write` 等）
- 反编译代码中的注释可能误导了实现
- 原版 TUI 更注重简洁性和实用性

### 下一步需要确认的问题

1. **状态栏渲染**：需要确认原版状态栏的具体格式
2. **消息边框**：是否需要为消息添加边框或背景
3. **可复制性**：原版无法复制可能是因为使用了特殊的渲染模式

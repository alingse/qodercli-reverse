# TUI 和日志系统修复总结

## 修复内容

### 1. 新增日志系统 (`core/log/log.go`)

创建了一个完整的日志系统，支持：
- **多级别日志**：DEBUG, INFO, WARN, ERROR, FATAL
- **双输出目标**：同时输出到 stderr 和文件
- **调用者追踪**：自动记录文件名和行号
- **时间戳**：精确到毫秒的时间戳

**使用方法**：
```go
import "github.com/alingse/qodercli-reverse/decompiled/core/log"

// 初始化（在 main 函数中）
log.Init("~/.qoder/qodercli.log", log.LevelDebug)
defer log.Close()

// 使用示例
log.Debug("Debug message: %s", data)
log.Info("Info message")
log.Warn("Warning message")
log.Error("Error: %v", err)
log.Fatal("Fatal error")
```

**CLI 参数**：
- `--debug`: 启用 DEBUG 级别日志
- `--log-file <path>`: 指定日志文件路径（默认：`~/.qoder/qodercli.log`）

### 2. TUI 组件增强

#### 2.1 Chat 组件 (`tui/components/chat/chat.go`)
- **集成 Markdown 渲染**：使用 `charmbracelet/glamour` 库渲染 Markdown 格式
- **渲染缓存**：避免重复渲染，提高性能
- **动态换行**：根据窗口大小自动调整

#### 2.2 Model 组件 (`tui/app/model.go`)
- **日志集成**：在所有关键操作处添加日志记录
- **错误订阅**：订阅 `EventTypeAgentError` 事件并显示错误
- **事件追踪**：记录所有重要事件用于调试

### 3. Provider 日志 (`core/agent/provider/openai.go`)
- **请求日志**：记录请求 URL、模型、body 大小
- **响应日志**：记录 HTTP 状态码、token 使用情况
- **错误日志**：详细记录 API 错误和网络错误

### 4. CLI 命令集成 (`cmd/root.go`)
- **日志初始化**：在 TUI 模式和 Print 模式下都正确初始化日志
- **错误处理**：使用新的日志系统替代标准 log 包
- **环境检测**：记录 API key 检测和 Provider 选择过程

## 文件修改清单

### 新建文件
| 文件 | 描述 |
|------|------|
| `core/log/log.go` | 日志系统核心实现 |

### 修改文件
| 文件 | 修改内容 |
|------|----------|
| `tui/components/chat/chat.go` | 添加 Markdown 渲染支持 |
| `tui/app/model.go` | 集成日志系统和错误处理 |
| `core/agent/provider/openai.go` | 添加 API 调用日志 |
| `cmd/root.go` | 集成日志系统初始化 |
| `go.mod` | 添加 glamour 依赖 |

## 日志文件位置

默认日志文件路径：`~/.qoder/qodercli.log`

可以通过以下命令查看实时日志：
```bash
tail -f ~/.qoder/qodercli.log
```

## 使用示例

### 1. TUI 模式（带调试日志）
```bash
export OPENAI_API_KEY="your-api-key"
./qodercli --debug
```

### 2. Print 模式（非交互）
```bash
export OPENAI_API_KEY="your-api-key"
./qodercli --debug --print "解释一下 Go 的 context 包"
```

### 3. 自定义日志文件
```bash
./qodercli --debug --log-file ./my-log.txt --print "Hello"
```

## 日志输出示例

**控制台输出**（stderr）：
```
[2026-03-09 20:53:01.110] [INFO] [root.go:138] Starting qodercli in print mode
[2026-03-09 20:53:01.283] [ERROR] [openai.go:113] API error response: status=401, body={...}
```

**文件输出**（包含 DEBUG）：
```
[2026-03-09 20:53:01.110] [INFO] [root.go:138] Starting qodercli in print mode
[2026-03-09 20:53:01.110] [DEBUG] [root.go:139] Input: test
[2026-03-09 20:53:01.110] [DEBUG] [root.go:307] Checking environment variables for API keys
[2026-03-09 20:53:01.110] [DEBUG] [openai.go:56] Starting OpenAI stream request to https://api.deepseek.com/v1
[2026-03-09 20:53:01.283] [ERROR] [openai.go:113] API error response: status=401
```

## 问题排查流程

当遇到问题时：

1. **启用调试模式**：
   ```bash
   ./qodercli --debug --print "your question"
   ```

2. **查看日志文件**：
   ```bash
   cat ~/.qoder/qodercli.log
   ```

3. **关键日志点**：
   - API 请求是否发送（`Starting OpenAI stream request`）
   - HTTP 响应状态码（`HTTP response status`）
   - Token 使用情况（`Token usage`）
   - 工具调用记录（`Tool started`, `Tool result`）

## 注意事项

1. **隐私保护**：日志文件可能包含敏感信息（如 API 调用内容），请妥善保管
2. **日志轮转**：当前未实现日志轮转，建议定期清理 `~/.qoder/qodercli.log`
3. **性能影响**：DEBUG 级别会产生大量日志，生产环境建议使用 INFO 级别

## 编译说明

```bash
cd decompiled/
go mod tidy
go build -o qodercli .
```

## 依赖更新

新增了以下依赖：
- `github.com/charmbracelet/glamour v1.0.0` - Markdown 渲染
- 相关传递依赖（goldmark, chroma 等）

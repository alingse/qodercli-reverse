# qodercli 日志系统文档

> 更新日期：2026-03-09  
> 状态：已完成（超越原版）

---

## 1. 概述

qodercli 实现了一个完整的日志系统，提供比原版更强大的调试能力。

### 1.1 核心特性

- ✅ **多级别日志**: DEBUG, INFO, WARN, ERROR, FATAL
- ✅ **双输出目标**: 同时输出到 stderr 和文件
- ✅ **调用者追踪**: 自动记录文件名和行号
- ✅ **时间戳**: 精确到毫秒
- ✅ **CLI 控制**: 通过 `--debug` 和 `--log-file` 参数控制

### 1.2 与原版对比

| 特性 | 原版 | 当前实现 | 说明 |
|------|------|----------|------|
| 日志级别 | INFO only | 5 级别 | 更细粒度控制 |
| 输出目标 | stderr only | stderr + 文件 | 便于排查问题 |
| 调用者信息 | ❌ | ✅ | 快速定位问题 |
| CLI 参数 | ❌ | ✅ | 灵活控制 |
| 总体评分 | 2/5 | 5/5 | **超越原版** |

---

## 2. 架构设计

### 2.1 包结构

```
core/
└── log/
    └── log.go        # 日志系统核心实现
```

### 2.2 日志级别

```go
type Level int

const (
    LevelDebug Level = iota  // 0 - 调试信息
    LevelInfo                // 1 - 一般信息
    LevelWarn                // 2 - 警告
    LevelError               // 3 - 错误
    LevelFatal               // 4 - 致命错误（退出程序）
)
```

### 2.3 输出格式

```
[时间戳] [级别] [调用者] 消息内容
```

**示例**:
```
[2026-03-09 20:53:01.110] [INFO] [root.go:138] Starting qodercli in TUI mode
[2026-03-09 20:53:01.110] [DEBUG] [openai.go:56] Starting OpenAI stream request to https://api.openai.com/v1
[2026-03-09 20:53:01.283] [ERROR] [openai.go:113] API error response: status=401
```

---

## 3. 使用方法

### 3.1 初始化日志

```go
import "github.com/alingse/qodercli-reverse/decompiled/core/log"

// 在 main 或入口函数中初始化
func main() {
    // 初始化日志（默认级别 INFO）
    log.Init("~/.qoder/qodercli.log", log.LevelInfo)
    defer log.Close()
    
    // ... 程序逻辑
}
```

### 3.2 记录日志

```go
// DEBUG - 调试信息（不输出到 stderr）
log.Debug("Debug message: %s", data)
log.Debugf("Request body size: %d bytes", len(body))

// INFO - 一般信息
log.Info("Starting qodercli in TUI mode")
log.Infof("Log file: %s", logFile)

// WARN - 警告
log.Warn("Configuration file not found, using defaults")

// ERROR - 错误
log.Error("API error: %v", err)
log.Errorf("Failed to create provider: %v", err)

// FATAL - 致命错误（会退出程序）
log.Fatal("Critical error: %v", err)
log.Fatalf("Cannot continue: %v", err)
```

### 3.3 CLI 参数控制

**启用调试模式**:
```bash
./qodercli --debug
```

**指定日志文件**:
```bash
./qodercli --log-file ./my-log.txt
```

**组合使用**:
```bash
./qodercli --debug --log-file ./debug.log --print "Hello"
```

---

## 4. 日志覆盖范围

### 4.1 TUI 模式日志

**启动阶段**:
```go
log.Info("Starting qodercli in TUI mode")
log.Debug("Log file: %s", logFile)
log.Debug("Checking environment variables for API keys")
```

**Agent 交互**:
```go
log.Debug("Sending user input to agent: %s", input)
log.Debug("Received agent response: %d chars", len(content))
log.Debug("Agent processing completed")
log.Error("Agent processing error: %v", err)
```

**事件订阅**:
```go
log.Debug("Setting up event subscriptions")
log.Debug("Tool started: %s", info["name"])
log.Debug("Token usage: input=%d, output=%d", usage.InputTokens, usage.OutputTokens)
log.Error("Event error: %v", err)
```

### 4.2 Print 模式日志

```go
log.Info("Starting qodercli in print mode")
log.Debug("Input: %s", input)
log.Error("Failed to create provider: %v", err)
log.Info("Request completed with finish reason: %s", reason)
```

### 4.3 Provider 层日志

**OpenAI Provider**:
```go
// 请求开始
log.Debug("Starting OpenAI stream request to %s", c.baseURL)
log.Debug("Request model: %s, max_tokens: %d", req.Model, req.MaxTokens)
log.Debug("Request body size: %d bytes", len(body))

// HTTP 响应
log.Debug("HTTP response status: %d", resp.StatusCode)

// 错误处理
log.Error("HTTP request failed: %v", err)
log.Error("API error response: status=%d, body=%s", resp.StatusCode, string(resp.Body))

// 完成
log.Debug("Stream processing completed")
```

**响应解析**:
```go
log.Debug("Response: model=%s, tokens=%d+%d", 
    apiResp.Model, 
    apiResp.Usage.InputTokens, 
    apiResp.Usage.OutputTokens)
```

### 4.4 工具调用日志

```go
log.Debug("Tool call: %s", call.Name)
log.Debug("Tool result: %s", result.ToolCallID)
log.Error("Callback error: %v", err)
```

---

## 5. 日志文件管理

### 5.1 默认位置

```
~/.qoder/qodercli.log
```

### 5.2 日志轮转（手动）

当前未实现自动日志轮转，建议定期清理：

```bash
# 查看日志大小
ls -lh ~/.qoder/qodercli.log

# 清空日志
> ~/.qoder/qodercli.log

# 或删除旧日志
rm ~/.qoder/qodercli.log
```

### 5.3 日志分析

**查看错误日志**:
```bash
grep "\[ERROR\]" ~/.qoder/qodercli.log
```

**查看 API 调用**:
```bash
grep "OpenAI" ~/.qoder/qodercli.log
```

**查看工具调用**:
```bash
grep "Tool" ~/.qoder/qodercli.log
```

**实时跟踪**:
```bash
tail -f ~/.qoder/qodercli.log
```

---

## 6. 故障排查

### 6.1 常见问题

**问题 1**: 看不到 DEBUG 日志

**解决**:
```bash
# 确保使用了 --debug 参数
./qodercli --debug
```

**问题 2**: 日志文件为空

**解决**:
```bash
# 检查日志级别（只有 INFO+ 才会写入 stderr）
# DEBUG 只写入文件
cat ~/.qoder/qodercli.log
```

**问题 3**: 性能问题（日志太多）

**解决**:
```bash
# 不使用 --debug，只用 INFO 级别
./qodercli

# 或禁用文件日志
./qodercli --log-file ""
```

### 6.2 调试技巧

**技巧 1**: 过滤特定组件日志
```bash
grep "openai.go" ~/.qoder/qodercli.log
```

**技巧 2**: 统计错误数量
```bash
grep -c "\[ERROR\]" ~/.qoder/qodercli.log
```

**技巧 3**: 查看最近 100 行
```bash
tail -100 ~/.qoder/qodercli.log
```

**技巧 4**: 时间范围筛选
```bash
# 查看特定时间的日志
grep "20:53:" ~/.qoder/qodercli.log
```

---

## 7. 最佳实践

### 7.1 日志级别选择

| 场景 | 推荐级别 | 说明 |
|------|----------|------|
| 生产环境 | INFO | 记录关键信息 |
| 开发调试 | DEBUG | 详细调试信息 |
| 问题排查 | DEBUG | 完整日志追踪 |
| CI/CD | WARN | 只记录警告和错误 |

### 7.2 日志内容规范

**好的日志**:
```go
// ✅ 包含上下文信息
log.Debug("HTTP response status: %d", resp.StatusCode)
log.Error("API error: status=%d, body=%s", resp.StatusCode, string(resp.Body))

// ✅ 使用一致的格式
log.Info("Starting qodercli in %s mode", mode)
log.Info("Request completed with finish reason: %s", reason)
```

**不好的日志**:
```go
// ❌ 信息不足
log.Debug("Error occurred")

// ❌ 格式不一致
log.Info("Started")
log.Info("starting...")
```

### 7.3 敏感信息处理

**避免记录**:
```go
// ❌ 不要记录完整的 API Key
log.Debug("API Key: %s", apiKey)

// ✅ 只记录部分信息
log.Debug("Using API Key ending with: %s", apiKey[len(apiKey)-4:])
```

---

## 8. 扩展开发

### 8.1 添加新的日志输出目标

```go
// 自定义日志处理器
func init() {
    // 可以添加网络日志、数据库日志等
    log.AddOutput(os.Stdout)
    log.AddOutput(networkWriter)
}
```

### 8.2 自定义日志格式

```go
// 修改 formatLog 函数
func formatLog(level Level, callerInfo, format string, args ...interface{}) string {
    // 自定义格式
    return fmt.Sprintf("[%s] %s: %s", 
        level.String(), 
        callerInfo, 
        fmt.Sprintf(format, args...))
}
```

### 8.3 日志钩子

```go
// 在特定事件时触发
log.SetHook(func(entry *log.Entry) {
    if entry.Level == log.LevelError {
        // 发送告警通知
        sendAlert(entry.Message)
    }
})
```

---

## 9. API 参考

### 9.1 初始化函数

```go
func Init(logFile string, level Level) error
func SetLevel(level Level)
func GetLevel() Level
func Close() error
```

### 9.2 日志记录函数

```go
func Debug(format string, args ...interface{})
func Info(format string, args ...interface{})
func Warn(format string, args ...interface{})
func Error(format string, args ...interface{})
func Fatal(format string, args ...interface{})

// 带格式的别名
func Debugf(format string, args ...interface{})
func Infof(format string, args ...interface{})
func Warnf(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Fatalf(format string, args ...interface{})
```

### 9.3 辅助函数

```go
func GetLogFile() string
func WithPrefix(prefix string) func(string, ...interface{})
```

---

## 10. 总结

qodercli 的日志系统提供了以下优势：

✅ **完整的日志级别** - 满足不同场景需求  
✅ **双输出目标** - 便于调试和监控  
✅ **调用者追踪** - 快速定位问题源头  
✅ **CLI 灵活控制** - 按需调整日志行为  
✅ **超越原版** - 比原版提供更强的调试能力  

推荐在日常开发中使用 `--debug` 参数，获得最详细的日志信息用于问题排查。

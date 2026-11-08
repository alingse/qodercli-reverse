# qodercli-reverse Logging System Documentation

> Last updated: 2026-03-09  
> Status: Completed (surpassed the original)

---

## 1. Overview

qodercli-reverse implements a complete logging system that provides more powerful debugging capabilities than the original.

### 1.1 Core Features

- ✅ **Multi-level logging**: DEBUG, INFO, WARN, ERROR, FATAL
- ✅ **Dual output targets**: Outputs to both stderr and a file
- ✅ **Caller tracking**: Automatically records file name and line number
- ✅ **Timestamp**: Precise to milliseconds
- ✅ **CLI control**: Controlled via the `--debug` and `--log-file` parameters

### 1.2 Comparison with the Original

| Feature | Original | Current Implementation | Description |
|------|------|----------|------|
| Log level | INFO only | 5 levels | Finer-grained control |
| Output target | stderr only | stderr + file | Easier to troubleshoot |
| Caller info | ❌ | ✅ | Quickly locate the issue |
| CLI parameter | ❌ | ✅ | Flexible control |
| Overall score | 2/5 | 5/5 | **Surpassed the original** |

---

## 2. Architecture Design

### 2.1 Package Structure

```
core/
└── log/
    └── log.go        # Core implementation of the logging system
```

### 2.2 Log Levels

```go
type Level int

const (
    LevelDebug Level = iota  // 0 - Debug information
    LevelInfo                // 1 - Info
    LevelWarn                // 2 - Warning
    LevelError               // 3 - Error
    LevelFatal               // 4 - Fatal error (exits the program)
)
```

### 2.3 Output Format

```
[Timestamp] [Level] [Caller] Message
```

**Example**:
```
[2026-03-09 20:53:01.110] [INFO] [root.go:138] Starting qodercli-reverse in TUI mode
[2026-03-09 20:53:01.110] [DEBUG] [openai.go:56] Starting OpenAI stream request to https://api.openai.com/v1
[2026-03-09 20:53:01.283] [ERROR] [openai.go:113] API error response: status=401
```

---

## 3. Usage

### 3.1 Initialize Logging

```go
import "github.com/alingse/qodercli-reverse/decompiled/core/log"

// Initialize in main or the entry function
func main() {
    // Initialize logging (default level INFO)
    log.Init("~/.qoder/qodercli-reverse.log", log.LevelInfo)
    defer log.Close()
    
    // ... program logic
}
```

### 3.2 Record Logs

```go
// DEBUG - Debug information (does not output to stderr)
log.Debug("Debug message: %s", data)
log.Debugf("Request body size: %d bytes", len(body))

// INFO - Info
log.Info("Starting qodercli-reverse in TUI mode")
log.Infof("Log file: %s", logFile)

// WARN - Warning
log.Warn("Configuration file not found, using defaults")

// ERROR - Error
log.Error("API error: %v", err)
log.Errorf("Failed to create provider: %v", err)

// FATAL - Fatal error (exits the program)
log.Fatal("Critical error: %v", err)
log.Fatalf("Cannot continue: %v", err)
```

### 3.3 CLI Parameter Control

**Enable debug mode**:
```bash
./qodercli-reverse --debug
```

**Specify log file**:
```bash
./qodercli-reverse --log-file ./my-log.txt
```

**Use together**:
```bash
./qodercli-reverse --debug --log-file ./debug.log --print "Hello"
```

---

## 4. Logging Scope

### 4.1 TUI Mode Logging

**Startup phase**:
```go
log.Info("Starting qodercli-reverse in TUI mode")
log.Debug("Log file: %s", logFile)
log.Debug("Checking environment variables for API keys")
```

**Agent Interaction**:
```go
log.Debug("Sending user input to agent: %s", input)
log.Debug("Received agent response: %d chars", len(content))
log.Debug("Agent processing completed")
log.Error("Agent processing error: %v", err)
```

**Event Subscription**:
```go
log.Debug("Setting up event subscriptions")
log.Debug("Tool started: %s", info["name"])
log.Debug("Token usage: input=%d, output=%d", usage.InputTokens, usage.OutputTokens)
log.Error("Event error: %v", err)
```

### 4.2 Print Mode Logging

```go
log.Info("Starting qodercli-reverse in print mode")
log.Debug("Input: %s", input)
log.Error("Failed to create provider: %v", err)
log.Info("Request completed with finish reason: %s", reason)
```

### 4.3 Provider Layer Logging

**OpenAI Provider**:
```go
// Request start
log.Debug("Starting OpenAI stream request to %s", c.baseURL)
log.Debug("Request model: %s, max_tokens: %d", req.Model, req.MaxTokens)
log.Debug("Request body size: %d bytes", len(body))

// HTTP response
log.Debug("HTTP response status: %d", resp.StatusCode)

// Error handling
log.Error("HTTP request failed: %v", err)
log.Error("API error response: status=%d, body=%s", resp.StatusCode, string(resp.Body))

// Completion
log.Debug("Stream processing completed")
```

**Response parsing**:
```go
log.Debug("Response: model=%s, tokens=%d+%d", 
    apiResp.Model, 
    apiResp.Usage.InputTokens, 
    apiResp.Usage.OutputTokens)
```

### 4.4 Tool Call Logging

```go
log.Debug("Tool call: %s", call.Name)
log.Debug("Tool result: %s", result.ToolCallID)
log.Error("Callback error: %v", err)
```

---

## 5. Log File Management

### 5.1 Default Location

```
~/.qoder/qodercli-reverse.log
```

### 5.2 Log Rotation (Manual)

Automatic log rotation is not implemented yet; it is recommended to clean up periodically:

```bash
# View log size
ls -lh ~/.qoder/qodercli-reverse.log

# Clear the log
> ~/.qoder/qodercli-reverse.log

# Or delete old logs
rm ~/.qoder/qodercli-reverse.log
```

### 5.3 Log Analysis

**View error logs**:
```bash
grep "\[ERROR\]" ~/.qoder/qodercli-reverse.log
```

**View API calls**:
```bash
grep "OpenAI" ~/.qoder/qodercli-reverse.log
```

**View tool calls**:
```bash
grep "Tool" ~/.qoder/qodercli-reverse.log
```

**Real-time tracking**:
```bash
tail -f ~/.qoder/qodercli-reverse.log
```

---

## 6. Troubleshooting

### 6.1 Common Issues

**Issue 1**: DEBUG logs are not visible

**Resolution**:
```bash
# Make sure to use the --debug parameter
./qodercli-reverse --debug
```

**Issue 2**: The log file is empty

**Resolution**:
```bash
# Check the log level (only INFO+ is written to stderr)
# DEBUG is written to the file only
cat ~/.qoder/qodercli-reverse.log
```

**Issue 3**: Performance issues (too many logs)

**Resolution**:
```bash
# Do not use --debug; use INFO level only
./qodercli-reverse

# Or disable file logging
./qodercli-reverse --log-file ""
```

### 6.2 Debugging Tips

**Tip 1**: Filter logs for specific components
```bash
grep "openai.go" ~/.qoder/qodercli-reverse.log
```

**Tip 2**: Count errors
```bash
grep -c "\[ERROR\]" ~/.qoder/qodercli-reverse.log
```

**Tip 3**: View the last 100 lines
```bash
tail -100 ~/.qoder/qodercli-reverse.log
```

**Tip 4**: Time range filtering
```bash
# View logs from a specific time
grep "20:53:" ~/.qoder/qodercli-reverse.log
```

---

## 7. Best Practices

### 7.1 Log Level Selection

| Scenario | Recommended Level | Description |
|------|----------|------|
| Production | INFO | Record key information |
| Development | DEBUG | Detailed debug information |
| Troubleshooting | DEBUG | Full log tracing |
| CI/CD | WARN | Only record warnings and errors |

### 7.2 Log Content Guidelines

**Good logs**:
```go
// ✅ Include contextual information
log.Debug("HTTP response status: %d", resp.StatusCode)
log.Error("API error: status=%d, body=%s", resp.StatusCode, string(resp.Body))

// ✅ Use a consistent format
log.Info("Starting qodercli-reverse in %s mode", mode)
log.Info("Request completed with finish reason: %s", reason)
```

**Bad logs**:
```go
// ❌ Insufficient information
log.Debug("Error occurred")

// ❌ Inconsistent format
log.Info("Started")
log.Info("starting...")
```

### 7.3 Sensitive Information Handling

**Avoid recording**:
```go
// ❌ Do not record the full API Key
log.Debug("API Key: %s", apiKey)

// ✅ Only record partial information
log.Debug("Using API Key ending with: %s", apiKey[len(apiKey)-4:])
```

---

## 8. Extension Development

### 8.1 Adding a New Log Output Target

```go
// Custom log handler
func init() {
    // You can add network logs, database logs, etc.
    log.AddOutput(os.Stdout)
    log.AddOutput(networkWriter)
}
```

### 8.2 Custom Log Format

```go
// Modify the formatLog function
func formatLog(level Level, callerInfo, format string, args ...interface{}) string {
    // Custom format
    return fmt.Sprintf("[%s] %s: %s", 
        level.String(), 
        callerInfo, 
        fmt.Sprintf(format, args...))
}
```

### 8.3 Log Hooks

```go
// Trigger at specific events
log.SetHook(func(entry *log.Entry) {
    if entry.Level == log.LevelError {
        // Send alert notifications
        sendAlert(entry.Message)
    }
})
```

---

## 9. API Reference

### 9.1 Initialization Functions

```go
func Init(logFile string, level Level) error
func SetLevel(level Level)
func GetLevel() Level
func Close() error
```

### 9.2 Log Recording Functions

```go
func Debug(format string, args ...interface{})
func Info(format string, args ...interface{})
func Warn(format string, args ...interface{})
func Error(format string, args ...interface{})
func Fatal(format string, args ...interface{})

// Formatted aliases
func Debugf(format string, args ...interface{})
func Infof(format string, args ...interface{})
func Warnf(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Fatalf(format string, args ...interface{})
```

### 9.3 Helper Functions

```go
func GetLogFile() string
func WithPrefix(prefix string) func(string, ...interface{})
```

---

## 10. Summary

The qodercli-reverse logging system provides the following advantages:

✅ **Complete log levels** - Meets different scenario requirements  
✅ **Dual output targets** - Facilitates debugging and monitoring  
✅ **Caller tracking** - Quickly locate the source of the issue  
✅ **Flexible CLI control** - Adjust log behavior as needed  
✅ **Surpassed the original** - Provides stronger debugging capabilities than the original  

It is recommended to use the `--debug` parameter during daily development to obtain the most detailed log information for troubleshooting.

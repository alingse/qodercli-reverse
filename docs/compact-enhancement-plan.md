# Compact Feature Enhancement Plan

> **Analysis Date**: 2026-03-09  
> **Analysis Method**: Binary symbol analysis (`strings` + `grep`)  
> **Related Document**: [compact-architecture-analysis.md](./compact-architecture-analysis.md)

## Current Status Summary

By locating the binary via `which qodercli-reverse`, extracting 157,441 lines of strings with `strings`, filtering 364 compact-related lines, and analyzing the complete function symbol table.

### Key Findings

**The original compact is a complex system containing:**
- 5 compression strategies (Micro/Full/SessionMemory/Summarize/Traditional)
- 3 trigger methods (Auto/Manual/Reload)
- Complete token counting and threshold detection
- Session Memory independent persistence module
- Pre-Compact Hook extension system
- Dedicated TUI components
- ACP protocol integration

**The current decompiled code has only ~50 lines of simplified implementation, missing all of the above advanced features.**

## Table of Contents

- [Phase 1: Basic Architecture (P0)](#phase-1-basic-architecture-p0)
- [Phase 2: MicroCompact (P1)](#phase-2-microcompact-p1)
- [Phase 3: FullCompact and LLM Summary (P1-P2)](#phase-3-fullcompact-and-llm-summary-p1-p2)
- [Phase 4: Session Memory Basics (P1)](#phase-4-session-memory-basics-p1)
- [Phase 5: Pre-Compact Hook System (P1)](#phase-5-pre-compact-hook-system-p1)
- [Phase 6: Configuration System (P2)](#phase-6-configuration-system-p2)
- [Phase 7: TUI Integration (P2)](#phase-7-tui-integration-p2)
- [Implementation Order Suggestions](#implementation-order-suggestions)
- [Risks and Challenges](#risks-and-challenges)
- [Acceptance Criteria](#acceptance-criteria)

---

## Phase 1: Basic Architecture (P0)

### 1.1 Define CompactTrigger Type

**File**: `core/types/compact.go`

```go
package types

// CompactTriggerType compression trigger type
type CompactTriggerType string

const (
    CompactTriggerAuto   CompactTriggerType = "auto"
    CompactTriggerManual CompactTriggerType = "manual"
    CompactTriggerReload CompactTriggerType = "reload"
)

// CompactTrigger compression trigger
type CompactTrigger struct {
    Type        CompactTriggerType `json:"type"`
    TokenCount  int                `json:"token_count,omitempty"`
    Threshold   int                `json:"threshold,omitempty"`
    Instruction string             `json:"instruction,omitempty"` // user instruction (when manually triggered)
}

// IsAuto whether it's an auto trigger
func (c *CompactTrigger) IsAuto() bool {
    return c.Type == CompactTriggerAuto
}

// IsManual whether it's a manual trigger
func (c *CompactTrigger) IsManual() bool {
    return c.Type == CompactTriggerManual
}

// IsReload whether it's a reload trigger
func (c *CompactTrigger) IsReload() bool {
    return c.Type == CompactTriggerReload
}

// String returns string representation
func (c *CompactTrigger) String() string {
    return string(c.Type)
}
```

### 1.2 Add Token Counting to State

**File**: `core/agent/state/state.go`

```go
// New methods needed
type State struct {
    messages       []types.Message
    toolResults    map[string]*types.ToolResult
    tokenCount     int  // new: current token count
    mu             sync.RWMutex
}

// GetTokenCount get current token count
func (s *State) GetTokenCount() int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.tokenCount
}

// UpdateTokenCount update token count
func (s *State) UpdateTokenCount(count int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.tokenCount = count
}
```

### 1.3 Implement Threshold Detection

**File**: `core/agent/compact.go` (new)

```go
package agent

const (
    // DefaultCompactThreshold default compression threshold (80% of context)
    DefaultCompactThreshold = 0.8
    
    // WarningThresholdK warning threshold (in K tokens)
    WarningThresholdK = 100 // 100K tokens
)

// getCompactThreshold get compression threshold
func getCompactThreshold(modelMaxTokens int) int {
    return int(float64(modelMaxTokens) * DefaultCompactThreshold)
}

// shouldRunMicroCompact determine whether to run micro compact
func shouldRunMicroCompact(currentTokens, threshold int) bool {
    // Slight overflow (less than 10%) uses micro compact
    return currentTokens > threshold && currentTokens < int(float64(threshold)*1.1)
}
```

### 1.4 Auto Compact Trigger Mechanism

**File**: `core/agent/agent.go` (modify generate method)

```go
// Add at the beginning of the generate() method loop
func (a *Agent) generate(ctx context.Context) error {
    for {
        // === new: check if auto compact needed ===
        if err := a.checkAndAutoCompact(ctx); err != nil {
            log.Debug("Auto compact check failed: %v", err)
        }
        // ===================================
        
        // ... existing logic
    }
}

// checkAndAutoCompact check and execute auto compact
func (a *Agent) checkAndAutoCompact(ctx context.Context) error {
    // 1. Get current token count
    currentTokens := a.state.GetTokenCount()
    
    // 2. Get model config
    maxTokens := a.config.MaxTokens
    threshold := getCompactThreshold(maxTokens)
    
    // 3. Determine if compaction needed
    if currentTokens <= threshold {
        return nil // no compaction needed
    }
    
    // 4. Choose compaction strategy
    trigger := &types.CompactTrigger{
        Type:       types.CompactTriggerAuto,
        TokenCount: currentTokens,
        Threshold:  threshold,
    }
    
    if shouldRunMicroCompact(currentTokens, threshold) {
        return a.MicroCompact(ctx, trigger)
    }
    
    return a.fullCompact(ctx, trigger)
}
```

---

## Phase 2: MicroCompact (P1)

### 2.1 Implement MicroCompact

**File**: `core/agent/compact.go`

```go
// MicroCompact micro compression - fast and lightweight
func (a *Agent) MicroCompact(ctx context.Context, trigger *types.CompactTrigger) error {
    messages := a.state.GetMessages()
    if len(messages) <= 3 {
        return nil
    }
    
    var compressed []types.Message
    
    // 1. Keep system message
    if len(messages) > 0 && messages[0].Role == types.RoleSystem {
       compressed = append(compressed, messages[0])
    }
    
    // 2. Keep last N complete conversations (e.g. 5 rounds)
    keepRounds := 5
    keepCount := keepRounds * 2 // user + assistant
    if len(messages) > keepCount+1 { // +1 for system message
        // 3. Generate brief summary for removed messages
        summary := a.generateMicroSummary(messages[1 : len(messages)-keepCount])
       compressed = append(compressed, types.Message{
            Role:    types.RoleSystem,
            Content: []types.ContentPart{{Type: "text", Text: summary}},
        })
    }
    
    // 4. Keep recent conversation
    if len(messages) > keepCount {
       compressed = append(compressed, messages[len(messages)-keepCount:]...)
    }
    
    // 5. Update state
    a.state.SetMessages(compressed)
    
    // 6. Report statistics
    savedTokens := trigger.TokenCount - a.calculateTokenCount(compressed)
    log.Debug("micro compact complete, compressedResults: %d, savedTokens: %d", 
        len(messages)-len(compressed), savedTokens)
    
    return nil
}

// generateMicroSummary generate micro summary (extract key points only)
func (a *Agent) generateMicroSummary(messages []types.Message) string {
    var summary strings.Builder
    summary.WriteString("Previous conversation summary:\n")
    
    for _, msg := range messages {
        if msg.Role == types.RoleUser {
            summary.WriteString("- User: ")
            for _, part := range msg.Content {
                if part.Type == "text" {
                    // Take first 100 chars
                    text := part.Text
                    if len(text) > 100 {
                        text = text[:100] + "..."
                    }
                    summary.WriteString(text)
                    break
                }
            }
            summary.WriteString("\n")
        }
    }
    
    return summary.String()
}
```

---

## Phase 3: FullCompact and LLM Summary (P1-P2)

### 3.1 FullCompact Framework

```go
// fullCompact full compression
func (a *Agent) fullCompact(ctx context.Context, trigger *types.CompactTrigger) error {
    // 1. Send compaction start notification
    a.reportCompactStart(trigger)
    
    // 2. Execute compression (may call LLM)
    var err error
    if a.shouldSummarizeCompact(trigger) {
        err = a.summarizeCompact(ctx, trigger)
    } else {
        err = a.traditionalSummarizeCompact(ctx, trigger)
    }
    
    // 3. Send completion notification
    a.reportCompactComplete(trigger, err)
    
    return err
}
```

### 3.2 LLM Summary (summarizeCompact)

```go
// summarizeCompact use LLM to generate semantic summary
func (a *Agent) summarizeCompact(ctx context.Context, trigger *types.CompactTrigger) error {
    messages := a.state.GetMessages()
    
    // Build summary request
    summaryReq := &provider.ModelRequest{
        Model:      a.config.Model,
        Messages:   messages,
        SystemPrompt: "Please summarize the key points of this conversation...",
        MaxTokens:  1000,
    }
    
    // Call LLM to generate summary
    // ... (need to call provider.Stream or similar method)
    
    return nil
}
```

---

## Phase 4: Session Memory Basics (P1)

### 4.1 Session Memory State Interface

**File**: `core/agent/state/session_memory/state.go` (new directory)

```go
package session_memory

import (
    "github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// State Session Memory state interface
type State interface {
    // config
    GetCompactConfig() *SMCompactConfig
    
    // memory management
    GetCurrentMemory() string
    GetLastSummarizedMsgId() string
    
    // update control
    TriggerUpdate() error
    WaitForUpdate(ctx context.Context) error
    MarkUpdateStart()
    MarkUpdateComplete()
    
    // state check
    ShouldUpdate() bool
    HasPendingContent() bool
    HasPendingToolUse() bool
    
    // persistence
    GetMemoryPath() string
    EnsureTemplate() error
}

// SMCompactConfig Session Memory compression config
type SMCompactConfig struct {
    Enabled           bool    `json:"enabled"`
    AutoCompact       bool    `json:"auto_compact"`
    ThresholdPercent  float64 `json:"threshold_percent"`
    KeepLastRounds    int     `json:"keep_last_rounds"`
}
```

### 4.2 SerialQueue Serial Queue

```go
// SerialQueue ensures update operations execute in order
type SerialQueue struct {
    queue    chan func()
    running  bool
    mu       sync.Mutex
}

// NewSerialQueue create serial queue
func NewSerialQueue() *SerialQueue {
    return &SerialQueue{
        queue:   make(chan func(), 100),
        running: true,
    }
}

// Enqueue enqueue
func (q *SerialQueue) Enqueue(task func()) {
    select {
    case q.queue <- task:
    default:
        // queue full, block and wait
        q.queue <- task
    }
}

// processQueue process queue
func (q *SerialQueue) processQueue() {
    for task := range q.queue {
        if !q.running {
            return
        }
        task()
    }
}
```

---

## Phase 5: Pre-Compact Hook System (P1)

### 5.1 Hook Definition

**File**: `core/agent/hooks/pre_compact/pre_compact.go` (new directory)

```go
package pre_compact

import (
    "context"
    "github.com/alingse/qodercli-reverse/decompiled/core/resource/hook"
)

// PreCompactHook pre-compact Hook interface
type PreCompactHook interface {
    Execute(ctx context.Context, input *hook.PreCompactInput) (*hook.Result, error)
}

// PreCompactExternalHook external Hook implementation
type PreCompactExternalHook struct {
    handler func(ctx context.Context, input *hook.PreCompactInput) (*hook.Result, error)
}

// Execute execute Hook
func (h *PreCompactExternalHook) Execute(ctx context.Context, input *hook.PreCompactInput) (*hook.Result, error) {
    return h.handler(ctx, input)
}

// NewPreCompactExternalHook create external Hook
func NewPreCompactExternalHook(handler func(context.Context, *hook.PreCompactInput) (*hook.Result, error)) *PreCompactExternalHook {
    return &PreCompactExternalHook{handler: handler}
}
```

### 5.2 Hook Input

**File**: `core/resource/hook/hook.go`

```go
// PreCompactInput pre-compact Hook input
type PreCompactInput struct {
    TriggerType string
    TokenCount   int
    Threshold    int
    MessageCount int
    ToolUseID    string // if there's a tool call
}

// EventName returns event name
func (p *PreCompactInput) EventName() string {
    return "pre_compact"
}

// ToolUseID returns tool call ID
func (p *PreCompactInput) ToolUseID() string {
    return p.ToolUseID
}
```

---

## Phase 6: Configuration System (P2)

### 6.1 Environment Configuration

**File**: `core/utils/env/env.go`

```go
// IsSessionMemoryCompactEnabled check if Session Memory is enabled
func IsSessionMemoryCompactEnabled() bool {
    return os.Getenv("ENABLE_SESSION_MEMORY_COMPACT") == "true"
}

// GetAutoCompactEnabled get auto compact switch
func GetAutoCompactEnabled() bool {
    val := os.Getenv("AUTO_COMPACT_ENABLED")
    return val == "" || val == "true" // enabled by default
}
```

### 6.2 Agent Configuration

```go
// Config add compact-related config
type Config struct {
    // ... existing fields
    
    AutoCompactEnabled bool    `json:"auto_compact_enabled"`
    CompactThreshold    float64 `json:"compact_threshold"` // 0.0-1.0
    KeepLastRounds      int     `json:"keep_last_rounds"`
    UseLLMSummary       bool    `json:"use_llm_summary"`
}
```

---

## Phase 7: TUI Integration (P2)

### 7.1 CompactResult Message View

**File**: `tui/components/messages/compact_result.go`

```go
package messages

// CompactResult compaction result
type CompactResult struct {
    CompressedCount int
    SavedTokens     int
    Error           error
}

// Render render result
func (c *CompactResult) Render(width int) string {
    if c.Error != nil {
        return fmt.Sprintf("❌ Compact failed: %v", c.Error)
    }
    
    return fmt.Sprintf(
        "✓ Compacted (%d messages → %d, saved ~%d tokens)",
        c.CompressedCount,
        c.CompressedCount/3, // estimate
        c.SavedTokens,
    )
}

// IsError check if there's an error
func (c *CompactResult) IsError() bool {
    return c.Error != nil
}

// NoContentTips return no content tip
func (c *CompactResult) NoContentTips() string {
    return "Not enough messages to compact"
}
```

---

## Implementation Order Suggestions

1. **Week 1**: Phase 1 (Basic Architecture)
   - Define CompactTrigger
   - Add token counting
   - Implement threshold detection

2. **Week 2**: Phase 2 (MicroCompact)
   - Implement fast compression
   - Add statistics reporting

3. **Week 3**: Phase 4 (Session Memory basics)
   - State interface
   - Serial queue

4. **Week 4**: Phase 5 (Hook system)
   - Pre-Compact Hook
   - External Hook support

5. **Week 5-6**: Phase 3 & 6 (LLM summary + config)
   - Complete compression flow
   - Config system

6. **Week 7-8**: Phase 7 (TUI) + testing
   - UI components
   - Integration tests

---

## Risks and Challenges

1. **Token counting accuracy**: Requires precise token calculation library
2. **LLM summary cost**: Calling LLM increases API costs
3. **Session Memory persistence**: Need to handle concurrency and crash recovery
4. **Backward compatibility**: Ensure no impact on existing functionality

---

## Acceptance Criteria

- [ ] `/compact` command works normally
- [ ] Auto compact triggers at threshold
- [ ] Token counting accurate
- [ ] MicroCompact completes within 1 second
- [ ] Session Memory persists correctly
- [ ] Hook system extensible
- [ ] TUI displays compaction statistics
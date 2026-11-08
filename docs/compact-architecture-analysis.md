# Deep Analysis of Original qodercli-reverse Compact Implementation

> **Analysis Method**: Locate the binary via `which qodercli-reverse`, use `strings` to extract the symbol table, and analyze compiled function signatures and package structure with grep.

## Executive Summary

This document presents conclusions from reverse engineering analysis of the qodercli-reverse binary (`/Users/zhihu/.local/bin/qodercli-reverse`). The `strings` tool extracted 157,441 lines of strings, from which 364 compact-related lines were filtered, and the complete Go function symbol table was analyzed.

**Key Finding**: The original compact is a complex system of ~4000+ lines, while the current decompiled code has only ~50 lines of simplified implementation.

## Table of Contents

- [1. Core Package Structure](#1-core-package-structure-inferred-from-symbol-paths)
- [2. CompactTrigger Type](#2-compacttrigger-type-typescompacttrigger)
- [3. Multiple Compression Strategies](#3-multiple-compression-strategies)
- [4. Token Management Mechanism](#4-token-management-mechanism)
- [5. Session Memory Detailed Architecture](#5-session-memory-detailed-architecture)
- [6. Pre-Compact Hook System](#6-pre-compact-hook-system)
- [7. Configuration System](#7-configuration-system)
- [8. TUI Integration](#8-tui-integration)
- [9. ACP Protocol Integration](#9-acp-protocol-integration)
- [10. Current Implementation Comparison](#10-current-implementation-comparison)

---

## Complete Architecture Analysis Based on Binary Symbols

### 1. Core Package Structure (inferred from symbol paths)

```
core/agent/
├── compact.go                     # Compact core implementation
├── session_memory.go              # Session Memory integration
├── hooks/pre_compact/              # Pre-Compact Hook system
│   ├── external.go                # External Hook implementation
│   └── pre_compact.go             # PreCompactHook definition
├── state/
│   ├── agent_state.go             # Agent state management
│   ├── model_context.go           # Model context management
│   └── session_memory/             # Session Memory independent package
│       ├── state.go               # State management
│       ├── queue.go               # Serial queue
│       └── limits.go              # Limit configuration
└── prompt/
    └── session_memory.go          # Session Memory prompts
```

### 2. CompactTrigger Type (types.CompactTrigger)

Three trigger modes inferred from symbols:
- `IsAuto()` - Automatic trigger (when context exceeds threshold)
- `IsManual()` - Manual trigger (/compact command)
- `IsReload()` - Reload trigger (on session recovery)

### 3. Multiple Compression Strategies

#### 3.1 MicroCompact (Micro Compression)
- Fast, lightweight compression
- Independent threshold judgment (`shouldRunMicroCompact`)
- Suitable for slight overflow scenarios

#### 3.2 FullCompact (Full Compression)
- Complete compression flow
- May call LLM to generate summary

#### 3.3 SessionMemoryCompact (Session Memory Compression)
- Uses Session Memory mechanism
- Persists compression results
- Supports resumption

#### 3.4 SummarizeCompact (Summary Compression)
- Calls LLM to generate semantic summary
- Has configuration options (`shouldSummarizeCompact`)

#### 3.5 TraditionalSummarizeCompact (Traditional Summary)
- Legacy summary method (possibly fallback)
- Has multiple configurators:
  - WithMeta
  - WithMetadata
  - WithModelConfig
  - WithPostProcessor
  - WithSystemPrompt
  - WithTools

### 4. Token Management Mechanism

From string evidence:
- `getCompactThreshold` - Get compression threshold
- `WarningThreshold %dK` - Warning threshold (in K units)
- `Skip micro compact: total history tokens %d less than WarningThreshold %dK`
- `compact_input_tokens` - Input token count
- `compact_output_tokens` - Output token count
- `savedTokens` - Tokens saved
- `compressed_count` - Number of compressed messages
- `total_candidates` - Total candidate messages
- `kept_messages` - Number of kept messages

### 5. Session Memory Detailed Architecture

#### 5.1 Core Types
- `session_memory.State` - State interface
- `session_memory.SMCompactConfig` - Compression config
- `session_memory.SMUpdateConfig` - Update config
- `session_memory.SerialQueue` - Serial queue (guarantees order)
- `session_memory.SectionTokenInfo` - Chunk token info

#### 5.2 Core Methods
- `GetCompactConfig()` - Get compression config
- `GetCurrentMemory()` - Get current memory
- `GetLastSummarizedMsgId()` - Get last summarized message ID
- `EnsureTemplate()` - Ensure template exists
- `TriggerUpdate()` - Trigger update
- `WaitForUpdate()` - Wait for update completion
- `MarkUpdateStart/Complete()` - Mark update status
- `ResetState()` - Reset state
- `countToolCallsSince()` - Count tool calls
- `hasPendingContent()` - Check pending content
- `hasPendingToolUse()` - Check pending tool use

#### 5.3 Truncation Mechanism
- `TruncateContent` - Truncate content
- `truncateSections` - Truncate chunks
- `GenerateOverLimitWarning` - Generate over-limit warning
- `GetOverLimitInfo` - Get over-limit info
- `ParseSectionTokens` - Parse chunk tokens

#### 5.4 Persistence
- `GetMemoryPath()` - Get memory file path
- `getMemoryDir()` - Get memory directory
- `IsTemplateEmpty()` - Check if template is empty

### 6. Pre-Compact Hook System

#### 6.1 Hook Types
- `pre_compact.PreCompactHook` - Base Hook interface
- `pre_compact.PreCompactExternalHook` - External Hook

#### 6.2 Hook Input
- `hook.PreCompactInput` - Hook input data
  - `EventName()` - Event name
  - `ToolUseID()` - Tool call ID

#### 6.3 Agent Runner Methods
- `handlePreCompact()` - Handle pre-compact logic
- `handleSessionStartAfterCompact()` - Session recovery after compact

### 7. Configuration System

#### 7.1 Environment Configuration
- `IsSessionMemoryCompactEnabled` - Session Memory switch
- `AutoCompactEnabled` - Auto compact switch (JSON config)
- `GetAutoCompactEnabled()` / `SetAutoCompactEnabled()` - Config access

#### 7.2 Threshold Configuration
- `getCompactThreshold()` - Get compression threshold

### 8. TUI Integration

#### 8.1 CompactCommand Component
- `tui/components/command/compact.go`
- Complete UI component supporting:
  - Keyboard handling (AddKeyHandler, Default*Handler)
  - Rendering (RenderView, RenderItem, RenderProgress)
  - State management (GetCheckboxState, SetCheckboxStates)
  - Refresh mechanism (SetRefreshHandler, CancelRefreshHandler)

#### 8.2 Message View
- `tui/components/messages.CompactResult`
  - `Render()` - Render result
  - `IsError()` - Error check
  - `NoContentTips()` - No content tip

#### 8.3 Processing Flow
- `processCompactPayload` - Process compact payload
- `handleCompactComplete` - Handle compact completion

### 9. ACP (Agent Communication Protocol) Integration

- `acp.(*Session).Compact` - ACP Session compact
- `acp.(*Session).handleCompactEvent` - Handle compact event
- `acp.(*Session).sendCompactionNotification` - Send compaction notification

### 10. Key Log/Status Strings

```
"compaction_triggered"      - compaction triggered
"compaction_start"          - compaction started
"compaction_cancel"         - compaction cancelled
"compact_complete"          - compaction completed
"micro compact complete, compressedResults: %d, savedTokens: %d"
"Starting summarize compact"
"Finished summarize compact"
"Session memory update completed, backtracked to complete turn"
"Truncate compact history until tokens below 0.98"
"SM state reset after traditional compact fallback"
```

---

## Current Decompiled Code vs Original Comparison

### Current Implementation (agent.go:502-557)

```go
func (a *Agent) compactContext(ctx context.Context) error {
    messages := a.state.GetMessages()
    if len(messages) <= 2 {
        return nil
    }
    
    // 1. Keep system messages
    // 2. Simple truncation, take first 50 chars as summary
    // 3. Add summary as system message
    // 4. Keep last 2 messages
    // 5. Update state
    
    return nil
}
```

### Feature Gap Analysis

| Feature Module | Original | Current Decompiled | Gap |
|----------------|----------|-------------------|-----|
| **Trigger Methods** | Auto/Manual/Reload | Manual only | 🔴 Missing auto trigger |
| **Token Counting** | Full tracking | None | 🔴 Completely missing |
| **Threshold Detection** | getCompactThreshold | None | 🔴 Completely missing |
| **MicroCompact** | Yes | No | 🔴 Completely missing |
| **Session Memory** | Independent package, full impl | None | 🔴 Completely missing |
| **LLM Summary** | summarizeCompact | None | 🔴 Completely missing |
| **Pre-Compact Hook** | Full Hook system | None | 🔴 Completely missing |
| **Config System** | AutoCompactEnabled etc. | None | 🔴 Completely missing |
| **Persistence** | Session Memory files | None | 🔴 Completely missing |
| **TUI Component** | CompactCommand | None | 🔴 Completely missing |
| **ACP Integration** | sendCompactionNotification | None | 🔴 Completely missing |
| **Compression Strategy** | 5 strategies | 1 simple strategy | 🔴 Severely simplified |
| **State Management** | SerialQueue ensures order | Simple slice | 🟡 Simplified |
| **Chunk Processing** | SectionTokenInfo | None | 🔴 Completely missing |

### Code Line Count Comparison

| Project | Original (estimated) | Current |
|---------|---------------------|---------|
| Compact core | ~2000+ lines | ~50 lines |
| Session Memory | ~1000+ lines | 0 lines |
| Hooks | ~500+ lines | 0 lines |
| TUI components | ~800+ lines | 0 lines |
| **Total** | **~4000+ lines** | **~50 lines** |

---

## Enhancement Priority Suggestions

### P0 - Core Features (must implement)
1. Token counting and threshold detection
2. Auto Compact trigger mechanism
3. CompactTrigger type definition

### P1 - Important Enhancements
4. MicroCompact fast compression
5. Pre-Compact Hook system
6. Session Memory basic framework

### P2 - UX Optimization
7. TUI CompactCommand component
8. Config system (AutoCompactEnabled)
9. Compression statistics (savedTokens, etc.)

### P3 - Advanced Features
10. LLM summary (summarizeCompact)
11. ACP integration
12. Persistence and resumption
# qodercli-reverse Session Management and Persistence Analysis

## 1. Session Architecture

### 1.1 Package Structure

```
core/agent/state/session/
├── cache/          # Session cache
├── options/        # Session options
├── session/        # Session core
└── storage/        # Session storage
```

## 2. Session Lifecycle

```
session/new      # Create new session
session/load     # Load session
session/update   # Update session
session/cancel   # Cancel session
session/prompt   # Session prompt
session/set_mode # Set session mode
```

## 3. Session Storage

### 3.1 Storage Location

```
~/.qoder/
├── sessions/
│   └── *-session.json    # Session files
├── conversations/
│   └── conversation-*.txt # Exported conversations
└── projects/
    └── {project}/
        └── .qoder/       # Project-level configuration
```

### 3.2 Session File Format

```json
{
  "session_id": "uuid",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "messages": [...],
  "state": {...},
  "metadata": {
    "model": "claude-sonnet-4",
    "workspace": "/path/to/project"
  }
}
```

## 4. Message Management

```
core/agent/state/message/
├── cache/      # Message cache
├── message/    # Message core
└── storage/    # Message storage
```

## 5. Context Compaction (Compact)

### 5.1 Trigger Conditions

- Context approaches model limit
- Manual trigger (`/compact`)
- Automatic trigger (auto compact)

### 5.2 Compaction Flow

```
1. Analyze message history
2. Identify important information
3. Generate summary
4. Replace history messages
5. Save compaction metadata
```

### 5.3 Related Metrics

```
compact_input_tokens     # Tokens before compaction
compact_output_tokens    # Tokens after compaction
compressionOutputTokens  # Compaction output
compact_cache_read       # Cache read
compact_cache_creation   # Cache creation
compaction_triggered     # Compaction triggered
compact_manual           # Manual compaction
```

## 6. Session Recovery

### 6.1 Recovery Methods

```bash
qodercli-reverse -c                    # Continue most recent session
qodercli-reverse -r <session-id>       # Resume specified session
/resume                        # Resume within TUI
```

### 6.2 Recovery Flow

```
1. Load session file
2. Restore message history
3. Restore shell state (if any)
4. Restore file state
5. Continue conversation
```

## 7. Session Export

```bash
/export [filename]   # Export current session as text
```

Export format:
```
conversation-{timestamp}.txt
```

## 8. State Management

### 8.1 State Types

```
core/agent/state/
├── active/           # Active state
├── context/          # Context state
├── file/             # File state
│   ├── file_state/   # File state management
│   ├── snapshot/     # File snapshot
│   └── storage/      # File storage
├── memory/           # Memory management
├── session/          # Session state
├── shell/            # Shell state
├── state/            # State core
├── state_execution/  # Execution state
├── state_quest/      # Quest state
├── state_todo/       # Todo state
└── state_vars/       # State variables
```

### 8.2 File Snapshots

```
file/snapshot/
├── Record file modification history
├── Support rollback
└── Used for diff display
```

## 9. Memory System

```
core/agent/state/memory/
├── memory/          # Memory entry point
├── memory_state/    # Memory state
└── qoder_rule/      # Qoder rules
```

Memory file locations:
```
~/.qoder/memory          # Global Memory
./.qoder/memory          # Project Memory
```

## 10. Concurrent Sessions (Jobs)

### 10.1 Job Types

- **Local Job**: local process
- **Worktree Job**: Git worktree isolation
- **Container Job**: Docker container
- **Kubernetes Job**: K8s Pod

### 10.2 Job Management

```bash
qodercli-reverse jobs           # List Jobs
qodercli-reverse jobs rm <id>   # Remove Job
qodercli-reverse --worktree     # Create Worktree Job
```

### 10.3 Job Storage

```
~/.qoder/
└── worktrees/
    └── {job-id}/
```
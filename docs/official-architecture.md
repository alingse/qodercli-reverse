# Official qodercli Architecture

This document describes the package structure of the official qodercli binary, extracted through reverse engineering analysis using `strings` and `nm` tools.

## Package Structure

The official qodercli follows a well-organized package structure:

```
code.alibaba-inc.com/qoder-core/qodercli/
├── cmd/                    # Command-line interface
│   ├── jobs/              # Job management commands
│   ├── start/             # Start command implementation
│   ├── utils/             # CLI utilities
│   ├── update/            # Update command
│   └── message_io/        # Message I/O handling
│
├── core/                   # Core functionality
│   ├── agent/             # Agent implementation
│   │   └── tools/         # Tool definitions and handlers
│   ├── auth/              # Authentication
│   │   └── model/         # Auth models
│   ├── config/            # Configuration management
│   ├── types/             # Type definitions
│   ├── pubsub/            # Pub/sub messaging
│   ├── account/           # Account management
│   ├── logging/           # Logging system
│   ├── generator/         # Code generation
│   ├── resource/          # Resource management
│   │   └── mcp/           # MCP (Model Context Protocol) integration
│   └── utils/             # Core utilities
│       ├── env/           # Environment utilities
│       └── sls/           # SLS (logging service) utilities
│
├── tui/                    # Terminal UI
│   ├── components/        # UI components
│   │   └── common/        # Common components
│   │       └── textarea/  # Text area component
│   ├── event/             # Event handling
│   ├── state/             # State management
│   ├── texts/             # Text resources
│   ├── theme/             # Theme system
│   └── util/              # TUI utilities
│
├── acp/                    # ACP integration
├── sdk/                    # SDK components
└── profile/                # Profiling utilities
```

## Key Observations

### 1. Separation of Concerns

The official architecture clearly separates:
- **CLI layer** (`cmd/`): Command definitions and entry points
- **Business logic** (`core/`): Core functionality and agent implementation
- **UI layer** (`tui/`): Terminal user interface components

### 2. Core Package Organization

The `core/` package is well-structured with clear responsibilities:
- `agent/`: Agent orchestration and tool management
- `auth/`: Authentication and authorization
- `config/`: Configuration loading and management
- `types/`: Shared type definitions
- `resource/`: External resource management (MCP servers, etc.)

### 3. TUI Architecture

The TUI follows a component-based architecture:
- `components/`: Reusable UI components
- `event/`: Event-driven architecture
- `state/`: Centralized state management
- `theme/`: Consistent theming system

### 4. Tool System

Tools are organized under `core/agent/tools/` with:
- Type-safe parameter definitions
- Permission parameter types
- Response metadata types
- Consistent naming: `{Tool}Params`, `{Tool}PermissionsParams`, `{Tool}ResponseMetadata`

## Comparison with Decompiled Code

### Current Decompiled Structure

```
decompiled/
├── cmd/
│   └── root.go           # 600+ lines, contains everything
├── core/
│   ├── agent/
│   ├── config/
│   ├── log/
│   ├── provider/
│   ├── pubsub/
│   └── types/
└── tui/
    ├── app/
    └── components/
```

### Issues with Current Structure

1. **Monolithic root.go**: All CLI logic, print mode, TUI mode, provider creation, and config loading in one file
2. **Missing separation**: No clear separation between print mode and TUI mode logic
3. **Utility functions mixed**: Helper functions like `formatToolCallArgs`, `truncateString` mixed with command logic

## Recommended Refactoring

To align with the official architecture:

### Phase 1: Extract Print Mode Logic

```
cmd/
├── root.go              # CLI definition only (~150 lines)
├── print/
│   ├── print.go         # Print mode orchestration
│   ├── formatter.go     # Tool call/result formatting
│   └── callbacks.go     # Streaming callbacks
└── utils/
    ├── provider.go      # Provider creation
    └── config.go        # Config loading
```

### Phase 2: Separate TUI Initialization

```
cmd/
├── root.go
├── print/
├── tui/
│   └── tui.go          # TUI mode initialization
└── utils/
```

### Phase 3: Enhance Core Packages

```
core/
├── agent/
│   ├── agent.go
│   └── tools/          # Tool implementations
├── auth/               # Add authentication
├── resource/
│   └── mcp/           # MCP server management
└── utils/
    ├── env/           # Environment utilities
    └── format/        # String formatting utilities
```

## Benefits of Refactoring

1. **Maintainability**: Each file has a single, clear responsibility
2. **Testability**: Smaller, focused functions are easier to test
3. **Extensibility**: New features can be added without modifying existing code
4. **Readability**: Developers can quickly find relevant code
5. **Alignment**: Closer to official architecture makes future updates easier

## Implementation Priority

1. **High Priority**: Extract print mode logic (fixes UTF-8 issues are in this area)
2. **Medium Priority**: Separate TUI initialization
3. **Low Priority**: Add missing packages (auth, resource management)

## Notes

- This analysis is based on reverse engineering the official binary
- Package paths use `code.alibaba-inc.com/qoder-core/qodercli` as the base
- The official code likely has additional internal packages not visible in the binary
- Refactoring should be done incrementally to maintain stability

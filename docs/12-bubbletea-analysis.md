# Bubble Tea Usage Analysis

## Summary of Findings

Through reverse engineering of the official qodercli-reverse binary, the following key information was discovered:

### 1. Dual Dependency

qodercli-reverse depends on two versions of Bubble Tea simultaneously:

```
github.com/charmbracelet/bubbletea v1.3.4
code.alibaba-inc.com/qoder-core/bubbletea v0.0.2
```

### 2. Custom Fork

**Key finding**: qodercli-reverse uses Alibaba's internal fork of Bubble Tea:
- Package path: `code.alibaba-inc.com/qoder-core/bubbletea`
- Version: v0.0.2
- This is a custom modified version, not directly using official charmbracelet/bubbletea

### 3. Related Dependencies

Also uses other components from the Charm ecosystem:

```
github.com/charmbracelet/bubbles v0.21.0
github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc
github.com/charmbracelet/glamour
github.com/charmbracelet/lipgloss
github.com/charmbracelet/x/ansi
github.com/charmbracelet/x/cellbuf
github.com/charmbracelet/x/term
```

## TUI Component Structure

TUI package structure extracted from the binary:

```
code.alibaba-inc.com/qoder-core/qodercli-reverse/tui/
├── components/
│   ├── command/
│   │   └── github/
│   ├── askuser/
│   ├── messages/
│   ├── filepicker/
│   ├── permission/
│   ├── interaction/
│   │   ├── editor/
│   │   ├── status/
│   │   ├── progress/
│   │   └── selectors/
│   └── common/
│       ├── dialog/
│       ├── editor/
│       └── textarea/
├── event/
├── state/
├── texts/
├── theme/
└── util/
```

## Why Use a Custom Fork?

Possible reasons:

### 1. Customization Needs
- Need to modify Bubble Tea's core behavior to fit qodercli-reverse's special requirements
- May have added custom event handling or rendering logic
- From error message `bubbletea: error creating cancel reader: %w` there is a custom cancel mechanism

### 2. Internal Dependency Management
- Alibaba internally may have a unified dependency management strategy
- Need security audit and modification of open source components
- Facilitates internal version control and bug fixes

### 3. Performance Optimization
- May have performance optimizations for large-scale TUI applications
- Custom input handling and context management

## Evidence Analysis

### Error Messages

```
bubbletea: error creating cancel reader: %w
found context error while reading input: %w
```

These error messages indicate the custom version enhances:
- Cancel reader mechanism
- Context error handling

### Component Architecture

Custom components are very rich:
- **Interaction components**: editor, status, progress, selectors
- **Dialog components**: askuser, permission, dialog
- **File components**: filepicker
- **Message components**: messages
- **Command components**: command (including GitHub integration)

These components are all custom implementations built on the Bubble Tea framework.

## Impact on Reverse Engineering

### 1. Cannot Directly Use Official Bubble Tea

Our decompiled code cannot simply use `github.com/charmbracelet/bubbletea`, because:
- Official version may lack custom features
- APIs may differ
- Behavior may be inconsistent

### 2. Need to Simulate Custom Behavior

In reverse implementation, we need to:
- Use official Bubble Tea as base
- Infer custom modifications from binary behavior
- Implement compatible interfaces and behavior

### 3. Component Implementation Strategy

For TUI components:
- Can use official `github.com/charmbracelet/bubbles` as reference
- Need to implement qodercli-reverse-specific components (askuser, permission, etc.) ourselves
- Keep component structure consistent with official architecture

## Recommendations

### Short-term Plan
1. Use official `github.com/charmbracelet/bubbletea` v1.3.4
2. Implement custom features as needed
3. Focus on cancel mechanism and context handling

### Long-term Plan
1. Try to obtain or infer specific modifications in the custom fork
2. Create compatibility layer bridging official version and custom behavior
3. Fully implement all custom components

## Related Files

- `docs/official-architecture.md` - Official architecture analysis
- `decompiled/tui/` - TUI implementation directory
- `decompiled/tui/components/` - Component implementations

## Changelog

- 2026-03-10: Initial analysis, discovered custom Bubble Tea fork
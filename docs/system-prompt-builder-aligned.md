# System Prompt Builder - Official Architecture Aligned Version

## Aligned with Official qodercli-reverse Architecture

Based on reverse engineering of the official binary, strictly aligned with the official architecture design.

## Official Behavior

### Command-line Flags

The official qodercli-reverse supports the following system-prompt-related flags:

```bash
# Official flags
--system-prompt string        # Set system prompt text directly
--with-claude-config          # Load .claude/ directory configuration
-w, --workspace string        # Set working directory
```

### Official Behavior Logic

```
if --system-prompt provided:
    Use user-provided system prompt
else:
    Internally auto-build system prompt (server-side or local build)
```

## Our Implementation

### Command-line Flags (Aligned)

```bash
# Implemented (consistent with official)
--system-prompt string        # Set system prompt text directly
--with-claude-config          # Load .claude/ directory configuration
-w, --workspace string        # Set working directory

# Removed (custom flags)
# --minimal-prompt           # Removed
# --no-project-info          # Removed
# --system-prompt-file       # Removed
```

### Code Implementation Logic

```go
// cmd/root.go
systemPromptText := systemPrompt
if systemPromptText == "" {
    // Internally auto-build system prompt (aligned with official behavior)
    systemPromptText, err = utils.BuildSystemPromptAuto(workDir, withClaudeConfig)
    if err != nil {
        // Fall back to default prompt
        systemPromptText = utils.GetDefaultSystemPrompt()
    }
}
```

### Auto-build Flow

```go
// cmd/utils/system_prompt.go
func BuildSystemPromptAuto(workDir string, withClaudeConfig bool) (string, error) {
    builder := prompts.NewSystemPromptBuilderV2(vars)
    
    // Standard components (enabled by default)
    builder.WithRoleDefinition(true).
        WithToolGuide(true).
        WithPermissionRules(true).
        WithEnvironmentInfo(true).
        WithProjectContext(true).
        WithCodingStandards(true)
    
    // Collect environment info
    builder.CollectEnvironment()
    
    // Collect project context (including .claude/ config)
    builder.CollectProjectContext(workDir)
    
    // If --with-claude-config is enabled, ensure .claude/ config is loaded
    if withClaudeConfig {
        // The project context loader automatically loads the .claude/ directory
        log.Debug("Loading .claude configuration")
    }
    
    return builder.Build()
}
```

## Architecture Alignment Comparison

| Official Architecture | Our Implementation | Alignment Status |
|----------------------|--------------------|------------------|
| `--system-prompt` | `--system-prompt` | ✓ |
| `--with-claude-config` | `--with-claude-config` | ✓ |
| `-w, --workspace` | `-w, --workspace` | ✓ |
| Server-side prompt building | `BuildSystemPromptAuto()` | ✓ |
| `acp.buildPrompt` | `SystemPromptBuilderV2` | ✓ |
| `tui/texts.GetText` | `ProjectContextLoader` | ✓ |

## Usage Examples

### Official Usage

```bash
# Approach 1: Use default auto-build
qodercli-reverse

# Approach 2: Use custom system prompt
qodercli-reverse --system-prompt "You are a React expert."

# Approach 3: Enable Claude config
qodercli-reverse --with-claude-config

# Approach 4: Specify working directory
qodercli-reverse -w /path/to/project

# Approach 5: Combined usage
qodercli-reverse -w /path/to/project --with-claude-config
```

### Project-specific Instructions

The official supports auto-loading project-specific instructions from the following files:

```
project-root/
├── AGENTS.md              # Project guide (auto-loaded)
├── .claude/
│   ├── CLAUDE.md         # Claude config (loaded when --with-claude-config)
│   ├── instructions.md   # Additional instructions
│   └── context.md        # Context
└── .cursorrules          # Cursor rules (auto-loaded)
```

## Core Components

### SystemPromptBuilderV2

Reference to official `acp.(*QoderAcpAgent).buildPrompt`:

```go
type SystemPromptBuilderV2 struct {
    vars *TemplateVars
    envCollector *EnvironmentCollector
    projectLoader *ProjectContextLoader
}

func (b *SystemPromptBuilderV2) Build() (string, error) {
    // 1. Role definition
    // 2. Core instructions
    // 3. Tool guide
    // 4. Permission rules
    // 5. Environment info (OS, Git)
    // 6. Project context (AGENTS.md, .claude/)
    // 7. Coding standards
    return renderedPrompt, nil
}
```

### EnvironmentCollector

Collects environment information:
- OS, Architecture, Shell
- Git branch, commit, status
- Go, Node.js, Python, Java, Rust versions

### ProjectContextLoader

Loads project context:
- Detects project type (Go, Node, Python, Rust, Java)
- Loads AGENTS.md
- Loads .claude/ directory (when --with-claude-config)
- Loads .cursorrules
- Loads coding standard configurations

## Differences from Official Binary

| Aspect | Official | Our Implementation | Notes |
|--------|----------|-------------------|-------|
| Prompt storage | Server-side + local cache | Local build | Official fetches part of prompts from server |
| OSS loading | Supported | Not yet implemented | Official loads part of config from Alibaba Cloud OSS |
| Dynamic update | Supported | Static build | Official supports runtime prompt updates |

## Test Verification

```bash
# Compile
cd decompiled
go build ./...

# Verify flags
./qodercli-reverse --help | grep -E "(system-prompt|with-claude|workspace)"

# Test default build
echo "Hello" | ./qodercli-reverse -p "Test"

# Test custom prompt
echo "Hello" | ./qodercli-reverse -p "Test" --system-prompt "You are a Go expert."

# Test Claude config
./qodercli-reverse --with-claude-config
```

## References

- Official binary: `/Users/zhihu/.local/bin/qodercli-reverse`
- Official help: `qodercli-reverse --help`
- Architecture doc: `docs/official-architecture.md`

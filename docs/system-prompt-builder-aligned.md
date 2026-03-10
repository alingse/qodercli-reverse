# System Prompt Builder - 官方架构对齐版本

## 与官方 qodercli 架构对齐

基于官方二进制逆向分析，严格对齐官方架构设计。

## 官方行为

### 命令行标志

官方 qodercli 支持以下与系统提示词相关的标志：

```bash
# 官方标志
--system-prompt string        # 直接设置系统提示词文本
--with-claude-config          # 加载 .claude/ 目录配置
-w, --workspace string        # 设置工作目录
```

### 官方行为逻辑

```
if --system-prompt 提供了:
    使用用户提供的系统提示词
else:
    内部自动构建系统提示词（服务器端或本地构建）
```

## 我们的实现

### 命令行标志（已对齐）

```bash
# 已实现（与官方一致）
--system-prompt string        # 直接设置系统提示词文本
--with-claude-config          # 加载 .claude/ 目录配置
-w, --workspace string        # 设置工作目录

# 已移除（自定义标志）
# --minimal-prompt           # 已移除
# --no-project-info          # 已移除
# --system-prompt-file       # 已移除
```

### 代码实现逻辑

```go
// cmd/root.go
systemPromptText := systemPrompt
if systemPromptText == "" {
    // 内部自动构建系统提示词（对齐官方行为）
    systemPromptText, err = utils.BuildSystemPromptAuto(workDir, withClaudeConfig)
    if err != nil {
        // 回退到默认提示词
        systemPromptText = utils.GetDefaultSystemPrompt()
    }
}
```

### 自动构建流程

```go
// cmd/utils/system_prompt.go
func BuildSystemPromptAuto(workDir string, withClaudeConfig bool) (string, error) {
    builder := prompts.NewSystemPromptBuilderV2(vars)
    
    // 标准组件（默认启用）
    builder.WithRoleDefinition(true).
        WithToolGuide(true).
        WithPermissionRules(true).
        WithEnvironmentInfo(true).
        WithProjectContext(true).
        WithCodingStandards(true)
    
    // 收集环境信息
    builder.CollectEnvironment()
    
    // 收集项目上下文（包括 .claude/ 配置）
    builder.CollectProjectContext(workDir)
    
    // 如果启用了 --with-claude-config，确保加载 .claude/ 配置
    if withClaudeConfig {
        // 项目上下文加载器会自动加载 .claude/ 目录
        log.Debug("Loading .claude configuration")
    }
    
    return builder.Build()
}
```

## 架构对齐对比

| 官方架构 | 我们的实现 | 对齐状态 |
|----------|-----------|----------|
| `--system-prompt` | `--system-prompt` | ✓ |
| `--with-claude-config` | `--with-claude-config` | ✓ |
| `-w, --workspace` | `-w, --workspace` | ✓ |
| 服务器端构建提示词 | `BuildSystemPromptAuto()` | ✓ |
| `acp.buildPrompt` | `SystemPromptBuilderV2` | ✓ |
| `tui/texts.GetText` | `ProjectContextLoader` | ✓ |

## 使用示例

### 官方使用方式

```bash
# 方式1：使用默认自动构建
qodercli

# 方式2：使用自定义系统提示词
qodercli --system-prompt "You are a React expert."

# 方式3：启用 Claude 配置
qodercli --with-claude-config

# 方式4：指定工作目录
qodercli -w /path/to/project

# 方式5：组合使用
qodercli -w /path/to/project --with-claude-config
```

### 项目特定指令

官方支持从以下文件自动加载项目特定指令：

```
project-root/
├── AGENTS.md              # 项目指南（自动加载）
├── .claude/
│   ├── CLAUDE.md         # Claude 配置（--with-claude-config 时加载）
│   ├── instructions.md   # 额外指令
│   └── context.md        # 上下文
└── .cursorrules          # Cursor 规则（自动加载）
```

## 核心组件

### SystemPromptBuilderV2

参考官方 `acp.(*QoderAcpAgent).buildPrompt`：

```go
type SystemPromptBuilderV2 struct {
    vars *TemplateVars
    envCollector *EnvironmentCollector
    projectLoader *ProjectContextLoader
}

func (b *SystemPromptBuilderV2) Build() (string, error) {
    // 1. 角色定义
    // 2. 核心指令
    // 3. 工具指南
    // 4. 权限规则
    // 5. 环境信息（OS, Git）
    // 6. 项目上下文（AGENTS.md, .claude/）
    // 7. 编码规范
    return renderedPrompt, nil
}
```

### EnvironmentCollector

收集环境信息：
- OS, Architecture, Shell
- Git branch, commit, status
- Go, Node.js, Python, Java, Rust 版本

### ProjectContextLoader

加载项目上下文：
- 检测项目类型（Go, Node, Python, Rust, Java）
- 加载 AGENTS.md
- 加载 .claude/ 目录（--with-claude-config 时）
- 加载 .cursorrules
- 加载编码规范配置

## 与官方二进制的差异

| 方面 | 官方 | 我们的实现 | 说明 |
|------|------|-----------|------|
| 提示词存储 | 服务器端 + 本地缓存 | 本地构建 | 官方部分提示词从服务器获取 |
| OSS 加载 | 支持 | 暂未实现 | 官方从阿里云 OSS 加载部分配置 |
| 动态更新 | 支持 | 静态构建 | 官方支持运行时更新提示词 |

## 测试验证

```bash
# 编译
cd decompiled
go build ./...

# 验证标志
./qodercli --help | grep -E "(system-prompt|with-claude|workspace)"

# 测试默认构建
echo "Hello" | ./qodercli -p "Test"

# 测试自定义提示词
echo "Hello" | ./qodercli -p "Test" --system-prompt "You are a Go expert."

# 测试 Claude 配置
./qodercli --with-claude-config
```

## 参考

- 官方二进制：`/Users/zhihu/.local/bin/qodercli`
- 官方帮助：`qodercli --help`
- 架构文档：`docs/official-architecture.md`

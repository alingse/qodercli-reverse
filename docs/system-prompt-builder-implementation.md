# System Prompt Builder 实现总结

## 概述

实现了 qodercli 的 System Prompt Builder，替代原有的硬编码系统提示词：

```go
// 之前
systemPrompt := "You are a helpful AI assistant."

// 现在
builder := prompts.NewSystemPromptBuilderV2(vars)
prompt, err := builder.Build() // 包含角色定义、工具指南、权限规则、环境信息、项目上下文
```

## 实现文件

### Core Prompts 包 (`decompiled/core/prompts/`)

| 文件 | 功能 | 行数 |
|------|------|------|
| `prompts.go` | 核心类型定义 (Prompt, TemplateVars, Manager 接口) | ~280 |
| `builtin.go` | 内置提示词 (主 Agent、子 Agent、IDE 集成等 30+ 类型) | ~680 |
| `manager.go` | 提示词管理器实现 (加载、注册、渲染) | ~260 |
| `registry.go` | 提示词注册表和组合器 | ~280 |
| `helpers.go` | 辅助函数 (便捷获取、系统提醒、思考触发词) | ~290 |
| `builder.go` | **SystemPromptBuilderV2 核心实现** | ~580 |
| `env_collector.go` | 环境信息收集器 (OS, Git, 开发环境) | ~130 |
| `project_loader.go` | 项目上下文加载器 (AGENTS.md, .claude/, .cursorrules) | ~430 |
| `README.md` | 使用文档 | ~340 |

### 集成文件

| 文件 | 修改内容 |
|------|----------|
| `cmd/utils/system_prompt.go` | 封装 System Prompt Builder 使用 |
| `cmd/utils/config.go` | 添加 Flags 字段 |
| `cmd/root.go` | 集成新的系统提示词构建流程 |
| `cmd/print/print.go` | 使用 BuildSystemPromptFromFlags |
| `cmd/tui/tui.go` | 使用 BuildSystemPromptFromFlags |

## 核心功能

### 1. 模块化提示词构建

```go
builder := prompts.NewSystemPromptBuilderV2(vars)
builder.
    WithRoleDefinition(true).      // 角色定义
    WithToolGuide(true).           // 工具使用指南
    WithPermissionRules(true).     // 权限规则
    WithEnvironmentInfo(true).     // 环境信息 (OS, Git)
    WithProjectContext(true).      // 项目上下文 (AGENTS.md)
    WithCodingStandards(true).     // 编码规范
    Build()
```

### 2. 项目上下文自动检测

自动检测并加载：
- **项目类型**: Go, Node, Python, Rust, Java
- **项目指令**: AGENTS.md, .claude/CLAUDE.md, .cursorrules
- **编码规范**: .golangci.yml, .eslintrc, prettier.config.js
- **README 摘要**: 自动提取项目描述

### 3. 环境信息收集

自动收集：
- **系统**: OS, Architecture, Shell, Working Directory
- **Git**: Branch, Commit, Remote, Status (modified/untracked)
- **开发环境**: Go, Node.js, Python, Java, Rust 版本

### 4. 命令行接口

```bash
# 默认模式 - 自动构建包含项目上下文的提示词
qodercli

# 最小化模式 - 不包含项目信息
qodercli --minimal-prompt

# 禁用项目信息
qodercli --no-project-info

# 使用自定义提示词文件
qodercli --system-prompt-file ./custom.md
```

## 架构对齐

与官方二进制分析对齐：

```
官方路径                                    我们的实现
────────────────────────────────────────────────────────────────
acp.(*QoderAcpAgent).buildPrompt       →   SystemPromptBuilderV2.Build()
core/agent.GetSystemPrompt             →   utils.BuildSystemPrompt()
core/resource/command.loadBuiltinPrompts → builtin.go
tui/texts.(*Service).GetText           →   ProjectContextLoader.Load()
```

## 使用示例

### 基础使用

```go
// 使用默认配置
config := utils.DefaultSystemPromptConfig()
prompt, err := utils.BuildSystemPrompt(config)
```

### 自定义配置

```go
config := &utils.SystemPromptConfig{
    AppName:              "mycli",
    EnableProjectContext: true,
    EnableEnvironment:    true,
    WorkDir:              "/path/to/project",
    CustomInstructions:   "Always write tests first.",
}
prompt, err := utils.BuildSystemPrompt(config)
```

### 手动构建

```go
vars := prompts.DefaultTemplateVars()
vars.AppName = "mycli"

builder := prompts.NewSystemPromptBuilderV2(vars)
builder.WithCustomRole("You are a React expert.")
builder.AddCustomSection("Custom Rules", "1. Use TypeScript\n2. Write tests", 100)

prompt, err := builder.Build()
```

## 内置提示词类型

### 主 Agent 类型
- `PromptTypeMainAgent` - 默认主 Agent
- `PromptTypeMainAgentEducational` - 教育版
- `PromptTypeMainAgentPractice` - 实践学习版
- `PromptTypeGenericTask` - 通用任务

### 子 Agent 类型 (18种)
- `PromptTypeBrowserSubagent` - 浏览器自动化
- `PromptTypeCodeImplement` - 代码实现
- `PromptTypeTaskExecutor` - 任务执行专家
- `PromptTypeDesignAgent` - 设计 Agent
- `PromptTypeSystemDesign` - 系统设计
- `PromptTypeSoftwareArchitect` - 软件架构师
- `PromptTypeDesignReview` - 设计评审
- `PromptTypeRequirements` - 需求分析
- `PromptTypeTestAutomation` - 自动化测试
- `PromptTypeCodeReviewer` - 代码审查
- `PromptTypeDebugger` - 调试专家
- `PromptTypeFileSearch` - 文件搜索
- `PromptTypeWorkflowOrchestrator` - 工作流编排
- `PromptTypeBehaviorAnalyzer` - 行为分析
- `PromptTypeSkepticalValidator` - 怀疑验证器
- `PromptTypeSecurityAuditor` - 安全审计
- `PromptTypeDataScientist` - 数据科学家
- `PromptTypeGuideAgent` - 引导 Agent
- `PromptTypeQuestHandler` - Quest 处理器

### IDE 集成类型
- `PromptTypeQoderWork` - QoderWork
- `PromptTypeQoderStudio` - Qoder Studio
- `PromptTypeQoderDesktop` - Qoder 桌面版

## 模板变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `{{.AppName}}` | 应用名称 | qodercli |
| `{{.BrandName}}` | 品牌名称 | Qoder |
| `{{.ProductName}}` | 产品名称 | Qoder CLI |
| `{{.ReadToolName}}` | 读取工具名 | Read |
| `{{.BashToolName}}` | Bash 工具名 | Bash |
| `{{.SearchCodebaseTool}}` | 搜索工具名 | SearchCodebase |
| `{{.SearchSymbolTool}}` | 符号搜索工具名 | SearchSymbol |

## 项目特定指令文件

支持以下文件自动加载：

```
project-root/
├── AGENTS.md              # 项目指南
├── .claude/
│   ├── CLAUDE.md         # Claude 配置
│   ├── instructions.md   # 额外指令
│   └── context.md        # 上下文
├── .cursorrules          # Cursor 规则
├── CONTRIBUTING.md       # 贡献指南
└── STYLE.md              # 风格指南
```

## 测试验证

```bash
# 编译验证
cd decompiled
go build ./...

# 运行帮助查看新标志
./qodercli --help

# 测试系统提示词构建
echo "Test" | ./qodercli --debug -p "Hello"
```

## 后续优化方向

1. **缓存机制**: 缓存项目上下文和环境信息，避免重复收集
2. **远程加载**: 支持从远程 URL 加载提示词模板
3. **YAML 支持**: 完整的 YAML frontmatter 解析
4. **多语言**: 根据用户语言自动切换提示词语言
5. **版本控制**: 提示词版本管理和迁移

## 参考文档

- `extracted_system_prompts.md` - 从官方二进制提取的提示词 v1
- `extracted_system_prompts_v2.md` - 提取的提示词 v2（完整版）
- `docs/official-architecture.md` - 官方架构分析
- `decompiled/core/prompts/README.md` - 使用文档

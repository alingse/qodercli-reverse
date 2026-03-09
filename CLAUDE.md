# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是 qodercli (v0.1.29) 的逆向分析项目，目标是还原其架构和实现。qodercli 是一个 AI 编程助手 CLI 工具，支持多种 LLM Provider、MCP 协议、工具系统和 TUI 界面。

## 核心文档

**必读文档**（按优先级）：

1. `docs/official-architecture.md` - 官方架构分析（重构参考）
2. `docs/architecture-overview.md` - 完整架构概览
3. `docs/11-decompiled-gap-analysis.md` - 反编译差异分析
4. `README.md` - 项目说明和开发工作流程

**模块文档**：
- `docs/01-package-structure.md` - 包结构和依赖
- `docs/02-cli-commands.md` - CLI 命令系统
- `docs/03-llm-integration.md` - LLM 集成
- `docs/04-mcp-protocol.md` - MCP 协议
- `docs/05-tools-system.md` - 工具系统
- `docs/06-session-management.md` - 会话管理
- `docs/07-permission-security.md` - 权限安全
- `docs/08-config-api.md` - 配置和 API
- `docs/09-github-integration.md` - GitHub 集成
- `docs/10-subagent-skill.md` - SubAgent 和 Skill
- `docs/12-bubbletea-analysis.md` - Bubble Tea 使用分析

## 开发工作流程

**重要**：在修改或扩展代码前，必须先进行逆向分析，确保实现与官方架构一致。

### 1. 定位官方二进制

```bash
which qodercli
# 输出示例: /Users/zhihu/.qoder/bin/qodercli/qodercli-0.1.29
```

### 2. 逆向分析目标模块

```bash
# 提取符号表
nm -gU $(which qodercli) | grep "关键字"

# 提取字符串
strings $(which qodercli) | grep "关键字"

# 反汇编特定函数
go tool objdump -s "函数名" $(which qodercli)
```

### 3. 对照官方架构

参考 `docs/official-architecture.md` 中的包结构，确保修改符合官方的分层架构。

### 4. 实现与验证

- 按照官方架构组织代码
- 使用官方命名约定（如 `{Tool}Params`）
- 实现后对比官方二进制行为

## 构建和测试

### 编译 decompiled 代码

```bash
cd decompiled
go mod tidy
go build -o qodercli .
```

**注意**：完整代码可能无法直接编译，详见 `COMPILATION_GUIDE.md`。

### 运行测试

```bash
cd decompiled
go test ./...
```

### 逆向分析命令

```bash
# 查看二进制信息
file $(which qodercli)
otool -L $(which qodercli)

# 符号表分析
nm -gU $(which qodercli) | grep "mcp"
nm $(which qodercli) | wc -l

# 字符串提取
strings $(which qodercli) | grep "code.alibaba-inc.com"

# Go 特定分析
go tool nm $(which qodercli) | grep "包名"
```

## 架构概览

```
┌─────────────────┐
│   CLI / TUI     │  ← cmd/ (Cobra + 自定义 Bubble Tea fork)
└────────┬────────┘
         │
┌────────▼────────┐
│   Agent Core    │  ← core/agent/ (主循环、Provider、Tools、Permission)
└────────┬────────┘
         │
┌────────▼────────┐
│   Resources     │  ← core/resource/ (MCP、Skill、SubAgent)
└────────┬────────┘
         │
┌────────▼────────┐
│   Utilities     │  ← core/utils/ (HTTP、Token、Storage、API)
└────────┬────────┘
```

**关键模块**：
- **多 Provider 支持**：Qoder、OpenAI、Anthropic、IdeaLab、DashScope
- **工具系统**：Bash、Read、Write、Edit、Grep、Glob 等
- **MCP 协议**：stdio/SSE 传输，可扩展工具资源
- **权限系统**：细粒度规则引擎，Hook 拦截
- **运行模式**：交互/非交互/SDK/容器化
- **TUI 框架**：使用阿里内部 fork 的 Bubble Tea (code.alibaba-inc.com/qoder-core/bubbletea v0.0.2)

## 代码组织

```
decompiled/
├── cmd/                    # CLI 命令入口
│   ├── root.go            # 主命令
│   ├── print/             # Print 模式
│   ├── tui/               # TUI 模式
│   └── utils/             # CLI 工具函数
├── core/                   # 核心逻辑
│   ├── agent/             # Agent 系统
│   │   ├── agent/         # 主循环
│   │   ├── provider/      # LLM Provider
│   │   ├── tools/         # 工具实现
│   │   ├── permission/    # 权限控制
│   │   ├── state/         # 状态管理
│   │   └── compact/       # 上下文压缩
│   ├── config/            # 配置加载
│   ├── resource/mcp/      # MCP 客户端
│   ├── pubsub/            # 事件系统
│   ├── log/               # 日志系统
│   └── types/             # 类型定义
├── tui/                    # TUI 界面
│   ├── app/               # 应用启动
│   └── components/        # UI 组件
└── main.go                # 程序入口
```

## 重要约定

1. **包导入路径**：使用 `github.com/alingse/qodercli-reverse/decompiled` 作为模块路径
2. **命名约定**：工具参数类型使用 `{Tool}Params` 格式
3. **架构对齐**：所有修改必须参考 `docs/official-architecture.md`
4. **逆向优先**：先分析官方实现，再编写代码

## 常见任务

### 添加新工具

1. 在 `decompiled/core/agent/tools/` 创建新文件
2. 实现 `Tool` 接口
3. 在 `tools.go` 注册工具
4. 参考官方二进制中的工具实现

### 修改 Provider

1. 查看 `decompiled/core/agent/provider/`
2. 参考 `docs/03-llm-integration.md`
3. 使用 `nm` 分析官方 Provider 实现

### 扩展 MCP 功能

1. 查看 `decompiled/core/resource/mcp/`
2. 参考 `docs/04-mcp-protocol.md`
3. 分析官方 MCP 客户端实现

## 注意事项

- 反编译代码基于二进制推导，可能与原始实现有差异
- 代码未经完整测试，不建议用于生产环境
- 仅供学习和架构参考使用
- 修改代码前务必先阅读相关文档和进行逆向分析

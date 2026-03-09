# qodercli 逆向分析项目

本项目对 Qoder CLI (v0.1.29) 进行深度逆向分析，还原其代码架构、模块设计和运行机制。

## 项目目标

- **分析对象**: qodercli v0.1.29 (Mach-O arm64 Go binary, ~37MB)
- **分析方法**: 符号表分析、字符串提取、行为观察、配置分析
- **输出物**: 完整架构文档 + 反编译代码示例

## 当前进展

已完成全部逆向分析任务：

| Phase | 任务 | 状态 |
|-------|------|------|
| Phase 1 | 基础信息提取（包结构、依赖图） | ✅ |
| Phase 2 | 核心模块分析（CLI/AI/MCP/Tools/Session/Permission/GitHub/SubAgent） | ✅ |
| Phase 3 | 整合输出（架构概览文档） | ✅ |

**完成时间**: 2026-03-09

## 文档目录

```
docs/
├── architecture-overview.md      # 完整架构概览
├── official-architecture.md      # 官方 qodercli 架构分析（重构参考）
├── 01-package-structure.md       # 包结构和依赖
├── 02-cli-commands.md            # CLI 命令系统
├── 03-llm-integration.md         # LLM 集成
├── 04-mcp-protocol.md            # MCP 协议
├── 05-tools-system.md            # 工具系统
├── 06-session-management.md      # 会话管理
├── 07-permission-security.md     # 权限安全
├── 08-config-api.md              # 配置和 API
├── 09-github-integration.md      # GitHub 集成
├── 10-subagent-skill.md          # SubAgent 和 Skill
├── 11-decompiled-gap-analysis.md # 反编译差异分析
└── 12-bubbletea-analysis.md      # Bubble Tea 使用分析
```

## decompiled 目录架构

```
decompiled/
├── cmd/                          # CLI 命令入口
│   └── root.go
├── core/                         # 核心逻辑
│   ├── agent/                    # Agent 系统
│   │   ├── agent/agent.go        # 主循环
│   │   ├── provider/             # LLM Provider
│   │   ├── tools/                # 工具实现
│   │   ├── permission/           # 权限控制
│   │   └── state/                # 状态管理
│   ├── config/                   # 配置加载
│   ├── resource/mcp/             # MCP 客户端
│   ├── pubsub/                   # 事件系统
│   └── types/                    # 类型定义
├── tui/                          # TUI 界面
│   ├── app/                      # 应用启动
│   └── components/               # UI 组件
├── main.go                       # 程序入口
└── go.mod/go.sum                 # 依赖管理
```

**代码统计**: 约 15 个文件，~3350 行代码

## 核心架构

```
┌─────────────────┐
│   CLI / TUI     │
└────────┬────────┘
         │
┌────────▼────────┐
│   Agent Core    │  ← 主循环、Provider、Tools、Permission
└────────┬────────┘
         │
┌────────▼────────┐
│   Resources     │  ← MCP、Skill、SubAgent
└────────┬────────┘
         │
┌────────▼────────┐
│   Utilities     │  ← HTTP、Token、Storage、API
└────────┬────────┘
```

## 关键发现

| 模块 | 说明 |
|------|------|
| **多 Provider** | 支持 Qoder/OpenAI/Anthropic/IdeaLab/DashScope |
| **工具系统** | Bash/Read/Write/Edit/Grep/Glob/Delete 等 |
| **MCP 协议** | stdio/SSE 传输，可扩展工具资源 |
| **权限系统** | 细粒度规则引擎，Hook 拦截 |
| **GitHub 集成** | PR 创建、Actions、Code Review |
| **运行模式** | 交互/非交互/SDK/容器化 |

## 开发工作流程

在修改或扩展代码前，必须遵循以下逆向分析流程，确保实现与官方架构保持一致：

### 1. 定位官方二进制

```bash
# 找到官方 qodercli 二进制文件位置
which qodercli
# 输出示例: /Users/zhihu/.qoder/bin/qodercli/qodercli-0.1.29
```

### 2. 逆向分析目标模块

使用反编译工具分析官方实现的函数、方法、调用关系：

```bash
# 提取符号表，查找相关包和函数
nm -gU /path/to/qodercli | grep "包名或功能关键字"

# 提取字符串，定位功能模块
strings /path/to/qodercli | grep "关键字"

# 使用 Go 工具反汇编特定函数
go tool objdump -s "函数名" /path/to/qodercli

# 示例：分析 MCP 相关实现
nm -gU $(which qodercli) | grep mcp
strings $(which qodercli) | grep -i "mcp"
```

### 3. 对照官方架构

参考 `docs/official-architecture.md` 中的官方包结构：

```
code.alibaba-inc.com/qoder-core/qodercli/
├── cmd/                    # 命令行接口层
├── core/                   # 核心业务逻辑
│   ├── agent/             # Agent 实现
│   ├── auth/              # 认证
│   ├── config/            # 配置管理
│   └── resource/mcp/      # MCP 集成
└── tui/                    # 终端 UI
```

确保你的修改符合官方的分层架构和职责划分。

### 4. 规划实现方案

在动手编码前：

1. **阅读相关文档**: 查看 `docs/` 中对应模块的分析文档
2. **查看差距分析**: 参考 `docs/11-decompiled-gap-analysis.md` 了解当前缺失的功能
3. **制定计划**: 明确需要新建/修改哪些文件，如何组织代码结构
4. **验证设计**: 确保方案与官方架构一致，避免偏离

### 5. 实现与验证

- 按照官方架构的包结构组织代码
- 使用官方的命名约定（如工具参数类型：`{Tool}Params`）
- 实现后对比官方二进制的行为，确保功能一致

## 使用方法

1. **查看架构**: 从 `docs/architecture-overview.md` 开始
2. **官方架构参考**: 查看 `docs/official-architecture.md` 了解官方代码结构（用于重构参考）
3. **深入模块**: 阅读对应模块文档 (01-11)
4. **参考代码**: 查看 `decompiled/` 中的反编译示例

## 重构计划

当前 `decompiled/cmd/root.go` 包含 600+ 行代码，混合了多种职责。参考 `docs/official-architecture.md` 中的官方架构，建议进行以下重构：

**重要**: 每个 Phase 开始前，必须先执行"开发工作流程"中的逆向分析步骤：
- 使用 `which qodercli` 定位官方二进制
- 使用 `nm`、`strings`、`go tool objdump` 分析官方实现
- 对照官方架构规划方案

### 重构阶段

1. **Phase 1**: 提取 Print Mode 逻辑到 `cmd/print/` 包
   - 分析官方 `cmd/start/` 和 `cmd/message_io/` 的实现
   - 提取格式化和流式输出逻辑

2. **Phase 2**: 分离 TUI 初始化到 `cmd/tui/` 包
   - 分析官方 TUI 启动流程
   - 分离交互模式和非交互模式

3. **Phase 3**: 提取工具函数到 `cmd/utils/` 包
   - 对照官方 `cmd/utils/` 包结构
   - 提取 Provider 创建和配置加载逻辑

详见 [官方架构文档](docs/official-architecture.md) 和 [差距分析](docs/11-decompiled-gap-analysis.md)。

## 注意事项

- 反编译代码基于二进制推导，可能与原始实现有差异
- 代码未经测试，不建议直接用于生产环境
- 仅供学习和架构参考使用

## 常用逆向分析命令

```bash
# 查看二进制基本信息
file $(which qodercli)
otool -L $(which qodercli)  # 查看依赖库

# 符号表分析
nm -gU $(which qodercli) | grep "关键字"  # 查找导出符号
nm $(which qodercli) | wc -l              # 统计符号数量

# 字符串提取
strings $(which qodercli) | grep "关键字"
strings $(which qodercli) | grep "code.alibaba-inc.com"  # 查找包路径

# Go 特定分析
go tool nm $(which qodercli) | grep "包名"
go tool objdump -s "函数名" $(which qodercli)

# 查看帮助信息（了解 CLI 结构）
qodercli --help
qodercli mcp --help
qodercli jobs --help
```

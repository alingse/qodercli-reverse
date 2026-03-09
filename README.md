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
└── 11-decompiled-gap-analysis.md # 反编译差异分析
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

## 使用方法

1. **查看架构**: 从 `docs/architecture-overview.md` 开始
2. **深入模块**: 阅读对应模块文档 (01-11)
3. **参考代码**: 查看 `decompiled/` 中的反编译示例

## 注意事项

- 反编译代码基于二进制推导，可能与原始实现有差异
- 代码未经测试，不建议直接用于生产环境
- 仅供学习和架构参考使用

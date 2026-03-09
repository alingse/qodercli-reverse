# qodercli 逆向分析任务规划

## 项目目标

对 qodercli (v0.1.29, Mach-O arm64 Go binary) 进行深度逆向分析，还原其可能的代码架构、模块设计和运行机制。

## 分析对象

- **二进制文件**: `/Users/zhihu/.qoder/bin/qodercli/qodercli-0.1.29`
- **文件类型**: Mach-O 64-bit executable arm64
- **编程语言**: Go 1.25.7
- **文件大小**: ~37.2 MB

## 任务清单

### Phase 1: 基础信息提取
- [x] 任务 1.1: 创建任务规划文档和 docs/ 目录结构
- [x] 任务 1.2: 深度分析二进制符号表和包结构 — 提取所有 Go 包路径，绘制完整依赖图

### Phase 2: 核心模块分析
- [x] 任务 2.1: CLI 命令系统和入口流程 — cobra 命令树、Flag 定义、子命令关系
- [x] 任务 2.2: AI/LLM 集成模块 — 模型路由、流式传输、消息协议
- [x] 任务 2.3: MCP 协议实现 — 传输层、工具注册、资源管理
- [x] 任务 2.4: 工具系统 (Tools) — Bash/Read/Write/Edit/Grep/Glob 等工具实现
- [x] 任务 2.5: 会话管理和持久化 — 会话生命周期、压缩、导出
- [x] 任务 2.6: 权限和安全机制 — 权限规则引擎、Hook 系统
- [x] 任务 2.7: 配置系统和 API 通信 — 配置文件加载、API 客户端
- [x] 任务 2.8: GitHub 集成和 CI/CD 功能 — PR/Review/Actions
- [x] 任务 2.9: SubAgent 和 Skill 系统 — 子代理调度、技能加载

### Phase 3: 整合输出
- [x] 任务 3.1: 整合分析结果，生成完整架构文档

---

**状态**: 所有任务已完成 (2026-03-09)

## 输出物

每个分析任务完成后在 `docs/` 下生成对应文档：

```
docs/
├── PLAN.md                      # 本文件 — 任务规划
├── 01-package-structure.md      # 包结构和依赖图
├── 02-cli-commands.md           # CLI 命令系统
├── 03-llm-integration.md       # AI/LLM 集成
├── 04-mcp-protocol.md          # MCP 协议实现
├── 05-tools-system.md          # 工具系统
├── 06-session-management.md    # 会话管理
├── 07-permission-security.md   # 权限和安全
├── 08-config-api.md            # 配置和 API
├── 09-github-integration.md    # GitHub 集成
├── 10-subagent-skill.md        # SubAgent 和 Skill
└── architecture-overview.md    # 完整架构概览 (最终输出)
```

## 分析方法

1. **符号表分析**: `nm`, `go tool objdump` 提取函数和包信息
2. **字符串分析**: 提取嵌入字符串定位功能模块
3. **行为观察**: 运行时观察文件操作、网络通信、进程行为
4. **配置分析**: 分析相关配置文件格式和内容
5. **交叉引用**: 结合已知的 system prompt 信息辅助分析

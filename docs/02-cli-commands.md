# qodercli CLI 命令系统分析

## 1. 入口与架构

### 1.1 入口点

```
main.main -> cmd/root.go -> cobra.Command
```

入口函数位于 `main` 包，使用 `github.com/spf13/cobra` 框架构建 CLI。

### 1.2 命令层级结构

```
qodercli (root)
├── [flags]                      # 全局标志
├── jobs                         # Job 管理命令组
│   └── rm                       # 删除 Job
├── mcp                          # MCP 服务器管理命令组
│   ├── add                      # 添加 MCP 服务器
│   ├── auth                     # MCP OAuth 认证
│   ├── get                      # 获取 MCP 服务器详情
│   ├── list                     # 列出所有 MCP 服务器
│   └── remove                   # 移除 MCP 服务器
├── completion                   # Shell 自动补全生成
├── feedback                     # 提交反馈
├── status                       # 显示账户和 CLI 状态
└── update                       # 自更新到最新版本
```

## 2. 全局 Flags

| Flag | 简写 | 类型 | 说明 |
|------|------|------|------|
| `--workspace` | `-w` | string | 工作目录 |
| `--version` | `-v` | - | 显示版本 |
| `--print` | `-p` | string | 非交互模式，打印响应后退出 |
| `--output-format` | `-f` | string | 输出格式 (text/json/stream-json) |
| `--input-format` | - | string | 输入格式 (text/stream-json) |
| `--resume` | `-r` | string | 恢复指定会话 |
| `--continue` | `-c` | - | 继续最近会话 |
| `--model` | - | string | 模型级别选择 |
| `--max-turns` | - | int | 最大 agent 循环次数 |
| `--max-output-tokens` | - | string | 最大输出 token (16k/32k) |
| `--attachment` | - | stringArray | 附件路径 |
| `--agents` | - | string | JSON 自定义 agent 定义 |
| `--allowed-tools` | - | stringArray | 允许的工具列表 |
| `--disallowed-tools` | - | stringArray | 禁用的工具列表 |
| `--dangerously-skip-permissions` | - | - | 跳过所有权限检查 |
| `--yolo` | - | - | 同上 (别名) |
| `--branch` | - | string | 分支名 (配合 --worktree) |
| `--path` | - | string | worktree 路径 |
| `--worktree` | - | - | 通过 git worktree 启动并发 Job |
| `--with-claude-config` | - | - | 加载 Claude Code 配置 |
| `--quiet` | `-q` | - | 隐藏 spinner |

## 3. 模型选择选项

`--model` 参数支持以下值：

| 值 | 说明 |
|----|------|
| `auto` | 自动选择 |
| `efficient` | 高效模式 |
| `performance` | 性能模式 |
| `ultimate` | 终极模式 |
| `lite` | 轻量模式 |
| `qmodel` | Q 模型 |
| `q35model` | Q3.5 模型 |
| `gmodel` | G 模型 |
| `kmodel` | K 模型 |
| `mmodel` | M 模型 |

## 4. TUI 内置斜杠命令

在交互模式下，支持以下斜杠命令（定义于 `core/resource/command/`）：

### 4.1 会话管理
| 命令 | 说明 |
|------|------|
| `/clear` | 清除对话历史 |
| `/resume` | 恢复历史会话 |
| `/export [filename]` | 导出当前会话 |
| `/compact` | 压缩上下文 |

### 4.2 账户与认证
| 命令 | 说明 |
|------|------|
| `/login` | 登录 |
| `/logout` | 登出 |
| `/status` | 显示状态 |
| `/usage` | 显示用量 |

### 4.3 模型与配置
| 命令 | 说明 |
|------|------|
| `/model` | 切换模型 |
| `/config` | 配置管理 |
| `/prompt` | Prompt 管理 |
| `/memory` | Memory 管理 |

### 4.4 Agent 与 Skill
| 命令 | 说明 |
|------|------|
| `/agents` | Agent 列表和管理 |
| `/skills` | Skill 列表和管理 |
| `/bashes` | 后台 Bash 进程管理 |

### 4.5 Quest 模式
| 命令 | 说明 |
|------|------|
| `/quest-on` | 进入 Quest 模式 |
| `/quest-off` | 退出 Quest 模式 |

### 4.6 GitHub 集成
| 命令 | 说明 |
|------|------|
| `/github` | GitHub 操作 |
| `/setup-github` | 设置 GitHub 集成 |

### 4.7 其他
| 命令 | 说明 |
|------|------|
| `/help` | 帮助信息 |
| `/quit` | 退出 |
| `/vim` | Vim 模式切换 |
| `/upgrade` | 升级 CLI |
| `/release-notes` | 查看发布说明 |
| `/feedback` | 提交反馈 |

## 5. 环境变量

### 5.1 认证相关
| 变量 | 说明 |
|------|------|
| `QODER_PERSONAL_ACCESS_TOKEN` | Qoder 个人访问令牌 |
| `QODER_AUTH_TOKEN` | 认证 Token |
| `QODER_ANTHROPIC_API_KEY` | Anthropic API Key |
| `QODER_OPENAI_API_KEY` | OpenAI API Key |
| `QODER_DASHSCOPE_API_KEY` | DashScope API Key |
| `QODER_IDEALAB_API_KEY` | IdeaLab API Key |
| `QODER_OPENAI_BASE_URL` | OpenAI API Base URL |

### 5.2 运行时配置
| 变量 | 说明 |
|------|------|
| `QODER_ENV` | 运行环境 |
| `QODER_CLI` | CLI 模式标识 (=1) |
| `QODER_CLI_INSTALL` | CLI 安装标识 |
| `QODER_CLI_RUNNING_MODEL` | 当前运行模型 |
| `QODER_MACHINE_ID` | 机器唯一标识 |
| `QODER_ENTRYPOINT` | 入口点 |
| `QODER_WORK_VM` | Work VM 配置 |

### 5.3 工具配置
| 变量 | 说明 |
|------|------|
| `QODER_BASH_PATH` | Bash 路径 |
| `QODER_BASH_TIMEOUT` | Bash 超时时间 |
| `QODER_PROJECT_DIR` | 项目目录 |
| `QODER_CURRENT_WORKDIR` | 当前工作目录 |

### 5.4 调试与显示
| 变量 | 说明 |
|------|------|
| `QODER_SHOW_TOKEN_USAGE` | 显示 Token 用量 |
| `QODER_EXPOSE_TOKEN_USAGE` | 暴露 Token 用量 |
| `QODER_CHANGELOG_URL` | Changelog URL |
| `QODER_ALIBABA_PREVIEW` | 阿里巴巴预览模式 |
| `QODER_AGENT_SDK_VERSION` | Agent SDK 版本 |
| `QODER_USER_INFO` | 用户信息 |

## 6. 运行模式

### 6.1 交互模式 (默认)
```bash
qodercli                    # 启动交互式 TUI
qodercli -w /path/to/repo   # 指定工作目录
```

### 6.2 非交互模式 (Print)
```bash
qodercli -p "解释 Go context 的用法"
qodercli -p "..." -f json           # JSON 输出
qodercli -p "..." --max-turns 25    # 限制循环次数
```

### 6.3 SDK 模式
通过 stdin/stdout 进行 JSON 协议通信，用于 IDE 集成。

### 6.4 ACP Server 模式
作为 ACP (Agent Communication Protocol) 服务器运行。

### 6.5 Worktree 模式
```bash
qodercli --worktree --branch feature-x
```

### 6.6 Container/Kubernetes 模式
```bash
qodercli --kubeconfig ...
```

## 7. 启动流程

```
1. cmd/root.go: 解析命令行参数
2. cmd/start/start_local.go (或其他启动器):
   - 初始化配置 (core/config)
   - 初始化日志 (core/logging)
   - 初始化认证 (core/auth)
   - 加载 MCP 服务器 (core/resource/mcp)
   - 加载 Skills (core/resource/skill)
   - 加载 SubAgents (core/resource/subagent)
   - 加载 Hooks (core/resource/hook)
   - 加载 Plugins (core/resource/plugin)
3. tui/app: 启动 TUI 应用
   - 初始化 Bubble Tea 程序
   - 设置事件订阅 (core/pubsub)
4. core/agent: Agent 主循环
   - 等待用户输入
   - 调用 Provider 进行推理
   - 执行工具调用
   - 更新状态
```

## 8. 更新机制

支持多种更新方式 (`cmd/update/`):

| 方式 | 包 |
|------|-----|
| Homebrew | `homebrew_updater` |
| NPM | `npm_updater` |
| curl\|bash | `curlbash_updater` |
| 自动更新 | `auto` |

## 9. Job 管理系统

Job 是并发执行的 Agent 实例:

```
jobs/
├── jobs.go      # 列出 Jobs
├── attach/      # 附加到 Job
├── fetch/       # 获取 Job 状态
├── rm/          # 删除 Job
└── stop/        # 停止 Job
```

Job 类型:
- **Local Job**: 本地进程
- **Worktree Job**: Git worktree 隔离
- **Container Job**: Docker 容器
- **Kubernetes Job**: K8s Job

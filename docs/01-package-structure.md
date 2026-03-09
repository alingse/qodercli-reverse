# qodercli 包结构与依赖分析

## 1. 内部 Go Module 路径

```
code.alibaba-inc.com/qoder-core/qodercli
```

> 这是 qodercli 的内部 Go module 路径，表明项目托管在阿里巴巴内部 GitLab 上。

## 2. 完整包结构树

以下是从二进制符号表中提取的完整内部包列表（共约 250+ 子包）：

```
code.alibaba-inc.com/qoder-core/qodercli/
│
├── cmd/                                    # CLI 命令层 (入口)
│   ├── root/                              # 根命令定义
│   ├── start/                             # 启动命令
│   │   ├── cleanup/                       # 资源清理
│   │   ├── job_startup/                   # Job 启动
│   │   ├── output_format/                 # 输出格式化
│   │   ├── start_container/               # 容器模式启动
│   │   ├── start_kubernetes/              # Kubernetes 模式启动
│   │   ├── start_local/                   # 本地模式启动
│   │   ├── start_worktree/                # Git worktree 模式启动
│   │   └── subscriptions/                 # 事件订阅
│   ├── update/                            # 自更新命令
│   │   ├── auto/                          # 自动更新
│   │   ├── common/                        # 更新公共逻辑
│   │   ├── curlbash_updater/              # curl|bash 方式更新
│   │   ├── homebrew_updater/              # Homebrew 方式更新
│   │   ├── npm_updater/                   # NPM 方式更新
│   │   └── update/                        # 更新入口
│   ├── mcp/                               # MCP 服务器管理命令
│   │   ├── add/                           # 添加 MCP 服务器
│   │   ├── auth/                          # MCP OAuth 认证
│   │   ├── get/                           # 获取 MCP 服务器信息
│   │   ├── list/                          # 列出 MCP 服务器
│   │   ├── mcp/                           # MCP 命令入口
│   │   └── remove/                        # 移除 MCP 服务器
│   ├── jobs/                              # 远程 Job 管理命令
│   │   ├── attach/                        # 附加到 Job
│   │   ├── fetch/                         # 获取 Job
│   │   ├── jobs/                          # Jobs 命令入口
│   │   ├── rm/                            # 删除 Job
│   │   └── stop/                          # 停止 Job
│   ├── feedback/                          # 反馈命令
│   │   └── feedback/
│   ├── machineinfo/                       # 机器信息命令
│   │   └── machine_info/
│   ├── status/                            # 状态命令
│   │   └── status/
│   ├── message_io/                        # 消息输入输出 (SDK 模式)
│   │   ├── out/                           # 输出处理
│   │   ├── streaming/                     # 流式处理
│   │   └── transform/                     # 消息转换
│   └── utils/                             # 命令工具
│       ├── archive/                       # 压缩归档
│       ├── docker/                        # Docker 工具
│       ├── flags/                         # 命令行 Flag
│       ├── git/                           # Git 工具
│       ├── hash/                          # 哈希工具
│       ├── initializer/                   # 初始化器
│       └── kubernetes/                    # Kubernetes 工具
│
├── core/                                   # 核心业务逻辑层
│   ├── account/                           # 账户管理
│   │   └── account/
│   ├── agent/                             # AI Agent 核心
│   │   ├── agent/                         # Agent 入口和主循环
│   │   ├── agent_stop_hook/               # Agent 停止钩子
│   │   ├── command/                       # Agent 命令处理
│   │   ├── compact/                       # 上下文压缩
│   │   ├── generate/                      # 推理生成
│   │   ├── hooks/                         # Hook 系统
│   │   │   ├── agent_stop/               # Agent 停止时 Hook
│   │   │   │   ├── empty_reply_checker/  # 空回复检查
│   │   │   │   ├── external/             # 外部停止 Hook
│   │   │   │   └── task_completion_checker/ # 任务完成检查
│   │   │   ├── post_tool_use/            # 工具使用后 Hook
│   │   │   │   ├── external/             # 外部 Hook
│   │   │   │   ├── result_persister/     # 结果持久化
│   │   │   │   └── result_truncator/     # 结果截断
│   │   │   ├── pre_tool_use/             # 工具使用前 Hook
│   │   │   │   ├── code_reviewer_checker/ # 代码审查检查
│   │   │   │   └── external/             # 外部 Hook
│   │   │   ├── session_start/            # 会话启动 Hook
│   │   │   │   └── external/
│   │   │   └── user_prompt_submit/       # 用户提交 Hook
│   │   │       └── external/
│   │   ├── mcp/                          # Agent 内 MCP 集成
│   │   ├── option/                       # Agent 配置选项
│   │   │   └── options/
│   │   ├── options/                      # 备选项
│   │   ├── permission/                   # 权限系统
│   │   │   ├── bash_permission_gitignore/ # Bash gitignore 权限
│   │   │   ├── bash_rule_matcher/        # Bash 规则匹配
│   │   │   ├── file_rule_matcher/        # 文件规则匹配
│   │   │   ├── mcp_rule_matcher/         # MCP 规则匹配
│   │   │   ├── path_checker/             # 路径检查
│   │   │   ├── permission_checker/       # 权限检查器
│   │   │   ├── permission_coordinator/   # 权限协调器
│   │   │   ├── permission_mode/          # 权限模式
│   │   │   └── webfetch_rule_matcher/    # Web 获取规则匹配
│   │   ├── prompt/                       # Prompt 管理
│   │   │   ├── coder/                    # Coder 模式 Prompt
│   │   │   ├── quest/                    # Quest 模式 Prompt
│   │   │   ├── summarizer/              # 摘要 Prompt
│   │   │   ├── task/                     # Task 模式 Prompt
│   │   │   └── utils/                    # Prompt 工具
│   │   ├── provider/                     # LLM Provider 层
│   │   │   ├── idealab/                  # IdeaLab Provider
│   │   │   ├── openai/                   # OpenAI Provider
│   │   │   ├── options/                  # Provider 选项
│   │   │   ├── provider/                 # Provider 接口
│   │   │   ├── qoder/                    # Qoder 自有 Provider
│   │   │   └── think/                    # 思考/推理层
│   │   ├── settings/                     # Agent 设置
│   │   ├── state/                        # Agent 状态管理
│   │   │   ├── active/                   # 活跃状态
│   │   │   ├── context/                  # 上下文状态
│   │   │   ├── file/                     # 文件状态
│   │   │   │   ├── file_state/           # 文件状态管理
│   │   │   │   ├── snapshot/             # 文件快照
│   │   │   │   └── storage/              # 文件存储
│   │   │   ├── max/                      # 最大限制
│   │   │   ├── memory/                   # Memory 管理
│   │   │   │   ├── memory/               # Memory 入口
│   │   │   │   ├── memory_state/         # Memory 状态
│   │   │   │   └── qoder_rule/           # Qoder 规则
│   │   │   ├── message/                  # 消息管理
│   │   │   │   ├── cache/                # 消息缓存
│   │   │   │   ├── message/              # 消息核心
│   │   │   │   └── storage/              # 消息存储
│   │   │   ├── session/                  # 会话状态
│   │   │   │   ├── cache/                # 会话缓存
│   │   │   │   ├── options/              # 会话选项
│   │   │   │   ├── session/              # 会话核心
│   │   │   │   └── storage/              # 会话存储
│   │   │   ├── shell/                    # Shell 管理
│   │   │   │   ├── adapt_unix/           # Unix 适配
│   │   │   │   ├── background_shell/     # 后台 Shell
│   │   │   │   ├── grep/                 # Grep 工具
│   │   │   │   │   ├── grep_darwin_arm64/ # arm64 平台 grep
│   │   │   │   │   └── grep_resource/   # grep 资源
│   │   │   │   ├── manager/              # Shell 管理器
│   │   │   │   ├── quote/                # 引号处理
│   │   │   │   ├── shell/                # Shell 核心
│   │   │   │   └── snapshot/             # Shell 快照
│   │   │   ├── state/                    # 状态核心
│   │   │   ├── state_execution/          # 执行状态
│   │   │   ├── state_quest/              # Quest 状态
│   │   │   ├── state_shell/              # Shell 状态
│   │   │   ├── state_todo/               # Todo 状态
│   │   │   └── state_vars/               # 状态变量
│   │   ├── task/                         # SubAgent Task 调度
│   │   ├── title/                        # 标题生成
│   │   └── tools/                        # 内置工具
│   │       ├── abstract/                 # 工具抽象基类
│   │       ├── ask_user_question/        # AskUserQuestion 工具
│   │       ├── bash/                     # Bash 工具
│   │       ├── bashkill/                 # BashKill (KillBash) 工具
│   │       ├── bashoutput/               # BashOutput 工具
│   │       ├── check_runtime/            # CheckRuntime 工具
│   │       ├── edit/                     # Edit 工具
│   │       ├── enter_plan_in_quest/      # Quest 进入 Plan 模式
│   │       ├── exit_plan_in_quest/       # Quest 退出 Plan 模式
│   │       ├── glob/                     # Glob 工具
│   │       ├── grep/                     # Grep 工具
│   │       ├── imagegen/                 # ImageGen 工具
│   │       ├── ls/                       # LS 工具
│   │       ├── multiedit/                # MultiEdit 工具
│   │       ├── read/                     # Read 工具
│   │       ├── skill/                    # Skill 工具
│   │       ├── task/                     # Task (SubAgent) 工具
│   │       ├── todowrite/                # TodoWrite 工具
│   │       ├── utils/                    # 工具通用工具
│   │       │   ├── file/                 # 文件工具
│   │       │   ├── format/               # 格式化工具
│   │       │   └── repair/               # 修复工具
│   │       ├── webfetch/                 # WebFetch 工具
│   │       ├── webfetch_trusted/         # WebFetch (受信任) 工具
│   │       ├── websearch/                # WebSearch 工具
│   │       └── write/                    # Write 工具
│   ├── auth/                             # 认证模块
│   │   ├── default/                      # 默认认证
│   │   └── model/                        # 模型认证
│   │       ├── anthropic/                # Anthropic 认证
│   │       ├── dashscope/                # DashScope 认证
│   │       ├── filter/                   # 模型过滤
│   │       ├── idealab/                  # IdeaLab 认证
│   │       ├── models/                   # 模型列表
│   │       ├── openai/                   # OpenAI 认证
│   │       └── qoder/                    # Qoder 认证
│   ├── config/                           # 配置管理
│   │   ├── config/                       # 配置核心
│   │   ├── qoder/                        # Qoder 配置
│   │   └── settings/                     # 设置管理
│   ├── generator/                        # 生成器
│   │   ├── command/                      # 命令生成器
│   │   └── subagent/                     # SubAgent 生成器
│   ├── logging/                          # 日志系统
│   │   ├── broker/                       # 日志中间件
│   │   ├── config/                       # 日志配置
│   │   └── logger/                       # 日志核心
│   ├── monitoring/                       # 监控
│   │   └── run/                          # 运行监控
│   ├── pubsub/                           # 发布/订阅
│   │   └── broker/                       # 消息代理
│   ├── resource/                         # 资源管理
│   │   ├── command/                      # 斜杠命令
│   │   │   ├── builtin/                  # 内置命令
│   │   │   └── commands/                 # 命令列表
│   │   ├── hook/                         # Hook 系统
│   │   │   ├── executor/                 # Hook 执行器
│   │   │   ├── hook/                     # Hook 核心
│   │   │   ├── hooks/                    # Hook 列表
│   │   │   └── log/                      # Hook 日志
│   │   ├── mcp/                          # MCP 资源
│   │   │   ├── mcp/                      # MCP 核心
│   │   │   ├── mcp_ipc_client_unix/      # Unix IPC 客户端
│   │   │   ├── mcps/                     # MCP 列表
│   │   │   ├── process/                  # MCP 进程管理
│   │   │   └── token_store/              # MCP Token 存储
│   │   ├── output_style/                 # 输出样式
│   │   │   ├── built_in/                 # 内置样式
│   │   │   └── output_style/             # 样式核心
│   │   ├── plugin/                       # 插件系统
│   │   │   ├── bootstrap/                # 插件引导
│   │   │   ├── config/                   # 插件配置
│   │   │   ├── errors/                   # 插件错误
│   │   │   ├── integration/              # 插件集成
│   │   │   ├── loader/                   # 插件加载
│   │   │   ├── manifest/                 # 插件清单
│   │   │   ├── registry/                 # 插件注册
│   │   │   └── resolver/                 # 插件解析
│   │   ├── skill/                        # Skill 系统
│   │   │   ├── builtin/                  # 内置 Skill
│   │   │   └── skills/                   # Skill 列表
│   │   └── subagent/                     # SubAgent 系统
│   │       ├── builtin/                  # 内置 SubAgent
│   │       └── subagent/                 # SubAgent 核心
│   ├── runtime/                          # 运行时检测
│   ├── types/                            # 类型定义
│   │   ├── compact/                      # 压缩类型
│   │   ├── error/                        # 错误类型
│   │   ├── location/                     # 位置类型
│   │   ├── message/                      # 消息类型
│   │   ├── permission/                   # 权限类型
│   │   ├── sdk_marshal/                  # SDK 序列化
│   │   └── sdk_protocol/                 # SDK 协议
│   └── utils/                            # 核心工具库
│       ├── converter/                    # 类型转换器
│       │   ├── anthropic/                # Anthropic 格式转换
│       │   └── qoder/                    # Qoder 格式转换
│       ├── diff/                         # Diff 工具
│       │   └── diff/
│       ├── env/                          # 环境变量
│       │   └── env/
│       ├── envprep/                      # 环境准备
│       │   ├── download/                 # 下载
│       │   ├── envprep/                  # 环境准备核心
│       │   ├── extract/                  # 解压
│       │   ├── gitbash_preparer/         # Git Bash 准备
│       │   └── lock/                     # 文件锁
│       ├── feedback/                     # 反馈收集
│       │   ├── collector/                # 收集器
│       │   ├── feedback/                 # 反馈核心
│       │   ├── image_upload/             # 图片上传
│       │   ├── sign/                     # 签名
│       │   └── truncate/                 # 截断
│       ├── fileutil/                     # 文件工具
│       │   └── fileutil/
│       ├── goroutine/                    # Goroutine 管理
│       │   └── run/
│       ├── http/                         # HTTP 客户端
│       │   ├── client/                   # 客户端核心
│       │   └── init/                     # 初始化
│       ├── imageutil/                    # 图片处理
│       │   └── processor/
│       ├── install/                      # 安装工具
│       │   └── source/
│       ├── jsonrepair/                   # JSON 修复
│       │   ├── const/
│       │   ├── jsonrepair/
│       │   └── utils/
│       ├── pathutil/                     # 路径工具
│       │   └── path/
│       ├── qoder/                        # Qoder 服务交互
│       │   ├── auth_callback/            # 认证回调
│       │   ├── business_tracker/         # 业务追踪
│       │   ├── codebase_api/             # 代码库 API
│       │   ├── codebase_service/         # 代码库服务
│       │   ├── codebase_types/           # 代码库类型
│       │   ├── device_info/              # 设备信息
│       │   ├── device_polling/           # 设备轮询
│       │   ├── device_token/             # 设备 Token
│       │   ├── endpoints/                # API 端点
│       │   ├── github/                   # GitHub 集成
│       │   ├── json/                     # JSON 工具
│       │   ├── manager/                  # 管理器
│       │   ├── oauth/                    # OAuth 认证
│       │   ├── privacy_policy/           # 隐私政策
│       │   ├── qoder_support/            # Qoder 支持
│       │   ├── quota/                    # 配额管理
│       │   ├── region_config/            # 区域配置
│       │   ├── request/                  # 请求构建
│       │   ├── sse/                      # SSE 客户端
│       │   │   ├── client/
│       │   │   └── event/
│       │   └── storage/                  # 本地存储
│       ├── runner/                       # 运行器
│       │   └── runner/
│       ├── shellutil/                    # Shell 工具
│       │   └── detect/
│       ├── sls/                          # SLS (日志服务)
│       │   ├── command_tracker/          # 命令追踪
│       │   ├── data_format/              # 数据格式
│       │   ├── heartbeat/                # 心跳
│       │   ├── initializer/              # 初始化
│       │   └── reporter/                 # 上报
│       ├── storage/                      # 通用存储
│       │   └── storage/
│       ├── system/                       # 系统工具
│       │   ├── os/
│       │   └── util/
│       ├── template/                     # 模板工具
│       │   └── template/
│       ├── timeutil/                     # 时间工具
│       │   └── timeutil/
│       ├── tokens/                       # Token 计算
│       │   └── tokens/
│       ├── umid/                         # 设备唯一 ID
│       │   ├── types/
│       │   ├── umid/
│       │   ├── umid_common/
│       │   └── umid_unix/
│       ├── version/                      # 版本管理
│       │   └── version/
│       └── yamlrepair/                   # YAML 修复
│           ├── const/
│           ├── utils/
│           └── yamlrepair/
│
├── acp/                                    # ACP (Agent Communication Protocol) 层
│   ├── acp/                              # ACP 核心
│   ├── acp_file/                         # ACP 文件服务
│   ├── json_utils/                       # JSON 工具
│   ├── mcp_server/                       # MCP 服务器
│   ├── option/                           # ACP 选项
│   ├── qoder_agent/                      # Qoder Agent
│   ├── session/                          # ACP 会话
│   └── tools/                            # ACP 工具
│
├── sdk/                                    # SDK 模式 (非交互式)
│   ├── control/                          # 控制协议
│   ├── process_query/                    # 进程查询
│   └── runtime/                          # SDK 运行时
│
├── tui/                                    # TUI (终端用户界面) 层
│   ├── app/                              # 应用入口
│   ├── cmd/                              # TUI 命令处理
│   ├── components/                       # UI 组件
│   │   ├── askuser/                      # 用户问题对话框
│   │   │   └── askuser/
│   │   ├── command/                      # 斜杠命令 UI
│   │   │   ├── agents/                   # /agents 命令
│   │   │   ├── bashes/                   # /bashes 命令
│   │   │   ├── clear/                    # /clear 命令
│   │   │   ├── commands/                 # 命令列表
│   │   │   ├── compact/                  # /compact 命令
│   │   │   ├── config/                   # /config 命令
│   │   │   ├── export/                   # /export 命令
│   │   │   ├── feedback/                 # /feedback 命令
│   │   │   ├── github/                   # /github 命令
│   │   │   │   ├── assets/               # GitHub Actions 模板
│   │   │   │   └── installer/            # 安装器
│   │   │   ├── help/                     # /help 命令
│   │   │   ├── login/                    # /login 命令
│   │   │   ├── logout/                   # /logout 命令
│   │   │   ├── mcp/                      # /mcp 命令
│   │   │   ├── memory/                   # /memory 命令
│   │   │   ├── model/                    # /model 命令
│   │   │   ├── prompt/                   # /prompt 命令
│   │   │   ├── quest_off/                # /quest-off 命令
│   │   │   ├── quest_on/                 # /quest-on 命令
│   │   │   ├── quit/                     # /quit 命令
│   │   │   ├── release_notes/            # /release-notes 命令
│   │   │   ├── resume/                   # /resume 命令
│   │   │   ├── setup_github/             # /setup-github 命令
│   │   │   ├── skills/                   # /skills 命令
│   │   │   ├── status/                   # /status 命令
│   │   │   ├── upgrade/                  # /upgrade 命令
│   │   │   ├── usage/                    # /usage 命令
│   │   │   └── vim/                      # /vim 命令
│   │   ├── common/                       # 通用 UI 组件
│   │   │   ├── dialog/                   # 对话框
│   │   │   │   ├── command/              # 命令对话框
│   │   │   │   ├── simple/               # 简单对话框
│   │   │   │   └── support/              # 支持对话框
│   │   │   ├── editor/                   # 编辑器组件
│   │   │   │   └── editor/
│   │   │   └── textarea/                 # 文本区域组件
│   │   │       └── textarea/
│   │   ├── filepicker/                   # 文件选择器
│   │   │   ├── filepicker/
│   │   │   └── filetree/
│   │   ├── interaction/                  # 交互面板
│   │   │   ├── bash/                     # Bash 交互
│   │   │   ├── board/                    # 面板
│   │   │   ├── editor/                   # 输入编辑器
│   │   │   │   ├── attachment/           # 附件处理
│   │   │   │   ├── bash_handler/         # Bash 处理
│   │   │   │   ├── cache/                # 编辑器缓存
│   │   │   │   ├── editor/               # 编辑器核心
│   │   │   │   ├── hint/                 # 提示
│   │   │   │   ├── history/              # 历史记录
│   │   │   │   ├── pending/              # 待处理
│   │   │   │   └── windows_paste_handler/ # Windows 粘贴
│   │   │   ├── memory/                   # 内存面板
│   │   │   ├── progress/                 # 进度条
│   │   │   │   └── progress/
│   │   │   ├── selectors/                # 选择器
│   │   │   │   ├── at_selector/          # @ 选择器
│   │   │   │   ├── command_selector/     # 命令选择器
│   │   │   │   └── memory_selector/      # 内存选择器
│   │   │   └── status/                   # 状态栏
│   │   │       └── status/
│   │   ├── messages/                     # 消息显示
│   │   │   ├── bash/                     # Bash 输出
│   │   │   ├── command/                  # 命令消息
│   │   │   ├── log/                      # 日志消息
│   │   │   ├── memory/                   # 内存消息
│   │   │   ├── message/                  # 消息核心
│   │   │   ├── process/                  # 进程消息
│   │   │   ├── render/                   # 渲染器
│   │   │   ├── tool/                     # 工具消息
│   │   │   ├── view/                     # 视图
│   │   │   └── welcome/                  # 欢迎消息
│   │   └── permission/                   # 权限 UI
│   │       ├── permission/
│   │       └── quest_switch/
│   ├── event/                            # TUI 事件
│   ├── render/                           # 渲染
│   ├── state/                            # TUI 状态
│   │   └── global_state/
│   ├── styles/                           # 样式
│   │   └── styles/
│   ├── theme/                            # 主题
│   │   ├── catppuccin/
│   │   ├── dracula/
│   │   ├── flexoki/
│   │   ├── gruvbox/
│   │   ├── init/
│   │   ├── manager/
│   │   ├── monokai/
│   │   ├── onedark/
│   │   ├── qoder/                        # Qoder 默认主题
│   │   ├── theme/
│   │   ├── tokyonight/
│   │   └── tron/
│   └── util/                             # TUI 工具
│       ├── code/
│       ├── diff/
│       ├── markdown/
│       ├── path/
│       ├── text/
│       └── util/
│
└── profile/                                # 性能分析
    ├── init/
    ├── profile/
    └── trigger/
```

## 3. 第三方依赖列表

### 3.1 核心框架与 SDK

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/spf13/cobra` | - | CLI 框架 |
| `github.com/anthropics/anthropic-sdk-go` | - | Anthropic Claude SDK |
| `github.com/google/go-github/v57` | v57.0.0 | GitHub API 客户端 |
| `github.com/Masterminds/semver/v3` | v3 | 语义版本控制 |

### 3.2 TUI 与终端

| 依赖 | 用途 |
|------|------|
| `github.com/charmbracelet/bubbletea` | TUI 框架 (推断, Bubble Tea) |
| `github.com/charmbracelet/lipgloss` | 终端样式 (推断) |
| `github.com/alecthomas/chroma/v2` | 代码语法高亮 |
| `github.com/atotto/clipboard` | 剪贴板操作 |
| `github.com/aymanbagabas/go-osc52/v2` | OSC52 终端协议 |
| `github.com/aymanbagabas/go-udiff` | Unified Diff |

### 3.3 数据处理

| 依赖 | 用途 |
|------|------|
| `github.com/BurntSushi/toml` | v1.5.0 | TOML 解析 |
| `github.com/JohannesKaufmann/html-to-markdown/v2` | HTML 转 Markdown |
| `github.com/JohannesKaufmann/dom` | v0.2.0 | DOM 操作 |
| `github.com/adrg/frontmatter` | Frontmatter 解析 |
| `github.com/andybalholm/brotli` | Brotli 压缩/解压 |

### 3.4 基础设施

| 依赖 | 用途 |
|------|------|
| `github.com/aliyun/alicloud-httpdns-go-sdk` | 阿里云 HTTPDNS |
| `golang.org/x/crypto` | 加密库 |
| `golang.org/x/net` | 网络库 |
| `golang.org/x/sync` | 并发原语 |
| `golang.org/x/sys` | 系统调用 |
| `golang.org/x/term` | 终端操作 |
| `golang.org/x/text` | 文本处理 |
| `golang.org/x/exp` | 实验性库 |

## 4. 核心 types 包类型定义

从二进制字符串中提取的关键类型（`core/types` 包）：

```
types.Role              types.Type              types.MIME
types.Error             types.Delta             types.Finish
types.Memory            types.Status            types.ModelId
types.Message           types.ToolCall          types.Location
types.ToolInfo          types.DeltaType         types.ErrorData
types.ToolResult        types.TokenUsage        types.Attachment
types.SdkMessage        types.ResultData        types.SystemData
types.ContentPart       types.TextContent       types.TypedFinish
types.ImageSource       types.FeatureName       types.FinishReason
types.ContentBlock      types.BusinessType      types.BusinessInfo
types.BinaryContent     types.ModelProvider      types.ReasoningItem
types.InterruptData     types.BusinessStage     types.SystemContent
types.SdkMessageType    types.PermissionRule    types.ControlRequest
types.CanUseToolData    types.InitializeData    types.McpMessageData
types.CompactTrigger    types.ImageUrlContent   types.ReminderContent
types.TypedToolResult   types.RewindFilesData   types.ControlResponse
types.UserMessageData   types.BusinessProduct   types.CompactMetadata
types.StreamEventData   types.ReasoningContent  types.HookCallbackData
types.TrackingEventData types.ControlRequestData
```

## 5. qoder 服务端交互类型

从二进制中提取的 `qoder` 包类型（`core/utils/qoder`）：

```
qoder.AuthManager           qoder.AuthService          qoder.AuthStatus
qoder.CachedUserInfo        qoder.ChatResponse         qoder.CodebaseService
qoder.CodebaseStatus        qoder.CosyUserInfo         qoder.DataPolicyResponse
qoder.DeviceTokenResponse   qoder.ErrorResponse        qoder.FinishReason
qoder.FunctionCall          qoder.FunctionDefinition   qoder.HttpPayload
qoder.LoginMethod           qoder.LoginRequestResult   qoder.Message
qoder.ModelConfig           qoder.Organization         qoder.QuotaUsage
qoder.RegionConfigCache     qoder.RemoteRegionConfig   qoder.RequestBuilder
qoder.StreamedChatResponsePayload  qoder.Tool           qoder.ToolCall
qoder.UserFeatureResponse   qoder.UserInfoResponse     qoder.UserPlan
qoder.WhitelistStatus       qoder.DeviceTokenPollingManager
qoder.ExchangeGitHubAppTokenParam  qoder.GitHubAppRepoStatusResponse
```

## 6. Provider 层类型

```
provider.Client             # Provider 客户端接口
provider.ClientBuilder      # 客户端构建器
provider.Event              # 流式事件
provider.EventType          # 事件类型
provider.ModelRequest       # 模型请求
provider.Response           # 模型响应
provider.TokenUsage         # Token 使用量
provider.ThinkLevel         # 思考级别
provider.ReasoningEffort    # 推理努力程度

# 具体实现
provider.OpenAIClient       # OpenAI 客户端
provider.QoderClient        # Qoder 客户端
provider.IdeaLabClient      # IdeaLab 客户端
```

## 7. ACP (Agent Communication Protocol) 层类型

```
acp.QoderAcpAgent           acp.Session            acp.SessionConfig
acp.AgentOption             acp.FileDiff           acp.MessageBuffer
acp.MessageStreamBuffer     acp.SearchResultType
```

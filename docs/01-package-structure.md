# qodercli-reverse Package Structure and Dependency Analysis

## 1. Internal Go Module Path

```
code.alibaba-inc.com/qoder-core/qodercli-reverse
```

> This is the internal Go module path of qodercli-reverse, indicating the project is hosted on Alibaba's internal GitLab.

## 2. Full Package Structure Tree

The following is the complete internal package list extracted from the binary symbol table (approximately 250+ subpackages):

```
code.alibaba-inc.com/qoder-core/qodercli-reverse/
│
├── cmd/                                    # CLI command layer (entry point)
│   ├── root/                              # Root command definition
│   ├── start/                             # Start command
│   │   ├── cleanup/                       # Resource cleanup
│   │   ├── job_startup/                   # Job startup
│   │   ├── output_format/                 # Output formatting
│   │   ├── start_container/               # Container mode startup
│   │   ├── start_kubernetes/              # Kubernetes mode startup
│   │   ├── start_local/                   # Local mode startup
│   │   ├── start_worktree/                # Git worktree mode startup
│   │   └── subscriptions/                 # Event subscriptions
│   ├── update/                            # Self-update command
│   │   ├── auto/                          # Automatic update
│   │   ├── common/                        # Update common logic
│   │   ├── curlbash_updater/              # curl|bash update method
│   │   ├── homebrew_updater/              # Homebrew update method
│   │   ├── npm_updater/                   # NPM update method
│   │   └── update/                        # Update entry point
│   ├── mcp/                               # MCP server management command
│   │   ├── add/                           # Add MCP server
│   │   ├── auth/                          # MCP OAuth authentication
│   │   ├── get/                           # Get MCP server info
│   │   ├── list/                          # List MCP servers
│   │   ├── mcp/                           # MCP command entry point
│   │   └── remove/                        # Remove MCP server
│   ├── jobs/                              # Remote Job management command
│   │   ├── attach/                        # Attach to Job
│   │   ├── fetch/                         # Fetch Job
│   │   ├── jobs/                          # Jobs command entry point
│   │   ├── rm/                            # Remove Job
│   │   └── stop/                          # Stop Job
│   ├── feedback/                          # Feedback command
│   │   └── feedback/
│   ├── machineinfo/                       # Machine info command
│   │   └── machine_info/
│   ├── status/                            # Status command
│   │   └── status/
│   ├── message_io/                        # Message I/O (SDK mode)
│   │   ├── out/                           # Output handling
│   │   ├── streaming/                     # Streaming handling
│   │   └── transform/                     # Message transformation
│   └── utils/                             # Command utilities
│       ├── archive/                       # Archive/compression
│       ├── docker/                        # Docker utilities
│       ├── flags/                         # Command-line flags
│       ├── git/                           # Git utilities
│       ├── hash/                          # Hash utilities
│       ├── initializer/                   # Initializer
│       └── kubernetes/                    # Kubernetes utilities
│
├── core/                                   # Core business logic layer
│   ├── account/                           # Account management
│   │   └── account/
│   ├── agent/                             # AI Agent core
│   │   ├── agent/                         # Agent entry point and main loop
│   │   ├── agent_stop_hook/               # Agent stop hook
│   │   ├── command/                       # Agent command handling
│   │   ├── compact/                       # Context compaction
│   │   ├── generate/                      # Inference generation
│   │   ├── hooks/                         # Hook system
│   │   │   ├── agent_stop/               # Agent stop hooks
│   │   │   │   ├── empty_reply_checker/  # Empty reply checker
│   │   │   │   ├── external/             # External stop hook
│   │   │   │   └── task_completion_checker/ # Task completion checker
│   │   │   ├── post_tool_use/            # Post-tool-use hook
│   │   │   │   ├── external/             # External hook
│   │   │   │   ├── result_persister/     # Result persister
│   │   │   │   └── result_truncator/     # Result truncator
│   │   │   ├── pre_tool_use/             # Pre-tool-use hook
│   │   │   │   ├── code_reviewer_checker/ # Code review checker
│   │   │   │   └── external/             # External hook
│   │   │   ├── session_start/            # Session start hook
│   │   │   │   └── external/
│   │   │   └── user_prompt_submit/       # User prompt submit hook
│   │   │       └── external/
│   │   ├── mcp/                          # Agent-internal MCP integration
│   │   ├── option/                       # Agent configuration options
│   │   │   └── options/
│   │   ├── options/                      # Alternative options
│   │   ├── permission/                   # Permission system
│   │   │   ├── bash_permission_gitignore/ # Bash gitignore permission
│   │   │   ├── bash_rule_matcher/        # Bash rule matcher
│   │   │   ├── file_rule_matcher/        # File rule matcher
│   │   │   ├── mcp_rule_matcher/         # MCP rule matcher
│   │   │   ├── path_checker/             # Path checker
│   │   │   ├── permission_checker/       # Permission checker
│   │   │   ├── permission_coordinator/   # Permission coordinator
│   │   │   ├── permission_mode/          # Permission mode
│   │   │   └── webfetch_rule_matcher/    # Web fetch rule matcher
│   │   ├── prompt/                       # Prompt management
│   │   │   ├── coder/                    # Coder mode prompt
│   │   │   ├── quest/                    # Quest mode prompt
│   │   │   ├── summarizer/              # Summarizer prompt
│   │   │   ├── task/                     # Task mode prompt
│   │   │   └── utils/                    # Prompt utilities
│   │   ├── provider/                     # LLM Provider layer
│   │   │   ├── idealab/                  # IdeaLab Provider
│   │   │   ├── openai/                   # OpenAI Provider
│   │   │   ├── options/                  # Provider options
│   │   │   ├── provider/                 # Provider interface
│   │   │   ├── qoder/                    # Qoder proprietary Provider
│   │   │   └── think/                    # Thinking/reasoning layer
│   │   ├── settings/                     # Agent settings
│   │   ├── state/                        # Agent state management
│   │   │   ├── active/                   # Active state
│   │   │   ├── context/                  # Context state
│   │   │   ├── file/                     # File state
│   │   │   │   ├── file_state/           # File state management
│   │   │   │   ├── snapshot/             # File snapshot
│   │   │   │   └── storage/              # File storage
│   │   │   ├── max/                      # Maximum limits
│   │   │   ├── memory/                   # Memory management
│   │   │   │   ├── memory/               # Memory entry point
│   │   │   │   ├── memory_state/         # Memory state
│   │   │   │   └── qoder_rule/           # Qoder rules
│   │   │   ├── message/                  # Message management
│   │   │   │   ├── cache/                # Message cache
│   │   │   │   ├── message/              # Message core
│   │   │   │   └── storage/              # Message storage
│   │   │   ├── session/                  # Session state
│   │   │   │   ├── cache/                # Session cache
│   │   │   │   ├── options/              # Session options
│   │   │   │   ├── session/              # Session core
│   │   │   │   └── storage/              # Session storage
│   │   │   ├── shell/                    # Shell management
│   │   │   │   ├── adapt_unix/           # Unix adapter
│   │   │   │   ├── background_shell/     # Background shell
│   │   │   │   ├── grep/                 # Grep tool
│   │   │   │   │   ├── grep_darwin_arm64/ # arm64 platform grep
│   │   │   │   │   └── grep_resource/   # grep resources
│   │   │   │   ├── manager/              # Shell manager
│   │   │   │   ├── quote/                # Quote handling
│   │   │   │   ├── shell/                # Shell core
│   │   │   │   └── snapshot/             # Shell snapshot
│   │   │   ├── state/                    # State core
│   │   │   ├── state_execution/          # Execution state
│   │   │   ├── state_quest/              # Quest state
│   │   │   ├── state_shell/              # Shell state
│   │   │   ├── state_todo/               # Todo state
│   │   │   └── state_vars/               # State variables
│   │   ├── task/                         # SubAgent Task scheduling
│   │   ├── title/                        # Title generation
│   │   └── tools/                        # Built-in tools
│   │       ├── abstract/                 # Tool abstract base class
│   │       ├── ask_user_question/        # AskUserQuestion tool
│   │       ├── bash/                     # Bash tool
│   │       ├── bashkill/                 # BashKill (KillBash) tool
│   │       ├── bashoutput/               # BashOutput tool
│   │       ├── check_runtime/            # CheckRuntime tool
│   │       ├── edit/                     # Edit tool
│   │       ├── enter_plan_in_quest/      # Quest enter Plan mode
│   │       ├── exit_plan_in_quest/       # Quest exit Plan mode
│   │       ├── glob/                     # Glob tool
│   │       ├── grep/                     # Grep tool
│   │       ├── imagegen/                 # ImageGen tool
│   │       ├── ls/                       # LS tool
│   │       ├── multiedit/                # MultiEdit tool
│   │       ├── read/                     # Read tool
│   │       ├── skill/                    # Skill tool
│   │       ├── task/                     # Task (SubAgent) tool
│   │       ├── todowrite/                # TodoWrite tool
│   │       ├── utils/                    # Tool common utilities
│   │       │   ├── file/                 # File utilities
│   │       │   ├── format/               # Formatting utilities
│   │       │   └── repair/               # Repair utilities
│   │       ├── webfetch/                 # WebFetch tool
│   │       ├── webfetch_trusted/         # WebFetch (trusted) tool
│   │       ├── websearch/                # WebSearch tool
│   │       └── write/                    # Write tool
│   ├── auth/                             # Authentication module
│   │   ├── default/                      # Default authentication
│   │   └── model/                        # Model authentication
│   │       ├── anthropic/                # Anthropic authentication
│   │       ├── dashscope/                # DashScope authentication
│   │       ├── filter/                   # Model filter
│   │       ├── idealab/                  # IdeaLab authentication
│   │       ├── models/                   # Model list
│   │       ├── openai/                   # OpenAI authentication
│   │       └── qoder/                    # Qoder authentication
│   ├── config/                           # Configuration management
│   │   ├── config/                       # Configuration core
│   │   ├── qoder/                        # Qoder configuration
│   │   └── settings/                     # Settings management
│   ├── generator/                        # Generators
│   │   ├── command/                      # Command generator
│   │   └── subagent/                     # SubAgent generator
│   ├── logging/                          # Logging system
│   │   ├── broker/                       # Logging middleware
│   │   ├── config/                       # Logging configuration
│   │   └── logger/                       # Logging core
│   ├── monitoring/                       # Monitoring
│   │   └── run/                          # Run monitoring
│   ├── pubsub/                           # Publish/subscribe
│   │   └── broker/                       # Message broker
│   ├── resource/                         # Resource management
│   │   ├── command/                      # Slash commands
│   │   │   ├── builtin/                  # Built-in commands
│   │   │   └── commands/                 # Command list
│   │   ├── hook/                         # Hook system
│   │   │   ├── executor/                 # Hook executor
│   │   │   ├── hook/                     # Hook core
│   │   │   ├── hooks/                    # Hook list
│   │   │   └── log/                      # Hook log
│   │   ├── mcp/                          # MCP resources
│   │   │   ├── mcp/                      # MCP core
│   │   │   ├── mcp_ipc_client_unix/      # Unix IPC client
│   │   │   ├── mcps/                     # MCP list
│   │   │   ├── process/                  # MCP process management
│   │   │   └── token_store/              # MCP Token storage
│   │   ├── output_style/                 # Output styles
│   │   │   ├── built_in/                 # Built-in styles
│   │   │   └── output_style/             # Style core
│   │   ├── plugin/                       # Plugin system
│   │   │   ├── bootstrap/                # Plugin bootstrap
│   │   │   ├── config/                   # Plugin configuration
│   │   │   ├── errors/                   # Plugin errors
│   │   │   ├── integration/              # Plugin integration
│   │   │   ├── loader/                   # Plugin loader
│   │   │   ├── manifest/                 # Plugin manifest
│   │   │   ├── registry/                 # Plugin registry
│   │   │   └── resolver/                 # Plugin resolver
│   │   ├── skill/                        # Skill system
│   │   │   ├── builtin/                  # Built-in skills
│   │   │   └── skills/                   # Skill list
│   │   └── subagent/                     # SubAgent system
│   │       ├── builtin/                  # Built-in SubAgents
│   │       └── subagent/                 # SubAgent core
│   ├── runtime/                          # Runtime detection
│   ├── types/                            # Type definitions
│   │   ├── compact/                      # Compaction types
│   │   ├── error/                        # Error types
│   │   ├── location/                     # Location types
│   │   ├── message/                      # Message types
│   │   ├── permission/                   # Permission types
│   │   ├── sdk_marshal/                  # SDK serialization
│   │   └── sdk_protocol/                 # SDK protocol
│   └── utils/                            # Core utility library
│       ├── converter/                    # Type converters
│       │   ├── anthropic/                # Anthropic format conversion
│       │   └── qoder/                    # Qoder format conversion
│       ├── diff/                         # Diff utilities
│       │   └── diff/
│       ├── env/                          # Environment variables
│       │   └── env/
│       ├── envprep/                      # Environment preparation
│       │   ├── download/                 # Download
│       │   ├── envprep/                  # Environment preparation core
│       │   ├── extract/                  # Extraction
│       │   ├── gitbash_preparer/         # Git Bash preparation
│       │   └── lock/                     # File lock
│       ├── feedback/                     # Feedback collection
│       │   ├── collector/                # Collector
│       │   ├── feedback/                 # Feedback core
│       │   ├── image_upload/             # Image upload
│       │   ├── sign/                     # Signature
│       │   └── truncate/                 # Truncation
│       ├── fileutil/                     # File utilities
│       │   └── fileutil/
│       ├── goroutine/                    # Goroutine management
│       │   └── run/
│       ├── http/                         # HTTP client
│       │   ├── client/                   # Client core
│       │   └── init/                     # Initialization
│       ├── imageutil/                    # Image processing
│       │   └── processor/
│       ├── install/                      # Installation utilities
│       │   └── source/
│       ├── jsonrepair/                   # JSON repair
│       │   ├── const/
│       │   ├── jsonrepair/
│       │   └── utils/
│       ├── pathutil/                     # Path utilities
│       │   └── path/
│       ├── qoder/                        # Qoder service interaction
│       │   ├── auth_callback/            # Authentication callback
│       │   ├── business_tracker/         # Business tracking
│       │   ├── codebase_api/             # Codebase API
│       │   ├── codebase_service/         # Codebase service
│       │   ├── codebase_types/           # Codebase types
│       │   ├── device_info/              # Device info
│       │   ├── device_polling/           # Device polling
│       │   ├── device_token/             # Device token
│       │   ├── endpoints/                # API endpoints
│       │   ├── github/                   # GitHub integration
│       │   ├── json/                     # JSON utilities
│       │   ├── manager/                  # Manager
│       │   ├── oauth/                    # OAuth authentication
│       │   ├── privacy_policy/           # Privacy policy
│       │   ├── qoder_support/            # Qoder support
│       │   ├── quota/                    # Quota management
│       │   ├── region_config/            # Region configuration
│       │   ├── request/                  # Request building
│       │   ├── sse/                      # SSE client
│       │   │   ├── client/
│       │   │   └── event/
│       │   └── storage/                  # Local storage
│       ├── runner/                       # Runner
│       │   └── runner/
│       ├── shellutil/                    # Shell utilities
│       │   └── detect/
│       ├── sls/                          # SLS (logging service)
│       │   ├── command_tracker/          # Command tracking
│       │   ├── data_format/              # Data format
│       │   ├── heartbeat/                # Heartbeat
│       │   ├── initializer/              # Initializer
│       │   └── reporter/                 # Reporting
│       ├── storage/                      # Generic storage
│       │   └── storage/
│       ├── system/                       # System utilities
│       │   ├── os/
│       │   └── util/
│       ├── template/                     # Template utilities
│       │   └── template/
│       ├── timeutil/                     # Time utilities
│       │   └── timeutil/
│       ├── tokens/                       # Token calculation
│       │   └── tokens/
│       ├── umid/                         # Device unique ID
│       │   ├── types/
│       │   ├── umid/
│       │   ├── umid_common/
│       │   └── umid_unix/
│       ├── version/                      # Version management
│       │   └── version/
│       └── yamlrepair/                   # YAML repair
│           ├── const/
│           ├── utils/
│           └── yamlrepair/
│
├── acp/                                    # ACP (Agent Communication Protocol) layer
│   ├── acp/                              # ACP core
│   ├── acp_file/                         # ACP file service
│   ├── json_utils/                       # JSON utilities
│   ├── mcp_server/                       # MCP server
│   ├── option/                           # ACP options
│   ├── qoder_agent/                      # Qoder Agent
│   ├── session/                          # ACP session
│   └── tools/                            # ACP tools
│
├── sdk/                                    # SDK mode (non-interactive)
│   ├── control/                          # Control protocol
│   ├── process_query/                    # Process query
│   └── runtime/                          # SDK runtime
│
├── tui/                                    # TUI (terminal user interface) layer
│   ├── app/                              # Application entry point
│   ├── cmd/                              # TUI command handling
│   ├── components/                       # UI components
│   │   ├── askuser/                      # User question dialog
│   │   │   └── askuser/
│   │   ├── command/                      # Slash command UI
│   │   │   ├── agents/                   # /agents command
│   │   │   ├── bashes/                   # /bashes command
│   │   │   ├── clear/                    # /clear command
│   │   │   ├── commands/                 # Command list
│   │   │   ├── compact/                  # /compact command
│   │   │   ├── config/                   # /config command
│   │   │   ├── export/                   # /export command
│   │   │   ├── feedback/                 # /feedback command
│   │   │   ├── github/                   # /github command
│   │   │   │   ├── assets/               # GitHub Actions templates
│   │   │   │   └── installer/            # Installer
│   │   │   ├── help/                     # /help command
│   │   │   ├── login/                    # /login command
│   │   │   ├── logout/                   # /logout command
│   │   │   ├── mcp/                      # /mcp command
│   │   │   ├── memory/                   # /memory command
│   │   │   ├── model/                    # /model command
│   │   │   ├── prompt/                   # /prompt command
│   │   │   ├── quest_off/                # /quest-off command
│   │   │   ├── quest_on/                 # /quest-on command
│   │   │   ├── quit/                     # /quit command
│   │   │   ├── release_notes/            # /release-notes command
│   │   │   ├── resume/                   # /resume command
│   │   │   ├── setup_github/             # /setup-github command
│   │   │   ├── skills/                   # /skills command
│   │   │   ├── status/                   # /status command
│   │   │   ├── upgrade/                  # /upgrade command
│   │   │   ├── usage/                    # /usage command
│   │   │   └── vim/                      # /vim command
│   │   ├── common/                       # Common UI components
│   │   │   ├── dialog/                   # Dialog
│   │   │   │   ├── command/              # Command dialog
│   │   │   │   ├── simple/               # Simple dialog
│   │   │   │   └── support/              # Support dialog
│   │   │   ├── editor/                   # Editor component
│   │   │   │   └── editor/
│   │   │   └── textarea/                 # Textarea component
│   │   │       └── textarea/
│   │   ├── filepicker/                   # File picker
│   │   │   ├── filepicker/
│   │   │   └── filetree/
│   │   ├── interaction/                  # Interaction panel
│   │   │   ├── bash/                     # Bash interaction
│   │   │   ├── board/                    # Board
│   │   │   ├── editor/                   # Input editor
│   │   │   │   ├── attachment/           # Attachment handling
│   │   │   │   ├── bash_handler/         # Bash handling
│   │   │   │   ├── cache/                # Editor cache
│   │   │   │   ├── editor/               # Editor core
│   │   │   │   ├── hint/                 # Hint
│   │   │   │   ├── history/              # History
│   │   │   │   ├── pending/              # Pending
│   │   │   │   └── windows_paste_handler/ # Windows paste
│   │   │   ├── memory/                   # Memory panel
│   │   │   ├── progress/                 # Progress bar
│   │   │   │   └── progress/
│   │   │   ├── selectors/                # Selectors
│   │   │   │   ├── at_selector/          # @ selector
│   │   │   │   ├── command_selector/     # Command selector
│   │   │   │   └── memory_selector/      # Memory selector
│   │   │   └── status/                   # Status bar
│   │   │       └── status/
│   │   ├── messages/                     # Message display
│   │   │   ├── bash/                     # Bash output
│   │   │   ├── command/                  # Command message
│   │   │   ├── log/                      # Log message
│   │   │   ├── memory/                   # Memory message
│   │   │   ├── message/                  # Message core
│   │   │   ├── process/                  # Process message
│   │   │   ├── render/                   # Renderer
│   │   │   ├── tool/                     # Tool message
│   │   │   ├── view/                     # View
│   │   │   └── welcome/                  # Welcome message
│   │   └── permission/                   # Permission UI
│   │       ├── permission/
│   │       └── quest_switch/
│   ├── event/                            # TUI events
│   ├── render/                           # Rendering
│   ├── state/                            # TUI state
│   │   └── global_state/
│   ├── styles/                           # Styles
│   │   └── styles/
│   ├── theme/                            # Theme
│   │   ├── catppuccin/
│   │   ├── dracula/
│   │   ├── flexoki/
│   │   ├── gruvbox/
│   │   ├── init/
│   │   ├── manager/
│   │   ├── monokai/
│   │   ├── onedark/
│   │   ├── qoder/                        # Qoder default theme
│   │   ├── theme/
│   │   ├── tokyonight/
│   │   └── tron/
│   └── util/                             # TUI utilities
│       ├── code/
│       ├── diff/
│       ├── markdown/
│       ├── path/
│       ├── text/
│       └── util/
│
└── profile/                                # Performance profiling
    ├── init/
    ├── profile/
    └── trigger/
```

## 3. Third-Party Dependencies

### 3.1 Core Frameworks and SDKs

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/spf13/cobra` | - | CLI framework |
| `github.com/anthropics/anthropic-sdk-go` | - | Anthropic Claude SDK |
| `github.com/google/go-github/v57` | v57.0.0 | GitHub API client |
| `github.com/Masterminds/semver/v3` | v3 | Semantic versioning |

### 3.2 TUI and Terminal

| Dependency | Purpose |
|------------|---------|
| `github.com/charmbracelet/bubbletea` | TUI framework (inferred, Bubble Tea) |
| `github.com/charmbracelet/lipgloss` | Terminal styling (inferred) |
| `github.com/alecthomas/chroma/v2` | Code syntax highlighting |
| `github.com/atotto/clipboard` | Clipboard operations |
| `github.com/aymanbagabas/go-osc52/v2` | OSC52 terminal protocol |
| `github.com/aymanbagabas/go-udiff` | Unified Diff |

### 3.3 Data Processing

| Dependency | Purpose |
|------------|---------|
| `github.com/BurntSushi/toml` | v1.5.0 | TOML parsing |
| `github.com/JohannesKaufmann/html-to-markdown/v2` | HTML to Markdown |
| `github.com/JohannesKaufmann/dom` | v0.2.0 | DOM manipulation |
| `github.com/adrg/frontmatter` | Frontmatter parsing |
| `github.com/andybalholm/brotli` | Brotli compression/decompression |

### 3.4 Infrastructure

| Dependency | Purpose |
|------------|---------|
| `github.com/aliyun/alicloud-httpdns-go-sdk` | Alibaba Cloud HTTPDNS |
| `golang.org/x/crypto` | Cryptography library |
| `golang.org/x/net` | Networking library |
| `golang.org/x/sync` | Concurrency primitives |
| `golang.org/x/sys` | System calls |
| `golang.org/x/term` | Terminal operations |
| `golang.org/x/text` | Text processing |
| `golang.org/x/exp` | Experimental library |

## 4. Core Types in the `types` Package

Key types extracted from binary strings (in the `core/types` package):

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

## 5. Qoder Service Interaction Types

`qoder` package types (in `core/utils/qoder`):

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

## 6. Provider Layer Types

```
provider.Client             # Provider client interface
provider.ClientBuilder      # Client builder
provider.Event              # Streaming event
provider.EventType          # Event type
provider.ModelRequest       # Model request
provider.Response           # Model response
provider.TokenUsage         # Token usage
provider.ThinkLevel         # Thinking level
provider.ReasoningEffort    # Reasoning effort

# Concrete implementations
provider.OpenAIClient       # OpenAI client
provider.QoderClient        # Qoder client
provider.IdeaLabClient      # IdeaLab client
```

## 7. ACP (Agent Communication Protocol) Layer Types

```
acp.QoderAcpAgent           acp.Session            acp.SessionConfig
acp.AgentOption             acp.FileDiff           acp.MessageBuffer
acp.MessageStreamBuffer     acp.SearchResultType
```
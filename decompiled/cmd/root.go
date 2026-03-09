package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/config"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/pubsub"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
	"github.com/alingse/qodercli-reverse/decompiled/tui/app"
)

var (
	// 全局标志变量
	model              string
	maxTokens          int
	temperature        float64
	maxTurns           int
	permissionMode     string
	systemPrompt       string
	printMode          string
	continueSession    bool
	resumeSession      string
	workspace          string
	outputFormat       string
	inputFormat        string
	maxOutputTokens    string
	allowedTools       []string
	disallowedTools    []string
	attachments        []string
	agents             string
	skipPermissions    bool
	yolo               bool
	worktree           bool
	branch             string
	path               string
	withClaudeConfig   bool
	quiet              bool
	version            bool
	debug              bool
	logFile            string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qodercli",
	Short: "Qoder CLI - AI-powered development assistant",
	Long: `Qoder CLI is an AI-powered development assistant that helps you with coding tasks,
	file operations, and system interactions through natural language commands.`,
	
	// 不带参数时运行 TUI 模式
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			fmt.Println("qodercli version 0.1.29 (decompiled)")
			return
		}

		// 如果有 print 参数，运行非交互模式
		if printMode != "" {
			runPrintMode(printMode)
			return
		}

		// 运行 TUI 模式
		runTUI()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&model, "model", "", "auto", "Model to use (auto/efficient/gmodel/kmodel/lite/mmodel/performance/q35model/qmodel/ultimate)")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "", 4096, "Maximum tokens for response")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "", 0.7, "Temperature for generation")
	rootCmd.PersistentFlags().IntVarP(&maxTurns, "max-turns", "", 25, "Maximum turns for agent")
	rootCmd.PersistentFlags().StringVarP(&permissionMode, "permission-mode", "", "ask", "Permission mode (ask/allow/deny)")
	rootCmd.PersistentFlags().StringVarP(&systemPrompt, "system-prompt", "", "You are a helpful AI assistant.", "System prompt for the agent")
	
	// 日志标志
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "", "", "Log file path (default: ~/.qoder/qodercli.log)")
	
	// 主要标志
	rootCmd.Flags().StringVarP(&printMode, "print", "p", "", "Non-interactive mode: process single prompt and exit")
	rootCmd.Flags().BoolVarP(&continueSession, "continue", "c", false, "Continue last conversation")
	rootCmd.Flags().StringVarP(&resumeSession, "resume", "r", "", "Resume specific session")
	rootCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace directory")
	rootCmd.Flags().StringVarP(&outputFormat, "output-format", "f", "text", "Output format (text/json/stream-json)")
	rootCmd.Flags().StringVarP(&inputFormat, "input-format", "", "text", "Input format (text/stream-json)")
	rootCmd.Flags().StringVarP(&maxOutputTokens, "max-output-tokens", "", "", "Maximum output tokens (16k/32k)")
	rootCmd.Flags().StringSliceVarP(&allowedTools, "allowed-tools", "", []string{}, "Allowed tools list")
	rootCmd.Flags().StringSliceVarP(&disallowedTools, "disallowed-tools", "", []string{}, "Disallowed tools list")
	rootCmd.Flags().StringSliceVarP(&attachments, "attachment", "", []string{}, "Attachment paths (can be specified multiple times)")
	rootCmd.Flags().StringVarP(&agents, "agents", "", "", "Custom agent JSON definition")
	rootCmd.Flags().BoolVarP(&skipPermissions, "dangerously-skip-permissions", "", false, "Skip permission checks")
	rootCmd.Flags().BoolVarP(&yolo, "yolo", "", false, "Alias for dangerously-skip-permissions")
	rootCmd.Flags().BoolVarP(&worktree, "worktree", "", false, "Start concurrent task via git worktree")
	rootCmd.Flags().StringVarP(&branch, "branch", "", "", "Worktree branch name")
	rootCmd.Flags().StringVarP(&path, "path", "", "", "Worktree path")
	rootCmd.Flags().BoolVarP(&withClaudeConfig, "with-claude-config", "", false, "Load .claude configuration")
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Hide spinner in non-interactive mode")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "Show version information")
	
	// 隐藏标志（用于兼容性）
	rootCmd.Flags().BoolP("help", "h", false, "Help for qodercli")
}

// runPrintMode 运行非交互打印模式
func runPrintMode(input string) {
	// 初始化日志
	logLevel := log.LevelInfo
	if debug {
		logLevel = log.LevelDebug
	}
	if logFile == "" {
		logFile = getDefaultLogFile()
	}
	if err := log.Init(logFile, logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init log: %v\n", err)
	}
	defer log.Close()

	log.Info("Starting qodercli in print mode")
	log.Debug("Input: %s", input)

	cfg := loadConfig()
	prov, err := createProviderFromEnv()
	if err != nil {
		log.Error("Failed to create provider: %v", err)
		log.Fatalf("Failed to create provider: %v", err)
	}

	agentCfg := &agent.Config{
		Model:          cfg.Model,
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		MaxTurns:       maxTurns,
		PermissionMode: permission.Mode(permissionMode),
		SystemPrompt:   systemPrompt,
	}

	ag, err := agent.NewAgent(agentCfg, prov)
	if err != nil {
		log.Error("Failed to create agent: %v", err)
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 简单的输出处理 - 流式增量打印
	var lastPrintedLen int
	var fullText strings.Builder
	var currentToolCall *types.ToolCall // 记录当前工具调用

	ag.SetCallbacks(
		func(msg *types.Message) {
			if !quiet {
				// 构建完整的文本内容
				fullText.Reset()
				for _, part := range msg.Content {
					if part.Type == "text" && part.Text != "" {
						fullText.WriteString(part.Text)
					}
				}

				// 只打印新增的文本部分
				fullStr := fullText.String()
				if len(fullStr) > lastPrintedLen {
					fmt.Print(fullStr[lastPrintedLen:])
					lastPrintedLen = len(fullStr)
				}
			}
		},
		func(call *types.ToolCall) {
			currentToolCall = call // 保存当前工具调用
			log.Debug("Tool call: %s", call.Name)
			if !quiet && debug {
				// 打印工具调用信息到标准输出，使用优雅的格式
				fmt.Fprintf(os.Stderr, "\n● %s\n", call.Name)

				// 对于 Read 工具，只显示文件路径
				if call.Name == "Read" {
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(call.Arguments), &args); err == nil {
						if filePath, ok := args["file_path"].(string); ok {
							fmt.Fprintf(os.Stderr, "  ⎿ %s\n", filePath)
						}
					}
				} else if call.Arguments != "" {
					fmt.Fprintf(os.Stderr, "  ⎿ %s\n", call.Arguments)
				}
			}
		},
		func(result *types.ToolResult) {
			log.Debug("Tool result: %s", result.ToolCallID)
			if !quiet && debug {
				// 打印工具结果信息到标准输出，使用优雅的格式
				if result.IsError {
					fmt.Fprintf(os.Stderr, "  ⎿ ✗ Error: %s\n", result.Content)
				} else {
					// 对于 Read 工具，只显示行数统计
					if currentToolCall != nil && currentToolCall.Name == "Read" {
						// 统计行数
						lineCount := strings.Count(result.Content, "\n")
						if lineCount > 0 {
							fmt.Fprintf(os.Stderr, "  ⎿ Read %d lines\n", lineCount)
						}
					} else {
						// 其他工具显示简短摘要
						content := result.Content
						maxLen := 100
						if len(content) > maxLen {
							// 尝试在合适的位置截断
							truncated := content[:maxLen]
							if idx := strings.LastIndex(truncated, "\n"); idx > 50 {
								truncated = truncated[:idx]
							}
							fmt.Fprintf(os.Stderr, "  ⎿ %s... (truncated)\n", truncated)
						} else if content != "" {
							// 如果内容较短，显示第一行
							firstLine := content
							if idx := strings.Index(content, "\n"); idx > 0 {
								firstLine = content[:idx]
							}
							if len(firstLine) > maxLen {
								firstLine = firstLine[:maxLen] + "..."
							}
							fmt.Fprintf(os.Stderr, "  ⎿ %s\n", firstLine)
						}
					}
				}
			}
		},
		func(err error) {
			log.Error("Callback error: %v", err)
		},
		func(reason types.FinishReason) {
			if !quiet {
				fmt.Println()
			}
			log.Info("Request completed with finish reason: %s", reason)
		},
	)

	ctx := context.Background()
	if err := ag.ProcessUserInput(ctx, input); err != nil {
		log.Error("Error processing input: %v", err)
		log.Fatalf("Error processing input: %v", err)
	}
}

// runTUI 运行 TUI 模式
func runTUI() {
	// 初始化日志
	logLevel := log.LevelInfo
	if debug {
		logLevel = log.LevelDebug
	}
	if logFile == "" {
		logFile = getDefaultLogFile()
	}
	if err := log.Init(logFile, logLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init log: %v\n", err)
	}

	log.Info("Starting qodercli in TUI mode")
	log.Debug("Log file: %s", logFile)

	cfg := loadConfig()
	prov, err := createProviderFromEnv()
	if err != nil {
		log.Error("Failed to create provider: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to create provider: %v\n", err)
		os.Exit(1)
	}

	ps := pubsub.New()
	_ = ps // 保留 ps 用于未来扩展

	agentCfg := &agent.Config{
		Model:          cfg.Model,
		MaxTokens:      maxTokens,
		Temperature:    temperature,
		MaxTurns:       maxTurns,
		PermissionMode: permission.Mode(permissionMode),
		SystemPrompt:   systemPrompt,
	}

	ag, err := agent.NewAgent(agentCfg, prov)
	if err != nil {
		log.Error("Failed to create agent: %v", err)
		fmt.Fprintf(os.Stderr, "Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// 启动 TUI
	opts := app.Options{
		Config: cfg,
		Agent:  ag,
		PubSub: ps,
	}

	if err := app.Run(opts); err != nil {
		log.Error("TUI error: %v", err)
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}

	log.Close()
}

// loadConfig 从标志和环境变量加载配置
func loadConfig() *config.Config {
	cfg := config.LoadFromEnv()
	
	// CLI 标志覆盖环境变量
	if model != "" && model != "auto" {
		cfg.Model = model
	}
	if maxTokens > 0 {
		cfg.MaxTokens = maxTokens
	}
	if temperature > 0 {
		cfg.Temperature = temperature
	}
	if maxTurns > 0 {
		cfg.MaxTurns = maxTurns
	}
	if permissionMode != "" {
		cfg.PermissionMode = permissionMode
	}
	if outputFormat != "" {
		cfg.OutputFormat = outputFormat
	}
	if len(allowedTools) > 0 {
		cfg.AllowedTools = allowedTools
	}
	if len(disallowedTools) > 0 {
		cfg.DisallowedTools = disallowedTools
	}
	if workspace != "" {
		cfg.Workspace = workspace
	}
	
	return cfg
}

// createProviderFromEnv 根据环境变量创建 Provider
// 优先级：标准环境变量 > QODER_* 环境变量
func createProviderFromEnv() (provider.Client, error) {
	// 首先检查标准 OpenAI 环境变量（优先级最高）
	standardOpenAIKey := os.Getenv("OPENAI_API_KEY")
	standardBaseURL := os.Getenv("OPENAI_BASE_URL")
	standardModel := os.Getenv("OPENAI_MODEL")

	// 检查 QODER_* 格式的环境变量
	qoderToken := os.Getenv("QODER_PERSONAL_ACCESS_TOKEN")
	anthropicKey := os.Getenv("QODER_ANTHROPIC_API_KEY")
	qoderOpenAIKey := os.Getenv("QODER_OPENAI_API_KEY")
	qoderBaseURL := os.Getenv("QODER_OPENAI_BASE_URL")
	qoderModel := os.Getenv("QODER_OPENAI_MODEL")
	dashscopeKey := os.Getenv("QODER_DASHSCOPE_API_KEY")
	idealabKey := os.Getenv("QODER_IDEALAB_API_KEY")

	log.Debug("Checking environment variables for API keys")
	if standardOpenAIKey != "" {
		log.Debug("Found OPENAI_API_KEY, using standard OpenAI provider")
	}
	if qoderOpenAIKey != "" {
		log.Debug("Found QODER_OPENAI_API_KEY, using Qoder OpenAI provider")
	}

	// 优先使用标准 OpenAI 环境变量（支持任意 OpenAI 兼容服务）
	if standardOpenAIKey != "" {
		var opts []provider.ClientOption
		opts = append(opts, provider.WithAPIKey(standardOpenAIKey))
		
		if standardBaseURL != "" {
			opts = append(opts, provider.WithBaseURL(standardBaseURL))
			log.Debug("Using custom base URL: %s", standardBaseURL)
		}
		
		client, err := provider.NewOpenAIClient(opts...)
		if err != nil {
			return nil, fmt.Errorf("create OpenAI client: %w", err)
		}
		
		// 如果指定了模型，更新配置
		if standardModel != "" {
			model = standardModel
			log.Debug("Using model from OPENAI_MODEL: %s", standardModel)
		}
		
		return client, nil
	}

	// 其次检查 QODER_OPENAI_* 格式
	if qoderOpenAIKey != "" {
		var opts []provider.ClientOption
		opts = append(opts, provider.WithAPIKey(qoderOpenAIKey))
		
		if qoderBaseURL != "" {
			opts = append(opts, provider.WithBaseURL(qoderBaseURL))
			log.Debug("Using Qoder base URL: %s", qoderBaseURL)
		}
		
		client, err := provider.NewOpenAIClient(opts...)
		if err != nil {
			return nil, fmt.Errorf("create OpenAI client: %w", err)
		}
		
		if qoderModel != "" {
			model = qoderModel
			log.Debug("Using model from QODER_OPENAI_MODEL: %s", qoderModel)
		}
		
		return client, nil
	}

	// Qoder 官方服务
	if qoderToken != "" {
		log.Debug("Using Qoder personal access token")
		return provider.NewQoderClient(provider.WithAPIKey(qoderToken))
	}

	// 其他 Provider（暂未实现）
	if anthropicKey != "" {
		return nil, fmt.Errorf("Anthropic provider: use OPENAI_API_KEY with compatible base URL instead")
	}
	if dashscopeKey != "" {
		return nil, fmt.Errorf("DashScope provider: use OPENAI_API_KEY with https://dashscope.aliyuncs.com/compatible-mode/v1 as base URL")
	}
	if idealabKey != "" {
		return nil, fmt.Errorf("IdeaLab provider not implemented yet")
	}

	return nil, fmt.Errorf("no API key found. Set OPENAI_API_KEY or QODER_PERSONAL_ACCESS_TOKEN")
}

// getDefaultLogFile 获取默认日志文件路径
func getDefaultLogFile() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "qodercli.log"
	}
	
	logDir := filepath.Join(homeDir, ".qoder")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "qodercli.log"
	}
	
	return filepath.Join(logDir, "qodercli.log")
}
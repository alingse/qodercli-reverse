package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/print"
	"github.com/alingse/qodercli-reverse/decompiled/cmd/tui"
	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
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

		// 初始化日志
		if err := utils.InitLogger(logFile, debug); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init log: %v\n", err)
		}
		defer log.Close()

		// 构建 flags 结构
		flags := &utils.Flags{
			Model:           model,
			MaxTokens:       maxTokens,
			Temperature:     temperature,
			MaxTurns:        maxTurns,
			PermissionMode:  permissionMode,
			OutputFormat:    outputFormat,
			AllowedTools:    allowedTools,
			DisallowedTools: disallowedTools,
			Workspace:       workspace,
		}

		// 如果有 print 参数，运行非交互模式
		if printMode != "" {
			if err := print.Run(printMode, flags, quiet, systemPrompt); err != nil {
				log.Fatalf("Print mode error: %v", err)
			}
			return
		}

		// 运行 TUI 模式
		if err := tui.Run(flags, systemPrompt); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
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

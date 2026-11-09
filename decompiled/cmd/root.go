package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/print"
	"github.com/alingse/qodercli-reverse/decompiled/cmd/tui"
	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/version"
)

var (
	// Global flag variables - strictly aligned with official qodercli-reverse
	model            string
	maxTokens        int
	temperature      float64
	maxTurns         int
	permissionMode   string
	systemPrompt     string
	printMode        string
	continueSession  bool
	resumeSession    string
	workspace        string
	planMode         bool // Enable plan mode: agent outputs a plan before executing tools
goalMode         bool // Enable goal mode: agent works toward a verifiable goal
goal             string // Goal to achieve (natural language description)
	outputFormat     string
	inputFormat      string
	maxOutputTokens  string
	allowedTools     []string
	disallowedTools  []string
	attachments      []string
	agents           string
	skipPermissions  bool
	yolo             bool
	worktree         bool
	branch           string
	path             string
	withClaudeConfig bool
	quiet            bool
	showVersion      bool
	debug            bool
	logFile          string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qodercli-reverse",
	Short: "Qoder CLI - AI-powered development assistant",
	Long: `Qoder CLI is an AI-powered development assistant that helps you with coding tasks,
file operations, and system interactions through natural language commands.`,

	// Run TUI mode when no arguments provided
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("qodercli-reverse version %s (decompiled)\n", version.Version)
			return
		}

		// Initialize logging
		if err := utils.InitLogger(logFile, debug); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init log: %v\n", err)
		}
		defer log.Close()

		// Determine working directory
		workDir := workspace
		if workDir == "" {
			workDir, _ = os.Getwd()
		}

		// Build system prompt
		// Official behavior: if user specified --system-prompt, use it; otherwise auto-build internally
		systemPromptText := systemPrompt
		if systemPromptText == "" {
			// Auto-build system prompt internally
			var buildErr error
			systemPromptText, buildErr = utils.BuildSystemPromptAuto(workDir, withClaudeConfig)
			if buildErr != nil {
				log.Error("Failed to build system prompt: %v", buildErr)
				// Fall back to default prompt
				systemPromptText = utils.GetDefaultSystemPrompt()
			}
			log.Debug("Auto-built system prompt, length: %d", len(systemPromptText))
		} else {
			log.Debug("Using user-provided system prompt")
		}

		// Build flags struct
		flags := &utils.Flags{
			Model:            model,
			MaxTokens:        maxTokens,
			Temperature:      temperature,
			MaxTurns:         maxTurns,
			PermissionMode:   permissionMode,
			OutputFormat:     outputFormat,
			AllowedTools:     allowedTools,
			DisallowedTools:  disallowedTools,
			Workspace:        workspace,
			WithClaudeConfig: withClaudeConfig,
			Debug:            debug,
		}

		// If print arg present, run non-interactive mode
		if printMode != "" {
			if err := print.Run(printMode, flags, quiet, systemPromptText, planMode, goalMode, goal); err != nil {
				log.Fatalf("Print mode error: %v", err)
			}
			return
		}

		// Run TUI mode
		if err := tui.Run(flags, systemPromptText, planMode, goalMode, goal); err != nil {
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
	// Global flags - strictly aligned with official
	rootCmd.PersistentFlags().StringVarP(&model, "model", "", "auto", "Model to use (auto/efficient/gmodel/kmodel/lite/mmodel/performance/q35model/qmodel/ultimate)")
	rootCmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "", 4096, "Maximum tokens for response")
	rootCmd.PersistentFlags().Float64VarP(&temperature, "temperature", "", 0.7, "Temperature for generation")
	rootCmd.PersistentFlags().IntVarP(&maxTurns, "max-turns", "", 25, "Maximum turns for agent")
	rootCmd.PersistentFlags().BoolVarP(&planMode, "plan-mode", "", false, "Enable plan mode: agent outputs a plan before executing tools")
	rootCmd.PersistentFlags().StringVarP(&permissionMode, "permission-mode", "", "ask", "Permission mode (ask/allow/deny)")
rootCmd.PersistentFlags().BoolVarP(&goalMode, "goal-mode", "", false, "Enable goal mode: agent works toward a verifiable goal")
rootCmd.PersistentFlags().StringVarP(&goal, "goal", "", "", "Goal to achieve (natural language description)")

	// System prompt flags - official standard
	rootCmd.PersistentFlags().StringVarP(&systemPrompt, "system-prompt", "", "", "System prompt for the agent")

	// Logging flags
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "", "", "Log file path (default: ~/.qodercli-reverse/qodercli-reverse.log)")

	// Main flags - strictly aligned with official
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
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")

	// Hidden flags (for compatibility)
	rootCmd.Flags().BoolP("help", "h", false, "Help for qodercli-reverse")
}

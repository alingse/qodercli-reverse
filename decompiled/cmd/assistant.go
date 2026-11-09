package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/github"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

var (
	assistantRepo    string
	assistantIssue   int
	assistantPR      int
	assistantComment string
	assistantAuthor  string
	assistantPost    bool
)

var assistantCmd = &cobra.Command{
	Use:   "assistant",
	Short: "Interactive GitHub assistant for issue/PR comments",
	Long: `Handle @qoder mentions on GitHub issues or pull requests. Fetches the
relevant context, runs it through the AI agent, and optionally posts a response.
Never executes arbitrary commands requested by untrusted commenters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.InitLogger("", false)
		defer log.Close()

		if assistantRepo == "" {
			return fmt.Errorf("--repo is required")
		}
		parts := strings.SplitN(assistantRepo, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --repo %q", assistantRepo)
		}
		owner, repo := parts[0], parts[1]

		number := assistantIssue
		if number == 0 {
			number = assistantPR
		}
		if number == 0 {
			return fmt.Errorf("--issue or --pr is required")
		}

		// Bot filtering
		if strings.HasSuffix(assistantAuthor, "[bot]") {
			fmt.Println("Bot comment, skipping.")
			return nil
		}

		intent := parseQoderIntent(assistantComment)
		if intent == "" {
			fmt.Println("No recognized @qoder intent found; nothing to do.")
			return nil
		}

		client := github.NewClient()

		var contextBuilder strings.Builder
		fmt.Fprintf(&contextBuilder, "Repository: %s\n", assistantRepo)
		if assistantPR > 0 {
			fmt.Fprintf(&contextBuilder, "Pull request: #%d\n", assistantPR)
			pr, err := client.GetPullRequest(owner, repo, assistantPR)
			if err == nil {
				fmt.Fprintf(&contextBuilder, "Title: %s\n", pr.Title)
				fmt.Fprintf(&contextBuilder, "Body: %s\n", pr.Body)
			}
		} else {
			fmt.Fprintf(&contextBuilder, "Issue: #%d\n", assistantIssue)
		}
		if assistantAuthor != "" {
			fmt.Fprintf(&contextBuilder, "Requester: %s\n", assistantAuthor)
		}
		fmt.Fprintf(&contextBuilder, "Comment: %s\n", assistantComment)
		fmt.Fprintf(&contextBuilder, "Intent: %s\n", intent)

		cfg := utils.LoadConfig(nil)
		prov, modelOverride, perr := utils.CreateProvider()
		if perr != nil {
			log.Error("Failed to create provider: %v", perr)
			fmt.Println(contextBuilder.String())
		} else {
			if modelOverride != "" {
				cfg.Model = modelOverride
			}
			agentCfg := &agent.Config{
				Model:          cfg.Model,
				MaxTokens:      4096,
				Temperature:    0.7,
				MaxTurns:       25,
				PermissionMode: permission.Mode("allow"),
				SystemPrompt:   utils.GetDefaultSystemPrompt(),
				PlanMode:       planMode,
			}
			ag, aerr := agent.NewAgent(agentCfg, prov)
			if aerr != nil {
				log.Error("Failed to create agent: %v", aerr)
				fmt.Println(contextBuilder.String())
			} else {
				var lastCount int
				var full strings.Builder
				ag.SetCallbacks(
					func(msg *types.Message) {
						full.Reset()
						for _, part := range msg.Content {
							if part.Type == "text" && part.Text != "" {
								full.WriteString(part.Text)
							}
						}
						s := full.String()
						runes := []rune(s)
						if len(runes) > lastCount {
							fmt.Print(string(runes[lastCount:]))
							lastCount = len(runes)
						}
					},
					func(call *types.ToolCall) { log.Debug("tool call: %s", call.Name) },
					func(result *types.ToolResult) { log.Debug("tool result: %s", result.ToolCallID) },
					func(e error) {
						if e != nil {
							log.Error("callback error: %v", e)
						}
					},
					func(reason types.FinishReason) { fmt.Println() },
				)
				ctx := context.Background()
				if err := ag.ProcessUserInput(ctx, contextBuilder.String()); err != nil {
					log.Error("Agent error: %v", err)
				}
			}
		}

		if assistantPost && client.Token != "" {
			if err := client.CreateIssueComment(owner, repo, number, "AI assistant response generated. See CLI output."); err != nil {
				log.Warn("Failed to post comment: %v", err)
			} else {
				fmt.Println("Posted assistant response to GitHub.")
			}
		}
		return nil
	},
}

func parseQoderIntent(comment string) string {
	lower := strings.ToLower(comment)
	if !strings.Contains(lower, "@qoder") {
		return ""
	}
	allowed := []string{"explain", "fix", "refactor", "tests", "generate tests"}
	for _, v := range allowed {
		if strings.Contains(lower, v) {
			return v
		}
	}
	return "explain"
}

func init() {
	assistantCmd.Flags().StringVar(&assistantRepo, "repo", "", "Repository in owner/name form")
	assistantCmd.Flags().IntVar(&assistantIssue, "issue", 0, "Issue number")
	assistantCmd.Flags().IntVar(&assistantPR, "pr", 0, "Pull request number")
	assistantCmd.Flags().StringVar(&assistantComment, "comment", "", "Comment text containing @qoder")
	assistantCmd.Flags().StringVar(&assistantAuthor, "author", "", "Comment author")
	assistantCmd.Flags().BoolVar(&assistantPost, "post", false, "Post the response back to GitHub")
	rootCmd.AddCommand(assistantCmd)
}

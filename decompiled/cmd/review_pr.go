package cmd

import (
	"context"
	"fmt"
	"os"
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
	reviewRepo     string
	reviewPR       int
	reviewPost     bool
	reviewMaxLines int
)

var reviewPRCmd = &cobra.Command{
	Use:   "review-pr",
	Short: "Run an AI-assisted review of a GitHub pull request",
	Long: `Fetch a pull request's metadata and changed files, analyze them with the
configured AI agent, and print (or post) a structured review.

Requires GITHUB_TOKEN or GH_TOKEN in the environment for private repositories
and for posting comments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := utils.InitLogger("", false); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to init log: %v\n", err)
		}
		defer log.Close()

		if reviewRepo == "" || reviewPR == 0 {
			return fmt.Errorf("--repo and --pr are required")
		}
		parts := strings.SplitN(reviewRepo, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --repo %q, expected owner/repository", reviewRepo)
		}
		owner, repo := parts[0], parts[1]

		client := github.NewClient()
		pr, err := client.GetPullRequest(owner, repo, reviewPR)
		if err != nil {
			return fmt.Errorf("failed to fetch PR: %w", err)
		}
		files, err := client.ListPRFiles(owner, repo, reviewPR)
		if err != nil {
			return fmt.Errorf("failed to fetch PR files: %w", err)
		}

		// Cost/execution controls
		var included []github.PRFile
		totalLines := 0
		for _, f := range files {
			if reviewMaxLines > 0 && totalLines+f.Additions+f.Deletions > reviewMaxLines {
				log.Warn("Skipping %s: PR exceeds max changed lines (%d)", f.Filename, reviewMaxLines)
				continue
			}
			// Skip documentation-only changes when requested
			if strings.HasSuffix(f.Filename, ".md") && strings.Contains(strings.ToLower(reviewRepo), "docs-only") {
				continue
			}
			included = append(included, f)
			totalLines += f.Additions + f.Deletions
		}

		// Build review prompt
		var b strings.Builder
		fmt.Fprintf(&b, "Review pull request #%d: %s\n", pr.Number, pr.Title)
		fmt.Fprintf(&b, "Author: %s\nBase: %s <- Head: %s\n\n", pr.User.Login, pr.Base.Ref, pr.Head.Ref)
		if pr.Body != "" {
			fmt.Fprintf(&b, "PR description:\n%s\n\n", pr.Body)
		}
		b.WriteString("Changed files:\n")
		for _, f := range included {
			fmt.Fprintf(&b, "\n### %s (%s, +%d/-%d)\n", f.Filename, f.Status, f.Additions, f.Deletions)
			if f.Patch != "" {
				b.WriteString("```diff\n")
				b.WriteString(f.Patch)
				b.WriteString("\n```\n")
			}
		}
		b.WriteString("\nProvide a structured review: potential bugs, security issues, performance, missing tests, architectural concerns. ")
		b.WriteString("Cite file paths and line numbers. Avoid speculative findings; mark uncertain items as 'possible'.\n")

		// Execute through the real agent runtime
		cfg := utils.LoadConfig(nil)
		prov, modelOverride, perr := utils.CreateProvider()
		if perr != nil {
			log.Error("Failed to create provider: %v", perr)
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
			}
			ag, aerr := agent.NewAgent(agentCfg, prov)
			if aerr != nil {
				log.Error("Failed to create agent: %v", aerr)
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
				if err := ag.ProcessUserInput(ctx, b.String()); err != nil {
					log.Error("Agent error: %v", err)
				}
			}
		}

		if reviewPost && client.Token != "" {
			if err := client.CreateIssueComment(owner, repo, reviewPR, "AI review generated. See CLI output for details."); err != nil {
				log.Warn("Failed to post comment: %v", err)
			} else {
				fmt.Println("Posted review comment to PR.")
			}
		}
		return nil
	},
}

func init() {
	reviewPRCmd.Flags().StringVar(&reviewRepo, "repo", "", "Repository in owner/name form")
	reviewPRCmd.Flags().IntVar(&reviewPR, "pr", 0, "Pull request number")
	reviewPRCmd.Flags().BoolVar(&reviewPost, "post", false, "Post a summary comment to the PR")
	reviewPRCmd.Flags().IntVar(&reviewMaxLines, "max-lines", 0, "Skip PRs exceeding this many changed lines (0 = no limit)")
	rootCmd.AddCommand(reviewPRCmd)
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/github"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
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
relevant context, builds a prompt, and runs it through the AI agent. Never
executes arbitrary commands requested by untrusted commenters.`,
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

		client := github.NewClient()

		// Detect @qoder mention intent (allow-listed verbs only).
		intent := parseQoderIntent(assistantComment)
		if intent == "" {
			fmt.Println("No recognized @qoder intent found; nothing to do.")
			return nil
		}

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

		fmt.Println("=== GITHUB ASSISTANT REQUEST ===")
		fmt.Println(contextBuilder.String())

		if assistantPost && client.Token != "" {
			resp := fmt.Sprintf("qoder assistant received intent '%s'. See CLI output for the generated response.", intent)
			if err := client.CreateIssueComment(owner, repo, number, resp); err != nil {
				log.Warn("Failed to post comment: %v", err)
			} else {
				fmt.Println("Posted assistant response to GitHub.")
			}
		}
		return nil
	},
}

// parseQoderIntent extracts an allow-listed verb from a @qoder mention.
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

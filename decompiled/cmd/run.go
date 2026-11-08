package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/agent"
	"github.com/alingse/qodercli-reverse/decompiled/core/agent/permission"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the agent in GitHub Actions compatible mode",
	Long: `Reads INPUT_PROMPT, INPUT_FLAGS, GITHUB_WORKSPACE and GITHUB_OUTPUT from the
environment, executes the agent with stream-json output, writes a unique output
file path to GITHUB_OUTPUT, and propagates the process exit code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		utils.InitLogger("", false)
		defer log.Close()

		workspace := os.Getenv("GITHUB_WORKSPACE")
		if workspace == "" {
			workspace, _ = os.Getwd()
		}
		prompt := os.Getenv("INPUT_PROMPT")
		flagsInput := os.Getenv("INPUT_FLAGS")

		timestamp := time.Now().Unix()
		outputFile := filepath.Join(os.TempDir(), fmt.Sprintf("qoder-output-%d.log", timestamp))
		errorFile := filepath.Join(os.TempDir(), fmt.Sprintf("qoder-error-%d.log", timestamp))

		outF, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		errF, err := os.Create(errorFile)
		if err != nil {
			outF.Close()
			return err
		}

		emit := func(ev map[string]any) {
			b, _ := json.Marshal(ev)
			line := string(b) + "\n"
			outF.WriteString(line)
			fmt.Print(line)
		}

		emit(map[string]any{"type": "start", "workspace": workspace})

		// Build agent flags
		flags := &utils.Flags{
			Workspace:    workspace,
			OutputFormat: "stream-json",
		}
		// Parse additional flags (one per line, simple split)
		for _, raw := range strings.Split(flagsInput, "\n") {
			line := strings.TrimSpace(raw)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			flags.AllowedTools = append(flags.AllowedTools, fields...)
		}
		_ = flags

		cfg := utils.LoadConfig(flags)
		prov, modelOverride, err := utils.CreateProvider()
		if err != nil {
			emit(map[string]any{"type": "error", "error": err.Error()})
			outF.Close()
			errF.Close()
			return err
		}
		if modelOverride != "" {
			cfg.Model = modelOverride
		}

		systemPrompt, _ := utils.BuildSystemPromptAuto(workspace, false)
		if systemPrompt == "" {
			systemPrompt = utils.GetDefaultSystemPrompt()
		}

		agentCfg := &agent.Config{
			Model:          cfg.Model,
			MaxTokens:      flags.MaxTokens,
			Temperature:    flags.Temperature,
			MaxTurns:       flags.MaxTurns,
			PermissionMode: permission.Mode(flags.PermissionMode),
			SystemPrompt:   systemPrompt,
		}
		ag, err := agent.NewAgent(agentCfg, prov)
		if err != nil {
			emit(map[string]any{"type": "error", "error": err.Error()})
			outF.Close()
			errF.Close()
			return err
		}

		ag.SetCallbacks(
			func(msg *types.Message) {
				for _, part := range msg.Content {
					if part.Type == "text" && part.Text != "" {
						emit(map[string]any{"type": "text", "content": part.Text})
					}
				}
			},
			func(call *types.ToolCall) {
				emit(map[string]any{"type": "tool_call", "tool_name": call.Name, "tool_id": call.ID})
			},
			func(result *types.ToolResult) {
				emit(map[string]any{"type": "tool_result", "tool_id": result.ToolCallID, "is_error": result.IsError})
			},
			func(e error) {
				if e != nil {
					emit(map[string]any{"type": "error", "error": e.Error()})
					fmt.Fprintln(errF, e.Error())
				}
			},
			func(reason types.FinishReason) {
				emit(map[string]any{"type": "finish", "reason": string(reason)})
			},
		)

		ctx := context.Background()
		if err := ag.ProcessUserInput(ctx, prompt); err != nil {
			emit(map[string]any{"type": "error", "error": err.Error()})
			fmt.Fprintln(errF, err.Error())
			outF.Close()
			errF.Close()
			writeGitHubOutput(outputFile, errorFile)
			return err
		}

		outF.Close()
		errF.Close()
		writeGitHubOutput(outputFile, errorFile)
		emit(map[string]any{"type": "done", "output_file": outputFile})
		return nil
	},
}

func writeGitHubOutput(outputFile, errorFile string) {
	ghOut := os.Getenv("GITHUB_OUTPUT")
	if ghOut == "" {
		return
	}
	f, err := os.OpenFile(ghOut, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "output_file=%s\n", outputFile)
	if info, _ := os.Stat(errorFile); info != nil && info.Size() > 0 {
		data, _ := os.ReadFile(errorFile)
		fmt.Fprintf(f, "error<<QODER_ERROR_EOF\n%s\nQODER_ERROR_EOF\n", string(data))
	} else {
		fmt.Fprintln(f, "error=")
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}

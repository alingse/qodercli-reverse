package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/core/commands"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "List discovered slash commands and subagents",
	Run: func(cmd *cobra.Command, args []string) {
		workDir, _ := os.Getwd()
		reg := commands.NewRegistry()
		reg.Discover(workDir)
		if len(os.Args) > 0 && contains(os.Args, "--json") {
			b, _ := json.MarshalIndent(reg, "", "  ")
			fmt.Println(string(b))
			return
		}
		fmt.Println("Slash commands:")
		for _, c := range reg.Commands {
			fmt.Printf("  %-20s %s (%s)\n", c.Name, c.Description, c.Source)
		}
		fmt.Println("\nSubagents:")
		for _, s := range reg.Subagents {
			fmt.Printf("  %-20s %s (%s)\n", s.Name, s.Description, s.Source)
		}
	},
}

func contains(slice []string, v string) bool {
	for _, s := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func init() {
	commandsCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(commandsCmd)
}

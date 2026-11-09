package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/cmd/utils"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics for provider and environment",
	Run: func(cmd *cobra.Command, args []string) {
		utils.InitLogger("", false)
		defer log.Close()

		fmt.Println("=== QoderCLI Reverse Doctor ===")

		// Check API key
		hasKey := false
		for _, k := range []string{"OPENAI_API_KEY", "QODER_PERSONAL_ACCESS_TOKEN", "QODER_OPENAI_API_KEY"} {
			if os.Getenv(k) != "" {
				hasKey = true
				fmt.Printf("API key present (%s): yes\n", k)
			}
		}
		if !hasKey {
			fmt.Println("API key present: no")
		}

		baseURL := os.Getenv("OPENAI_BASE_URL")
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		fmt.Printf("Base URL: %s\n", baseURL)
		model := os.Getenv("OPENAI_MODEL")
		if model == "" {
			model = os.Getenv("QODER_MODEL")
		}
		if model == "" {
			model = "auto"
		}
		fmt.Printf("Model: %s\n", model)

		// Try provider creation
		prov, override, err := utils.CreateProvider()
		if err != nil {
			fmt.Printf("Provider creation: FAILED (%v)\n", err)
			return
		}
		fmt.Println("Provider creation: OK")
		if override != "" {
			fmt.Printf("Model override: %s\n", override)
		}
		_ = prov
		fmt.Println("Diagnostics complete.")
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

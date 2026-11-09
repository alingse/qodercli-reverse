package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/core/skills"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage agent skills",
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Run: func(cmd *cobra.Command, args []string) {
		workDir, _ := os.Getwd()
		sks, _ := skills.DiscoverSkills(skills.SkillDirs(workDir)...)
		for _, s := range sks {
			fmt.Printf("%-20s %s (%s)\n", s.Name, s.Description, s.Source)
		}
		if len(sks) == 0 {
			fmt.Println("No skills installed.")
		}
	},
}

var skillSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search installed skills",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workDir, _ := os.Getwd()
		sks, _ := skills.DiscoverSkills(skills.SkillDirs(workDir)...)
		res := skills.Search(sks, strings.Join(args, " "))
		for _, s := range res {
			fmt.Printf("%-20s %s\n", s.Name, s.Description)
		}
		if len(res) == 0 {
			fmt.Println("No matching skills.")
		}
	},
}

var skillInspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Inspect a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workDir, _ := os.Getwd()
		sks, _ := skills.DiscoverSkills(skills.SkillDirs(workDir)...)
		for _, s := range sks {
			if s.Name == args[0] {
				b, _ := json.MarshalIndent(s, "", "  ")
				fmt.Println(string(b))
				fmt.Println("---")
				fmt.Println(s.Content)
				return
			}
		}
		fmt.Printf("Skill %q not found.\n", args[0])
	},
}

var skillInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a skill from a local path or source",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Local install: copy SKILL.md from given path into .qoder/skills/<name>
		src := args[0]
		info, err := os.Stat(src)
		if err != nil {
			fmt.Printf("Source not found: %s\n", src)
			return
		}
		workDir, _ := os.Getwd()
		targetDir := filepath.Join(workDir, ".qoder", "skills", filepath.Base(src))
		if info.IsDir() {
			targetDir = filepath.Join(workDir, ".qoder", "skills", filepath.Base(src))
		} else {
			targetDir = filepath.Join(workDir, ".qoder", "skills", strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)))
		}
		os.MkdirAll(targetDir, 0755)
		if info.IsDir() {
			// copy SKILL.md
			data, _ := os.ReadFile(filepath.Join(src, "SKILL.md"))
			os.WriteFile(filepath.Join(targetDir, "SKILL.md"), data, 0644)
		} else {
			data, _ := os.ReadFile(src)
			os.WriteFile(filepath.Join(targetDir, "SKILL.md"), data, 0644)
		}
		fmt.Printf("Installed skill to %s\n", targetDir)
	},
}

var skillRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an installed skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workDir, _ := os.Getwd()
		targetDir := filepath.Join(workDir, ".qoder", "skills", args[0])
		if err := os.RemoveAll(targetDir); err != nil {
			fmt.Printf("Failed to remove: %v\n", err)
			return
		}
		fmt.Printf("Removed skill %s\n", args[0])
	},
}

var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Manage community skill sources",
}

var communitySourcesCmd = &cobra.Command{
	Use:   "sources",
	Short: "Manage community sources",
}

var communitySourcesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List community sources",
	Run: func(cmd *cobra.Command, args []string) {
		reg := skills.NewRegistry()
		for _, s := range reg.Sources {
			fmt.Printf("%-20s %s (%s)\n", s.Name, s.URL, s.Ref)
		}
	},
}

var communitySourcesAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a community source",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		reg := skills.NewRegistry()
		ref, _ := cmd.Flags().GetString("ref")
		if err := reg.AddSource(args[0], args[1], ref); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Added source %s\n", args[0])
	},
}

var communitySourcesSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync community sources (metadata fetch is offline-cached)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Community sources configured. Offline cache used for installed skills.")
	},
}

func init() {
	skillCmd.AddCommand(skillListCmd, skillSearchCmd, skillInspectCmd, skillInstallCmd, skillRemoveCmd)
	communitySourcesCmd.AddCommand(communitySourcesListCmd, communitySourcesAddCmd, communitySourcesSyncCmd)
	communityCmd.AddCommand(communitySourcesCmd)
	rootCmd.AddCommand(skillCmd, communityCmd)
	communitySourcesAddCmd.Flags().String("ref", "main", "Branch/tag/commit reference")
}

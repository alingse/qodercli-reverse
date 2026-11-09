package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alingse/qodercli-reverse/decompiled/core/resource/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage Model Context Protocol servers",
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := mcp.LoadGlobalConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if len(cfg.Servers) == 0 {
			fmt.Println("No MCP servers configured")
			return
		}

		fmt.Println("Configured MCP servers:")
		for name, server := range cfg.Servers {
			fmt.Printf("\n  %s:\n", name)
			fmt.Printf("    Command: %s\n", server.Command)
			if server.Type != "" {
				fmt.Printf("    Type: %s\n", server.Type)
			}
			if server.URL != "" {
				fmt.Printf("    URL: %s\n", server.URL)
			}
			if len(server.Args) > 0 {
				argsStr := make([]string, len(server.Args))
				for i, a := range server.Args {
					argsStr[i] = a
				}
				fmt.Printf("    Args: %v\n", argsStr)
			}
			if len(server.Env) > 0 {
				envVars := make([]string, 0, len(server.Env))
				for k, v := range server.Env {
					envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
				}
				fmt.Printf("    Env: %v\n", envVars)
			}
			if server.Cwd != "" {
				fmt.Printf("    Cwd: %s\n", server.Cwd)
			}
		}
	},
}

var mcpAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := mcp.LoadGlobalConfig()
		if err != nil {
			cfg = &mcp.ServersConfig{Servers: make(map[string]*mcp.ServerConfig)}
		}

		// Check if server already exists
		if _, exists := cfg.Servers[name]; exists {
			fmt.Printf("MCP server %q already exists\n", name)
			return
		}

		// Prompt for server details
		fmt.Printf("Adding MCP server: %s\n", name)

		var command, serverType, url string
		var argsList []string
		env := make(map[string]string)
		var cwd string

		fmt.Print("Command (e.g., qoder-github-mcp-server): ")
		fmt.Scanln(&command)

		fmt.Print("Type (stdio/sse/http, default stdio): ")
		fmt.Scanln(&serverType)
		if serverType == "" {
			serverType = "stdio"
		}

		fmt.Print("URL (if HTTP type, leave empty for stdio): ")
		fmt.Scanln(&url)

		fmt.Print("Args (comma-separated, optional): ")
		argsInput := ""
		fmt.Scanln(&argsInput)
		if argsInput != "" {
			for _, a := range strings.Split(argsInput, ",") {
				a = strings.TrimSpace(a)
				if a != "" {
					argsList = append(argsList, a)
				}
			}
		}

		fmt.Print("Env vars (KEY=VALUE, comma-separated, optional): ")
		envInput := ""
		fmt.Scanln(&envInput)
		if envInput != "" {
			for _, e := range strings.Split(envInput, ",") {
				e = strings.TrimSpace(e)
				if strings.Contains(e, "=") {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
					}
				}
			}
		}

		fmt.Print("Cwd (optional, press enter to skip): ")
		fmt.Scanln(&cwd)
		if cwd == "" {
			cwd = ""
		}

		// Create server config
		cfg.Servers[name] = &mcp.ServerConfig{
			Name:    name,
			Command: command,
			Args:    argsList,
			Env:     env,
			Cwd:     cwd,
			Type:    serverType,
			URL:     url,
		}

		// Save config
		if err := mcp.SaveGlobalConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✓ MCP server %q added\n", name)
	},
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := mcp.LoadGlobalConfig()
		if err != nil {
			fmt.Println("No MCP config found")
			return
		}

		if _, exists := cfg.Servers[name]; !exists {
			fmt.Printf("MCP server %q not found\n", name)
			return
		}

		delete(cfg.Servers, name)
		if err := mcp.SaveGlobalConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✓ MCP server %q removed\n", name)
	},
}

var mcpEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an MCP server configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := mcp.LoadGlobalConfig()
		if err != nil {
			fmt.Println("No MCP config found")
			return
		}

		server, exists := cfg.Servers[name]
		if !exists {
			fmt.Printf("MCP server %q not found\n", name)
			return
		}

		fmt.Printf("Editing MCP server: %s\n", name)

		var command, serverType, url string
		var argsList []string
		env := server.Env
		var cwd string

		if server.Args != nil {
			argsList = make([]string, len(server.Args))
			copy(argsList, server.Args)
		}

		fmt.Print("Command (current: " + server.Command + ", press enter to keep): ")
		fmt.Scanln(&command)
		if command == "" {
			command = server.Command
		}

		fmt.Print("Type (current: " + server.Type + ", press enter to keep): ")
		fmt.Scanln(&serverType)
		if serverType == "" {
			serverType = server.Type
		}

		fmt.Print("URL (current: " + server.URL + ", press enter to keep): ")
		fmt.Scanln(&url)
		if url == "" {
			url = server.URL
		}

		fmt.Printf("Args (current: %v, comma-separated, press enter to keep): ", server.Args)
		argsInput := ""
		fmt.Scanln(&argsInput)
		if argsInput != "" {
			for _, a := range strings.Split(argsInput, ",") {
				a = strings.TrimSpace(a)
				if a != "" {
					argsList = append(argsList, a)
				}
			}
			if argsList == nil {
				argsList = server.Args
			}
		} else {
			argsList = server.Args
		}

		fmt.Print("Env vars (comma-separated, press enter to keep current): len(server.Env)")
		envInput := ""
		fmt.Scanln(&envInput)
		if envInput != "" {
			for _, e := range strings.Split(envInput, ",") {
				e = strings.TrimSpace(e)
				if strings.Contains(e, "=") {
					parts := strings.SplitN(e, "=", 2)
					if len(parts) == 2 {
						env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
					}
				}
			}
		}
		if env == nil {
			env = server.Env
		}

		fmt.Print("Cwd (current: " + server.Cwd + ", press enter to keep): ")
		fmt.Scanln(&cwd)
		if cwd == "" {
			cwd = server.Cwd
		}

		// Update server config
		server.Command = command
		server.Type = serverType
		server.URL = url
		server.Args = argsList
		server.Env = env
		server.Cwd = cwd

		if err := mcp.SaveGlobalConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("✓ MCP server %q updated\n", name)
	},
}

func init() {
	mcpCmd.AddCommand(mcpListCmd, mcpAddCmd, mcpRemoveCmd, mcpEditCmd)
	rootCmd.AddCommand(mcpCmd)
}
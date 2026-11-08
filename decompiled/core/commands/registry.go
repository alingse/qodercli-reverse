// Package commands provides slash-command and subagent discovery for qodercli-reverse.
package commands

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SlashCommand represents a discoverable slash command.
type SlashCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"` // "builtin", "project", "user"
	Path        string `json:"path,omitempty"`
}

// Subagent represents a discoverable subagent definition.
type Subagent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Path        string `json:"path,omitempty"`
}

// Registry holds discovered commands and subagents.
type Registry struct {
	Commands  []SlashCommand
	Subagents []Subagent
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) loadCommandsDir(dir, source string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		cmdName := strings.TrimSuffix(name, filepath.Ext(name))
		desc := extractDescription(string(data))
		r.Commands = append(r.Commands, SlashCommand{Name: "/" + cmdName, Description: desc, Source: source, Path: path})
	}
}

func (r *Registry) loadSubagentsDir(dir, source string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		agentName := strings.TrimSuffix(name, filepath.Ext(name))
		desc := extractDescription(string(data))
		r.Subagents = append(r.Subagents, Subagent{Name: agentName, Description: desc, Source: source, Path: path})
	}
}

// Discover populates the registry from built-ins, project-level (.qoder/) and user-level (~/.qoder/) extensions.
func (r *Registry) Discover(workDir string) {
	for _, c := range builtinCommands() {
		r.Commands = append(r.Commands, c)
	}
	for _, s := range builtinSubagents() {
		r.Subagents = append(r.Subagents, s)
	}

	if workDir != "" {
		r.loadCommandsDir(filepath.Join(workDir, ".qoder", "commands"), "project")
		r.loadSubagentsDir(filepath.Join(workDir, ".qoder", "subagents"), "project")
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		r.loadCommandsDir(filepath.Join(home, ".qoder", "commands"), "user")
		r.loadSubagentsDir(filepath.Join(home, ".qoder", "subagents"), "user")
	}

	r.dedup()
}

func builtinCommands() []SlashCommand {
	return []SlashCommand{
		{Name: "/review-pr", Description: "Run an AI-assisted review of a GitHub pull request", Source: "builtin"},
		{Name: "/assistant", Description: "Interactive GitHub assistant for issues and PRs", Source: "builtin"},
		{Name: "/security-audit", Description: "Run a security audit on the codebase", Source: "builtin"},
		{Name: "/generate-tests", Description: "Generate tests for a module", Source: "builtin"},
		{Name: "/explain-code", Description: "Explain selected code", Source: "builtin"},
		{Name: "/refactor", Description: "Refactor a function or module", Source: "builtin"},
		{Name: "/setup-github", Description: "Set up GitHub integration", Source: "builtin"},
	}
}

func builtinSubagents() []Subagent {
	return []Subagent{
		{Name: "code-reviewer", Description: "Reviews code for bugs and maintainability", Source: "builtin"},
		{Name: "security-auditor", Description: "Audits code for security vulnerabilities", Source: "builtin"},
		{Name: "test-engineer", Description: "Generates and improves tests", Source: "builtin"},
		{Name: "documentation-agent", Description: "Writes and updates documentation", Source: "builtin"},
		{Name: "refactoring-agent", Description: "Refactors code safely", Source: "builtin"},
	}
}

func extractDescription(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

func (r *Registry) dedup() {
	seenCmd := map[string]bool{}
	var cmds []SlashCommand
	for _, c := range r.Commands {
		if seenCmd[c.Name] {
			continue
		}
		seenCmd[c.Name] = true
		cmds = append(cmds, c)
	}
	r.Commands = cmds

	seenSub := map[string]bool{}
	var subs []Subagent
	for _, s := range r.Subagents {
		if seenSub[s.Name] {
			continue
		}
		seenSub[s.Name] = true
		subs = append(subs, s)
	}
	r.Subagents = subs

	sort.Slice(r.Commands, func(i, j int) bool { return r.Commands[i].Name < r.Commands[j].Name })
	sort.Slice(r.Subagents, func(i, j int) bool { return r.Subagents[i].Name < r.Subagents[j].Name })
}

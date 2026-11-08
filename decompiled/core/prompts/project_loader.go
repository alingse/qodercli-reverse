// Package prompts project context loader
// Responsible for loading project-specific instructions such as AGENTS.md, .claude/, .cursorrules, etc.
package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectContextLoader project context loader.
type ProjectContextLoader struct {
	workDir     string
	maxFileSize int64
	maxDepth    int
	cache       *ProjectContext
}

// NewProjectContextLoader creates a new loader.
func NewProjectContextLoader(workDir string) *ProjectContextLoader {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	return &ProjectContextLoader{
		workDir:     workDir,
		maxFileSize: 1024 * 1024, // 1MB
		maxDepth:    3,
	}
}

// SetMaxFileSize sets the maximum file size.
func (l *ProjectContextLoader) SetMaxFileSize(size int64) {
	l.maxFileSize = size
}

// Load loads the project context.
func (l *ProjectContextLoader) Load() (*ProjectContext, error) {
	if l.cache != nil {
		return l.cache, nil
	}

	ctx := &ProjectContext{
		RootPath:          l.workDir,
		GitignorePatterns: make([]string, 0),
	}

	// detect project type
	l.detectProjectType(ctx)

	// load AGENTS.md
	l.loadAgentsMD(ctx)

	// load .qoder/ directory
	l.loadQoderDir(ctx)

	// load .claude/ directory
	l.loadClaudeDir(ctx)

	// load .cursorrules
	l.loadCursorRules(ctx)

	// load README summary
	l.loadReadmeSummary(ctx)

	// load coding standards files
	l.loadCodingStandards(ctx)

	// load .gitignore
	l.loadGitignore(ctx)

	l.cache = ctx
	return ctx, nil
}

// detectProjectType detects the project type.
func (l *ProjectContextLoader) detectProjectType(ctx *ProjectContext) {
	// check various project config files
	if l.fileExists("go.mod") {
		ctx.Type = "go"
		ctx.Language = "Go"
		ctx.HasGoMod = true
		ctx.Name = l.readGoModuleName()
	} else if l.fileExists("package.json") {
		ctx.Type = "node"
		ctx.Language = "JavaScript/TypeScript"
		ctx.HasPackageJSON = true
		ctx.Name = l.readNodeProjectName()
	} else if l.fileExists("pyproject.toml") || l.fileExists("setup.py") || l.fileExists("requirements.txt") {
		ctx.Type = "python"
		ctx.Language = "Python"
		ctx.HasPyProject = true
		ctx.Name = l.readPythonProjectName()
	} else if l.fileExists("Cargo.toml") {
		ctx.Type = "rust"
		ctx.Language = "Rust"
		ctx.HasCargoToml = true
		ctx.Name = l.readRustProjectName()
	} else if l.fileExists("pom.xml") || l.fileExists("build.gradle") {
		ctx.Type = "java"
		ctx.Language = "Java"
		ctx.HasPomXML = true
	}

	// if no name detected, use directory name
	if ctx.Name == "" {
		ctx.Name = filepath.Base(l.workDir)
	}
}

// loadAgentsMD loads the AGENTS.md file (and case variants).
func (l *ProjectContextLoader) loadAgentsMD(ctx *ProjectContext) {
	variants := []string{"AGENTS.md", "Agents.md", "agents.md"}
	ctx.AgentsMdVariants = make(map[string]string)
	for _, name := range variants {
		content := l.readFileWithLimit(name)
		if content == "" {
			continue
		}
		ctx.AgentsMdVariants[name] = l.extractRelevantSections(content)
		if ctx.AgentsMDContent == "" {
			ctx.AgentsMDContent = ctx.AgentsMdVariants[name]
		}
	}
}

// loadClaudeDir loads the .claude/ directory.
func (l *ProjectContextLoader) loadClaudeDir(ctx *ProjectContext) {
	claudeDir := filepath.Join(l.workDir, ".claude")

	info, err := os.Stat(claudeDir)
	if err != nil || !info.IsDir() {
		return
	}

	var parts []string

	// read CLAUDE.md
	if content := l.readFileWithLimit(filepath.Join(".claude", "CLAUDE.md")); content != "" {
		parts = append(parts, "CLAUDE.md:")
		parts = append(parts, l.extractRelevantSections(content))
	}

	// read instructions.md
	if content := l.readFileWithLimit(filepath.Join(".claude", "instructions.md")); content != "" {
		parts = append(parts, "Instructions:")
		parts = append(parts, l.extractRelevantSections(content))
	}

	// read context.md
	if content := l.readFileWithLimit(filepath.Join(".claude", "context.md")); content != "" {
		parts = append(parts, "Context:")
		parts = append(parts, l.extractRelevantSections(content))
	}

	// read all .md files
	entries, err := os.ReadDir(claudeDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".md") &&
				name != "CLAUDE.md" &&
				name != "instructions.md" &&
				name != "context.md" {
				if content := l.readFileWithLimit(filepath.Join(".claude", name)); content != "" {
					parts = append(parts, fmt.Sprintf("%s:", name))
					parts = append(parts, l.extractRelevantSections(content))
				}
			}
		}
	}

	if len(parts) > 0 {
		ctx.ClaudeDirContent = strings.Join(parts, "\n\n")
	}
}

// loadQoderDir loads the .qoder/ directory (commands, subagents, settings, instructions).
func (l *ProjectContextLoader) loadQoderDir(ctx *ProjectContext) {
	qoderDir := filepath.Join(l.workDir, ".qoder")
	info, err := os.Stat(qoderDir)
	if err != nil || !info.IsDir() {
		return
	}

	var parts []string

	// Project-level instructions
	for _, name := range []string{"AGENTS.md", "instructions.md", "context.md", "README.md"} {
		if content := l.readFileWithLimit(filepath.Join(".qoder", name)); content != "" {
			parts = append(parts, fmt.Sprintf("%s:", name))
			parts = append(parts, l.extractRelevantSections(content))
		}
	}

	// commands/
	if entries, err := os.ReadDir(filepath.Join(qoderDir, "commands")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".md") {
				continue
			}
			if content := l.readFileWithLimit(filepath.Join(".qoder", "commands", name)); content != "" {
				parts = append(parts, fmt.Sprintf("command %s:", name))
				parts = append(parts, l.extractRelevantSections(content))
			}
		}
	}

	// subagents/
	if entries, err := os.ReadDir(filepath.Join(qoderDir, "subagents")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".json") {
				continue
			}
			if content := l.readFileWithLimit(filepath.Join(".qoder", "subagents", name)); content != "" {
				parts = append(parts, fmt.Sprintf("subagent %s:", name))
				parts = append(parts, l.extractRelevantSections(content))
			}
		}
	}

	// settings/
	if entries, err := os.ReadDir(filepath.Join(qoderDir, "settings")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if content := l.readFileWithLimit(filepath.Join(".qoder", "settings", name)); content != "" {
				parts = append(parts, fmt.Sprintf("setting %s:", name))
				parts = append(parts, l.truncateString(content, 1000))
			}
		}
	}

	if len(parts) > 0 {
		ctx.QoderDirContent = strings.Join(parts, "\n\n")
	}
}

// loadCursorRules loads the .cursorrules file.
func (l *ProjectContextLoader) loadCursorRules(ctx *ProjectContext) {
	content := l.readFileWithLimit(".cursorrules")
	if content != "" {
		// merge into coding standards
		if ctx.CodingStandards != "" {
			ctx.CodingStandards += "\n\n"
		}
		ctx.CodingStandards += "Cursor Rules:\n" + content
	}
}

// loadReadmeSummary loads the README summary.
func (l *ProjectContextLoader) loadReadmeSummary(ctx *ProjectContext) {
	content := l.readFileWithLimit("README.md")
	if content == "" {
		content = l.readFileWithLimit("README.rst")
	}
	if content == "" {
		content = l.readFileWithLimit("README")
	}

	if content != "" {
		// extract first 500 characters as summary
		ctx.ReadmeContent = l.truncateString(content, 500)
	}
}

// loadCodingStandards loads coding standards files.
func (l *ProjectContextLoader) loadCodingStandards(ctx *ProjectContext) {
	// load corresponding standards file based on project type
	switch ctx.Type {
	case "go":
		if content := l.readFileWithLimit(".golangci.yml"); content != "" {
			ctx.StyleGuide = "Using golangci-lint configuration"
		}
	case "node":
		if content := l.readFileWithLimit(".eslintrc.js"); content != "" {
			ctx.StyleGuide = "Using ESLint configuration"
		} else if content := l.readFileWithLimit(".eslintrc.json"); content != "" {
			ctx.StyleGuide = "Using ESLint configuration"
		}
		if content := l.readFileWithLimit("prettier.config.js"); content != "" {
			if ctx.StyleGuide != "" {
				ctx.StyleGuide += "\n"
			}
			ctx.StyleGuide += "Using Prettier configuration"
		}
	case "python":
		if content := l.readFileWithLimit("pyproject.toml"); content != "" {
			ctx.StyleGuide = "Using pyproject.toml configuration"
		} else if content := l.readFileWithLimit("setup.cfg"); content != "" {
			ctx.StyleGuide = "Using setup.cfg configuration"
		}
	case "rust":
		if content := l.readFileWithLimit("rustfmt.toml"); content != "" {
			ctx.StyleGuide = "Using rustfmt configuration"
		}
	}

	// generic standards files
	if content := l.readFileWithLimit("CONTRIBUTING.md"); content != "" {
		if ctx.CodingStandards != "" {
			ctx.CodingStandards += "\n\n"
		}
		ctx.CodingStandards += "Contributing Guidelines:\n" + l.extractRelevantSections(content)
	}

	if content := l.readFileWithLimit("STYLE.md"); content != "" {
		if ctx.CodingStandards != "" {
			ctx.CodingStandards += "\n\n"
		}
		ctx.CodingStandards += "Style Guide:\n" + content
	}
}

// loadGitignore loads .gitignore patterns.
func (l *ProjectContextLoader) loadGitignore(ctx *ProjectContext) {
	content := l.readFileWithLimit(".gitignore")
	if content == "" {
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// ignore comments and empty lines
		if line != "" && !strings.HasPrefix(line, "#") {
			ctx.GitignorePatterns = append(ctx.GitignorePatterns, line)
		}
	}
}

// ==================== helper methods ====================

func (l *ProjectContextLoader) fileExists(name string) bool {
	path := filepath.Join(l.workDir, name)
	_, err := os.Stat(path)
	return err == nil
}

func (l *ProjectContextLoader) readFileWithLimit(path string) string {
	fullPath := filepath.Join(l.workDir, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return ""
	}

	// check file size
	if info.Size() > l.maxFileSize {
		return fmt.Sprintf("[File too large: %d bytes]", info.Size())
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}

	return string(data)
}

func (l *ProjectContextLoader) extractRelevantSections(content string) string {
	// extract relevant sections, filtering out irrelevant content
	lines := strings.Split(content, "\n")
	var relevant []string

	for _, line := range lines {
		// keep headings, list items, and important paragraphs
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "1.") ||
			strings.HasPrefix(trimmed, "IMPORTANT:") ||
			strings.HasPrefix(trimmed, "CRITICAL:") ||
			strings.HasPrefix(trimmed, "ALWAYS:") ||
			strings.HasPrefix(trimmed, "NEVER:") {
			relevant = append(relevant, line)
		} else if len(trimmed) > 0 && len(relevant) > 0 {
			// keep non-empty lines; they may be part of a paragraph
			relevant = append(relevant, line)
		}
	}

	return strings.Join(relevant, "\n")
}

func (l *ProjectContextLoader) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ==================== project-specific read methods ====================

func (l *ProjectContextLoader) readGoModuleName() string {
	content := l.readFileWithLimit("go.mod")
	if content == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return ""
}

func (l *ProjectContextLoader) readNodeProjectName() string {
	content := l.readFileWithLimit("package.json")
	if content == "" {
		return ""
	}

	// simple string extraction
	const nameKey = `"name"`
	idx := strings.Index(content, nameKey)
	if idx == -1 {
		return ""
	}

	// find value
	rest := content[idx+len(nameKey):]
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, ":") {
		rest = strings.TrimSpace(rest[1:])
		// extract value in quotes
		if strings.HasPrefix(rest, `"`) {
			rest = rest[1:]
			endIdx := strings.Index(rest, `"`)
			if endIdx != -1 {
				return rest[:endIdx]
			}
		}
	}
	return ""
}

func (l *ProjectContextLoader) readPythonProjectName() string {
	// try reading from pyproject.toml
	content := l.readFileWithLimit("pyproject.toml")
	if content != "" {
		const nameKey = `name`
		idx := strings.Index(content, nameKey)
		if idx != -1 {
			rest := content[idx+len(nameKey):]
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "=") {
				rest = strings.TrimSpace(rest[1:])
				rest = strings.Trim(rest, `"'`)
				return rest
			}
		}
	}
	return ""
}

func (l *ProjectContextLoader) readRustProjectName() string {
	content := l.readFileWithLimit("Cargo.toml")
	if content == "" {
		return ""
	}

	const nameKey = `name`
	idx := strings.Index(content, nameKey)
	if idx == -1 {
		return ""
	}

	rest := content[idx+len(nameKey):]
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "=") {
		rest = strings.TrimSpace(rest[1:])
		rest = strings.Trim(rest, `"'`)
		return rest
	}
	return ""
}

// ClearCache clears the cache.
func (l *ProjectContextLoader) ClearCache() {
	l.cache = nil
}

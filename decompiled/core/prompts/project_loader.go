// Package prompts 项目上下文加载器
// 负责加载 AGENTS.md, .claude/, .cursorrules 等项目特定指令
package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectContextLoader 项目上下文加载器
type ProjectContextLoader struct {
	workDir         string
	maxFileSize     int64
	maxDepth        int
	cache           *ProjectContext
}

// NewProjectContextLoader 创建新的加载器
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

// SetMaxFileSize 设置最大文件大小
func (l *ProjectContextLoader) SetMaxFileSize(size int64) {
	l.maxFileSize = size
}

// Load 加载项目上下文
func (l *ProjectContextLoader) Load() (*ProjectContext, error) {
	if l.cache != nil {
		return l.cache, nil
	}

	ctx := &ProjectContext{
		RootPath:          l.workDir,
		GitignorePatterns: make([]string, 0),
	}

	// 检测项目类型
	l.detectProjectType(ctx)

	// 加载 AGENTS.md
	l.loadAgentsMD(ctx)

	// 加载 .claude/ 目录
	l.loadClaudeDir(ctx)

	// 加载 .cursorrules
	l.loadCursorRules(ctx)

	// 加载 README 摘要
	l.loadReadmeSummary(ctx)

	// 加载编码规范文件
	l.loadCodingStandards(ctx)

	// 加载 .gitignore
	l.loadGitignore(ctx)

	l.cache = ctx
	return ctx, nil
}

// detectProjectType 检测项目类型
func (l *ProjectContextLoader) detectProjectType(ctx *ProjectContext) {
	// 检查各种项目配置文件
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

	// 如果没有检测到名称，使用目录名
	if ctx.Name == "" {
		ctx.Name = filepath.Base(l.workDir)
	}
}

// loadAgentsMD 加载 AGENTS.md 文件
func (l *ProjectContextLoader) loadAgentsMD(ctx *ProjectContext) {
	content := l.readFileWithLimit("AGENTS.md")
	if content == "" {
		// 尝试小写版本
		content = l.readFileWithLimit("agents.md")
	}

	if content != "" {
		ctx.AgentsMDContent = l.extractRelevantSections(content)
	}
}

// loadClaudeDir 加载 .claude/ 目录
func (l *ProjectContextLoader) loadClaudeDir(ctx *ProjectContext) {
	claudeDir := filepath.Join(l.workDir, ".claude")
	
	info, err := os.Stat(claudeDir)
	if err != nil || !info.IsDir() {
		return
	}

	var parts []string

	// 读取 CLAUDE.md
	if content := l.readFileWithLimit(filepath.Join(".claude", "CLAUDE.md")); content != "" {
		parts = append(parts, "CLAUDE.md:")
		parts = append(parts, l.extractRelevantSections(content))
	}

	// 读取 instructions.md
	if content := l.readFileWithLimit(filepath.Join(".claude", "instructions.md")); content != "" {
		parts = append(parts, "Instructions:")
		parts = append(parts, l.extractRelevantSections(content))
	}

	// 读取 context.md
	if content := l.readFileWithLimit(filepath.Join(".claude", "context.md")); content != "" {
		parts = append(parts, "Context:")
		parts = append(parts, l.extractRelevantSections(content))
	}

	// 读取所有 .md 文件
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

// loadCursorRules 加载 .cursorrules 文件
func (l *ProjectContextLoader) loadCursorRules(ctx *ProjectContext) {
	content := l.readFileWithLimit(".cursorrules")
	if content != "" {
		// 合并到编码规范中
		if ctx.CodingStandards != "" {
			ctx.CodingStandards += "\n\n"
		}
		ctx.CodingStandards += "Cursor Rules:\n" + content
	}
}

// loadReadmeSummary 加载 README 摘要
func (l *ProjectContextLoader) loadReadmeSummary(ctx *ProjectContext) {
	content := l.readFileWithLimit("README.md")
	if content == "" {
		content = l.readFileWithLimit("README.rst")
	}
	if content == "" {
		content = l.readFileWithLimit("README")
	}

	if content != "" {
		// 提取前 500 个字符作为摘要
		ctx.ReadmeContent = l.truncateString(content, 500)
	}
}

// loadCodingStandards 加载编码规范文件
func (l *ProjectContextLoader) loadCodingStandards(ctx *ProjectContext) {
	// 根据项目类型加载对应的规范文件
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

	// 通用规范文件
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

// loadGitignore 加载 .gitignore 模式
func (l *ProjectContextLoader) loadGitignore(ctx *ProjectContext) {
	content := l.readFileWithLimit(".gitignore")
	if content == "" {
		return
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 忽略注释和空行
		if line != "" && !strings.HasPrefix(line, "#") {
			ctx.GitignorePatterns = append(ctx.GitignorePatterns, line)
		}
	}
}

// ==================== 辅助方法 ====================

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

	// 检查文件大小
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
	// 提取相关章节，过滤掉不相关的内容
	lines := strings.Split(content, "\n")
	var relevant []string

	for _, line := range lines {
		// 保留标题、列表项和重要段落
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
			// 保留非空行，可能是段落的一部分
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

// ==================== 项目特定读取方法 ====================

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

	// 简单的字符串提取
	const nameKey = `"name"`
	idx := strings.Index(content, nameKey)
	if idx == -1 {
		return ""
	}

	// 找到值
	rest := content[idx+len(nameKey):]
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, ":") {
		rest = strings.TrimSpace(rest[1:])
		// 提取引号中的值
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
	// 尝试从 pyproject.toml 读取
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

// ClearCache 清除缓存
func (l *ProjectContextLoader) ClearCache() {
	l.cache = nil
}

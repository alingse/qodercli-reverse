// Package tools 文件操作工具实现
// 反编译自 qodercli v0.1.29
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ReadTool Read 工具
type ReadTool struct {
	BaseTool
	maxFileSize int64
}

// ReadParams Read 参数
type ReadParams struct {
	FilePath  string `json:"file_path"`
	Offset    int    `json:"offset,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	ReadImage bool   `json:"read_image,omitempty"`
}

// NewReadTool 创建 Read 工具
func NewReadTool(maxFileSize int64) *ReadTool {
	return &ReadTool{
		BaseTool: BaseTool{
			name:        "Read",
			description: "Read the contents of a file",
			inputSchema: BuildFileSchema(),
		},
		maxFileSize: maxFileSize,
	}
}

// Execute 执行 Read
func (t *ReadTool) Execute(ctx context.Context, input string) (string, error) {
	var params ReadParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证路径
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// 检查文件是否存在
	info, err := os.Stat(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, use LS tool instead")
	}

	// 检查文件大小
	if info.Size() > t.maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), t.maxFileSize)
	}

	// 读取文件
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 处理行号
	lines := strings.Split(string(content), "\n")
	start := params.Offset
	if start < 0 {
		start = 0
	}
	if start >= len(lines) {
		return "", fmt.Errorf("offset exceeds file length")
	}

	end := len(lines)
	if params.Limit > 0 {
		end = start + params.Limit
		if end > len(lines) {
			end = len(lines)
		}
	}

	// 添加行号
	var result strings.Builder
	for i := start; i < end; i++ {
		result.WriteString(fmt.Sprintf("%d\t%s\n", i+1, lines[i]))
	}

	return result.String(), nil
}

// WriteTool Write 工具
type WriteTool struct {
	BaseTool
}

// WriteParams Write 参数
type WriteParams struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// NewWriteTool 创建 Write 工具
func NewWriteTool() *WriteTool {
	return &WriteTool{
		BaseTool: BaseTool{
			name:        "Write",
			description: "Create or overwrite a file",
			inputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "The absolute path to the file",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"file_path", "content"},
			},
		},
	}
}

// Execute 执行 Write
func (t *WriteTool) Execute(ctx context.Context, input string) (string, error) {
	var params WriteParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证路径
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// 创建目录
	dir := filepath.Dir(params.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// 原子写入
	tmpFile := params.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(params.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if err := os.Rename(tmpFile, params.FilePath); err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	return fmt.Sprintf("File written successfully: %s", params.FilePath), nil
}

// EditTool Edit 工具
type EditTool struct {
	BaseTool
	fileHistory map[string][]string
}

// EditParams Edit 参数
type EditParams struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// NewEditTool 创建 Edit 工具
func NewEditTool() *EditTool {
	return &EditTool{
		BaseTool: BaseTool{
			name:        "Edit",
			description: "Replace text in a file using search and replace",
			inputSchema: BuildEditSchema(),
		},
		fileHistory: make(map[string][]string),
	}
}

// Execute 执行 Edit
func (t *EditTool) Execute(ctx context.Context, input string) (string, error) {
	var params EditParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证路径
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// 读取文件
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 保存历史
	t.fileHistory[params.FilePath] = append(t.fileHistory[params.FilePath], string(content))

	// 执行替换
	oldContent := string(content)
	var newContent string

	if params.ReplaceAll {
		newContent = strings.ReplaceAll(oldContent, params.OldString, params.NewString)
	} else {
		newContent = strings.Replace(oldContent, params.OldString, params.NewString, 1)
	}

	if newContent == oldContent {
		return "", fmt.Errorf("old_string not found in file")
	}

	// 写入文件
	tmpFile := params.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if err := os.Rename(tmpFile, params.FilePath); err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	return fmt.Sprintf("File edited successfully: %s", params.FilePath), nil
}

// DeleteFileTool DeleteFile 工具
type DeleteFileTool struct {
	BaseTool
}

// DeleteFileParams DeleteFile 参数
type DeleteFileParams struct {
	FilePath string `json:"file_path"`
}

// NewDeleteFileTool 创建 DeleteFile 工具
func NewDeleteFileTool() *DeleteFileTool {
	return &DeleteFileTool{
		BaseTool: BaseTool{
			name:        "DeleteFile",
			description: "Delete a file",
			inputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "The absolute path to the file to delete",
					},
				},
				"required": []string{"file_path"},
			},
		},
	}
}

// Execute 执行 DeleteFile
func (t *DeleteFileTool) Execute(ctx context.Context, input string) (string, error) {
	var params DeleteFileParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证路径
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// 检查文件是否存在
	if _, err := os.Stat(params.FilePath); err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// 删除文件
	if err := os.Remove(params.FilePath); err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("File deleted successfully: %s", params.FilePath), nil
}

// GlobTool Glob 工具
type GlobTool struct {
	BaseTool
}

// GlobParams Glob 参数
type GlobParams struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

// NewGlobTool 创建 Glob 工具
func NewGlobTool() *GlobTool {
	return &GlobTool{
		BaseTool: BaseTool{
			name:        "Glob",
			description: "Find files matching a glob pattern",
			inputSchema: BuildGlobSchema(),
		},
	}
}

// Execute 执行 Glob
func (t *GlobTool) Execute(ctx context.Context, input string) (string, error) {
	var params GlobParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 设置搜索路径
	searchPath := params.Path
	if searchPath == "" {
		searchPath = "."
	}

	// 验证路径
	if !IsSafePath(searchPath) {
		return "", fmt.Errorf("unsafe path: %s", searchPath)
	}

	// 执行 glob
	pattern := filepath.Join(searchPath, params.Pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob error: %w", err)
	}

	// 格式化输出
	var result strings.Builder
	for _, match := range matches {
		result.WriteString(match)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// GrepTool Grep 工具
type GrepTool struct {
	BaseTool
	ripgrepPath string
}

// GrepParams Grep 参数
type GrepParams struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path,omitempty"`
	OutputMode string `json:"output_mode,omitempty"` // content, files_with_matches, count
	Type       string `json:"type,omitempty"`        // 文件类型
	Glob       string `json:"glob,omitempty"`
	HeadLimit  int    `json:"head_limit,omitempty"`
	Multiline  bool   `json:"multiline,omitempty"`
	Before     int    `json:"-B,omitempty"`
	After      int    `json:"-A,omitempty"`
	Context    int    `json:"-C,omitempty"`
}

// GrepResult Grep 结果
type GrepResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Column  int    `json:"column,omitempty"`
	Content string `json:"content,omitempty"`
}

// NewGrepTool 创建 Grep 工具
func NewGrepTool(ripgrepPath string) *GrepTool {
	return &GrepTool{
		BaseTool: BaseTool{
			name:        "Grep",
			description: "Search file contents using regex",
			inputSchema: BuildGrepSchema(),
		},
		ripgrepPath: ripgrepPath,
	}
}

// Execute 执行 Grep
func (t *GrepTool) Execute(ctx context.Context, input string) (string, error) {
	var params GrepParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 设置搜索路径
	searchPath := params.Path
	if searchPath == "" {
		searchPath = "."
	}

	// 验证路径
	if !IsSafePath(searchPath) {
		return "", fmt.Errorf("unsafe path: %s", searchPath)
	}

	// 构建 ripgrep 命令
	args := []string{
		"--json",
		"--line-number",
		"--column",
	}

	// 输出模式
	switch params.OutputMode {
	case "files_with_matches":
		args = append(args, "-l")
	case "count":
		args = append(args, "-c")
	}

	// 限制结果数
	if params.HeadLimit > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", params.HeadLimit))
	}

	// 多行模式
	if params.Multiline {
		args = append(args, "--multiline")
	}

	// 上下文
	if params.Context > 0 {
		args = append(args, "-C", fmt.Sprintf("%d", params.Context))
	} else {
		if params.Before > 0 {
			args = append(args, "-B", fmt.Sprintf("%d", params.Before))
		}
		if params.After > 0 {
			args = append(args, "-A", fmt.Sprintf("%d", params.After))
		}
	}

	// 文件类型
	if params.Type != "" {
		args = append(args, "-t", params.Type)
	}

	// glob 模式
	if params.Glob != "" {
		args = append(args, "-g", params.Glob)
	}

	// 添加模式和路径
	args = append(args, params.Pattern, searchPath)

	// 执行 ripgrep
	cmd := exec.CommandContext(ctx, t.ripgrepPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return "", fmt.Errorf("grep failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// GetRipgrepPath 获取 ripgrep 路径
func GetRipgrepPath() string {
	// 尝试常见路径
	paths := []string{
		"rg",
		"/usr/local/bin/rg",
		"/opt/homebrew/bin/rg",
		"/usr/bin/rg",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}

	return "rg"
}

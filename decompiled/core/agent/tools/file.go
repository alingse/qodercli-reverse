// Package tools file operation tool implementation
// Decompiled from qodercli-reverse v0.1.29
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

// ReadTool Read tool
type ReadTool struct {
	BaseTool
	maxFileSize int64
}

// ReadParams Read parameters
type ReadParams struct {
	FilePath  string `json:"file_path"`
	Offset    int    `json:"offset,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	ReadImage bool   `json:"read_image,omitempty"`
}

// NewReadTool creates the Read tool
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

// Execute runs Read
func (t *ReadTool) Execute(ctx context.Context, input string) (string, error) {
	var params ReadParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// validate path
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// check if file exists
	info, err := os.Stat(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, use LS tool instead")
	}

	// check file size
	if info.Size() > t.maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), t.maxFileSize)
	}

	// read file
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// handle line numbers
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

	// add line numbers
	var result strings.Builder
	for i := start; i < end; i++ {
		result.WriteString(fmt.Sprintf("%d\t%s\n", i+1, lines[i]))
	}

	return result.String(), nil
}

// WriteTool Write tool
type WriteTool struct {
	BaseTool
}

// WriteParams Write parameters
type WriteParams struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// NewWriteTool creates the Write tool
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

// Execute runs Write
func (t *WriteTool) Execute(ctx context.Context, input string) (string, error) {
	var params WriteParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// validate path
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// create directory
	dir := filepath.Dir(params.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// atomic write
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

// EditTool Edit tool
type EditTool struct {
	BaseTool
	fileHistory map[string][]string
}

// EditParams Edit parameters
type EditParams struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// NewEditTool creates the Edit tool
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

// Execute runs Edit
func (t *EditTool) Execute(ctx context.Context, input string) (string, error) {
	var params EditParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// validate path
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// read file
	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// save history
	t.fileHistory[params.FilePath] = append(t.fileHistory[params.FilePath], string(content))

	// perform replacement
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

	// write file
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

// DeleteFileTool DeleteFile tool
type DeleteFileTool struct {
	BaseTool
}

// DeleteFileParams DeleteFile parameters
type DeleteFileParams struct {
	FilePath string `json:"file_path"`
}

// NewDeleteFileTool creates the DeleteFile tool
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

// Execute runs DeleteFile
func (t *DeleteFileTool) Execute(ctx context.Context, input string) (string, error) {
	var params DeleteFileParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// validate path
	if !IsSafePath(params.FilePath) {
		return "", fmt.Errorf("unsafe path: %s", params.FilePath)
	}

	// check if file exists
	if _, err := os.Stat(params.FilePath); err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	// delete file
	if err := os.Remove(params.FilePath); err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("File deleted successfully: %s", params.FilePath), nil
}

// GlobTool Glob tool
type GlobTool struct {
	BaseTool
}

// GlobParams Glob parameters
type GlobParams struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

// NewGlobTool creates the Glob tool
func NewGlobTool() *GlobTool {
	return &GlobTool{
		BaseTool: BaseTool{
			name:        "Glob",
			description: "Find files matching a glob pattern",
			inputSchema: BuildGlobSchema(),
		},
	}
}

// Execute runs Glob
func (t *GlobTool) Execute(ctx context.Context, input string) (string, error) {
	var params GlobParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// set search path
	searchPath := params.Path
	if searchPath == "" {
		searchPath = "."
	}

	// validate path
	if !IsSafePath(searchPath) {
		return "", fmt.Errorf("unsafe path: %s", searchPath)
	}

	// execute glob
	pattern := filepath.Join(searchPath, params.Pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob error: %w", err)
	}

	// format output
	var result strings.Builder
	for _, match := range matches {
		result.WriteString(match)
		result.WriteString("\n")
	}

	return result.String(), nil
}

// GrepTool Grep tool
type GrepTool struct {
	BaseTool
	ripgrepPath string
}

// GrepParams Grep parameters
type GrepParams struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path,omitempty"`
	OutputMode string `json:"output_mode,omitempty"` // content, files_with_matches, count
	Type       string `json:"type,omitempty"`        // file type
	Glob       string `json:"glob,omitempty"`
	HeadLimit  int    `json:"head_limit,omitempty"`
	Multiline  bool   `json:"multiline,omitempty"`
	Before     int    `json:"-B,omitempty"`
	After      int    `json:"-A,omitempty"`
	Context    int    `json:"-C,omitempty"`
}

// GrepResult Grep result
type GrepResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Column  int    `json:"column,omitempty"`
	Content string `json:"content,omitempty"`
}

// NewGrepTool creates the Grep tool
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

// Execute runs Grep
func (t *GrepTool) Execute(ctx context.Context, input string) (string, error) {
	var params GrepParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// set search path
	searchPath := params.Path
	if searchPath == "" {
		searchPath = "."
	}

	// validate path
	if !IsSafePath(searchPath) {
		return "", fmt.Errorf("unsafe path: %s", searchPath)
	}

	// build ripgrep command
	args := []string{
		"--json",
		"--line-number",
		"--column",
	}

	// output mode
	switch params.OutputMode {
	case "files_with_matches":
		args = append(args, "-l")
	case "count":
		args = append(args, "-c")
	}

	// limit result count
	if params.HeadLimit > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", params.HeadLimit))
	}

	// multiline mode
	if params.Multiline {
		args = append(args, "--multiline")
	}

	// context
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

	// file type
	if params.Type != "" {
		args = append(args, "-t", params.Type)
	}

	// glob pattern
	if params.Glob != "" {
		args = append(args, "-g", params.Glob)
	}

	// add pattern and path
	args = append(args, params.Pattern, searchPath)

	// execute ripgrep
	cmd := exec.CommandContext(ctx, t.ripgrepPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return "", fmt.Errorf("grep failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// GetRipgrepPath gets ripgrep path
func GetRipgrepPath() string {
	// try common paths
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

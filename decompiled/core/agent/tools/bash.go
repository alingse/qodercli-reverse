// Package tools Bash 工具实现
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
	"time"
)

// BashTool Bash 工具
type BashTool struct {
	BaseTool
	shellManager ShellManager
	timeout      time.Duration
}

// BashParams Bash 参数
type BashParams struct {
	Command         string `json:"command"`
	Description     string `json:"description,omitempty"`
	Timeout         int    `json:"timeout,omitempty"`
	RunInBackground bool   `json:"run_in_background,omitempty"`
}

// BashResult Bash 执行结果
type BashResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ShellID  string `json:"shell_id,omitempty"`
}

// ShellManager Shell 管理器接口
type ShellManager interface {
	Execute(ctx context.Context, command string, timeout time.Duration) (*ShellResult, error)
	ExecuteBackground(command string) (string, error)
	Kill(shellID string) error
	GetOutput(shellID string) (*ShellResult, error)
	Snapshot() (*ShellSnapshot, error)
	Restore(snapshot *ShellSnapshot) error
}

// ShellResult Shell 执行结果
type ShellResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	ShellID  string
	Running  bool
}

// ShellSnapshot Shell 快照
type ShellSnapshot struct {
	ID        string
	Directory string
	Env       map[string]string
}

// NewBashTool 创建 Bash 工具
func NewBashTool(manager ShellManager, defaultTimeout time.Duration) *BashTool {
	return &BashTool{
		BaseTool: BaseTool{
			name:        "Bash",
			description: "Execute shell commands",
			inputSchema: BuildBashSchema(),
		},
		shellManager: manager,
		timeout:      defaultTimeout,
	}
}

// Execute 执行 Bash 命令
func (t *BashTool) Execute(ctx context.Context, input string) (string, error) {
	var params BashParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 验证命令
	if err := t.validateCommand(params.Command); err != nil {
		return "", err
	}

	// 设置超时
	timeout := t.timeout
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Millisecond
	}

	// 后台执行
	if params.RunInBackground {
		shellID, err := t.shellManager.ExecuteBackground(params.Command)
		if err != nil {
			return "", fmt.Errorf("failed to start background command: %w", err)
		}

		result := BashResult{
			ShellID: shellID,
			Stdout:  fmt.Sprintf("Command started in background with ID: %s", shellID),
		}

		output, _ := json.Marshal(result)
		return string(output), nil
	}

	// 前台执行
	shellResult, err := t.shellManager.Execute(ctx, params.Command, timeout)
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	result := BashResult{
		ExitCode: shellResult.ExitCode,
		Stdout:   shellResult.Stdout,
		Stderr:   shellResult.Stderr,
	}

	output, _ := json.Marshal(result)
	return string(output), nil
}

// validateCommand 验证命令
func (t *BashTool) validateCommand(command string) error {
	// 检查危险命令
	dangerousPatterns := []string{
		"rm -rf /",
		"> /dev/sda",
		"mkfs.",
		"dd if=/dev/zero of=/dev/sda",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(command, pattern) {
			return fmt.Errorf("dangerous command detected: %s", pattern)
		}
	}

	return nil
}

// BashOutputTool BashOutput 工具
type BashOutputTool struct {
	BaseTool
	shellManager ShellManager
}

// BashOutputParams BashOutput 参数
type BashOutputParams struct {
	ShellID string `json:"shell_id"`
}

// NewBashOutputTool 创建 BashOutput 工具
func NewBashOutputTool(manager ShellManager) *BashOutputTool {
	return &BashOutputTool{
		BaseTool: BaseTool{
			name:        "BashOutput",
			description: "Get output from a background bash command",
			inputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"shell_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the background shell",
					},
				},
				"required": []string{"shell_id"},
			},
		},
		shellManager: manager,
	}
}

// Execute 执行 BashOutput
func (t *BashOutputTool) Execute(ctx context.Context, input string) (string, error) {
	var params BashOutputParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	result, err := t.shellManager.GetOutput(params.ShellID)
	if err != nil {
		return "", fmt.Errorf("failed to get output: %w", err)
	}

	bashResult := BashResult{
		ExitCode: result.ExitCode,
		Stdout:   result.Stdout,
		Stderr:   result.Stderr,
		ShellID:  result.ShellID,
	}

	output, _ := json.Marshal(bashResult)
	return string(output), nil
}

// KillBashTool KillBash 工具
type KillBashTool struct {
	BaseTool
	shellManager ShellManager
}

// KillBashParams KillBash 参数
type KillBashParams struct {
	ShellID string `json:"shell_id"`
}

// NewKillBashTool 创建 KillBash 工具
func NewKillBashTool(manager ShellManager) *KillBashTool {
	return &KillBashTool{
		BaseTool: BaseTool{
			name:        "KillBash",
			description: "Kill a background bash command",
			inputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"shell_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the background shell to kill",
					},
				},
				"required": []string{"shell_id"},
			},
		},
		shellManager: manager,
	}
}

// Execute 执行 KillBash
func (t *KillBashTool) Execute(ctx context.Context, input string) (string, error) {
	var params KillBashParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if err := t.shellManager.Kill(params.ShellID); err != nil {
		return "", fmt.Errorf("failed to kill shell: %w", err)
	}

	return fmt.Sprintf("Shell %s killed successfully", params.ShellID), nil
}

// defaultShellManager 默认 Shell 管理器
type defaultShellManager struct {
	shells   map[string]*runningShell
	bashPath string
}

// runningShell 运行中的 Shell
type runningShell struct {
	ID       string
	Cmd      *exec.Cmd
	Stdout   strings.Builder
	Stderr   strings.Builder
	Running  bool
	ExitCode int
	Done     chan struct{}
}

// NewDefaultShellManager 创建默认 Shell 管理器
func NewDefaultShellManager() ShellManager {
	bashPath := os.Getenv("QODER_BASH_PATH")
	if bashPath == "" {
		bashPath = "/bin/bash"
	}

	return &defaultShellManager{
		shells:   make(map[string]*runningShell),
		bashPath: bashPath,
	}
}

// Execute 执行命令
func (m *defaultShellManager) Execute(ctx context.Context, command string, timeout time.Duration) (*ShellResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, m.bashPath, "-c", command)
	cmd.Dir, _ = os.Getwd()
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()

	result := &ShellResult{
		ExitCode: cmd.ProcessState.ExitCode(),
		Stdout:   string(output),
		Stderr:   "",
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %v", timeout)
	}

	return result, nil
}

// ExecuteBackground 后台执行命令
func (m *defaultShellManager) ExecuteBackground(command string) (string, error) {
	shellID := generateShellID()

	cmd := exec.Command(m.bashPath, "-c", command)
	cmd.Dir, _ = os.Getwd()
	cmd.Env = os.Environ()

	shell := &runningShell{
		ID:      shellID,
		Cmd:     cmd,
		Running: true,
		Done:    make(chan struct{}),
	}

	m.shells[shellID] = shell

	go func() {
		output, _ := cmd.CombinedOutput()
		shell.Stdout.Write(output)
		shell.Running = false
		shell.ExitCode = cmd.ProcessState.ExitCode()
		close(shell.Done)
	}()

	return shellID, nil
}

// Kill 终止 Shell
func (m *defaultShellManager) Kill(shellID string) error {
	shell, exists := m.shells[shellID]
	if !exists {
		return fmt.Errorf("shell %s not found", shellID)
	}

	if shell.Running && shell.Cmd.Process != nil {
		return shell.Cmd.Process.Kill()
	}

	return nil
}

// GetOutput 获取输出
func (m *defaultShellManager) GetOutput(shellID string) (*ShellResult, error) {
	shell, exists := m.shells[shellID]
	if !exists {
		return nil, fmt.Errorf("shell %s not found", shellID)
	}

	return &ShellResult{
		ExitCode: shell.ExitCode,
		Stdout:   shell.Stdout.String(),
		Stderr:   shell.Stderr.String(),
		ShellID:  shellID,
		Running:  shell.Running,
	}, nil
}

// Snapshot 创建快照
func (m *defaultShellManager) Snapshot() (*ShellSnapshot, error) {
	dir, _ := os.Getwd()
	return &ShellSnapshot{
		ID:        generateShellID(),
		Directory: dir,
		Env:       make(map[string]string),
	}, nil
}

// Restore 恢复快照
func (m *defaultShellManager) Restore(snapshot *ShellSnapshot) error {
	if snapshot.Directory != "" {
		os.Chdir(snapshot.Directory)
	}
	return nil
}

// generateShellID 生成 Shell ID
func generateShellID() string {
	return fmt.Sprintf("shell-%d", time.Now().UnixNano())
}

// GetBashPath 获取 Bash 路径
func GetBashPath() string {
	if path := os.Getenv("QODER_BASH_PATH"); path != "" {
		return path
	}
	return "/bin/bash"
}

// IsSafePath 检查路径是否安全
func IsSafePath(path string) bool {
	// 检查是否在允许的路径范围内
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" && strings.HasPrefix(absPath, homeDir) {
		return true
	}

	// 检查是否在项目目录内
	if cwd, err := os.Getwd(); err == nil {
		if strings.HasPrefix(absPath, cwd) {
			return true
		}
	}

	return false
}

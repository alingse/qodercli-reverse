// Package prompts 环境信息收集器
package prompts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// EnvironmentCollector 环境信息收集器
type EnvironmentCollector struct {
	cache *EnvironmentInfo
}

// NewEnvironmentCollector 创建新的收集器
func NewEnvironmentCollector() *EnvironmentCollector {
	return &EnvironmentCollector{}
}

// Collect 收集环境信息
func (c *EnvironmentCollector) Collect() (*EnvironmentInfo, error) {
	if c.cache != nil {
		return c.cache, nil
	}

	info := &EnvironmentInfo{}

	// 收集系统信息
	c.collectSystemInfo(info)

	// 收集 Git 信息
	c.collectGitInfo(info)

	// 收集开发环境版本
	c.collectDevEnvironments(info)

	// 时间信息
	info.Timezone = time.Local.String()
	info.CurrentTime = time.Now().Format(time.RFC3339)

	c.cache = info
	return info, nil
}

// collectSystemInfo 收集系统信息
func (c *EnvironmentCollector) collectSystemInfo(info *EnvironmentInfo) {
	info.OS = runtime.GOOS
	info.Architecture = runtime.GOARCH
	info.HomeDir = c.getHomeDir()
	info.WorkingDir = c.getWorkingDir()
	info.TempDir = os.TempDir()
	info.Shell = c.detectShell()
}

// collectGitInfo 收集 Git 信息
func (c *EnvironmentCollector) collectGitInfo(info *EnvironmentInfo) {
	// 检查是否在 Git 仓库中
	if !c.isGitRepo() {
		return
	}

	info.GitRepo = true
	info.GitBranch = c.execGit("rev-parse", "--abbrev-ref", "HEAD")
	info.GitCommit = c.execGit("rev-parse", "--short", "HEAD")
	info.GitRemote = c.execGit("remote", "get-url", "origin")

	// 获取 Git 状态摘要
	status := c.execGit("status", "--porcelain")
	if status != "" {
		lines := strings.Split(status, "\n")
		modified := 0
		untracked := 0
		for _, line := range lines {
			if len(line) > 0 {
				if strings.HasPrefix(line, "??") {
					untracked++
				} else {
					modified++
				}
			}
		}
		if modified > 0 || untracked > 0 {
			info.GitStatus = fmt.Sprintf("%d modified, %d untracked", modified, untracked)
		}
	} else {
		info.GitStatus = "clean"
	}
}

// collectDevEnvironments 收集开发环境版本
func (c *EnvironmentCollector) collectDevEnvironments(info *EnvironmentInfo) {
	info.GoVersion = c.getCommandVersion("go", "version")
	info.NodeVersion = c.getCommandVersion("node", "--version")
	info.PythonVersion = c.getCommandVersion("python", "--version")
	if info.PythonVersion == "" {
		info.PythonVersion = c.getCommandVersion("python3", "--version")
	}
	info.JavaVersion = c.getCommandVersion("java", "-version")
	info.RustVersion = c.getCommandVersion("rustc", "--version")
}

// getHomeDir 获取用户主目录
func (c *EnvironmentCollector) getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// getWorkingDir 获取当前工作目录
func (c *EnvironmentCollector) getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// detectShell 检测当前 shell
func (c *EnvironmentCollector) detectShell() string {
	// 从环境变量检测
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}

	// Windows
	if runtime.GOOS == "windows" {
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		return "cmd"
	}

	return "unknown"
}

// isGitRepo 检查是否在 Git 仓库中
func (c *EnvironmentCollector) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

// execGit 执行 Git 命令
func (c *EnvironmentCollector) execGit(args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getCommandVersion 获取命令版本
func (c *EnvironmentCollector) getCommandVersion(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// 取第一行
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

// ClearCache 清除缓存
func (c *EnvironmentCollector) ClearCache() {
	c.cache = nil
}

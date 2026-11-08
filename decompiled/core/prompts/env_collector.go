// Package prompts Environment info collector
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

// EnvironmentCollector Environment info collector
type EnvironmentCollector struct {
	cache *EnvironmentInfo
}

// NewEnvironmentCollector Creates a new collector
func NewEnvironmentCollector() *EnvironmentCollector {
	return &EnvironmentCollector{}
}

// Collect Collects environment info
func (c *EnvironmentCollector) Collect() (*EnvironmentInfo, error) {
	if c.cache != nil {
		return c.cache, nil
	}

	info := &EnvironmentInfo{}

	// Collect system info
	c.collectSystemInfo(info)

	// Collect Git info
	c.collectGitInfo(info)

	// Collect dev environment versions
	c.collectDevEnvironments(info)

	// Time info
	info.Timezone = time.Local.String()
	info.CurrentTime = time.Now().Format(time.RFC3339)

	c.cache = info
	return info, nil
}

// collectSystemInfo Collects system info
func (c *EnvironmentCollector) collectSystemInfo(info *EnvironmentInfo) {
	info.OS = runtime.GOOS
	info.Architecture = runtime.GOARCH
	info.HomeDir = c.getHomeDir()
	info.WorkingDir = c.getWorkingDir()
	info.TempDir = os.TempDir()
	info.Shell = c.detectShell()
}

// collectGitInfo Collects Git info
func (c *EnvironmentCollector) collectGitInfo(info *EnvironmentInfo) {
	// Check whether in a Git repo
	if !c.isGitRepo() {
		return
	}

	info.GitRepo = true
	info.GitBranch = c.execGit("rev-parse", "--abbrev-ref", "HEAD")
	info.GitCommit = c.execGit("rev-parse", "--short", "HEAD")
	info.GitRemote = c.execGit("remote", "get-url", "origin")

	// Get Git status summary
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

// collectDevEnvironments Collects dev environment versions
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

// getHomeDir Gets the user home directory
func (c *EnvironmentCollector) getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// getWorkingDir Gets the current working directory
func (c *EnvironmentCollector) getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

// detectShell Detects the current shell
func (c *EnvironmentCollector) detectShell() string {
	// Detect from environment variables
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

// isGitRepo Checks whether in a Git repo
func (c *EnvironmentCollector) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

// execGit Executes a Git command
func (c *EnvironmentCollector) execGit(args ...string) string {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getCommandVersion Gets a command version
func (c *EnvironmentCollector) getCommandVersion(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// Take the first line
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

// ClearCache Clears the cache
func (c *EnvironmentCollector) ClearCache() {
	c.cache = nil
}

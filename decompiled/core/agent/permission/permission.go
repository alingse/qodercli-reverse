// Package permission 权限系统实现
// 反编译自 qodercli v0.1.29
package permission

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Mode 权限模式
type Mode string

const (
	ModeAsk         Mode = "ask"
	ModeAutoApprove Mode = "auto-approve"
	ModeAutoDeny    Mode = "auto-deny"
)

// Decision 权限决策
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

// Rule 权限规则
type Rule struct {
	Pattern string   `json:"pattern"`
	Action  Decision `json:"action"`
	Type    string   `json:"type,omitempty"` // file, bash, mcp, web
}

// Request 权限请求
type Request struct {
	ToolName    string            `json:"tool_name"`
	ToolInput   string            `json:"tool_input,omitempty"`
	FilePath    string            `json:"file_path,omitempty"`
	Command     string            `json:"command,omitempty"`
	MCPName     string            `json:"mcp_name,omitempty"`
	MCPMethod   string            `json:"mcp_method,omitempty"`
	URL         string            `json:"url,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Result 权限结果
type Result struct {
	Decision Decision `json:"decision"`
	Reason   string   `json:"reason,omitempty"`
}

// Checker 权限检查器接口
type Checker interface {
	Check(ctx context.Context, req *Request) (*Result, error)
}

// Coordinator 权限协调器
type Coordinator struct {
	mode            Mode
	allowedTools    []string
	disallowedTools []string
	fileMatcher     *FileRuleMatcher
	bashMatcher     *BashRuleMatcher
	mcpMatcher      *MCPRuleMatcher
	webMatcher      *WebFetchRuleMatcher
	askCallback     func(ctx context.Context, req *Request) (Decision, error)
}

// NewCoordinator 创建权限协调器
func NewCoordinator(mode Mode) *Coordinator {
	return &Coordinator{
		mode:        mode,
		fileMatcher: NewFileRuleMatcher(),
		bashMatcher: NewBashRuleMatcher(),
		mcpMatcher:  NewMCPRuleMatcher(),
		webMatcher:  NewWebFetchRuleMatcher(),
	}
}

// SetAllowedTools 设置允许的工具
func (c *Coordinator) SetAllowedTools(tools []string) {
	c.allowedTools = tools
}

// SetDisallowedTools 设置禁止的工具
func (c *Coordinator) SetDisallowedTools(tools []string) {
	c.disallowedTools = tools
}

// SetAskCallback 设置询问回调
func (c *Coordinator) SetAskCallback(callback func(ctx context.Context, req *Request) (Decision, error)) {
	c.askCallback = callback
}

// Check 检查权限
func (c *Coordinator) Check(ctx context.Context, req *Request) (*Result, error) {
	// 1. 检查工具级别权限
	toolDecision := c.checkToolPermission(req.ToolName)
	if toolDecision == DecisionDeny {
		return &Result{
			Decision: DecisionDeny,
			Reason:   fmt.Sprintf("Tool %s is in disallowed list", req.ToolName),
		}, nil
	}
	
	// 2. 根据工具类型检查具体权限
	var decision Decision
	var reason string
	
	switch req.ToolName {
	case "Bash", "BashOutput", "KillBash":
		decision, reason = c.checkBashPermission(req)
	case "Read", "Write", "Edit", "DeleteFile":
		decision, reason = c.checkFilePermission(req)
	case "WebFetch", "WebSearch":
		decision, reason = c.checkWebPermission(req)
	default:
		// 检查是否是 MCP 工具
		if strings.HasPrefix(req.ToolName, "mcp__") {
			decision, reason = c.checkMCPPermission(req)
		} else {
			// 其他工具默认允许
			decision = DecisionAllow
		}
	}
	
	// 3. 应用权限模式
	if decision == DecisionAsk {
		switch c.mode {
		case ModeAutoApprove:
			decision = DecisionAllow
			reason = "auto-approved by mode"
		case ModeAutoDeny:
			decision = DecisionDeny
			reason = "auto-denied by mode"
		case ModeAsk:
			if c.askCallback != nil {
				userDecision, err := c.askCallback(ctx, req)
				if err != nil {
					return nil, err
				}
				decision = userDecision
				reason = "user decision"
			}
		}
	}
	
	return &Result{
		Decision: decision,
		Reason:   reason,
	}, nil
}

// checkToolPermission 检查工具级别权限
func (c *Coordinator) checkToolPermission(toolName string) Decision {
	// 检查禁止列表
	for _, t := range c.disallowedTools {
		if t == toolName {
			return DecisionDeny
		}
	}
	
	// 检查允许列表
	if len(c.allowedTools) > 0 {
		for _, t := range c.allowedTools {
			if t == toolName {
				return DecisionAllow
			}
		}
		return DecisionDeny
	}
	
	return DecisionAllow
}

// checkFilePermission 检查文件权限
func (c *Coordinator) checkFilePermission(req *Request) (Decision, string) {
	d := c.fileMatcher.Match(req.FilePath)
	return d, ""
}

// checkBashPermission 检查 Bash 权限
func (c *Coordinator) checkBashPermission(req *Request) (Decision, string) {
	return c.bashMatcher.Match(req.Command)
}

// checkMCPPermission 检查 MCP 权限
func (c *Coordinator) checkMCPPermission(req *Request) (Decision, string) {
	return c.mcpMatcher.Match(req.MCPName, req.MCPMethod)
}

// checkWebPermission 检查 Web 权限
func (c *Coordinator) checkWebPermission(req *Request) (Decision, string) {
	return c.webMatcher.Match(req.URL)
}

// FileRuleMatcher 文件规则匹配器
type FileRuleMatcher struct {
	rules []Rule
}

// NewFileRuleMatcher 创建文件规则匹配器
func NewFileRuleMatcher() *FileRuleMatcher {
	return &FileRuleMatcher{
		rules: loadDefaultFileRules(),
	}
}

// AddRule 添加规则
func (m *FileRuleMatcher) AddRule(rule Rule) {
	m.rules = append(m.rules, rule)
}

// Match 匹配路径
func (m *FileRuleMatcher) Match(path string) Decision {
	if path == "" {
		return DecisionAllow
	}
	
	// 规范化路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return DecisionAsk
	}
	
	// 检查特殊保护路径
	if isProtectedPath(absPath) {
		return DecisionAsk
	}
	
	// 检查规则
	for _, rule := range m.rules {
		if match, _ := filepath.Match(rule.Pattern, absPath); match {
			return rule.Action
		}
		// 尝试 glob 匹配
		if matched, _ := filepath.Match(rule.Pattern, filepath.Base(absPath)); matched {
			return rule.Action
		}
	}
	
	// 默认策略：项目内允许，项目外询问
	if isInProject(absPath) {
		return DecisionAllow
	}
	
	return DecisionAsk
}

// BashRuleMatcher Bash 规则匹配器
type BashRuleMatcher struct {
	denyPatterns []*regexp.Regexp
}

// NewBashRuleMatcher 创建 Bash 规则匹配器
func NewBashRuleMatcher() *BashRuleMatcher {
	return &BashRuleMatcher{
		denyPatterns: loadDefaultBashDenyPatterns(),
	}
}

// Match 匹配命令
func (m *BashRuleMatcher) Match(command string) (Decision, string) {
	if command == "" {
		return DecisionAllow, ""
	}
	
	// 检查危险命令模式
	for _, pattern := range m.denyPatterns {
		if pattern.MatchString(command) {
			return DecisionDeny, fmt.Sprintf("matches deny pattern: %s", pattern.String())
		}
	}
	
	// 检查 sudo
	if strings.Contains(command, "sudo ") {
		return DecisionAsk, "sudo command requires confirmation"
	}
	
	// 检查 rm -rf
	if strings.Contains(command, "rm -rf") {
		return DecisionAsk, "rm -rf requires confirmation"
	}
	
	return DecisionAllow, ""
}

// MCPRuleMatcher MCP 规则匹配器
type MCPRuleMatcher struct {
	rules []Rule
}

// NewMCPRuleMatcher 创建 MCP 规则匹配器
func NewMCPRuleMatcher() *MCPRuleMatcher {
	return &MCPRuleMatcher{
		rules: loadDefaultMCPRules(),
	}
}

// Match 匹配 MCP 调用
func (m *MCPRuleMatcher) Match(serverName, method string) (Decision, string) {
	fullName := fmt.Sprintf("mcp__%s__%s", serverName, method)
	
	for _, rule := range m.rules {
		if matched, _ := filepath.Match(rule.Pattern, fullName); matched {
			return rule.Action, fmt.Sprintf("matches rule: %s", rule.Pattern)
		}
	}
	
	return DecisionAllow, ""
}

// WebFetchRuleMatcher WebFetch 规则匹配器
type WebFetchRuleMatcher struct {
	allowList []string
	denyList  []string
}

// NewWebFetchRuleMatcher 创建 WebFetch 规则匹配器
func NewWebFetchRuleMatcher() *WebFetchRuleMatcher {
	return &WebFetchRuleMatcher{
		allowList: loadDefaultWebAllowList(),
		denyList:  loadDefaultWebDenyList(),
	}
}

// Match 匹配 URL
func (m *WebFetchRuleMatcher) Match(url string) (Decision, string) {
	if url == "" {
		return DecisionAllow, ""
	}
	
	// 检查拒绝列表
	for _, pattern := range m.denyList {
		if strings.Contains(url, pattern) {
			return DecisionDeny, fmt.Sprintf("URL in deny list: %s", pattern)
		}
	}
	
	// 检查允许列表
	if len(m.allowList) > 0 {
		for _, pattern := range m.allowList {
			if strings.Contains(url, pattern) {
				return DecisionAllow, ""
			}
		}
		return DecisionAsk, "URL not in allow list"
	}
	
	return DecisionAllow, ""
}

// 辅助函数

func isProtectedPath(path string) bool {
	protected := []string{
		".qoder",
		".mcp.json",
		"settings.json",
		"settings.local.json",
	}
	
	for _, p := range protected {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

func isInProject(path string) bool {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	
	// 检查路径是否在工作目录内
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return false
	}
	
	return !strings.HasPrefix(rel, "..")
}

func loadDefaultFileRules() []Rule {
	return []Rule{
		{Pattern: "*.env", Action: DecisionAsk},
		{Pattern: "*.key", Action: DecisionAsk},
		{Pattern: "*.pem", Action: DecisionAsk},
		{Pattern: ".ssh/*", Action: DecisionDeny},
		{Pattern: ".aws/*", Action: DecisionAsk},
	}
}

func loadDefaultBashDenyPatterns() []*regexp.Regexp {
	patterns := []string{
		`rm\s+-rf\s+/`,
		`>\s*/dev/sda`,
		`mkfs\.`,
		`dd\s+if=/dev/zero\s+of=/dev/sda`,
	}
	
	var result []*regexp.Regexp
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil {
			result = append(result, re)
		}
	}
	return result
}

func loadDefaultMCPRules() []Rule {
	return []Rule{
		// 默认允许所有 MCP 工具
	}
}

func loadDefaultWebAllowList() []string {
	return []string{
		"github.com",
		"stackoverflow.com",
		"docs.python.org",
		"pkg.go.dev",
		"developer.mozilla.org",
	}
}

func loadDefaultWebDenyList() []string {
	return []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
		"file://",
	}
}

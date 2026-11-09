// Package permission Permission system implementation
// Decompiled from qodercli-reverse v0.1.29
package permission

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Mode Permission mode
type Mode string

const (
	ModeAsk         Mode = "ask"
	ModeAutoApprove Mode = "auto-approve"
	ModeAutoDeny    Mode = "auto-deny"
)

// Decision Permission decision
type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
	DecisionAsk   Decision = "ask"
)

// Rule Permission rule
type Rule struct {
	Pattern string   `json:"pattern"`
	Action  Decision `json:"action"`
	Type    string   `json:"type,omitempty"` // file, bash, mcp, web
}

// Request Permission request
type Request struct {
	ToolName  string            `json:"tool_name"`
	ToolInput string            `json:"tool_input,omitempty"`
	FilePath  string            `json:"file_path,omitempty"`
	Command   string            `json:"command,omitempty"`
	MCPName   string            `json:"mcp_name,omitempty"`
	MCPMethod string            `json:"mcp_method,omitempty"`
	URL       string            `json:"url,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Result Permission result
type Result struct {
	Decision Decision `json:"decision"`
	Reason   string   `json:"reason,omitempty"`
}

// Checker Permission checker interface
type Checker interface {
	Check(ctx context.Context, req *Request) (*Result, error)
}

// Coordinator Permission coordinator
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

// NewCoordinator Creates a permission coordinator
func NewCoordinator(mode Mode) *Coordinator {
	return &Coordinator{
		mode:        mode,
		fileMatcher: NewFileRuleMatcher(),
		bashMatcher: NewBashRuleMatcher(),
		mcpMatcher:  NewMCPRuleMatcher(),
		webMatcher:  NewWebFetchRuleMatcher(),
	}
}

// SetAllowedTools Sets the allowed tools
func (c *Coordinator) SetAllowedTools(tools []string) {
	c.allowedTools = tools
}

// SetDisallowedTools Sets the disallowed tools
func (c *Coordinator) SetDisallowedTools(tools []string) {
	c.disallowedTools = tools
}

// SetAskCallback Sets the ask callback
func (c *Coordinator) SetAskCallback(callback func(ctx context.Context, req *Request) (Decision, error)) {
	c.askCallback = callback
}

// Check Checks permission
func (c *Coordinator) Check(ctx context.Context, req *Request) (*Result, error) {
	// 1. Check tool-level permission
	toolDecision := c.checkToolPermission(req.ToolName)
	if toolDecision == DecisionDeny {
		return &Result{
			Decision: DecisionDeny,
			Reason:   fmt.Sprintf("Tool %s is in disallowed list", req.ToolName),
		}, nil
	}

	// 2. Check specific permissions based on tool type
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
		// Check whether it is an MCP tool
		if strings.HasPrefix(req.ToolName, "mcp__") {
			decision, reason = c.checkMCPPermission(req)
		} else {
			// Other tools are allowed by default
			decision = DecisionAllow
		}
	}

	// 3. Apply permission mode
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

// checkToolPermission Checks tool-level permission
func (c *Coordinator) checkToolPermission(toolName string) Decision {
	// Check disallowed list
	for _, t := range c.disallowedTools {
		if t == toolName {
			return DecisionDeny
		}
	}

	// Check allowed list
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

// checkFilePermission Checks file permission
func (c *Coordinator) checkFilePermission(req *Request) (Decision, string) {
	d := c.fileMatcher.Match(req.FilePath)
	return d, ""
}

// checkBashPermission Checks Bash permission
func (c *Coordinator) checkBashPermission(req *Request) (Decision, string) {
	return c.bashMatcher.Match(req.Command)
}

// checkMCPPermission Checks MCP permission
func (c *Coordinator) checkMCPPermission(req *Request) (Decision, string) {
	return c.mcpMatcher.Match(req.MCPName, req.MCPMethod)
}

// checkWebPermission Checks Web permission
func (c *Coordinator) checkWebPermission(req *Request) (Decision, string) {
	return c.webMatcher.Match(req.URL)
}

// FileRuleMatcher File rule matcher
type FileRuleMatcher struct {
	rules []Rule
}

// NewFileRuleMatcher Creates a file rule matcher
func NewFileRuleMatcher() *FileRuleMatcher {
	return &FileRuleMatcher{
		rules: loadDefaultFileRules(),
	}
}

// AddRule Adds a rule
func (m *FileRuleMatcher) AddRule(rule Rule) {
	m.rules = append(m.rules, rule)
}

// Match Matches a path
func (m *FileRuleMatcher) Match(path string) Decision {
	if path == "" {
		return DecisionAllow
	}

	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return DecisionAsk
	}

	// Check special protected paths
	if isProtectedPath(absPath) {
		return DecisionAsk
	}

	// Check rules
	for _, rule := range m.rules {
		if match, _ := filepath.Match(rule.Pattern, absPath); match {
			return rule.Action
		}
		// Try glob match
		if matched, _ := filepath.Match(rule.Pattern, filepath.Base(absPath)); matched {
			return rule.Action
		}
	}

	// Default policy: allow inside project, ask outside
	if isInProject(absPath) {
		return DecisionAllow
	}

	return DecisionAsk
}

// BashRuleMatcher Bash rule matcher
type BashRuleMatcher struct {
	denyPatterns []*regexp.Regexp
}

// NewBashRuleMatcher Creates a Bash rule matcher
func NewBashRuleMatcher() *BashRuleMatcher {
	return &BashRuleMatcher{
		denyPatterns: loadDefaultBashDenyPatterns(),
	}
}

// Match Matches a command
func (m *BashRuleMatcher) Match(command string) (Decision, string) {
	if command == "" {
		return DecisionAllow, ""
	}

	// Check dangerous command patterns
	for _, pattern := range m.denyPatterns {
		if pattern.MatchString(command) {
			return DecisionDeny, fmt.Sprintf("matches deny pattern: %s", pattern.String())
		}
	}

	// Check sudo
	if strings.Contains(command, "sudo ") {
		return DecisionAsk, "sudo command requires confirmation"
	}

	// Check rm -rf
	if strings.Contains(command, "rm -rf") {
		return DecisionAsk, "rm -rf requires confirmation"
	}

	return DecisionAllow, ""
}

// MCPRuleMatcher MCP rule matcher
type MCPRuleMatcher struct {
	rules []Rule
}

// NewMCPRuleMatcher Creates an MCP rule matcher
func NewMCPRuleMatcher() *MCPRuleMatcher {
	return &MCPRuleMatcher{
		rules: loadDefaultMCPRules(),
	}
}

// Match Matches an MCP call
func (m *MCPRuleMatcher) Match(serverName, method string) (Decision, string) {
	fullName := fmt.Sprintf("mcp__%s__%s", serverName, method)

	for _, rule := range m.rules {
		if matched, _ := filepath.Match(rule.Pattern, fullName); matched {
			return rule.Action, fmt.Sprintf("matches rule: %s", rule.Pattern)
		}
	}

	return DecisionAllow, ""
}

// WebFetchRuleMatcher WebFetch rule matcher
type WebFetchRuleMatcher struct {
	allowList []string
	denyList  []string
}

// NewWebFetchRuleMatcher Creates a WebFetch rule matcher
func NewWebFetchRuleMatcher() *WebFetchRuleMatcher {
	return &WebFetchRuleMatcher{
		allowList: loadDefaultWebAllowList(),
		denyList:  loadDefaultWebDenyList(),
	}
}

// Match Matches a URL
func (m *WebFetchRuleMatcher) Match(url string) (Decision, string) {
	if url == "" {
		return DecisionAllow, ""
	}

	// Check deny list
	for _, pattern := range m.denyList {
		if strings.Contains(url, pattern) {
			return DecisionDeny, fmt.Sprintf("URL in deny list: %s", pattern)
		}
	}

	// Check allow list
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

// Helper functions

func isProtectedPath(path string) bool {
	protected := []string{
		".qoder",
		".qodercli-reverse",
		".qodercli-reverse",
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
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Check whether the path is within the working directory
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
		// Allow all MCP tools by default
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

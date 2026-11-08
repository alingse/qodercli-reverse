// Package mcp MCP (Model Context Protocol) client implementation
// Reverse-engineered from qodercli-reverse v0.1.29
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ServerConfig MCP server configuration.
type ServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Type    string            `json:"type,omitempty"` // stdio, sse, http
	URL     string            `json:"url,omitempty"`
}

// ServersConfig MCP server configuration list.
type ServersConfig struct {
	Servers map[string]*ServerConfig `json:"mcpServers"`
}

// Client MCP client.
type Client struct {
	config       *ServerConfig
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	requestID    int
	mu           sync.Mutex
	pending      map[int]chan *Response
	closed       bool
	closeChan    chan struct{}
	initDone     bool
	capabilities ServerCapabilities
}

// ServerCapabilities server capabilities.
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// ToolsCapability tool capability.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability resource capability.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability prompt capability.
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Request JSON-RPC request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response JSON-RPC response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error JSON-RPC error.
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Notification JSON-RPC notification.
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewClient creates an MCP client.
func NewClient(config *ServerConfig) (*Client, error) {
	return &Client{
		config:    config,
		pending:   make(map[int]chan *Response),
		closeChan: make(chan struct{}),
	}, nil
}

// Start starts the MCP server.
func (c *Client) Start(ctx context.Context) error {
	if c.config.Type != "" && c.config.Type != "stdio" {
		return fmt.Errorf("unsupported server type: %s", c.config.Type)
	}

	// build command
	cmd := exec.CommandContext(ctx, c.config.Command, c.config.Args...)

	// set working directory
	if c.config.Cwd != "" {
		cmd.Dir = c.config.Cwd
	} else {
		cmd.Dir, _ = os.Getwd()
	}

	// set environment variables
	cmd.Env = os.Environ()
	for k, v := range c.config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// get pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr

	// start message read loop
	go c.readLoop(ctx)

	// initialize
	if err := c.initialize(ctx); err != nil {
		c.Stop()
		return fmt.Errorf("failed to initialize: %w", err)
	}

	return nil
}

// Stop stops the MCP server.
func (c *Client) Stop() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.closeChan)
	c.mu.Unlock()

	// close pipes
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	// terminate process
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}

	return nil
}

// readLoop message read loop.
func (c *Client) readLoop(ctx context.Context) {
	scanner := bufio.NewScanner(c.stdout)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closeChan:
			return
		default:
		}

		if !scanner.Scan() {
			return
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// parse message
		var msg struct {
			JSONRPC string `json:"jsonrpc"`
			ID      int    `json:"id"`
			Method  string `json:"method"`
		}

		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		// handle response
		if msg.ID > 0 && msg.Method == "" {
			var resp Response
			if err := json.Unmarshal([]byte(line), &resp); err == nil {
				c.handleResponse(&resp)
			}
		}

		// handle notification
		if msg.Method != "" && msg.ID == 0 {
			var notif Notification
			if err := json.Unmarshal([]byte(line), &notif); err == nil {
				c.handleNotification(&notif)
			}
		}
	}
}

// handleResponse handles a response.
func (c *Client) handleResponse(resp *Response) {
	c.mu.Lock()
	ch, exists := c.pending[resp.ID]
	if exists {
		delete(c.pending, resp.ID)
	}
	c.mu.Unlock()

	if exists {
		ch <- resp
	}
}

// handleNotification handles a notification.
func (c *Client) handleNotification(notif *Notification) {
	switch notif.Method {
	case "notifications/tools/list_changed":
		// tool list changed
	case "notifications/resources/list_changed":
		// resource list changed
	case "notifications/prompts/list_changed":
		// prompt list changed
	}
}

// call invokes a method.
func (c *Client) call(ctx context.Context, method string, params interface{}) (*Response, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}

	c.requestID++
	id := c.requestID

	// build request
	req := &Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}

	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			c.mu.Unlock()
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		req.Params = data
	}

	// create response channel
	ch := make(chan *Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	// send request
	data, err := json.Marshal(req)
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	if _, err := fmt.Fprintln(c.stdin, string(data)); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("write request: %w", err)
	}

	// wait for response
	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp, nil
	case <-time.After(30 * time.Second):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

// initialize initializes the connection.
func (c *Client) initialize(ctx context.Context) error {
	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{
				"listChanged": true,
			},
			"resources": map[string]bool{
				"subscribe":   true,
				"listChanged": true,
			},
			"prompts": map[string]bool{
				"listChanged": true,
			},
		},
		"clientInfo": map[string]string{
			"name":    "qodercli-reverse",
			"version": "0.1.29",
		},
	}

	resp, err := c.call(ctx, "initialize", params)
	if err != nil {
		return err
	}

	var result struct {
		ProtocolVersion string             `json:"protocolVersion"`
		Capabilities    ServerCapabilities `json:"capabilities"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("unmarshal result: %w", err)
	}

	c.capabilities = result.Capabilities
	c.initDone = true

	// send initialized notification
	notif := &Notification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	data, _ := json.Marshal(notif)
	fmt.Fprintln(c.stdin, string(data))

	return nil
}

// ListTools lists tools.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	resp, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tools []Tool `json:"tools"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	return result.Tools, nil
}

// CallTool calls a tool.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResult, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	resp, err := c.call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}

	var result ToolResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	return &result, nil
}

// ListResources lists resources.
func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	resp, err := c.call(ctx, "resources/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Resources []Resource `json:"resources"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	return result.Resources, nil
}

// ReadResource reads a resource.
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	params := map[string]string{
		"uri": uri,
	}

	resp, err := c.call(ctx, "resources/read", params)
	if err != nil {
		return nil, err
	}

	var result ResourceContent
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	return &result, nil
}

// ListPrompts lists prompts.
func (c *Client) ListPrompts(ctx context.Context) ([]Prompt, error) {
	resp, err := c.call(ctx, "prompts/list", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Prompts []Prompt `json:"prompts"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	return result.Prompts, nil
}

// GetPrompt gets a prompt.
func (c *Client) GetPrompt(ctx context.Context, name string, arguments map[string]string) (*PromptMessage, error) {
	params := map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	}

	resp, err := c.call(ctx, "prompts/get", params)
	if err != nil {
		return nil, err
	}

	var result PromptMessage
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("unmarshal result: %w", err)
	}

	return &result, nil
}

// Tool MCP tool.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult tool result.
type ToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content content.
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Resource MCP resource.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent resource content.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// Prompt MCP prompt.
type Prompt struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Arguments   []Argument `json:"arguments,omitempty"`
}

// Argument argument.
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptMessage prompt message.
type PromptMessage struct {
	Description string    `json:"description,omitempty"`
	Messages    []Message `json:"messages"`
}

// Message message.
type Message struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

// LoadConfig loads MCP config.
func LoadConfig(path string) (*ServersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config ServersConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &config, nil
}

// LoadGlobalConfig loads global config.
func LoadGlobalConfig() (*ServersConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".mcp.json")
	return LoadConfig(configPath)
}

// LoadProjectConfig loads project config.
func LoadProjectConfig() (*ServersConfig, error) {
	configPath := ".mcp.json"
	return LoadConfig(configPath)
}

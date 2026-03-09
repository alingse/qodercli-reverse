// Package mcp MCP (Model Context Protocol) 客户端实现
// 反编译自 qodercli v0.1.29
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

// ServerConfig MCP 服务器配置
type ServerConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Cwd     string            `json:"cwd,omitempty"`
	Type    string            `json:"type,omitempty"` // stdio, sse, http
	URL     string            `json:"url,omitempty"`
}

// ServersConfig MCP 服务器配置列表
type ServersConfig struct {
	Servers map[string]*ServerConfig `json:"mcpServers"`
}

// Client MCP 客户端
type Client struct {
	config     *ServerConfig
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	requestID  int
	mu         sync.Mutex
	pending    map[int]chan *Response
	closed     bool
	closeChan  chan struct{}
	initDone   bool
	capabilities ServerCapabilities
}

// ServerCapabilities 服务器能力
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// ToolsCapability 工具能力
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability 资源能力
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability 提示能力
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Request JSON-RPC 请求
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response JSON-RPC 响应
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error JSON-RPC 错误
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Notification JSON-RPC 通知
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// NewClient 创建 MCP 客户端
func NewClient(config *ServerConfig) (*Client, error) {
	return &Client{
		config:    config,
		pending:   make(map[int]chan *Response),
		closeChan: make(chan struct{}),
	}, nil
}

// Start 启动 MCP 服务器
func (c *Client) Start(ctx context.Context) error {
	if c.config.Type != "" && c.config.Type != "stdio" {
		return fmt.Errorf("unsupported server type: %s", c.config.Type)
	}
	
	// 构建命令
	cmd := exec.CommandContext(ctx, c.config.Command, c.config.Args...)
	
	// 设置工作目录
	if c.config.Cwd != "" {
		cmd.Dir = c.config.Cwd
	} else {
		cmd.Dir, _ = os.Getwd()
	}
	
	// 设置环境变量
	cmd.Env = os.Environ()
	for k, v := range c.config.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	
	// 获取管道
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
	
	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr
	
	// 启动消息读取循环
	go c.readLoop(ctx)
	
	// 初始化
	if err := c.initialize(ctx); err != nil {
		c.Stop()
		return fmt.Errorf("failed to initialize: %w", err)
	}
	
	return nil
}

// Stop 停止 MCP 服务器
func (c *Client) Stop() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.closeChan)
	c.mu.Unlock()
	
	// 关闭管道
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}
	
	// 终止进程
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
	
	return nil
}

// readLoop 消息读取循环
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
		
		// 解析消息
		var msg struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      int             `json:"id"`
			Method  string          `json:"method"`
		}
		
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		
		// 处理响应
		if msg.ID > 0 && msg.Method == "" {
			var resp Response
			if err := json.Unmarshal([]byte(line), &resp); err == nil {
				c.handleResponse(&resp)
			}
		}
		
		// 处理通知
		if msg.Method != "" && msg.ID == 0 {
			var notif Notification
			if err := json.Unmarshal([]byte(line), &notif); err == nil {
				c.handleNotification(&notif)
			}
		}
	}
}

// handleResponse 处理响应
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

// handleNotification 处理通知
func (c *Client) handleNotification(notif *Notification) {
	switch notif.Method {
	case "notifications/tools/list_changed":
		// 工具列表变更
	case "notifications/resources/list_changed":
		// 资源列表变更
	case "notifications/prompts/list_changed":
		// 提示列表变更
	}
}

// call 调用方法
func (c *Client) call(ctx context.Context, method string, params interface{}) (*Response, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("client closed")
	}
	
	c.requestID++
	id := c.requestID
	
	// 构建请求
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
	
	// 创建响应通道
	ch := make(chan *Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()
	
	// 发送请求
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
	
	// 等待响应
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

// initialize 初始化连接
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
			"name":    "qodercli",
			"version": "0.1.29",
		},
	}
	
	resp, err := c.call(ctx, "initialize", params)
	if err != nil {
		return err
	}
	
	var result struct {
		ProtocolVersion string                `json:"protocolVersion"`
		Capabilities    ServerCapabilities    `json:"capabilities"`
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
	
	// 发送 initialized 通知
	notif := &Notification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	data, _ := json.Marshal(notif)
	fmt.Fprintln(c.stdin, string(data))
	
	return nil
}

// ListTools 列出工具
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

// CallTool 调用工具
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

// ListResources 列出资源
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

// ReadResource 读取资源
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

// ListPrompts 列出提示
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

// GetPrompt 获取提示
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

// Tool MCP 工具
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult 工具结果
type ToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content 内容
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Resource MCP 资源
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent 资源内容
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// Prompt MCP 提示
type Prompt struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Arguments   []Argument  `json:"arguments,omitempty"`
}

// Argument 参数
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptMessage 提示消息
type PromptMessage struct {
	Description string    `json:"description,omitempty"`
	Messages    []Message `json:"messages"`
}

// Message 消息
type Message struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

// LoadConfig 加载 MCP 配置
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

// LoadGlobalConfig 加载全局配置
func LoadGlobalConfig() (*ServersConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	configPath := filepath.Join(homeDir, ".mcp.json")
	return LoadConfig(configPath)
}

// LoadProjectConfig 加载项目配置
func LoadProjectConfig() (*ServersConfig, error) {
	configPath := ".mcp.json"
	return LoadConfig(configPath)
}

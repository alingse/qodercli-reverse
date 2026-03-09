// Package provider Qoder Provider 实现
// 反编译自 qodercli v0.1.29
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// QoderClient Qoder 客户端
type QoderClient struct {
	*baseClient
	endpoint    string
	accessToken string
}

// NewQoderClient 创建 Qoder 客户端
func NewQoderClient(opts ...ClientOption) (*QoderClient, error) {
	client := &QoderClient{
		baseClient: &baseClient{
			httpClient: &defaultHTTPClient{},
		},
		endpoint: "https://openapi.qoder.sh",
	}
	
	for _, opt := range opts {
		opt(client)
	}
	
	return client, nil
}

// ClientOption 客户端选项
type ClientOption func(Client)

// WithAPIKey 设置 API Key
func WithAPIKey(apiKey string) ClientOption {
	return func(c Client) {
		switch client := c.(type) {
		case *QoderClient:
			client.apiKey = apiKey
		case *OpenAIClient:
			client.apiKey = apiKey
		case *IdeaLabClient:
			client.apiKey = apiKey
		}
	}
}

// WithBaseURL 设置 Base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c Client) {
		switch client := c.(type) {
		case *QoderClient:
			client.endpoint = baseURL
		case *OpenAIClient:
			client.baseURL = baseURL
		case *IdeaLabClient:
			client.baseURL = baseURL
		}
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) ClientOption {
	return func(c Client) {
		switch client := c.(type) {
		case *QoderClient:
			client.debug = debug
		case *OpenAIClient:
			client.debug = debug
		case *IdeaLabClient:
			client.debug = debug
		}
	}
}

// WithAgentName 设置 Agent 名称
func WithAgentName(name string) ClientOption {
	return func(c Client) {
		switch client := c.(type) {
		case *QoderClient:
			client.agentName = name
		case *OpenAIClient:
			client.agentName = name
		case *IdeaLabClient:
			client.agentName = name
		}
	}
}

// WithSubagent 设置是否为子代理
func WithSubagent(isSubagent bool) ClientOption {
	return func(c Client) {
		switch client := c.(type) {
		case *QoderClient:
			client.isSubagent = isSubagent
		case *OpenAIClient:
			client.isSubagent = isSubagent
		case *IdeaLabClient:
			client.isSubagent = isSubagent
		}
	}
}

// Stream 流式请求
func (c *QoderClient) Stream(ctx context.Context, req *ModelRequest) <-chan Event {
	eventChan := make(chan Event, 100)
	
	go func() {
		defer close(eventChan)
		
		// 构建请求
		body, err := json.Marshal(req)
		if err != nil {
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "internal_error",
					Message: err.Error(),
				},
			}
			return
		}
		
		httpReq := &HTTPRequest{
			Method:  "POST",
			URL:     fmt.Sprintf("%s/v1/chat/completions", c.endpoint),
			Headers: c.buildHeaders(),
			Body:    body,
			Timeout: 5 * time.Minute,
		}
		
		// 发送请求
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "request_error",
					Message: err.Error(),
				},
			}
			return
		}
		
		// 处理 SSE 流
		c.processSSEStream(ctx, bytes.NewReader(resp.Body), eventChan)
	}()
	
	return eventChan
}

// Send 同步请求
func (c *QoderClient) Send(ctx context.Context, req *ModelRequest) (*Response, error) {
	req.Stream = false
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	httpReq := &HTTPRequest{
		Method:  "POST",
		URL:     fmt.Sprintf("%s/v1/chat/completions", c.endpoint),
		Headers: c.buildHeaders(),
		Body:    body,
		Timeout: 5 * time.Minute,
	}
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(resp.Body))
	}
	
	var apiResp qoderAPIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	
	return c.parseResponse(&apiResp), nil
}

// Close 关闭客户端
func (c *QoderClient) Close() error {
	return nil
}

// buildHeaders 构建请求头
func (c *QoderClient) buildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":      "application/json",
		"Accept":            "text/event-stream",
		"Authorization":     fmt.Sprintf("Bearer %s", c.apiKey),
		"X-Request-Id":      generateRequestID(),
		"X-Machine-Id":      getMachineID(),
		"X-Machine-OS":      getMachineOS(),
		"X-Client-Timestamp": time.Now().Format(time.RFC3339),
	}
	
	if c.agentName != "" {
		headers["X-Agent-Name"] = c.agentName
	}
	if c.isSubagent {
		headers["X-Is-Subagent"] = "true"
	}
	
	return headers
}

// processSSEStream 处理 SSE 流
func (c *QoderClient) processSSEStream(ctx context.Context, reader io.Reader, eventChan chan<- Event) {
	decoder := json.NewDecoder(reader)
	
	for {
		select {
		case <-ctx.Done():
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "context_canceled",
					Message: ctx.Err().Error(),
				},
			}
			return
		default:
		}
		
		var sseEvent sseEvent
		if err := decoder.Decode(&sseEvent); err != nil {
			if err == io.EOF {
				return
			}
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "decode_error",
					Message: err.Error(),
				},
			}
			return
		}
		
		event := c.parseSSEEvent(&sseEvent)
		if event != nil {
			eventChan <- *event
		}
	}
}

// parseSSEEvent 解析 SSE 事件
func (c *QoderClient) parseSSEEvent(sse *sseEvent) *Event {
	switch sse.Event {
	case "message_start":
		return &Event{Type: EventTypeMessageStart}
	case "content_block_start":
		return &Event{Type: EventTypeContentBlockStart}
	case "content_block_delta":
		return c.parseContentDelta(sse.Data)
	case "content_block_stop":
		return &Event{Type: EventTypeContentBlockStop}
	case "tool_use_start":
		return c.parseToolUseStart(sse.Data)
	case "tool_use_delta":
		return c.parseToolUseDelta(sse.Data)
	case "tool_use_stop":
		return &Event{Type: EventTypeToolUseStop}
	case "message_stop":
		return &Event{Type: EventTypeMessageStop}
	case "message_delta":
		return c.parseMessageDelta(sse.Data)
	case "error":
		return c.parseError(sse.Data)
	}
	return nil
}

// parseContentDelta 解析内容增量
func (c *QoderClient) parseContentDelta(data json.RawMessage) *Event {
	var delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &delta); err != nil {
		return nil
	}
	
	eventType := EventTypeContentBlockDelta
	if delta.Type == "thinking_delta" {
		eventType = EventTypeThinkingDelta
	}
	
	return &Event{
		Type:    eventType,
		Content: delta.Text,
	}
}

// parseToolUseStart 解析工具使用开始
func (c *QoderClient) parseToolUseStart(data json.RawMessage) *Event {
	var toolUse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &toolUse); err != nil {
		return nil
	}
	return &Event{
		Type: EventTypeToolUseStart,
		ToolUse: &ToolUse{
			ID:   toolUse.ID,
			Name: toolUse.Name,
		},
	}
}

// parseToolUseDelta 解析工具使用增量
func (c *QoderClient) parseToolUseDelta(data json.RawMessage) *Event {
	var delta struct {
		ToolCallID string `json:"tool_call_id"`
		Name       string `json:"name"`
		Arguments  string `json:"arguments"`
	}
	if err := json.Unmarshal(data, &delta); err != nil {
		return nil
	}
	return &Event{
		Type: EventTypeToolUseDelta,
		ToolCall: &ToolCall{
			ID:        delta.ToolCallID,
			Name:      delta.Name,
			Arguments: delta.Arguments,
		},
	}
}

// parseMessageDelta 解析消息增量
func (c *QoderClient) parseMessageDelta(data json.RawMessage) *Event {
	var delta struct {
		FinishReason string       `json:"finish_reason"`
		TokenUsage   *TokenUsage  `json:"usage"`
	}
	if err := json.Unmarshal(data, &delta); err != nil {
		return nil
	}
	return &Event{
		Type:         EventTypeMessageDelta,
		FinishReason: types.FinishReason(delta.FinishReason),
		TokenUsage:   delta.TokenUsage,
	}
}

// parseError 解析错误
func (c *QoderClient) parseError(data json.RawMessage) *Event {
	var errData struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    string `json:"code"`
	}
	if err := json.Unmarshal(data, &errData); err != nil {
		return nil
	}
	return &Event{
		Type: EventTypeError,
		Error: &ErrorData{
			Type:    errData.Type,
			Message: errData.Message,
			Code:    errData.Code,
		},
	}
}

// parseResponse 解析响应
func (c *QoderClient) parseResponse(resp *qoderAPIResponse) *Response {
	result := &Response{
		ID:           resp.ID,
		Model:        resp.Model,
		FinishReason: types.FinishReason(resp.Choices[0].FinishReason),
		TokenUsage: &TokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}
	
	// 解析内容
	for _, content := range resp.Choices[0].Message.Content {
		if content.Type == "text" {
			result.Content += content.Text
		} else if content.Type == "thinking" {
			result.Thinking += content.Thinking
		}
	}
	
	// 解析工具调用
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		result.ToolCalls = make([]types.ToolCall, len(resp.Choices[0].Message.ToolCalls))
		for i, tc := range resp.Choices[0].Message.ToolCalls {
			result.ToolCalls[i] = types.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}
	
	return result
}

// sseEvent SSE 事件
type sseEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// qoderAPIResponse Qoder API 响应
type qoderAPIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      qoderMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// qoderMessage Qoder 消息
type qoderMessage struct {
	Role      string `json:"role"`
	Content   []struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		Thinking string `json:"thinking,omitempty"`
	} `json:"content"`
	ToolCalls []struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls,omitempty"`
}

// defaultHTTPClient 默认 HTTP 客户端
type defaultHTTPClient struct{}

// IdeaLabClient IdeaLab 客户端（stub）
type IdeaLabClient struct {
	*baseClient
}

func (c *IdeaLabClient) Stream(ctx context.Context, req *ModelRequest) <-chan Event { return nil }
func (c *IdeaLabClient) Send(ctx context.Context, req *ModelRequest) (*Response, error) {
	return nil, fmt.Errorf("IdeaLab client not implemented")
}
func (c *IdeaLabClient) Close() error { return nil }

func (c *defaultHTTPClient) Do(req *HTTPRequest) (*HTTPResponse, error) {
	httpReq, err := http.NewRequest(req.Method, req.URL, strings.NewReader(string(req.Body)))
	if err != nil {
		return nil, err
	}
	
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	
	client := &http.Client{Timeout: req.Timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = v[0]
	}
	
	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// getMachineID 获取机器 ID
func getMachineID() string {
	// 实际实现会从 umid 模块获取
	return "unknown"
}

// getMachineOS 获取机器操作系统
func getMachineOS() string {
	return "darwin"
}

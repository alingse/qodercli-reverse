// Package provider Qoder Provider implementation
// Decompiled from qodercli-reverse v0.1.29
package provider

import (
	"bufio"
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

// QoderClient Qoder client
type QoderClient struct {
	*baseClient
	endpoint    string
	accessToken string
}

// NewQoderClient creates the Qoder client
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

// ClientOption client option
type ClientOption func(Client)

// WithAPIKey sets API Key
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

// WithBaseURL sets Base URL
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

// WithDebug sets debug mode
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

// WithAgentName sets agent name
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

// WithSubagent sets whether subagent
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

// Stream streaming request
func (c *QoderClient) Stream(ctx context.Context, req *ModelRequest) <-chan Event {
	eventChan := make(chan Event, 100)

	go func() {
		defer close(eventChan)

		// build request
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

		// send request
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

		// handle SSE stream
		c.processSSEStream(ctx, bytes.NewReader(resp.Body), eventChan)
	}()

	return eventChan
}

// Send sync request
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

// Close closes the client
func (c *QoderClient) Close() error {
	return nil
}

// buildHeaders builds request headers
func (c *QoderClient) buildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":       "application/json",
		"Accept":             "text/event-stream",
		"Authorization":      fmt.Sprintf("Bearer %s", c.apiKey),
		"X-Request-Id":       generateRequestID(),
		"X-Machine-Id":       getMachineID(),
		"X-Machine-OS":       getMachineOS(),
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

// processSSEStream handles SSE stream
func (c *QoderClient) processSSEStream(ctx context.Context, reader io.Reader, eventChan chan<- Event) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
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

		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			break
		}

		var sseEvent sseEvent
		if err := json.Unmarshal([]byte(data), &sseEvent); err != nil {
			continue
		}

		event := c.parseSSEEvent(&sseEvent)
		if event != nil {
			eventChan <- *event
		}
	}
}

// parseSSEEvent parses SSE event
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

// parseContentDelta parses content delta
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

// parseToolUseStart parses tool use start
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

// parseToolUseDelta parses tool use delta
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

// parseMessageDelta parses message delta
func (c *QoderClient) parseMessageDelta(data json.RawMessage) *Event {
	var delta struct {
		FinishReason string      `json:"finish_reason"`
		TokenUsage   *TokenUsage `json:"usage"`
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

// parseError parses error
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

// parseResponse parses response
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

	// parse content
	for _, content := range resp.Choices[0].Message.Content {
		if content.Type == "text" {
			result.Content += content.Text
		} else if content.Type == "thinking" {
			result.Thinking += content.Thinking
		}
	}

	// parse tool calls
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

// sseEvent SSE event
type sseEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// qoderAPIResponse Qoder API response
type qoderAPIResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int          `json:"index"`
		Message      qoderMessage `json:"message"`
		FinishReason string       `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// qoderMessage Qoder message
type qoderMessage struct {
	Role    string `json:"role"`
	Content []struct {
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

// defaultHTTPClient default HTTP client
type defaultHTTPClient struct{}

// IdeaLabClient IdeaLab client (stub)
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

// generateRequestID generates request ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

// getMachineID gets machine ID
func getMachineID() string {
	// actual implementation fetches from umid module
	return "unknown"
}

// getMachineOS gets machine OS
func getMachineOS() string {
	return "darwin"
}

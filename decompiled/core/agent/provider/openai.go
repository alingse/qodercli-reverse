// Package provider OpenAI Provider 实现
// 支持标准的 OPENAI_API_KEY, OPENAI_MODEL, OPENAI_BASE_URL 环境变量
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// OpenAIClient OpenAI 兼容客户端
type OpenAIClient struct {
	*baseClient
	baseURL    string
	model      string
	httpClient HTTPClient
}

// NewOpenAIClient 创建 OpenAI 客户端
func NewOpenAIClient(opts ...ClientOption) (*OpenAIClient, error) {
	client := &OpenAIClient{
		baseClient: &baseClient{
			apiKey:     getEnvOrDefault("OPENAI_API_KEY", ""),
			baseURL:    getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			debug:      false,
			agentName:  "",
			isSubagent: false,
		},
		model:      getEnvOrDefault("OPENAI_MODEL", "gpt-4o"),
		httpClient: &defaultHTTPClient{},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// Stream 流式请求
func (c *OpenAIClient) Stream(ctx context.Context, req *ModelRequest) <-chan Event {
	eventChan := make(chan Event, 100)

	go func() {
		defer close(eventChan)

		log.Debug("Starting OpenAI stream request to %s", c.baseURL)
		log.Debug("Request model: %s, max_tokens: %d", req.Model, req.MaxTokens)

		// 构建 OpenAI 格式的请求
		openAIReq, err := c.buildOpenAIRequest(req)
		if err != nil {
			log.Error("Failed to build OpenAI request: %v", err)
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "internal_error",
					Message: err.Error(),
				},
			}
			return
		}

		body, err := json.Marshal(openAIReq)
		if err != nil {
			log.Error("Failed to marshal request: %v", err)
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "internal_error",
					Message: err.Error(),
				},
			}
			return
		}

		log.Debug("Request body size: %d bytes", len(body))

		httpReq := &HTTPRequest{
			Method:  "POST",
			URL:     fmt.Sprintf("%s/chat/completions", c.baseURL),
			Headers: c.buildHeaders(),
			Body:    body,
			Timeout: 5 * time.Minute,
		}

		// 发送请求
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			log.Error("HTTP request failed: %v", err)
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "request_error",
					Message: err.Error(),
				},
			}
			return
		}

		log.Debug("HTTP response status: %d", resp.StatusCode)

		if resp.StatusCode >= 400 {
			log.Error("API error response: status=%d, body=%s", resp.StatusCode, string(resp.Body))
			eventChan <- Event{
				Type: EventTypeError,
				Error: &ErrorData{
					Type:    "http_error",
					Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(resp.Body)),
				},
			}
			return
		}

		// 处理 SSE 流
		c.processSSEStream(ctx, bytes.NewReader(resp.Body), eventChan)
		log.Debug("Stream processing completed")
	}()

	return eventChan
}

// Send 同步请求
func (c *OpenAIClient) Send(ctx context.Context, req *ModelRequest) (*Response, error) {
	log.Debug("Starting OpenAI sync request to %s", c.baseURL)
	req.Stream = false

	// 构建 OpenAI 格式的请求
	openAIReq, err := c.buildOpenAIRequest(req)
	if err != nil {
		log.Error("Failed to build OpenAI request: %v", err)
		return nil, fmt.Errorf("build request: %w", err)
	}

	body, err := json.Marshal(openAIReq)
	if err != nil {
		log.Error("Failed to marshal request: %v", err)
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	log.Debug("Request body size: %d bytes", len(body))

	httpReq := &HTTPRequest{
		Method:  "POST",
		URL:     fmt.Sprintf("%s/chat/completions", c.baseURL),
		Headers: c.buildHeaders(),
		Body:    body,
		Timeout: 5 * time.Minute,
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Error("HTTP request failed: %v", err)
		return nil, fmt.Errorf("send request: %w", err)
	}

	log.Debug("HTTP response status: %d", resp.StatusCode)

	if resp.StatusCode >= 400 {
		log.Error("API error response: status=%d, body=%s", resp.StatusCode, string(resp.Body))
		return nil, fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(resp.Body))
	}

	var apiResp openAIChatResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		log.Error("Failed to unmarshal response: %v", err)
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	log.Debug("Response: model=%s, tokens=%d+%d", apiResp.Model, apiResp.Usage.InputTokens, apiResp.Usage.OutputTokens)

	return c.parseResponse(&apiResp), nil
}

// Close 关闭客户端
func (c *OpenAIClient) Close() error {
	return nil
}

// buildOpenAIRequest 构建 OpenAI 格式的请求
func (c *OpenAIClient) buildOpenAIRequest(req *ModelRequest) (*openAIChatRequest, error) {
	// 转换消息格式
	messages := make([]openAIMessage, 0, len(req.Messages)+1)

	// 添加 system prompt
	if req.SystemPrompt != "" {
		messages = append(messages, openAIMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// 转换用户消息
	for _, msg := range req.Messages {
		openAIMsg := openAIMessage{
			Role: string(msg.Role),
		}

		// 处理内容 - 将 ContentPart 数组转换为文本或混合内容
		if len(msg.Content) > 0 {
			textParts := make([]string, 0)
			hasImages := false

			for _, part := range msg.Content {
				if part.Type == "text" || part.Type == "thinking" {
					textParts = append(textParts, part.Text)
				} else if part.Type == "image" && part.ImageSource != nil {
					hasImages = true
				}
			}

			fullText := strings.Join(textParts, "\n")

			if hasImages && len(msg.Content) > 0 {
				// 多模态内容
				contentParts := make([]interface{}, 0)
				for _, part := range msg.Content {
					if part.Type == "text" && part.Text != "" {
						contentParts = append(contentParts, map[string]string{
							"type": "text",
							"text": part.Text,
						})
					} else if part.Type == "image" && part.ImageSource != nil {
						imageURL := part.ImageSource.Data
						if part.ImageSource.Type == "base64" {
							imageURL = fmt.Sprintf("data:%s;base64,%s", part.ImageSource.MediaType, part.ImageSource.Data)
						}
						contentParts = append(contentParts, map[string]interface{}{
							"type": "image_url",
							"image_url": map[string]string{
								"url": imageURL,
							},
						})
					}
				}
				openAIMsg.Content = contentParts
			} else {
				// 纯文本内容
				openAIMsg.Content = fullText
			}
		}

		// 处理工具调用
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]openAIToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				toolCalls[i] = openAIToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: openAIFunction{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
			openAIMsg.ToolCalls = toolCalls
		}

		// 处理工具调用结果
		if msg.ToolCallID != "" {
			openAIMsg.ToolCallID = msg.ToolCallID
			openAIMsg.Name = msg.Name
		}

		messages = append(messages, openAIMsg)
	}

	// 转换工具定义
	var tools []openAITool
	if len(req.Tools) > 0 {
		tools = make([]openAITool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = openAITool{
				Type: "function",
				Function: openAIFunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.InputSchema,
				},
			}
		}
	}

	// 设置温度
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	// 设置最大 tokens
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// 使用配置中的模型（优先）或请求中的模型
	model := c.model
	if req.Model != "" {
		model = req.Model
	}

	return &openAIChatRequest{
		Model:           model,
		Messages:        messages,
		Tools:           tools,
		MaxTokens:       maxTokens,
		Temperature:     temperature,
		TopP:            req.TopP,
		Stream:          req.Stream,
		Stop:            req.StopSequences,
		ReasoningEffort: string(req.ReasoningEffort),
	}, nil
}

// buildHeaders 构建请求头
func (c *OpenAIClient) buildHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.apiKey),
		"X-Request-ID":  generateRequestID(),
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
func (c *OpenAIClient) processSSEStream(ctx context.Context, reader io.Reader, eventChan chan<- Event) {
	scanner := bufio.NewScanner(reader)

	// 发送消息开始事件
	eventChan <- Event{Type: EventTypeMessageStart}
	eventChan <- Event{Type: EventTypeContentBlockStart}

	var currentContent strings.Builder
	var currentToolCall *openAIToolCall
	lastToolCallIndex := -1 // 初始化为 -1，这样第一个工具调用（index=0）也能被正确处理

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

		var chunk openAIChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// 处理内容增量
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta

			// Content 可能是 string 或 []interface{}
			if contentStr, ok := delta.Content.(string); ok && contentStr != "" {
				currentContent.WriteString(contentStr)
				eventChan <- Event{
					Type:    EventTypeContentBlockDelta,
					Content: contentStr,
				}
			} else if contentArr, ok := delta.Content.([]interface{}); ok {
				// 处理数组格式的内容
				for _, item := range contentArr {
					if m, ok := item.(map[string]interface{}); ok {
						if text, ok := m["text"].(string); ok && text != "" {
							currentContent.WriteString(text)
							eventChan <- Event{
								Type:    EventTypeContentBlockDelta,
								Content: text,
							}
						}
					}
				}
			}

			// 处理思考内容（某些模型如 o1）
			if delta.ReasoningContent != "" {
				eventChan <- Event{
					Type:     EventTypeThinkingDelta,
					Thinking: delta.ReasoningContent,
				}
			}
		}

		// 处理工具调用
		if len(chunk.Choices) > 0 && len(chunk.Choices[0].Delta.ToolCalls) > 0 {
			for _, tc := range chunk.Choices[0].Delta.ToolCalls {
				log.Debug("Processing tool call: index=%d, name=%s, args=%s", tc.Index, tc.Function.Name, tc.Function.Arguments)

				if tc.Index != lastToolCallIndex {
					// 新的工具调用 - 先发送前一个工具调用的停止事件
					if currentToolCall != nil {
						eventChan <- Event{
							Type: EventTypeToolUseStop,
						}
						log.Debug("Sent ToolUseStop for previous tool call")
					}

					// 更新索引
					lastToolCallIndex = tc.Index

					if tc.Function.Name != "" {
						currentToolCall = &openAIToolCall{
							ID:   tc.ID,
							Type: "function",
							Function: openAIFunction{
								Name: tc.Function.Name,
							},
						}
						eventChan <- Event{
							Type: EventTypeToolUseStart,
							ToolUse: &ToolUse{
								ID:   tc.ID,
								Name: tc.Function.Name,
							},
						}
						log.Debug("Sent ToolUseStart for tool: %s", tc.Function.Name)
					}
				}

				if currentToolCall != nil {
					if tc.Function.Arguments != "" {
						currentToolCall.Function.Arguments += tc.Function.Arguments
					}

					eventChan <- Event{
						Type: EventTypeToolUseDelta,
						ToolCall: &ToolCall{
							ID:        tc.ID,
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
					log.Debug("Sent ToolUseDelta for tool: %s, args len: %d", tc.Function.Name, len(tc.Function.Arguments))
				}
			}
		}

		// 处理完成原因
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			finishReason := types.FinishReason(chunk.Choices[0].FinishReason)
			log.Debug("Finish reason: %s", finishReason)

			// 如果有未完成的工具调用，发送停止事件
			if currentToolCall != nil {
				eventChan <- Event{
					Type: EventTypeToolUseStop,
				}
				log.Debug("Sent final ToolUseStop")
				currentToolCall = nil
			}

			// 发送 token 使用信息
			if chunk.Usage != nil {
				eventChan <- Event{
					Type: EventTypeMessageDelta,
					TokenUsage: &TokenUsage{
						InputTokens:  chunk.Usage.InputTokens,
						OutputTokens: chunk.Usage.OutputTokens,
						TotalTokens:  chunk.Usage.TotalTokens,
					},
					FinishReason: finishReason,
				}
			} else {
				eventChan <- Event{
					Type:         EventTypeMessageDelta,
					FinishReason: finishReason,
				}
			}
		}
	}

	// 发送停止事件
	eventChan <- Event{Type: EventTypeContentBlockStop}
	eventChan <- Event{Type: EventTypeMessageStop}
}

// parseResponse 解析响应
func (c *OpenAIClient) parseResponse(resp *openAIChatResponse) *Response {
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
	choice := resp.Choices[0]

	// Content 可能是 string 或 []interface{}
	if contentStr, ok := choice.Message.Content.(string); ok {
		result.Content = contentStr
	} else if contentArr, ok := choice.Message.Content.([]interface{}); ok {
		// 处理数组格式的内容
		for _, item := range contentArr {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					result.Content += text
				}
			}
		}
	}

	// 添加思考内容
	if choice.Message.ReasoningContent != "" {
		result.Thinking = choice.Message.ReasoningContent
	}

	// 解析工具调用
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]types.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = types.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			}
		}
	}

	return result
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// OpenAI API 数据结构

// openAIChatRequest OpenAI 聊天请求
type openAIChatRequest struct {
	Model           string          `json:"model"`
	Messages        []openAIMessage `json:"messages"`
	Tools           []openAITool    `json:"tools,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
	Temperature     float64         `json:"temperature,omitempty"`
	TopP            float64         `json:"top_p,omitempty"`
	Stream          bool            `json:"stream"`
	Stop            []string        `json:"stop,omitempty"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
}

// openAIMessage OpenAI 消息
type openAIMessage struct {
	Role             string           `json:"role"`
	Content          interface{}      `json:"content,omitempty"` // string 或 []interface{}
	ToolCalls        []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string           `json:"tool_call_id,omitempty"`
	Name             string           `json:"name,omitempty"`
	ReasoningContent string           `json:"reasoning_content,omitempty"` // 思考内容（某些模型）
}

// openAITool OpenAI 工具
type openAITool struct {
	Type     string                   `json:"type"`
	Function openAIFunctionDefinition `json:"function"`
}

// openAIToolCall OpenAI 工具调用
type openAIToolCall struct {
	Index    int            `json:"index,omitempty"`
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

// openAIFunction OpenAI 函数
type openAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

// openAIFunctionDefinition OpenAI 函数定义
type openAIFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// openAIChatResponse OpenAI 聊天响应
type openAIChatResponse struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

// openAIChoice OpenAI 选择
type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// openAIChunk OpenAI 流式响应块
type openAIChunk struct {
	ID      string              `json:"id"`
	Model   string              `json:"model"`
	Choices []openAIChunkChoice `json:"choices"`
	Usage   *openAIUsage        `json:"usage,omitempty"`
}

// openAIChunkChoice OpenAI 流式选择
type openAIChunkChoice struct {
	Index        int           `json:"index"`
	Delta        openAIMessage `json:"delta"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

// openAIUsage OpenAI 使用情况
type openAIUsage struct {
	InputTokens  int `json:"prompt_tokens"`
	OutputTokens int `json:"completion_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

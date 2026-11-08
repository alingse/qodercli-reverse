// Package provider OpenAI Provider implementation
// Supports standard OPENAI_API_KEY, OPENAI_MODEL, OPENAI_BASE_URL env vars
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

// OpenAIClient OpenAI compatible client
type OpenAIClient struct {
	*baseClient
	baseURL    string
	model      string
	httpClient HTTPClient
}

// NewOpenAIClient creates OpenAI client
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

// Stream streaming request
func (c *OpenAIClient) Stream(ctx context.Context, req *ModelRequest) <-chan Event {
	eventChan := make(chan Event, 100)

	go func() {
		defer close(eventChan)

		log.Debug("Starting OpenAI stream request to %s", c.baseURL)
		log.Debug("Request model: %s, max_tokens: %d", req.Model, req.MaxTokens)

		// build OpenAI-format request
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

		// send request
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

		// handle SSE stream
		c.processSSEStream(ctx, bytes.NewReader(resp.Body), eventChan)
		log.Debug("Stream processing completed")
	}()

	return eventChan
}

// Send sync request
func (c *OpenAIClient) Send(ctx context.Context, req *ModelRequest) (*Response, error) {
	log.Debug("Starting OpenAI sync request to %s", c.baseURL)
	req.Stream = false

	// build OpenAI-format request
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

// Close closes the client
func (c *OpenAIClient) Close() error {
	return nil
}

// buildOpenAIRequest builds OpenAI-format request
func (c *OpenAIClient) buildOpenAIRequest(req *ModelRequest) (*openAIChatRequest, error) {
	// convert message format
	messages := make([]openAIMessage, 0, len(req.Messages)+1)

	// add system prompt
	if req.SystemPrompt != "" {
		messages = append(messages, openAIMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	// convert user messages
	for _, msg := range req.Messages {
		openAIMsg := openAIMessage{
			Role: string(msg.Role),
		}

		// handle content - convert ContentPart array to text or mixed content
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
				// multimodal content
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
				// plain text content
				openAIMsg.Content = fullText
			}
		}

		// handle tool calls
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

		// handle tool call results
		if msg.ToolCallID != "" {
			openAIMsg.ToolCallID = msg.ToolCallID
			openAIMsg.Name = msg.Name
		}

		messages = append(messages, openAIMsg)
	}

	// convert tool definitions
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

	// set temperature
	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.7
	}

	// set max tokens
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// use configured model (priority) or model from request
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

// buildHeaders builds request headers
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

// processSSEStream handles SSE stream
func (c *OpenAIClient) processSSEStream(ctx context.Context, reader io.Reader, eventChan chan<- Event) {
	scanner := bufio.NewScanner(reader)

	// send message start event
	eventChan <- Event{Type: EventTypeMessageStart}
	eventChan <- Event{Type: EventTypeContentBlockStart}

	var currentContent strings.Builder
	var currentToolCall *openAIToolCall
	lastToolCallIndex := -1 // initialized to -1 so first tool call (index=0) is handled correctly

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

		// handle content delta
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta

			// Content may be string or []interface{}
			if contentStr, ok := delta.Content.(string); ok && contentStr != "" {
				currentContent.WriteString(contentStr)
				eventChan <- Event{
					Type:    EventTypeContentBlockDelta,
					Content: contentStr,
				}
			} else if contentArr, ok := delta.Content.([]interface{}); ok {
				// handle array-format content
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

			// handle thinking content (some models like o1)
			if delta.ReasoningContent != "" {
				eventChan <- Event{
					Type:     EventTypeThinkingDelta,
					Thinking: delta.ReasoningContent,
				}
			}
		}

		// handle tool calls
		if len(chunk.Choices) > 0 && len(chunk.Choices[0].Delta.ToolCalls) > 0 {
			for _, tc := range chunk.Choices[0].Delta.ToolCalls {
				log.Debug("Processing tool call: index=%d, name=%s, args=%s", tc.Index, tc.Function.Name, tc.Function.Arguments)

				if tc.Index != lastToolCallIndex {
					// new tool call - send stop event for previous tool call first
					if currentToolCall != nil {
						eventChan <- Event{
							Type: EventTypeToolUseStop,
						}
						log.Debug("Sent ToolUseStop for previous tool call")
					}

					// update index
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

		// handle finish reason
		if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
			finishReason := types.FinishReason(chunk.Choices[0].FinishReason)
			log.Debug("Finish reason: %s", finishReason)

			// if there are unfinished tool calls, send stop event
			if currentToolCall != nil {
				eventChan <- Event{
					Type: EventTypeToolUseStop,
				}
				log.Debug("Sent final ToolUseStop")
				currentToolCall = nil
			}

			// send token usage info
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

	// send stop event
	eventChan <- Event{Type: EventTypeContentBlockStop}
	eventChan <- Event{Type: EventTypeMessageStop}
}

// parseResponse parses response
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

	// parse content
	choice := resp.Choices[0]

	// Content may be string or []interface{}
	if contentStr, ok := choice.Message.Content.(string); ok {
		result.Content = contentStr
	} else if contentArr, ok := choice.Message.Content.([]interface{}); ok {
		// handle array-format content
		for _, item := range contentArr {
			if m, ok := item.(map[string]interface{}); ok {
				if text, ok := m["text"].(string); ok {
					result.Content += text
				}
			}
		}
	}

	// add thinking content
	if choice.Message.ReasoningContent != "" {
		result.Thinking = choice.Message.ReasoningContent
	}

	// parse tool calls
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

// getEnvOrDefault gets env var or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

// OpenAI API data structures

// openAIChatRequest OpenAI chat request
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

// openAIMessage OpenAI message
type openAIMessage struct {
	Role             string           `json:"role"`
	Content          interface{}      `json:"content,omitempty"` // string or []interface{}
	ToolCalls        []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string           `json:"tool_call_id,omitempty"`
	Name             string           `json:"name,omitempty"`
	ReasoningContent string           `json:"reasoning_content,omitempty"` // thinking content (some models)
}

// openAITool OpenAI tool
type openAITool struct {
	Type     string                   `json:"type"`
	Function openAIFunctionDefinition `json:"function"`
}

// openAIToolCall OpenAI tool call
type openAIToolCall struct {
	Index    int            `json:"index,omitempty"`
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

// openAIFunction OpenAI function
type openAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

// openAIFunctionDefinition OpenAI function definition
type openAIFunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// openAIChatResponse OpenAI chat response
type openAIChatResponse struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

// openAIChoice OpenAI choice
type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// openAIChunk OpenAI streaming chunk
type openAIChunk struct {
	ID      string              `json:"id"`
	Model   string              `json:"model"`
	Choices []openAIChunkChoice `json:"choices"`
	Usage   *openAIUsage        `json:"usage,omitempty"`
}

// openAIChunkChoice OpenAI streaming choice
type openAIChunkChoice struct {
	Index        int           `json:"index"`
	Delta        openAIMessage `json:"delta"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

// openAIUsage OpenAI usage
type openAIUsage struct {
	InputTokens  int `json:"prompt_tokens"`
	OutputTokens int `json:"completion_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

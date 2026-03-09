// Package compact 提供上下文压缩功能
package compact

import (
	"context"
	"fmt"
	"strings"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// SummaryCompactor 摘要压缩器 - 使用 LLM 生成摘要
type SummaryCompactor struct {
	provider  provider.Client
	tokenizer Tokenizer
}

// Compact 执行摘要压缩
func (c *SummaryCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategySummary,
		}, nil
	}

	// 分离系统消息
	var systemMessages []types.Message
	var conversationMessages []types.Message

	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			conversationMessages = append(conversationMessages, msg)
		}
	}

	// 保留最后 N 轮对话
	recentTurns := options.KeepRecentTurns
	if recentTurns <= 0 {
		recentTurns = 5
	}

	// 计算需要保留的消息数（每轮对话包含用户和助手消息）
	keepCount := recentTurns * 2
	if keepCount > len(conversationMessages) {
		keepCount = len(conversationMessages)
	}

	// 提取需要压缩的消息
	messagesToSummarize := conversationMessages[:len(conversationMessages)-keepCount]
	if len(messagesToSummarize) == 0 {
		// 无需压缩
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategySummary,
		}, nil
	}

	// 构建摘要请求
	summaryText, err := c.generateSummary(ctx, messagesToSummarize, options)
	if err != nil {
		log.Error("Failed to generate summary: %v", err)
		// 降级为简单摘要
		summaryText = c.simpleSummary(messagesToSummarize)
	}

	// 构建摘要消息
	summaryMsg := types.Message{
		Role:    types.RoleSystem,
		Content: []types.ContentPart{{Type: "text", Text: summaryText}},
	}

	// 构建压缩后的消息
	var compressed []types.Message

	// 添加原始系统消息
	if options.KeepSystemMessages {
		compressed = append(compressed, systemMessages...)
	}

	// 添加摘要消息
	compressed = append(compressed, summaryMsg)

	// 添加保留的最近对话
	compressed = append(compressed, conversationMessages[len(conversationMessages)-keepCount:]...)

	originalTokens := c.estimateTokens(messages)
	compressedTokens := c.estimateTokens(compressed)
	savedTokens := originalTokens - compressedTokens

	ratio := 0.0
	if originalTokens > 0 {
		ratio = float64(savedTokens) / float64(originalTokens) * 100
	}

	return &CompactResult{
		OriginalMessageCount:   len(messages),
		CompressedMessageCount: len(compressed),
		OriginalTokens:         originalTokens,
		CompressedTokens:       compressedTokens,
		SavedTokens:            savedTokens,
		CompressionRatio:       ratio,
		Strategy:               StrategySummary,
		Summary:                summaryText,
	}, nil
}

// generateSummary 使用 LLM 生成摘要
func (c *SummaryCompactor) generateSummary(ctx context.Context, messages []types.Message, options *CompactOptions) (string, error) {
	if c.provider == nil {
		return "", fmt.Errorf("no provider available")
	}

	// 构建摘要请求的消息
	var builder strings.Builder
	builder.WriteString("请总结以下对话的关键信息，包括：\n")
	builder.WriteString("1. 用户的主要需求和问题\n")
	builder.WriteString("2. 已解决的关键点\n")
	builder.WriteString("3. 重要的上下文信息\n")
	builder.WriteString("4. 待完成的任务\n\n")
	builder.WriteString("对话内容：\n")

	for i, msg := range messages {
		switch msg.Role {
		case types.RoleUser:
			builder.WriteString(fmt.Sprintf("[用户 %d]: ", i+1))
			for _, part := range msg.Content {
				if part.Type == "text" {
					builder.WriteString(part.Text)
				}
			}
			builder.WriteString("\n")
		case types.RoleAssistant:
			builder.WriteString(fmt.Sprintf("[助手 %d]: ", i+1))
			for _, part := range msg.Content {
				if part.Type == "text" && part.Thinking == "" {
					builder.WriteString(part.Text)
				}
			}
			builder.WriteString("\n")
		}
	}

	// 使用自定义提示词（如果有）
	prompt := builder.String()
	if options.CustomPrompt != "" {
		prompt = options.CustomPrompt + "\n\n对话内容：\n" + prompt
	}

	// 构建请求
	req := &provider.ModelRequest{
		Model:      "default",
		SystemPrompt: "你是一个专业的对话摘要助手。请用简洁的语言总结对话内容，保留关键信息。",
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: []types.ContentPart{{Type: "text", Text: prompt}},
			},
		},
		MaxTokens:   2000,
		Temperature: 0.3,
		Stream:      false,
	}

	// 发送请求（非流式）
	resp, err := c.provider.Send(ctx, req)
	if err != nil {
		return "", err
	}

	// 提取响应
	var summary strings.Builder
	for _, part := range resp.Choices[0].Message.Content {
		if part.Type == "text" {
			summary.WriteString(part.Text)
		}
	}

	return summary.String(), nil
}

// simpleSummary 简单的摘要生成（降级方案）
func (c *SummaryCompactor) simpleSummary(messages []types.Message) string {
	var builder strings.Builder
	builder.WriteString("Previous conversation summary:\n\n")

	userTopics := make(map[string]int)
	
	for _, msg := range messages {
		if msg.Role == types.RoleUser {
			for _, part := range msg.Content {
				if part.Type == "text" {
					// 提取前 50 个字符作为主题
					text := part.Text
					if len(text) > 50 {
						text = text[:50] + "..."
					}
					userTopics[text]++
				}
			}
		}
	}

	count := 0
	for topic := range userTopics {
		if count >= 10 {
			break
		}
		builder.WriteString(fmt.Sprintf("- User mentioned: %s\n", topic))
		count++
	}

	if len(userTopics) > 10 {
		builder.WriteString(fmt.Sprintf("\n... and %d more topics", len(userTopics)-10))
	}

	return builder.String()
}

// estimateTokens 估算 Token 数
func (c *SummaryCompactor) estimateTokens(messages []types.Message) int {
	if c.tokenizer != nil {
		total := 0
		for _, msg := range messages {
			for _, part := range msg.Content {
				total += c.tokenizer.Count(part.Text)
				total += c.tokenizer.Count(part.Thinking)
			}
		}
		return total
	}
	return estimateTokensSimple(messages)
}

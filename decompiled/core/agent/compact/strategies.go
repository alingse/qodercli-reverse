// Package compact provides context compaction - compaction strategy implementations
package compact

import (
	"context"
	"fmt"
	"strings"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// SummaryCompactor summary compactor - uses LLM to generate a summary
type SummaryCompactor struct {
	provider  provider.Client
	tokenizer Tokenizer
}

// Compact performs summary compaction
func (c *SummaryCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategySummary,
		}, nil
	}

	// separate system messages
	var systemMessages []types.Message
	var conversationMessages []types.Message

	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			conversationMessages = append(conversationMessages, msg)
		}
	}

	// keep the last N conversation turns
	recentTurns := options.KeepRecentTurns
	if recentTurns <= 0 {
		recentTurns = 5
	}

	// calculate number of messages to keep (each turn includes user and assistant messages)
	keepCount := recentTurns * 2
	if keepCount > len(conversationMessages) {
		keepCount = len(conversationMessages)
	}

	// extract messages to compress
	messagesToSummarize := conversationMessages[:len(conversationMessages)-keepCount]
	if len(messagesToSummarize) == 0 {
		// no compaction needed
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategySummary,
		}, nil
	}

	// build summary request
	summaryText, err := c.generateSummary(ctx, messagesToSummarize, options)
	if err != nil {
		log.Error("Failed to generate summary: %v", err)
		// fall back to simple summary
		summaryText = c.simpleSummary(messagesToSummarize)
	}

	// build summary message
	summaryMsg := types.Message{
		Role:    types.RoleSystem,
		Content: []types.ContentPart{{Type: "text", Text: summaryText}},
	}

	// build compressed messages
	var compressed []types.Message

	// add original system messages
	if options.KeepSystemMessages {
		compressed = append(compressed, systemMessages...)
	}

	// add summary message
	compressed = append(compressed, summaryMsg)

	// add kept recent conversation
	compressed = append(compressed, conversationMessages[len(conversationMessages)-keepCount:]...)

	originalTokens := c.EstimateTokens(messages)
	compressedTokens := c.EstimateTokens(compressed)
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

// generateSummary uses LLM to generate a summary
func (c *SummaryCompactor) generateSummary(ctx context.Context, messages []types.Message, options *CompactOptions) (string, error) {
	if c.provider == nil {
		return "", fmt.Errorf("no provider available")
	}

	// build summary request messages
	var builder strings.Builder
	builder.WriteString("Please summarize the key information of the following conversation, including:\n")
	builder.WriteString("1. The user's main needs and questions\n")
	builder.WriteString("2. Key points already resolved\n")
	builder.WriteString("3. Important context information\n")
	builder.WriteString("4. Tasks yet to be completed\n\n")
	builder.WriteString("Conversation content:\n")

	for i, msg := range messages {
		switch msg.Role {
		case types.RoleUser:
			builder.WriteString(fmt.Sprintf("[User %d]: ", i+1))
			for _, part := range msg.Content {
				if part.Type == "text" {
					builder.WriteString(part.Text)
				}
			}
			builder.WriteString("\n")
		case types.RoleAssistant:
			builder.WriteString(fmt.Sprintf("[Assistant %d]: ", i+1))
			for _, part := range msg.Content {
				if part.Type == "text" && part.Thinking == "" {
					builder.WriteString(part.Text)
				}
			}
			builder.WriteString("\n")
		}
	}

	// use custom prompt if provided
	prompt := builder.String()
	if options.CustomPrompt != "" {
		prompt = options.CustomPrompt + "\n\nConversation content:\n" + prompt
	}

	// build request
	req := &provider.ModelRequest{
		Model:        "default",
		SystemPrompt: "You are a professional conversation summary assistant. Please summarize the conversation content concisely while retaining key information.",
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

	// send request (non-streaming)
	resp, err := c.provider.Send(ctx, req)
	if err != nil {
		return "", err
	}

	// extract response (simplified handling)
	return resp.Content, nil
}

// simpleSummary simple summary generation (fallback)
func (c *SummaryCompactor) simpleSummary(messages []types.Message) string {
	var builder strings.Builder
	builder.WriteString("Previous conversation summary:\n\n")

	userTopics := make(map[string]int)

	for _, msg := range messages {
		if msg.Role == types.RoleUser {
			for _, part := range msg.Content {
				if part.Type == "text" {
					// extract first 50 characters as topic
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

// EstimateTokens estimates token count
func (c *SummaryCompactor) EstimateTokens(messages []types.Message) int {
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

// TruncateCompactor truncation compactor - keeps the last N messages
type TruncateCompactor struct {
	tokenizer Tokenizer
}

// Compact performs truncation compaction
func (c *TruncateCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTruncate,
		}, nil
	}

	// separate system messages
	var systemMessages []types.Message
	var conversationMessages []types.Message

	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			conversationMessages = append(conversationMessages, msg)
		}
	}

	// calculate number of messages to keep
	keepCount := options.KeepRecentTurns * 2
	if keepCount <= 0 {
		keepCount = 10 // default keep 5 conversation turns
	}

	if keepCount >= len(conversationMessages) {
		// no compaction needed
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTruncate,
		}, nil
	}

	// build compressed messages
	var compressed []types.Message

	// add system messages
	if options.KeepSystemMessages {
		compressed = append(compressed, systemMessages...)
	}

	// add recent conversation
	startIndex := len(conversationMessages) - keepCount
	compressed = append(compressed, conversationMessages[startIndex:]...)

	originalTokens := c.EstimateTokens(messages)
	compressedTokens := c.EstimateTokens(compressed)
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
		Strategy:               StrategyTruncate,
	}, nil
}

// EstimateTokens estimates token count
func (c *TruncateCompactor) EstimateTokens(messages []types.Message) int {
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

// SelectiveCompactor selective compactor - selects messages based on importance
type SelectiveCompactor struct {
	tokenizer Tokenizer
}

// messageImportance message importance score
type messageImportance struct {
	message    types.Message
	importance int
	index      int
}

// Compact performs selective compaction
func (c *SelectiveCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategySelective,
		}, nil
	}

	// evaluate importance of each message
	var scored []messageImportance
	for i, msg := range messages {
		score := c.scoreImportance(msg)
		scored = append(scored, messageImportance{
			message:    msg,
			importance: score,
			index:      i,
		})
	}

	// separate system messages (always kept)
	var systemMessages []messageImportance
	var conversationScored []messageImportance

	for _, s := range scored {
		if s.message.Role == types.RoleSystem {
			systemMessages = append(systemMessages, s)
		} else {
			conversationScored = append(conversationScored, s)
		}
	}

	// sort by importance
	sortByImportance(conversationScored)

	// calculate number of messages to keep
	targetTokens := options.TargetTokens
	if targetTokens <= 0 {
		targetTokens = c.EstimateTokens(messages) / 2
	}

	// accumulate tokens until target reached
	var selected []messageImportance
	currentTokens := 0

	// first keep the most important messages
	for _, s := range conversationScored {
		msgTokens := c.EstimateTokens([]types.Message{s.message})
		if currentTokens+msgTokens <= targetTokens || len(selected) < 3 {
			selected = append(selected, s)
			currentTokens += msgTokens
		}
	}

	// re-sort by original order
	sortByIndex(selected)

	// build compressed messages
	var compressed []types.Message

	if options.KeepSystemMessages {
		for _, s := range systemMessages {
			compressed = append(compressed, s.message)
		}
	}

	for _, s := range selected {
		compressed = append(compressed, s.message)
	}

	originalTokens := c.EstimateTokens(messages)
	compressedTokens := c.EstimateTokens(compressed)
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
		Strategy:               StrategySelective,
	}, nil
}

// scoreImportance evaluates message importance
func (c *SelectiveCompactor) scoreImportance(msg types.Message) int {
	score := 0

	// user messages are usually more important
	if msg.Role == types.RoleUser {
		score += 10
	}

	// messages containing tool calls are important
	if len(msg.ToolCalls) > 0 {
		score += 20
	}

	// long messages may contain more information
	contentLen := 0
	for _, part := range msg.Content {
		contentLen += len(part.Text) + len(part.Thinking)
	}

	if contentLen > 500 {
		score += 5
	} else if contentLen > 100 {
		score += 3
	}

	// messages containing code are important
	for _, part := range msg.Content {
		if strings.Contains(part.Text, "```") {
			score += 15
			break
		}
	}

	// messages containing keywords are important
	keywords := []string{"important", "key", "remember", "note", "summary", "key", "note"}
	for _, part := range msg.Content {
		for _, kw := range keywords {
			if strings.Contains(strings.ToLower(part.Text), kw) {
				score += 5
			}
		}
	}

	return score
}

// sortByImportance sorts by importance descending
func sortByImportance(scored []messageImportance) {
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].importance > scored[i].importance {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
}

// sortByIndex sorts by index ascending
func sortByIndex(scored []messageImportance) {
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].index < scored[i].index {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
}

// EstimateTokens estimates token count
func (c *SelectiveCompactor) EstimateTokens(messages []types.Message) int {
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

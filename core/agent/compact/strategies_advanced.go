// Package compact 提供上下文压缩功能 - 其他压缩策略
package compact

import (
	"context"
	"strings"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// TruncateCompactor 截断压缩器 - 保留最近 N 条消息
type TruncateCompactor struct {
	tokenizer Tokenizer
}

// Compact 执行截断压缩
func (c *TruncateCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTruncate,
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

	// 计算保留的消息数
	keepCount := options.KeepRecentTurns * 2
	if keepCount <= 0 {
		keepCount = 10 // 默认保留 5 轮对话
	}

	if keepCount >= len(conversationMessages) {
		// 无需压缩
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTruncate,
		}, nil
	}

	// 构建压缩后的消息
	var compressed []types.Message

	// 添加系统消息
	if options.KeepSystemMessages {
		compressed = append(compressed, systemMessages...)
	}

	// 添加最近的对话
	startIndex := len(conversationMessages) - keepCount
	compressed = append(compressed, conversationMessages[startIndex:]...)

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
		Strategy:               StrategyTruncate,
	}, nil
}

// estimateTokens 估算 Token 数
func (c *TruncateCompactor) estimateTokens(messages []types.Message) int {
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

// SelectiveCompactor 选择性压缩器 - 基于重要性选择消息
type SelectiveCompactor struct {
	tokenizer Tokenizer
}

// messageImportance 消息重要性评分
type messageImportance struct {
	message   types.Message
	importance int
	index      int
}

// Compact 执行选择性压缩
func (c *SelectiveCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategySelective,
		}, nil
	}

	// 评估每条消息的重要性
	var scored []messageImportance
	for i, msg := range messages {
		score := c.scoreImportance(msg)
		scored = append(scored, messageImportance{
			message:    msg,
			importance: score,
			index:      i,
		})
	}

	// 分离系统消息（总是保留）
	var systemMessages []types.Message
	var conversationScored []messageImportance

	for _, s := range scored {
		if s.message.Role == types.RoleSystem {
			systemMessages = append(systemMessages, s.message)
		} else {
			conversationScored = append(conversationScored, s)
		}
	}

	// 按重要性排序
	sortByImportance(conversationScored)

	// 计算需要保留的消息数
	targetTokens := options.TargetTokens
	if targetTokens <= 0 {
		targetTokens = c.estimateTokens(messages) / 2
	}

	// 累加 Token 直到达到目标
	var selected []messageImportance
	currentTokens := 0

	// 首先保留最重要的消息
	for _, s := range conversationScored {
		msgTokens := c.estimateTokens([]types.Message{s.message})
		if currentTokens+msgTokens <= targetTokens || len(selected) < 3 {
			selected = append(selected, s)
			currentTokens += msgTokens
		}
	}

	// 按原始顺序重新排序
	sortByIndex(selected)

	// 构建压缩后的消息
	var compressed []types.Message

	if options.KeepSystemMessages {
		compressed = append(compressed, systemMessages...)
	}

	for _, s := range selected {
		compressed = append(compressed, s.message)
	}

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
		Strategy:               StrategySelective,
	}, nil
}

// scoreImportance 评估消息重要性
func (c *SelectiveCompactor) scoreImportance(msg types.Message) int {
	score := 0

	// 用户消息通常更重要
	if msg.Role == types.RoleUser {
		score += 10
	}

	// 包含工具调用的消息重要
	if len(msg.ToolCalls) > 0 {
		score += 20
	}

	// 长消息可能包含更多信息
	contentLen := 0
	for _, part := range msg.Content {
		contentLen += len(part.Text) + len(part.Thinking)
	}

	if contentLen > 500 {
		score += 5
	} else if contentLen > 100 {
		score += 3
	}

	// 包含代码的消息重要
	for _, part := range msg.Content {
		if strings.Contains(part.Text, "```") {
			score += 15
			break
		}
	}

	// 包含关键词的消息重要
	keywords := []string{"important", "key", "remember", "note", "总结", "关键", "注意"}
	for _, part := range msg.Content {
		for _, kw := range keywords {
			if strings.Contains(strings.ToLower(part.Text), kw) {
				score += 5
			}
		}
	}

	return score
}

// sortByImportance 按重要性降序排序
func sortByImportance(scored []messageImportance) {
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].importance > scored[i].importance {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
}

// sortByIndex 按索引升序排序
func sortByIndex(scored []messageImportance) {
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].index < scored[i].index {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
}

// estimateTokens 估算 Token 数
func (c *SelectiveCompactor) estimateTokens(messages []types.Message) int {
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

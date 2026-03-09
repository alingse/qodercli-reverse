// Package compact 提供上下文压缩功能 - 高级压缩策略
package compact

import (
	"context"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// HybridCompactor 混合压缩器 - 结合多种策略
type HybridCompactor struct {
	provider  provider.Client
	tokenizer Tokenizer
}

// Compact 执行混合压缩
func (c *HybridCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyHybrid,
		}, nil
	}

	currentTokens := c.estimateTokens(messages)
	targetTokens := options.TargetTokens
	
	if targetTokens <= 0 {
		targetTokens = currentTokens/ 2
	}

	// 第一阶段：使用截断策略快速减少消息
	truncateCompactor := &TruncateCompactor{tokenizer: c.tokenizer}
	truncateOpts := &CompactOptions{
		KeepRecentTurns:    options.KeepRecentTurns,
		KeepSystemMessages: options.KeepSystemMessages,
	}
	
	truncateResult, err := truncateCompactor.Compact(ctx, messages, truncateOpts)
	if err != nil {
		return nil, err
	}

	// 如果截断后已经达到目标，返回结果
	if truncateResult.CompressedTokens <= targetTokens {
		truncateResult.Strategy = StrategyHybrid
		return truncateResult, nil
	}

	// 第二阶段：对截断后的消息使用摘要策略
	summaryCompactor := &SummaryCompactor{provider: c.provider, tokenizer: c.tokenizer}
	summaryOpts := &CompactOptions{
		KeepRecentTurns:    options.KeepRecentTurns,
		KeepSystemMessages: options.KeepSystemMessages,
		CustomPrompt:       options.CustomPrompt,
	}

	return summaryCompactor.Compact(ctx, messages, summaryOpts)
}

// estimateTokens 估算 Token 数
func (c *HybridCompactor) estimateTokens(messages []types.Message) int {
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

// TokenAwareCompactor Token 感知压缩器 - 基于 Token 使用量优化
type TokenAwareCompactor struct {
	provider  provider.Client
	tokenizer Tokenizer
}

// Compact 执行 Token 感知压缩
func (c *TokenAwareCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.estimateTokens(messages),
			CompressedTokens:       c.estimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTokenAware,
		}, nil
	}

	currentTokens := c.estimateTokens(messages)
	maxTokens := options.MaxTokens
	targetTokens := options.TargetTokens

	if targetTokens <= 0 {
		// 默认目标是最大 Token 数的 50%
		targetTokens = maxTokens/ 2
	}

	// 如果当前 Token 数已经小于目标，无需压缩
	if currentTokens <= targetTokens {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         currentTokens,
			CompressedTokens:       currentTokens,
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTokenAware,
		}, nil
	}

	// 分析每条消息的 Token 使用
	type messageTokenInfo struct {
		message types.Message
		tokens int
		index   int
	}

	var tokenInfo []messageTokenInfo
	for i, msg := range messages {
		tokens := c.estimateTokens([]types.Message{msg})
		tokenInfo = append(tokenInfo, messageTokenInfo{
			message: msg,
			tokens:  tokens,
			index:   i,
		})
	}

	// 分离系统消息
	var systemMessages []messageTokenInfo
	var conversationMessages []messageTokenInfo

	for _, info := range tokenInfo {
		if info.message.Role == types.RoleSystem {
			systemMessages = append(systemMessages, info)
		} else {
			conversationMessages = append(conversationMessages, info)
		}
	}

	// 按 Token 数降序排序（Token 多的优先压缩）
	sortByTokens(conversationMessages)

	// 计算需要移除的 Token 数
	tokensToRemove := currentTokens- targetTokens

	// 标记需要移除的消息
	var toRemove []int
	removedTokens := 0

	for _, info := range conversationMessages {
		if removedTokens >= tokensToRemove {
			break
		}
		toRemove = append(toRemove, info.index)
		removedTokens += info.tokens
	}

	// 构建压缩后的消息
	var compressed []types.Message

	// 添加系统消息
	if options.KeepSystemMessages {
		for _, info := range systemMessages {
			compressed = append(compressed, info.message)
		}
	}

	// 添加未标记移除的消息
	removeSet := make(map[int]bool)
	for _, idx := range toRemove {
		removeSet[idx] = true
	}

	for _, info := range tokenInfo {
		if !removeSet[info.index] && info.message.Role != types.RoleSystem {
			compressed = append(compressed, info.message)
		}
	}

	// 按原始顺序排序
	sortByMessageIndex(compressed, messages)

	originalTokens := currentTokens
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
		Strategy:               StrategyTokenAware,
	}, nil
}

// sortByTokens 按 Token 数降序排序
func sortByTokens(infos []messageTokenInfo) {
	for i := 0; i < len(infos)-1; i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[j].tokens > infos[i].tokens {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}
}

// sortByMessageIndex 按原始消息索引排序
func sortByMessageIndex(compressed []types.Message, original []types.Message) {
	indexMap := make(map[string]int)
	for i, msg := range original {
		key := messageKey(msg)
		indexMap[key] = i
	}

	for i := 0; i < len(compressed)-1; i++ {
		for j := i + 1; j < len(compressed); j++ {
			keyI := messageKey(compressed[i])
			keyJ := messageKey(compressed[j])
			if indexMap[keyI] > indexMap[keyJ] {
				compressed[i], compressed[j] = compressed[j], compressed[i]
			}
		}
	}
}

// messageKey 生成消息的唯一键
func messageKey(msg types.Message) string {
	var key string
	for _, part := range msg.Content {
		key += part.Text
	}
	return key
}

// estimateTokens 估算 Token 数
func (c *TokenAwareCompactor) estimateTokens(messages []types.Message) int {
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

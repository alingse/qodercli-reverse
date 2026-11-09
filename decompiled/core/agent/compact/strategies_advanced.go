// Package compact provides context compaction - advanced compaction strategies
package compact

import (
	"context"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// messageTokenInfo message token info (used by TokenAwareCompactor)
type messageTokenInfo struct {
	message types.Message
	tokens  int
	index   int
}

// HybridCompactor hybrid compactor - combines multiple strategies
type HybridCompactor struct {
	provider  provider.Client
	tokenizer Tokenizer
}

// Compact performs hybrid compaction
func (c *HybridCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyHybrid,
		}, nil
	}

	currentTokens := c.EstimateTokens(messages)
	targetTokens := options.TargetTokens

	if targetTokens <= 0 {
		targetTokens = currentTokens / 2
	}

	// Stage 1: use truncation strategy to quickly reduce messages
	truncateCompactor := &TruncateCompactor{tokenizer: c.tokenizer}
	truncateOpts := &CompactOptions{
		KeepRecentTurns:    options.KeepRecentTurns,
		KeepSystemMessages: options.KeepSystemMessages,
	}

	truncateResult, err := truncateCompactor.Compact(ctx, messages, truncateOpts)
	if err != nil {
		return nil, err
	}

	// if truncation already reaches target, return result
	if truncateResult.CompressedTokens <= targetTokens {
		truncateResult.Strategy = StrategyHybrid
		return truncateResult, nil
	}

	// Stage 2: use summary strategy on truncated messages
	summaryCompactor := &SummaryCompactor{provider: c.provider, tokenizer: c.tokenizer}
	summaryOpts := &CompactOptions{
		KeepRecentTurns:    options.KeepRecentTurns,
		KeepSystemMessages: options.KeepSystemMessages,
		CustomPrompt:       options.CustomPrompt,
	}

	return summaryCompactor.Compact(ctx, messages, summaryOpts)
}

// EstimateTokens estimates token count
func (c *HybridCompactor) EstimateTokens(messages []types.Message) int {
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

// TokenAwareCompactor token-aware compactor - optimizes based on token usage
type TokenAwareCompactor struct {
	provider  provider.Client
	tokenizer Tokenizer
}

// Compact performs token-aware compaction
func (c *TokenAwareCompactor) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	if len(messages) <= 2 {
		return &CompactResult{
			OriginalMessageCount:   len(messages),
			CompressedMessageCount: len(messages),
			OriginalTokens:         c.EstimateTokens(messages),
			CompressedTokens:       c.EstimateTokens(messages),
			SavedTokens:            0,
			CompressionRatio:       0,
			Strategy:               StrategyTokenAware,
		}, nil
	}

	currentTokens := c.EstimateTokens(messages)
	maxTokens := options.MaxTokens
	targetTokens := options.TargetTokens

	if targetTokens <= 0 {
		// default target is 50% of max tokens
		targetTokens = maxTokens / 2
	}

	// if current token count is already below target, no compaction needed
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

	// analyze token usage of each message
	var tokenInfo []messageTokenInfo
	for i, msg := range messages {
		tokens := c.EstimateTokens([]types.Message{msg})
		tokenInfo = append(tokenInfo, messageTokenInfo{
			message: msg,
			tokens:  tokens,
			index:   i,
		})
	}

	// separate system messages
	var systemMessages []messageTokenInfo
	var conversationMessages []messageTokenInfo

	for _, info := range tokenInfo {
		if info.message.Role == types.RoleSystem {
			systemMessages = append(systemMessages, info)
		} else {
			conversationMessages = append(conversationMessages, info)
		}
	}

	// sort by token count descending (high-token messages prioritized for compaction)
	sortByTokens(conversationMessages)

	// calculate tokens to remove
	tokensToRemove := currentTokens - targetTokens

	// mark messages to remove
	var toRemove []int
	removedTokens := 0

	for _, info := range conversationMessages {
		if removedTokens >= tokensToRemove {
			break
		}
		toRemove = append(toRemove, info.index)
		removedTokens += info.tokens
	}

	// build compressed messages
	var compressed []types.Message

	// add system messages
	if options.KeepSystemMessages {
		for _, info := range systemMessages {
			compressed = append(compressed, info.message)
		}
	}

	// add messages not marked for removal
	removeSet := make(map[int]bool)
	for _, idx := range toRemove {
		removeSet[idx] = true
	}

	for _, info := range tokenInfo {
		if !removeSet[info.index] && info.message.Role != types.RoleSystem {
			compressed = append(compressed, info.message)
		}
	}

	// sort by original order
	sortByMessageIndex(compressed, messages)

	originalTokens := currentTokens
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
		Strategy:               StrategyTokenAware,
	}, nil
}

// sortByTokens sorts by token count descending
func sortByTokens(infos []messageTokenInfo) {
	for i := 0; i < len(infos)-1; i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[j].tokens > infos[i].tokens {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}
}

// sortByMessageIndex sorts by original message index
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

// messageKey generates a unique key for a message
func messageKey(msg types.Message) string {
	var key string
	for _, part := range msg.Content {
		key += part.Text
	}
	return key
}

// EstimateTokens estimates token count
func (c *TokenAwareCompactor) EstimateTokens(messages []types.Message) int {
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

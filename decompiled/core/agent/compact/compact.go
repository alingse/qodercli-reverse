// Package compact provides context compaction
// Supports multiple compaction strategies, token management, and Session Memory integration
package compact

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/provider"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// CompactTrigger compaction trigger type
type CompactTrigger string

const (
	// TriggerAuto auto trigger (context close to model limit)
	TriggerAuto CompactTrigger = "auto"
	// TriggerManual manual trigger (user runs /compact command)
	TriggerManual CompactTrigger = "manual"
	// TriggerReload reload trigger (session restore)
	TriggerReload CompactTrigger = "reload"
)

// CompactStrategy compaction strategy
type CompactStrategy int

const (
	// StrategySummary summary strategy - uses LLM to generate a summary
	StrategySummary CompactStrategy = iota
	// StrategyTruncate truncation strategy - keeps the last N messages
	StrategyTruncate
	// StrategySelective selective strategy - selects messages based on importance
	StrategySelective
	// StrategyHybrid hybrid strategy - combines multiple strategies
	StrategyHybrid
	// StrategyTokenAware token-aware strategy - optimizes based on token usage
	StrategyTokenAware
)

// CompactOptions compaction options
type CompactOptions struct {
	// compaction strategy
	Strategy CompactStrategy
	// trigger type
	Trigger CompactTrigger
	// target token count (after compaction)
	TargetTokens int
	// max token count (model limit)
	MaxTokens int
	// keep system messages
	KeepSystemMessages bool
	// keep tool calls
	KeepToolCalls bool
	// keep last N conversation turns
	KeepRecentTurns int
	// custom prompt (for LLM summary)
	CustomPrompt string
}

// DefaultOptions default compaction options
func DefaultOptions() *CompactOptions {
	return &CompactOptions{
		Strategy:           StrategyHybrid,
		Trigger:            TriggerManual,
		TargetTokens:       0,
		MaxTokens:          200000,
		KeepSystemMessages: true,
		KeepToolCalls:      false,
		KeepRecentTurns:    5,
		CustomPrompt:       "",
	}
}

// CompactResult compaction result
type CompactResult struct {
	// original message count
	OriginalMessageCount int
	// compressed message count
	CompressedMessageCount int
	// original token count
	OriginalTokens int
	// compressed token count
	CompressedTokens int
	// saved token count
	SavedTokens int
	// compression ratio
	CompressionRatio float64
	// strategy used
	Strategy CompactStrategy
	// summary content (if any)
	Summary string
	// compaction metadata
	Metadata *CompactMetadata
}

// CompactMetadata compaction metadata
type CompactMetadata struct {
	// compaction time
	Timestamp time.Time `json:"timestamp"`
	// trigger type
	Trigger CompactTrigger `json:"trigger"`
	// compaction strategy
	Strategy CompactStrategy `json:"strategy"`
	// model info
	Model string `json:"model"`
	// compaction version
	Version string `json:"version"`
}

// Compactor compactor interface
type Compactor interface {
	// Compact performs compaction
	Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error)
	// EstimateTokens estimates token count
	EstimateTokens(messages []types.Message) int
}

// Manager compaction manager
type Manager struct {
	mu          sync.RWMutex
	provider    provider.Client
	tokenizer   Tokenizer
	hooks       []PreCompactHook
	strategy    CompactStrategy
	lastCompact *CompactResult
}

// Tokenizer token counter interface
type Tokenizer interface {
	Count(text string) int
}

// PreCompactHook pre-compaction hook interface
type PreCompactHook interface {
	Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error
}

// PostCompactHook post-compaction hook interface
type PostCompactHook interface {
	Execute(ctx context.Context, result *CompactResult) error
}

// ThresholdHook token threshold hook interface
type ThresholdHook interface {
	Execute(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool
}

// NewManager creates a compaction manager
func NewManager(provider provider.Client, tokenizer Tokenizer) *Manager {
	return &Manager{
		provider:  provider,
		tokenizer: tokenizer,
		strategy:  StrategyHybrid,
	}
}

// SetStrategy sets compaction strategy
func (m *Manager) SetStrategy(strategy CompactStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategy = strategy
}

// AddHook adds a pre-compaction hook
func (m *Manager) AddHook(hook PreCompactHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// Compact performs compaction
func (m *Manager) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	m.mu.RLock()
	strategy := m.strategy
	hooks := m.hooks
	m.mu.RUnlock()

	if options == nil {
		options = DefaultOptions()
	}

	// use configured compaction strategy
	if options.Strategy == 0 {
		options.Strategy = strategy
	}

	// execute pre-compaction hooks
	for _, hook := range hooks {
		if err := hook.Execute(ctx, messages, options); err != nil {
			return nil, fmt.Errorf("pre-compact hook error: %w", err)
		}
	}

	// record start time
	startTime := time.Now()
	log.Debug("Starting compact: %d messages, strategy: %v", len(messages), options.Strategy)

	// get compactor
	compactor := m.getCompactor(options.Strategy)

	// perform compaction
	result, err := compactor.Compact(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("compact error: %w", err)
	}

	// fill metadata
	result.Metadata = &CompactMetadata{
		Timestamp: startTime,
		Trigger:   options.Trigger,
		Strategy:  options.Strategy,
		Version:   "1.0.0",
	}

	// save last compaction result
	m.mu.Lock()
	m.lastCompact = result
	m.mu.Unlock()

	log.Debug("Compact completed: %d -> %d messages, saved %d tokens (%.2f%%)",
		result.OriginalMessageCount,
		result.CompressedMessageCount,
		result.SavedTokens,
		result.CompressionRatio)

	return result, nil
}

// getCompactor gets a compactor
func (m *Manager) getCompactor(strategy CompactStrategy) Compactor {
	switch strategy {
	case StrategySummary:
		return &SummaryCompactor{provider: m.provider, tokenizer: m.tokenizer}
	case StrategyTruncate:
		return &TruncateCompactor{tokenizer: m.tokenizer}
	case StrategySelective:
		return &SelectiveCompactor{tokenizer: m.tokenizer}
	case StrategyHybrid:
		return &HybridCompactor{provider: m.provider, tokenizer: m.tokenizer}
	case StrategyTokenAware:
		return &TokenAwareCompactor{provider: m.provider, tokenizer: m.tokenizer}
	default:
		return &HybridCompactor{provider: m.provider, tokenizer: m.tokenizer}
	}
}

// EstimateTokens estimates token count
func (m *Manager) EstimateTokens(messages []types.Message) int {
	if m.tokenizer == nil {
		return estimateTokensSimple(messages)
	}

	total := 0
	for _, msg := range messages {
		for _, part := range msg.Content {
			if part.Type == "text" {
				total += m.tokenizer.Count(part.Text)
			}
			if part.Thinking != "" {
				total += m.tokenizer.Count(part.Thinking)
			}
		}
		// extra overhead for tool calls
		for _, tc := range msg.ToolCalls {
			total += m.tokenizer.Count(tc.Name) + m.tokenizer.Count(tc.Arguments)
		}
	}
	return total
}

// estimateTokensSimple simple token estimate (fallback)
func estimateTokensSimple(messages []types.Message) int {
	total := 0
	for _, msg := range messages {
		for _, part := range msg.Content {
			// rough estimate by character count (English ~4 chars/token, Chinese ~2 chars/token)
			total += len(part.Text) / 3
			total += len(part.Thinking) / 3
		}
		// base overhead per message
		total += 4
	}
	return total
}

// GetLastCompact gets the last compaction result
func (m *Manager) GetLastCompact() *CompactResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCompact
}

// ShouldTriggerAuto whether auto compaction should trigger
func (m *Manager) ShouldTriggerAuto(currentTokens int, maxTokens int, threshold float64) bool {
	return float64(currentTokens)/float64(maxTokens) > threshold
}

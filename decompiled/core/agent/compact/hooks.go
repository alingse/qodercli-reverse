// Package compact provides context compaction - Hook system
package compact

import (
	"context"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// HookType hook type
type HookType string

const (
	// HookTypePreCompact pre-compaction hook
	HookTypePreCompact HookType = "pre_compact"
	// HookTypePostCompact post-compaction hook
	HookTypePostCompact HookType = "post_compact"
	// HookTypeOnTokenThreshold token threshold hook
	HookTypeOnTokenThreshold HookType = "on_token_threshold"
)

// Note: PreCompactHook, PostCompactHook, ThresholdHook interfaces are defined in compact.go

// HookFunc functional hook
type HookFunc func(ctx context.Context, messages []types.Message, options *CompactOptions) error

// Execute executes the hook
func (f HookFunc) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	return f(ctx, messages, options)
}

// PostHookFunc functional post-compaction hook
type PostHookFunc func(ctx context.Context, result *CompactResult) error

// Execute executes the post-compaction hook
func (f PostHookFunc) Execute(ctx context.Context, result *CompactResult) error {
	return f(ctx, result)
}

// ThresholdHookFunc functional threshold hook
type ThresholdHookFunc func(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool

// Execute executes the threshold hook
func (f ThresholdHookFunc) Execute(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool {
	return f(ctx, currentTokens, maxTokens, threshold)
}

// BuiltinHooks built-in hook collection
type BuiltinHooks struct {
	preHooks       []PreCompactHook
	postHooks      []PostCompactHook
	thresholdHooks []ThresholdHook
}

// NewBuiltinHooks creates a built-in hook collection
func NewBuiltinHooks() *BuiltinHooks {
	return &BuiltinHooks{
		preHooks:       make([]PreCompactHook, 0),
		postHooks:      make([]PostCompactHook, 0),
		thresholdHooks: make([]ThresholdHook, 0),
	}
}

// AddPreHook adds a pre-compaction hook
func (h *BuiltinHooks) AddPreHook(hook PreCompactHook) {
	h.preHooks = append(h.preHooks, hook)
}

// AddPostHook adds a post-compaction hook
func (h *BuiltinHooks) AddPostHook(hook PostCompactHook) {
	h.postHooks = append(h.postHooks, hook)
}

// AddThresholdHook adds a threshold hook
func (h *BuiltinHooks) AddThresholdHook(hook ThresholdHook) {
	h.thresholdHooks = append(h.thresholdHooks, hook)
}

// ExecutePreHooks executes all pre-compaction hooks
func (h *BuiltinHooks) ExecutePreHooks(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	for _, hook := range h.preHooks {
		if err := hook.Execute(ctx, messages, options); err != nil {
			log.Error("Pre-compact hook error: %v", err)
			return err
		}
	}
	return nil
}

// ExecutePostHooks executes all post-compaction hooks
func (h *BuiltinHooks) ExecutePostHooks(ctx context.Context, result *CompactResult) error {
	for _, hook := range h.postHooks {
		if err := hook.Execute(ctx, result); err != nil {
			log.Error("Post-compact hook error: %v", err)
			return err
		}
	}
	return nil
}

// ExecuteThresholdHooks executes all threshold hooks
func (h *BuiltinHooks) ExecuteThresholdHooks(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool {
	shouldTrigger := false
	for _, hook := range h.thresholdHooks {
		if hook.Execute(ctx, currentTokens, maxTokens, threshold) {
			shouldTrigger = true
		}
	}
	return shouldTrigger
}

// ========== built-in hook implementations ==========

// LoggingHook logging hook
type LoggingHook struct{}

// Execute performs logging
func (h *LoggingHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	log.Debug("Compact started: trigger=%s, strategy=%s, messages=%d",
		options.Trigger, strategyToString(options.Strategy), len(messages))
	return nil
}

// ValidationHook validation hook
type ValidationHook struct{}

// Execute performs validation
func (h *ValidationHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	if len(messages) == 0 {
		return fmt.Errorf("empty messages")
	}

	if options.MaxTokens <= 0 {
		return fmt.Errorf("invalid max tokens: %d", options.MaxTokens)
	}

	return nil
}

// MetricsHook metrics collection hook
type MetricsHook struct {
	compactCount     int
	totalSavedTokens int
	lastError        error
}

// Execute performs metrics collection
func (h *MetricsHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	// record pre-compaction state
	return nil
}

// GetMetrics gets metrics
func (h *MetricsHook) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"compact_count":      h.compactCount,
		"total_saved_tokens": h.totalSavedTokens,
		"last_error":         h.lastError,
	}
}

// CacheHook cache hook - used to cache compaction results
type CacheHook struct {
	cache map[string]*CompactResult
}

// NewCacheHook creates a cache hook
func NewCacheHook() *CacheHook {
	return &CacheHook{
		cache: make(map[string]*CompactResult),
	}
}

// Execute performs cache check
func (h *CacheHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	// generate a hash of messages as cache key
	key := hashMessages(messages)

	// check cache
	if result, ok := h.cache[key]; ok {
		log.Debug("Cache hit for compact result")
		// can return a special error to indicate using cached result
		// more complex mechanism needed in actual usage
		_ = result
	}

	return nil
}

// Store stores compaction result
func (h *CacheHook) Store(messages []types.Message, result *CompactResult) {
	key := hashMessages(messages)
	h.cache[key] = result
}

// Clear clears cache
func (h *CacheHook) Clear() {
	h.cache = make(map[string]*CompactResult)
}

// hashMessages generates a message hash
func hashMessages(messages []types.Message) string {
	// simplified implementation; should use a real hash function in practice
	var hash string
	for _, msg := range messages {
		for _, part := range msg.Content {
			hash += part.Text
		}
	}
	return fmt.Sprintf("%d", len(hash))
}

// strategyToString converts strategy to string
func strategyToString(strategy CompactStrategy) string {
	switch strategy {
	case StrategySummary:
		return "summary"
	case StrategyTruncate:
		return "truncate"
	case StrategySelective:
		return "selective"
	case StrategyHybrid:
		return "hybrid"
	case StrategyTokenAware:
		return "token_aware"
	default:
		return "unknown"
	}
}

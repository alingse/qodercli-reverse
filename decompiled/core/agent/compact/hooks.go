// Package compact 提供上下文压缩功能 - Hook 系统
package compact

import (
	"context"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// HookType Hook 类型
type HookType string

const (
	// HookTypePreCompact 压缩前 Hook
	HookTypePreCompact HookType = "pre_compact"
	// HookTypePostCompact 压缩后 Hook
	HookTypePostCompact HookType = "post_compact"
	// HookTypeOnTokenThreshold Token 阈值 Hook
	HookTypeOnTokenThreshold HookType = "on_token_threshold"
)

// Note: PreCompactHook, PostCompactHook, ThresholdHook interfaces are defined in compact.go

// HookFunc 函数式 Hook
type HookFunc func(ctx context.Context, messages []types.Message, options *CompactOptions) error

// Execute 执行 Hook
func (f HookFunc) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	return f(ctx, messages, options)
}

// PostHookFunc 函数式压缩后 Hook
type PostHookFunc func(ctx context.Context, result *CompactResult) error

// Execute 执行压缩后 Hook
func (f PostHookFunc) Execute(ctx context.Context, result *CompactResult) error {
	return f(ctx, result)
}

// ThresholdHookFunc 函数式阈值 Hook
type ThresholdHookFunc func(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool

// Execute 执行阈值 Hook
func (f ThresholdHookFunc) Execute(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool {
	return f(ctx, currentTokens, maxTokens, threshold)
}

// BuiltinHooks 内置 Hook 集合
type BuiltinHooks struct {
	preHooks       []PreCompactHook
	postHooks      []PostCompactHook
	thresholdHooks []ThresholdHook
}

// NewBuiltinHooks 创建内置 Hook 集合
func NewBuiltinHooks() *BuiltinHooks {
	return &BuiltinHooks{
		preHooks:       make([]PreCompactHook, 0),
		postHooks:      make([]PostCompactHook, 0),
		thresholdHooks: make([]ThresholdHook, 0),
	}
}

// AddPreHook 添加压缩前 Hook
func (h *BuiltinHooks) AddPreHook(hook PreCompactHook) {
	h.preHooks = append(h.preHooks, hook)
}

// AddPostHook 添加压缩后 Hook
func (h *BuiltinHooks) AddPostHook(hook PostCompactHook) {
	h.postHooks = append(h.postHooks, hook)
}

// AddThresholdHook 添加阈值 Hook
func (h *BuiltinHooks) AddThresholdHook(hook ThresholdHook) {
	h.thresholdHooks = append(h.thresholdHooks, hook)
}

// ExecutePreHooks 执行所有压缩前 Hook
func (h *BuiltinHooks) ExecutePreHooks(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	for _, hook := range h.preHooks {
		if err := hook.Execute(ctx, messages, options); err != nil {
			log.Error("Pre-compact hook error: %v", err)
			return err
		}
	}
	return nil
}

// ExecutePostHooks 执行所有压缩后 Hook
func (h *BuiltinHooks) ExecutePostHooks(ctx context.Context, result *CompactResult) error {
	for _, hook := range h.postHooks {
		if err := hook.Execute(ctx, result); err != nil {
			log.Error("Post-compact hook error: %v", err)
			return err
		}
	}
	return nil
}

// ExecuteThresholdHooks 执行所有阈值 Hook
func (h *BuiltinHooks) ExecuteThresholdHooks(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool {
	shouldTrigger := false
	for _, hook := range h.thresholdHooks {
		if hook.Execute(ctx, currentTokens, maxTokens, threshold) {
			shouldTrigger = true
		}
	}
	return shouldTrigger
}

// ========== 内置 Hook 实现 ==========

// LoggingHook 日志记录 Hook
type LoggingHook struct{}

// Execute 执行日志记录
func (h *LoggingHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	log.Debug("Compact started: trigger=%s, strategy=%s, messages=%d",
		options.Trigger, strategyToString(options.Strategy), len(messages))
	return nil
}

// ValidationHook 验证 Hook
type ValidationHook struct{}

// Execute 执行验证
func (h *ValidationHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	if len(messages) == 0 {
		return fmt.Errorf("empty messages")
	}

	if options.MaxTokens <= 0 {
		return fmt.Errorf("invalid max tokens: %d", options.MaxTokens)
	}

	return nil
}

// MetricsHook 指标收集 Hook
type MetricsHook struct {
	compactCount     int
	totalSavedTokens int
	lastError        error
}

// Execute 执行指标收集
func (h *MetricsHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	// 记录压缩前的状态
	return nil
}

// GetMetrics 获取指标
func (h *MetricsHook) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"compact_count":      h.compactCount,
		"total_saved_tokens": h.totalSavedTokens,
		"last_error":         h.lastError,
	}
}

// CacheHook 缓存 Hook - 用于缓存压缩结果
type CacheHook struct {
	cache map[string]*CompactResult
}

// NewCacheHook 创建缓存 Hook
func NewCacheHook() *CacheHook {
	return &CacheHook{
		cache: make(map[string]*CompactResult),
	}
}

// Execute 执行缓存检查
func (h *CacheHook) Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error {
	// 生成消息的哈希作为缓存键
	key := hashMessages(messages)

	// 检查缓存
	if result, ok := h.cache[key]; ok {
		log.Debug("Cache hit for compact result")
		// 这里可以返回特殊错误来表示使用缓存结果
		// 实际使用中需要更复杂的机制
		_ = result
	}

	return nil
}

// Store 存储压缩结果
func (h *CacheHook) Store(messages []types.Message, result *CompactResult) {
	key := hashMessages(messages)
	h.cache[key] = result
}

// Clear 清空缓存
func (h *CacheHook) Clear() {
	h.cache = make(map[string]*CompactResult)
}

// hashMessages 生成消息哈希
func hashMessages(messages []types.Message) string {
	// 简化实现，实际应该使用真正的哈希函数
	var hash string
	for _, msg := range messages {
		for _, part := range msg.Content {
			hash += part.Text
		}
	}
	return fmt.Sprintf("%d", len(hash))
}

// strategyToString 策略转字符串
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

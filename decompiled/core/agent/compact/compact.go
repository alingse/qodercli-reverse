// Package compact 提供上下文压缩功能
// 支持多种压缩策略、Token 管理、Session Memory 集成
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

// CompactTrigger 压缩触发类型
type CompactTrigger string

const (
	// TriggerAuto 自动触发（上下文接近模型限制）
	TriggerAuto CompactTrigger = "auto"
	// TriggerManual 手动触发（用户执行 /compact 命令）
	TriggerManual CompactTrigger = "manual"
	// TriggerReload 重载触发（会话恢复时）
	TriggerReload CompactTrigger = "reload"
)

// CompactStrategy 压缩策略
type CompactStrategy int

const (
	// StrategySummary 摘要策略 - 使用 LLM 生成摘要
	StrategySummary CompactStrategy = iota
	// StrategyTruncate 截断策略 - 保留最近 N 条消息
	StrategyTruncate
	// StrategySelective 选择性策略 - 基于重要性选择消息
	StrategySelective
	// StrategyHybrid 混合策略 - 结合多种策略
	StrategyHybrid
	// StrategyTokenAware Token 感知策略 - 基于 Token 使用量优化
	StrategyTokenAware
)

// CompactOptions 压缩选项
type CompactOptions struct {
	// 压缩策略
	Strategy CompactStrategy
	// 触发类型
	Trigger CompactTrigger
	// 目标 Token 数（压缩后）
	TargetTokens int
	// 最大 Token 数（模型限制）
	MaxTokens int
	// 保留系统消息
	KeepSystemMessages bool
	// 保留工具调用
	KeepToolCalls bool
	// 保留最近 N 轮对话
	KeepRecentTurns int
	// 自定义提示词（用于 LLM 摘要）
	CustomPrompt string
}

// DefaultOptions 默认压缩选项
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

// CompactResult 压缩结果
type CompactResult struct {
	// 原始消息数
	OriginalMessageCount int
	// 压缩后消息数
	CompressedMessageCount int
	// 原始 Token 数
	OriginalTokens int
	// 压缩后 Token 数
	CompressedTokens int
	// 节省的 Token 数
	SavedTokens int
	// 压缩率
	CompressionRatio float64
	// 使用的策略
	Strategy CompactStrategy
	// 摘要内容（如果有）
	Summary string
	// 压缩元数据
	Metadata *CompactMetadata
}

// CompactMetadata 压缩元数据
type CompactMetadata struct {
	// 压缩时间
	Timestamp time.Time `json:"timestamp"`
	// 触发类型
	Trigger CompactTrigger `json:"trigger"`
	// 压缩策略
	Strategy CompactStrategy `json:"strategy"`
	// 模型信息
	Model string `json:"model"`
	// 压缩版本
	Version string `json:"version"`
}

// Compactor 压缩器接口
type Compactor interface {
	// Compact 执行压缩
	Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error)
	// EstimateTokens 估算 Token 数
	EstimateTokens(messages []types.Message) int
}

// Manager 压缩管理器
type Manager struct {
	mu          sync.RWMutex
	provider    provider.Client
	tokenizer   Tokenizer
	hooks       []PreCompactHook
	strategy    CompactStrategy
	lastCompact *CompactResult
}

// Tokenizer Token 计算器接口
type Tokenizer interface {
	Count(text string) int
}

// PreCompactHook Pre-Compact Hook 接口
type PreCompactHook interface {
	Execute(ctx context.Context, messages []types.Message, options *CompactOptions) error
}

// PostCompactHook Post-Compact Hook 接口
type PostCompactHook interface {
	Execute(ctx context.Context, result *CompactResult) error
}

// ThresholdHook Token 阈值 Hook 接口
type ThresholdHook interface {
	Execute(ctx context.Context, currentTokens int, maxTokens int, threshold float64) bool
}

// NewManager 创建压缩管理器
func NewManager(provider provider.Client, tokenizer Tokenizer) *Manager {
	return &Manager{
		provider:  provider,
		tokenizer: tokenizer,
		strategy:  StrategyHybrid,
	}
}

// SetStrategy 设置压缩策略
func (m *Manager) SetStrategy(strategy CompactStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategy = strategy
}

// AddHook 添加 Pre-Compact Hook
func (m *Manager) AddHook(hook PreCompactHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// Compact 执行压缩
func (m *Manager) Compact(ctx context.Context, messages []types.Message, options *CompactOptions) (*CompactResult, error) {
	m.mu.RLock()
	strategy := m.strategy
	hooks := m.hooks
	m.mu.RUnlock()

	if options == nil {
		options = DefaultOptions()
	}

	// 使用配置的压缩策略
	if options.Strategy == 0 {
		options.Strategy = strategy
	}

	// 执行 Pre-Compact Hooks
	for _, hook := range hooks {
		if err := hook.Execute(ctx, messages, options); err != nil {
			return nil, fmt.Errorf("pre-compact hook error: %w", err)
		}
	}

	// 记录开始时间
	startTime := time.Now()
	log.Debug("Starting compact: %d messages, strategy: %v", len(messages), options.Strategy)

	// 获取压缩器
	compactor := m.getCompactor(options.Strategy)

	// 执行压缩
	result, err := compactor.Compact(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("compact error: %w", err)
	}

	// 填充元数据
	result.Metadata = &CompactMetadata{
		Timestamp: startTime,
		Trigger:   options.Trigger,
		Strategy:  options.Strategy,
		Version:   "1.0.0",
	}

	// 保存最后一次压缩结果
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

// getCompactor 获取压缩器
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

// EstimateTokens 估算 Token 数
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
		// 工具调用的额外开销
		for _, tc := range msg.ToolCalls {
			total += m.tokenizer.Count(tc.Name) + m.tokenizer.Count(tc.Arguments)
		}
	}
	return total
}

// estimateTokensSimple 简单的 Token 估算（备用方案）
func estimateTokensSimple(messages []types.Message) int {
	total := 0
	for _, msg := range messages {
		for _, part := range msg.Content {
			// 按字符数粗略估算（英文约 4 字符/token，中文约 2 字符/token）
			total += len(part.Text) / 3
			total += len(part.Thinking) / 3
		}
		// 每条消息的基础开销
		total += 4
	}
	return total
}

// GetLastCompact 获取最后一次压缩结果
func (m *Manager) GetLastCompact() *CompactResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCompact
}

// ShouldTriggerAuto 是否应该触发自动压缩
func (m *Manager) ShouldTriggerAuto(currentTokens int, maxTokens int, threshold float64) bool {
	return float64(currentTokens)/float64(maxTokens) > threshold
}

// Package agent Agent 集成 compact 功能
package agent

import (
	"context"
	"fmt"

	"github.com/alingse/qodercli-reverse/core/agent/compact"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// CompactManager Compact 管理器（扩展 Agent 功能）
type CompactManager struct {
	manager   *compact.Manager
	session  *compact.SessionMemory
	tokenMgr  *compact.TokenManager
	config    *CompactConfig
}

// CompactConfig Compact 配置
type CompactConfig struct {
	// 自动压缩启用
	AutoCompactEnabled bool
	// 自动压缩阈值（超过此比例触发）
	AutoCompactThreshold float64
	// 警告阈值
	WarningThreshold float64
	// 压缩策略
	Strategy compact.CompactStrategy
	// 保留最近轮次
	KeepRecentTurns int
	// 最大 Token 数
	MaxTokens int
	// 目标 Token 数
	TargetTokens int
}

// DefaultCompactConfig 默认 Compact 配置
func DefaultCompactConfig() *CompactConfig {
	return &CompactConfig{
		AutoCompactEnabled:   true,
		AutoCompactThreshold: 0.85,
		WarningThreshold:     0.75,
		Strategy:             compact.StrategyHybrid,
		KeepRecentTurns:      5,
		MaxTokens:            200000,
		TargetTokens:         100000,
	}
}

// NewCompactManager 创建 Compact 管理器
func NewCompactManager(sessionID string) *CompactManager {
	tokenizer := compact.NewSimpleTokenizer()
	manager := compact.NewManager(nil, tokenizer)

	return &CompactManager{
		manager:  manager,
		session:  compact.NewSessionMemory(sessionID, 10),
		tokenMgr: compact.NewTokenManager(tokenizer, 200000),
		config:   DefaultCompactConfig(),
	}
}

// SetConfig 设置配置
func (cm *CompactManager) SetConfig(config *CompactConfig) {
	cm.config = config
	cm.manager.SetStrategy(config.Strategy)
	cm.tokenMgr.SetThresholds(config.WarningThreshold, cm.config.AutoCompactThreshold)
	cm.tokenMgr = compact.NewTokenManager(compact.NewSimpleTokenizer(), config.MaxTokens)
}

// CheckAndCompact 检查是否需要压缩并执行压缩
func (cm *CompactManager) CheckAndCompact(ctx context.Context, messages []types.Message) (*compact.CompactResult, error) {
	// 更新 Token 计数
	cm.tokenMgr.Update(messages)

	// 检查是否需要压缩
	if !cm.tokenMgr.ShouldCompact() && !cm.config.AutoCompactEnabled {
		return nil, nil // 不需要压缩
	}

	// 检查是否达到阈值
	current, max, ratio := cm.tokenMgr.GetUsage()
	log.Debug("Token usage: %d/%d (%.2f%%)", current, max, ratio*100)

	if ratio < cm.config.AutoCompactThreshold {
		return nil, nil // 未达到阈值
	}

	// 执行压缩
	options := &compact.CompactOptions{
		Strategy:           cm.config.Strategy,
		Trigger:            compact.TriggerAuto,
		TargetTokens:       cm.config.TargetTokens,
		MaxTokens:          cm.config.MaxTokens,
		KeepSystemMessages: true,
		KeepRecentTurns:    cm.config.KeepRecentTurns,
	}

	result, err := cm.manager.Compact(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("compact error: %w", err)
	}

	// 保存历史
	if result.Metadata != nil {
		cm.session.AddCompactHistory(result.Metadata)
	}

	// 记录结果
	compact.LogCompactResult(result)

	return result, nil
}

// ManualCompact 手动压缩
func (cm *CompactManager) ManualCompact(ctx context.Context, messages []types.Message) (*compact.CompactResult, error) {
	options := &compact.CompactOptions{
		Strategy:           cm.config.Strategy,
		Trigger:            compact.TriggerManual,
		TargetTokens:       cm.config.TargetTokens,
		MaxTokens:          cm.config.MaxTokens,
		KeepSystemMessages: true,
		KeepRecentTurns:    cm.config.KeepRecentTurns,
	}

	result, err := cm.manager.Compact(ctx, messages, options)
	if err != nil {
		return nil, fmt.Errorf("manual compact error: %w", err)
	}

	// 保存历史
	if result.Metadata != nil {
		cm.session.AddCompactHistory(result.Metadata)
	}

	// 记录结果
	compact.LogCompactResult(result)

	return result, nil
}

// GetTokenUsage 获取 Token 使用情况
func (cm *CompactManager) GetTokenUsage() (current int, max int, ratio float64) {
	return cm.tokenMgr.GetUsage()
}

// GetSessionStats 获取会话统计
func (cm *CompactManager) GetSessionStats() map[string]interface{} {
	return cm.session.GetStats()
}

// GetLastCompactResult 获取最后一次压缩结果
func (cm *CompactManager) GetLastCompactResult() *compact.CompactResult {
	return cm.manager.GetLastCompact()
}

// ExportSession 导出会话
func (cm *CompactManager) ExportSession() ([]byte, error) {
	return cm.session.Export()
}

// ImportSession 导入会话
func (cm *CompactManager) ImportSession(data []byte) error {
	return cm.session.Import(data)
}

// ClearSession 清空会话
func (cm *CompactManager) ClearSession() {
	cm.session.Clear()
}

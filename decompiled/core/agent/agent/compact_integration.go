// Package agent Agent compact integration
package agent

import (
	"context"
	"fmt"

	"github.com/alingse/qodercli-reverse/decompiled/core/agent/compact"
	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// CompactManager Compact manager (extends Agent features)
type CompactManager struct {
	manager  *compact.Manager
	session  *compact.SessionMemory
	tokenMgr *compact.TokenManager
	config   *CompactConfig
}

// CompactConfig Compact config
type CompactConfig struct {
	// auto compact enabled
	AutoCompactEnabled bool
	// auto compact threshold (triggered when exceeded)
	AutoCompactThreshold float64
	// warning threshold
	WarningThreshold float64
	// compact strategy
	Strategy compact.CompactStrategy
	// keep recent turns
	KeepRecentTurns int
	// max token count
	MaxTokens int
	// target token count
	TargetTokens int
}

// DefaultCompactConfig default Compact config
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

// NewCompactManager creates Compact manager
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

// SetConfig sets config
func (cm *CompactManager) SetConfig(config *CompactConfig) {
	cm.config = config
	cm.manager.SetStrategy(config.Strategy)
	cm.tokenMgr.SetThresholds(config.WarningThreshold, cm.config.AutoCompactThreshold)
	cm.tokenMgr = compact.NewTokenManager(compact.NewSimpleTokenizer(), config.MaxTokens)
}

// CheckAndCompact checks if compaction needed and executes
func (cm *CompactManager) CheckAndCompact(ctx context.Context, messages []types.Message) (*compact.CompactResult, error) {
	// update token count
	cm.tokenMgr.Update(messages)

	// check if compaction needed
	if !cm.tokenMgr.ShouldCompact() && !cm.config.AutoCompactEnabled {
		return nil, nil // no compaction needed
	}

	// check if threshold reached
	current, max, ratio := cm.tokenMgr.GetUsage()
	log.Debug("Token usage: %d/%d (%.2f%%)", current, max, ratio*100)

	if ratio < cm.config.AutoCompactThreshold {
		return nil, nil // threshold not reached
	}

	// execute compaction
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

	// save history
	if result.Metadata != nil {
		cm.session.AddCompactHistory(result.Metadata)
	}

	// record result
	compact.LogCompactResult(result)

	return result, nil
}

// ManualCompact manual compact
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

	// save history
	if result.Metadata != nil {
		cm.session.AddCompactHistory(result.Metadata)
	}

	// record result
	compact.LogCompactResult(result)

	return result, nil
}

// GetTokenUsage get token usage
func (cm *CompactManager) GetTokenUsage() (current int, max int, ratio float64) {
	return cm.tokenMgr.GetUsage()
}

// GetSessionStats get session stats
func (cm *CompactManager) GetSessionStats() map[string]interface{} {
	return cm.session.GetStats()
}

// GetLastCompactResult get last compact result
func (cm *CompactManager) GetLastCompactResult() *compact.CompactResult {
	return cm.manager.GetLastCompact()
}

// ExportSession export session
func (cm *CompactManager) ExportSession() ([]byte, error) {
	return cm.session.Export()
}

// ImportSession import session
func (cm *CompactManager) ImportSession(data []byte) error {
	return cm.session.Import(data)
}

// ClearSession clear session
func (cm *CompactManager) ClearSession() {
	cm.session.Clear()
}

// Package compact 提供上下文压缩功能 - Session Memory 集成
package compact

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// SessionMemory 会话内存管理
type SessionMemory struct {
	mu            sync.RWMutex
	sessionID     string
	memory       map[string]interface{}
	compactHistory []*CompactMetadata
	maxHistory   int
}

// NewSessionMemory 创建会话内存
func NewSessionMemory(sessionID string, maxHistory int) *SessionMemory {
	if maxHistory <= 0 {
		maxHistory = 10
	}
	return &SessionMemory{
		sessionID:     sessionID,
		memory:        make(map[string]interface{}),
		compactHistory: make([]*CompactMetadata, 0),
		maxHistory:    maxHistory,
	}
}

// Store 存储数据
func (sm *SessionMemory) Store(key string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	sm.memory[key] = raw
	return nil
}

// Load 加载数据
func (sm *SessionMemory) Load(key string, target interface{}) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	value, ok := sm.memory[key]
	if !ok {
		return fmt.Errorf("key not found: %s", key)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// Delete 删除数据
func (sm *SessionMemory) Delete(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.memory, key)
}

// Clear 清空内存
func (sm *SessionMemory) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.memory = make(map[string]interface{})
	sm.compactHistory = make([]*CompactMetadata, 0)
}

// AddCompactHistory 添加压缩历史
func (sm *SessionMemory) AddCompactHistory(metadata *CompactMetadata) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.compactHistory = append(sm.compactHistory, metadata)

	// 限制历史记录数量
	if len(sm.compactHistory) > sm.maxHistory {
		sm.compactHistory = sm.compactHistory[1:]
	}
}

// GetCompactHistory 获取压缩历史
func (sm *SessionMemory) GetCompactHistory() []*CompactMetadata {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	history := make([]*CompactMetadata, len(sm.compactHistory))
	copy(history, sm.compactHistory)
	return history
}

// GetStats 获取统计信息
func (sm *SessionMemory) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	totalSavedTokens := 0
	for _, h := range sm.compactHistory {
		if h != nil {
			// 这里需要从完整的 CompactResult 中获取，简化处理
		}
	}

	return map[string]interface{}{
		"session_id":       sm.sessionID,
		"memory_keys":      len(sm.memory),
		"compact_count":    len(sm.compactHistory),
		"total_saved_tokens": totalSavedTokens,
	}
}

// Export 导出会话内存
func (sm *SessionMemory) Export() ([]byte, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	data := map[string]interface{}{
		"session_id":      sm.sessionID,
		"memory":          sm.memory,
		"compact_history": sm.compactHistory,
		"exported_at":     time.Now().Format(time.RFC3339),
	}

	return json.MarshalIndent(data, "", "  ")
}

// Import 导入会话内存
func (sm *SessionMemory) Import(data []byte) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var imported map[string]interface{}
	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	if sid, ok := imported["session_id"].(string); ok {
		sm.sessionID = sid
	}

	if memory, ok := imported["memory"].(map[string]interface{}); ok {
		sm.memory = memory
	}

	if history, ok := imported["compact_history"].([]interface{}); ok {
		sm.compactHistory = make([]*CompactMetadata, len(history))
		for i, h := range history {
			if hm, ok := h.(*CompactMetadata); ok {
				sm.compactHistory[i] = hm
			}
		}
	}

	return nil
}

// ========== Token 管理器 ==========

// TokenManager Token 管理器
type TokenManager struct {
	mu              sync.RWMutex
	tokenizer       Tokenizer
	currentTokens   int
	maxTokens       int
	warningThreshold float64
	criticalThreshold float64
}

// NewTokenManager 创建 Token 管理器
func NewTokenManager(tokenizer Tokenizer, maxTokens int) *TokenManager {
	return &TokenManager{
		tokenizer:         tokenizer,
		maxTokens:         maxTokens,
		warningThreshold:  0.8,
		criticalThreshold: 0.9,
	}
}

// SetThresholds 设置阈值
func (tm *TokenManager) SetThresholds(warning, critical float64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.warningThreshold = warning
	tm.criticalThreshold = critical
}

// Update 更新 Token 计数
func (tm *TokenManager) Update(messages []types.Message) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.tokenizer != nil {
		tm.currentTokens = tm.calculateTokens(messages)
	}
}

// calculateTokens 计算 Token 数
func (tm *TokenManager) calculateTokens(messages []types.Message) int {
	total := 0
	for _, msg := range messages {
		for _, part := range msg.Content {
			if tm.tokenizer != nil {
				total += tm.tokenizer.Count(part.Text)
				total += tm.tokenizer.Count(part.Thinking)
			} else {
				total += estimateSimpleTokens(part.Text) + estimateSimpleTokens(part.Thinking)
			}
		}
	}
	return total
}

// GetUsage 获取使用情况
func (tm *TokenManager) GetUsage() (current int, max int, ratio float64) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ratio = float64(tm.currentTokens) / float64(tm.maxTokens)
	return tm.currentTokens, tm.maxTokens, ratio
}

// ShouldWarn 是否应该警告
func (tm *TokenManager) ShouldWarn() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ratio := float64(tm.currentTokens) / float64(tm.maxTokens)
	return ratio >= tm.warningThreshold
}

// ShouldCompact 是否应该压缩
func (tm *TokenManager) ShouldCompact() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	ratio := float64(tm.currentTokens) / float64(tm.maxTokens)
	return ratio >= tm.criticalThreshold
}

// GetRemainingTokens 获取剩余 Token 数
func (tm *TokenManager) GetRemainingTokens() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.maxTokens - tm.currentTokens
}

// estimateSimpleTokens 简单的 Token 估算
func estimateSimpleTokens(text string) int {
	if text == "" {
		return 0
	}
	// 中文约 2 字符/token，英文约 4 字符/token
	return len(text) / 3
}

// ========== 工具函数 ==========

// FilterMessages 过滤消息
func FilterMessages(messages []types.Message, filter func(types.Message) bool) []types.Message {
	var result []types.Message
	for _, msg := range messages {
		if filter(msg) {
			result = append(result, msg)
		}
	}
	return result
}

// GroupByTurns 按对话轮次分组
func GroupByTurns(messages []types.Message) [][]types.Message {
	var turns [][]types.Message
	var currentTurn []types.Message

	for _, msg := range messages {
		if msg.Role == types.RoleUser && len(currentTurn) > 0 {
			turns = append(turns, currentTurn)
			currentTurn = nil
		}
		currentTurn = append(currentTurn, msg)
	}

	if len(currentTurn) > 0 {
		turns = append(turns, currentTurn)
	}

	return turns
}

// KeepLastTurns 保留最后 N 轮对话
func KeepLastTurns(messages []types.Message, turns int) []types.Message {
	grouped := GroupByTurns(messages)
	
	if turns >= len(grouped) {
		return messages
	}

	startIndex := len(grouped) - turns
	var result []types.Message
	
	for i := startIndex; i < len(grouped); i++ {
		result = append(result, grouped[i]...)
	}

	return result
}

// ExtractKeyInfo 提取关键信息
func ExtractKeyInfo(messages []types.Message) map[string]interface{} {
	info := map[string]interface{}{
		"user_queries":    make([]string, 0),
		"tool_calls":      make([]string, 0),
		"code_blocks":     make([]string, 0),
		"important_files": make([]string, 0),
	}

	for _, msg := range messages {
		switch msg.Role {
		case types.RoleUser:
			for _, part := range msg.Content {
				if part.Type == "text" && len(part.Text) < 200 {
					info["user_queries"] = append(info["user_queries"].([]string), part.Text)
				}
			}
		case types.RoleAssistant:
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					info["tool_calls"] = append(info["tool_calls"].([]string), tc.Name)
				}
			}
			for _, part := range msg.Content {
				if part.Type == "text" {
					// 提取代码块
					extractCodeBlocks(part.Text, &info)
				}
			}
		}
	}

	return info
}

// extractCodeBlocks 提取代码块
func extractCodeBlocks(text string, info *map[string]interface{}) {
	// 简化实现，查找 ``` 标记
	start := -1
	for i, ch := range text {
		if ch == '`' {
			if start == -1 {
				start = i
			}
		} else if start != -1 {
			if i-start >= 3 {
				// 找到代码块开始
				end := findCodeBlockEnd(text, i)
				if end> 0 {
					code := text[start:end]
					(*info)["code_blocks"] = append((*info)["code_blocks"].([]string), code)
				}
				start = -1
			} else {
				start = -1
			}
		}
	}
}

// findCodeBlockEnd 查找代码块结束位置
func findCodeBlockEnd(text string, start int) int {
	// 简化实现
	for i := start; i < len(text)-2; i++ {
		if text[i:i+3] == "```" {
			return i + 3
		}
	}
	return -1
}

// LogCompactResult 记录压缩结果
func LogCompactResult(result *CompactResult) {
	log.Info("Compact completed: original_messages=%d, compressed_messages=%d, saved_tokens=%d, compression_ratio=%.2f%%, strategy=%v",
		result.OriginalMessageCount,
		result.CompressedMessageCount,
		result.SavedTokens,
		result.CompressionRatio,
		result.Strategy)
}

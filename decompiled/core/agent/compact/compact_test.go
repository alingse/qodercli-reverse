// Package compact 提供上下文压缩功能 - 单元测试
package compact

import (
	"context"
	"fmt"
	"testing"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// TestSimpleCompact 测试简单压缩
func TestSimpleCompact(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	manager := NewManager(nil, tokenizer)

	messages := []types.Message{
		{Role: types.RoleSystem, Content: []types.ContentPart{{Type: "text", Text: "You are a helpful assistant."}}},
		{Role: types.RoleUser, Content: []types.ContentPart{{Type: "text", Text: "Hello, how are you?"}}},
		{Role: types.RoleAssistant, Content: []types.ContentPart{{Type: "text", Text: "I'm doing well, thank you!"}}},
		{Role: types.RoleUser, Content: []types.ContentPart{{Type: "text", Text: "What's the weather like?"}}},
		{Role: types.RoleAssistant, Content: []types.ContentPart{{Type: "text", Text: "I don't have access to real-time weather data."}}},
	}

	options := &CompactOptions{
		Strategy:           StrategyTruncate,
		Trigger:            TriggerManual,
		KeepRecentTurns:    1,
		KeepSystemMessages: true,
	}

	result, err := manager.Compact(context.Background(), messages, options)
	if err != nil {
		t.Fatalf("Compact failed: %v", err)
	}

	t.Logf("Compact result: %+v", result)

	if result.OriginalMessageCount != 5 {
		t.Errorf("Expected original count 5, got %d", result.OriginalMessageCount)
	}

	if result.CompressedMessageCount >= result.OriginalMessageCount {
		t.Errorf("Expected compressed count < %d, got %d", result.OriginalMessageCount, result.CompressedMessageCount)
	}
}

// TestTokenEstimation 测试 Token 估算
func TestTokenEstimation(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"Hello", 2},                         // 5 字符/ 3 ≈ 2
		{"你好", 3},                           // 2 字符 * 1.5 = 3
		{"Hello, 世界!", 4},                   // 混合文本（实际计算）
	}

	for _, tt := range tests {
		got := tokenizer.Count(tt.text)
		if got < tt.expected-1 || got > tt.expected+1 {
			t.Errorf("Count(%q) = %d, want around %d", tt.text, got, tt.expected)
		}
	}
}

// TestHybridStrategy 测试混合策略
func TestHybridStrategy(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	manager := NewManager(nil, tokenizer)

	// 创建更多消息
	var messages []types.Message
	messages = append(messages, types.Message{Role: types.RoleSystem, Content: []types.ContentPart{{Type: "text", Text: "System message"}}})
	
	for i := 0; i < 20; i++ {
		q := fmt.Sprintf("Question %d", i)
		a := fmt.Sprintf("Answer %d", i)
		messages = append(messages, 
			types.Message{Role: types.RoleUser, Content: []types.ContentPart{{Type: "text", Text: q}}},
			types.Message{Role: types.RoleAssistant, Content: []types.ContentPart{{Type: "text", Text: a}}},
		)
	}

	options := &CompactOptions{
		Strategy:           StrategyHybrid,
		Trigger:            TriggerManual,
		KeepRecentTurns:    3,
		KeepSystemMessages: true,
	}

	result, err := manager.Compact(context.Background(), messages, options)
	if err != nil {
		t.Fatalf("Compact failed: %v", err)
	}

	t.Logf("Hybrid compact result: %+v", result)

	if result.SavedTokens <= 0 {
		t.Error("Expected positive saved tokens")
	}

	if result.CompressionRatio <= 0 {
		t.Error("Expected positive compression ratio")
	}
}

// TestSessionMemory 测试 Session Memory
func TestSessionMemory(t *testing.T) {
	session := NewSessionMemory("test-session-1", 5)

	// 存储数据
	err := session.Store("key1", "value1")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// 加载数据
	var value string
	err = session.Load("key1", &value)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	// 测试导出
	data, err := session.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty export")
	}
}

// TestTokenManager 测试 Token Manager
func TestTokenManager(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	tokenMgr := NewTokenManager(tokenizer, 10000)

	messages := []types.Message{
		{Role: types.RoleUser, Content: []types.ContentPart{{Type: "text", Text: "Hello world"}}},
		{Role: types.RoleAssistant, Content: []types.ContentPart{{Type: "text", Text: "Hi there"}}},
	}

	tokenMgr.Update(messages)

	current, max, ratio := tokenMgr.GetUsage()
	
	t.Logf("Token usage: %d/%d (%.2f%%)", current, max, ratio*100)

	if current <= 0 {
		t.Error("Expected positive token count")
	}

	if max != 10000 {
		t.Errorf("Expected max 10000, got %d", max)
	}
}

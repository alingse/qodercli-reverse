// Package compact provides context compaction - unit tests
package compact

import (
	"context"
	"fmt"
	"testing"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// TestSimpleCompact tests simple compaction
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

// TestTokenEstimation tests token estimation
func TestTokenEstimation(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	tests := []struct {
		text     string
		expected int
	}{
		{"", 0},
		{"Hello", 2},      // 5 chars / 3 ≈ 2
		{"Ni hao", 3},       // 2 words * 1.3 ≈ 3
		{"Hello, world!", 2}, // mixed text (actual calculation)
	}

	for _, tt := range tests {
		got := tokenizer.Count(tt.text)
		if got < tt.expected-1 || got > tt.expected+1 {
			t.Errorf("Count(%q) = %d, want around %d", tt.text, got, tt.expected)
		}
	}
}

// TestHybridStrategy tests hybrid strategy
func TestHybridStrategy(t *testing.T) {
	tokenizer := NewSimpleTokenizer()
	manager := NewManager(nil, tokenizer)

	// create more messages
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

// TestSessionMemory tests Session Memory
func TestSessionMemory(t *testing.T) {
	session := NewSessionMemory("test-session-1", 5)

	// store data
	err := session.Store("key1", "value1")
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// load data
	var value string
	err = session.Load("key1", &value)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	// test export
	data, err := session.Export()
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty export")
	}
}

// TestTokenManager tests Token Manager
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

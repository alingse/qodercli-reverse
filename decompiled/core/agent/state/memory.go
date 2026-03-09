// Package state 会话内存存储
// 反编译自 qodercli v0.1.29
package state

import (
	"encoding/json"
	"sync"
)

// SessionMemory 会话内存存储 - 用于存储临时数据如 Todo 列表
type SessionMemory struct {
	mu   sync.RWMutex
	data map[string]json.RawMessage
}

// NewSessionMemory 创建新的会话内存
func NewSessionMemory() *SessionMemory {
	return &SessionMemory{
		data: make(map[string]json.RawMessage),
	}
}

// Get 获取数据
func (m *SessionMemory) Get(key string) (json.RawMessage, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

// Set 设置数据
func (m *SessionMemory) Set(key string, value json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Delete 删除数据
func (m *SessionMemory) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Clear 清空所有数据
func (m *SessionMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]json.RawMessage)
}

// Keys 返回所有键
func (m *SessionMemory) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

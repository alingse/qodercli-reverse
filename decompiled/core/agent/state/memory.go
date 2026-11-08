// Package state Session memory storage
// Decompiled from qodercli-reverse v0.1.29
package state

import (
	"encoding/json"
	"sync"
)

// SessionMemory Session memory storage - used to store temporary data such as Todo lists
type SessionMemory struct {
	mu   sync.RWMutex
	data map[string]json.RawMessage
}

// NewSessionMemory Creates a new session memory
func NewSessionMemory() *SessionMemory {
	return &SessionMemory{
		data: make(map[string]json.RawMessage),
	}
}

// Get Retrieves data
func (m *SessionMemory) Get(key string) (json.RawMessage, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.data[key]
	return value, ok
}

// Set Stores data
func (m *SessionMemory) Set(key string, value json.RawMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Delete Deletes data
func (m *SessionMemory) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Clear Clears all data
func (m *SessionMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]json.RawMessage)
}

// Keys Returns all keys
func (m *SessionMemory) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Package app TUI 应用缓冲区管理
package app

import (
	"sync"
)

// BufferManager 线程安全的缓冲区管理器
type BufferManager struct {
	content string
	mu      sync.RWMutex
}

// NewBufferManager 创建新的缓冲区管理器
func NewBufferManager() *BufferManager {
	return &BufferManager{}
}

// Append 追加内容（线程安全）
func (b *BufferManager) Append(text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content += text
}

// Get 获取当前所有内容（线程安全）
func (b *BufferManager) Get() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.content
}

// Len 获取内容长度（线程安全）
func (b *BufferManager) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.content)
}

// Reset 清空缓冲区（线程安全）
func (b *BufferManager) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.content = ""
}

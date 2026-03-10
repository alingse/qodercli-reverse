// Package app TUI 应用输入队列管理
package app

import (
	"errors"
	"sync"
)

var (
	// ErrQueueFull 队列已满
	ErrQueueFull = errors.New("input queue is full")
	// ErrQueueEmpty 队列为空
	ErrQueueEmpty = errors.New("input queue is empty")
)

// InputQueue 输入队列管理器
type InputQueue struct {
	items   []string
	maxSize int
	mu      sync.RWMutex
}

// NewInputQueue 创建新的输入队列
func NewInputQueue(maxSize int) *InputQueue {
	if maxSize <= 0 {
		maxSize = 50 // 默认最大 50 条
	}
	return &InputQueue{
		items:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

// Enqueue 入队（线程安全）
func (q *InputQueue) Enqueue(input string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.maxSize {
		return ErrQueueFull
	}

	q.items = append(q.items, input)
	return nil
}

// Dequeue 出队（线程安全）
func (q *InputQueue) Dequeue() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return "", ErrQueueEmpty
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}

// Len 获取队列长度（线程安全）
func (q *InputQueue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items)
}

// Clear 清空队列（线程安全）
func (q *InputQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = q.items[:0]
}

// IsEmpty 检查队列是否为空（线程安全）
func (q *InputQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.items) == 0
}

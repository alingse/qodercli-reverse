// Package app TUI 应用状态机
package app

import (
	"fmt"
	"sync"

	"github.com/alingse/qodercli-reverse/decompiled/core/log"
)

// StateMachine 状态机管理器
type StateMachine struct {
	current AppState
	mu      sync.RWMutex
}

// NewStateMachine 创建新的状态机
func NewStateMachine(initial AppState) *StateMachine {
	return &StateMachine{
		current: initial,
	}
}

// Current 获取当前状态（线程安全）
func (sm *StateMachine) Current() AppState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// TransitionTo 转换到新状态（线程安全）
func (sm *StateMachine) TransitionTo(next AppState, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.canTransitionTo(next) {
		return fmt.Errorf("invalid state transition: %s -> %s", sm.current, next)
	}

	old := sm.current
	sm.current = next
	log.Debug(fmt.Sprintf("State transition: %s -> %s (reason: %s)", old, next, reason))
	return nil
}

// CanTransitionTo 检查是否可以转换到目标状态（线程安全）
func (sm *StateMachine) CanTransitionTo(next AppState) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.canTransitionTo(next)
}

// canTransitionTo 内部方法：检查状态转换是否合法（不加锁）
func (sm *StateMachine) canTransitionTo(next AppState) bool {
	// 任意状态都可以转换到 Quitting
	if next == StateQuitting {
		return true
	}

	// 定义合法的状态转换规则
	switch sm.current {
	case StateReady:
		// Ready -> Processing
		return next == StateProcessing

	case StateProcessing:
		// Processing -> Ready, Error
		return next == StateReady || next == StateError

	case StateError:
		// Error -> Ready
		return next == StateReady

	case StateQuitting:
		// Quitting 是终态，不能转换到其他状态
		return false

	default:
		return false
	}
}

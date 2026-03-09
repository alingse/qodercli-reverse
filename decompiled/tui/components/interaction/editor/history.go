package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

// HistoryHandler 历史记录处理器
type HistoryHandler struct {
	history   []string
	index     int
	isNavigating bool
}

// NewHistoryHandler 创建新的历史记录处理器
func NewHistoryHandler() *HistoryHandler {
	return &HistoryHandler{
		history: make([]string, 0),
		index:   -1,
	}
}

// Init 初始化
func (hh *HistoryHandler) Init() tea.Cmd {
	return nil
}

// AddHistory 添加历史记录
func (hh *HistoryHandler) AddHistory(input string) {
	if input == "" {
		return
	}
	// 避免重复添加相同的连续输入
	if len(hh.history) > 0 && hh.history[len(hh.history)-1] == input {
		return
	}
	hh.history = append(hh.history, input)
	hh.index = len(hh.history) // 重置索引到末尾
}

// Navigate 导航历史记录
// direction: -1 向上(上一个), 1 向下(下一个)
func (hh *HistoryHandler) Navigate(direction int) string {
	if len(hh.history) == 0 {
		return ""
	}

	newIndex := hh.index + direction
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > len(hh.history) {
		newIndex = len(hh.history)
	}

	hh.index = newIndex
	hh.isNavigating = true

	if hh.index >= len(hh.history) {
		return "" // 返回空表示恢复到当前输入
	}
	return hh.history[hh.index]
}

// ReloadHistory 重新加载历史记录（用于从文件恢复）
func (hh *HistoryHandler) ReloadHistory(history []string) {
	hh.history = history
	hh.index = len(history)
}

// GetHistory 获取所有历史记录
func (hh *HistoryHandler) GetHistory() []string {
	return hh.history
}

// ResetNavigation 重置导航状态
func (hh *HistoryHandler) ResetNavigation() {
	hh.index = len(hh.history)
	hh.isNavigating = false
}

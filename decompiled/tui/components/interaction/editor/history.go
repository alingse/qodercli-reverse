package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

// HistoryHandler history handler
type HistoryHandler struct {
	history      []string
	index        int
	isNavigating bool
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler() *HistoryHandler {
	return &HistoryHandler{
		history: make([]string, 0),
		index:   -1,
	}
}

// Init initializes
func (hh *HistoryHandler) Init() tea.Cmd {
	return nil
}

// AddHistory adds history record
func (hh *HistoryHandler) AddHistory(input string) {
	if input == "" {
		return
	}
	// avoid duplicate consecutive identical inputs
	if len(hh.history) > 0 && hh.history[len(hh.history)-1] == input {
		return
	}
	hh.history = append(hh.history, input)
	hh.index = len(hh.history) // reset index to end
}

// Navigate navigates history
// direction: -1 up (previous), 1 down (next)
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
		return "" // empty means restore to current input
	}
	return hh.history[hh.index]
}

// ReloadHistory reloads history (for restoring from file)
func (hh *HistoryHandler) ReloadHistory(history []string) {
	hh.history = history
	hh.index = len(history)
}

// GetHistory gets all history records
func (hh *HistoryHandler) GetHistory() []string {
	return hh.history
}

// ResetNavigation resets navigation state
func (hh *HistoryHandler) ResetNavigation() {
	hh.index = len(hh.history)
	hh.isNavigating = false
}

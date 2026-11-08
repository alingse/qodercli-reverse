package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AttachmentHandler attachment handler
type AttachmentHandler struct {
	attachments      []string
	highlightedIndex int
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler() *AttachmentHandler {
	return &AttachmentHandler{
		attachments:      make([]string, 0),
		highlightedIndex: -1,
	}
}

// AddAttachment adds an attachment
func (ah *AttachmentHandler) AddAttachment(path string) error {
	if ah.ValidateAttachment(path) {
		ah.attachments = append(ah.attachments, path)
		return nil
	}
	return fmt.Errorf("invalid attachment: %s", path)
}

// SetAttachments sets attachment list
func (ah *AttachmentHandler) SetAttachments(paths []string) {
	ah.attachments = paths
}

// GetAttachments gets attachment list
func (ah *AttachmentHandler) GetAttachments() []string {
	return ah.attachments
}

// HasAttachments whether has attachments
func (ah *AttachmentHandler) HasAttachments() bool {
	return len(ah.attachments) > 0
}

// DeleteAttachment deletes specified attachment
func (ah *AttachmentHandler) DeleteAttachment(index int) {
	if index >= 0 && index < len(ah.attachments) {
		ah.attachments = append(ah.attachments[:index], ah.attachments[index+1:]...)
		if ah.highlightedIndex >= len(ah.attachments) {
			ah.highlightedIndex = len(ah.attachments) - 1
		}
	}
}

// DeleteHighlightedAttachment deletes highlighted attachment
func (ah *AttachmentHandler) DeleteHighlightedAttachment() {
	if ah.highlightedIndex >= 0 {
		ah.DeleteAttachment(ah.highlightedIndex)
	}
}

// HasHighlightedAttachment whether has highlighted attachment
func (ah *AttachmentHandler) HasHighlightedAttachment() bool {
	return ah.highlightedIndex >= 0 && ah.highlightedIndex < len(ah.attachments)
}

// CycleHighlightAttachment cycles highlighted attachment
func (ah *AttachmentHandler) CycleHighlightAttachment() {
	if len(ah.attachments) == 0 {
		ah.highlightedIndex = -1
		return
	}
	ah.highlightedIndex = (ah.highlightedIndex + 1) % len(ah.attachments)
}

// Reset resets
func (ah *AttachmentHandler) Reset() {
	ah.attachments = make([]string, 0)
	ah.highlightedIndex = -1
}

// Render renders attachment list
func (ah *AttachmentHandler) Render() string {
	if !ah.HasAttachments() {
		return ""
	}

	var sb strings.Builder
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("75")).
		Italic(true)

	sb.WriteString("Attachments: ")
	for i, att := range ah.attachments {
		if i > 0 {
			sb.WriteString(", ")
		}
		if i == ah.highlightedIndex {
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")).
				Bold(true).
				Render(att))
		} else {
			sb.WriteString(style.Render(att))
		}
	}
	return sb.String()
}

// ValidateAttachment validates attachment
func (ah *AttachmentHandler) ValidateAttachment(path string) bool {
	return path != "" && !strings.Contains(path, "\x00")
}

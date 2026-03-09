package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// AttachmentHandler 附件处理器
type AttachmentHandler struct {
	attachments        []string
	highlightedIndex   int
}

// NewAttachmentHandler 创建新的附件处理器
func NewAttachmentHandler() *AttachmentHandler {
	return &AttachmentHandler{
		attachments:      make([]string, 0),
		highlightedIndex: -1,
	}
}

// AddAttachment 添加附件
func (ah *AttachmentHandler) AddAttachment(path string) error {
	if ah.ValidateAttachment(path) {
		ah.attachments = append(ah.attachments, path)
		return nil
	}
	return fmt.Errorf("invalid attachment: %s", path)
}

// SetAttachments 设置附件列表
func (ah *AttachmentHandler) SetAttachments(paths []string) {
	ah.attachments = paths
}

// GetAttachments 获取附件列表
func (ah *AttachmentHandler) GetAttachments() []string {
	return ah.attachments
}

// HasAttachments 是否有附件
func (ah *AttachmentHandler) HasAttachments() bool {
	return len(ah.attachments) > 0
}

// DeleteAttachment 删除指定附件
func (ah *AttachmentHandler) DeleteAttachment(index int) {
	if index >= 0 && index < len(ah.attachments) {
		ah.attachments = append(ah.attachments[:index], ah.attachments[index+1:]...)
		if ah.highlightedIndex >= len(ah.attachments) {
			ah.highlightedIndex = len(ah.attachments) - 1
		}
	}
}

// DeleteHighlightedAttachment 删除高亮的附件
func (ah *AttachmentHandler) DeleteHighlightedAttachment() {
	if ah.highlightedIndex >= 0 {
		ah.DeleteAttachment(ah.highlightedIndex)
	}
}

// HasHighlightedAttachment 是否有高亮的附件
func (ah *AttachmentHandler) HasHighlightedAttachment() bool {
	return ah.highlightedIndex >= 0 && ah.highlightedIndex < len(ah.attachments)
}

// CycleHighlightAttachment 循环高亮附件
func (ah *AttachmentHandler) CycleHighlightAttachment() {
	if len(ah.attachments) == 0 {
		ah.highlightedIndex = -1
		return
	}
	ah.highlightedIndex = (ah.highlightedIndex + 1) % len(ah.attachments)
}

// Reset 重置
func (ah *AttachmentHandler) Reset() {
	ah.attachments = make([]string, 0)
	ah.highlightedIndex = -1
}

// Render 渲染附件列表
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

// ValidateAttachment 验证附件
func (ah *AttachmentHandler) ValidateAttachment(path string) bool {
	return path != "" && !strings.Contains(path, "\x00")
}

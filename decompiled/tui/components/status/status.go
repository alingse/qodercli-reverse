// Package status status bar component
// displays current status, token usage, model info, etc.
package status

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alingse/qodercli-reverse/decompiled/core/types"
)

// color scheme - use standard ANSI color codes for better compatibility
var (
	statusColorIdle       = lipgloss.Color("82")  // idle - green
	statusColorConnecting = lipgloss.Color("215") // connecting - orange
	statusColorThinking   = lipgloss.Color("135") // thinking - purple
	statusColorTool       = lipgloss.Color("75")  // tool execution - blue
	statusColorError      = lipgloss.Color("203") // error - red
	statusColorBg         = lipgloss.Color("237") // status bar background - dark blue-gray
	statusColorText       = lipgloss.Color("251") // text - light gray
	statusColorMuted      = lipgloss.Color("248") // muted text - gray
	statusColorVim        = lipgloss.Color("215") // Vim mode - orange
)

// Status application status
type Status int

const (
	StatusIdle Status = iota
	StatusConnecting
	StatusThinking
	StatusToolExecuting
	StatusError
)

// Component status bar component
type Component struct {
	width       int
	height      int
	status      Status
	tokenUsage  *types.TokenUsage
	model       string
	mode        string
	message     string
	lastUpdated time.Time
}

// New creates a new status bar component
func New() *Component {
	return &Component{
		status:      StatusIdle,
		model:       "auto",
		mode:        "normal",
		lastUpdated: time.Now(),
	}
}

// Init initializes the component
func (c *Component) Init() tea.Cmd {
	return nil
}

// Update updates component state
func (c *Component) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return c, nil
}

// View renders the view
func (c *Component) View() string {
	if c.height <= 0 {
		c.height = 1
	}

	// left: status indicator
	leftContent := c.renderStatus()

	// center: model info
	centerContent := c.renderModelInfo()

	// right: token usage
	rightContent := c.renderTokenUsage()

	// compose status bar
	leftWidth := 20
	rightWidth := 25
	centerWidth := c.width - leftWidth - rightWidth - 4

	if centerWidth < 10 {
		centerWidth = 10
	}

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Align(lipgloss.Left)

	centerStyle := lipgloss.NewStyle().
		Width(centerWidth).
		Align(lipgloss.Center)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Align(lipgloss.Right)

	barStyle := lipgloss.NewStyle().
		Background(statusColorBg).
		Foreground(statusColorText).
		Width(c.width).
		Padding(0, 1)

	content := fmt.Sprintf("%s %s %s",
		leftStyle.Render(leftContent),
		centerStyle.Render(centerContent),
		rightStyle.Render(rightContent))

	return barStyle.Render(content)
}

// SetSize sets component size
func (c *Component) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetStatus sets status
func (c *Component) SetStatus(status Status) {
	c.status = status
	c.lastUpdated = time.Now()
}

// SetModel sets model
func (c *Component) SetModel(model string) {
	c.model = model
}

// SetMode sets mode
func (c *Component) SetMode(mode string) {
	c.mode = mode
}

// SetMessage sets status message
func (c *Component) SetMessage(msg string) {
	c.message = msg
}

// UpdateTokenUsage updates token usage
func (c *Component) UpdateTokenUsage(usage *types.TokenUsage) {
	c.tokenUsage = usage
	c.lastUpdated = time.Now()
}

// renderStatus renders status indicator
func (c *Component) renderStatus() string {
	switch c.status {
	case StatusIdle:
		return lipgloss.NewStyle().
			Foreground(statusColorIdle).
			Render("● Ready")

	case StatusConnecting:
		return lipgloss.NewStyle().
			Foreground(statusColorConnecting).
			Render("◐ Connecting...")

	case StatusThinking:
		return lipgloss.NewStyle().
			Foreground(statusColorThinking).
			Render("◐ Thinking...")

	case StatusToolExecuting:
		return lipgloss.NewStyle().
			Foreground(statusColorTool).
			Render("◐ Tool...")

	case StatusError:
		return lipgloss.NewStyle().
			Foreground(statusColorError).
			Render("● Error")

	default:
		return "Ready"
	}
}

// renderModelInfo renders model info
func (c *Component) renderModelInfo() string {
	var modeIndicator string
	if c.mode == "vim" {
		modeIndicator = lipgloss.NewStyle().
			Foreground(statusColorVim).
			Render("[VIM] ")
	}

	modelStyle := lipgloss.NewStyle().
		Foreground(statusColorMuted)

	return fmt.Sprintf("%s%s", modeIndicator, modelStyle.Render(c.model))
}

// renderTokenUsage renders token usage
func (c *Component) renderTokenUsage() string {
	if c.tokenUsage == nil {
		return ""
	}

	style := lipgloss.NewStyle().
		Foreground(statusColorMuted)

	total := c.tokenUsage.InputTokens + c.tokenUsage.OutputTokens
	return style.Render(fmt.Sprintf("Tokens: %d", total))
}

// GetStatus gets current status
func (c *Component) GetStatus() Status {
	return c.status
}

// IsProcessing whether currently processing
func (c *Component) IsProcessing() bool {
	return c.status == StatusThinking || c.status == StatusToolExecuting
}

package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nanocode/internal/mathutil"
	"nanocode/ui/theme"
)

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	if !m.hasScrollableContent() {
		m.layout.scrollbarDragging = false
		return m, cmd
	}
	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft && m.isOnScrollbar(msg.X, msg.Y) {
			m.layout.scrollbarDragging = true
			m.scrollToMouseY(msg.Y)
			return m, cmd
		}
	case tea.MouseActionMotion:
		if m.layout.scrollbarDragging {
			m.scrollToMouseY(msg.Y)
			return m, cmd
		}
	case tea.MouseActionRelease:
		m.layout.scrollbarDragging = false
	}
	return m, cmd
}

func (m Model) hasScrollableContent() bool {
	return m.viewport.TotalLineCount() > m.viewport.Height
}

const scrollbarRightOffset = 1

func (m Model) isOnScrollbar(x, y int) bool {
	if !m.hasScrollableContent() {
		return false
	}
	if y < m.layout.viewportTop || y >= m.layout.viewportTop+m.viewport.Height {
		return false
	}
	return x >= m.scrollbarColumn()
}

func (m *Model) scrollToMouseY(y int) {
	if !m.hasScrollableContent() || m.viewport.Height <= 1 {
		return
	}
	height := mathutil.Max(1, m.viewport.Height)
	trackY := mathutil.Clamp(y-m.layout.viewportTop, 0, height-1)
	maxOffset := mathutil.Max(0, m.viewport.TotalLineCount()-height)
	if maxOffset == 0 {
		return
	}
	target := trackY * maxOffset / mathutil.Max(1, height-1)
	m.viewport.SetYOffset(target)
}

func (m Model) viewportWithScrollbar() string {
	content := m.viewport.View()
	if !m.hasScrollableContent() {
		return content
	}
	height := mathutil.Max(1, m.viewport.Height)
	contentColumn := lipgloss.NewStyle().
		Width(m.viewport.Width).
		Height(height).
		MaxHeight(height).
		Render(content)
	scrollbarColumn := m.scrollbarColumnView(height)
	return lipgloss.JoinHorizontal(lipgloss.Top, contentColumn, scrollbarColumn)
}

func (m Model) scrollbarColumnView(height int) string {
	thumbSize := mathutil.Max(1, (height*height)/mathutil.Max(1, m.viewport.TotalLineCount()))
	maxThumbTop := mathutil.Max(0, height-thumbSize)
	maxOffset := mathutil.Max(1, m.viewport.TotalLineCount()-height)
	thumbTop := m.viewport.YOffset * maxThumbTop / maxOffset
	trackStyle := lipgloss.NewStyle().Foreground(theme.MutedText)
	thumbStyle := lipgloss.NewStyle().Foreground(theme.PrimaryAccent)
	rendered := make([]string, 0, height)
	for i := 0; i < height; i++ {
		bar := trackStyle.Render("│")
		if i >= thumbTop && i < thumbTop+thumbSize {
			bar = thumbStyle.Render("█")
		}
		rendered = append(rendered, bar)
	}
	return strings.Join(rendered, "\n")
}

func (m Model) scrollbarColumn() int {
	return m.layout.width - scrollbarRightOffset
}

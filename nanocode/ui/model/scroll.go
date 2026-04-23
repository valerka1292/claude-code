package model

import (
	"math"
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
	trackY := mathutil.Clamp(y-m.layout.viewportTop, 0, m.viewport.Height-1)
	maxOffset := mathutil.Max(0, m.viewport.TotalLineCount()-m.viewport.Height)
	target := int(math.Round(float64(trackY) / float64(m.viewport.Height-1) * float64(maxOffset)))
	m.viewport.SetYOffset(target)
}

func (m Model) viewportWithScrollbar() string {
	content := m.viewport.View()
	if !m.hasScrollableContent() {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) < m.viewport.Height {
		padding := make([]string, m.viewport.Height-len(lines))
		lines = append(lines, padding...)
	}

	thumbSize := mathutil.Max(1, (m.viewport.Height*m.viewport.Height)/mathutil.Max(1, m.viewport.TotalLineCount()))
	maxThumbTop := mathutil.Max(0, m.viewport.Height-thumbSize)
	maxOffset := mathutil.Max(1, m.viewport.TotalLineCount()-m.viewport.Height)
	thumbTop := int(math.Round(float64(m.viewport.YOffset) / float64(maxOffset) * float64(maxThumbTop)))

	trackStyle := lipgloss.NewStyle().Foreground(theme.MutedText)
	thumbStyle := lipgloss.NewStyle().Foreground(theme.PrimaryAccent)
	rendered := make([]string, 0, len(lines))
	for i, line := range lines {
		bar := trackStyle.Render("│")
		if i >= thumbTop && i < thumbTop+thumbSize {
			bar = thumbStyle.Render("█")
		}
		rendered = append(rendered, line+bar)
	}
	return strings.Join(rendered, "\n")
}

func (m Model) scrollbarColumn() int {
	return m.layout.width - scrollbarRightOffset
}

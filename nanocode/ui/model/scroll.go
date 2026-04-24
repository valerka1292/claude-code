package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nanocode/internal/mathutil"
	"nanocode/ui/theme"
)

var (
	scrollbarTrackStyle = lipgloss.NewStyle().Foreground(theme.MutedText)
	scrollbarThumbStyle = lipgloss.NewStyle().Foreground(theme.PrimaryAccent)
)

func (m Model) hasScrollableContent() bool {
	return m.viewport.TotalLineCount() > m.viewport.Height
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
		m.layout.scrollbarDragging = false
	}

	if m.shouldHandleScrollbarMouse(msg) {
		m.updateViewportFromScrollbar(msg.Y - m.layout.viewportTop)
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			m.layout.scrollbarDragging = true
		}
		return m, nil
	}

	if m.layout.scrollbarDragging && msg.Action == tea.MouseActionMotion {
		if msg.Y >= m.layout.viewportTop && msg.Y < m.layout.viewportTop+m.viewport.Height {
			m.updateViewportFromScrollbar(msg.Y - m.layout.viewportTop)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) shouldHandleScrollbarMouse(msg tea.MouseMsg) bool {
	if !m.hasScrollableContent() {
		return false
	}
	if msg.X != m.layout.width-1 {
		return false
	}
	if msg.Y < m.layout.viewportTop || msg.Y >= m.layout.viewportTop+m.viewport.Height {
		return false
	}
	if msg.Action == tea.MouseActionMotion {
		return m.layout.scrollbarDragging
	}
	return msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft
}

func (m *Model) updateViewportFromScrollbar(localY int) {
	trackTop, thumbHeight := scrollbarThumbGeometry(m.viewport.Height, m.viewport.TotalLineCount(), m.viewport.YOffset)
	maxOffset := mathutil.Max(0, m.viewport.TotalLineCount()-m.viewport.Height)
	if maxOffset == 0 {
		m.viewport.SetYOffset(0)
		return
	}

	rangeHeight := m.viewport.Height - thumbHeight
	if rangeHeight <= 0 {
		m.viewport.SetYOffset(0)
		return
	}

	targetTop := mathutil.Clamp(localY-thumbHeight/2, 0, rangeHeight)
	if localY >= trackTop && localY < trackTop+thumbHeight {
		targetTop = mathutil.Clamp(trackTop, 0, rangeHeight)
	}
	newOffset := int(float64(targetTop) / float64(rangeHeight) * float64(maxOffset))
	m.viewport.SetYOffset(mathutil.Clamp(newOffset, 0, maxOffset))
}

func (m Model) viewportWithScrollbar() string {
	content := m.viewport.View()
	lines := strings.Split(content, "\n")
	height := mathutil.Max(1, m.viewport.Height)
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}

	trackTop, thumbHeight := scrollbarThumbGeometry(height, m.viewport.TotalLineCount(), m.viewport.YOffset)
	for i := range lines {
		symbol := " "
		if m.hasScrollableContent() {
			symbol = scrollbarTrackStyle.Render("│")
			if i >= trackTop && i < trackTop+thumbHeight {
				symbol = scrollbarThumbStyle.Render("█")
			}
		}
		lines[i] += " " + symbol
	}

	return strings.Join(lines, "\n")
}

func scrollbarThumbGeometry(height int, totalLines int, yOffset int) (int, int) {
	if height <= 0 || totalLines <= height {
		return 0, mathutil.Max(1, height)
	}

	thumbHeight := mathutil.Clamp((height*height)/totalLines, 1, height)
	maxOffset := totalLines - height
	maxTrackTop := height - thumbHeight
	if maxOffset == 0 || maxTrackTop <= 0 {
		return 0, thumbHeight
	}

	clampedOffset := mathutil.Clamp(yOffset, 0, maxOffset)
	thumbTop := int(float64(clampedOffset) / float64(maxOffset) * float64(maxTrackTop))
	return thumbTop, thumbHeight
}

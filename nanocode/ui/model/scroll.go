package model

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) viewportWithScrollbar() string {
	return m.viewport.View()
}

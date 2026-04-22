package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/config"
)

func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.settings.open = false
		m.resizeViewport()
		return m, nil
	case "up":
		m.settings.selectedStyle = clamp(m.settings.selectedStyle-1, 0, 1)
		return m, nil
	case "down":
		m.settings.selectedStyle = clamp(m.settings.selectedStyle+1, 0, 1)
		return m, nil
	case "enter":
		style := spinnerStyleFor(m.settings.selectedStyle)
		m.settings.values.SpinnerStyle = style
		m.settings.open = false
		m.resizeViewport()
		m.refreshViewport(false)
		return m, func() tea.Msg {
			if err := config.SaveSettings(m.settings.values); err != nil {
				return spinnerChangedMsg{saved: false}
			}
			return spinnerChangedMsg{saved: true}
		}
	}
	return m, nil
}

func spinnerStyleFor(index int) string {
	if index == 1 {
		return config.SpinnerCircles
	}
	return config.SpinnerHexagons
}

func spinnerIndexFor(style string) int {
	if style == config.SpinnerCircles {
		return 1
	}
	return 0
}

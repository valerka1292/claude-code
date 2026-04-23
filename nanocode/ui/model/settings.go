package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/config"
)

var timeoutOptions = []int{30, 60, 90, 120, 180, 240, 300}

func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.settings.open = false
		m.resizeViewport()
		return m, nil
	case "up":
		m.settings.selectedRow = clamp(m.settings.selectedRow-1, 0, 1)
		return m, nil
	case "down":
		m.settings.selectedRow = clamp(m.settings.selectedRow+1, 0, 1)
		return m, nil
	case "left":
		m.shiftCurrentSetting(-1)
		return m, nil
	case "right":
		m.shiftCurrentSetting(1)
		return m, nil
	case "enter":
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

func (m *Model) shiftCurrentSetting(delta int) {
	switch m.settings.selectedRow {
	case 0:
		styles := []string{config.SpinnerHexagons, config.SpinnerCircles}
		idx := spinnerIndexFor(m.settings.values.SpinnerStyle)
		idx = clamp(idx+delta, 0, len(styles)-1)
		m.settings.values.SpinnerStyle = styles[idx]
	case 1:
		idx := timeoutIndexFor(m.settings.values.APITimeoutSeconds)
		idx = clamp(idx+delta, 0, len(timeoutOptions)-1)
		m.settings.values.APITimeoutSeconds = timeoutOptions[idx]
	}
}

func spinnerIndexFor(style string) int {
	if style == config.SpinnerCircles {
		return 1
	}
	return 0
}

func timeoutIndexFor(seconds int) int {
	for i, option := range timeoutOptions {
		if option == seconds {
			return i
		}
	}
	return 4
}

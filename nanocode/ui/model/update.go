package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/internal/mathutil"
	"nanocode/ui/components/nobby"
	"nanocode/ui/types"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout.width = msg.Width
		m.layout.height = msg.Height
		m.input.Width = mathutil.Max(10, msg.Width-6)
		m.providers.input.Width = mathutil.Max(30, msg.Width/2)
		m.resizeViewport()
		m.refreshViewport(false)
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case nobbyTickMsg:
		m.nobbyStep++
		return m, nobbyTickCmd(m.nobbyPose, m.nobbyStep)
	case spinnerChangedMsg:
		if msg.saved {
			m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: "Settings saved: spinner style updated."})
			m.refreshViewport(true)
		}
		return m, nil
	case providerSavedMsg:
		if msg.saved {
			m.providers.open = false
			m.providers.mode = providerModeMenu
			m.providers.input.Blur()
			m.reloadProviderNames()
			m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: msg.message, Timestamp: time.Now()})
			m.refreshViewport(true)
		}
		return m, nil
	case streamStartedMsg:
		m.stream.ch = msg.ch
		m.setNobbyPose(nobby.PoseReading)
		return m, pollStreamCmd(m.stream.ch)
	case streamEventMsg:
		return m.handleStreamEvent(msg)
	case tea.KeyMsg:
		if m.settings.open {
			return m.handleSettingsKeys(msg)
		}
		if m.providers.open {
			return m.handleProviderKeys(msg)
		}
		if nextModel, cmd, handled := m.handleKeyMsg(msg); handled {
			return nextModel, cmd
		}
	case spinnerTickMsg:
		if !m.chat.thinking || !m.chat.showInferring {
			return m, nil
		}
		m.chat.spinnerStep++
		m.refreshViewport(true)
		return m, spinnerTickCmd(m.settings.values.SpinnerStyle)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.updateCommandSuggestions()
	m.resizeViewport()
	m.refreshViewport(false)
	return m, cmd
}

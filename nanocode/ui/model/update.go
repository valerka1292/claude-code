package model

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/components/spinner"
	"nanocode/ui/types"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout.width = msg.Width
		m.layout.height = msg.Height
		m.input.Width = max(10, msg.Width-6)
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
			m.chat.messages = append(m.chat.messages, types.Message{
				Role: types.RoleAssistant,
				Text: "Settings saved: spinner style updated.",
			})
			m.refreshViewport(true)
		}
		return m, nil

	case tea.KeyMsg:
		if m.settings.open {
			return m.handleSettingsKeys(msg)
		}
		if nextModel, cmd, handled := m.handleKeyMsg(msg); handled {
			return nextModel, cmd
		}

	case spinnerTickMsg:
		if !m.chat.thinking {
			return m, nil
		}
		m.chat.spinnerStep++
		m.refreshViewport(true)
		return m, spinnerTickCmd(m.settings.values.SpinnerStyle)

	case assistantReplyMsg:
		if !m.chat.thinking {
			return m, nil
		}
		userText := ""
		for i := len(m.chat.messages) - 1; i >= 0; i-- {
			if m.chat.messages[i].Role == types.RoleUser {
				userText = m.chat.messages[i].Text
				break
			}
		}
		m.chat.messages = append(m.chat.messages, types.Message{
			Role:      types.RoleAssistant,
			Text:      fmt.Sprintf("Got it: %q. This is a mock nanocode response after a 2-second wait.", userText),
			Timestamp: time.Now(),
		})
		m.chat.thinking = false
		m.chat.spinnerStep = 0
		m.chat.spinnerVerb = ""
		m.refreshViewport(true)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.updateCommandSuggestions()
	m.resizeViewport()
	m.refreshViewport(false)
	return m, cmd
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if len(m.commands.suggestions) > 0 {
		switch msg.String() {
		case "up":
			m.commands.selected = clamp(m.commands.selected-1, 0, len(m.commands.suggestions)-1)
			return m, nil, true
		case "down":
			m.commands.selected = clamp(m.commands.selected+1, 0, len(m.commands.suggestions)-1)
			return m, nil, true
		case "tab":
			m.input.SetValue(m.commands.suggestions[m.commands.selected] + " ")
			m.clearCommandSuggestions()
			m.resizeViewport()
			return m, nil, true
		}
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit, true
	case "pgup":
		m.viewport.HalfViewUp()
		return m, nil, true
	case "pgdown":
		m.viewport.HalfViewDown()
		return m, nil, true
	case "up":
		if len(m.commands.suggestions) == 0 {
			m.viewport.LineUp(1)
			return m, nil, true
		}
	case "down":
		if len(m.commands.suggestions) == 0 {
			m.viewport.LineDown(1)
			return m, nil, true
		}
	case "enter":
		if m.chat.thinking {
			return m, nil, true
		}
		if len(m.commands.suggestions) > 0 {
			selected := m.commands.suggestions[m.commands.selected]
			m.input.SetValue(selected)
			m.clearCommandSuggestions()
			m.resizeViewport()
			nextModel, cmd := m.executeInput()
			return nextModel, cmd, true
		}
		nextModel, cmd := m.executeInput()
		return nextModel, cmd, true
	}

	return m, nil, false
}

func (m Model) executeInput() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}

	if text == "/settings" {
		m.settings.open = true
		m.settings.selectedStyle = spinnerIndexFor(m.settings.values.SpinnerStyle)
		m.input.SetValue("")
		m.clearCommandSuggestions()
		m.resizeViewport()
		return m, nil
	}

	m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleUser, Text: text, Timestamp: time.Now()})
	m.input.SetValue("")
	m.clearCommandSuggestions()
	m.chat.thinking = true
	m.chat.spinnerStep = 0
	m.chat.spinnerVerb = spinner.RandomVerb()
	m.resizeViewport()
	m.refreshViewport(true)
	return m, tea.Batch(spinnerTickCmd(m.settings.values.SpinnerStyle), mockReplyCmd())
}

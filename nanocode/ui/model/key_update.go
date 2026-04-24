package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/internal/mathutil"
	"nanocode/ui/components/nobby"
)

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if len(m.commands.suggestions) > 0 {
		switch msg.String() {
		case "up":
			m.commands.selected = mathutil.Clamp(m.commands.selected-1, 0, len(m.commands.suggestions)-1)
			return m, nil, true
		case "down":
			m.commands.selected = mathutil.Clamp(m.commands.selected+1, 0, len(m.commands.suggestions)-1)
			return m, nil, true
		case "tab":
			m.input.SetValue(m.commands.suggestions[m.commands.selected] + " ")
			m.clearCommandSuggestions()
			m.resizeViewport()
			return m, nil, true
		}
	}

	switch msg.String() {
	case "shift+tab":
		if m.chat.mode == ModeAsk {
			m.chat.mode = ModeCode
		} else {
			m.chat.mode = ModeAsk
		}
		m.resizeViewport()
		return m, nil, true
	case "ctrl+c":
		return m.handleCtrlC()
	case "esc":
		return m.handleEscKey()
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
			m.input.SetValue(m.commands.suggestions[m.commands.selected])
			m.clearCommandSuggestions()
		}
		nextModel, cmd := m.executeInput()
		return nextModel, cmd, true
	}
	return m, nil, false
}

func (m Model) handleCtrlC() (tea.Model, tea.Cmd, bool) {
	// Double-press Ctrl+C to quit.
	if m.isPendingConfirmationFor("ctrl+c") {
		m.clearPendingConfirmation()
		return m, tea.Quit, true
	}
	m.setPendingConfirmation("ctrl+c")
	return m, nil, true
}

func (m Model) handleEscKey() (tea.Model, tea.Cmd, bool) {
	// During thinking mode, ESC interrupts the stream with confirmation.
	if m.chat.thinking {
		if m.isPendingConfirmationFor("esc") {
			m.chat.interrupted = true
			m.chat.thinking = false
			m.chat.showInferring = false
			close(m.chat.abortChan)
			m.chat.abortChan = nil
			m.setNobbyPose(nobby.PoseIdle)
			if !m.chat.cycleStartedAt.IsZero() {
				m.chat.lastWorkedForSec = int(time.Since(m.chat.cycleStartedAt).Seconds())
			}
			m.clearPendingConfirmation()
			m.refreshViewport(true)
			return m, nil, true
		}
		m.setPendingConfirmation("esc")
		return m, nil, true
	}

	// Outside generation ESC should do nothing (reserved for canceling generation only).
	if m.chat.confirmKey == "esc" {
		m.clearPendingConfirmation()
	}
	return m, nil, true
}

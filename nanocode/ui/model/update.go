package model

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/internal/mathutil"
	"nanocode/ui/components/nobby"
	"nanocode/ui/types"
)

const confirmWindow = 800 * time.Millisecond

func (m *Model) isPendingConfirmationFor(key string) bool {
	if !m.chat.confirmPending || m.chat.confirmKey != key {
		return false
	}
	return time.Since(m.chat.confirmPressTime) <= confirmWindow
}

func (m *Model) setPendingConfirmation(key string) {
	m.chat.confirmPending = true
	m.chat.confirmKey = key
	m.chat.confirmPressTime = time.Now()
}

func (m *Model) clearPendingConfirmation() {
	m.chat.confirmPending = false
	m.chat.confirmKey = ""
	m.chat.confirmPressTime = time.Time{}
}

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
		if msg.done {
			m.setNobbyPose(nobby.PoseIdle)
			m.chat.thinking = false
			m.chat.spinnerVerb = ""
			m.chat.spinnerStep = 0
			m.chat.showInferring = false
			if !m.chat.cycleStartedAt.IsZero() {
				m.chat.lastWorkedForSec = int(time.Since(m.chat.cycleStartedAt).Seconds())
			}
			if strings.TrimSpace(m.chat.streamingText) != "" {
				m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: m.chat.streamingText, Timestamp: time.Now()})
			}
			m.chat.streamingText = ""
			m.chat.streamingThought = ""
			m.refreshViewport(true)
			return m, nil
		}
		if msg.event.ErrorText != "" {
			m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: "Error: " + msg.event.ErrorText, Timestamp: time.Now()})
			m.setNobbyPose(nobby.PoseAPIError)
			m.chat.thinking = false
			m.chat.showInferring = false
			m.chat.streamingText = ""
			m.chat.streamingThought = ""
			m.refreshViewport(true)
			return m, nil
		}
		if msg.event.ReconnectNote != "" {
			m.setNobbyPose(nobby.PoseAPIErrorReconnect)
			m.chat.messages = append(m.chat.messages, types.Message{
				Role:      types.RoleAssistant,
				Text:      msg.event.ReconnectNote,
				Timestamp: time.Now(),
			})
			m.refreshViewport(true)
			return m, pollStreamCmd(m.stream.ch)
		}
		if msg.event.ReasoningDelta != "" {
			m.setNobbyPose(nobby.PoseThinking)
			m.chat.streamingThought += msg.event.ReasoningDelta
			m.chat.estimatedTokensStream++
			m.chat.estimatedReasoningTokens++
		}
		if msg.event.ContentDelta != "" {
			m.setNobbyPose(nobby.PoseWriting)
			m.chat.showInferring = false
			m.chat.streamingText += msg.event.ContentDelta
			m.chat.estimatedTokensStream++
		}
		if msg.event.RefusalDelta != "" {
			m.chat.estimatedTokensStream++
		}
		if msg.event.ToolDelta != "" {
			m.chat.estimatedTokensStream++
		}
		liveEstimatedTotal := m.chat.usage.PromptTokens + m.chat.estimatedTokensStream
		if liveEstimatedTotal > m.chat.contextTokenFloor {
			m.chat.contextTokenFloor = liveEstimatedTotal
		}
		if msg.event.Usage != nil {
			m.chat.usage = *msg.event.Usage
			if m.chat.usage.TotalTokens > m.chat.contextTokenFloor {
				m.chat.contextTokenFloor = m.chat.usage.TotalTokens
			}
		}
		m.refreshViewport(true)
		return m, pollStreamCmd(m.stream.ch)
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
	case "ctrl+c":
		// Double-press Ctrl+C to quit.
		if m.isPendingConfirmationFor("ctrl+c") {
			m.clearPendingConfirmation()
			return m, tea.Quit, true
		}
		m.setPendingConfirmation("ctrl+c")
		return m, nil, true
	case "esc":
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

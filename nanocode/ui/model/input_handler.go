package model

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/components/nobby"
	"nanocode/ui/components/spinner"
	"nanocode/ui/config"
	"nanocode/ui/model/agent"
	"nanocode/ui/types"
)

func (m Model) executeInput() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}
	if text == "/settings" {
		m.settings.open = true
		m.settings.selectedRow = 0
		m.input.SetValue("")
		m.clearCommandSuggestions()
		m.resizeViewport()
		return m, nil
	}
	if text == "/provider" {
		m.openProviderPanel()
		m.input.SetValue("")
		m.clearCommandSuggestions()
		m.resizeViewport()
		return m, nil
	}

	active, ok := config.ActiveProvider(m.providers.data)
	if !ok {
		m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: "No active provider. Run /provider and create one first."})
		m.input.SetValue("")
		m.clearCommandSuggestions()
		m.refreshViewport(true)
		return m, nil
	}

	m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleUser, Text: text, Timestamp: time.Now()})
	m.input.SetValue("")
	m.clearCommandSuggestions()
	m.chat.thinking = true
	m.chat.spinnerStep = 0
	m.chat.spinnerVerb = spinner.RandomVerb()
	m.chat.streamingText = ""
	m.chat.streamingThought = ""
	m.chat.cycleStartedAt = time.Now()
	m.chat.estimatedTokensStream = 0
	m.chat.estimatedReasoningTokens = 0
	m.chat.showInferring = true
	m.chat.lastWorkedForSec = 0
	m.chat.interrupted = false
	m.clearPendingConfirmation()
	m.chat.abortChan = make(chan struct{})
	promptTokens := estimatePromptTokens(m.chat.messages, m.chat.mode)
	if promptTokens < m.chat.contextTokenFloor {
		promptTokens = m.chat.contextTokenFloor
	}
	if promptTokens > m.chat.contextTokenFloor {
		m.chat.contextTokenFloor = promptTokens
	}
	m.chat.usage = agent.UsageState{
		PromptTokens: promptTokens,
		TotalTokens:  promptTokens,
	}
	m.setNobbyPose(nobby.PoseReading)
	m.resizeViewport()
	m.refreshViewport(true)
	return m, tea.Batch(
		spinnerTickCmd(m.settings.values.SpinnerStyle),
		startAgentStreamCmd(active, m.chat.messages, m.settings.values, m.chat.abortChan, m.chat.mode),
	)
}

func (m *Model) setNobbyPose(pose nobby.Pose) {
	if m.nobbyPose == pose {
		return
	}
	m.nobbyPose = pose
	m.nobbyStep = 0
}

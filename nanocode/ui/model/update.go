package model

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/internal/mathutil"
	"nanocode/ui/components/nobby"
	"nanocode/ui/components/spinner"
	"nanocode/ui/config"
	"nanocode/ui/model/provider"
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
			m.chat.liveDownTokens += estimateTokens(msg.event.ReasoningDelta)
		}
		if msg.event.ContentDelta != "" {
			m.setNobbyPose(nobby.PoseWriting)
			m.chat.showInferring = false
			m.chat.streamingText += msg.event.ContentDelta
			m.chat.liveDownTokens += estimateTokens(msg.event.ContentDelta)
		}
		if msg.event.Usage != nil {
			m.chat.usage = *msg.event.Usage
			if msg.event.Usage.CompletionTokens > 0 {
				m.chat.liveDownTokens = msg.event.Usage.CompletionTokens
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
			m.input.SetValue(m.commands.suggestions[m.commands.selected])
			m.clearCommandSuggestions()
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
	m.chat.liveDownTokens = 0
	m.chat.showInferring = true
	m.chat.lastWorkedForSec = 0
	m.setNobbyPose(nobby.PoseReading)
	m.resizeViewport()
	m.refreshViewport(true)
	return m, tea.Batch(
		spinnerTickCmd(m.settings.values.SpinnerStyle),
		startAgentStreamCmd(active, m.chat.messages, m.settings.values),
	)
}

func (m *Model) setNobbyPose(pose nobby.Pose) {
	if m.nobbyPose == pose {
		return
	}
	m.nobbyPose = pose
	m.nobbyStep = 0
}

func (m Model) handleProviderKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.providers.mode {
	case providerModeInputValue, providerModeCreate, providerModeEditInput:
		if msg.String() == "esc" {
			m.providers.open = false
			m.providers.mode = providerModeMenu
			m.providers.input.Blur()
			m.resizeViewport()
			return m, nil
		}
		if msg.String() == "enter" {
			return m.submitProviderInput()
		}
		var cmd tea.Cmd
		m.providers.input, cmd = m.providers.input.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "esc":
		m.providers.open = false
		m.providers.mode = providerModeMenu
		m.providers.input.Blur()
		m.resizeViewport()
		return m, nil
	case "up":
		switch m.providers.mode {
		case providerModeMenu:
			m.providers.menuIndex = mathutil.Clamp(m.providers.menuIndex-1, 0, 3)
		case providerModeSelect, providerModeEditPick, providerModeDelete:
			m.providers.selectedProvider = mathutil.Clamp(m.providers.selectedProvider-1, 0, len(m.providers.names)-1)
		case providerModeEditField:
			m.providers.selectedField = provider.Field(mathutil.Clamp(int(m.providers.selectedField)-1, 0, 4))
		}
		return m, nil
	case "down":
		switch m.providers.mode {
		case providerModeMenu:
			m.providers.menuIndex = mathutil.Clamp(m.providers.menuIndex+1, 0, 3)
		case providerModeSelect, providerModeEditPick, providerModeDelete:
			m.providers.selectedProvider = mathutil.Clamp(m.providers.selectedProvider+1, 0, len(m.providers.names)-1)
		case providerModeEditField:
			m.providers.selectedField = provider.Field(mathutil.Clamp(int(m.providers.selectedField)+1, 0, 4))
		}
		return m, nil
	case "enter":
		return m.handleProviderSelect()
	}
	return m, nil
}

func (m Model) handleProviderSelect() (tea.Model, tea.Cmd) {
	switch m.providers.mode {
	case providerModeMenu:
		switch m.providers.menuIndex {
		case 0:
			m.beginProviderCreate()
		case 1:
			m.providers.mode = providerModeSelect
		case 2:
			m.providers.mode = providerModeEditPick
		case 3:
			m.providers.mode = providerModeDelete
		}
		return m, nil
	case providerModeSelect:
		if len(m.providers.names) == 0 {
			return m, nil
		}
		selected := m.providers.names[m.providers.selectedProvider]
		for name, provider := range m.providers.data.Providers {
			provider.Active = name == selected
			m.providers.data.Providers[name] = provider
		}
		return m, m.saveProvidersCmd(fmt.Sprintf("Active provider set to %s.", selected))
	case providerModeDelete:
		if len(m.providers.names) == 0 {
			return m, nil
		}
		selected := m.providers.names[m.providers.selectedProvider]
		delete(m.providers.data.Providers, selected)
		return m, m.saveProvidersCmd(fmt.Sprintf("Provider %s deleted.", selected))
	case providerModeEditPick:
		if len(m.providers.names) == 0 {
			return m, nil
		}
		m.providers.currentProviderRef = m.providers.names[m.providers.selectedProvider]
		m.providers.mode = providerModeEditField
		m.providers.selectedField = 0
		return m, nil
	case providerModeEditField:
		return m.startEditFieldInput()
	}
	return m, nil
}

func (m Model) startEditFieldInput() (tea.Model, tea.Cmd) {
	p := m.providers.data.Providers[m.providers.currentProviderRef]
	m.providers.mode = providerModeEditInput
	m.providers.input.Focus()
	switch m.providers.selectedField {
	case provider.FieldName:
		m.providers.inputPrompt = "Edit name"
		m.providers.inputField = provider.FieldName
		m.providers.input.SetValue(p.Name)
	case provider.FieldBaseURL:
		m.providers.inputPrompt = "Edit base URL"
		m.providers.inputField = provider.FieldBaseURL
		m.providers.input.SetValue(p.BaseURL)
	case provider.FieldModel:
		m.providers.inputPrompt = "Edit model"
		m.providers.inputField = provider.FieldModel
		m.providers.input.SetValue(p.Model)
	case provider.FieldAPIKey:
		m.providers.inputPrompt = "Edit API key"
		m.providers.inputField = provider.FieldAPIKey
		m.providers.input.SetValue(p.APIKey)
	case provider.FieldContextSize:
		m.providers.inputPrompt = "Edit context size"
		m.providers.inputField = provider.FieldContextSize
		m.providers.input.SetValue(fmt.Sprintf("%d", p.ContextSize))
	}
	return m, nil
}

func (m Model) submitProviderInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.providers.input.Value())
	switch m.providers.inputField {
	case provider.FieldName:
		if value == "" {
			return m, nil
		}
		if m.providers.mode == providerModeEditInput {
			p := m.providers.data.Providers[m.providers.currentProviderRef]
			delete(m.providers.data.Providers, m.providers.currentProviderRef)
			p.Name = value
			m.providers.data.Providers[value] = p
			m.providers.currentProviderRef = value
			return m, m.saveProvidersCmd("Provider name updated.")
		}
		m.providers.form.Name = value
		m.providers.inputField = provider.FieldBaseURL
		m.providers.inputPrompt = "Enter base URL"
		m.providers.input.SetValue("")
		return m, nil
	case provider.FieldBaseURL:
		if value == "" {
			return m, nil
		}
		if m.providers.mode == providerModeEditInput {
			p := m.providers.data.Providers[m.providers.currentProviderRef]
			p.BaseURL = config.NormalizeBaseURL(value)
			m.providers.data.Providers[m.providers.currentProviderRef] = p
			return m, m.saveProvidersCmd("Provider base URL updated.")
		}
		m.providers.form.BaseURL = config.NormalizeBaseURL(value)
		m.providers.inputField = provider.FieldModel
		m.providers.inputPrompt = "Enter model"
		m.providers.input.SetValue("")
		return m, nil
	case provider.FieldModel:
		if value == "" {
			return m, nil
		}
		if m.providers.mode == providerModeEditInput {
			p := m.providers.data.Providers[m.providers.currentProviderRef]
			p.Model = value
			m.providers.data.Providers[m.providers.currentProviderRef] = p
			return m, m.saveProvidersCmd("Provider model updated.")
		}
		m.providers.form.Model = value
		m.providers.inputField = provider.FieldAPIKey
		m.providers.inputPrompt = "Enter API key"
		m.providers.input.SetValue("")
		return m, nil
	case provider.FieldAPIKey:
		if value == "" {
			return m, nil
		}
		if m.providers.mode == providerModeEditInput {
			p := m.providers.data.Providers[m.providers.currentProviderRef]
			p.APIKey = value
			m.providers.data.Providers[m.providers.currentProviderRef] = p
			return m, m.saveProvidersCmd("Provider API key updated.")
		}
		m.providers.form.APIKey = value
		m.providers.inputField = provider.FieldContextSize
		m.providers.inputPrompt = "Enter context size"
		m.providers.input.SetValue("")
		return m, nil
	case provider.FieldContextSize:
		size, err := parseContextSize(value)
		if err != nil || size <= 0 {
			return m, nil
		}
		if m.providers.mode == providerModeEditInput {
			p := m.providers.data.Providers[m.providers.currentProviderRef]
			p.ContextSize = size
			m.providers.data.Providers[m.providers.currentProviderRef] = p
			return m, m.saveProvidersCmd("Provider context size updated.")
		}
		m.providers.form.ContextSize = value
		for name, provider := range m.providers.data.Providers {
			provider.Active = false
			m.providers.data.Providers[name] = provider
		}
		m.providers.data.Providers[m.providers.form.Name] = config.Provider{
			Name:        m.providers.form.Name,
			BaseURL:     m.providers.form.BaseURL,
			Model:       m.providers.form.Model,
			APIKey:      m.providers.form.APIKey,
			ContextSize: size,
			Active:      true,
		}
		return m, m.saveProvidersCmd(fmt.Sprintf("Provider %s created and activated.", m.providers.form.Name))
	}
	return m, nil
}

func (m Model) saveProvidersCmd(success string) tea.Cmd {
	return func() tea.Msg {
		if err := config.SaveProviders(m.providers.data); err != nil {
			return providerSavedMsg{saved: false, message: err.Error()}
		}
		return providerSavedMsg{saved: true, message: success}
	}
}

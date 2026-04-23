package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/internal/mathutil"
	"nanocode/ui/config"
	"nanocode/ui/model/provider"
)

const providerMenuItemCount = 4

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
			m.providers.menuIndex = mathutil.Clamp(m.providers.menuIndex-1, 0, providerMenuItemCount-1)
		case providerModeSelect, providerModeEditPick, providerModeDelete:
			m.providers.selectedProvider = mathutil.Clamp(m.providers.selectedProvider-1, 0, len(m.providers.names)-1)
		case providerModeEditField:
			m.providers.selectedField = provider.Field(mathutil.Clamp(int(m.providers.selectedField)-1, 0, providerFieldCount()-1))
		}
		return m, nil
	case "down":
		switch m.providers.mode {
		case providerModeMenu:
			m.providers.menuIndex = mathutil.Clamp(m.providers.menuIndex+1, 0, providerMenuItemCount-1)
		case providerModeSelect, providerModeEditPick, providerModeDelete:
			m.providers.selectedProvider = mathutil.Clamp(m.providers.selectedProvider+1, 0, len(m.providers.names)-1)
		case providerModeEditField:
			m.providers.selectedField = provider.Field(mathutil.Clamp(int(m.providers.selectedField)+1, 0, providerFieldCount()-1))
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
		for name, p := range m.providers.data.Providers {
			p.Active = name == selected
			m.providers.data.Providers[name] = p
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
		for name, p := range m.providers.data.Providers {
			p.Active = false
			m.providers.data.Providers[name] = p
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

package model

import (
	"nanocode/internal/mathutil"
	"nanocode/ui/config"
	"nanocode/ui/model/provider"
)

func (m *Model) reloadProviderNames() {
	m.providers.names = config.ProviderNames(m.providers.data)
	if len(m.providers.names) == 0 {
		m.providers.selectedProvider = 0
		return
	}
	m.providers.selectedProvider = mathutil.Clamp(m.providers.selectedProvider, 0, len(m.providers.names)-1)
}

func (m *Model) beginProviderCreate() {
	m.providers.mode = providerModeCreate
	m.providers.form.Reset()
	m.providers.inputPrompt = "Enter provider name"
	m.providers.input.SetValue("")
	m.providers.input.Focus()
	m.providers.inputField = provider.FieldName
}

func (m *Model) openProviderPanel() {
	m.providers.open = true
	m.providers.mode = providerModeMenu
	m.providers.menuIndex = 0
	m.reloadProviderNames()
}

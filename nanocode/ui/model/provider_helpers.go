package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"nanocode/ui/components/providers"
	"nanocode/ui/components/spinner"
	"nanocode/ui/config"
)

func (m *Model) reloadProviderNames() {
	m.providers.names = config.ProviderNames(m.providers.data)
	if len(m.providers.names) == 0 {
		m.providers.selectedProvider = 0
		return
	}
	m.providers.selectedProvider = clamp(m.providers.selectedProvider, 0, len(m.providers.names)-1)
}

func (m Model) activeProviderName() string {
	p, ok := config.ActiveProvider(m.providers.data)
	if !ok {
		return "No Provider"
	}
	return p.Name
}

func (m Model) activeModelName() string {
	p, ok := config.ActiveProvider(m.providers.data)
	if !ok {
		return "No Model"
	}
	return p.Model
}

func (m Model) usageLine() string {
	active, ok := config.ActiveProvider(m.providers.data)
	if !ok || active.ContextSize <= 0 {
		return ""
	}
	used := m.chat.usage.TotalTokens
	percent := (float64(used) / float64(active.ContextSize)) * 100
	return fmt.Sprintf("%s / %s (%.1f%% ctx)", formatCompact(used), formatCompact(active.ContextSize), percent)
}

func (m Model) footerStatusText() string {
	parts := make([]string, 0, 2)
	if usage := m.usageLine(); usage != "" {
		parts = append(parts, usage)
	}
	if status := m.agentStatusLine(); status != "" {
		parts = append(parts, status)
	}
	return strings.Join(parts, "  ")
}

func (m Model) agentStatusLine() string {
	if m.chat.showInferring && !m.chat.cycleStartedAt.IsZero() {
		elapsed := int(time.Since(m.chat.cycleStartedAt).Seconds())
		if elapsed < 0 {
			elapsed = 0
		}
		return fmt.Sprintf(
			"%s Inferring… (%ds · ↓ %s tokens · thinking)",
			spinner.Indicator(m.settings.values.SpinnerStyle, m.chat.spinnerStep),
			elapsed,
			formatCompact(m.chat.liveDownTokens),
		)
	}
	if m.chat.lastWorkedForSec > 0 {
		return fmt.Sprintf(
			"%s Worked for %ds",
			spinner.Indicator(m.settings.values.SpinnerStyle, m.chat.spinnerStep),
			m.chat.lastWorkedForSec,
		)
	}
	return ""
}

func formatCompact(value int) string {
	if value < 1000 {
		return fmt.Sprintf("%d", value)
	}
	return fmt.Sprintf("%.1fk", float64(value)/1000.0)
}

func (m Model) providerPanelViewData() (string, string, []string, int, string) {
	switch m.providers.mode {
	case providerModeCreate, providerModeInputValue:
		return "Provider Setup", m.providers.inputPrompt, nil, 0, m.providers.input.View()
	case providerModeSelect:
		options := m.providerOptions(true)
		return "Set Active Provider", "Choose active provider", options, m.providers.selectedProvider, ""
	case providerModeDelete:
		options := m.providerOptions(false)
		return "Delete Provider", "Choose provider to delete", options, m.providers.selectedProvider, ""
	case providerModeEditPick:
		options := m.providerOptions(false)
		return "Edit Provider", "Choose provider to edit", options, m.providers.selectedProvider, ""
	case providerModeEditField:
		fields := []string{"Name", "Base URL", "Model", "API Key", "Context Size"}
		return "Edit Field", "Choose field to change", fields, m.providers.selectedField, ""
	case providerModeEditInput:
		return "Edit Value", m.providers.inputPrompt, nil, 0, m.providers.input.View()
	default:
		options := []string{"Create provider", "Set active provider", "Edit provider", "Delete provider"}
		return "Providers", "OpenAI-compatible providers", options, m.providers.menuIndex, ""
	}
}

func (m Model) providerOptions(withActive bool) []string {
	out := make([]string, 0, len(m.providers.names))
	for _, name := range m.providers.names {
		p := m.providers.data.Providers[name]
		if withActive {
			out = append(out, providers.ProviderSummary(name, p.Model, p.ContextSize, p.Active))
		} else {
			out = append(out, providers.ProviderSummary(name, p.Model, p.ContextSize, false))
		}
	}
	if len(out) == 0 {
		out = []string{"No providers configured"}
	}
	return out
}

func (m *Model) beginProviderCreate() {
	m.providers.mode = providerModeCreate
	m.providers.formName = ""
	m.providers.formBaseURL = ""
	m.providers.formModel = ""
	m.providers.formAPIKey = ""
	m.providers.formContextSize = ""
	m.providers.inputPrompt = "Enter provider name"
	m.providers.input.SetValue("")
	m.providers.input.Focus()
	m.providers.inputField = "name"
}

func (m *Model) openProviderPanel() {
	m.providers.open = true
	m.providers.mode = providerModeMenu
	m.providers.menuIndex = 0
	m.reloadProviderNames()
}

func parseContextSize(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	return strconv.Atoi(value)
}

package model

import (
	"fmt"
	"time"

	"nanocode/ui/components/providers"
	"nanocode/ui/components/spinner"
	"nanocode/ui/config"
)

var providerMenuOptions = []string{
	"Create provider",
	"Set active provider",
	"Edit provider",
	"Delete provider",
}

var providerEditFieldLabels = []string{
	"Name",
	"Base URL",
	"Model",
	"API Key",
	"Context Size",
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

	// Если есть точные данные от API - используем их
	if m.chat.usage.TotalTokens > 0 && m.chat.usage.CompletionTokens > 0 {
		percent := (float64(m.chat.usage.TotalTokens) / float64(active.ContextSize)) * 100
		return fmt.Sprintf("%s / %s (%.2f%% ctx)",
			formatCompact(m.chat.usage.TotalTokens),
			formatCompact(active.ContextSize),
			percent)
	}

	// До получения usage показываем оценку
	estimated := m.chat.usage.PromptTokens + m.chat.estimatedTokensStream
	if estimated == 0 {
		return fmt.Sprintf("0 / %s (0.00%% ctx)", formatCompact(active.ContextSize))
	}
	percent := (float64(estimated) / float64(active.ContextSize)) * 100
	return fmt.Sprintf("~%s / %s (%.2f%% ctx)",
		formatCompact(estimated),
		formatCompact(active.ContextSize),
		percent)
}

func (m Model) agentStatusLine() string {
	if m.chat.interrupted {
		return fmt.Sprintf(
			"%s Interrupted",
			spinner.StaticIndicator(m.settings.values.SpinnerStyle),
		)
	}
	if m.chat.showInferring && !m.chat.cycleStartedAt.IsZero() {
		elapsed := int(time.Since(m.chat.cycleStartedAt).Seconds())
		if elapsed < 0 {
			elapsed = 0
		}

		// Показываем reasoning токены если доступны
		thinkingTokens := m.chat.usage.ReasoningTokens
		thinkingLabel := ""

		if thinkingTokens > 0 {
			// Есть точные reasoning токены от API
			thinkingLabel = fmt.Sprintf(" · ↓ %s tokens · thinking", formatCompact(thinkingTokens))
		} else if m.chat.estimatedTokensStream > 0 {
			// Пока нет точных данных - показываем оценку
			thinkingLabel = fmt.Sprintf(" · ↓ %s tokens · thinking", formatCompact(m.chat.estimatedTokensStream))
		}

		durationStr := formatDuration(int(time.Since(m.chat.cycleStartedAt).Milliseconds()))

		return fmt.Sprintf(
			"%s %s... (**esc to interrupt** · %s%s)",
			spinner.Indicator(m.settings.values.SpinnerStyle, m.chat.spinnerStep),
			m.chat.spinnerVerb,
			durationStr,
			thinkingLabel,
		)
	}
	if m.chat.lastWorkedForSec > 0 {
		return fmt.Sprintf(
			"%s Worked for %ds",
			spinner.StaticIndicator(m.settings.values.SpinnerStyle),
			m.chat.lastWorkedForSec,
		)
	}
	return ""
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
		return "Edit Field", "Choose field to change", providerEditFieldLabels, int(m.providers.selectedField), ""
	case providerModeEditInput:
		return "Edit Value", m.providers.inputPrompt, nil, 0, m.providers.input.View()
	default:
		return "Providers", "OpenAI-compatible providers", providerMenuOptions, m.providers.menuIndex, ""
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

package model

import (
	"nanocode/internal/mathutil"

	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/components/header"
	"nanocode/ui/components/nobby"
	"nanocode/ui/components/prompt"
	"nanocode/ui/components/providers"
	"nanocode/ui/components/settings"
	"nanocode/ui/components/suggestions"
)

func (m *Model) resizeViewport() {
	if m.layout.width == 0 || m.layout.height == 0 {
		return
	}
	headerHeight := lipgloss.Height(header.View(m.cwd, nobby.Render(m.nobbyPose, m.nobbyStep), m.activeProviderName(), m.activeModelName()))
	inputHeight := lipgloss.Height(prompt.InputBar(m.input.View(), m.layout.width))
	footerHeight := lipgloss.Height(prompt.Footer(m.layout.width, m.usageLine(), m.confirmationHint()))
	reserved := headerHeight + inputHeight + footerHeight + 3
	if len(m.commands.suggestions) > 0 {
		reserved += lipgloss.Height(suggestions.CommandList(m.layout.width, m.commands.suggestions, m.commands.selected))
	}
	if m.settings.open {
		reserved += lipgloss.Height(settings.Panel(
			m.layout.width,
			m.settings.selectedRow,
			m.settings.values.SpinnerStyle,
			m.settings.values.APITimeoutSeconds,
		))
	}
	if m.providers.open {
		title, desc, options, selected, inputView := m.providerPanelViewData()
		reserved += lipgloss.Height(providers.Panel(m.layout.width, title, desc, options, selected, inputView))
	}
	vHeight := m.layout.height - reserved
	if vHeight < 6 {
		vHeight = 6
	}
	m.layout.viewportTop = headerHeight + 1
	m.layout.viewportMaxHeight = vHeight
	m.viewport.Width = mathutil.Max(10, m.layout.width-1)
	m.viewport.Height = vHeight
}

package model

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nanocode/internal/mathutil"
	"nanocode/ui/components/header"
	"nanocode/ui/components/messages"
	"nanocode/ui/components/nobby"
	"nanocode/ui/components/prompt"
	"nanocode/ui/components/providers"
	"nanocode/ui/components/settings"
	"nanocode/ui/components/spinner"
	"nanocode/ui/components/suggestions"
)

func (m *Model) refreshViewport(forceBottom bool) {
	if m.layout.width == 0 {
		return
	}
	spinnerLine := m.agentStatusLine()
	wasBottom := m.viewport.AtBottom()
	content := messages.View(m.chat.messages, m.viewport.Width, spinnerLine, "", m.chat.streamingText)
	m.viewport.Height = mathutil.Max(1, m.layout.viewportMaxHeight)
	m.viewport.SetContent(content)
	m.viewport.SetYOffset(m.viewport.YOffset)
	if forceBottom || wasBottom {
		m.viewport.GotoBottom()
	}
}

func (m Model) View() string {
	if m.layout.width == 0 || m.layout.height == 0 {
		return "Loading..."
	}

	nobbyView := nobby.Render(m.nobbyPose, m.nobbyStep)
	headerView := header.View(m.layout.width, m.cwd, nobbyView, m.activeProviderName(), m.activeModelName())
	inputView := prompt.InputBar(m.input.View(), m.layout.width)
	footerView := prompt.Footer(m.layout.width, m.usageLine(), m.confirmationHint(), string(m.chat.mode))
	headerHeight := lipgloss.Height(headerView)
	inputHeight := lipgloss.Height(inputView)
	footerHeight := lipgloss.Height(footerView)
	totalReserved := headerHeight + inputHeight + footerHeight + 1
	debugInfo := fmt.Sprintf(
		"RAW W:%d H:%d | HeaderH:%d InputH:%d FooterH:%d | ViewportW:%d ViewportH:%d ViewTop:%d | ScrollbarX:%d | Reserved:%d",
		m.layout.width,
		m.layout.height,
		headerHeight,
		inputHeight,
		footerHeight,
		m.viewport.Width,
		m.viewport.Height,
		m.layout.viewportTop,
		m.scrollbarColumn(),
		totalReserved,
	)
	parts := []string{headerView, debugInfo, m.viewportWithScrollbar(), inputView}

	if len(m.commands.suggestions) > 0 {
		suggestionsView := suggestions.CommandList(m.layout.width, m.commands.suggestions, m.commands.selected)
		totalReserved += lipgloss.Height(suggestionsView)
		parts = append(parts, suggestionsView)
	}
	if m.settings.open {
		settingsView := settings.Panel(
			m.layout.width,
			m.settings.selectedRow,
			m.settings.values.SpinnerStyle,
			m.settings.values.APITimeoutSeconds,
		)
		totalReserved += lipgloss.Height(settingsView)
		parts = append(parts, settingsView)
	}
	if m.providers.open {
		title, desc, options, selected, inputView := m.providerPanelViewData()
		providersView := providers.Panel(m.layout.width, title, desc, options, selected, inputView)
		totalReserved += lipgloss.Height(providersView)
		parts = append(parts, providersView)
	}

	debugInfo = fmt.Sprintf(
		"%s | Reserved+Panels:%d | V+R:%d",
		debugInfo,
		totalReserved,
		m.viewport.Height+totalReserved,
	)
	parts[1] = debugInfo
	parts = append(parts, footerView)
	root := lipgloss.NewStyle().
		Width(m.layout.width).
		Height(m.layout.height)
	return root.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func nobbyTickCmd(pose nobby.Pose, step int) tea.Cmd {
	return tea.Tick(nobby.DurationFor(pose, step), func(t time.Time) tea.Msg {
		return nobbyTickMsg(t)
	})
}

func spinnerTickCmd(style string) tea.Cmd {
	return tea.Tick(spinner.Interval(style), func(t time.Time) tea.Msg {
		return spinnerTickMsg(t)
	})
}

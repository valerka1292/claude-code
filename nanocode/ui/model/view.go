package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/components/header"
	"nanocode/ui/components/messages"
	"nanocode/ui/components/nobby"
	"nanocode/ui/components/prompt"
	"nanocode/ui/components/providers"
	"nanocode/ui/components/settings"
	"nanocode/ui/components/spinner"
	"nanocode/ui/components/suggestions"
	"nanocode/ui/theme"
)

func (m *Model) refreshViewport(forceBottom bool) {
	if m.layout.width == 0 {
		return
	}
	spinnerLine := m.agentStatusLine()
	wasBottom := m.viewport.AtBottom()
	content := messages.View(m.chat.messages, m.viewport.Width, spinnerLine, "", m.chat.streamingText)
	m.viewport.SetContent(content)
	targetHeight := min(max(1, m.viewport.TotalLineCount()), m.layout.viewportMaxHeight)
	if targetHeight < 1 {
		targetHeight = 1
	}
	m.viewport.Height = targetHeight
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
	headerView := header.View(m.cwd, nobbyView, m.activeProviderName(), m.activeModelName())
	inputView := prompt.InputBar(m.input.View(), m.layout.width)
	parts := []string{headerView, "", m.viewportWithScrollbar(), inputView}

	if len(m.commands.suggestions) > 0 {
		parts = append(parts, suggestions.CommandList(m.layout.width, m.commands.suggestions, m.commands.selected))
	}
	if m.settings.open {
		parts = append(parts, settings.Panel(
			m.layout.width,
			m.settings.selectedRow,
			m.settings.values.SpinnerStyle,
			m.settings.values.APITimeoutSeconds,
		))
	}
	if m.providers.open {
		title, desc, options, selected, inputView := m.providerPanelViewData()
		parts = append(parts, providers.Panel(m.layout.width, title, desc, options, selected, inputView))
	}

	parts = append(parts, prompt.Footer(m.layout.width, m.footerStatusText()))
	root := lipgloss.NewStyle().Background(theme.AppBackground).Foreground(theme.PrimaryText)
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

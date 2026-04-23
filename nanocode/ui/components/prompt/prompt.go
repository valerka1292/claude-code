package prompt

import (
	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/theme"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, true, false).
			BorderForeground(theme.MutedText).
			Foreground(theme.PrimaryText).
			Background(theme.AppBackground)
	footerStyle = lipgloss.NewStyle().Foreground(theme.SecondaryAccent).PaddingLeft(1)
)

func InputBar(text string, width int) string {
	return boxStyle.Width(width).Render("❯ " + text)
}

func Footer(width int, usage string) string {
	left := footerStyle.Render("⏵⏵ accept edits on (shift+tab to cycle)")
	right := footerStyle.Foreground(theme.MutedText).Render(usage)
	if usage == "" {
		return left
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Top, left, lipgloss.NewStyle().Width(max(0, width-lipgloss.Width(left)-lipgloss.Width(right))).Render(""), right))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

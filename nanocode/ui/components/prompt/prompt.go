package prompt

import (
	"nanocode/internal/mathutil"
	"nanocode/ui/theme"

	"github.com/charmbracelet/lipgloss"
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

func Footer(width int, usage string, hint string) string {
	leftText := "⏵⏵ accept edits on (shift+tab to cycle)"
	if hint != "" {
		leftText = hint
	}
	left := footerStyle.Render(leftText)
	right := footerStyle.Foreground(theme.MutedText).Render(usage)
	if usage == "" {
		return left
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Top, left, lipgloss.NewStyle().Width(mathutil.Max(0, width-lipgloss.Width(left)-lipgloss.Width(right))).Render(""), right))
}

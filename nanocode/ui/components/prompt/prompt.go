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

func Footer() string {
	return footerStyle.Render("⏵⏵ accept edits on (shift+tab to cycle)")
}

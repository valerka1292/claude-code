package prompt

import "github.com/charmbracelet/lipgloss"

var (
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, true, false).BorderForeground(lipgloss.Color("240"))
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).PaddingLeft(1)
)

func InputBar(text string, width int) string {
	return boxStyle.Width(width).Render("❯ " + text)
}

func Footer() string {
	return footerStyle.Render("⏵⏵ accept edits on (shift+tab to cycle)")
}

package header

import (
	"nanocode/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryText)
	metaStyle  = lipgloss.NewStyle().Foreground(theme.MutedText)
	noteStyle  = lipgloss.NewStyle().Foreground(theme.SecondaryAccent).Bold(true)
)

func View(cwd string, mascot string) string {

	infoRows := []string{
		titleStyle.Render("nanocode v0.0.1"),
		metaStyle.Render("Mock Model · API Usage Billing"),
		metaStyle.Render(cwd),
	}
	info := lipgloss.NewStyle().PaddingLeft(1).Render(lipgloss.JoinVertical(lipgloss.Left, infoRows...))

	topSection := lipgloss.JoinHorizontal(lipgloss.Top, mascot, info)

	note := noteStyle.Render("↑ Fastest coding agent· 5x faster than other agents, 5x cheaper!")

	content := lipgloss.JoinVertical(lipgloss.Left, topSection, "", note)

	wrap := lipgloss.NewStyle().Background(theme.AppBackground).Padding(0, 1)
	return wrap.Render(content)
}

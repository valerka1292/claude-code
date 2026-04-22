package header

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/theme"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryText)
	metaStyle  = lipgloss.NewStyle().Foreground(theme.MutedText)
	noteStyle  = lipgloss.NewStyle().Foreground(theme.SecondaryAccent).Bold(true)
)

func View(cwd string, mascot string) string {
	rows := []string{
		titleStyle.Render("nanocode v0.0.1"),
		metaStyle.Render("Mock Model · API Usage Billing"),
		metaStyle.Render(cwd),
		"",
		noteStyle.Render("↑ Fastest coding agent· 5x faster than other agents, 5x cheaper!"),
	}

	info := lipgloss.NewStyle().PaddingLeft(1).Render(fmt.Sprintf("%s", lipgloss.JoinVertical(lipgloss.Left, rows...)))

	wrap := lipgloss.NewStyle().Background(theme.AppBackground).Padding(0, 1)
	return wrap.Render(lipgloss.JoinHorizontal(lipgloss.Top, mascot, info))
}

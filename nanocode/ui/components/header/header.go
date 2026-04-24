package header

import (
	"nanocode/ui/theme"
	"nanocode/version"

	"github.com/charmbracelet/lipgloss"
)

var (
	nameStyle    = lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryText)
	versionStyle = lipgloss.NewStyle().Foreground(theme.MutedText)
	metaStyle    = lipgloss.NewStyle().Foreground(theme.MutedText)
	noteStyle    = lipgloss.NewStyle().Foreground(theme.SecondaryAccent).Bold(true)
)

func View(width int, cwd string, mascot string, providerName string, modelName string) string {
	title := lipgloss.JoinHorizontal(
		lipgloss.Left,
		nameStyle.Render(version.Name),
		" ",
		versionStyle.Render("v"+version.Current),
	)

	infoRows := []string{
		title,
		metaStyle.Render(modelName + " · " + providerName),
		metaStyle.Render(cwd),
	}
	info := lipgloss.NewStyle().PaddingLeft(1).Render(lipgloss.JoinVertical(lipgloss.Left, infoRows...))

	topSection := lipgloss.JoinHorizontal(lipgloss.Top, mascot, info)

	note := noteStyle.Render("↑ Fastest coding agent · 5x faster than other agents, 5x cheaper!")

	content := lipgloss.JoinVertical(lipgloss.Left, topSection, "", note)

	wrap := lipgloss.NewStyle().Background(theme.AppBackground).Padding(0, 1).Width(width)
	return wrap.Render(content)
}

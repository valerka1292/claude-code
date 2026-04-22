package header

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	logoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("209"))
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	metaStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	noteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
)

func View(cwd string) string {
	art := []string{
		" ▐▛███▜▌   ",
		"▝▜█████▛▘  ",
		"  ▘▘ ▝▝    ",
	}

	rows := []string{
		logoStyle.Render(art[0]) + titleStyle.Render("nanocode v0.0.1"),
		logoStyle.Render(art[1]) + metaStyle.Render("Mock Model · API Usage Billing"),
		logoStyle.Render(art[2]) + metaStyle.Render(cwd),
		"",
		noteStyle.Render(" ↑ Fastest coding agent· 5x faster than other agents, 5x cheaper!"),
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(fmt.Sprintf("%s", lipgloss.JoinVertical(lipgloss.Left, rows...)))
}

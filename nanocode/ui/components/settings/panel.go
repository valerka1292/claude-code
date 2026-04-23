package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/config"
	"nanocode/ui/theme"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(theme.SecondaryAccent).
			Padding(1, 2).
			Background(theme.AppBackground)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryAccent)
	mutedText  = lipgloss.NewStyle().Foreground(theme.MutedText)
	selected   = lipgloss.NewStyle().Foreground(theme.AccentContrast).Background(theme.PrimaryAccent)
)

func Panel(width int, selectedRow int, spinnerStyle string, timeoutSeconds int) string {
	spinnerLabel := "Hexagons"
	if spinnerStyle == config.SpinnerCircles {
		spinnerLabel = "Circles"
	}

	rows := []string{
		titleStyle.Render("Settings"),
		mutedText.Render("↑/↓ select setting • ←/→ change value • Enter save • Esc close"),
		"",
	}

	items := []string{
		"Spinner style:   < " + spinnerLabel + " >",
		"API timeout:     < " + formatTimeout(timeoutSeconds) + " >",
	}
	for i, line := range items {
		if i == selectedRow {
			rows = append(rows, selected.Render(line))
			continue
		}
		rows = append(rows, line)
	}

	return boxStyle.Width(max(56, width*2/3)).Render(strings.Join(rows, "\n"))
}

func formatTimeout(seconds int) string {
	return fmt.Sprintf("%ds", seconds)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package settings

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func Panel(width int, selectedIndex int, current string) string {
	options := []string{"Hexagons", "Circles"}
	keys := []string{"hexagons", "circles"}
	rows := []string{
		titleStyle.Render("Settings"),
		mutedText.Render("Choose spinner animation style for thinking state:"),
		"",
	}
	for i, name := range options {
		prefix := "  "
		if keys[i] == current {
			prefix = "✓ "
		}
		line := prefix + name
		if i == selectedIndex {
			line = selected.Render(line)
		}
		rows = append(rows, line)
	}
	rows = append(rows, "", mutedText.Render("↑/↓ move • Enter save • Esc close"))
	return boxStyle.Width(max(48, width*2/3)).Render(strings.Join(rows, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

package providers

import (
	"fmt"
	"strings"

	"nanocode/internal/mathutil"
	"nanocode/ui/theme"

	"github.com/charmbracelet/lipgloss"
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

func Panel(width int, title string, description string, options []string, selectedIndex int, inputView string) string {
	rows := []string{
		titleStyle.Render(title),
		mutedText.Render(description),
		"",
	}

	for i, option := range options {
		line := "  " + option
		if i == selectedIndex {
			line = selected.Render(line)
		}
		rows = append(rows, line)
	}

	if inputView != "" {
		rows = append(rows, "", mutedText.Render("Input:"), inputView)
	}

	rows = append(rows, "", mutedText.Render("↑/↓ move • Enter select • Esc close"))
	return boxStyle.Width(mathutil.Max(62, width*3/4)).Render(strings.Join(rows, "\n"))
}

func ProviderSummary(name string, model string, contextSize int, active bool) string {
	activeMark := " "
	if active {
		activeMark = "✓"
	}
	return fmt.Sprintf("%s %s · %s · ctx %d", activeMark, name, model, contextSize)
}

package prompt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, true, false).
			BorderForeground(lipgloss.Color("240"))
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).PaddingLeft(1)

	suggestionBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)
	selectedSuggestion = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))

	settingsBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(0, 1)
	settingsTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
)

func InputBar(text string, width int) string {
	return boxStyle.Width(width).Render("❯ " + text)
}

func Footer() string {
	return footerStyle.Render("⏵⏵ accept edits on (shift+tab to cycle)")
}

func CommandSuggestions(width int, items []string, selected int) string {
	if len(items) == 0 {
		return ""
	}
	rows := make([]string, 0, len(items)+1)
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Commands"))
	for i, item := range items {
		line := fmt.Sprintf("%s", item)
		if i == selected {
			line = selectedSuggestion.Render(line)
		}
		rows = append(rows, line)
	}
	return suggestionBox.Width(max(24, width/3)).Render(strings.Join(rows, "\n"))
}

func SettingsPanel(width int, selected int, current string) string {
	options := []string{"Hexagons", "Circles"}
	keys := []string{"hexagons", "circles"}
	rows := []string{settingsTitle.Render("Settings"), "Spinner style"}
	for i, name := range options {
		prefix := "  "
		if keys[i] == current {
			prefix = "✓ "
		}
		line := prefix + name
		if i == selected {
			line = selectedSuggestion.Render(line)
		}
		rows = append(rows, line)
	}
	rows = append(rows, "", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("↑/↓ choose • Enter save • Esc close"))
	return settingsBox.Width(max(34, width/2)).Render(strings.Join(rows, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

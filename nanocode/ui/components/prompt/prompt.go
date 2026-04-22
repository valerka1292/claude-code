package prompt

import (
	"fmt"
	"strings"

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

	suggestionBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(theme.MutedText).
			Padding(0, 1).
			Background(theme.AppBackground)
	selectedSuggestion = lipgloss.NewStyle().Foreground(theme.AccentContrast).Background(theme.PrimaryAccent)
	mutedText          = lipgloss.NewStyle().Foreground(theme.MutedText)

	settingsBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(theme.SecondaryAccent).
			Padding(1, 2).
			Background(theme.AppBackground)
	settingsTitle = lipgloss.NewStyle().Bold(true).Foreground(theme.PrimaryAccent)
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
	rows := make([]string, 0, len(items))
	for i, item := range items {
		line := fmt.Sprintf(" %s", item)
		if i == selected {
			line = selectedSuggestion.Render(line)
		}
		rows = append(rows, line)
	}
	contentWidth := max(18, min(width-4, maxItemWidth(items)+4))
	return suggestionBox.Width(contentWidth).Render(strings.Join(rows, "\n"))
}

func SettingsPanel(width int, selected int, current string) string {
	options := []string{"Hexagons", "Circles"}
	keys := []string{"hexagons", "circles"}
	rows := []string{
		settingsTitle.Render("Settings"),
		mutedText.Render("Choose spinner animation style for thinking state:"),
		"",
	}
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
	rows = append(rows, "", mutedText.Render("↑/↓ move • Enter save • Esc close"))
	return settingsBox.Width(max(48, width*2/3)).Render(strings.Join(rows, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxItemWidth(items []string) int {
	w := 0
	for _, item := range items {
		if len(item) > w {
			w = len(item)
		}
	}
	return w
}

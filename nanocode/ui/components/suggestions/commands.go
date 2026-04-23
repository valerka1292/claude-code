package suggestions

import (
	"fmt"
	"strings"

	"nanocode/internal/mathutil"
	"nanocode/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(theme.MutedText).
			Padding(0, 1).
			Background(theme.AppBackground)
	selectedStyle = lipgloss.NewStyle().Foreground(theme.AccentContrast).Background(theme.PrimaryAccent)
)

func CommandList(width int, items []string, selected int) string {
	if len(items) == 0 {
		return ""
	}
	rows := make([]string, 0, len(items))
	for i, item := range items {
		line := fmt.Sprintf(" %s", item)
		if i == selected {
			line = selectedStyle.Render(line)
		}
		rows = append(rows, line)
	}
	contentWidth := mathutil.Max(18, mathutil.Min(width-4, maxItemWidth(items)+4))
	return boxStyle.Width(contentWidth).Render(strings.Join(rows, "\n"))
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

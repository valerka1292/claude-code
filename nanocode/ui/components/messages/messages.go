package messages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/types"
)

var (
	panelStyle = lipgloss.NewStyle().Padding(0, 0, 1, 0)
	userStyle  = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("255")).Padding(0, 1)
	agentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).PaddingLeft(1)
	dotStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

func View(list []types.Message, width int, spinnerLine string) string {
	var lines []string
	for _, msg := range list {
		switch msg.Role {
		case types.RoleUser:
			lines = append(lines, userStyle.Width(max(width-2, 10)).Render("❯ "+msg.Text))
		case types.RoleAssistant:
			lines = append(lines, agentStyle.Render(dotStyle.Render("• ")+msg.Text))
		}
		lines = append(lines, "")
	}

	if spinnerLine != "" {
		lines = append(lines, agentStyle.Render(spinnerLine), "")
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

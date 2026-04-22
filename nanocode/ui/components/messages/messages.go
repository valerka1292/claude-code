package messages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/theme"
	"nanocode/ui/types"
)

var (
	panelStyle = lipgloss.NewStyle().Padding(0, 0, 1, 0).Background(theme.AppBackground)
	userStyle  = lipgloss.NewStyle().Background(theme.SurfaceBackground).Foreground(theme.PrimaryText).Padding(0, 1)
	agentStyle = lipgloss.NewStyle().Foreground(theme.PrimaryText).PaddingLeft(1)
	thinkStyle = lipgloss.NewStyle().Foreground(theme.MutedText).PaddingLeft(3)
	dotStyle   = lipgloss.NewStyle().Foreground(theme.PrimaryAccent)
)

func View(list []types.Message, width int, spinnerLine string, thinking string, streamingText string) string {
	var lines []string
	for _, msg := range list {
		switch msg.Role {
		case types.RoleUser:
			lines = append(lines, userStyle.Width(max(width-2, 10)).Render("❯ "+msg.Text))
		case types.RoleAssistant:
			lines = append(lines, agentStyle.Render(renderAssistantBlock(msg.Text, width, false)))
		}
		lines = append(lines, "")
	}

	if spinnerLine != "" {
		lines = append(lines, agentStyle.Render(spinnerLine), "")
	}
	if thinking != "" {
		lines = append(lines, thinkStyle.Render("thinking: "+thinking), "")
	}
	if streamingText != "" {
		lines = append(lines, agentStyle.Render(renderAssistantBlock(streamingText, width, true)), "")
	}

	if len(lines) == 0 {
		lines = append(lines, "")
	}

	return panelStyle.Width(width).Render(strings.Join(lines, "\n"))
}

func renderAssistantBlock(text string, width int, streaming bool) string {
	rendered := renderMarkdown(text, max(width-4, minMarkdownWidth), streaming)
	if rendered == "" {
		return dotStyle.Render("•")
	}

	rows := strings.Split(rendered, "\n")
	rows[0] = dotStyle.Render("• ") + rows[0]
	for i := 1; i < len(rows); i++ {
		rows[i] = "  " + rows[i]
	}
	return strings.Join(rows, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

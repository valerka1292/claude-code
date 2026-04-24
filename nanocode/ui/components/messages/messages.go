package messages

import (
	"strings"

	"nanocode/internal/mathutil"
	"nanocode/ui/theme"
	"nanocode/ui/types"

	"github.com/charmbracelet/lipgloss"
)

var (
	panelStyle = lipgloss.NewStyle().Padding(0, 0, 1, 0)
	userStyle  = lipgloss.NewStyle().Foreground(theme.PrimaryText).Padding(0, 1)
	agentStyle = lipgloss.NewStyle().Foreground(theme.PrimaryText)
	thinkStyle = lipgloss.NewStyle().Foreground(theme.MutedText).PaddingLeft(3)
	dotStyle   = lipgloss.NewStyle().Foreground(theme.PrimaryAccent)

	toolRunIconStyle     = lipgloss.NewStyle().Foreground(theme.SecondaryAccent).Bold(true)
	toolSuccessIconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50fa7b")).Bold(true)
	toolErrorIconStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f5f")).Bold(true)
	toolTextStyle        = lipgloss.NewStyle().Foreground(theme.MutedText)
	toolResultStyle      = lipgloss.NewStyle().Foreground(theme.SecondaryAccent)

	toolBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(theme.SurfaceBackground).
			PaddingLeft(1).
			MarginLeft(2)
)

func View(list []types.Message, width int, spinnerLine string, thinking string, streamingText string) string {
	var lines []string
	for _, msg := range list {
		switch msg.Role {
		case types.RoleUser:
			lines = append(lines, userStyle.Width(mathutil.Max(width-2, 10)).Render("❯ "+msg.Text))
		case types.RoleAssistant:
			lines = append(lines, agentStyle.Render(renderAssistantBlock(msg.Text, width, false)))
		case types.RoleTool:
			text := msg.Text
			formattedText := text

			if strings.HasPrefix(text, "⚙") {
				formattedText = toolRunIconStyle.Render("⚙") + toolTextStyle.Render(text[3:])
			} else if strings.HasPrefix(text, "✓") {
				parts := strings.SplitN(text, "\n", 2)
				firstLine := toolSuccessIconStyle.Render("✓") + toolTextStyle.Render(parts[0][3:])
				if len(parts) > 1 {
					secondLine := toolResultStyle.Render(parts[1])
					formattedText = firstLine + "\n" + secondLine
				} else {
					formattedText = firstLine
				}
			} else if strings.HasPrefix(text, "✗") {
				parts := strings.SplitN(text, "\n", 2)
				firstLine := toolErrorIconStyle.Render("✗") + toolTextStyle.Render(parts[0][3:])
				if len(parts) > 1 {
					secondLine := toolErrorIconStyle.Render(parts[1])
					formattedText = firstLine + "\n" + secondLine
				} else {
					formattedText = firstLine
				}
			}

			lines = append(lines, toolBoxStyle.Width(mathutil.Max(width-4, 10)).Render(formattedText))
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
	rendered := renderMarkdown(text, mathutil.Max(width-4, minMarkdownWidth), streaming)
	rendered = strings.Trim(rendered, "\n\r")

	if rendered == "" {
		return dotStyle.Render("●")
	}

	rows := strings.Split(rendered, "\n")
	rows[0] = dotStyle.Render("●") + " " + rows[0]
	for i := 1; i < len(rows); i++ {
		rows[i] = "  " + rows[i]
	}
	return strings.Join(rows, "\n")
}

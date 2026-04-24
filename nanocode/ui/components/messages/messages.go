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
	userStyle  = lipgloss.NewStyle().Foreground(theme.PrimaryText).Background(theme.SurfaceBackground).Padding(0, 1).MarginBottom(1)
	agentStyle = lipgloss.NewStyle().Foreground(theme.PrimaryText).MarginBottom(1)
	thinkStyle = lipgloss.NewStyle().Foreground(theme.MutedText).PaddingLeft(3).MarginBottom(1)
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
			MarginLeft(2).
			MarginBottom(1)
	horizontalFrame = toolBoxStyle.GetHorizontalFrameSize()
)

func availableWidth(totalWidth int, frameSize int) int {
	return mathutil.Max(10, totalWidth-frameSize)
}

func View(list []types.Message, width int, spinnerLine string, thinking string, streamingText string) string {
	var blocks []string
	for _, msg := range list {
		switch msg.Role {
		case types.RoleUser:
			blocks = append(blocks, userStyle.Width(availableWidth(width, horizontalFrame)).Render("❯ "+msg.Text))
		case types.RoleAssistant:
			blocks = append(blocks, agentStyle.Render(renderAssistantBlock(msg.Text, width, false)))
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

			blocks = append(blocks, toolBoxStyle.Width(availableWidth(width, horizontalFrame+2)).Render(formattedText))
		}
	}

	if spinnerLine != "" {
		blocks = append(blocks, agentStyle.Render(spinnerLine))
	}
	if thinking != "" {
		blocks = append(blocks, thinkStyle.Render("thinking: "+thinking))
	}
	if streamingText != "" {
		blocks = append(blocks, agentStyle.Render(renderAssistantBlock(streamingText, width, true)))
	}

	if len(blocks) == 0 {
		blocks = append(blocks, "")
	}

	return panelStyle.Width(width).Render(strings.Join(blocks, "\n"))
}

func renderAssistantBlock(text string, width int, streaming bool) string {
	contentWidth := availableWidth(width, horizontalFrame+4)
	rendered := renderMarkdown(text, contentWidth, streaming)
	lines := strings.Split(rendered, "\n")
	for _, line := range lines {
		if lipgloss.Width(line) > contentWidth {
			return "!!! OVERFLOW: " + line
		}
	}
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

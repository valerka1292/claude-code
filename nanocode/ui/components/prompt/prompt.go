package prompt

import (
	"nanocode/internal/mathutil"
	"nanocode/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, true, false).
			BorderForeground(theme.MutedText).
			Foreground(theme.PrimaryText).
			Background(theme.AppBackground)
	footerStyle = lipgloss.NewStyle().Foreground(theme.SecondaryAccent).PaddingLeft(1)
)

func InputBar(text string, width int) string {
	return boxStyle.Width(width).Render("❯ " + text)
}

func Footer(width int, usage string, hint string, mode string) string {
	var badge string
	if mode == "code" {
		badge = lipgloss.NewStyle().
			Foreground(theme.AccentContrast).
			Background(theme.PrimaryAccent).
			Padding(0, 1).Bold(true).Render("CODE")
	} else {
		badge = lipgloss.NewStyle().
			Foreground(theme.AccentContrast).
			Background(theme.ModeAsk).
			Padding(0, 1).Bold(true).Render("ASK")
	}

	badgePadded := lipgloss.NewStyle().PaddingLeft(1).Render(badge)
	hintStyle := lipgloss.NewStyle().Foreground(theme.SecondaryAccent)

	leftText := hintStyle.Render("(shift+tab to cycle)")
	if hint != "" {
		leftText = hintStyle.Render(hint)
	}
	left := lipgloss.JoinHorizontal(lipgloss.Left, badgePadded, " ", leftText)

	right := footerStyle.Foreground(theme.MutedText).Render(usage)
	if usage == "" {
		return lipgloss.NewStyle().Width(width).Background(theme.AppBackground).Render(left)
	}
	spacerWidth := mathutil.Max(0, width-lipgloss.Width(left)-lipgloss.Width(right))
	spacer := lipgloss.NewStyle().Width(spacerWidth).Render("")
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, spacer, right)
	return lipgloss.NewStyle().Width(width).Background(theme.AppBackground).Render(content)
}

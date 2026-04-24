package messages

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nanocode/internal/mathutil"
	"nanocode/ui/theme"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

var (
	addBgStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#163320")).Foreground(lipgloss.Color("#2EEA78"))
	subBgStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#3d1a1a")).Foreground(lipgloss.Color("#EA4646"))
	addPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2EEA78")).Bold(true)
	subPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EA4646")).Bold(true)
	numStyle       = lipgloss.NewStyle().Foreground(theme.MutedText)
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
)

func RenderDiff(filePath string, diffText string, width int) string {
	lines := strings.Split(diffText, "\n")
	var cleanLines []string
	var lineTypes []string

	for _, line := range lines {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "\\") {
			lineTypes = append(lineTypes, "skip")
			continue
		}
		if strings.HasPrefix(line, "@@") {
			lineTypes = append(lineTypes, "@@")
			continue
		}
		if len(line) > 0 {
			char := line[0:1]
			if char == "+" || char == "-" || char == " " {
				lineTypes = append(lineTypes, char)
				cleanLines = append(cleanLines, line[1:])
				continue
			}
		}
		lineTypes = append(lineTypes, " ")
		cleanLines = append(cleanLines, line)
	}

	// Syntax-highlight everything. For +/- lines we strip background codes
	// so our own lipgloss background takes effect cleanly.
	cleanText := strings.Join(cleanLines, "\n")
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	if ext == "" {
		ext = "text"
	}
	var buf strings.Builder
	err := quick.Highlight(&buf, cleanText, ext, "terminal256", "dracula")
	highlightedText := cleanText
	if err == nil && buf.Len() > 0 {
		highlightedText = buf.String()
	}
	hlLines := strings.Split(highlightedText, "\n")

	// Gutter width
	maxLine := 0
	for _, line := range lines {
		if m := hunkHeaderRegex.FindStringSubmatch(line); len(m) == 3 {
			if nl, err2 := strconv.Atoi(m[2]); err2 == nil && nl > maxLine {
				maxLine = nl
			}
		}
	}
	maxLine += len(lines)
	digits := len(strconv.Itoa(maxLine))
	if digits < 2 {
		digits = 2
	}

	gutterWidth := digits + 3 // "%Nd │ "
	innerWidth := mathutil.Max(40, width-8)
	contentWidth := mathutil.Max(10, innerWidth-gutterWidth)

	var formatted strings.Builder
	oldLine, newLine, hlIndex := 0, 0, 0

	for i, lType := range lineTypes {
		if lType == "skip" {
			continue
		}
		if lType == "@@" {
			if m := hunkHeaderRegex.FindStringSubmatch(lines[i]); len(m) == 3 {
				oldLine, _ = strconv.Atoi(m[1])
				newLine, _ = strconv.Atoi(m[2])
			}
			gutter := strings.Repeat(" ", digits) + " │ "
			formatted.WriteString(numStyle.Render(gutter) + infoStyle.Render(lines[i]) + "\n")
			continue
		}

		hlContent := ""
		rawContent := ""
		if hlIndex < len(cleanLines) {
			rawContent = cleanLines[hlIndex]
		}
		if hlIndex < len(cleanLines) {
			hlContent = cleanLines[hlIndex]
		}
		if hlIndex < len(hlLines) {
			hlContent = hlLines[hlIndex]
		}
		hlIndex++

		var numStr, lineContent string

		switch lType {
		case "+":
			numStr = numStyle.Render(fmt.Sprintf("%*d │ ", digits, newLine))
			newLine++
			// Не используем chroma ANSI здесь: reset-коды ломают сплошной фон.
			pfx := addPrefixStyle.Render("+ ")
			lineContent = addBgStyle.Width(contentWidth).Render(pfx + rawContent)
		case "-":
			numStr = numStyle.Render(fmt.Sprintf("%*d │ ", digits, oldLine))
			oldLine++
			pfx := subPrefixStyle.Render("- ")
			lineContent = subBgStyle.Width(contentWidth).Render(pfx + rawContent)
		default:
			numStr = numStyle.Render(fmt.Sprintf("%*d │ ", digits, newLine))
			oldLine++
			newLine++
			lineContent = "  " + hlContent
		}

		formatted.WriteString(numStr + lineContent + "\n")
	}

	content := strings.TrimRight(formatted.String(), "\n")

	header := lipgloss.NewStyle().
		Background(theme.PrimaryAccent).
		Foreground(theme.AccentContrast).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2).
		Render(" WRITE: " + filePath + " ")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.SurfaceBackground).
		Width(innerWidth).
		Render(content)

	return header + "\n" + box
}

package messages

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nanocode/ui/theme"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

var (
	addBgStyle = lipgloss.NewStyle().Background(lipgloss.Color("#163320"))
	subBgStyle = lipgloss.NewStyle().Background(lipgloss.Color("#3d1a1a"))

	addPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#2EEA78")).Bold(true)
	subPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EA4646")).Bold(true)

	numStyle = lipgloss.NewStyle().Foreground(theme.MutedText)

	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
)

func RenderDiff(filePath string, diffText string) string {
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

	var formatted strings.Builder

	oldLine := 0
	newLine := 0
	hlIndex := 0

	maxLine := 0
	for _, line := range lines {
		if matches := hunkHeaderRegex.FindStringSubmatch(line); len(matches) == 3 {
			if nl, err := strconv.Atoi(matches[2]); err == nil {
				if nl > maxLine {
					maxLine = nl
				}
			}
		}
	}
	maxLine += len(lines)
	digits := len(strconv.Itoa(maxLine))
	if digits < 2 {
		digits = 2
	}

	for i, lType := range lineTypes {
		if lType == "skip" {
			continue
		}

		if lType == "@@" {
			matches := hunkHeaderRegex.FindStringSubmatch(lines[i])
			if len(matches) == 3 {
				oldLine, _ = strconv.Atoi(matches[1])
				newLine, _ = strconv.Atoi(matches[2])
			}
			gutter := strings.Repeat(" ", digits) + " │ "
			formatted.WriteString(numStyle.Render(gutter) + infoStyle.Render(lines[i]) + "\n")
			continue
		}

		var numStr string
		var lineContent string

		hlContent := ""
		if hlIndex < len(hlLines) {
			hlContent = hlLines[hlIndex]
		}
		hlIndex++

		if hlContent == "" {
			hlContent = " "
		}

		if lType == "+" {
			numStr = numStyle.Render(fmt.Sprintf("%*d │ ", digits, newLine))
			newLine++
			prefix := addPrefixStyle.Render("+ ")
			lineContent = addBgStyle.Render(prefix + hlContent)
		} else if lType == "-" {
			numStr = numStyle.Render(fmt.Sprintf("%*d │ ", digits, oldLine))
			oldLine++
			prefix := subPrefixStyle.Render("- ")
			lineContent = subBgStyle.Render(prefix + hlContent)
		} else {
			numStr = numStyle.Render(fmt.Sprintf("%*d │ ", digits, newLine))
			oldLine++
			newLine++
			prefix := "  "
			lineContent = prefix + hlContent
		}

		formatted.WriteString(numStr + lineContent + "\n")
	}

	content := strings.TrimRight(formatted.String(), "\n")

	headerText := " WRITE: " + filePath + " "
	header := lipgloss.NewStyle().
		Background(theme.PrimaryAccent).
		Foreground(theme.AccentContrast).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2).
		Render(headerText)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.SurfaceBackground).
		Render(content)

	return header + "\n" + box
}
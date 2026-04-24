package messages

import (
	"fmt"
	"os"
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
var ansiSGRRe = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

type EditLineType int

const (
	EditLineContext EditLineType = iota
	EditLineAdd
	EditLineDel
)

type EditLine struct {
	Type    EditLineType
	Content string
	OldNum  int
	NewNum  int
}

type EditHunk struct {
	OldStart int
	NewStart int
	Lines    []EditLine
}

type ParsedEditDiff struct {
	Hunks []EditHunk
}

func ParseEditDiff(diffText string) ParsedEditDiff {
	var result ParsedEditDiff
	diffText = strings.ReplaceAll(diffText, "\r\n", "\n")
	lines := strings.Split(diffText, "\n")

	var currentHunk *EditHunk
	oldLineNum := 0
	newLineNum := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") {
			continue
		}

		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				result.Hunks = append(result.Hunks, *currentHunk)
			}

			m := hunkHeaderRegex.FindStringSubmatch(line)
			if len(m) >= 3 {
				oldStart, _ := strconv.Atoi(m[1])
				newStart, _ := strconv.Atoi(m[2])
				oldLineNum = oldStart
				newLineNum = newStart

				currentHunk = &EditHunk{
					OldStart: oldStart,
					NewStart: newStart,
					Lines:    make([]EditLine, 0),
				}
			}
			continue
		}

		if currentHunk == nil {
			continue
		}

		char := line[0:1]
		content := ""
		if len(line) > 1 {
			content = line[1:]
		}

		switch char {
		case " ":
			currentHunk.Lines = append(currentHunk.Lines, EditLine{
				Type:    EditLineContext,
				Content: content,
				OldNum:  oldLineNum,
				NewNum:  newLineNum,
			})
			oldLineNum++
			newLineNum++
		case "-":
			currentHunk.Lines = append(currentHunk.Lines, EditLine{
				Type:    EditLineDel,
				Content: content,
				OldNum:  oldLineNum,
				NewNum:  0,
			})
			oldLineNum++
		case "+":
			currentHunk.Lines = append(currentHunk.Lines, EditLine{
				Type:    EditLineAdd,
				Content: content,
				OldNum:  0,
				NewNum:  newLineNum,
			})
			newLineNum++
		case "\\":
			continue
		}
	}

	if currentHunk != nil {
		result.Hunks = append(result.Hunks, *currentHunk)
	}

	return result
}

var (
	addBgStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#163320"))
	subBgStyle     = lipgloss.NewStyle().Background(lipgloss.Color("#3d1a1a"))
	addPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#63FF47")).Bold(true)
	subPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EA4646")).Bold(true)
	numStyle       = lipgloss.NewStyle().Foreground(theme.MutedText)
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Italic(true)
)

func stripBgSGR(s string) string {
	return ansiSGRRe.ReplaceAllStringFunc(s, func(code string) string {
		m := ansiSGRRe.FindStringSubmatch(code)
		if len(m) != 2 {
			return code
		}
		if m[1] == "" {
			return code
		}

		parts := strings.Split(m[1], ";")
		filtered := make([]string, 0, len(parts))
		for i := 0; i < len(parts); i++ {
			p := parts[i]
			if p == "" {
				continue
			}

			n, err := strconv.Atoi(p)
			if err != nil {
				filtered = append(filtered, p)
				continue
			}

			if (n >= 40 && n <= 49) || (n >= 100 && n <= 109) {
				continue
			}
			if n == 48 && i+1 < len(parts) {
				mode := parts[i+1]
				if mode == "5" && i+2 < len(parts) {
					i += 2
					continue
				}
				if mode == "2" && i+4 < len(parts) {
					i += 4
					continue
				}
			}

			filtered = append(filtered, p)
		}

		if len(filtered) == 0 {
			return ""
		}
		return "\x1b[" + strings.Join(filtered, ";") + "m"
	})
}

func keepFgAndReapplyBg(s string, bgANSI string) string {
	withoutBg := stripBgSGR(s)
	withoutBg = strings.ReplaceAll(withoutBg, "\x1b[0m", "\x1b[0m"+bgANSI)
	withoutBg = strings.ReplaceAll(withoutBg, "\x1b[m", "\x1b[m"+bgANSI)
	return withoutBg
}

func displayPath(filePath string) string {
	if filePath == "" {
		return filePath
	}
	cwd, err := os.Getwd()
	if err != nil {
		return filePath
	}
	rel, err := filepath.Rel(cwd, filePath)
	if err != nil {
		return filePath
	}
	if rel == "." {
		return filepath.Base(cwd)
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return filePath
	}
	return rel
}

func RenderDiff(filePath string, diffText string, width int, operation string) string {
	parsed := ParseEditDiff(diffText)

	if len(parsed.Hunks) == 0 {
		return lipgloss.NewStyle().Foreground(theme.MutedText).Render("No changes to display.")
	}

	var cleanLines []string
	for _, hunk := range parsed.Hunks {
		for _, line := range hunk.Lines {
			cleanLines = append(cleanLines, line.Content)
		}
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

	maxLine := 0
	for _, hunk := range parsed.Hunks {
		for _, line := range hunk.Lines {
			if line.NewNum > maxLine {
				maxLine = line.NewNum
			}
			if line.OldNum > maxLine {
				maxLine = line.OldNum
			}
		}
	}

	digits := len(strconv.Itoa(maxLine))
	if digits < 1 {
		digits = 1
	}

	gutterWidth := digits + 3 + digits + 3
	innerWidth := mathutil.Max(40, width-8)
	contentWidth := mathutil.Max(10, innerWidth-gutterWidth)

	var formatted strings.Builder
	hlIndex := 0

	for hIndex, hunk := range parsed.Hunks {
		if hIndex > 0 {
			sepWidth := digits*2 + 3
			text := "..."

			leftPad := (sepWidth - len(text)) / 2
			rightPad := sepWidth - len(text) - leftPad

			centeredDots := strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
			sepGutter := numStyle.Render(fmt.Sprintf("%s │ ", centeredDots))

			formatted.WriteString(sepGutter + "\n")
		}

		for _, line := range hunk.Lines {
			hlContent := line.Content
			if hlIndex < len(hlLines) {
				hlContent = hlLines[hlIndex]
			}
			hlIndex++

			var numStr, lineContent string

			switch line.Type {
			case EditLineAdd:
				numStr = numStyle.Render(fmt.Sprintf("%*s │ %*d │ ", digits, "", digits, line.NewNum))
				pfx := keepFgAndReapplyBg(addPrefixStyle.Render("+ "), "\x1b[48;2;22;51;32m")
				lineContent = addBgStyle.Width(contentWidth).Render(
					pfx + keepFgAndReapplyBg(hlContent, "\x1b[48;2;22;51;32m"),
				)
			case EditLineDel:
				numStr = numStyle.Render(fmt.Sprintf("%*d │ %*s │ ", digits, line.OldNum, digits, ""))
				pfx := keepFgAndReapplyBg(subPrefixStyle.Render("- "), "\x1b[48;2;61;26;26m")
				lineContent = subBgStyle.Width(contentWidth).Render(
					pfx + keepFgAndReapplyBg(hlContent, "\x1b[48;2;61;26;26m"),
				)
			case EditLineContext:
				numStr = numStyle.Render(fmt.Sprintf("%*d │ %*d │ ", digits, line.OldNum, digits, line.NewNum))
				lineContent = "  " + hlContent
			}

			formatted.WriteString(numStr + lineContent + "\n")
		}
	}

	content := strings.TrimRight(formatted.String(), "\n")

	opTitle := " " + strings.ToUpper(operation) + ": " + displayPath(filePath) + " "
	header := lipgloss.NewStyle().
		Background(theme.PrimaryAccent).
		Foreground(theme.AccentContrast).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2).
		Render(opTitle)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.SurfaceBackground).
		Width(innerWidth).
		Render(content)

	return header + "\n" + box
}

package messages

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

const minMarkdownWidth = 20

var rendererCache sync.Map

func stripANSI(str string) string {
	var out strings.Builder
	inEscape := false
	for _, r := range str {
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func removeVisibleIndent(line string, count int) string {
	removed := 0
	var out strings.Builder
	inEscape := false

	for _, r := range line {
		if inEscape {
			out.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		if r == '\x1b' {
			inEscape = true
			out.WriteRune(r)
			continue
		}
		if removed < count && r == ' ' {
			removed++
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func renderMarkdown(text string, width int, streaming bool) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	if streaming {
		normalized = stabilizeStreamingMarkdown(normalized)
	}
	normalized = addCodeLineNumbers(normalized)

	renderer, err := getRenderer(width)
	if err != nil {
		return normalized
	}

	rendered, err := renderer.Render(normalized)
	if err != nil {
		return normalized
	}

	return dedentRendered(strings.TrimRight(rendered, "\n"))
}

func stabilizeStreamingMarkdown(text string) string {
	if strings.Count(text, "```")%2 == 1 {
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += "```"
	}
	return text
}

func dedentRendered(text string) string {
	lines := strings.Split(text, "\n")
	minIndent := -1
	for _, line := range lines {
		plainLine := stripANSI(line)
		if strings.TrimSpace(plainLine) == "" {
			continue
		}
		indent := 0
		for _, r := range plainLine {
			if r != ' ' {
				break
			}
			indent++
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return text
	}
	for i, line := range lines {
		lines[i] = removeVisibleIndent(line, minIndent)
	}
	return strings.Join(lines, "\n")
}

func addCodeLineNumbers(text string) string {
	lines := strings.Split(text, "\n")
	var out []string
	inFence := false
	fenceToken := ""
	fenceLen := 0
	var block []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inFence {
			token := fenceTokenFromStart(trimmed)
			if token != "" {
				inFence = true
				fenceToken = token
				fenceLen = len(token)
				block = block[:0]
			}
			out = append(out, line)
			continue
		}

		if isFenceEnd(trimmed, fenceToken, fenceLen) {
			out = append(out, withLineNumbers(block)...)
			out = append(out, line)
			inFence = false
			fenceToken = ""
			fenceLen = 0
			block = block[:0]
			continue
		}

		block = append(block, line)
	}

	if inFence {
		out = append(out, withLineNumbers(block)...)
	}

	return strings.Join(out, "\n")
}

func withLineNumbers(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	digits := len(strconv.Itoa(len(lines)))
	if digits < 2 {
		digits = 2
	}
	result := make([]string, 0, len(lines))
	for i, line := range lines {
		result = append(result, fmt.Sprintf("%*d │ %s", digits, i+1, line))
	}
	return result
}

func fenceTokenFromStart(line string) string {
	if strings.HasPrefix(line, "```") {
		return "```"
	}
	if strings.HasPrefix(line, "~~~") {
		return "~~~"
	}
	return ""
}

func isFenceEnd(line string, token string, minLen int) bool {
	if token == "" || len(line) < minLen {
		return false
	}
	if !strings.HasPrefix(line, token) {
		return false
	}
	for _, r := range line {
		if rune(token[0]) != r && r != ' ' {
			return false
		}
	}
	return true
}

func getRenderer(width int) (*glamour.TermRenderer, error) {
	wrap := width
	if wrap < minMarkdownWidth {
		wrap = minMarkdownWidth
	}

	if cached, ok := rendererCache.Load(wrap); ok {
		return cached.(*glamour.TermRenderer), nil
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(wrap),
		glamour.WithStandardStyle("dark"),
	)
	if err != nil {
		return nil, err
	}

	actual, _ := rendererCache.LoadOrStore(wrap, renderer)
	return actual.(*glamour.TermRenderer), nil
}

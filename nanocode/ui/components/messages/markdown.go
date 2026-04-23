// ui/components/messages/markdown.go
package messages

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"charm.land/glamour/v2"
	"github.com/charmbracelet/lipgloss"
)

const minMarkdownWidth = 20

// ── renderer cache (keyed by wrap width) ────────────────────────────────────

var rendererCache sync.Map

// ── public entry-points ──────────────────────────────────────────────────────

func renderMarkdown(text string, width int, streaming bool) string {
	normalized := normalizeNewlines(text)
	if streaming {
		normalized = stabilizeStreamingMarkdown(normalized)
	}

	// Split into prose segments and code-block segments, render each correctly.
	segments := splitSegments(normalized)
	var sb strings.Builder
	for _, seg := range segments {
		if seg.isCode {
			sb.WriteString(renderCodeBlock(seg.lang, seg.body, width))
		} else {
			sb.WriteString(renderProse(seg.body, width))
		}
	}
	return dedentRendered(strings.TrimRight(sb.String(), "\n"))
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

// ── segment splitting ────────────────────────────────────────────────────────

type segment struct {
	isCode bool
	lang   string
	body   string
}

// splitSegments splits markdown into alternating prose / fenced-code segments.
func splitSegments(text string) []segment {
	lines := strings.Split(text, "\n")
	var segs []segment
	var cur strings.Builder
	inFence := false
	fenceToken := ""
	lang := ""

	flush := func(code bool, l string) {
		body := cur.String()
		cur.Reset()
		if body == "" && !code {
			return
		}
		segs = append(segs, segment{isCode: code, lang: l, body: body})
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inFence {
			tok := fenceTokenFromStart(trimmed)
			if tok != "" {
				flush(false, "")
				inFence = true
				fenceToken = tok
				lang = strings.TrimSpace(strings.TrimPrefix(trimmed, tok))
				continue
			}
			cur.WriteString(line)
			cur.WriteByte('\n')
			continue
		}
		// inside fence
		if isFenceEnd(trimmed, fenceToken, len(fenceToken)) {
			flush(true, lang)
			inFence = false
			fenceToken = ""
			lang = ""
			continue
		}
		cur.WriteString(line)
		cur.WriteByte('\n')
	}
	// leftover
	if inFence {
		flush(true, lang)
	} else {
		flush(false, "")
	}
	return segs
}

// ── code block renderer ──────────────────────────────────────────────────────

var (
	codeHeaderStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2A2A2A")).
			Foreground(lipgloss.Color("#808080")).
			PaddingLeft(1)

	codeBodyBg = lipgloss.Color("#1A1A1A")

	lineNumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A4A4A")).
			Background(codeBodyBg)

	gutterSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Background(codeBodyBg)

	codeLineStyle = lipgloss.NewStyle().
			Background(codeBodyBg).
			Foreground(lipgloss.Color("#E5E7EB"))

	codeBorderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333"))

	langBadgeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FBFA56")).
			Foreground(lipgloss.Color("#181818")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	langBarBg = lipgloss.NewStyle().
			Background(lipgloss.Color("#2A2A2A"))
)

func renderCodeBlock(lang, body string, width int) string {
	// Clamp width
	if width < minMarkdownWidth {
		width = minMarkdownWidth
	}

	// Syntax-highlight the body via glamour (renders as a standalone code block)
	highlighted := highlightCode(lang, body, width)

	rawLines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	hlLines := strings.Split(strings.TrimRight(highlighted, "\n"), "\n")

	// Ensure same line count (glamour may add/remove blank lines)
	for len(hlLines) < len(rawLines) {
		hlLines = append(hlLines, "")
	}

	digits := len(fmt.Sprintf("%d", len(rawLines)))
	if digits < 2 {
		digits = 2
	}

	innerWidth := width - digits - 4 // gutter: " NN │ "

	var sb strings.Builder

	// ── top bar: language badge ──────────────────────────────────────────
	langLabel := lang
	if langLabel == "" {
		langLabel = "text"
	}
	badge := langBadgeStyle.Render(langLabel)
	barPad := langBarBg.Width(width - lipgloss.Width(badge)).Render("")
	sb.WriteString(badge + barPad)
	sb.WriteByte('\n')

	// ── top border ───────────────────────────────────────────────────────
	sb.WriteString(codeBorderStyle.Render("┌" + strings.Repeat("─", width-2) + "┐"))
	sb.WriteByte('\n')

	// ── code lines ───────────────────────────────────────────────────────
	for i, hl := range hlLines {
		num := lineNumStyle.Render(fmt.Sprintf(" %*d ", digits, i+1))
		sep := gutterSepStyle.Render("│")
		// Pad/truncate the highlighted line to innerWidth visible chars
		visLen := visibleWidth(hl)
		var codePart string
		if visLen < innerWidth {
			codePart = codeLineStyle.Render(hl + strings.Repeat(" ", innerWidth-visLen))
		} else {
			codePart = codeLineStyle.Render(hl)
		}
		sb.WriteString(codeBorderStyle.Render("│") + num + sep + " " + codePart + codeBorderStyle.Render("│"))
		sb.WriteByte('\n')
	}

	// ── bottom border ────────────────────────────────────────────────────
	sb.WriteString(codeBorderStyle.Render("└" + strings.Repeat("─", width-2) + "┘"))
	sb.WriteByte('\n')

	return sb.String()
}

// highlightCode renders only the code body through glamour so chroma applies
// syntax colouring, then strips the surrounding box glamour adds.
func highlightCode(lang, body string, width int) string {
	// Wrap in a fenced block and render
	md := "```" + lang + "\n" + strings.TrimRight(body, "\n") + "\n```\n"

	renderer, err := getRenderer(width)
	if err != nil {
		return body
	}
	rendered, err := renderer.Render(md)
	if err != nil {
		return body
	}

	// glamour wraps code blocks – extract just the inner lines
	lines := strings.Split(rendered, "\n")
	// Drop leading/trailing empty lines, find the code content lines
	var inner []string
	started := false
	for _, l := range lines {
		plain := stripANSI(l)
		trimmed := strings.TrimSpace(plain)
		// Skip the decorative lines glamour adds around the block
		if !started && trimmed == "" {
			continue
		}
		started = true
		inner = append(inner, l)
	}
	// Remove trailing blank lines
	for len(inner) > 0 && strings.TrimSpace(stripANSI(inner[len(inner)-1])) == "" {
		inner = inner[:len(inner)-1]
	}
	return strings.Join(inner, "\n")
}

// ── prose renderer ───────────────────────────────────────────────────────────

func renderProse(text string, width int) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	renderer, err := getRenderer(width)
	if err != nil {
		return text
	}
	rendered, err := renderer.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimRight(rendered, "\n") + "\n"
}

// ── glamour renderer (prose-only style, no code-block chrome) ───────────────

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
		glamour.WithStylesFromJSONBytes([]byte(glamourStyle())),
	)
	if err != nil {
		return nil, err
	}
	actual, _ := rendererCache.LoadOrStore(wrap, renderer)
	return actual.(*glamour.TermRenderer), nil
}

// ── style JSON ───────────────────────────────────────────────────────────────

func glamourStyle() string {
	return `{
	"document": { "block_prefix": "", "block_suffix": "" },

	"block_quote": {
		"color": "#9CA3AF",
		"italic": true,
		"indent": 1,
		"indent_token": "▎ "
	},

	"list": { "level_indent": 2 },

	"heading": {
		"bold": true,
		"block_suffix": "\n"
	},
	"h1": {
		"prefix": "󰊠  ",
		"color": "#FBFA56",
		"bold": true,
		"block_suffix": "\n"
	},
	"h2": {
		"prefix": "  ",
		"color": "#EBDC2F",
		"bold": true
	},
	"h3": {
		"prefix": " ▸ ",
		"color": "#D4D4D4",
		"bold": true
	},
	"h4": { "prefix": "   · ", "color": "#9CA3AF", "bold": true },
	"h5": { "prefix": "     ‣ ", "color": "#6B7280" },
	"h6": { "prefix": "       – ", "color": "#4B5563" },

	"text":          { "color": "#E5E7EB" },
	"strong":        { "bold": true, "color": "#F9FAFB" },
	"italic":        { "italic": true, "color": "#D1D5DB" },
	"strikethrough": { "crossed_out": true, "color": "#6B7280" },

	"code": {
		"color": "#FBFA56",
		"background_color": "#2A2A2A",
		"prefix": " ",
		"suffix": " "
	},

	"code_block": {
		"color": "#E5E7EB",
		"background_color": "#1A1A1A",
		"margin": 0,
		"chroma": {
			"text":                 { "color": "#E5E7EB" },
			"error":                { "color": "#FF5555", "bold": true },
			"comment":              { "color": "#6A9955", "italic": true },
			"comment_preproc":      { "color": "#C586C0" },
			"keyword":              { "color": "#569CD6", "bold": true },
			"keyword_reserved":     { "color": "#569CD6", "bold": true },
			"keyword_namespace":    { "color": "#C586C0", "bold": true },
			"keyword_type":         { "color": "#4EC9B0" },
			"operator":             { "color": "#D4D4D4" },
			"punctuation":          { "color": "#D4D4D4" },
			"name":                 { "color": "#E5E7EB" },
			"name_builtin":         { "color": "#DCDCAA" },
			"name_tag":             { "color": "#4EC9B0" },
			"name_attribute":       { "color": "#9CDCFE" },
			"name_class":           { "color": "#4EC9B0", "bold": true },
			"name_constant":        { "color": "#4FC1FF" },
			"name_decorator":       { "color": "#DCDCAA" },
			"name_exception":       { "color": "#F44747" },
			"name_function":        { "color": "#DCDCAA" },
			"name_other":           { "color": "#9CDCFE" },
			"name_variable":        { "color": "#9CDCFE" },
			"literal_number":       { "color": "#B5CEA8" },
			"literal_string":       { "color": "#CE9178" },
			"literal_string_doc":   { "color": "#6A9955", "italic": true },
			"literal_string_escape":{ "color": "#D7BA7D" },
			"literal_string_interpol":{ "color": "#569CD6" },
			"generic_deleted":      { "color": "#FF5555" },
			"generic_inserted":     { "color": "#50FA7B" },
			"generic_heading":      { "color": "#FBFA56", "bold": true },
			"generic_subheading":   { "color": "#EBDC2F" },
			"generic_strong":       { "bold": true },
			"generic_emph":         { "italic": true }
		}
	},

	"paragraph": { "margin": 0 },

	"table": {
		"center_separator": "┼",
		"column_separator": "│",
		"row_separator": "─"
	},

	"item":        { "block_prefix": "  ● " },
	"enumeration": { "block_prefix": ". " },
	"task": {
		"ticked":   "[✓] ",
		"unticked": "[ ] "
	},

	"hr": {
		"color": "#333333",
		"format": "\n─────────────────────────────────────────────────────────────────\n"
	},

	"link":       { "color": "#60A5FA", "underline": true },
	"link_text":  { "color": "#93C5FD", "bold": true },
	"image":      { "color": "#60A5FA", "underline": true },
	"image_text": { "color": "#93C5FD", "bold": true, "format": "🖼  {{.text}}" }
}`
}

// ── helpers ──────────────────────────────────────────────────────────────────

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

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

// visibleWidth returns the number of visible characters (ignoring ANSI escapes).
func visibleWidth(s string) int {
	return utf8.RuneCountInString(stripANSI(s))
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

func fenceTokenFromStart(line string) string {
	if strings.HasPrefix(line, "```") {
		return "```"
	}
	if strings.HasPrefix(line, "~~~") {
		return "~~~"
	}
	return ""
}

func isFenceEnd(line, token string, minLen int) bool {
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

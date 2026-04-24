package messages

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"nanocode/internal/mathutil"
	"nanocode/ui/theme"

	"charm.land/glamour/v2"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/lipgloss"
)

const minMarkdownWidth = 20

type rendererCacheState struct {
	mu       sync.Mutex
	width    int
	renderer *glamour.TermRenderer
}

var rendererCache rendererCacheState

// Фикс для LLM, которые иногда забывают ставить пробел после решеток (например: ##Заголовок)
var headingRegex = regexp.MustCompile(`(?m)^(#{1,6})([^#\s].*)$`)

type segment struct {
	isCode   bool
	language string
	content  string
}

// getCustomStyle создает премиальную тему рендера.
// - Заголовки стали более контрастными.
// - Инлайн-код теперь использует черный текст (AccentContrast) на акцентном желтом фоне.
func getCustomStyle() (string, error) {
	style := map[string]map[string]any{
		"document":      {"margin": 0, "color": theme.PrimaryTextHex},
		"block_quote":   {"indent": 1, "indent_token": "┃ ", "color": theme.MutedTextHex},
		"paragraph":     {},
		"list":          {"level_indent": 2},
		"heading":       {"block_suffix": "\n", "bold": true},
		"h1":            {"color": theme.AccentContrastHex, "background_color": theme.PrimaryAccentHex, "bold": true, "prefix": " "},
		"h2":            {"color": theme.PrimaryAccentHex, "bold": true},
		"h3":            {"color": theme.SecondaryAccentHex, "bold": true},
		"h4":            {"color": theme.PrimaryTextHex, "bold": true, "italic": true},
		"h5":            {"color": theme.MutedTextHex, "bold": true},
		"h6":            {"color": theme.MutedTextHex, "bold": true},
		"strikethrough": {"crossed_out": true},
		"emph":          {"italic": true, "color": theme.PrimaryTextHex},
		"strong":        {"bold": true, "color": theme.PrimaryAccentHex},
		"hr":            {"color": theme.SurfaceBackgroundHex, "format": "\n────────────────────────────────────────────────────────────\n"},
		"item":          {"block_prefix": "• "},
		"enumeration":   {"block_prefix": ". "},
		"task":          {"ticked": "[✓] ", "unticked": "[ ] "},
		"link":          {"color": theme.SecondaryAccentHex, "underline": true},
		"link_text":     {"color": theme.PrimaryAccentHex, "bold": true},
		"image":         {"color": theme.PrimaryAccentHex, "underline": true},
		"image_text":    {"format": "🖼️ {{.text}}"},
		"code":          {"color": theme.AccentContrastHex, "background_color": theme.PrimaryAccentHex},
		"table":         {"center_separator": "┼", "column_separator": "│", "row_separator": "─"},
	}
	raw, err := json.Marshal(style)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func renderMarkdown(text string, width int, streaming bool) string {
	text = preprocessMarkdown(text, streaming)
	segments := parseSegments(text)
	var result strings.Builder

	renderer, err := getRenderer(width)
	if err != nil {
		return strings.TrimSpace(text)
	}

	for i, seg := range segments {
		if seg.isCode {
			ui := renderCodeUI(seg.language, seg.content, width)
			result.WriteString(ui)
		} else {
			textSeg := strings.TrimSpace(seg.content)
			if textSeg == "" {
				continue
			}

			rendered, err := renderer.Render(textSeg)
			if err != nil {
				result.WriteString(textSeg)
			} else {
				result.WriteString(strings.TrimRight(rendered, "\n"))
			}
		}
		if i < len(segments)-1 {
			result.WriteString("\n\n") // Отступ между блоками
		}
	}

	return strings.TrimSpace(result.String())
}

func preprocessMarkdown(text string, streaming bool) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = headingRegex.ReplaceAllString(text, "$1 $2") // Лечим слипшиеся заголовки

	if streaming {
		text = stabilizeStreamingMarkdown(text)
	}

	return text
}

// Закрывает парные блоки, если нейросеть оборвала генерацию на середине
func stabilizeStreamingMarkdown(text string) string {
	if strings.Count(text, "```")%2 == 1 {
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += "```"
	}
	return text
}

func parseSegments(text string) []segment {
	var segments []segment
	lines := strings.Split(text, "\n")
	var current strings.Builder
	inFence := false
	fenceToken := ""
	language := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inFence {
			if token := fenceTokenFromStart(trimmed); token != "" {
				if current.Len() > 0 {
					segments = append(segments, segment{isCode: false, content: current.String()})
					current.Reset()
				}
				inFence = true
				fenceToken = token
				language = strings.TrimSpace(trimmed[len(token):])
				continue
			}
			if strings.HasPrefix(trimmed, "$$") {
				if current.Len() > 0 {
					segments = append(segments, segment{isCode: false, content: current.String()})
					current.Reset()
				}
				inFence = true
				fenceToken = "$$"
				language = "math"

				if len(trimmed) > 2 && strings.HasSuffix(trimmed, "$$") {
					content := trimmed[2 : len(trimmed)-2]
					segments = append(segments, segment{isCode: true, language: "math", content: content + "\n"})
					inFence = false
					fenceToken = ""
					language = ""
				} else {
					content := trimmed[2:]
					if content != "" {
						current.WriteString(content + "\n")
					}
				}
				continue
			}
			current.WriteString(line + "\n")
		} else {
			if isFenceEnd(trimmed, fenceToken, len(fenceToken)) || (fenceToken == "$$" && strings.HasSuffix(trimmed, "$$")) {
				if fenceToken == "$$" {
					content := trimmed[:len(trimmed)-2]
					if content != "" {
						current.WriteString(content + "\n")
					}
				}
				segments = append(segments, segment{isCode: true, language: language, content: current.String()})
				current.Reset()
				inFence = false
				fenceToken = ""
				language = ""
				continue
			}
			current.WriteString(line + "\n")
		}
	}

	if current.Len() > 0 {
		segments = append(segments, segment{isCode: inFence, language: language, content: current.String()})
	}
	return segments
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

func renderCodeUI(lang, content string, width int) string {
	content = strings.TrimSuffix(content, "\n")
	if content == "" {
		return ""
	}

	var buf strings.Builder
	highlighted := content

	// Подсветка синтаксиса
	err := quick.Highlight(&buf, content, lang, "terminal256", "dracula")
	if err == nil && buf.Len() > 0 {
		highlighted = buf.String()
	}

	lines := strings.Split(highlighted, "\n")

	// ДИНАМИЧЕСКИЙ РАЗМЕР НОМЕРОВ СТРОК: отступ равен кол-ву цифр в самой длинной строке
	digits := len(strconv.Itoa(len(lines)))

	var formatted strings.Builder
	numStyle := lipgloss.NewStyle().Foreground(theme.MutedText) // MutedText отлично контрастирует и не сливается с рамкой

	for i, line := range lines {
		numStr := fmt.Sprintf("%*d │ ", digits, i+1)
		formatted.WriteString(numStyle.Render(numStr) + line)
		if i < len(lines)-1 {
			formatted.WriteString("\n")
		}
	}

	headerText := " " + strings.ToUpper(lang) + " "
	if lang == "" {
		headerText = " CODE "
	} else if lang == "math" {
		headerText = " MATHJAX "
	} else if lang == "mermaid" {
		headerText = " MERMAID "
	}

	// Элегантная "вкладка"
	header := lipgloss.NewStyle().
		Background(theme.PrimaryAccent).
		Foreground(theme.AccentContrast).
		Bold(true).
		Padding(0, 1).
		MarginLeft(2).
		Render(headerText)

	// Рамка вокруг кода
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.SurfaceBackground)
	horizontalFrame := boxStyle.GetHorizontalFrameSize()
	box := boxStyle.
		Width(mathutil.Max(20, width-horizontalFrame)).
		Render(formatted.String())

	return header + "\n" + box
}

func getRenderer(width int) (*glamour.TermRenderer, error) {
	wrap := width
	if wrap < minMarkdownWidth {
		wrap = minMarkdownWidth
	}

	rendererCache.mu.Lock()
	defer rendererCache.mu.Unlock()

	if rendererCache.renderer != nil && rendererCache.width == wrap {
		return rendererCache.renderer, nil
	}

	styleJSON, err := getCustomStyle()
	if err != nil {
		return nil, err
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(wrap),
		glamour.WithStylesFromJSONBytes([]byte(styleJSON)),
	)
	if err != nil {
		return nil, err
	}

	rendererCache.width = wrap
	rendererCache.renderer = renderer
	return renderer, nil
}

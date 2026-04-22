package messages

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

const minMarkdownWidth = 20

var rendererCache sync.Map

func renderMarkdown(text string, width int, streaming bool) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	if streaming {
		normalized = stabilizeStreamingMarkdown(normalized)
	}

	renderer, err := getRenderer(width)
	if err != nil {
		return normalized
	}

	rendered, err := renderer.Render(normalized)
	if err != nil {
		return normalized
	}

	return strings.TrimRight(rendered, "\n")
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

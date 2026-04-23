package messages

import (
	"strings"
	"testing"
)

// Локальный хелпер для тестов, чтобы очищать lipgloss/ansi стили и проверять сырой текст
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

func TestStabilizeStreamingMarkdown_ClosesOpenFence(t *testing.T) {
	in := "```python\nprint('x')"
	out := stabilizeStreamingMarkdown(in)
	if strings.Count(out, "```")%2 != 0 {
		t.Fatalf("expected balanced fences, got %q", out)
	}
}

func TestRenderMarkdown_RendersStrongText(t *testing.T) {
	out := renderMarkdown("**bold**", 80, false)
	if !strings.Contains(out, "bold") {
		t.Fatalf("expected markdown rendering to preserve text, got %q", out)
	}
}

// Тест успешно вызовет функцию из messages.go (в рамках одного пакета messages)
func TestRenderAssistantBlock_PrefixesBullet(t *testing.T) {
	out := renderAssistantBlock("hello", 80, false)
	if !strings.Contains(out, "●") {
		t.Fatalf("expected assistant output to include bullet prefix, got %q", out)
	}
}

func TestRenderCodeUI_ContainsLineNumbers(t *testing.T) {
	out := renderCodeUI("go", "fmt.Println(\"a\")\nfmt.Println(\"b\")\n", 80)
	plain := stripANSI(out)
	if !strings.Contains(plain, "1 │") || !strings.Contains(plain, "2 │") {
		t.Fatalf("expected line numbers in code block, got %q", plain)
	}
}

func TestRenderCodeUI_ContainsLangBadge(t *testing.T) {
	out := renderCodeUI("python", "x = 1\n", 80)
	plain := stripANSI(out)
	if !strings.Contains(plain, "PYTHON") {
		t.Fatalf("expected language badge in code block, got %q", plain)
	}
}

func TestParseSegments_SeparatesCodeAndProse(t *testing.T) {
	md := "hello\n```go\nfmt.Println()\n```\nworld\n"
	segs := parseSegments(md)
	var codeCount, proseCount int
	for _, s := range segs {
		if s.isCode {
			codeCount++
		} else {
			proseCount++
		}
	}
	if codeCount != 1 {
		t.Fatalf("expected 1 code segment, got %d", codeCount)
	}
	if proseCount < 1 {
		t.Fatalf("expected prose segments, got %d", proseCount)
	}
}

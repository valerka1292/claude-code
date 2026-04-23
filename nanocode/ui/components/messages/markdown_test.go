// ui/components/messages/markdown_test.go
package messages

import (
	"strings"
	"testing"
)

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

func TestRenderAssistantBlock_PrefixesBullet(t *testing.T) {
	out := renderAssistantBlock("hello", 80, false)
	if !strings.Contains(out, "●") {
		t.Fatalf("expected assistant output to include bullet prefix, got %q", out)
	}
}

func TestRenderCodeBlock_ContainsLineNumbers(t *testing.T) {
	out := renderCodeBlock("go", "fmt.Println(\"a\")\nfmt.Println(\"b\")\n", 80)
	if !strings.Contains(stripANSI(out), "1") || !strings.Contains(stripANSI(out), "2") {
		t.Fatalf("expected line numbers in code block, got %q", out)
	}
}

func TestRenderCodeBlock_ContainsLangBadge(t *testing.T) {
	out := renderCodeBlock("python", "x = 1\n", 80)
	if !strings.Contains(stripANSI(out), "python") {
		t.Fatalf("expected language badge in code block, got %q", out)
	}
}

func TestRenderCodeBlock_HasBorders(t *testing.T) {
	out := renderCodeBlock("go", "x := 1\n", 80)
	plain := stripANSI(out)
	if !strings.Contains(plain, "┌") || !strings.Contains(plain, "└") {
		t.Fatalf("expected box borders in code block, got %q", out)
	}
}

func TestSplitSegments_SeparatesCodeAndProse(t *testing.T) {
	md := "hello\n```go\nfmt.Println()\n```\nworld\n"
	segs := splitSegments(md)
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

func TestDedentRendered_RemovesSharedIndent(t *testing.T) {
	in := "    line1\n    line2"
	out := dedentRendered(in)
	if out != "line1\nline2" {
		t.Fatalf("expected common indent removed, got %q", out)
	}
}

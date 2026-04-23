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
	if !strings.Contains(out, "•") {
		t.Fatalf("expected assistant output to include bullet prefix, got %q", out)
	}
}

func TestAddCodeLineNumbers_NumberedFence(t *testing.T) {
	in := "```go\nfmt.Println(\"a\")\nfmt.Println(\"b\")\n```"
	out := addCodeLineNumbers(in)
	if !strings.Contains(out, " 1 │ fmt.Println(\"a\")") || !strings.Contains(out, " 2 │ fmt.Println(\"b\")") {
		t.Fatalf("expected numbered code lines, got %q", out)
	}
}

func TestDedentRendered_RemovesSharedIndent(t *testing.T) {
	in := "    line1\n    line2"
	out := dedentRendered(in)
	if out != "line1\nline2" {
		t.Fatalf("expected common indent removed, got %q", out)
	}
}

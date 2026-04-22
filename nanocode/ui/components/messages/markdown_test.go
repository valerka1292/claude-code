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

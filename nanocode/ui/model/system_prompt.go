package model

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"nanocode/ui/model/agent"
)

var (
	staticSystemPromptOnce sync.Once
	staticSystemPromptText string
)

func buildSystemPrompts(mode AgentMode) []string {
	return []string{
		buildStaticSystemPrompt(),
		buildDynamicSystemPrompt(mode),
	}
}

func buildStaticSystemPrompt() string {
	staticSystemPromptOnce.Do(func() {
		sections := []string{
			"# Identity",
			"You are **nanocode - autonomous coding agent**.",
			"",
			"# Tool Protocol",
			"- Solve tasks through available tool calls, not by guessing outcomes.",
			"- Treat tool output as the single source of truth.",
			"- Any read, write, search, or execution action must happen via tools.",
			"",
			"# Agentic Loop",
			"- Continue until user request is fully resolved or a blocker is hit.",
			"- Evaluate, choose best tool, execute.",
			"- Keep CLI text concise. Return short completion report.",
			"- In user-facing text, always use paths relative to CWD; use absolute paths only in tool calls.",
		}
		staticSystemPromptText = strings.Join(sections, "\n")
	})
	return staticSystemPromptText
}

func buildDynamicSystemPrompt(mode AgentMode) string {
	cwd := currentWorkingDirectory()
	shell := shellName()
	osName := runtime.GOOS

	var modeDesc string
	if mode == ModeAsk {
		modeDesc = "Current Mode: `ask` (READ-ONLY).\nWrite tools will explicitly fail. If writing is required, ask the user to press `shift+tab` to switch to `code` mode."
	} else {
		modeDesc = "Current Mode: `code` (READ/WRITE).\nAll tools are available."
	}

	registry := agent.NewDefaultToolRegistry()
	var toolsDesc []string
	toolsDesc = append(toolsDesc, "# Tools Status")
	for _, t := range registry.GetAllTools() {
		status := "✅ Available"
		if mode == ModeAsk && !t.ReadOnly {
			status = "❌ BLOCKED (Requires 'code' mode)"
		}
		toolsDesc = append(toolsDesc, fmt.Sprintf("- %s: %s", t.Name, status))
	}

	sections := []string{
		"# Environment",
		fmt.Sprintf("- CWD: `%s`", cwd),
		fmt.Sprintf("- OS: `%s`", osName),
		fmt.Sprintf("- Shell: `%s`", shell),
		"",
		"# Permissions",
		modeDesc,
		"",
		strings.Join(toolsDesc, "\n"),
	}
	return strings.Join(sections, "\n")
}

func currentWorkingDirectory() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return cwd
}

func shellName() string {
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" && runtime.GOOS == "windows" {
		shell = strings.TrimSpace(os.Getenv("COMSPEC"))
	}
	if shell == "" {
		return "unknown"
	}
	return shell
}

func buildSystemPrompt(mode AgentMode) string {
	return strings.Join(buildSystemPrompts(mode), "\n\n")
}

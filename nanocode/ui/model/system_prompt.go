package model

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	staticSystemPromptOnce sync.Once
	staticSystemPromptText string
)

func buildSystemPrompts() []string {
	return []string{
		buildStaticSystemPrompt(),
		buildDynamicSystemPrompt(),
	}
}

func buildStaticSystemPrompt() string {
	staticSystemPromptOnce.Do(func() {
		sections := []string{
			"# Identity",
			"You are **nanocode - autonomous coding agent**.",
			"",
			"# Tool Protocol",
			"- Solve tasks through available tool calls, not by guessing outcomes in plain text.",
			"- Treat tool output as the single source of truth for the next step.",
			"- Any read, write, search, or execution action must happen via tools.",
			"",
			"# Agentic Loop",
			"- Continue until the user request is fully resolved or a real blocker is hit.",
			"- After each tool result: evaluate, choose the best next tool, execute.",
			"- Keep CLI text concise: brief progress + next action.",
			"- Finish by returning a short completion report with outcome and key artifacts.",
		}
		staticSystemPromptText = strings.Join(sections, "\n")
	})
	return staticSystemPromptText
}

func buildDynamicSystemPrompt() string {
	cwd := currentWorkingDirectory()
	shell := shellName()
	osName := runtime.GOOS

	sections := []string{
		"# Environment",
		"Current runtime context:",
		fmt.Sprintf("- CWD: `%s`", cwd),
		fmt.Sprintf("- OS: `%s`", osName),
		fmt.Sprintf("- Shell: `%s`", shell),
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
	if shell == "" {
		if runtime.GOOS == "windows" {
			shell = strings.TrimSpace(os.Getenv("COMSPEC"))
		}
	}
	if shell == "" {
		return "unknown"
	}
	return shell
}

func buildSystemPrompt() string {
	return strings.Join(buildSystemPrompts(), "\n\n")
}

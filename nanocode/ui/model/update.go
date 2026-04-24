package model

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/internal/mathutil"
	"nanocode/ui/components/messages"
	"nanocode/ui/components/nobby"
	"nanocode/ui/types"
)

const confirmWindow = 800 * time.Millisecond

func (m *Model) isPendingConfirmationFor(key string) bool {
	if !m.chat.confirmPending || m.chat.confirmKey != key {
		return false
	}
	return time.Since(m.chat.confirmPressTime) <= confirmWindow
}

func (m *Model) setPendingConfirmation(key string) {
	m.chat.confirmPending = true
	m.chat.confirmKey = key
	m.chat.confirmPressTime = time.Now()
}

func (m *Model) clearPendingConfirmation() {
	m.chat.confirmPending = false
	m.chat.confirmKey = ""
	m.chat.confirmPressTime = time.Time{}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout.width = msg.Width
		m.layout.height = msg.Height
		m.input.Width = mathutil.Max(10, msg.Width-6)
		m.providers.input.Width = mathutil.Max(30, msg.Width/2)
		m.resizeViewport()
		m.refreshViewport(false)
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case nobbyTickMsg:
		m.nobbyStep++
		return m, nobbyTickCmd(m.nobbyPose, m.nobbyStep)
	case spinnerChangedMsg:
		if msg.saved {
			m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: "Settings saved: spinner style updated."})
			m.refreshViewport(true)
		}
		return m, nil
	case providerSavedMsg:
		if msg.saved {
			m.providers.open = false
			m.providers.mode = providerModeMenu
			m.providers.input.Blur()
			m.reloadProviderNames()
			m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: msg.message, Timestamp: time.Now()})
			m.refreshViewport(true)
		}
		return m, nil
	case streamStartedMsg:
		m.stream.ch = msg.ch
		m.setNobbyPose(nobby.PoseReading)
		return m, pollStreamCmd(m.stream.ch)
	case streamEventMsg:
		if msg.done {
			m.setNobbyPose(nobby.PoseIdle)
			m.chat.thinking = false
			m.chat.spinnerVerb = ""
			m.chat.spinnerStep = 0
			m.chat.showInferring = false
			if !m.chat.cycleStartedAt.IsZero() {
				m.chat.lastWorkedForSec = int(time.Since(m.chat.cycleStartedAt).Seconds())
			}
			if strings.TrimSpace(m.chat.streamingText) != "" {
				m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: m.chat.streamingText, Timestamp: time.Now()})
			}
			m.chat.streamingText = ""
			m.chat.streamingThought = ""
			m.refreshViewport(true)
			return m, nil
		}
		if msg.event.ErrorText != "" {
			m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: "Error: " + msg.event.ErrorText, Timestamp: time.Now()})
			m.setNobbyPose(nobby.PoseAPIError)
			m.chat.thinking = false
			m.chat.showInferring = false
			m.chat.streamingText = ""
			m.chat.streamingThought = ""
			m.refreshViewport(true)
			return m, nil
		}

		if msg.event.ToolCallStart != nil {
			name := msg.event.ToolCallStart.Name
			args := formatToolArgs(name, msg.event.ToolCallStart.Arguments, m.cwd)

			m.chat.messages = append(m.chat.messages, types.Message{
				Role:      types.RoleTool,
				Text:      fmt.Sprintf("⚙ %s %s", name, args),
				Timestamp: time.Now(),
			})
			m.refreshViewport(true)
			return m, pollStreamCmd(m.stream.ch)
		}

		if msg.event.ToolCallResult != nil {
			name := msg.event.ToolCallResult.Name
			isErr := msg.event.ToolCallResult.IsError
			summary := formatToolResult(name, msg.event.ToolCallResult.Result, isErr, m.layout.width)

			updated := false
			if len(m.chat.messages) > 0 {
				lastIdx := len(m.chat.messages) - 1
				lastMsg := m.chat.messages[lastIdx]
				if lastMsg.Role == types.RoleTool && strings.HasPrefix(lastMsg.Text, "⚙ "+name) {
					icon := "✓"
					if isErr {
						icon = "✗"
					}
					argsPart := strings.TrimPrefix(lastMsg.Text, "⚙ "+name)
					m.chat.messages[lastIdx].Text = fmt.Sprintf("%s %s%s\n  ↳ %s", icon, name, argsPart, summary)
					updated = true
				}
			}

			if !updated {
				icon := "✓"
				if isErr {
					icon = "✗"
				}
				m.chat.messages = append(m.chat.messages, types.Message{
					Role:      types.RoleTool,
					Text:      fmt.Sprintf("%s %s completed\n  ↳ %s", icon, name, summary),
					Timestamp: time.Now(),
				})
			}

			m.refreshViewport(true)
			return m, pollStreamCmd(m.stream.ch)
		}
		if msg.event.ReconnectNote != "" {
			m.setNobbyPose(nobby.PoseAPIErrorReconnect)
			m.chat.messages = append(m.chat.messages, types.Message{
				Role:      types.RoleAssistant,
				Text:      msg.event.ReconnectNote,
				Timestamp: time.Now(),
			})
			m.refreshViewport(true)
			return m, pollStreamCmd(m.stream.ch)
		}
		if msg.event.ReasoningDelta != "" {
			m.setNobbyPose(nobby.PoseThinking)
			m.chat.streamingThought += msg.event.ReasoningDelta
			m.chat.estimatedTokensStream++
			m.chat.estimatedReasoningTokens++
		}
		if msg.event.ContentDelta != "" {
			m.setNobbyPose(nobby.PoseWriting)
			m.chat.showInferring = false
			m.chat.streamingText += msg.event.ContentDelta
			m.chat.estimatedTokensStream++
		}
		if msg.event.RefusalDelta != "" {
			m.chat.estimatedTokensStream++
		}
		if msg.event.ToolDelta != "" {
			m.chat.estimatedTokensStream++
		}
		liveEstimatedTotal := m.chat.usage.PromptTokens + m.chat.estimatedTokensStream
		if liveEstimatedTotal > m.chat.contextTokenFloor {
			m.chat.contextTokenFloor = liveEstimatedTotal
		}
		if msg.event.Usage != nil {
			m.chat.usage = *msg.event.Usage
			if m.chat.usage.TotalTokens > m.chat.contextTokenFloor {
				m.chat.contextTokenFloor = m.chat.usage.TotalTokens
			}
		}
		m.refreshViewport(true)
		return m, pollStreamCmd(m.stream.ch)
	case tea.KeyMsg:
		if m.settings.open {
			return m.handleSettingsKeys(msg)
		}
		if m.providers.open {
			return m.handleProviderKeys(msg)
		}
		if nextModel, cmd, handled := m.handleKeyMsg(msg); handled {
			return nextModel, cmd
		}
	case spinnerTickMsg:
		if !m.chat.thinking || !m.chat.showInferring {
			return m, nil
		}
		m.chat.spinnerStep++
		m.refreshViewport(true)
		return m, spinnerTickCmd(m.settings.values.SpinnerStyle)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.updateCommandSuggestions()
	m.resizeViewport()
	m.refreshViewport(false)
	return m, cmd
}

func formatToolArgs(name string, raw string, cwd string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(trimmed), &args); err != nil {
		if len(trimmed) > 70 {
			return trimmed[:67] + "..."
		}
		return trimmed
	}

	relPath := func(p string) string {
		if p == "" {
			return ""
		}
		if rel, err := filepath.Rel(cwd, p); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
		return p
	}

	switch name {
	case "Write":
		path, _ := args["file_path"].(string)
		content, _ := args["content"].(string)
		base := relPath(path)
		lines := strings.Count(content, "\n")
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			lines++
		}
		return fmt.Sprintf("%s · %d lines", base, lines)

	case "Read":
		path, _ := args["file_path"].(string)
		offsetFloat, hasOffset := args["offset"].(float64)
		limitFloat, hasLimit := args["limit"].(float64)

		base := relPath(path)
		if hasOffset || hasLimit {
			start := 1
			if hasOffset {
				start = int(offsetFloat)
			}
			if hasLimit {
				return fmt.Sprintf("%s · lines %d-%d", base, start, start+int(limitFloat)-1)
			}
			return fmt.Sprintf("%s · from line %d", base, start)
		}
		return base

	case "Glob":
		pattern, _ := args["pattern"].(string)
		path, ok := args["path"].(string)
		if ok && path != "" && path != cwd {
			return fmt.Sprintf("%q in %s", pattern, relPath(path))
		}
		return fmt.Sprintf("%q", pattern)

	case "Grep":
		pattern, _ := args["pattern"].(string)
		path, ok := args["path"].(string)
		glob, _ := args["glob"].(string)
		mode, _ := args["output_mode"].(string)

		res := fmt.Sprintf("%q", pattern)
		if ok && path != "" && path != cwd {
			res += fmt.Sprintf(" in %s", relPath(path))
		}
		if glob != "" {
			res += fmt.Sprintf(" [%s]", glob)
		}
		if mode == "content" {
			res += " (content)"
		} else if mode == "count" {
			res += " (count)"
		}
		return res

	default:
		var parts []string
		for k, v := range args {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		joined := strings.Join(parts, " ")
		if len(joined) > 70 {
			return joined[:67] + "..."
		}
		return joined
	}
}

func formatToolResult(name string, raw string, isErr bool, width int) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "empty result"
	}

	if isErr {
		lines := strings.Split(trimmed, "\n")
		firstLine := strings.TrimPrefix(lines[0], "tool execution failed: ")
		if len(firstLine) > 120 {
			return firstLine[:117] + "..."
		}
		return firstLine
	}

	lines := strings.Count(trimmed, "\n") + 1
	bytes := len(trimmed)

	switch name {
	case "Write":
		var writeData struct {
			Type     string `json:"type"`
			FilePath string `json:"filePath"`
			Diff     string `json:"diff"`
		}
		if json.Unmarshal([]byte(trimmed), &writeData) == nil && writeData.Diff != "" {
			return messages.RenderDiff(writeData.FilePath, writeData.Diff, width, "write")
		}
		if strings.Contains(trimmed, `"type":"create"`) {
			return "File created"
		}
		return "File updated"
	case "Edit":
		var editData struct {
			Type       string `json:"type"`
			FilePath   string `json:"filePath"`
			ReplaceAll bool  `json:"replaceAll"`
			Diff       string `json:"diff"`
		}
		if json.Unmarshal([]byte(trimmed), &editData) == nil && editData.Diff != "" {
			op := "edit"
			if editData.ReplaceAll {
				op = "replace"
			}
			return messages.RenderDiff(editData.FilePath, editData.Diff, width, op)
		}
		return "File edited"
	case "Read":
		return fmt.Sprintf("Read %d lines (%s)", lines, formatBytes(bytes))
	case "Glob", "Grep":
		firstLine := strings.SplitN(trimmed, "\n", 2)[0]
		if strings.HasPrefix(firstLine, "Found ") || strings.HasPrefix(firstLine, "No ") {
			return firstLine
		}
		return fmt.Sprintf("%d results", lines)
	default:
		if lines == 1 && bytes < 80 {
			return trimmed
		}
		return fmt.Sprintf("%d lines, %s", lines, formatBytes(bytes))
	}
}

func formatBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	return fmt.Sprintf("%.1f KB", float64(b)/1024.0)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if len(m.commands.suggestions) > 0 {
		switch msg.String() {
		case "up":
			m.commands.selected = mathutil.Clamp(m.commands.selected-1, 0, len(m.commands.suggestions)-1)
			return m, nil, true
		case "down":
			m.commands.selected = mathutil.Clamp(m.commands.selected+1, 0, len(m.commands.suggestions)-1)
			return m, nil, true
		case "tab":
			m.input.SetValue(m.commands.suggestions[m.commands.selected] + " ")
			m.clearCommandSuggestions()
			m.resizeViewport()
			return m, nil, true
		}
	}

	switch msg.String() {
	case "shift+tab":
		if m.chat.mode == ModeAsk {
			m.chat.mode = ModeCode
		} else {
			m.chat.mode = ModeAsk
		}
		m.resizeViewport()
		return m, nil, true
	case "ctrl+c":
		// Double-press Ctrl+C to quit.
		if m.isPendingConfirmationFor("ctrl+c") {
			m.clearPendingConfirmation()
			return m, tea.Quit, true
		}
		m.setPendingConfirmation("ctrl+c")
		return m, nil, true
	case "esc":
		// During thinking mode, ESC interrupts the stream with confirmation.
		if m.chat.thinking {
			if m.isPendingConfirmationFor("esc") {
				m.chat.interrupted = true
				m.chat.thinking = false
				m.chat.showInferring = false
				close(m.chat.abortChan)
				m.chat.abortChan = nil
				m.setNobbyPose(nobby.PoseIdle)
				if !m.chat.cycleStartedAt.IsZero() {
					m.chat.lastWorkedForSec = int(time.Since(m.chat.cycleStartedAt).Seconds())
				}
				m.clearPendingConfirmation()
				m.refreshViewport(true)
				return m, nil, true
			}
			m.setPendingConfirmation("esc")
			return m, nil, true
		}
		// Outside generation ESC should do nothing (reserved for canceling generation only).
		if m.chat.confirmKey == "esc" {
			m.clearPendingConfirmation()
		}
		return m, nil, true
	case "pgup":
		m.viewport.HalfViewUp()
		return m, nil, true
	case "pgdown":
		m.viewport.HalfViewDown()
		return m, nil, true
	case "up":
		if len(m.commands.suggestions) == 0 {
			m.viewport.LineUp(1)
			return m, nil, true
		}
	case "down":
		if len(m.commands.suggestions) == 0 {
			m.viewport.LineDown(1)
			return m, nil, true
		}
	case "enter":
		if m.chat.thinking {
			return m, nil, true
		}
		if len(m.commands.suggestions) > 0 {
			m.input.SetValue(m.commands.suggestions[m.commands.selected])
			m.clearCommandSuggestions()
		}
		nextModel, cmd := m.executeInput()
		return nextModel, cmd, true
	}
	return m, nil, false
}

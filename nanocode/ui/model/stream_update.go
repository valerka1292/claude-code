package model

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/components/nobby"
	"nanocode/ui/types"
)

func (m Model) handleStreamEvent(msg streamEventMsg) (tea.Model, tea.Cmd) {
	if msg.done {
		return m.finishStreamEvent()
	}

	if msg.event.ErrorText != "" {
		return m.handleStreamError(msg.event.ErrorText)
	}

	if msg.event.ToolCallStart != nil {
		return m.handleToolCallStart(msg.event.ToolCallStart.Name, msg.event.ToolCallStart.Arguments)
	}

	if msg.event.ToolCallResult != nil {
		return m.handleToolCallResult(
			msg.event.ToolCallResult.Name,
			msg.event.ToolCallResult.Result,
			msg.event.ToolCallResult.IsError,
		)
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
}

func (m Model) finishStreamEvent() (tea.Model, tea.Cmd) {
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

func (m Model) handleStreamError(errText string) (tea.Model, tea.Cmd) {
	m.chat.messages = append(m.chat.messages, types.Message{Role: types.RoleAssistant, Text: "Error: " + errText, Timestamp: time.Now()})
	m.setNobbyPose(nobby.PoseAPIError)
	m.chat.thinking = false
	m.chat.showInferring = false
	m.chat.streamingText = ""
	m.chat.streamingThought = ""
	m.refreshViewport(true)
	return m, nil
}

func (m Model) handleToolCallStart(name, rawArgs string) (tea.Model, tea.Cmd) {
	args := formatToolArgs(name, rawArgs, m.cwd)
	m.chat.messages = append(m.chat.messages, types.Message{
		Role:      types.RoleTool,
		Text:      fmt.Sprintf("⚙ %s %s", name, args),
		Timestamp: time.Now(),
	})
	m.refreshViewport(true)
	return m, pollStreamCmd(m.stream.ch)
}

func (m Model) handleToolCallResult(name, rawResult string, isErr bool) (tea.Model, tea.Cmd) {
	summary := formatToolResult(name, rawResult, isErr, m.layout.width)

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

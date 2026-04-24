package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/config"
	"nanocode/ui/model/agent"
	"nanocode/ui/types"
)

func startAgentStreamCmd(provider config.Provider, history []types.Message, settings config.Settings, abortChan <-chan struct{}, mode AgentMode) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan agent.StreamEvent, 64)
		readOnly := mode == ModeAsk
		go agent.RunLoop(convertProvider(provider), convertMessages(history, mode), settings.APITimeoutSeconds, ch, abortChan, readOnly)
		return streamStartedMsg{ch: ch}
	}
}

func convertProvider(provider config.Provider) agent.ProviderConfig {
	return agent.ProviderConfig{
		BaseURL: config.NormalizeBaseURL(provider.BaseURL),
		Model:   provider.Model,
		APIKey:  provider.APIKey,
	}
}

func convertMessages(history []types.Message, mode AgentMode) []agent.APIMessage {
	systemPrompts := buildSystemPrompts(mode)
	result := make([]agent.APIMessage, 0, len(history)+len(systemPrompts))
	for _, prompt := range systemPrompts {
		if strings.TrimSpace(prompt) == "" {
			continue
		}
		result = append(result, agent.APIMessage{
			Role:    "system",
			Content: prompt,
		})
	}
	for _, msg := range history {
		if msg.Role == types.RoleTool {
			continue
		}
		result = append(result, agent.APIMessage{
			Role:    string(msg.Role),
			Content: msg.Text,
		})
	}
	return result
}

func pollStreamCmd(ch <-chan agent.StreamEvent) tea.Cmd {
	return func() tea.Msg {
		next, ok := <-ch
		if !ok {
			return streamEventMsg{done: true}
		}
		return streamEventMsg{event: next}
	}
}

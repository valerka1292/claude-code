package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/config"
	"nanocode/ui/model/agent"
	"nanocode/ui/types"
)

func startAgentStreamCmd(provider config.Provider, history []types.Message, settings config.Settings) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan agent.StreamEvent, 64)
		go agent.RunLoop(convertProvider(provider), convertMessages(history), settings.APITimeoutSeconds, ch)
		return streamStartedMsg{ch: ch}
	}
}

func convertProvider(provider config.Provider) agent.ProviderConfig {
	return agent.ProviderConfig{
		BaseURL: provider.BaseURL,
		Model:   provider.Model,
		APIKey:  provider.APIKey,
	}
}

func convertMessages(history []types.Message) []agent.APIMessage {
	result := make([]agent.APIMessage, 0, len(history)+1)
	result = append(result, agent.APIMessage{
		Role:    "system",
		Content: buildSystemPrompt(),
	})
	for _, msg := range history {
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

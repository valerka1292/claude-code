package model

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/config"
	"nanocode/ui/types"
)

type streamEvent struct {
	ContentDelta   string
	ReasoningDelta string
	Usage          *UsageState
	FinishReason   string
	ErrorText      string
}

type apiMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []apiToolCall `json:"tool_calls,omitempty"`
}

type apiToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function apiToolFunction `json:"function"`
}

type apiToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens           int `json:"prompt_tokens"`
		CompletionTokens       int `json:"completion_tokens"`
		TotalTokens            int `json:"total_tokens"`
		CompletionTokenDetails struct {
			ReasoningTokens int `json:"reasoning_tokens"`
		} `json:"completion_tokens_details"`
	} `json:"usage"`
}

type toolBuffer struct {
	ID        string
	Name      string
	Arguments strings.Builder
}

func startAgentStreamCmd(provider config.Provider, history []types.Message) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan streamEvent, 64)
		go runAgentLoop(provider, history, ch)
		return streamStartedMsg{ch: ch}
	}
}

func runAgentLoop(provider config.Provider, history []types.Message, out chan<- streamEvent) {
	defer close(out)
<<<<<<< 979ixv-codex/implement-streaming-agent-cycle-with-provider-settings
	apiHistory := make([]apiMessage, 0, len(history)+1)
	apiHistory = append(apiHistory, apiMessage{
		Role:    "system",
		Content: buildSystemPrompt(),
	})
=======
	apiHistory := make([]apiMessage, 0, len(history))
>>>>>>> main
	for _, msg := range history {
		apiHistory = append(apiHistory, apiMessage{Role: string(msg.Role), Content: msg.Text})
	}

	client := &http.Client{Timeout: 180 * time.Second}
	for turns := 0; turns < 8; turns++ {
		calls, usage, finishReason, err := streamOneTurn(client, provider, apiHistory, out)
		if err != nil {
			out <- streamEvent{ErrorText: err.Error()}
			return
		}
		if usage != nil {
			out <- streamEvent{Usage: usage}
		}
		if finishReason != "tool_calls" {
			out <- streamEvent{FinishReason: finishReason}
			return
		}

		assistantMsg := apiMessage{Role: "assistant", ToolCalls: calls}
		apiHistory = append(apiHistory, assistantMsg)
		for _, call := range calls {
			toolContent := fmt.Sprintf(`{"error":"tool %q not implemented yet","arguments":%q}`, call.Function.Name, call.Function.Arguments)
			apiHistory = append(apiHistory, apiMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    toolContent,
			})
		}
	}
	out <- streamEvent{ErrorText: "agent loop stopped: maximum turns reached"}
}

func streamOneTurn(client *http.Client, provider config.Provider, messages []apiMessage, out chan<- streamEvent) ([]apiToolCall, *UsageState, string, error) {
	requestBody := map[string]any{
		"model":    provider.Model,
		"messages": messages,
		"stream":   true,
		"stream_options": map[string]any{
			"include_usage": true,
		},
		"tools": []any{},
	}
	raw, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, "", err
	}
	endpoint := config.NormalizeBaseURL(provider.BaseURL) + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return nil, nil, "", fmt.Errorf("provider error: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}

	scanner := bufio.NewScanner(resp.Body)
	toolCalls := map[int]*toolBuffer{}
	var usage *UsageState
	finishReason := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			break
		}
		var chunk chatCompletionChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Usage.TotalTokens > 0 {
			usage = &UsageState{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				ReasoningTokens:  chunk.Usage.CompletionTokenDetails.ReasoningTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		if delta.ReasoningContent != "" {
			out <- streamEvent{ReasoningDelta: delta.ReasoningContent}
		}
		if delta.Content != "" {
			out <- streamEvent{ContentDelta: delta.Content}
		}
		for _, callDelta := range delta.ToolCalls {
			buf, ok := toolCalls[callDelta.Index]
			if !ok {
				buf = &toolBuffer{}
				toolCalls[callDelta.Index] = buf
			}
			if callDelta.ID != "" {
				buf.ID = callDelta.ID
			}
			if callDelta.Function.Name != "" {
				buf.Name = callDelta.Function.Name
			}
			if callDelta.Function.Arguments != "" {
				buf.Arguments.WriteString(callDelta.Function.Arguments)
			}
		}
		if chunk.Choices[0].FinishReason != "" {
			finishReason = chunk.Choices[0].FinishReason
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, usage, finishReason, err
	}

	resultCalls := make([]apiToolCall, 0, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		if tool, ok := toolCalls[i]; ok {
			resultCalls = append(resultCalls, apiToolCall{
				ID:   tool.ID,
				Type: "function",
				Function: apiToolFunction{
					Name:      tool.Name,
					Arguments: tool.Arguments.String(),
				},
			})
		}
	}
	return resultCalls, usage, finishReason, nil
}

func pollStreamCmd(ch <-chan streamEvent) tea.Cmd {
	return func() tea.Msg {
		next, ok := <-ch
		if !ok {
			return streamEventMsg{done: true}
		}
		return streamEventMsg{event: next}
	}
}

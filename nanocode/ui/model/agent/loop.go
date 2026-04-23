package agent

import (
	"fmt"
	"time"

	"nanocode/ui/config"
)

const (
	MaxTurns   = 8
	MaxRetries = 4
)

// RunLoop executes the agent loop, streaming events to the output channel.
func RunLoop(provider config.Provider, messages []APIMessage, timeoutSeconds int, out chan<- StreamEvent) {
	defer close(out)

	client := NewClient(timeoutSeconds)

	apiHistory := make([]APIMessage, 0, len(messages)+1)
	apiHistory = append(apiHistory, APIMessage{
		Role:    "system",
		Content: buildSystemPrompt(),
	})
	for _, msg := range messages {
		apiHistory = append(apiHistory, msg)
	}

	for turns := 0; turns < MaxTurns; turns++ {
		calls, usage, finishReason, err := streamOneTurnWithRetry(client, apiHistory, out)
		if err != nil {
			out <- StreamEvent{ErrorText: err.Error()}
			return
		}
		if usage != nil {
			// Convert internal UsageState to exported type
			out <- StreamEvent{Usage: &UsageState{
				PromptTokens:     usage.PromptTokens,
				CompletionTokens: usage.CompletionTokens,
				ReasoningTokens:  usage.ReasoningTokens,
				TotalTokens:      usage.TotalTokens,
			}}
		}
		if finishReason != "tool_calls" {
			out <- StreamEvent{FinishReason: finishReason}
			return
		}

		assistantMsg := APIMessage{Role: "assistant", ToolCalls: calls}
		apiHistory = append(apiHistory, assistantMsg)
		for _, call := range calls {
			toolContent := fmt.Sprintf(`{"error":"tool %q not implemented yet","arguments":%q}`, call.Function.Name, call.Function.Arguments)
			apiHistory = append(apiHistory, APIMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    toolContent,
			})
		}
	}
	out <- StreamEvent{ErrorText: "agent loop stopped: maximum turns reached"}
}

func streamOneTurnWithRetry(client *Client, messages []APIMessage, out chan<- StreamEvent) ([]APIToolCall, *UsageState, string, error) {
	var (
		calls        []APIToolCall
		usage        *UsageState
		finishReason string
		err          error
	)

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		calls, usage, finishReason, err = client.Stream(StreamConfig{Messages: messages}, out)
		if err == nil {
			return calls, usage, finishReason, nil
		}
		if !IsRetryableError(err) || attempt == MaxRetries {
			return nil, usage, finishReason, err
		}

		backoff := RetryBackoff(attempt)
		out <- StreamEvent{
			ReconnectNote: fmt.Sprintf(
				"API error/timeout. Reconnect %d/%d in %.1fs…",
				attempt+1,
				MaxRetries,
				backoff.Seconds(),
			),
		}
		time.Sleep(backoff)
	}
	return nil, usage, finishReason, err
}

// buildSystemPrompt returns the system prompt for the agent.
func buildSystemPrompt() string {
	return "You are nanocode - autonomous coding agent."
}

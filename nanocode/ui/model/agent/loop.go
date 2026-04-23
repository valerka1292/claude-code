package agent

import (
	"fmt"
	"time"
)

const (
	MaxTurns   = 8
	MaxRetries = 4
)

// RunLoop executes the agent loop, streaming events to the output channel.
func RunLoop(provider ProviderConfig, messages []APIMessage, timeoutSeconds int, out chan<- StreamEvent, abortChan <-chan struct{}) {
	defer close(out)

	client := NewClient(timeoutSeconds)

	apiHistory := make([]APIMessage, 0, len(messages))
	apiHistory = append(apiHistory, messages...)

	for turns := 0; turns < MaxTurns; turns++ {
		calls, usage, finishReason, err := streamOneTurnWithRetry(client, provider, apiHistory, out, abortChan)
		if err != nil {
			out <- StreamEvent{ErrorText: err.Error()}
			return
		}
		if usage != nil {
			out <- StreamEvent{Usage: usage}
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

func streamOneTurnWithRetry(client *Client, provider ProviderConfig, messages []APIMessage, out chan<- StreamEvent, abortChan <-chan struct{}) ([]APIToolCall, *UsageState, string, error) {
	var (
		calls        []APIToolCall
		usage        *UsageState
		finishReason string
		err          error
	)

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		calls, usage, finishReason, err = client.Stream(StreamConfig{
			Provider:  provider,
			Messages:  messages,
			AbortChan: abortChan,
		}, out)
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

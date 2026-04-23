package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"nanocode/ui/config"
)

// StreamEvent represents an event from the streaming API.
type StreamEvent struct {
	ContentDelta   string
	ReasoningDelta string
	Usage          *UsageState
	FinishReason   string
	ErrorText      string
	ReconnectNote  string
}

// UsageState holds token usage information.
type UsageState struct {
	PromptTokens     int
	CompletionTokens int
	ReasoningTokens  int
	TotalTokens      int
}

// APIMessage represents a message in the API format.
type APIMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []APIToolCall `json:"tool_calls,omitempty"`
}

// APIToolCall represents a tool call in the API format.
type APIToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function APIToolFunction `json:"function"`
}

// APIToolFunction represents a tool function in the API format.
type APIToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionChunk represents a chunk from the streaming API.
type ChatCompletionChunk struct {
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

// ToolBuffer accumulates tool call data during streaming.
type ToolBuffer struct {
	ID        string
	Name      string
	Arguments strings.Builder
}

// Client handles HTTP communication with the API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new API client with the given timeout.
func NewClient(timeoutSeconds int) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

// StreamConfig holds configuration for a streaming request.
type StreamConfig struct {
	Provider config.Provider
	Messages []APIMessage
}

// Stream performs a single streaming request and returns tool calls, usage, and finish reason.
func (c *Client) Stream(cfg StreamConfig, out chan<- StreamEvent) ([]APIToolCall, *UsageState, string, error) {
	requestBody := map[string]any{
		"model":    cfg.Provider.Model,
		"messages": cfg.Messages,
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
	endpoint := config.NormalizeBaseURL(cfg.Provider.BaseURL) + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Provider.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return nil, nil, "", fmt.Errorf("provider error: %s - %s", resp.Status, strings.TrimSpace(string(body)))
	}

	scanner := bufio.NewScanner(resp.Body)
	toolCalls := map[int]*ToolBuffer{}
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
		var chunk ChatCompletionChunk
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
			out <- StreamEvent{ReasoningDelta: delta.ReasoningContent}
		}
		if delta.Content != "" {
			out <- StreamEvent{ContentDelta: delta.Content}
		}
		for _, callDelta := range delta.ToolCalls {
			buf, ok := toolCalls[callDelta.Index]
			if !ok {
				buf = &ToolBuffer{}
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

	resultCalls := make([]APIToolCall, 0, len(toolCalls))
	for i := 0; i < len(toolCalls); i++ {
		if tool, ok := toolCalls[i]; ok {
			resultCalls = append(resultCalls, APIToolCall{
				ID:   tool.ID,
				Type: "function",
				Function: APIToolFunction{
					Name:      tool.Name,
					Arguments: tool.Arguments.String(),
				},
			})
		}
	}
	return resultCalls, usage, finishReason, nil
}

// IsRetryableError determines if an error is retryable.
func IsRetryableError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "504")
}

// RetryBackoff returns the backoff duration for a given retry attempt.
func RetryBackoff(attempt int) time.Duration {
	steps := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
	}
	if attempt < 0 {
		return steps[0]
	}
	if attempt >= len(steps) {
		return steps[len(steps)-1]
	}
	return steps[attempt]
}

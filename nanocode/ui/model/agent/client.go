package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
	Provider ProviderConfig
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
	}
	raw, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, "", err
	}
	endpoint := cfg.Provider.BaseURL + "/chat/completions"
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
			if callDelta.Function.Name != "" {
				out <- StreamEvent{ToolDelta: callDelta.Function.Name}
			}
			if callDelta.Function.Arguments != "" {
				out <- StreamEvent{ToolDelta: callDelta.Function.Arguments}
			}
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

package agent

import "strings"

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

// ProviderConfig is a transport-level provider configuration for the agent.
type ProviderConfig struct {
	BaseURL string
	Model   string
	APIKey  string
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

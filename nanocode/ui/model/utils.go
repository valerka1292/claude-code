package model

import (
	"unicode/utf8"

	"nanocode/ui/types"
)

func estimateTokens(text string) int {
	runes := utf8.RuneCountInString(text)
	if runes <= 0 {
		return 0
	}
	estimated := runes / 4
	if estimated < 1 {
		return 1
	}
	return estimated
}

func estimatePromptTokens(history []types.Message) int {
	total := estimateTokens(buildSystemPrompt())
	for _, msg := range history {
		total += estimateTokens(msg.Text)
	}
	return total
}

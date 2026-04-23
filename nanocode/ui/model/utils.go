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
	// Более точная оценка: 1 токен ≈ 4 руны, но без завышения для коротких строк
	estimated := float64(runes) / 4.0
	if estimated < 0.25 {
		// Очень короткие дельты (< 1 руны) считаем как 0.25 токена
		return 0
	}
	return int(estimated + 0.5) // округление вместо отсечения
}

func estimatePromptTokens(history []types.Message) int {
	total := estimateTokens(buildSystemPrompt())
	for _, msg := range history {
		total += estimateTokens(msg.Text)
	}
	return total
}

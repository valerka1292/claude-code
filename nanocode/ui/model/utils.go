package model

import (
	"unicode/utf8"
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

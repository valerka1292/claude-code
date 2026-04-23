package model

import (
	"nanocode/internal/mathutil"
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

func clampInt(v, lo, hi int) int {
	return mathutil.Clamp(v, lo, hi)
}

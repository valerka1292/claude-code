package model

import "unicode/utf8"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

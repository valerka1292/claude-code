package model

import (
	"fmt"
	"strconv"
	"strings"
)

func formatCompact(value int) string {
	if value < 1000 {
		return fmt.Sprintf("%d", value)
	}
	return fmt.Sprintf("%.1fk", float64(value)/1000.0)
}

// formatDuration formats milliseconds into human-readable duration
// Matches Claude Code format: Xm Ys for minutes, Xs for seconds
func formatDuration(ms int) string {
	if ms < 60000 {
		seconds := ms / 1000
		if seconds == 0 {
			return "0s"
		}
		return fmt.Sprintf("%ds", seconds)
	}

	minutes := ms / 60000
	seconds := (ms % 60000) / 1000

	if seconds == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

func parseContextSize(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	return strconv.Atoi(value)
}

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

func parseContextSize(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	return strconv.Atoi(value)
}

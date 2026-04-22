package spinner

import (
	"fmt"
	"math/rand"
	"strings"
)

// Nano-themed status verbs.
var verbs = []string{
	"Assembling nanoblocks",
	"Growing lattice",
	"Synthesizing structure",
	"Stitching microcode",
	"Aligning bits",
}

// Materialization animation: empty -> dense.
var blockFrames = []rune{'░', '▒', '▓', '█'}

func RandomVerb() string {
	return verbs[rand.Intn(len(verbs))]
}

func Status(frame int, verb string) string {
	phase := frame % len(blockFrames)
	barWidth := 6
	filled := (frame % (barWidth + 1))
	if filled == 0 {
		filled = 1
	}

	bar := strings.Repeat(string(blockFrames[phase]), filled)
	pad := strings.Repeat(" ", max(0, barWidth-filled))
	return fmt.Sprintf("[%s%s] %s...", bar, pad, verb)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

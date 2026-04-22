package spinner

import (
	"fmt"
	"math/rand"
	"time"

	"nanocode/ui/config"
)

var verbs = []string{
	"Assembling nanoblocks",
	"Growing lattice",
	"Synthesizing structure",
	"Stitching microcode",
	"Aligning bits",
}

var hexFrames = []string{
	"⬢ ⬡ ⬡",
	"⬡ ⬢ ⬡",
	"⬡ ⬡ ⬢",
	"⬡ ⬢ ⬡",
}

var circleFrames = []string{
	"·",
	"◦",
	"○",
	"◦",
}

func RandomVerb() string {
	return verbs[rand.Intn(len(verbs))]
}

func Frame(style string, i int) string {
	frames := circleFrames
	if style == config.SpinnerHexagons {
		frames = hexFrames
	}
	return frames[i%len(frames)]
}

func Interval(style string) time.Duration {
	if style == config.SpinnerHexagons {
		return 250 * time.Millisecond
	}
	return 200 * time.Millisecond
}

func Status(frame int, verb string, style string) string {
	return fmt.Sprintf("%s %s...", Frame(style, frame), verb)
}

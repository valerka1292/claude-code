package spinner

import (
	"fmt"
	"math/rand"
	"time"
)

var verbs = []string{
	"Analyzing",
	"Crafting",
	"Inferring",
	"Lollygagging",
	"Musing",
	"Orchestrating",
	"Perusing",
	"Pondering",
}

var frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func RandomVerb() string {
	return verbs[rand.Intn(len(verbs))]
}

func Frame(i int) string {
	if len(frames) == 0 {
		return "."
	}
	return frames[i%len(frames)]
}

func Tick() time.Time {
	return time.Now()
}

func Status(frame int, verb string) string {
	return fmt.Sprintf("%s %s...", Frame(frame), verb)
}

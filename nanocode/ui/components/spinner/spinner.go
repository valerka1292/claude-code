package spinner

import (
	"fmt"
	"math/rand"
	"time"

	"nanocode/ui/config"
)

// NanoCode-themed spinner verbs - organized by category
// Keeping brand identity while matching Claude's spirit (alliteration, playful terms)
var verbs = []string{
	// Nano-Tech (Micro-world & Physics)
	"Assembling",
	"Atomizing",
	"Bonding",
	"Colliding",
	"Condensing",
	"Crystallizing",
	"Etching",
	"Ionizing",
	"Layering",
	"Magnetizing",
	"Nanifying",
	"Nucleating",
	"Polarizing",
	"Quantumizing",
	"Synthesizing",

	// Nerd/Hacker Culture
	"Baffling",
	"Bit-flipping",
	"Bootstrapping",
	"Compiling",
	"Debugging",
	"Deploying",
	"Gitifying",
	"Hashing",
	"Hydrating",
	"Indexing",
	"Linting",
	"Optimizing",
	"Parsing",
	"Refactoring",
	"Reticulating",
	"Transpiling",

	// Absurd & Playful
	"Bamboozling",
	"Blooping",
	"Brainstorming",
	"Caffeinating",
	"Chai-sipping",
	"Deep-frying",
	"Hallucinating",
	"Meditating",
	"Napping",
	"Overthinking",
	"Pondering",
	"Rubber-ducking",
	"Shrugging",
	"Vibing",

	// Nano-Action (Dynamic verbs)
	"Accelerating",
	"Beaming",
	"Blasting",
	"Cascading",
	"Igniting",
	"Orbiting",
	"Sprinting",
	"Warping",
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

func Indicator(style string, frame int) string {
	return Frame(style, frame)
}

func StaticIndicator(style string) string {
	if style == config.SpinnerHexagons {
		return "⬢"
	}
	return "○"
}

package nobby

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/theme"
)

type Pose string

const (
	PoseIdle      Pose = "idle"
	PoseThinking  Pose = "thinking"
	PoseWorking   Pose = "working"
	PoseSearching Pose = "searching"
	PoseSuccess   Pose = "success"
	PoseError     Pose = "error"
	PoseWarning   Pose = "warning"
	PoseReading   Pose = "reading"
)

type AnimationFrame struct {
	Antenna string
	Face    string
	BodyL   string
	BodyC   string
	BodyR   string
	Legs    string
	Color   lipgloss.Color
}

type animation struct {
	Frames []AnimationFrame
	Speed  time.Duration
	Per    []time.Duration
}

var animations = map[Pose]animation{
	PoseIdle: {
		Frames: []AnimationFrame{
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [- -]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Per: []time.Duration{2400 * time.Millisecond, 100 * time.Millisecond},
	},
	PoseThinking: {
		Frames: []AnimationFrame{
			{Antenna: "   /     ", Face: "  [~ ~]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    –    ", Face: "  [~ ~]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "     \\   ", Face: "  [~ ~]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    –    ", Face: "  [~ ~]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 200 * time.Millisecond,
	},
	PoseWorking: {
		Frames: []AnimationFrame{
			{Antenna: "    •    ", Face: "  [o  ]  ", BodyL: " -", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [  o]  ", BodyL: "  ", BodyC: "█████", BodyR: "- ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 250 * time.Millisecond,
	},
	PoseSearching: {
		Frames: []AnimationFrame{
			{Antenna: "    ●    ", Face: "  [◉ ◉]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    ○    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    ●    ", Face: "  [◉  ]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    ○    ", Face: "  [  ◉]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 300 * time.Millisecond,
	},
	PoseSuccess: {
		Frames: []AnimationFrame{
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "  █   █  "},
			{Antenna: "    •    ", Face: "  [^ ^]  ", BodyL: " \\", BodyC: "█████", BodyR: "/ ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [◡ ◡]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 200 * time.Millisecond,
	},
	PoseError: {
		Frames: []AnimationFrame{
			{Antenna: "    ✗    ", Face: "  [X X]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: lipgloss.Color("#ff5f5f")},
			{Antenna: "    ✗    ", Face: "  [x x]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "  █   █  ", Color: lipgloss.Color("#ff5f5f")},
		},
		Speed: 300 * time.Millisecond,
	},
	PoseWarning: {
		Frames: []AnimationFrame{
			{Antenna: "    !    ", Face: "  [○ ○]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: theme.SecondaryAccent},
			{Antenna: "         ", Face: "  [○ ○]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: theme.SecondaryAccent},
		},
		Speed: 400 * time.Millisecond,
	},
	PoseReading: {
		Frames: []AnimationFrame{
			{Antenna: "    •    ", Face: "  [´  ]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [>  ]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [.  ]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [  <]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 350 * time.Millisecond,
	},
}

func FrameFor(pose Pose, step int) AnimationFrame {
	a := animations[pose]
	f := a.Frames[step%len(a.Frames)]
	if f.Color == "" {
		f.Color = theme.PrimaryAccent
	}
	return f
}

func DurationFor(pose Pose, step int) time.Duration {
	a := animations[pose]
	if len(a.Per) > 0 {
		return a.Per[step%len(a.Per)]
	}
	if a.Speed <= 0 {
		return 250 * time.Millisecond
	}
	return a.Speed
}

func Render(pose Pose, step int) string {
	f := FrameFor(pose, step)
	lineStyle := lipgloss.NewStyle().Foreground(f.Color)
	bodyStyle := lipgloss.NewStyle().Foreground(f.Color).Background(theme.SurfaceBackground)

	lines := []string{
		lineStyle.Render(f.Antenna),
		lineStyle.Render(f.Face),
		lineStyle.Render(f.BodyL) + bodyStyle.Render(f.BodyC) + lineStyle.Render(f.BodyR),
		lineStyle.Render(f.Legs),
	}

	container := lipgloss.NewStyle().
		Background(theme.SurfaceBackground).
		Padding(0, 1)
	return container.Render(strings.Join(lines, "\n"))
}

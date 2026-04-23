package nobby

import (
	"strings"
	"time"

	"nanocode/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

type Pose string

const (
	PoseIdle              Pose = "idle"
	PoseReading           Pose = "reading"
	PoseThinking          Pose = "thinking"
	PoseWriting           Pose = "writing"
	PoseToolCalling       Pose = "tool-calling"
	PoseToolSuccess       Pose = "tool-success"
	PoseToolError         Pose = "tool-error"
	PoseAPIErrorReconnect Pose = "api-error-reconnect"
	PoseAPIError          Pose = "api-error"
)

type AnimationFrame struct {
	Antenna string
	Face    string
	BodyL   string
	BodyC   string
	BodyR   string
	Legs    string
	Color   string
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
		Per: []time.Duration{2500 * time.Millisecond, 150 * time.Millisecond},
	},
	PoseReading: {
		Frames: []AnimationFrame{
			{Antenna: "    ▲    ", Face: "  [◉  ]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    ▲    ", Face: "  [◉ ◉]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    ▲    ", Face: "  [  ◉]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    ▲    ", Face: "  [◉ ◉]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 250 * time.Millisecond,
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
	PoseWriting: {
		Frames: []AnimationFrame{
			{Antenna: "    •    ", Face: "  [o  ]  ", BodyL: " -", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [  o]  ", BodyL: "  ", BodyC: "█████", BodyR: "- ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Speed: 300 * time.Millisecond,
	},
	PoseToolCalling: {
		Frames: []AnimationFrame{
			{Antenna: "    ●    ", Face: "  [• •]  ", BodyL: " -", BodyC: "█████", BodyR: "- ", Legs: "   █ █   "},
			{Antenna: "    ○    ", Face: "  [• •]  ", BodyL: " -", BodyC: "█████", BodyR: "- ", Legs: "   █ █   "},
		},
		Speed: 300 * time.Millisecond,
	},
	PoseToolSuccess: {
		Frames: []AnimationFrame{
			{Antenna: "    •    ", Face: "  [o o]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "  █   █  "},
			{Antenna: "    •    ", Face: "  [^ ^]  ", BodyL: " \\", BodyC: "█████", BodyR: "/ ", Legs: "   █ █   "},
			{Antenna: "    •    ", Face: "  [◡ ◡]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   "},
		},
		Per: []time.Duration{150 * time.Millisecond, 300 * time.Millisecond, 200 * time.Millisecond},
	},
	PoseToolError: {
		Frames: []AnimationFrame{
			{Antenna: "    ✗    ", Face: "  [X X]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: "#ff5f5f"},
			{Antenna: "    ✗    ", Face: "  [x x]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "  █   █  ", Color: "#ff5f5f"},
		},
		Speed: 400 * time.Millisecond,
	},
	PoseAPIErrorReconnect: {
		Frames: []AnimationFrame{
			{Antenna: "    │    ", Face: "  [@ @]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: "#ff5f5f"},
			{Antenna: "    ╱    ", Face: "  [@ @]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: "#ff5f5f"},
			{Antenna: "    ─    ", Face: "  [@ @]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: "#ff5f5f"},
			{Antenna: "    ╲    ", Face: "  [@ @]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: "#ff5f5f"},
		},
		Speed: 200 * time.Millisecond,
	},
	PoseAPIError: {
		Frames: []AnimationFrame{
			{Antenna: "    ✗    ", Face: "  [x x]  ", BodyL: "  ", BodyC: "█████", BodyR: "  ", Legs: "   █ █   ", Color: "#ff5f5f"},
			{Antenna: "    ·    ", Face: "  [. .]  ", BodyL: "  ", BodyC: "░░░░░", BodyR: "  ", Legs: "   █ █   ", Color: string(theme.MutedText)},
		},
		Speed: 600 * time.Millisecond,
	},
}

func FrameFor(pose Pose, step int) AnimationFrame {
	a := animations[pose]
	f := a.Frames[step%len(a.Frames)]
	if f.Color == "" {
		f.Color = string(theme.PrimaryAccent)
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

	// 1. Убрали Background(theme.SurfaceBackground)
	bodyStyle := lipgloss.NewStyle().Foreground(f.Color)

	lines := []string{
		lineStyle.Render(f.Antenna),
		lineStyle.Render(f.Face),
		lineStyle.Render(f.BodyL) + bodyStyle.Render(f.BodyC) + lineStyle.Render(f.BodyR),
		lineStyle.Render(f.Legs),
	}

	container := lipgloss.NewStyle().Padding(0, 1)
	return container.Render(strings.Join(lines, "\n"))
}

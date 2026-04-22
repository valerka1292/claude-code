package model

import (
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/components/nobby"
	"nanocode/ui/config"
	"nanocode/ui/types"
)

type spinnerTickMsg time.Time

type assistantReplyMsg struct{}

type spinnerChangedMsg struct{ saved bool }

type nobbyTickMsg time.Time

type command struct {
	Name string
}

var availableCommands = []command{{Name: "/settings"}}

type LayoutState struct {
	width             int
	height            int
	viewportMaxHeight int
	viewportTop       int
	scrollbarDragging bool
}

type ChatState struct {
	messages    []types.Message
	thinking    bool
	spinnerVerb string
	spinnerStep int
}

type CommandState struct {
	suggestions []string
	selected    int
}

type SettingsState struct {
	open          bool
	selectedStyle int
	values        config.Settings
}

type Model struct {
	cwd       string
	input     textinput.Model
	viewport  viewport.Model
	nobbyPose nobby.Pose
	nobbyStep int

	layout   LayoutState
	chat     ChatState
	commands CommandState
	settings SettingsState
}

func New() Model {
	rand.Seed(time.Now().UnixNano())
	in := textinput.New()
	in.Focus()
	in.Prompt = ""
	in.Placeholder = "Type a message..."
	in.CharLimit = 1000
	in.Width = 80

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "./"
	}

	cfg, err := config.LoadSettings()
	if err != nil {
		cfg = config.DefaultSettings()
	}

	vp := viewport.New(80, 10)

	m := Model{
		cwd:       cwd,
		input:     in,
		viewport:  vp,
		nobbyPose: nobby.PoseIdle,
		nobbyStep: 0,
		layout:    LayoutState{},
		chat:      ChatState{},
		commands:  CommandState{},
		settings: SettingsState{
			values:        cfg,
			selectedStyle: spinnerIndexFor(cfg.SpinnerStyle),
		},
	}
	m.refreshViewport(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, nobbyTickCmd(m.nobbyPose, m.nobbyStep))
}

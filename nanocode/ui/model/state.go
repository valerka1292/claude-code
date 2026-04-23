package model

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/components/nobby"
	"nanocode/ui/config"
	"nanocode/ui/model/agent"
	"nanocode/ui/model/provider"
	"nanocode/ui/types"
)

type providerMode string

const (
	providerModeMenu       providerMode = "menu"
	providerModeCreate     providerMode = "create"
	providerModeSelect     providerMode = "select"
	providerModeEditPick   providerMode = "edit_pick"
	providerModeEditField  providerMode = "edit_field"
	providerModeEditInput  providerMode = "edit_input"
	providerModeDelete     providerMode = "delete"
	providerModeInputValue providerMode = "input_value"
)

type spinnerTickMsg time.Time
type spinnerChangedMsg struct{ saved bool }
type providerSavedMsg struct {
	saved   bool
	message string
}
type nobbyTickMsg time.Time

type streamStartedMsg struct{ ch <-chan agent.StreamEvent }
type streamEventMsg struct {
	event agent.StreamEvent
	done  bool
}

type command struct {
	Name string
}

var availableCommands = []command{{Name: "/settings"}, {Name: "/provider"}}

type LayoutState struct {
	width             int
	height            int
	viewportMaxHeight int
	viewportTop       int
	scrollbarDragging bool
}

type ChatState struct {
	messages         []types.Message
	thinking         bool
	spinnerVerb      string
	spinnerStep      int
	streamingText    string
	streamingThought string
	usage            agent.UsageState
	cycleStartedAt   time.Time
	liveDownTokens   int
	showInferring    bool
	lastWorkedForSec int
}

type CommandState struct {
	suggestions []string
	selected    int
}

type SettingsState struct {
	open        bool
	selectedRow int
	values      config.Settings
}

type ProviderState struct {
	open               bool
	mode               providerMode
	menuIndex          int
	selectedProvider   int
	selectedField      provider.Field
	input              textinput.Model
	data               config.ProvidersFile
	names              []string
	form               provider.Form
	inputPrompt        string
	inputField         provider.Field
	currentProviderRef string
}

type StreamState struct {
	ch <-chan agent.StreamEvent
}

type Model struct {
	cwd       string
	input     textinput.Model
	viewport  viewport.Model
	nobbyPose nobby.Pose
	nobbyStep int

	layout    LayoutState
	chat      ChatState
	commands  CommandState
	settings  SettingsState
	providers ProviderState
	stream    StreamState
}

func New() Model {
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

	providers, err := config.LoadProviders()
	if err != nil {
		providers = config.ProvidersFile{Providers: map[string]config.Provider{}}
	}

	panelInput := textinput.New()
	panelInput.Prompt = ""
	panelInput.CharLimit = 2048
	panelInput.Width = 48

	vp := viewport.New(80, 10)
	m := Model{
		cwd:       cwd,
		input:     in,
		viewport:  vp,
		nobbyPose: nobby.PoseIdle,
		layout:    LayoutState{},
		chat:      ChatState{},
		commands:  CommandState{},
		settings: SettingsState{
			values:      cfg,
			selectedRow: 0,
		},
		providers: ProviderState{
			mode:  providerModeMenu,
			input: panelInput,
			data:  providers,
		},
	}
	m.providers.names = config.ProviderNames(m.providers.data)
	m.refreshViewport(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, nobbyTickCmd(m.nobbyPose, m.nobbyStep))
}

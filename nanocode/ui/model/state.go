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
type spinnerChangedMsg struct{ saved bool }
type providerSavedMsg struct {
	saved   bool
	message string
}
type nobbyTickMsg time.Time

type streamStartedMsg struct{ ch <-chan streamEvent }
type streamEventMsg struct {
	event streamEvent
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

type UsageState struct {
	PromptTokens     int
	CompletionTokens int
	ReasoningTokens  int
	TotalTokens      int
}

type ChatState struct {
	messages         []types.Message
	thinking         bool
	spinnerVerb      string
	spinnerStep      int
	streamingText    string
	streamingThought string
	usage            UsageState
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

type ProviderState struct {
	open               bool
	mode               providerMode
	menuIndex          int
	selectedProvider   int
	selectedField      int
	input              textinput.Model
	data               config.ProvidersFile
	names              []string
	formName           string
	formBaseURL        string
	formModel          string
	formAPIKey         string
	formContextSize    string
	inputPrompt        string
	inputField         string
	inputProviderName  string
	inputValuePrefill  string
	currentProviderRef string
}

type StreamState struct {
	ch <-chan streamEvent
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
			values:        cfg,
			selectedStyle: spinnerIndexFor(cfg.SpinnerStyle),
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

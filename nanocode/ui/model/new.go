package model

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/components/nobby"
	"nanocode/ui/config"
	"nanocode/ui/model/provider"
)

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

func providerFieldCount() int {
	return int(provider.FieldContextSize) + 1
}

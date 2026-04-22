package model

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/components/header"
	"nanocode/ui/components/messages"
	"nanocode/ui/components/prompt"
	"nanocode/ui/components/spinner"
	"nanocode/ui/config"
	"nanocode/ui/types"
)

type spinnerTickMsg time.Time

type assistantReplyMsg struct{}

type spinnerChangedMsg struct{ saved bool }

type command struct {
	Name string
}

var availableCommands = []command{{Name: "/settings"}}

type Model struct {
	width       int
	height      int
	cwd         string
	input       textinput.Model
	viewport    viewport.Model
	messages    []types.Message
	thinking    bool
	spinnerVerb string
	spinnerStep int
	settings    config.Settings

	commandSuggestions []string
	commandIndex       int
	settingsOpen       bool
	settingsIndex      int
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
		cwd:                cwd,
		input:              in,
		viewport:           vp,
		settings:           cfg,
		commandIndex:       0,
		settingsIndex:      spinnerIndexFor(cfg.SpinnerStyle),
		settingsOpen:       false,
		spinnerStep:        0,
		commandSuggestions: nil,
	}
	m.refreshViewport(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(10, msg.Width-6)
		m.resizeViewport()
		m.refreshViewport(false)
		return m, nil

	case spinnerChangedMsg:
		if msg.saved {
			m.messages = append(m.messages, types.Message{
				Role: types.RoleAssistant,
				Text: "Settings saved: spinner style updated.",
			})
			m.refreshViewport(true)
		}
		return m, nil

	case tea.KeyMsg:
		if m.settingsOpen {
			return m.handleSettingsKeys(msg)
		}

		if len(m.commandSuggestions) > 0 {
			switch msg.String() {
			case "up":
				m.commandIndex = clamp(m.commandIndex-1, 0, len(m.commandSuggestions)-1)
				return m, nil
			case "down":
				m.commandIndex = clamp(m.commandIndex+1, 0, len(m.commandSuggestions)-1)
				return m, nil
			case "tab":
				m.input.SetValue(m.commandSuggestions[m.commandIndex] + " ")
				m.clearCommandSuggestions()
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "pgup":
			m.viewport.HalfViewUp()
			return m, nil
		case "pgdown":
			m.viewport.HalfViewDown()
			return m, nil
		case "up":
			if len(m.commandSuggestions) == 0 {
				m.viewport.LineUp(1)
				return m, nil
			}
		case "down":
			if len(m.commandSuggestions) == 0 {
				m.viewport.LineDown(1)
				return m, nil
			}
		case "enter":
			if m.thinking {
				return m, nil
			}
			if len(m.commandSuggestions) > 0 {
				selected := m.commandSuggestions[m.commandIndex]
				m.input.SetValue(selected)
				m.clearCommandSuggestions()
				return m.executeInput()
			}
			return m.executeInput()
		}

	case spinnerTickMsg:
		if !m.thinking {
			return m, nil
		}
		m.spinnerStep++
		m.refreshViewport(true)
		return m, spinnerTickCmd(m.settings.SpinnerStyle)

	case assistantReplyMsg:
		if !m.thinking {
			return m, nil
		}
		userText := ""
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == types.RoleUser {
				userText = m.messages[i].Text
				break
			}
		}
		m.messages = append(m.messages, types.Message{
			Role:      types.RoleAssistant,
			Text:      fmt.Sprintf("Got it: %q. This is a mock nanocode response after a 2-second wait.", userText),
			Timestamp: time.Now(),
		})
		m.thinking = false
		m.spinnerStep = 0
		m.spinnerVerb = ""
		m.refreshViewport(true)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.updateCommandSuggestions()
	m.refreshViewport(false)
	return m, cmd
}

func (m Model) executeInput() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.input.Value())
	if text == "" {
		return m, nil
	}

	if text == "/settings" {
		m.settingsOpen = true
		m.settingsIndex = spinnerIndexFor(m.settings.SpinnerStyle)
		m.input.SetValue("")
		m.clearCommandSuggestions()
		return m, nil
	}

	m.messages = append(m.messages, types.Message{Role: types.RoleUser, Text: text, Timestamp: time.Now()})
	m.input.SetValue("")
	m.clearCommandSuggestions()
	m.thinking = true
	m.spinnerStep = 0
	m.spinnerVerb = spinner.RandomVerb()
	m.refreshViewport(true)
	return m, tea.Batch(spinnerTickCmd(m.settings.SpinnerStyle), mockReplyCmd())
}

func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.settingsOpen = false
		return m, nil
	case "up":
		m.settingsIndex = clamp(m.settingsIndex-1, 0, 1)
		return m, nil
	case "down":
		m.settingsIndex = clamp(m.settingsIndex+1, 0, 1)
		return m, nil
	case "enter":
		style := spinnerStyleFor(m.settingsIndex)
		m.settings.SpinnerStyle = style
		m.settingsOpen = false
		m.refreshViewport(false)
		return m, func() tea.Msg {
			if err := config.SaveSettings(m.settings); err != nil {
				return spinnerChangedMsg{saved: false}
			}
			return spinnerChangedMsg{saved: true}
		}
	}
	return m, nil
}

func (m *Model) updateCommandSuggestions() {
	value := strings.TrimSpace(m.input.Value())
	if strings.HasPrefix(value, "/") && !strings.Contains(value, " ") {
		m.commandSuggestions = m.commandSuggestions[:0]
		for _, cmd := range availableCommands {
			if strings.HasPrefix(cmd.Name, value) {
				m.commandSuggestions = append(m.commandSuggestions, cmd.Name)
			}
		}
		if len(m.commandSuggestions) == 0 {
			m.commandSuggestions = nil
			m.commandIndex = 0
			return
		}
		m.commandIndex = clamp(m.commandIndex, 0, len(m.commandSuggestions)-1)
		return
	}
	m.clearCommandSuggestions()
}

func (m *Model) clearCommandSuggestions() {
	m.commandSuggestions = nil
	m.commandIndex = 0
}

func (m *Model) resizeViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}
	headerHeight := lipgloss.Height(header.View(m.cwd))
	inputHeight := lipgloss.Height(prompt.InputBar(m.input.View(), m.width))
	footerHeight := lipgloss.Height(prompt.Footer())
	reserved := headerHeight + inputHeight + footerHeight + 3
	if len(m.commandSuggestions) > 0 {
		reserved += lipgloss.Height(prompt.CommandSuggestions(m.width, m.commandSuggestions, m.commandIndex))
	}
	if m.settingsOpen {
		reserved += lipgloss.Height(prompt.SettingsPanel(m.width, m.settingsIndex, m.settings.SpinnerStyle))
	}
	vHeight := m.height - reserved
	if vHeight < 6 {
		vHeight = 6
	}
	m.viewport.Width = m.width
	m.viewport.Height = vHeight
}

func (m *Model) refreshViewport(forceBottom bool) {
	if m.width == 0 {
		return
	}
	spinnerLine := ""
	if m.thinking {
		spinnerLine = spinner.Status(m.spinnerStep, m.spinnerVerb, m.settings.SpinnerStyle)
	}
	wasBottom := m.viewport.AtBottom()
	content := messages.View(m.messages, m.width, spinnerLine)
	m.viewport.SetContent(content)
	if forceBottom || wasBottom {
		m.viewport.GotoBottom()
	}
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	headerView := header.View(m.cwd)
	inputView := prompt.InputBar(m.input.View(), m.width)
	parts := []string{headerView, "", m.viewport.View(), inputView}

	if len(m.commandSuggestions) > 0 {
		parts = append(parts, prompt.CommandSuggestions(m.width, m.commandSuggestions, m.commandIndex))
	}
	if m.settingsOpen {
		parts = append(parts, prompt.SettingsPanel(m.width, m.settingsIndex, m.settings.SpinnerStyle))
	}

	parts = append(parts, prompt.Footer())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func spinnerTickCmd(style string) tea.Cmd {
	return tea.Tick(spinner.Interval(style), func(t time.Time) tea.Msg {
		return spinnerTickMsg(t)
	})
}

func mockReplyCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return assistantReplyMsg{}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func spinnerStyleFor(index int) string {
	if index == 1 {
		return config.SpinnerCircles
	}
	return config.SpinnerHexagons
}

func spinnerIndexFor(style string) int {
	if style == config.SpinnerCircles {
		return 1
	}
	return 0
}

package model

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nanocode/ui/components/header"
	"nanocode/ui/components/messages"
	"nanocode/ui/components/prompt"
	"nanocode/ui/components/spinner"
	"nanocode/ui/types"
)

type spinnerTickMsg time.Time

type assistantReplyMsg struct{}

type Model struct {
	width       int
	height      int
	cwd         string
	input       textinput.Model
	messages    []types.Message
	thinking    bool
	spinnerVerb string
	spinnerStep int
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

	return Model{
		cwd:   cwd,
		input: in,
	}
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
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.thinking {
				return m, nil
			}
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			m.messages = append(m.messages, types.Message{Role: types.RoleUser, Text: text, Timestamp: time.Now()})
			m.input.SetValue("")
			m.thinking = true
			m.spinnerStep = 0
			m.spinnerVerb = spinner.RandomVerb()
			return m, tea.Batch(spinnerTickCmd(), mockReplyCmd())
		}

	case spinnerTickMsg:
		if !m.thinking {
			return m, nil
		}
		m.spinnerStep++
		return m, spinnerTickCmd()

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
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	headerView := header.View(m.cwd)
	headerHeight := lipgloss.Height(headerView)

	inputView := prompt.InputBar(m.input.View(), m.width)
	footerView := prompt.Footer()
	bottomHeight := lipgloss.Height(inputView) + lipgloss.Height(footerView)

	messagesHeight := m.height - headerHeight - bottomHeight - 3
	if messagesHeight < 5 {
		messagesHeight = 5
	}

	spinnerLine := ""
	if m.thinking {
		spinnerLine = spinner.Status(m.spinnerStep, m.spinnerVerb)
	}

	messagesView := messages.View(m.messages, m.width, messagesHeight, spinnerLine)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerView,
		"",
		messagesView,
		inputView,
		footerView,
	)
}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
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

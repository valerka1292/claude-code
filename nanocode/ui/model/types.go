package model

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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
	messages              []types.Message
	thinking              bool
	spinnerVerb           string
	spinnerStep           int
	streamingText         string
	streamingThought      string
	usage                 agent.UsageState
	cycleStartedAt        time.Time
	estimatedTokensStream int
	showInferring         bool
	lastWorkedForSec      int
	interrupted           bool
	// Double-press ESC handling
	escapePressTime    time.Time
	escapePending      bool
	abortChan          chan struct{}
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

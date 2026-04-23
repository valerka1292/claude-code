package model

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"nanocode/ui/components/nobby"
)

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

package theme

import "github.com/charmbracelet/lipgloss"

const (
	PrimaryAccentHex   = "#FBFA56"
	SecondaryAccentHex = "#EBDC2F"
	AccentContrastHex  = "#181818"

	AppBackgroundHex     = "#181818"
	SurfaceBackgroundHex = "#333333"

	PrimaryTextHex = "#D4D4D4"
	MutedTextHex   = "#808080"

	ModeAskHex = "#8BE9FD"
)

var (
	PrimaryAccent     = lipgloss.Color(PrimaryAccentHex)
	SecondaryAccent   = lipgloss.Color(SecondaryAccentHex)
	AccentContrast    = lipgloss.Color(AccentContrastHex)
	AppBackground     = lipgloss.Color(AppBackgroundHex)
	SurfaceBackground = lipgloss.Color(SurfaceBackgroundHex)
	PrimaryText       = lipgloss.Color(PrimaryTextHex)
	MutedText         = lipgloss.Color(MutedTextHex)

	ModeAsk           = lipgloss.Color(ModeAskHex)
)

package model

import "strings"

func (m *Model) updateCommandSuggestions() {
	value := strings.TrimSpace(m.input.Value())
	if strings.HasPrefix(value, "/") && !strings.Contains(value, " ") {
		m.commands.suggestions = m.commands.suggestions[:0]
		for _, cmd := range availableCommands {
			if strings.HasPrefix(cmd.Name, value) {
				m.commands.suggestions = append(m.commands.suggestions, cmd.Name)
			}
		}
		if len(m.commands.suggestions) == 0 {
			m.commands.suggestions = nil
			m.commands.selected = 0
			return
		}
		m.commands.selected = clamp(m.commands.selected, 0, len(m.commands.suggestions)-1)
		return
	}
	m.clearCommandSuggestions()
}

func (m *Model) clearCommandSuggestions() {
	m.commands.suggestions = nil
	m.commands.selected = 0
}

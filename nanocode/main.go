package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"nanocode/ui/model"
)

func main() {
	p := tea.NewProgram(
		model.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nanocode failed: %v\n", err)
		os.Exit(1)
	}
}

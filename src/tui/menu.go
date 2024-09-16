package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func StartApp() {
	// init the program with the initial model
	p := tea.NewProgram(InitialModel())

	// run the program
	_, err := p.Run()
	if err != nil {
		fmt.Printf("!!!error starting the TUI: %v\n", err)
	}
}

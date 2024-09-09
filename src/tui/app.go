package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	state string
}

const (
	stateMenu        = "menu"
	stateScraping    = "scraping"
	stateViewing     = "viewing"
	stateOptions     = "options"
	stateForceScrape = "forceScrape"
)

func InitialModel() model {
	return model{state: stateMenu}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "1":
			m.state = stateScraping
		case "2":
			m.state = stateViewing
		case "3":
			m.state = stateOptions
		case "4":
			return m, tea.Quit
		}

	// submenu
	case string:
		if msg == "forceScrape" {
			m.state = stateForceScrape
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateMenu:
		return viewMainMenu()
	case stateScraping:
		return "Scraping data... (Press q to go back)\n"
	case stateViewing:
		return "Displaying cached data... (Press q to go back)\n"
	case stateOptions:
		return "Options Menu: Change settings (Press q to go back)\n"
	case stateForceScrape:
		return "Force scraping even with cache... (Press q to go back)\n"
	}
	return ""
}

func viewMainMenu() string {
	return `
RHUL Events Tracker

1. Run Scraper
2. View Current Data
3. Options
4. Exit

Choose an option: 
`
}

package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/bastiangx/rhs-scraper/src/scraper"
)

// model struct stores the state of the app
type model struct {
	state       string
	events      []scraper.Event // from scraper package
	scrapeError error
}

// Define application states
const (
	stateMenu         = "menu"         // Main menu
	stateScraping     = "scraping"     // Scraping in progress
	stateViewing      = "viewing"      // Viewing events
	stateOptions      = "options"      // Options menu
	stateForceScrape  = "forceScrape"  // Force scraping
	stateLoadingCache = "loadingCache" // Loading cached events
)

// Define custom message types
type ScrapeSuccessMsg struct {
	Events []scraper.Event
}

type ScrapeErrorMsg struct {
	Err error
}

// InitialModel function returns the initial model
func InitialModel() model {
	return model{state: stateMenu}
}

// Init function is called when the program starts
func (m model) Init() tea.Cmd {
	return nil
}

// Update function handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle custom scrape success message
	case ScrapeSuccessMsg:
		m.events = msg.Events
		m.scrapeError = nil
		m.state = stateViewing
		fmt.Printf("Loaded %d events into TUI\n", len(m.events)) // Debug print
		return m, nil

	// Handle custom scrape error message
	case ScrapeErrorMsg:
		m.scrapeError = msg.Err
		m.state = stateMenu // Return to menu or handle as desired
		return m, nil

	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case "q":
			m.state = stateMenu
			return m, nil

		case "1":
			// Option 1: Run Scraper
			// Before scraping, check if we should scrape or load from cache
			if scraper.ShouldScrape() {
				m.state = stateScraping
				return m, scrapeEventsCmd(10) // Pass the number of events to scrape
			} else {
				m.state = stateLoadingCache
				return m, loadCachedEventsCmd()
			}

		case "2":
			// Option 2: View Current Data
			m.state = stateLoadingCache
			return m, loadCachedEventsCmd()

		case "3":
			// Option 3: Options
			m.state = stateOptions
			return m, nil

		case "4":
			// Option 4: Exit
			return m, tea.Quit
		}

	// Handle submenu messages or other string messages
	case string:
		if msg == "done_scraping" || msg == "done_viewing" {
			m.state = stateViewing
		}
	}

	return m, nil
}

// View function renders the UI based on the current state
func (m model) View() string {
	switch m.state {
	case stateMenu:
		return viewMainMenu()
	case stateScraping:
		return "Scraping data... Please wait.\nPress 'q' to cancel.\n"
	case stateLoadingCache:
		return "Loading cached events... Please wait.\nPress 'q' to cancel.\n"
	case stateViewing:
		return viewEvents(m.events)
	case stateOptions:
		return viewOptionsMenu()
	default:
		if m.scrapeError != nil {
			return fmt.Sprintf("Error: %v\nPress 'q' to return to the menu.", m.scrapeError)
		}
		return ""
	}
}

// viewMainMenu function renders the main menu
func viewMainMenu() string {
	return `
RHUL Events Tracker

1. Run Scraper
2. View Current Data
3. Options
4. Exit

Choose an option: `
}

// viewOptionsMenu function renders the options menu
func viewOptionsMenu() string {
	return `
Options Menu:

1. Set Number of Events to Scrape
2. Set Cache Expiry Time (Min 1 hour)
3. [Future Option]
4. Back to Main Menu (q)

Choose an option: `
}

// viewEvents function renders the list of events
func viewEvents(events []scraper.Event) string {
	if len(events) == 0 {
		return "No events found...\n"
	}
	var output string
	for i, event := range events {
		output += fmt.Sprintf("Event %d\n", i+1)
		output += fmt.Sprintf(
			"Title: %s\nDate: %s\nLocation: %s\nCategory: %s\nDescription: %s\n\n",
			event.Title,
			event.Date,
			event.Location,
			event.Category,
			event.Description,
		)
	}
	fmt.Printf("scraped %d events\n", len(events))
	return output
}

// Command to scrape events
func scrapeEventsCmd(numEvents int) tea.Cmd {
	return func() tea.Msg {
		events, err := scraper.ScrapeEvents(numEvents)
		if err != nil {
			return ScrapeErrorMsg{Err: err}
		}
		fmt.Printf("scraped %d events\n", len(events))
		return ScrapeSuccessMsg{Events: events}
	}
}

// Command to load cached events
func loadCachedEventsCmd() tea.Cmd {
	return func() tea.Msg {
		events, err := scraper.LoadCachedEvents()
		if err != nil {
			return ScrapeErrorMsg{Err: err}
		}
		fmt.Printf("scraped %d events\n", len(events))
		return ScrapeSuccessMsg{Events: events}
	}
}

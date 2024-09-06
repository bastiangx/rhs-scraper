package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gocolly/colly"
)

// Event struct
type Event struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

func main() {
	// store the event temporarily
	eventsMap := make(map[string]*Event)
	// slice to store all events
	var events []Event

	// counters
	MaxEvents := 10
	eventCount := 0

	// main collector
	c := colly.NewCollector()

	c.OnHTML(".msl_event_name", func(e *colly.HTMLElement) {
		if eventCount >= MaxEvents {
			return
		}

		title := e.Text
		event, found := eventsMap[title]
		// if not found, create a new event
		if !found {
			event = &Event{}
			eventsMap[title] = event
		}
		// store the title
		event.Title = title

		fmt.Println("event title:", event.Title)
	})

	c.OnHTML(".msl_event_time", func(e *colly.HTMLElement) {
		if eventCount >= MaxEvents {
			return
		}
		// assuming title is added
		title := e.DOM.Parent().
			Find(".msl_event_name").
			Text()
		if event, found := eventsMap[title]; found {
			event.Date = e.Text

			fmt.Println("event date:", event.Date)
		}
	})

	c.OnHTML(".msl_event_description", func(e *colly.HTMLElement) {
		if eventCount >= MaxEvents {
			return
		}
		title := e.DOM.Parent().
			Find(".msl_event_name").
			Text()
		if event, found := eventsMap[title]; found {
			event.Description = e.Text

			fmt.Println("event description:", event.Description)
		}
	})

	c.OnHTML(".msl_event_location", func(e *colly.HTMLElement) {
		if eventCount >= MaxEvents {
			return
		}

		title := e.DOM.Parent().
			Find(".msl_event_name").
			Text()
		if event, found := eventsMap[title]; found {
			event.Location = e.Text

			fmt.Println("event location:", event.Location)
		}
	})

	// category field
	c.OnHTML(".msl_event_types", func(e *colly.HTMLElement) {
		if eventCount >= MaxEvents {
			return
		}

		title := e.DOM.Parent().Find(".msl_event_name").Text()
		if event, found := eventsMap[title]; found {
			// Grab only the first category using a loop but break after the first
			e.ForEach("a", func(i int, el *colly.HTMLElement) {
				if i == 0 { // First element (index 0)
					firstCat := el.Text
					event.Category = firstCat // Assign only the first category
					fmt.Println("event category:", event.Category)
				}
			})
		}
	})

	// handle request
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting right now...", r.URL)
	})

	// store the events in a json
	c.OnScraped(func(r *colly.Response) {
		for _, event := range eventsMap {
			if eventCount >= MaxEvents {
				break
			}
			events = append(events, *event)
			eventCount++
		}
		fmt.Println("found events: ", events)

		file, err := os.Create("events.json")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// init encoder
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(events); err != nil {
			log.Fatal("! Error encoding events to json: ", err)
		}

		// print success message
		fmt.Println("events saved to events.json")
	})

	// Start the collector
	c.Visit("http://su.rhul.ac.uk/events/calendar/")
}

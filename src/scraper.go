// // TODO:
// // [ x ] Timestamps on scraping events
// // [ x ] If last scrape > 12hrs, scrape again, if not, use the json cache
// // [ x ] Add a flag to force scrape
// // [ x ] Add a flag to specify the number of events to scrape

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gocolly/colly"
)

// Event struct to hold event details
type Event struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// CachedData struct to store events with a timestamp
type CachedData struct {
	LastScraped time.Time `json:"last_scraped"`
	Events      []Event   `json:"events"`
}

func main() {
	// Command-line flags
	forceScrape := flag.Bool("f", false, "Force scraping even if recent data is available.")
	numEvents := flag.Int("n", 10, "Number of events to scrape (min 2, max 25).")
	flag.Parse()

	// Ensure number of events is within limits
	if *numEvents < 2 {
		*numEvents = 2
	} else if *numEvents > 25 {
		*numEvents = 25
	}

	// Check if we should scrape or use cached data
	if *forceScrape || shouldScrape() {
		fmt.Println("Scraping...")

		// Scrape up to the specified number of events
		scrapeEvents(*numEvents)
	} else {
		fmt.Println("Recent data found.")
		// Ask user if they want to force a scrape
		fmt.Println("Do you want to force scrape? (y/n): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.ToLower(scanner.Text())

		if response == "y" {
			scrapeEvents(*numEvents)
		} else {
			fmt.Println("Using cached data.")
			loadCachedEvents()
		}
	}
}

// Function to scrape the events
func scrapeEvents(numEvents int) {
	// Store the event temporarily
	eventsMap := make(map[string]*Event)
	var events []Event

	// Counter
	eventCount := 0

	// Main collector
	c := colly.NewCollector()

	// Scrape event title
	c.OnHTML(".msl_event_name", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}

		title := e.Text
		event, found := eventsMap[title]
		if !found {
			event = &Event{}
			eventsMap[title] = event
		}
		event.Title = title
		fmt.Println("Event title:", event.Title)
	})

	// Scrape event date
	c.OnHTML(".msl_event_time", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}
		title := e.DOM.Parent().Find(".msl_event_name").Text()
		if event, found := eventsMap[title]; found {
			event.Date = e.Text
			fmt.Println("Event date:", event.Date)
		}
	})

	// Scrape event description
	c.OnHTML(".msl_event_description", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}
		title := e.DOM.Parent().Find(".msl_event_name").Text()
		if event, found := eventsMap[title]; found {
			event.Description = e.Text
			fmt.Println("Event description:", event.Description)
		}
	})

	// Scrape event location
	c.OnHTML(".msl_event_location", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}
		title := e.DOM.Parent().Find(".msl_event_name").Text()
		if event, found := eventsMap[title]; found {
			event.Location = e.Text
			fmt.Println("Event location:", event.Location)
		}
	})

	// Scrape event category
	c.OnHTML(".msl_event_types", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}
		parentEvent := e.DOM.Parent().Parent()
		title := parentEvent.Find(".msl_event_name").Text()

		if event, found := eventsMap[title]; found {
			validCategoryFound := false

			e.ForEach("a", func(i int, el *colly.HTMLElement) {
				categoryText := el.Text

				// Skip "Free" category
				if categoryText == "Free" || categoryText == "free" {
					return
				}

				// Assign the first valid category found
				event.Category = categoryText
				validCategoryFound = true
				fmt.Println("Event category:", event.Category)
				return
			})

			// Default category if none found
			if !validCategoryFound {
				event.Category = "Uncategorized"
				fmt.Println("No valid category found, assigning 'Uncategorized'")
			}
		} else {
			fmt.Println("Warning: Category found but no matching event title!")
		}
	})

	// Handle request
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting right now...", r.URL)
	})

	// Store the events in a JSON
	c.OnScraped(func(r *colly.Response) {
		for _, event := range eventsMap {
			if eventCount >= numEvents {
				break
			}
			events = append(events, *event)
			eventCount++
		}

		// Make cached data with timestamp
		cachedData := CachedData{
			LastScraped: time.Now(),
			Events:      events,
		}

		fmt.Println("Found events: ", events)

		cacheDir := getCacheDir()
		jsonFilePath := filepath.Join(cacheDir, "events.json")

		// Save all to JSON with timestamp
		file, err := os.Create(jsonFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// Init encoder
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(cachedData); err != nil {
			log.Fatal("!!! Error encoding events to JSON: ", err)
		}

		fmt.Println("Events saved to", jsonFilePath)
	})

	// Start scraping
	c.Visit("https://su.rhul.ac.uk/events/calendar/")
}

// Function to check if we should scrape again or use cached data
func shouldScrape() bool {
	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	// Load the cache
	file, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Println("No recent data found, scraping...")
		return true
	}
	defer file.Close()

	var cachedData CachedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cachedData); err != nil {
		fmt.Println("Error reading cache, scraping...")
		return true
	}

	// If last scraped time is within 12 hours
	if time.Since(cachedData.LastScraped) < 12*time.Hour {
		fmt.Printf(
			"*** Using cached data. Last scraped: %s\n",
			humanize.Time(cachedData.LastScraped),
		)
		return false
	}

	// Scrape again (more than 12 hours)
	return true
}

// Function to return the cache directory
func getCacheDir() string {
	// Based on OS
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("Error getting cache directory:", err)
	}

	// Append custom dir
	appCacheDir := filepath.Join(cacheDir, "rhs-scraper")

	// Create dir if it doesn't exist
	if _, err := os.Stat(appCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(appCacheDir, 0755) // permission: rwxr-xr-x
		if err != nil {
			log.Fatal("Error creating cache directory:", err)
		}
	}
	return appCacheDir
}

// Function to load cached events from the JSON file
func loadCachedEvents() {
	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	// Load the cache
	file, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatal("Error loading cache:", err)
	}
	defer file.Close()

	var cachedData CachedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cachedData); err != nil {
		log.Fatal("Error decoding cache:", err)
	}

	fmt.Println("Loaded cached events:", cachedData.Events)
}

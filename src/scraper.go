// package scraper implements a simple web scraper to get events from the Royal Holloway Students' Union Events Calendar.
// It uses the Colly library to scrape the events from the website.
// The events are stored in a JSON file in the user's cache directory.
// The scraper checks if the data is recent (within 12 hours) and uses the cached data if it is.
// The user can force scrape the data if they want to.
// The user can also specify the number of events to scrape (min 2, max 25).
// The scraped data will be handed to Bubbles package later for TUI display.
package scraper

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

// Event struct stores the event details
// Title, Date, Location, Category, Description
// can be marshalled to JSON
type Event struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// CachedData struct stores last timestamp
type CachedData struct {
	LastScraped time.Time `json:"last_scraped"`
	Events      []Event   `json:"events"`
}

func main() {
	// cmd flags
	// -f can be used to force scrape even if cache exists
	forceScrape := flag.Bool("f", false, "force scraping")

	// -n can be used to specify number of events to scrape
	numEvents := flag.Int("n", 10, "Number of events to scrape (min 2, max 25).")
	flag.Parse()

	// validate number of events
	if *numEvents < 2 {
		*numEvents = 2
	} else if *numEvents > 25 {
		*numEvents = 25
	}

	if *forceScrape || shouldScrape() {
		fmt.Println("Scraping --> ")

		scrapeEvents(*numEvents)
	} else {
		fmt.Println("recent data found!")
		fmt.Println("Do you want to Force scrape? (y/n): ")

		// scan user input
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.ToLower(scanner.Text())

		if response == "y" || response == "f" {
			scrapeEvents(*numEvents)
		} else {
			fmt.Println("using cached data.")
			loadCachedEvents()
		}
	}
}

////////////////////////////////////////////
//////////////// SCRAPING //////////////////
////////////////////////////////////////////

// scrapeEvents function scrapes the events from the SU Events Calendar
// uses colly collector
// using OnHTML since the data is in HTML
func scrapeEvents(numEvents int) {
	// store the events in a map
	eventsMap := make(map[string]*Event)

	// store the events in a slice
	var events []Event

	// Counter
	eventCount := 0

	// Main collector
	c := colly.NewCollector()

	// title of the event
  // the targeted string is in the class msl_event_name via css selector
	c.OnHTML(".msl_event_name", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}

		// set the title of the event
		title := e.Text

		// check if the event already exists
		event, found := eventsMap[title]
		if !found {
			// if not, create a new event
			event = &Event{}
			eventsMap[title] = event
		}
		event.Title = title
		fmt.Println("Event title:", event.Title)
	})

	// Scrape event date
	// all c.OnHTML functions are similar
	// onlu the class name changes / css selector
	// we also check the parent node to get the possible missing data
	c.OnHTML(".msl_event_time", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			return
		}
		// get the parent node and find class name
		title := e.DOM.Parent().Find(".msl_event_name").Text()
		// get the event from the map
		if event, found := eventsMap[title]; found {
			// set the date
			event.Date = e.Text
			fmt.Println("Event date:", event.Date)
		}
	})

	// event description
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

	// event location
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

	// event category
	// this one has some filtering
	// for spammed categories
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

				// skip "Free" category
				if categoryText == "Free" || categoryText == "free" {
					return
				}

				// assign the first valid category found
				event.Category = categoryText
				validCategoryFound = true
				fmt.Println("Event category:", event.Category)
				return
			})

			// Default category if no valid cat found
			if !validCategoryFound {
				event.Category = "Uncategorized"
				fmt.Println("no valid category found, assigning 'Uncategorized'")
			}
		} else {
			fmt.Println("Warning: Category found but no matching event title!")
		}
	})

	// handle request
	// temporary Println, will be bubbles later
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting right now...", r.URL)
	})

	////////////////////////////////////////////
	////////////// AFTER SCRAPING //////////////
	////////////////////////////////////////////

	// store the events in a JSON
	c.OnScraped(func(r *colly.Response) {
		for _, event := range eventsMap {
			if eventCount >= numEvents {
				break
			}
			events = append(events, *event)
			eventCount++
		}

		// cachedData strcut
		// appends events to the slice
		// has a timestamp
		// time.now is later converted to human readable format
		cachedData := CachedData{
			LastScraped: time.Now(),
			Events:      events,
		}

		fmt.Println("Found events: ", events)

		cacheDir := getCacheDir()
		jsonFilePath := filepath.Join(cacheDir, "events.json")

		// save all to JSON with timestamp
		file, err := os.Create(jsonFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// Init json encoder
		encoder := json.NewEncoder(file)
		// indent for readability
		encoder.SetIndent("", "  ")

		if err := encoder.Encode(cachedData); err != nil {
			log.Fatal("!!! error encoding events to JSON: ", err)
		}

		fmt.Println("Events saved to", jsonFilePath)
	})

	// start scraping
	// visit the events Calendar
	c.Visit("https://su.rhul.ac.uk/events/calendar/")
}

////////////////////////////////////////////
//////////////// CHECK CACHE ///////////////
////////////////////////////////////////////

// shouldScrape function checks if the data is recent
// if the data is older than 12 hours, it returns true
// retrieves the timestamp from the json file
func shouldScrape() bool {
	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	// load the cache
	file, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Println("no recent data found, scraping...")
		return true
	}
	defer file.Close()

	var cachedData CachedData
	// read and decode the json
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cachedData); err != nil {
		fmt.Println("!!! error reading cache, scraping...")
		return true
	}

	// if last scraped time is within 12 hours
	// uses humanized time format
	if time.Since(cachedData.LastScraped) < 12*time.Hour {
		fmt.Printf(
			"*** Using cached data. Last scraped: %s\n",
			humanize.Time(cachedData.LastScraped),
		)
		return false
	}

	// if older than 12 hours, scrape
	return true
}

////////////////////////////////////////////
//////////////// CACHE DIR /////////////////
////////////////////////////////////////////

// getCacheDir function gets the cache directory
// creates the directory if it doesn't exist
// for mac and windows, it uses the default cache directory
// for linux, it is /.cache/rhs-scraper under the home directory
func getCacheDir() string {
	// os method to get default cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal("!!! error getting cache directory:", err)
	}

	// append custom dir
	appCacheDir := filepath.Join(cacheDir, "rhs-scraper")

	// create dir if it doesnt exist
	if _, err := os.Stat(appCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(appCacheDir, 0755) // permission: rwxr-xr-x
		if err != nil {
			log.Fatal("!!! error creating cache directory:", err)
		}
	}
	return appCacheDir
}

////////////////////////////////////////////
//////////////// LOAD CACHE ////////////////
////////////////////////////////////////////

// loadCachedEvents function prints the events from json
// uses json decoder
func loadCachedEvents() {
	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	file, err := os.Open(jsonFilePath)
	if err != nil {
		log.Fatal("!!! error loading cache:", err)
	}
	defer file.Close()

	var cachedData CachedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cachedData); err != nil {
		log.Fatal("!!! error decoding cache:", err)
	}

	fmt.Println("loaded cached events:", cachedData.Events)
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

// Event struct contains event card info
// can have default values for each field
type Event struct {
	Title       string `json:"title"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// CachedData struct holds the timestamp of the last scrape and the list of events
type CachedData struct {
	LastScraped time.Time `json:"last_scraped"`
	Events      []Event   `json:"events"`
}

// ScrapeEvents func scrapes specified number of events
// returns a slice of Event and an Err if the scraping fails
func ScrapeEvents(numEvents int) ([]Event, error) {
	// validate num of events
	if numEvents < 2 {
		numEvents = 2
	} else if numEvents > 25 {
		numEvents = 25
	} // min max can be changed later

	// main collector
	c := colly.NewCollector()

	// slice of events
	var events []Event
	eventCount := 0

	// OnHTML handler func for each event container
	// returns the title, date, location, category and description
	// uses GoQuery selector for parsing the html, i think
	// used becuase of the structure of the page -> only css selectors and classes & no js or ajax rendering
	c.OnHTML(".msl_eventlist .event_item", func(e *colly.HTMLElement) {
		if eventCount >= numEvents {
			// stop scraping
			e.Request.Abort()
			return
		}

		// extract title
		title := strings.TrimSpace(e.ChildText("a.msl_event_name"))
		if title == "" {
			log.Println("!! event with an empty title.")
			return
		}

		// extract date
		// dd is the tag and msl_event_time is the class
		// from the html structure
		date := strings.TrimSpace(e.ChildText("dd.msl_event_time"))

		// extract location
		location := strings.TrimSpace(e.ChildText("dd.msl_event_location"))

		// extract desc
		description := strings.TrimSpace(e.ChildText("dd.msl_event_description"))

		// extract category
		// also checks if its a spammed cat -> "free", etc

		// default category
		category := "uncategorized"
		var categories []string
		// for each category
		e.ForEach("dd.msl_event_types a", func(_ int, el *colly.HTMLElement) {
			categoryText := strings.TrimSpace(el.Text)
			// check if its not a spam or empty
			if strings.ToLower(categoryText) != "free" && categoryText != "" {
				categories = append(categories, categoryText)
			}
		})
		if len(categories) > 0 {
			category = strings.Join(categories, ", ")
		}

		// append the valid events
		event := Event{
			Title:       title,
			Date:        date,
			Location:    location,
			Category:    category,
			Description: description,
		}
		events = append(events, event)
		eventCount++

		// log.Printf("appended event %d: %s\n", eventCount, title)
	})

	// OnRequest func requests the url
	// can be shown to the user
	// can be pure BubbleTea UI later
	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting %s\n", r.URL.String())
	})

	// OnResponse func logs the status code from the server
	// can be shown to the user if Err
	c.OnResponse(func(r *colly.Response) {
		// log.Printf("received response with code: %d\n", r.StatusCode)
	})

	// OnError func logs the error if any
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %s: %v\n", r.Request.URL.String(), err)
	})

	// visti func send a get request to the URL
	// at the end of collector since it needs
	// to compile or setup all the handlers first
	err := c.Visit("https://su.rhul.ac.uk/events/calendar/")
	if err != nil {
		return nil, fmt.Errorf("failed to visit events page: %w", err)
	}

	// wait for the collector to finish
	err = saveToCache(events)
	if err != nil {
		return nil, fmt.Errorf("failed to save events to cache: %w", err)
	}

	// log.Printf("scraping succesful. %d events found and saved.\n", len(events))
	return events, nil
}

// saveToCache func encodes the current scraped events
// along with the current timestamp to a JSON cache file.
// files' path are os defined, via UserCacheDir func
func saveToCache(events []Event) error {
	cachedData := CachedData{
		LastScraped: time.Now(),
		Events:      events,
	}

	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	// default os based
	file, err := os.Create(jsonFilePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(cachedData); err != nil {
		return fmt.Errorf("failed to encode events to JSON: %w", err)
	}

	// log.Printf("new events saved to %s\n", jsonFilePath)
	return nil
}

// ShouldScrape func determines whether a new scrape is necessary based on timestamp
// returns true if the cache is older than 12 hours or doesnt exist
func ShouldScrape() bool {
	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	file, err := os.Open(jsonFilePath)
	if err != nil {
		// log.Println("no recent data found, scraping new events...")
		return true
	}
	defer file.Close()

	// decode the cache file
	var cachedData CachedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cachedData); err != nil {
		// log.Println("!error reading cache, scraping...")
		return true
	}

	// timestamp check
	// for now, var is 12 hours
	var timeLimit time.Duration = 12
	if time.Since(cachedData.LastScraped) < timeLimit*time.Hour {
		// log.Printf("last scraped %s ago, no need to scrape new events.\n",
		// 	humanize.Time(cachedData.LastScraped))
		return false
	}

	// log.Printf("last scraped more than %d hours ago, scraping new events...\n", 12)
	return true
}

// LoadCachedEvents func loads events from the cache
// returns a slice of Event and an error if loading fails
func LoadCachedEvents() ([]Event, error) {
	cacheDir := getCacheDir()
	jsonFilePath := filepath.Join(cacheDir, "events.json")

	file, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("!error opening cache file: %w", err)
	}
	defer file.Close()

	var cachedData CachedData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cachedData); err != nil {
		return nil, fmt.Errorf("!error decoding cache file: %w", err)
	}

	log.Printf("--> Loaded %d events from cache.\n", len(cachedData.Events))
	return cachedData.Events, nil
}

// getCacheDir func retrieves the cache directory path
// also appends the app name to path
// returns the path as a string
func getCacheDir() string {
	// UserCacheDir, os method gets the default cache directory
	// in linux, its ~/.cache/...
	// in win, its %APPDATA%/...
	// in mac, its ~/Library/Caches/...
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("!!error getting user cache directory: %v", err)
	}

	appCacheDir := filepath.Join(cacheDir, "rhs-scraper")

	if _, err := os.Stat(appCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(appCacheDir, 0755) // permission: rwxr-xr-x
		if err != nil {
			log.Fatalf("!!!error creating cache directory: %v", err)
			log.Fatalf("make sure you have the correct .cache directory")
			log.Fatalf("visit the git page to see the correct cache directory")
		}
	}

	// return path as string
	return appCacheDir
}

func printer(events []Event) {
  for _, event := range events {
    fmt.Printf("Title: %s\n", event.Title)
    fmt.Printf("Date: %s\n", event.Date)
    fmt.Printf("Location: %s\n", event.Location)
    fmt.Printf("Category: %s\n", event.Category)
    fmt.Printf("Description: %s\n", event.Description)
    fmt.Println()
  }
}

func main() {
	shouldScrape := ShouldScrape()
	if shouldScrape {
		events, err := ScrapeEvents(10)
    printer(events)
		if err != nil {
			log.Fatalf("failed to scrape events: %v", err)
		}
		_ = events
	} else {
		events, err := LoadCachedEvents()
    printer(events)
		if err != nil {
			log.Fatalf("failed to load cached events: %v", err)
		}
		_ = events
	}
}

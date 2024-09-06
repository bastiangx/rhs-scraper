 i identified these css selectors for the actual targets. needs testing


```go

// Scraping event names
c.OnHTML(".msl_event_name", func(e *colly.HTMLElement) {
    eventName := e.Text
    fmt.Println("Event Name:", eventName)
})

// Scraping event dates
c.OnHTML(".msl_event_time", func(e *colly.HTMLElement) {
    eventDate := e.Text
    fmt.Println("Event Date:", eventDate)
})

// Scraping event descriptions
c.OnHTML(".msl_event_description", func(e *colly.HTMLElement) {
    eventDescription := e.Text
    fmt.Println("Event Description:", eventDescription)
})
```

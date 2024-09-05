# rhs-scraper
Student Union Events CLI written in GO

### Overview:
The **RHUL Event Tracker** terminal-based application designed to display upcoming
events from the Royal Holloway University of London (RHUL) Student Union. It
fetches event data from the Student Union events page, processes it, and
presents it in a clean, interactive terminal UI. Users can view events for the 
current and upcoming week, navigate through events, and see detailed information for 
each event.

## Technical Stack:
- **Language**: Golang
- **UI Library**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) for building the Text User Interface (TUI).
- **Scraping Library**: [GoColly](https://github.com/gocolly/colly) for
    scraping event data from the Student Union events webpage.
- **Additional Tools**:
  - **Golang `time` package**: To schedule daily fetching of event data.
  - **Golang `net/http` package**: For sending HTTP requests (in case of API usage or scraping).

## Key Features:
1. **Event List Display**:
   - Shows this week’s and next week’s student union events.
   - Uses **Bubble Tea**'s list component for navigation (up/down arrows to scroll through events).

2. **Event Details View**:
   - When an event is selected, shows additional details (event description, time, location, etc.).
   - Implemented using Bubble Tea’s TUI components (e.g., modals for detailed views).

3. **Scraping Logic**:
   - Uses **GoColly** to scrape event data (event titles, dates, times, descriptions) from the RHUL Student Union website.
   - Data is parsed and structured in Go for easy presentation in the terminal UI.
   - Scraper runs daily to fetch and update event data (via cron jobs or `time.Sleep` in Go).

4. **Rate Limiting & Error Handling**:
   - **Colly’s Limit Rules** ensure responsible scraping, with delays between requests to 
   avoid overwhelming the server.
   - **Error handling** for failed requests or site changes to handle dynamic content or downtime gracefully.

## Planned Extensions:
- **Event Filtering**: Filter events by type (e.g., sports, music, talks).
- **Search Functionality**: Allow users to search for events based on keywords.
- **Reminder System**: Option to set reminders for specific events.

---

This app will demonstrate your ability to combine **web scraping**, **data processing**,
and **user interface design** in a terminal environment, using modern tools like **Go** and **Bubble Tea**.

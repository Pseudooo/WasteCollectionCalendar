package main

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
)

type WasteCollectionEvent struct {
	Title string
	Date  time.Time
}

func main() {
	router := gin.Default()

	router.GET("/calendar", getCalendarHandler)

	router.Run("localhost:8080")
}

type CalendarQuery struct {
	Uprn     string `form:"uprn" binding:"required"`
	Postcode string `form:"postcode" binding:"required"`
}

func getCalendarHandler(c *gin.Context) {
	var query CalendarQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentTime := time.Now()
	events, err := getWasteCollectionEvents(query.Uprn, query.Postcode, int(currentTime.Month()), currentTime.Year())
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	calendar := buildCalendarFromWasteCollectionEvents(events)

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="calendar.ics"`)
	c.Header("Cache-Control", "no-cache; no-store; must-revalidate")

	c.String(http.StatusOK, calendar.Serialize())
}

func getWasteCollectionEvents(uprn string, postcode string, month int, year int) ([]WasteCollectionEvent, error) {
	endpoint := "https://ilambassadorformsprod.azurewebsites.net/wastecollectiondays/wastecollectioncalendar"
	data := url.Values{}
	data.Set("Postcode", postcode)
	data.Set("Month", strconv.Itoa(month))
	data.Set("Year", strconv.Itoa(year))
	data.Set("Uprn", uprn)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	events, err := parseWasteCollectionEventsFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func parseWasteCollectionEventsFromReader(reader io.Reader) ([]WasteCollectionEvent, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	var events []WasteCollectionEvent

	doc.Find(".cal-month-box .rc-event-container").Each(func(i int, s *goquery.Selection) {
		eventTitle := strings.TrimSpace(s.Find("span").Text())

		calInnerParent := s.Closest(".cal-inner")

		dateStr, exists := calInnerParent.Find(".day-no").Attr("data-cal-date")
		parsedTime, err := time.Parse("2006-01-02T15:04:05", dateStr)

		if exists && eventTitle != "" && err == nil {
			events = append(events, WasteCollectionEvent{
				Title: eventTitle,
				Date:  parsedTime,
			})
		}
	})

	return events, nil
}

func buildCalendarFromWasteCollectionEvents(events []WasteCollectionEvent) *ics.Calendar {
	calendar := ics.NewCalendar()

	for index, value := range events {
		calendarEvent := calendar.AddEvent(strconv.Itoa(index))
		calendarEvent.SetSummary(value.Title)
		calendarEvent.SetProperty(
			ics.ComponentPropertyDtStart,
			value.Date.Format("20060102"),
			ics.WithValue(string(ics.ValueDataTypeDate)),
		)
		calendarEvent.SetProperty(
			ics.ComponentPropertyDtEnd,
			value.Date.AddDate(0, 0, 1).Format("20060102"),
			ics.WithValue(string(ics.ValueDataTypeDate)),
		)
	}

	return calendar
}

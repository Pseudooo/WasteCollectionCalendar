package calendar

import (
	"fmt"
	"time"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/models"
	ics "github.com/arran4/golang-ical"
)

func buildCalendarFromWasteCollectionEvents(events []models.WasteCollectionEvent) *ics.Calendar {
	calendar := ics.NewCalendar()
	calendar.SetMethod(ics.MethodRequest)

	nowStr := time.Now().UTC().Format("20060102T150405Z")

	for _, value := range events {
		uniqueId := fmt.Sprintf("waste-collection-%d-%d", int(value.Date.Month()), value.Date.Day())
		calendarEvent := calendar.AddEvent(uniqueId)
		calendarEvent.SetSummary(value.Title)
		calendarEvent.SetProperty(ics.ComponentPropertyDtstamp, nowStr)
		calendarEvent.SetProperty(ics.ComponentPropertyCreated, nowStr)

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

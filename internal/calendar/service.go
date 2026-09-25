package calendar

import (
	"strconv"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/models"
	ics "github.com/arran4/golang-ical"
)

func buildCalendarFromWasteCollectionEvents(events []models.WasteCollectionEvent) *ics.Calendar {
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

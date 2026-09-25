package calendar

import (
	"net/http"
	"time"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/models"
	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
)

type CalendarQuery struct {
	Uprn     string `form:"uprn" binding:"required"`
	Postcode string `form:"postcode" binding:"required"`
}

func GetCalendarHandler(c *gin.Context) {
	var query CalendarQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var allEvents []models.WasteCollectionEvent
	currentTime := time.Now()
	for i := range 3 {
		evalTime := currentTime.AddDate(0, i, 0)
		events, err := getWasteCollectionEvents(query.Uprn, query.Postcode, int(evalTime.Month()), evalTime.Year())
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		allEvents = append(allEvents, events...)
	}

	calendar := buildCalendarFromWasteCollectionEvents(allEvents)

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="calendar.ics"`)
	c.Header("Cache-Control", "no-cache; no-store; must-revalidate")

	c.String(http.StatusOK, calendar.Serialize(ics.WithNewLineWindows))
}

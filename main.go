package main

import (
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/calendar", getCalendar)

	router.Run("localhost:8080")
}

func getCalendar(c *gin.Context) {
	calendar := ics.NewCalendar()
	event := calendar.AddEvent("my-id")
	event.SetCreatedTime(time.Now())
	event.SetDtStampTime(time.Now())
	event.SetStartAt(time.Now())
	event.SetEndAt(time.Now())

	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="calendar.ics"`)
	c.Header("Cache-Control", "no-cache; no-store; must-revalidate")

	c.String(http.StatusOK, calendar.Serialize())
}

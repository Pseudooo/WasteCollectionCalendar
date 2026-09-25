package main

import (
	"github.com/Pseudooo/WasteCollectionCalendar/internal/calendar"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/calendar", calendar.GetCalendarHandler)

	router.Run("localhost:8080")
}

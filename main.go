package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/calendar"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router := gin.New()
	router.Use(SlogMiddleware(logger))
	router.Use(gin.Recovery())

	router.GET("/calendar", calendar.GetCalendarHandler)

	router.Run("localhost:8080")
}

func SlogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		logger.Info(
			"request",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("error", slog.String("error", err.Error()))
			}
		}
	}
}

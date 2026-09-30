package main

import (
	"log/slog"
	"os"
	"time"
	"uuid"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/calendar"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/gin-gonic/gin"
)

const LoggerKey = "slog_logger"

func main() {
	globalLogAttributes := []slog.Attr{
		slog.String("service.name", "bindayapi"),
		slog.String("service.version", "0.1.0"),
	}
	loggingHandler := slog.NewJSONHandler(os.Stdout, nil).WithAttrs(globalLogAttributes)
	logger := slog.New(loggingHandler)

	calendarHandler := &calendar.CalendarHandler{Logger: logger}

	router := gin.New()
	router.Use(SlogMiddleware(logger))
	router.Use(gin.Recovery())

	router.GET("/calendar", calendarHandler.GetCalendar)

	router.Run("localhost:8080")
}

func SlogMiddleware(baseLogger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		correlationId := c.GetHeader("X-Correlation-Id")
		if correlationId == "" {
			correlationId = uuid.New().String()
		}
		c.Header("X-Correlation-Id", correlationId)

		requestLogger := baseLogger.With(
			slog.String("correlation_id", correlationId),
		)
		c.Set(LoggerKey, requestLogger)

		c.Next()

		requestLogger.Info(
			"Request Completed",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Duration("latency", time.Since(start)),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("error", slog.String("error", err.Error()))
			}
		}
	}
}

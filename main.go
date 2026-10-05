package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
	"uuid"

	telemetry "github.com/Pseudooo/WasteCollectionCalendar/internal"
	"github.com/Pseudooo/WasteCollectionCalendar/internal/calendar"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const LoggerKey = "slog_logger"

func main() {
	ctx := context.Background()
	shutdownMetrics, err := telemetry.InitMetrics(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := shutdownMetrics(ctx); err != nil {
			fmt.Printf("Error shutting down metrics%v\n", err)
		}
	}()

	globalLogAttributes := []slog.Attr{
		slog.String("service.name", "waste-collection-api"),
		slog.String("service.version", "0.1.0"),
	}
	loggingHandler := slog.NewJSONHandler(os.Stdout, nil).WithAttrs(globalLogAttributes)
	logger := slog.New(loggingHandler)

	calendarHandler := &calendar.CalendarHandler{Logger: logger}

	router := gin.New()
	router.Use(SlogMiddleware(logger))
	router.Use(otelgin.Middleware("waste-collection-api"))
	router.Use(gin.Recovery())

	router.GET("/calendar", calendarHandler.GetCalendar)

	router.Run(":8080")
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

		elapsed := time.Since(start)
		response_code := strconv.Itoa(c.Writer.Status())

		requestLogger.Info(
			"Request Completed",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Duration("latency", elapsed),
			slog.String("status_code", response_code),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				requestLogger.Error("error", slog.String("error", err.Error()))
			}
		}
	}
}

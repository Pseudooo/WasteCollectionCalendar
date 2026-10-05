package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
	"uuid"

	"github.com/Pseudooo/WasteCollectionCalendar/internal/calendar"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const LoggerKey = "slog_logger"

func main() {
	ctx := context.Background()
	shutdownMetrics, err := InitMetrics(ctx)
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

func InitMetrics(ctx context.Context) (func(context.Context) error, error) {
	otelHost := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelHost == "" {
		otel.SetMeterProvider(noop.NewMeterProvider())
		return func(ctx context.Context) error { return nil }, nil
	}

	exporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String("waste-collection-api"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	reader := metric.NewPeriodicReader(
		exporter,
		metric.WithInterval(15*time.Second),
		metric.WithProducer(runtime.NewProducer()),
	)

	customBuckets := []float64{
		0.005, 0.010, 0.025, 0.050, 0.075, 0.100,
		0.250, 0.500, 0.750, 1.000, 2.500, 5.000, 10.000,
	}

	durationView := metric.NewView(
		metric.Instrument{Name: "http.server.request.duration"},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: customBuckets,
			},
		},
	)

	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(reader),
		metric.WithView(durationView),
	)

	otel.SetMeterProvider(provider)

	err = runtime.Start(
		runtime.WithMeterProvider(provider),
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics collection: %w", err)
	}

	return provider.Shutdown, nil
}

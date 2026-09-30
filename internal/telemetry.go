package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const LoggerKey = "slog_logger"

var (
	meter = otel.Meter("calendar-service-router")

	HttpRequestsTotal, _ = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Total number of HTTP requests received"),
	)

	HttpRequestDuration, _ = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
)

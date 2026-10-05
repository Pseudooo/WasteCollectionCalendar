package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

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

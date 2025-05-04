package telemetry

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTelemetry initializes OpenTelemetry tracing, metrics and logging.
// Returns a single shutdown function that handles all components.
func InitTelemetry(ctx context.Context, serviceName, version string) (func(context.Context) error, error) {
	// Create resource with service name
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
		),
		resource.WithFromEnv(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "resource.New")
	}

	// Initialize tracing
	traceExporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "autoexport.NewSpanExporter")
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize metrics
	metricExporter, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "autoexport.NewMetricReader")
	}

	metricProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metricExporter),
	)
	otel.SetMeterProvider(metricProvider)

	// Initialize logging using our registered composite exporter
	logExporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "autoexport.NewLogExporter")
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	// Return combined shutdown function
	return func(ctx context.Context) error {
		var errs []error

		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, errors.Wrap(err, "tracerProvider.Shutdown"))
		}

		if err := metricProvider.Shutdown(ctx); err != nil {
			errs = append(errs, errors.Wrap(err, "metricProvider.Shutdown"))
		}

		if err := loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, errors.Wrap(err, "loggerProvider.Shutdown"))
		}

		return wrapMultiErrors(errs)
	}, nil
}

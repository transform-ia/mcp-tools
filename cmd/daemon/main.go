package main

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/transform-ia/mcp-tools/pkg/daemon"
)

func main() {
	// Set OpenTelemetry environment variables
	os.Setenv("OTEL_TRACES_EXPORTER", "console")
	os.Setenv("OTEL_METRICS_EXPORTER", "console")

	// Initialize OpenTelemetry
	tracerProvider, meterProvider, err := initTelemetry()
	if err != nil {
		panic(errors.Wrap(err, "initTelemetry"))
	}
	defer func() {
		_ = tracerProvider.Shutdown(context.Background())
		_ = meterProvider.Shutdown(context.Background())
	}()

	// Run the daemon with required parameters
	daemon.RunDaemon("daemon", "1.0.0", func() error {
		// Create instruments
		meter := meterProvider.Meter("daemon")
		counter, err := meter.Int64Counter("daemon.operations.count")
		if err != nil {
			return errors.Wrap(err, "meter.Int64Counter")
		}

		histogram, err := meter.Float64Histogram("daemon.operation.duration")
		if err != nil {
			return errors.Wrap(err, "meter.Float64Histogram")
		}

		// Main loop with telemetry
		tracer := tracerProvider.Tracer("daemon")
		for {
			ctx, span := tracer.Start(context.Background(), "daemon.operation")
			
			// Record metrics
			startTime := time.Now()
			counter.Add(ctx, 1)
			histogram.Record(ctx, time.Since(startTime).Seconds())

			// Add span attributes and events
			span.SetAttributes(attribute.String("status", "processing"))
			span.AddEvent("operation.started", trace.WithAttributes(
				attribute.Int64("timestamp", time.Now().Unix()),
			))

			// Simulate work
			time.Sleep(5 * time.Second)

			span.SetAttributes(attribute.String("status", "completed"))
			span.End()
		}
		return nil
	})
}

func initTelemetry() (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, error) {
	// Create console trace exporter
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, nil, errors.Wrap(err, "stdouttrace.New")
	}

	// Create console metric exporter
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, nil, errors.Wrap(err, "stdoutmetric.New")
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", "daemon"),
			attribute.String("service.version", "1.0.0"),
		),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "resource.New")
	}

	// Create trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return tracerProvider, meterProvider, nil
}

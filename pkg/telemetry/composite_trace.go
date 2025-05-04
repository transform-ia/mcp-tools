// Package telemetry implements composite OpenTelemetry exporters
package telemetry

import (
	"context"
	"os"
	"sync"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func init() {
	// Register our composite exporter type
	autoexport.RegisterSpanExporter("console+otlp", func(ctx context.Context) (trace.SpanExporter, error) {
		const onlyTwo = 2

		var (
			proto     = os.Getenv(otelExporterOTLPTracesProtoEnvKey)
			err       error
			composite = &compositeTraceExporter{
				exporters: make([]trace.SpanExporter, onlyTwo),
			}
		)

		if proto == "" {
			proto = os.Getenv(otelExporterOTLPProtoEnvKey)
		}

		// Fallback to default, http/protobuf.
		if proto == "" {
			proto = "http/protobuf"
		}

		composite.exporters[0], err = stdouttrace.New()
		if err != nil {
			return nil, errors.Wrap(err, "stdouttrace.New")
		}

		switch proto {
		case "grpc":
			if composite.exporters[1], err = otlptracegrpc.New(ctx); err != nil {
				return nil, errors.Wrap(err, "otlptracegrpc.New")
			}
			return composite, nil
		case "http/protobuf":
			if composite.exporters[1], err = otlptracehttp.New(ctx); err != nil {
				return nil, errors.Wrap(err, "otlptracehttp.New")
			}
			return composite, nil
		default:
			return nil, errors.New("invalid OTLP protocol - should be one of ['grpc', 'http/protobuf']")
		}
	})
}

// compositeTraceExporter implements trace.SpanExporter
type compositeTraceExporter struct {
	exporters []trace.SpanExporter
	mu        sync.Mutex
}

func (c *compositeTraceExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	for _, exp := range c.exporters {
		if err := exp.ExportSpans(ctx, spans); err != nil {
			errs = append(errs, err)
		}
	}

	return wrapMultiErrors(errs)
}

func (c *compositeTraceExporter) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	for _, exp := range c.exporters {
		if err := exp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return wrapMultiErrors(errs)
}

func (c *compositeTraceExporter) ForceFlush(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	for _, exp := range c.exporters {
		if f, ok := exp.(interface{ ForceFlush(context.Context) error }); ok {
			if err := f.ForceFlush(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return wrapMultiErrors(errs)
}

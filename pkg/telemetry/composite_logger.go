// Package telemetry initialize opentelemetry
package telemetry

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
)

// copied from https://raw.githubusercontent.com/open-telemetry/opentelemetry-go-contrib/refs/tags/exporters/autoexport/v0.60.0/exporters/autoexport/logs.go
//
//nolint:lll
const otelExporterOTLPLogsProtoEnvKey = "OTEL_EXPORTER_OTLP_LOGS_PROTOCOL"

// copied from https://github.com/open-telemetry/opentelemetry-go-contrib/blob/exporters/autoexport/v0.60.0/exporters/autoexport/registry.go
//
//nolint:lll
const otelExporterOTLPProtoEnvKey = "OTEL_EXPORTER_OTLP_PROTOCOL"

func init() {
	// Register our composite exporter type
	autoexport.RegisterLogExporter("console+otlp", func(ctx context.Context) (log.Exporter, error) {
		const onlyTwo = 2

		var (
			proto     = os.Getenv(otelExporterOTLPLogsProtoEnvKey)
			err       error
			composite = &compositeExporter{
				exporters: make([]log.Exporter, onlyTwo),
			}
		)

		if proto == "" {
			proto = os.Getenv(otelExporterOTLPProtoEnvKey)
		}

		// Fallback to default, http/protobuf.
		if proto == "" {
			proto = "http/protobuf"
		}

		composite.exporters[0], err = stdoutlog.New()
		if err != nil {
			return nil, errors.Wrap(err, "stdoutlog.New")
		}

		switch proto {
		case "grpc":
			if composite.exporters[1], err = otlploggrpc.New(ctx); err != nil {
				return nil, errors.Wrap(err, "otlploggrpc.New")
			}

			return composite, nil
		case "http/protobuf":
			if composite.exporters[1], err = otlploghttp.New(ctx); err != nil {
				return nil, errors.Wrap(err, "otlploghttp.New")
			}

			return composite, nil
		default:
			return nil, errors.New("invalid OTLP protocol - should be one of ['grpc', 'http/protobuf']")
		}
	})
}

// compositeExporter implements log.Exporter
type compositeExporter struct {
	exporters []log.Exporter
	mu        sync.Mutex
}

func wrapMultiErrors(errs []error) error {
	lenErrs := len(errs)

	if lenErrs == 0 {
		return nil
	}

	errsString := make([]string, lenErrs)

	for index, err := range errs {
		errsString[index] = err.Error()
	}

	return errors.Errorf("error(s) occurred: %s", strings.Join(errsString, ","))
}

func (c *compositeExporter) Export(ctx context.Context, records []log.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	for _, exp := range c.exporters {
		if err := exp.Export(ctx, records); err != nil {
			errs = append(errs, err)
		}
	}

	return wrapMultiErrors(errs)
}

func (c *compositeExporter) ForceFlush(ctx context.Context) error {
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

func (c *compositeExporter) Shutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	for _, exp := range c.exporters {
		if s, ok := exp.(interface{ Shutdown(context.Context) error }); ok {
			if err := s.Shutdown(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return wrapMultiErrors(errs)
}

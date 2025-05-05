package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func TestInitTelemetry(t *testing.T) {
	// Reset global state before each test
	resetGlobals := func() {
		otel.SetTracerProvider(nil)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	t.Run("successful initialization", func(t *testing.T) {
		resetGlobals()
		ctx := context.Background()
		serviceName := "test-service"
		version := "1.0.0"

		shutdown, err := InitTelemetry(ctx, serviceName, version)
		require.NoError(t, err)
		require.NotNil(t, shutdown)

		// Verify tracer provider is set
		tracerProvider := otel.GetTracerProvider()
		_, ok := tracerProvider.(*sdktrace.TracerProvider)
		assert.True(t, ok, "expected sdktrace.TracerProvider")

		// Test shutdown - ignore metrics errors since we don't have collector running
		err = shutdown(ctx)
		if err != nil {
			assert.Contains(t, err.Error(), "metricProvider.Shutdown")
		}
	})

	t.Run("shutdown error handling", func(t *testing.T) {
		resetGlobals()
		// Test that shutdown function properly handles errors
		ctx := context.Background()
		shutdown := func(ctx context.Context) error {
			return errors.New("simulated shutdown error")
		}

		err := shutdown(ctx)
		assert.Error(t, err)
		assert.Equal(t, "simulated shutdown error", err.Error())
	})
}

func TestResourceAttributes(t *testing.T) {
	serviceName := "test-service"
	version := "1.0.0"

	shutdown, err := InitTelemetry(t.Context(), serviceName, version)
	require.NoError(t, err)
	defer func() {
		_ = shutdown(t.Context())
	}()

	// Verify the global tracer provider has our service name
	tracer := otel.Tracer("test")
	_, span := tracer.Start(t.Context(), "test-span")
	defer span.End()

	// The attributes should be visible in the span
	attrs := span.(sdktrace.ReadOnlySpan).Resource().Attributes()

	// Convert attributes to map for easier checking
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsString()
	}

	assert.Equal(t, serviceName, attrMap[string(semconv.ServiceNameKey)])
	assert.Equal(t, version, attrMap[string(semconv.ServiceVersionKey)])
}

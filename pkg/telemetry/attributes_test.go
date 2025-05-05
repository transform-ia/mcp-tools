package telemetry

import (
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"github.com/stretchr/testify/assert"
)

func TestPrefixAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    []attribute.KeyValue
		prefix   string
		expected []attribute.KeyValue
	}{
		{
			name:     "empty attributes",
			input:    []attribute.KeyValue{},
			prefix:   "test.",
			expected: []attribute.KeyValue{},
		},
		{
			name: "single attribute",
			input: []attribute.KeyValue{
				attribute.String("key", "value"),
			},
			prefix: "test.",
			expected: []attribute.KeyValue{
				attribute.String("test.key", "value"),
			},
		},
		{
			name: "multiple attributes",
			input: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.Int("key2", 42),
				attribute.Bool("key3", true),
			},
			prefix: "service.",
			expected: []attribute.KeyValue{
				attribute.String("service.key1", "value1"),
				attribute.Int("service.key2", 42),
				attribute.Bool("service.key3", true),
			},
		},
		{
			name: "empty prefix",
			input: []attribute.KeyValue{
				attribute.String("key", "value"),
			},
			prefix: "",
			expected: []attribute.KeyValue{
				attribute.String("key", "value"),
			},
		},
		{
			name: "special characters in prefix",
			input: []attribute.KeyValue{
				attribute.String("key", "value"),
			},
			prefix: "test-@#.",
			expected: []attribute.KeyValue{
				attribute.String("test-@#.key", "value"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := PrefixAttributes(test.input, test.prefix)
			assert.Equal(t, test.expected, actual)
		})
	}
}

type mockAttributer struct {
	attrs []attribute.KeyValue
}

func (m *mockAttributer) Attributes() []attribute.KeyValue {
	return m.attrs
}

func TestPrefixAttributers(t *testing.T) {
	tests := []struct {
		name       string
		attributers []Attributer
		prefix     string
		expected   []attribute.KeyValue
	}{
		{
			name:       "empty attributers",
			attributers: []Attributer{},
			prefix:     "test.",
			expected:   []attribute.KeyValue{},
		},
		{
			name: "single attributer with single attribute",
			attributers: []Attributer{
				&mockAttributer{
					attrs: []attribute.KeyValue{
						attribute.String("key", "value"),
					},
				},
			},
			prefix: "svc",
			expected: []attribute.KeyValue{
				attribute.String("svc.0.key", "value"),
			},
		},
		{
			name: "multiple attributers with multiple attributes",
			attributers: []Attributer{
				&mockAttributer{
					attrs: []attribute.KeyValue{
						attribute.String("key1", "value1"),
						attribute.Int("key2", 42),
					},
				},
				&mockAttributer{
					attrs: []attribute.KeyValue{
						attribute.Bool("key3", true),
					},
				},
			},
			prefix: "service",
			expected: []attribute.KeyValue{
				attribute.String("service.0.key1", "value1"),
				attribute.Int("service.0.key2", 42),
				attribute.Bool("service.1.key3", true),
			},
		},
		{
			name: "empty prefix",
			attributers: []Attributer{
				&mockAttributer{
					attrs: []attribute.KeyValue{
						attribute.String("key", "value"),
					},
				},
			},
			prefix: "",
			expected: []attribute.KeyValue{
				attribute.String("0.key", "value"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := PrefixAttributers(test.attributers, test.prefix)
			assert.Equal(t, test.expected, actual)
		})
	}
}

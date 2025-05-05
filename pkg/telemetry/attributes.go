package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

type Attributer interface {
	Attributes() []attribute.KeyValue
}

// PrefixAttributes adds a prefix to all attribute names while preserving their values
func PrefixAttributes(attrs []attribute.KeyValue, prefix string) []attribute.KeyValue {
	prefixed := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		prefixed[i] = attribute.KeyValue{
			Key:   attribute.Key(prefix + string(attr.Key)),
			Value: attr.Value,
		}
	}
	return prefixed
}

// PrefixAttributers adds a prefix and slice position to all attribute names from multiple Attributers.
// The format is "prefix.position.originalName" where position is the index in the slice.
func PrefixAttributers[T Attributer](attributers []T, prefix string) []attribute.KeyValue {
	var allAttrs []attribute.KeyValue
	for i, attributer := range attributers {
		attrs := attributer.Attributes()
		for _, attr := range attrs {
			allAttrs = append(allAttrs, attribute.KeyValue{
				Key:   attribute.Key(fmt.Sprintf("%s.%d.%s", prefix, i, attr.Key)),
				Value: attr.Value,
			})
		}
	}
	return allAttrs
}

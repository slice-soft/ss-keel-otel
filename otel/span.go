package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	apitrace "go.opentelemetry.io/otel/trace"

	"github.com/slice-soft/ss-keel-core/contracts"
)

var _ contracts.Span = (*spanWrapper)(nil)

// spanWrapper bridges the OTel SDK trace.Span to contracts.Span.
type spanWrapper struct {
	span apitrace.Span
}

// SetAttribute sets a key/value attribute on the span.
func (s *spanWrapper) SetAttribute(key string, value any) {
	s.span.SetAttributes(anyAttribute(key, value))
}

// RecordError marks the span as errored and records the error event.
func (s *spanWrapper) RecordError(err error) {
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// End finalizes the span.
func (s *spanWrapper) End() {
	s.span.End()
}

// anyAttribute converts a Go value to an OTel attribute.KeyValue.
func anyAttribute(key string, value any) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

// noopSpan is returned when telemetry is disabled.
type noopSpan struct{}

func (noopSpan) SetAttribute(_ string, _ any) {}
func (noopSpan) RecordError(_ error)          {}
func (noopSpan) End()                         {}

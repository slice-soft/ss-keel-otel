package otel

import (
	"context"
	"time"

	"github.com/slice-soft/ss-keel-core/contracts"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/codes"
)

// PanelID implements contracts.Debuggable.
func (p *Provider) PanelID() string { return "otel" }

// PanelLabel implements contracts.Debuggable.
func (p *Provider) PanelLabel() string { return "OpenTelemetry" }

// PanelEvents implements contracts.Debuggable.
func (p *Provider) PanelEvents() <-chan contracts.PanelEvent { return p.events }

// spanEventProcessor is a no-op SpanProcessor that emits a PanelEvent when a span ends.
type spanEventProcessor struct {
	events  chan contracts.PanelEvent
	addonID string
}

func newSpanEventProcessor(events chan contracts.PanelEvent, addonID string) *spanEventProcessor {
	return &spanEventProcessor{events: events, addonID: addonID}
}

func (ep *spanEventProcessor) OnStart(_ context.Context, _ sdktrace.ReadWriteSpan) {}

func (ep *spanEventProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	level := "info"
	if s.Status().Code == codes.Error {
		level = "error"
	}

	detail := map[string]any{
		"trace_id": s.SpanContext().TraceID().String(),
		"span_id":  s.SpanContext().SpanID().String(),
		"duration": s.EndTime().Sub(s.StartTime()).String(),
	}
	for _, attr := range s.Attributes() {
		detail[string(attr.Key)] = attr.Value.AsInterface()
	}

	event := contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   ep.addonID,
		Label:     s.Name(),
		Detail:    detail,
		Level:     level,
	}

	select {
	case ep.events <- event:
	default:
		// channel full — drop rather than block the span pipeline
	}
}

func (ep *spanEventProcessor) Shutdown(_ context.Context) error   { return nil }
func (ep *spanEventProcessor) ForceFlush(_ context.Context) error { return nil }

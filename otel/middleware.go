package otel

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Middleware returns a Fiber handler that creates an OTel span per HTTP request.
// It extracts incoming trace context from request headers (W3C TraceContext + Baggage)
// and propagates it through the Fiber user context so handlers can start child spans.
//
// Usage:
//
//	app.Fiber().Use(otelProvider.Middleware())
func (p *Provider) Middleware() fiber.Handler {
	if !p.IsEnabled() {
		return func(c *fiber.Ctx) error { return c.Next() }
	}

	tracer := otel.Tracer(p.config.ServiceName)
	propagator := otel.GetTextMapPropagator()

	return func(c *fiber.Ctx) error {
		carrier := fiberCarrier{c: c}
		ctx := propagator.Extract(c.UserContext(), carrier)

		route := c.Route().Path
		spanName := c.Method() + " " + route

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.URLPath(c.Path()),
				semconv.ServerAddress(c.Hostname()),
				attribute.String("http.route", route),
				attribute.String("net.peer.ip", c.IP()),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)

		err := c.Next()

		status := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else if status >= 500 {
			span.SetStatus(codes.Error, strconv.Itoa(status))
		}

		return err
	}
}

// fiberCarrier adapts Fiber's request/response headers to propagation.TextMapCarrier.
type fiberCarrier struct{ c *fiber.Ctx }

func (f fiberCarrier) Get(key string) string { return f.c.Get(key) }

func (f fiberCarrier) Set(key, val string) { f.c.Set(key, val) }

func (f fiberCarrier) Keys() []string {
	headers := f.c.GetReqHeaders()
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	return keys
}

var _ propagation.TextMapCarrier = fiberCarrier{}

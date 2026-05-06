package otel

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	keelcore "github.com/slice-soft/ss-keel-core/core"
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

		// Start with a placeholder name; the real route is only available after c.Next()
		// because Fiber resolves the matched route during handler dispatch.
		ctx, span := tracer.Start(ctx, c.Method()+" <resolving>",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.URLPath(strings.Clone(c.Path())),
				semconv.ServerAddress(c.Hostname()),
				attribute.String("net.peer.ip", c.IP()),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)

		err := c.Next()

		// Route is resolved after c.Next().
		route := c.Route().Path
		span.SetName(c.Method() + " " + route)
		span.SetAttributes(attribute.String("http.route", route))

		status := resolveStatus(c, err)
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

// resolveStatus returns the true HTTP status code for the request.
// c.Response().StatusCode() reads 200 before Fiber's error handler runs,
// so we inspect the returned error directly when one is present.
func resolveStatus(c *fiber.Ctx, err error) int {
	if err != nil {
		var ke *keelcore.KError
		if errors.As(err, &ke) {
			return ke.StatusCode
		}
		if fe, ok := err.(*fiber.Error); ok {
			return fe.Code
		}
	}
	return c.Response().StatusCode()
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

package otel

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	apitrace "go.opentelemetry.io/otel/trace"

	"github.com/slice-soft/ss-keel-core/contracts"
)

// Provider wraps the OpenTelemetry SDK and implements contracts.Tracer.
// When disabled, all operations are no-ops and no SDK components are initialized.
type Provider struct {
	tracer         apitrace.Tracer
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	config         Config
	events         chan contracts.PanelEvent
}

var (
	_ contracts.Addon        = (*Provider)(nil)
	_ contracts.Tracer       = (*Provider)(nil)
	_ contracts.Manifestable = (*Provider)(nil)
	_ contracts.Debuggable   = (*Provider)(nil)
)

// New initializes the OpenTelemetry SDK with the given config.
// If cfg.Enabled is false, a disabled no-op Provider is returned — no SDK
// components are created and no network connections are attempted.
//
// The OTLP exporters automatically read OTEL_EXPORTER_OTLP_ENDPOINT and
// OTEL_EXPORTER_OTLP_HEADERS from the environment.
func New(cfg Config) (*Provider, error) {
	cfg.withDefaults()

	events := make(chan contracts.PanelEvent, 512)

	if !cfg.Enabled {
		return &Provider{config: cfg, events: events}, nil
	}

	ctx := context.Background()

	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ss-keel-otel: build resource: %w", err)
	}

	sampler, err := buildSampler(cfg.SamplerType, cfg.SamplerArg)
	if err != nil {
		return nil, fmt.Errorf("ss-keel-otel: build sampler: %w", err)
	}

	traceExporter, err := buildTraceExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("ss-keel-otel: trace exporter: %w", err)
	}

	metricExporter, err := buildMetricExporter(ctx, cfg)
	if err != nil {
		_ = traceExporter.Shutdown(ctx)
		return nil, fmt.Errorf("ss-keel-otel: metric exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithSpanProcessor(newSpanEventProcessor(events, "otel")),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	// Set as global providers so OTel instrumentation libraries work out of the box.
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if cfg.Logger != nil {
		cfg.Logger.Info("otel enabled [service=%s version=%s env=%s protocol=%s sampler=%s]",
			cfg.ServiceName, cfg.ServiceVersion, cfg.Environment, cfg.ExporterProtocol, cfg.SamplerType)
	}

	return &Provider{
		tracer:         tracerProvider.Tracer(cfg.ServiceName),
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		config:         cfg,
		events:         events,
	}, nil
}

// Shutdown flushes and stops all telemetry pipelines.
// Register this with app.OnShutdown for graceful termination.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("ss-keel-otel: tracer provider shutdown: %w", err)
		}
	}
	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("ss-keel-otel: meter provider shutdown: %w", err)
		}
	}
	return nil
}

// Start implements contracts.Tracer.
// Returns a no-op span when the provider is disabled.
func (p *Provider) Start(ctx context.Context, name string) (context.Context, contracts.Span) {
	if p.tracer == nil {
		return ctx, noopSpan{}
	}
	ctx, span := p.tracer.Start(ctx, name)
	return ctx, &spanWrapper{span: span}
}

// IsEnabled reports whether the OpenTelemetry provider is active.
func (p *Provider) IsEnabled() bool {
	return p.tracerProvider != nil
}

// --- internal builders ---

func buildResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
		resource.WithProcess(),
		resource.WithHost(),
	)
}

func buildSampler(samplerType, samplerArg string) (sdktrace.Sampler, error) {
	switch strings.ToLower(strings.TrimSpace(samplerType)) {
	case SamplerAlwaysOn:
		return sdktrace.AlwaysSample(), nil
	case SamplerAlwaysOff:
		return sdktrace.NeverSample(), nil
	case SamplerTraceIDRatio:
		ratio, err := parseSamplerRatio(samplerArg)
		if err != nil {
			return nil, err
		}
		return sdktrace.TraceIDRatioBased(ratio), nil
	case SamplerParentBasedTraceIDRatio:
		ratio, err := parseSamplerRatio(samplerArg)
		if err != nil {
			return nil, err
		}
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)), nil
	default: // parentbased_always_on and any unrecognised value
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	}
}

func parseSamplerRatio(arg string) (float64, error) {
	if arg == "" {
		return 1.0, nil
	}
	ratio, err := strconv.ParseFloat(strings.TrimSpace(arg), 64)
	if err != nil || ratio < 0 || ratio > 1 {
		return 0, fmt.Errorf("invalid sampler ratio %q: must be a float between 0.0 and 1.0", arg)
	}
	return ratio, nil
}

func buildTraceExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.isGRPC() {
		return otlptracegrpc.New(ctx)
	}
	return otlptracehttp.New(ctx)
}

func buildMetricExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	if cfg.isGRPC() {
		return otlpmetricgrpc.New(ctx)
	}
	return otlpmetrichttp.New(ctx)
}

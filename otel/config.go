package otel

import (
	"strings"

	"github.com/slice-soft/ss-keel-core/contracts"
)

const (
	SamplerAlwaysOn                = "always_on"
	SamplerAlwaysOff               = "always_off"
	SamplerParentBasedAlwaysOn     = "parentbased_always_on"
	SamplerTraceIDRatio            = "traceidratio"
	SamplerParentBasedTraceIDRatio = "parentbased_traceidratio"

	ProtocolHTTP = "http/protobuf"
	ProtocolGRPC = "grpc"
)

// Config configures the OpenTelemetry provider.
type Config struct {
	// Enabled controls whether telemetry is active. Defaults to false.
	Enabled bool `keel:"otel.enabled"`
	// ServiceName is the logical name of the service.
	ServiceName string `keel:"otel.service-name,required"`
	// ServiceVersion is the version of the service (e.g. "1.2.3").
	ServiceVersion string `keel:"otel.service-version"`
	// Environment is the deployment environment (e.g. "production", "staging").
	Environment string `keel:"otel.environment"`
	// ExporterProtocol selects the OTLP transport: "http/protobuf" (default) or "grpc".
	// Maps to OTEL_EXPORTER_OTLP_PROTOCOL. The OTLP exporters automatically read
	// OTEL_EXPORTER_OTLP_ENDPOINT and OTEL_EXPORTER_OTLP_HEADERS from the environment.
	ExporterProtocol string `keel:"otel.exporter-otlp-protocol"`
	// SamplerType is the trace sampler strategy. Defaults to "parentbased_always_on".
	// Maps to OTEL_TRACES_SAMPLER.
	SamplerType string `keel:"otel.traces-sampler"`
	// SamplerArg is the sampler argument (e.g. "0.1" for 10% ratio sampling).
	// Maps to OTEL_TRACES_SAMPLER_ARG.
	SamplerArg string `keel:"otel.traces-sampler-arg"`

	Logger contracts.Logger
}

func (c *Config) withDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "keel-app"
	}
	if c.ServiceVersion == "" {
		c.ServiceVersion = "0.0.0"
	}
	if c.Environment == "" {
		c.Environment = "development"
	}
	if c.ExporterProtocol == "" {
		c.ExporterProtocol = ProtocolHTTP
	}
	if c.SamplerType == "" {
		c.SamplerType = SamplerParentBasedAlwaysOn
	}
}

func (c *Config) isGRPC() bool {
	return strings.EqualFold(strings.TrimSpace(c.ExporterProtocol), ProtocolGRPC)
}

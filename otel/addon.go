package otel

import "github.com/slice-soft/ss-keel-core/contracts"

// ID implements contracts.Addon.
func (p *Provider) ID() string { return "otel" }

// Manifest implements contracts.Manifestable.
func (p *Provider) Manifest() contracts.AddonManifest {
	return contracts.AddonManifest{
		ID:           "otel",
		Version:      "0.1.0",
		Capabilities: []string{"observability"},
		Resources:    []string{},
		EnvVars: []contracts.EnvVar{
			{
				Key:         "OTEL_ENABLED",
				ConfigKey:   "otel.enabled",
				Description: "Enable OpenTelemetry instrumentation",
				Required:    false,
				Default:     "false",
				Source:      "otel",
			},
			{
				Key:         "OTEL_SERVICE_NAME",
				ConfigKey:   "otel.service-name",
				Description: "Logical service name reported in all telemetry signals",
				Required:    true,
				Source:      "otel",
			},
			{
				Key:         "OTEL_SERVICE_VERSION",
				ConfigKey:   "otel.service-version",
				Description: "Service version reported as a resource attribute",
				Default:     "0.0.0",
				Source:      "otel",
			},
			{
				Key:         "OTEL_ENVIRONMENT",
				ConfigKey:   "otel.environment",
				Description: "Deployment environment (e.g. production, staging)",
				Default:     "development",
				Source:      "otel",
			},
			{
				Key:         "OTEL_EXPORTER_OTLP_ENDPOINT",
				ConfigKey:   "otel.exporter-otlp-endpoint",
				Description: "OTLP collector endpoint. Read directly by the SDK exporter.",
				Default:     "http://localhost:4318",
				Source:      "otel",
			},
			{
				Key:         "OTEL_EXPORTER_OTLP_PROTOCOL",
				ConfigKey:   "otel.exporter-otlp-protocol",
				Description: "OTLP transport protocol: http/protobuf (default) or grpc",
				Default:     "http/protobuf",
				Source:      "otel",
			},
			{
				Key:         "OTEL_TRACES_SAMPLER",
				ConfigKey:   "otel.traces-sampler",
				Description: "Trace sampler: always_on, always_off, parentbased_always_on, traceidratio, parentbased_traceidratio",
				Default:     "parentbased_always_on",
				Source:      "otel",
			},
			{
				Key:         "OTEL_TRACES_SAMPLER_ARG",
				ConfigKey:   "otel.traces-sampler-arg",
				Description: "Sampler argument — ratio value (0.0–1.0) for traceidratio samplers",
				Source:      "otel",
			},
		},
	}
}

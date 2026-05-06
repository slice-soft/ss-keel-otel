package otel

import "testing"

func TestConfigWithDefaults_AppliesExpectedDefaults(t *testing.T) {
	cfg := Config{}
	cfg.withDefaults()

	if cfg.ServiceName != "keel-app" {
		t.Fatalf("expected ServiceName %q, got %q", "keel-app", cfg.ServiceName)
	}
	if cfg.ServiceVersion != "0.0.0" {
		t.Fatalf("expected ServiceVersion %q, got %q", "0.0.0", cfg.ServiceVersion)
	}
	if cfg.Environment != "development" {
		t.Fatalf("expected Environment %q, got %q", "development", cfg.Environment)
	}
	if cfg.ExporterProtocol != ProtocolHTTP {
		t.Fatalf("expected ExporterProtocol %q, got %q", ProtocolHTTP, cfg.ExporterProtocol)
	}
	if cfg.SamplerType != SamplerParentBasedAlwaysOn {
		t.Fatalf("expected SamplerType %q, got %q", SamplerParentBasedAlwaysOn, cfg.SamplerType)
	}
}

func TestConfigWithDefaults_PreservesConfiguredValues(t *testing.T) {
	cfg := Config{
		ServiceName:      "my-service",
		ServiceVersion:   "2.1.0",
		Environment:      "staging",
		ExporterProtocol: ProtocolGRPC,
		SamplerType:      SamplerTraceIDRatio,
		SamplerArg:       "0.5",
	}
	cfg.withDefaults()

	if cfg.ServiceName != "my-service" {
		t.Fatalf("expected ServiceName preserved, got %q", cfg.ServiceName)
	}
	if cfg.ServiceVersion != "2.1.0" {
		t.Fatalf("expected ServiceVersion preserved, got %q", cfg.ServiceVersion)
	}
	if cfg.Environment != "staging" {
		t.Fatalf("expected Environment preserved, got %q", cfg.Environment)
	}
	if cfg.ExporterProtocol != ProtocolGRPC {
		t.Fatalf("expected ExporterProtocol preserved, got %q", cfg.ExporterProtocol)
	}
	if cfg.SamplerType != SamplerTraceIDRatio {
		t.Fatalf("expected SamplerType preserved, got %q", cfg.SamplerType)
	}
	if cfg.SamplerArg != "0.5" {
		t.Fatalf("expected SamplerArg preserved, got %q", cfg.SamplerArg)
	}
}

func TestConfig_IsGRPC(t *testing.T) {
	tests := []struct {
		protocol string
		want     bool
	}{
		{ProtocolGRPC, true},
		{"GRPC", true},
		{ProtocolHTTP, false},
		{"http/json", false},
		{"", false},
	}

	for _, tt := range tests {
		cfg := Config{ExporterProtocol: tt.protocol}
		if got := cfg.isGRPC(); got != tt.want {
			t.Errorf("isGRPC(%q) = %v, want %v", tt.protocol, got, tt.want)
		}
	}
}

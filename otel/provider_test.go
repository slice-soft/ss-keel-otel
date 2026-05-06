package otel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/slice-soft/ss-keel-core/contracts"
)

// Compile-time assertions.
var (
	_ contracts.Addon        = (*Provider)(nil)
	_ contracts.Tracer       = (*Provider)(nil)
	_ contracts.Manifestable = (*Provider)(nil)
)

// --- New ---

func TestNew_DisabledReturnsProvider(t *testing.T) {
	p, err := New(Config{Enabled: false, ServiceName: "test"})
	if err != nil {
		t.Fatalf("New returned error for disabled config: %v", err)
	}
	if p == nil {
		t.Fatal("New returned nil provider")
	}
	if p.IsEnabled() {
		t.Fatal("expected disabled provider to report IsEnabled() = false")
	}
}

func TestNew_DisabledSkipsSDKInit(t *testing.T) {
	// Must not attempt any network connection — unreachable endpoint is fine.
	p, err := New(Config{
		Enabled:     false,
		ServiceName: "test",
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if p.tracerProvider != nil {
		t.Fatal("expected tracerProvider to be nil when disabled")
	}
	if p.meterProvider != nil {
		t.Fatal("expected meterProvider to be nil when disabled")
	}
}

func TestNew_InvalidSamplerRatio(t *testing.T) {
	_, err := New(Config{
		Enabled:     true,
		ServiceName: "test",
		SamplerType: SamplerTraceIDRatio,
		SamplerArg:  "not-a-number",
	})
	if err == nil {
		t.Fatal("expected error for invalid sampler ratio")
	}
}

func TestNew_SamplerRatioOutOfRange(t *testing.T) {
	_, err := New(Config{
		Enabled:     true,
		ServiceName: "test",
		SamplerType: SamplerTraceIDRatio,
		SamplerArg:  "1.5",
	})
	if err == nil {
		t.Fatal("expected error for ratio > 1.0")
	}
}

// --- Provider.Start ---

func TestProvider_StartReturnsNoopWhenDisabled(t *testing.T) {
	p, err := New(Config{Enabled: false, ServiceName: "test"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	ctx := context.Background()
	_, span := p.Start(ctx, "test-span")
	if span == nil {
		t.Fatal("Start returned nil span")
	}

	// noop span must not panic on any operation
	span.SetAttribute("key", "value")
	span.RecordError(errors.New("test error"))
	span.End()
}

// --- Provider.Shutdown ---

func TestProvider_Shutdown_WhenDisabled(t *testing.T) {
	p, err := New(Config{Enabled: false, ServiceName: "test"})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if err := p.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error for disabled provider: %v", err)
	}
}

// --- contracts.Addon ---

func TestProvider_IDReturnsOtel(t *testing.T) {
	p, _ := New(Config{Enabled: false, ServiceName: "test"})
	if got := p.ID(); got != "otel" {
		t.Errorf("ID() = %q, want %q", got, "otel")
	}
}

// --- contracts.Manifestable ---

func TestProvider_ManifestIsComplete(t *testing.T) {
	p, _ := New(Config{Enabled: false, ServiceName: "test"})
	m := p.Manifest()

	if m.ID != "otel" {
		t.Errorf("Manifest().ID = %q, want %q", m.ID, "otel")
	}
	if m.Version == "" {
		t.Error("Manifest().Version must not be empty")
	}
	if len(m.Capabilities) == 0 {
		t.Error("Manifest().Capabilities must not be empty")
	}
	if m.EnvVars == nil {
		t.Error("Manifest().EnvVars must not be nil")
	}
	if len(m.EnvVars) == 0 {
		t.Error("Manifest().EnvVars must declare at least one env var")
	}
}

// --- buildSampler ---

func TestBuildSampler_KnownTypes(t *testing.T) {
	types := []string{
		SamplerAlwaysOn,
		SamplerAlwaysOff,
		SamplerParentBasedAlwaysOn,
		SamplerTraceIDRatio,
		SamplerParentBasedTraceIDRatio,
		"unknown-type-falls-back-to-parentbased",
	}

	for _, st := range types {
		arg := ""
		if st == SamplerTraceIDRatio || st == SamplerParentBasedTraceIDRatio {
			arg = "0.5"
		}
		sampler, err := buildSampler(st, arg)
		if err != nil {
			t.Errorf("buildSampler(%q, %q) returned error: %v", st, arg, err)
		}
		if sampler == nil {
			t.Errorf("buildSampler(%q, %q) returned nil", st, arg)
		}
	}
}

func TestBuildSampler_InvalidRatio(t *testing.T) {
	cases := []string{"not-a-float", "1.5", "-0.1"}
	for _, arg := range cases {
		_, err := buildSampler(SamplerTraceIDRatio, arg)
		if err == nil {
			t.Errorf("expected error for sampler arg %q", arg)
		}
	}
}

func TestBuildSampler_EmptyRatioDefaultsToOne(t *testing.T) {
	s, err := buildSampler(SamplerTraceIDRatio, "")
	if err != nil {
		t.Fatalf("buildSampler returned error for empty ratio: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sampler")
	}
}

// --- spanWrapper ---

func TestSpanWrapper_SetAttribute_DoesNotPanic(t *testing.T) {
	types := []any{"string", 42, int64(99), 3.14, true, struct{}{}}
	for _, v := range types {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetAttribute panicked for value %T: %v", v, r)
				}
			}()
			kv := anyAttribute("key", v)
			_ = kv
		}()
	}
}

func TestAnyAttribute_StringTypes(t *testing.T) {
	cases := []struct {
		value    any
		wantType string
	}{
		{"hello", "STRING"},
		{42, "INT64"},
		{int64(99), "INT64"},
		{3.14, "FLOAT64"},
		{true, "BOOL"},
		{struct{ X int }{1}, "STRING"}, // fallback to fmt.Sprintf
	}

	for _, tc := range cases {
		kv := anyAttribute("k", tc.value)
		if kv.Value.Type().String() != tc.wantType {
			t.Errorf("anyAttribute(%T) type = %q, want %q", tc.value, kv.Value.Type().String(), tc.wantType)
		}
	}
}

// --- testLogger ---

type testLogger struct{ infos []string }

func (l *testLogger) Info(format string, args ...interface{}) {
	l.infos = append(l.infos, fmt.Sprintf(format, args...))
}
func (l *testLogger) Warn(string, ...interface{})  {}
func (l *testLogger) Error(string, ...interface{}) {}
func (l *testLogger) Debug(string, ...interface{}) {}

package myaddon

import (
	"testing"
	"time"

	"github.com/slice-soft/ss-keel-core/contracts"
)

// Compile-time assertions (repeated here as living documentation).
var (
	_ contracts.Addon        = (*MyAddon)(nil)
	_ contracts.Debuggable   = (*MyAddon)(nil)
	_ contracts.Manifestable = (*MyAddon)(nil)
)

// panelRegistryMock records which addons were registered.
type panelRegistryMock struct {
	registered []contracts.Debuggable
}

func (m *panelRegistryMock) RegisterAddon(d contracts.Debuggable) {
	m.registered = append(m.registered, d)
}

func TestAddon_ID(t *testing.T) {
	a := New()
	if got := a.ID(); got != "my-addon" {
		t.Errorf("ID() = %q, want %q", got, "my-addon")
	}
}

func TestAddon_PanelID(t *testing.T) {
	a := New()
	if got := a.PanelID(); got != "my-addon" {
		t.Errorf("PanelID() = %q, want %q", got, "my-addon")
	}
}

func TestAddon_PanelLabel(t *testing.T) {
	a := New()
	if got := a.PanelLabel(); got != "My Addon" {
		t.Errorf("PanelLabel() = %q, want %q", got, "My Addon")
	}
}

func TestAddon_Manifest(t *testing.T) {
	a := New()
	m := a.Manifest()

	if m.ID != "my-addon" {
		t.Errorf("Manifest().ID = %q, want %q", m.ID, "my-addon")
	}
	if m.Version == "" {
		t.Error("Manifest().Version must not be empty")
	}
	if m.Capabilities == nil {
		t.Error("Manifest().Capabilities must not be nil")
	}
	if m.Resources == nil {
		t.Error("Manifest().Resources must not be nil")
	}
	if m.EnvVars == nil {
		t.Error("Manifest().EnvVars must not be nil")
	}
}

func TestAddon_PanelEvents_ReturnsReadableChannel(t *testing.T) {
	a := New()
	ch := a.PanelEvents()
	if ch == nil {
		t.Fatal("PanelEvents() returned nil channel")
	}

	// Verify the channel is readable by sending and receiving an event.
	a.events <- contracts.PanelEvent{AddonID: "my-addon", Label: "test"}
	select {
	case e := <-ch:
		if e.AddonID != "my-addon" || e.Label != "test" {
			t.Errorf("received unexpected event: %+v", e)
		}
	default:
		t.Fatal("expected to receive event from PanelEvents channel")
	}
}

func TestAddon_TryEmit_EmitsEvent(t *testing.T) {
	a := New()

	want := contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "my-addon",
		Label:     "example",
		Detail:    map[string]any{"key": "value"},
		Level:     "info",
	}
	a.tryEmit(want)

	select {
	case got := <-a.PanelEvents():
		if got.AddonID != want.AddonID {
			t.Errorf("AddonID = %q, want %q", got.AddonID, want.AddonID)
		}
		if got.Label != want.Label {
			t.Errorf("Label = %q, want %q", got.Label, want.Label)
		}
		if got.Level != want.Level {
			t.Errorf("Level = %q, want %q", got.Level, want.Level)
		}
	default:
		t.Fatal("expected event to be received from PanelEvents channel")
	}
}

func TestAddon_TryEmit_DoesNotBlockWhenFull(t *testing.T) {
	a := New()

	// Send 300 events into a 256-buffer channel — must never block.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 300; i++ {
			a.tryEmit(contracts.PanelEvent{
				AddonID: "my-addon",
				Label:   "example",
				Level:   "info",
			})
		}
	}()

	select {
	case <-done:
		// success — tryEmit did not block
	case <-time.After(2 * time.Second):
		t.Fatal("tryEmit blocked when channel was full")
	}
}

func TestAddon_RegisterWithPanel(t *testing.T) {
	a := New()
	registry := &panelRegistryMock{}

	a.RegisterWithPanel(registry)

	if len(registry.registered) != 1 {
		t.Fatalf("expected 1 registered addon, got %d", len(registry.registered))
	}
	if registry.registered[0] != a {
		t.Error("RegisterWithPanel did not register the addon itself")
	}
}

func TestAddon_EmitEvent_SetsFields(t *testing.T) {
	a := New()

	detail := map[string]any{"duration_ms": int64(42)}
	a.emitEvent("example", detail, "info")

	select {
	case e := <-a.PanelEvents():
		if e.AddonID != "my-addon" {
			t.Errorf("AddonID = %q, want %q", e.AddonID, "my-addon")
		}
		if e.Label != "example" {
			t.Errorf("Label = %q, want %q", e.Label, "example")
		}
		if e.Level != "info" {
			t.Errorf("Level = %q, want %q", e.Level, "info")
		}
		if e.Timestamp.IsZero() {
			t.Error("Timestamp must not be zero")
		}
		if e.Detail["duration_ms"] != int64(42) {
			t.Errorf("Detail[duration_ms] = %v, want 42", e.Detail["duration_ms"])
		}
	default:
		t.Fatal("expected event to be received from PanelEvents channel")
	}
}

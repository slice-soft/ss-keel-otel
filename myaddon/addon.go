// Package myaddon is the scaffold for a Keel addon.
// Rename this package and the MyAddon type to match your addon's domain.
package myaddon

import (
	"time"

	"github.com/slice-soft/ss-keel-core/contracts"
)

// Compile-time assertions — the build will fail here if MyAddon stops
// satisfying any of the required contracts.
var (
	_ contracts.Addon        = (*MyAddon)(nil)
	_ contracts.Debuggable   = (*MyAddon)(nil)
	_ contracts.Manifestable = (*MyAddon)(nil)
)

// MyAddon is the main addon struct.
// Rename it and add your domain-specific fields (clients, config, etc.).
type MyAddon struct {
	events chan contracts.PanelEvent
}

// New creates a new MyAddon instance.
func New() *MyAddon {
	return &MyAddon{
		events: make(chan contracts.PanelEvent, 256),
	}
}

// --- contracts.Addon ---

// ID returns the unique identifier for this addon.
// It must match the "id" field in keel-addon.json.
func (a *MyAddon) ID() string { return "my-addon" }

// --- contracts.Debuggable ---

// PanelID returns the identifier used by the dev panel to key this addon.
func (a *MyAddon) PanelID() string { return "my-addon" }

// PanelLabel returns the human-readable name shown in the dev panel sidebar.
func (a *MyAddon) PanelLabel() string { return "My Addon" }

// PanelEvents returns the read-only channel of observable events.
// The dev panel consumes this channel and owns all rendering decisions.
func (a *MyAddon) PanelEvents() <-chan contracts.PanelEvent { return a.events }

// --- contracts.Manifestable ---

// Manifest describes the capabilities, resources, and config-facing env vars
// of this addon. The CLI and dev panel consume this metadata.
func (a *MyAddon) Manifest() contracts.AddonManifest {
	return contracts.AddonManifest{
		ID:      "my-addon",
		Version: "0.1.0",
		// Capabilities declares what this addon provides.
		// Common values: "database", "cache", "queue", "auth", "scheduler"
		Capabilities: []string{},
		// Resources declares external services this addon depends on.
		// Common values: "postgres", "redis", "mongodb", "rabbitmq"
		Resources: []string{},
		EnvVars:   []contracts.EnvVar{
			// Declare every environment variable your addon reads.
			// Example:
			// {
			// 	Key:         "MY_ADDON_URL",
			// 	ConfigKey:   "my-addon.url",
			// 	Description: "Connection URL for My Addon.",
			// 	Required:    false,
			// 	Secret:      true,
			// 	Default:     "http://localhost:8080",
			// 	Source:      "my-addon",
			// },
		},
	}
}

// --- Panel registration ---

// RegisterWithPanel registers this addon with the devpanel PanelRegistry so
// the panel can consume its event stream. Call this from your provider setup
// after the addon is fully initialized:
//
//	if panel, ok := app.GetAddon("devpanel").(contracts.PanelRegistry); ok {
//	    myAddon.RegisterWithPanel(panel)
//	}
func (a *MyAddon) RegisterWithPanel(r contracts.PanelRegistry) {
	r.RegisterAddon(a)
}

// --- Internal helpers ---

// tryEmit sends e to the events channel without blocking.
// If the 256-event buffer is full the event is dropped rather than blocking
// the caller — correctness of your addon must never depend on event delivery.
func (a *MyAddon) tryEmit(e contracts.PanelEvent) {
	select {
	case a.events <- e:
	default:
	}
}

// emitEvent is an example of how to publish an observable event.
// Replace this with real domain events (e.g. query timing, cache hit/miss).
// Never include sensitive values (passwords, tokens, PII) in label or detail.
func (a *MyAddon) emitEvent(label string, detail map[string]any, level string) {
	a.tryEmit(contracts.PanelEvent{
		Timestamp: time.Now(),
		AddonID:   "my-addon",
		Label:     label,
		Detail:    detail,
		Level:     level,
	})
}

// --- Optional: DebuggableWithView ---
//
// Implement contracts.DebuggableWithView if you want to render a custom view
// in the dev panel instead of the default key/value table. Requires the
// a-h/templ package (any templ.Component satisfies contracts.PanelComponent).
//
// Step 1 — add the compile-time assertion:
//   var _ contracts.DebuggableWithView = (*MyAddon)(nil)
//
// Step 2 — create a Templ component (e.g. views/panel.templ):
//   package views
//   templ Panel(events []SomeEvent) { ... }
//
// Step 3 — implement PanelView():
//   func (a *MyAddon) PanelView() contracts.PanelComponent {
//       return views.Panel(a.snapshot())
//   }

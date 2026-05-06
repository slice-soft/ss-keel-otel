<img src="https://cdn.slicesoft.dev/boat.svg" width="400" />

# ss-keel-otel
Official OpenTelemetry addon for Keel — traces, metrics, and automatic Fiber HTTP spans via OTLP.

[![CI](https://github.com/slice-soft/ss-keel-otel/actions/workflows/ci.yml/badge.svg)](https://github.com/slice-soft/ss-keel-otel/actions)
[![Release](https://img.shields.io/github/v/release/slice-soft/ss-keel-otel)](https://github.com/slice-soft/ss-keel-otel/releases)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/slice-soft/ss-keel-otel)](https://goreportcard.com/report/github.com/slice-soft/ss-keel-otel)
[![Go Reference](https://pkg.go.dev/badge/github.com/slice-soft/ss-keel-otel.svg)](https://pkg.go.dev/github.com/slice-soft/ss-keel-otel)
![License](https://img.shields.io/badge/License-MIT-green)
![Made in Colombia](https://img.shields.io/badge/Made%20in-Colombia-FCD116?labelColor=003893)


## OpenTelemetry addon for Keel

`ss-keel-otel` adds distributed tracing and metrics to a [Keel](https://keel-go.dev) application via the [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/).
It implements `contracts.Tracer` from `ss-keel-core`, registers automatic HTTP spans for every Fiber request, and exports telemetry through OTLP to any compatible observability backend — Grafana, Jaeger, Datadog, New Relic, AWS X-Ray, Honeycomb, and more.

When `OTEL_ENABLED=false` (the default) the addon is completely inert — no SDK components are initialized, no connections are made, and all spans are no-ops.

---

## 🚀 Installation

```bash
keel add otel
```

The Keel CLI will:
1. Add `github.com/slice-soft/ss-keel-otel` as a dependency.
2. Create `cmd/setup_otel.go` and inject initialization code into `cmd/main.go`.
3. Register the Fiber HTTP middleware and OTel shutdown hook automatically.
4. Add all required environment variable examples to `.env` and `.env.example`.

---

## ⚙️ Configuration

Configuration is loaded through Keel's typed config system (`application.properties` + env vars):

```go
provider, err := ssotel.New(ssotel.Config{
    Enabled:          true,
    ServiceName:      "my-api",
    ServiceVersion:   "1.4.2",
    Environment:      "production",
    ExporterProtocol: ssotel.ProtocolHTTP, // or ssotel.ProtocolGRPC
    SamplerType:      ssotel.SamplerParentBasedAlwaysOn,
    Logger:           log,
})
```

| Field | Default | Description |
|---|---|---|
| `Enabled` | `false` | Master switch — must be `true` to activate |
| `ServiceName` | `keel-app` | Logical service name shown in dashboards |
| `ServiceVersion` | `0.0.0` | Reported as a resource attribute |
| `Environment` | `development` | Deployment environment |
| `ExporterProtocol` | `http/protobuf` | `http/protobuf` or `grpc` |
| `SamplerType` | `parentbased_always_on` | Sampler strategy |
| `SamplerArg` | — | Ratio value for ratio-based samplers |

---

## 🌐 Environment variables

The OTLP exporters read standard OTel env vars directly — no extra config needed:

| Variable | Default | Description |
|---|---|---|
| `OTEL_ENABLED` | `false` | Enable telemetry |
| `OTEL_SERVICE_NAME` | `keel-app` | Service name |
| `OTEL_SERVICE_VERSION` | `0.0.0` | Service version |
| `OTEL_ENVIRONMENT` | `development` | Deployment environment |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4318` | Collector endpoint |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf` | `http/protobuf` or `grpc` |
| `OTEL_EXPORTER_OTLP_HEADERS` | — | Auth headers (e.g. `Authorization=Bearer <token>`) |
| `OTEL_TRACES_SAMPLER` | `parentbased_always_on` | Sampler strategy |
| `OTEL_TRACES_SAMPLER_ARG` | — | Sampler ratio (e.g. `0.1` = 10%) |

---

## 📡 HTTP middleware

The addon automatically registers a Fiber middleware that creates a server span for every HTTP request:

```go
// Registered automatically by setupOtel()
app.Fiber().Use(provider.Middleware())
```

Each request span includes: HTTP method, URL path, route pattern, status code, client IP, and error recording.
Incoming trace context is extracted from `traceparent` / `baggage` headers (W3C TraceContext).

---

## 🔍 Manual spans

Use `app.Tracer()` to start child spans anywhere in your application:

```go
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    ctx, span := app.Tracer().Start(ctx, "UserService.GetByID")
    defer span.End()

    span.SetAttribute("user.id", id)

    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }
    return user, nil
}
```

In Fiber handlers, pass `c.UserContext()` as the parent context:

```go
func (h *Handler) GetUser(c *httpx.Ctx) error {
    ctx, span := app.Tracer().Start(c.UserContext(), "GetUser")
    defer span.End()
    // ...
}
```

---

## 📊 Samplers

| Value | Description |
|---|---|
| `always_on` | Sample every trace (high volume — development only) |
| `always_off` | Drop all traces |
| `parentbased_always_on` | Follow parent decision; sample root spans always (default) |
| `traceidratio` | Sample a fraction of root spans independently |
| `parentbased_traceidratio` | Follow parent decision; sample root spans at the given ratio |

Production recommendation: `parentbased_traceidratio` with `OTEL_TRACES_SAMPLER_ARG=0.1` (10%).

---

## 🔌 Backends

The addon exports via OTLP — compatible with any standard backend:

| Backend | OTLP endpoint | Notes |
|---|---|---|
| **Grafana** | Grafana Alloy / Grafana Cloud OTLP | Traces → Tempo, Metrics → Prometheus/Mimir |
| **Jaeger** | `http://localhost:4318` | Enable OTLP receiver in Jaeger |
| **Datadog** | Datadog Agent OTLP receiver | Set `DD_OTLP_CONFIG_RECEIVER_PROTOCOLS_HTTP_ENDPOINT` |
| **New Relic** | `https://otlp.nr-data.net` | Add `api-key` header |
| **Honeycomb** | `https://api.honeycomb.io` | Add `x-honeycomb-team` header |
| **AWS X-Ray** | OTel Collector with X-Ray exporter | Use ADOT collector |
| **OTel Collector** | `http://localhost:4318` | Relay to any backend |

---

## 🤚 CI/CD and releases

- **CI** runs on every pull request targeting `main` via `.github/workflows/ci.yml`.
- **Releases** are created automatically on merge to `main` via `.github/workflows/release.yml` using Release Please.

---

## 💡 Recommendations

* Keep `OTEL_ENABLED=false` in development and enable per environment — telemetry should never break a startup.
* Use `parentbased_traceidratio` with a ratio of `0.01`–`0.1` in production to control cost.
* Accept `contracts.Tracer` in your services — not `*ssotel.Provider` — so you can swap the implementation in tests.

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for setup and repository-specific rules.
The base workflow, commit conventions, and community standards live in [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md).

## Community

| Document | |
|---|---|
| [CONTRIBUTING.md](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) | Workflow, commit conventions, and PR guidelines |
| [GOVERNANCE.md](https://github.com/slice-soft/ss-community/blob/main/GOVERNANCE.md) | Decision-making, roles, and release process |
| [CODE_OF_CONDUCT.md](https://github.com/slice-soft/ss-community/blob/main/CODE_OF_CONDUCT.md) | Community standards |
| [VERSIONING.md](https://github.com/slice-soft/ss-community/blob/main/VERSIONING.md) | SemVer policy and breaking changes |
| [SECURITY.md](https://github.com/slice-soft/ss-community/blob/main/SECURITY.md) | How to report vulnerabilities |
| [MAINTAINERS.md](https://github.com/slice-soft/ss-community/blob/main/MAINTAINERS.md) | Active maintainers |

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- Website: [keel-go.dev](https://keel-go.dev)
- GitHub: [github.com/slice-soft/ss-keel-otel](https://github.com/slice-soft/ss-keel-otel)
- Documentation: [docs.keel-go.dev](https://docs.keel-go.dev)

---

Made by [SliceSoft](https://slicesoft.dev) — Colombia 💙

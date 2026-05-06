# Contributing to ss-keel-otel

The base contributing guide — workflow, commit conventions, PR guidelines, and community standards — lives in the [ss-community](https://github.com/slice-soft/ss-community/blob/main/CONTRIBUTING.md) repository. Read it first.

This document covers only what is specific to this repository.

---

## Getting Started

> Requirements
> - Go 1.25+
> - Git
> - An OTLP-compatible collector (optional, only needed for integration testing with a real backend)

1. **Fork the repository**: [github.com/slice-soft/ss-keel-otel](https://github.com/slice-soft/ss-keel-otel)
2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ss-keel-otel.git
   cd ss-keel-otel
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Create a branch**
   ```bash
   git checkout -b feat/your-feature-name
   ```

### Testing

The test suite does not require any external service. The OTel SDK is tested
via the disabled-provider path (no network connections), sampler logic, and
span wrapper behaviour.

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

To test against a real OTLP collector locally, start one via Docker:

```bash
# Jaeger (all-in-one with OTLP HTTP on port 4318)
docker run --rm -p 4318:4318 -p 16686:16686 jaegertracing/all-in-one

# Then run your Keel app with:
OTEL_ENABLED=true OTEL_SERVICE_NAME=test OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 go run ./cmd
```

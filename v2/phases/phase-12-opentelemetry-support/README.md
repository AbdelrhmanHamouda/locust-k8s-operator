# Phase 12: OpenTelemetry Support

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Status:** Not Started  
**Dependencies:** Phase 7 (v2 API Types)

---

## Objective

Replace the Prometheus metrics exporter sidecar with native Locust OpenTelemetry integration. This enables richer observability with traces and metrics exported to any OTel-compatible backend, while maintaining backward compatibility with the existing sidecar approach.

---

## Requirements Reference

- **REQUIREMENTS.md §5.1.1:** Native OpenTelemetry Support
- **Issue #72:** Metrics Exporter Sidecar Enhancement
- **analysis/LOCUST_FEATURES.md §1.1:** OpenTelemetry Integration

---

## Background

### Locust OTel Support (v2.42.4+)

Locust natively supports OpenTelemetry for exporting traces and metrics:

```bash
pip install locust[otel]
locust --otel
```

**Configuration via Environment Variables:**
```bash
OTEL_TRACES_EXPORTER=otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_EXPORTER_OTLP_ENDPOINT=https://otel-collector:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc  # or http/protobuf
```

### Current Implementation

The operator deploys a `containersol/locust_exporter:v0.5.0` sidecar with the master pod to export Prometheus metrics. With native OTel support, this sidecar becomes optional.

---

## Scope

### In Scope

- Add `--otel` flag to Locust command when OTel is enabled
- Inject OTel environment variables (`OTEL_*`) into Locust containers
- Conditionally skip metrics exporter sidecar when OTel is enabled
- Add validation for OTel configuration (endpoint required when enabled)
- Unit and integration tests for OTel support

### Out of Scope

- Deploying an OTel Collector (user-provided infrastructure)
- OTel SDK auto-instrumentation beyond Locust's native support
- Custom span/metric processors

---

## Key Deliverables

| File | Description |
|------|-------------|
| `internal/resources/otel.go` | OTel environment variable building functions |
| `internal/resources/otel_test.go` | Unit tests for OTel builders |
| `internal/resources/command.go` | Updated to add `--otel` flag |
| `internal/resources/job.go` | Conditionally skip metrics sidecar |
| `internal/resources/service.go` | Conditionally exclude metrics port |
| `api/v2/locusttest_webhook.go` | OTel configuration validation |
| `config/samples/locust_v2_locusttest_with_otel.yaml` | Sample CR with OTel |

---

## Success Criteria

1. `--otel` flag added to master and worker commands when `observability.openTelemetry.enabled: true`
2. OTel environment variables correctly injected into all Locust containers
3. Metrics exporter sidecar NOT deployed when OTel is enabled
4. Metrics port excluded from Service when OTel is enabled
5. Metrics exporter sidecar still deployed when OTel is disabled (backward compatible)
6. Validation rejects OTel config with `enabled: true` but no endpoint
7. All tests pass with ≥80% coverage for new code

---

## API Types (Already Defined)

The v2 API types are already defined in `api/v2/locusttest_types.go`:

```go
// ObservabilityConfig defines observability settings.
type ObservabilityConfig struct {
    // OpenTelemetry configuration for native Locust OTel integration.
    // +optional
    OpenTelemetry *OpenTelemetryConfig `json:"openTelemetry,omitempty"`
}

// OpenTelemetryConfig defines OpenTelemetry integration settings.
type OpenTelemetryConfig struct {
    // Enabled enables OpenTelemetry integration.
    // When true, adds --otel flag to Locust command.
    // +kubebuilder:default=false
    Enabled bool `json:"enabled"`

    // Endpoint is the OTel collector endpoint (e.g., "otel-collector:4317").
    // Required when Enabled is true.
    // +optional
    Endpoint string `json:"endpoint,omitempty"`

    // Protocol for OTel export.
    // +optional
    // +kubebuilder:validation:Enum=grpc;http/protobuf
    // +kubebuilder:default=grpc
    Protocol string `json:"protocol,omitempty"`

    // Insecure skips TLS verification for the collector connection.
    // +optional
    // +kubebuilder:default=false
    Insecure bool `json:"insecure,omitempty"`

    // ExtraEnvVars for additional OTel SDK configuration.
    // +optional
    ExtraEnvVars map[string]string `json:"extraEnvVars,omitempty"`
}
```

---

## Example Usage

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: load-test-with-otel
spec:
  image: locustio/locust:2.32.0  # Must have otel extras installed
  master:
    command: "locust -f /lotest/src/locustfile.py"
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 3
  testFiles:
    configMapRef: locust-scripts
  
  # OpenTelemetry configuration
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"
      insecure: true  # For non-TLS collectors
      extraEnvVars:
        OTEL_RESOURCE_ATTRIBUTES: "service.name=locust-load-test,environment=staging"
        OTEL_TRACES_SAMPLER: "parentbased_traceidratio"
        OTEL_TRACES_SAMPLER_ARG: "0.1"
```

---

## Quick Start

```bash
# After implementation, verify with:
make generate
make manifests
make build
make test

# Test OTel configuration manually
kubectl apply -f config/samples/locust_v2_locusttest_with_otel.yaml
kubectl get pods -l performance-test-name=load-test-with-otel

# Verify --otel flag in command
kubectl get pod <master-pod> -o jsonpath='{.spec.containers[0].args}' | grep -- --otel

# Verify OTel env vars
kubectl exec -it <master-pod> -- env | grep OTEL

# Verify NO metrics exporter sidecar
kubectl get pod <master-pod> -o jsonpath='{.spec.containers[*].name}'
# Expected: only "load-test-with-otel-master" (no "locust-exporter")

# Test validation rejection (missing endpoint)
cat <<EOF | kubectl apply -f -
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: invalid-otel
spec:
  image: locustio/locust:2.32.0
  master:
    command: "locust"
  worker:
    command: "locust"
    replicas: 1
  observability:
    openTelemetry:
      enabled: true
      # endpoint missing - should fail validation
EOF
# Expected: Rejected with "endpoint is required when OpenTelemetry is enabled" error
```

---

## Image Requirements

When using OpenTelemetry, the Locust image must have OTel extras installed:

```dockerfile
# Custom Dockerfile
FROM locustio/locust:2.32.0
RUN pip install locust[otel]
```

Or use a pre-built image with OTel support. The operator documentation should clearly state this requirement.

---

## Related Documents

- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist
- [DESIGN.md](./DESIGN.md) - Technical design and code patterns

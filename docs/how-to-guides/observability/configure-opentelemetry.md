---
title: Configure OpenTelemetry integration
description: Enable native OpenTelemetry support for metrics and traces export from Locust tests
tags:
  - observability
  - opentelemetry
  - metrics
  - traces
  - monitoring
---

# Configure OpenTelemetry integration

Native OpenTelemetry support in v2 eliminates the need for a metrics exporter sidecar. Your Locust tests can export metrics and traces directly to an OTel Collector.

## Prerequisites

You need an OpenTelemetry Collector deployed in your cluster. The collector receives telemetry data from Locust and forwards it to your observability backend (Prometheus, Jaeger, Tempo, etc.).

## Step 1: Verify OTel Collector endpoint connectivity

Determine the correct endpoint for your OTel Collector:

| Scenario | Endpoint Format | Example |
|----------|-----------------|---------|
| Same namespace | `http://<service-name>:<port>` | `http://otel-collector:4317` |
| Different namespace | `http://<service-name>.<namespace>:<port>` | `http://otel-collector.monitoring:4317` |
| External collector | `https://<hostname>:<port>` | `https://otel.example.com:4317` |

**Test connectivity** from a debug pod:

```bash
kubectl run debug --image=busybox --rm -it -- nc -zv otel-collector.monitoring 4317
```

If the connection succeeds, you'll see: `otel-collector.monitoring (10.0.1.5:4317) open`

## Step 2: Configure in LocustTest CR

Add the `observability.openTelemetry` block to your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: otel-enabled-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  observability:
    openTelemetry:
      enabled: true                                    # Enable OTel integration
      endpoint: "http://otel-collector.monitoring:4317"  # OTel Collector endpoint
      protocol: "grpc"                                 # Use gRPC (or "http/protobuf")
      insecure: false                                  # Use TLS (set true for dev/testing)
      extraEnvVars:
        OTEL_SERVICE_NAME: "my-load-test"              # Service name in traces
        OTEL_RESOURCE_ATTRIBUTES: "environment=staging,team=platform"  # Resource attributes
```

**Configuration fields explained:**

- `enabled`: Set to `true` to activate OpenTelemetry integration
- `endpoint`: OTel Collector URL (scheme://hostname:port). Include the scheme (`http://` or `https://`) for compatibility across OTel SDK versions.
- `protocol`: Transport protocol
    - `grpc` (recommended, default): Use gRPC transport
    - `http/protobuf`: Use HTTP/protobuf transport
- `insecure`: TLS configuration
    - `false` (default): Use TLS for secure communication
    - `true`: Skip TLS verification (development/testing only)

    !!! note
        TLS behavior primarily depends on the endpoint scheme (`http://` vs `https://`). The `OTEL_EXPORTER_OTLP_INSECURE` environment variable is set by the operator but may not be recognized by all OTel SDK implementations (e.g., Python). Use `http://` endpoints for non-TLS connections.
- `extraEnvVars`: Additional OpenTelemetry environment variables
    - `OTEL_SERVICE_NAME`: Identifier for this test in traces
    - `OTEL_RESOURCE_ATTRIBUTES`: Metadata tags (key=value pairs, comma-separated)

## Step 3: Deploy and verify

Apply your LocustTest CR:

```bash
kubectl apply -f locusttest.yaml
```

**Check that OTel environment variables were injected:**

```bash
kubectl get pod -l performance-test-name=otel-enabled-test -o yaml | grep OTEL_
```

**Expected environment variables:**

| Variable | Value | Purpose |
|----------|-------|---------|
| `OTEL_TRACES_EXPORTER` | `otlp` | Enable OTLP trace export |
| `OTEL_METRICS_EXPORTER` | `otlp` | Enable OTLP metrics export |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Your endpoint | Collector address |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` or `http/protobuf` | Transport protocol |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` (if set) | Skip TLS verification |
| `OTEL_SERVICE_NAME` | Your service name | Service identifier |
| `OTEL_RESOURCE_ATTRIBUTES` | Your attributes | Resource metadata |

## Step 4: Query traces and metrics

Once your test is running, telemetry flows to your OTel Collector and downstream backends.

**Prometheus metrics** (if OTel Collector exports to Prometheus):

!!! note
    The exact metric names depend on your OTel Collector pipeline configuration and Locust's OTel instrumentation. The examples below assume the Collector exports to Prometheus with default naming.

```promql
# Request rate by service
rate(locust_requests_total{service_name="my-load-test"}[1m])

# Average response time
avg(locust_response_time_seconds{service_name="my-load-test"})
```

**Jaeger/Tempo traces** (if OTel Collector exports to tracing backend):

Filter by:
- Service name: `my-load-test`
- Resource attributes: `environment=staging`, `team=platform`

Look for:
- Request spans showing HTTP calls
- Duration metrics for performance analysis
- Error traces for debugging failures

## Troubleshooting

### No traces appearing in backend

**Check Locust logs for OTel errors:**

```bash
kubectl logs job/otel-enabled-test-master | grep -i otel
```

**Common issues:**

| Problem | Symptom | Solution |
|---------|---------|----------|
| Wrong endpoint | Connection refused | Verify endpoint with `nc -zv` test |
| TLS mismatch | TLS handshake errors | Set `insecure: true` for testing, or fix TLS certificates |
| Collector not receiving OTLP | No error in logs but no data | Check collector logs and verify protocol matches |
| Network policy blocking | Connection timeouts | Ensure NetworkPolicy allows egress to collector |

**Check OTel Collector logs:**

```bash
kubectl logs -n monitoring deployment/otel-collector | grep -i error
```

### Performance impact

OpenTelemetry adds overhead to your test execution:

- **Overhead:** Generally minimal overhead, varying with telemetry volume, sampling rate, and collector proximity.
- **Network overhead:** Depends on telemetry volume and sampling

**Recommendations:**

- **Use sampling** for high-volume tests:
  ```yaml
  extraEnvVars:
    OTEL_TRACES_SAMPLER: "traceidratio"
    OTEL_TRACES_SAMPLER_ARG: "0.1"  # Sample 10% of traces
  ```
- **Adjust collector resources** if experiencing backpressure
- **Monitor test pods** for resource saturation when OTel is enabled

## OTel vs Metrics Sidecar comparison

| Aspect | OpenTelemetry | Metrics Sidecar |
|--------|---------------|-----------------|
| **Traces** | Yes | No |
| **Metrics** | Yes | Yes |
| **Additional containers** | None | 1 sidecar per master pod |
| **Setup complexity** | Requires OTel Collector | Works with Prometheus directly |
| **Resource overhead** | Generally minimal, varies with config | Additional sidecar container |
| **Recommended for** | New deployments, distributed tracing needs | Legacy compatibility, Prometheus-only stacks |
| **v2 API** | Yes | Yes (default when OTel disabled) |
| **v1 API** | No | Yes |

**When OpenTelemetry is enabled:**

- The `--otel` flag is automatically added to Locust commands
- The metrics exporter sidecar is NOT deployed
- The metrics port (9646) is excluded from the Service

**When to use each:**

- **Use OpenTelemetry** if you need traces, want to reduce container count, or are building a new observability stack
- **Use Metrics Sidecar** if you only need Prometheus metrics, have existing Prometheus infrastructure, or need v1 API compatibility

## Related guides

- [Monitor test status and health](monitor-test-status.md) — Check test phase, conditions, and pod health
- [Configure resources](../configuration/configure-resources.md) — Adjust resource limits for OTel overhead
- [Metrics & Dashboards](../../metrics_and_dashboards.md) — Complete observability reference

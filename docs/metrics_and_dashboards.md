---
title: Metrics & Dashboards
description: Information on Metrics & Dashboards.
tags:
  - monitoring
  - metrics
  - dashboards
  - prometheus
  - observability
---

# Metrics & Dashboards


## :material-chart-timeline: OpenTelemetry Metrics & Traces

!!! info "New in v2.0"
    Native OpenTelemetry support is available in the v2 API.

### Native OpenTelemetry Support

Locust 2.x includes native OpenTelemetry support, which the operator can configure automatically. This provides both metrics and distributed tracing without requiring the metrics exporter sidecar.

### Configuring OTel

Enable OpenTelemetry in your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: otel-test
spec:
  image: locustio/locust:2.43.3
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  observability:
    openTelemetry:
      enabled: true
      endpoint: "http://otel-collector.monitoring:4317"
      protocol: "grpc"
```

See [Advanced Topics - OpenTelemetry](how-to-guides/observability/configure-opentelemetry.md) for detailed configuration options.

### OTel Collector Setup

For a complete observability setup, deploy an OTel Collector. Example configuration:

```yaml
# otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
  otlphttp:
    endpoint: http://jaeger-collector:4318
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [otlp]
      exporters: [prometheus]
    traces:
      receivers: [otlp]
      exporters: [otlphttp]
```

!!! tip
    The Helm chart includes an optional OTel Collector deployment. Enable it with `otelCollector.enabled: true`.

---

## :material-robot-outline: Operator Metrics

The Go operator can expose controller-runtime metrics (disabled by default). When enabled, metrics are served on the configured port (default: 8080):

| Metric | Description |
|--------|-------------|
| `controller_runtime_reconcile_total` | Total reconciliations |
| `controller_runtime_reconcile_errors_total` | Reconciliation errors |
| `controller_runtime_reconcile_time_seconds` | Reconciliation duration |
| `workqueue_depth` | Current queue depth |
| `workqueue_adds_total` | Items added to queue |

These metrics can be scraped by Prometheus using the standard `/metrics` endpoint on the operator pod.

### Enabling Operator Metrics

Enable metrics in your Helm values:

```yaml
metrics:
  enabled: true
```

Then configure Prometheus to scrape the operator:

```yaml
scrape_configs:
  - job_name: 'locust-operator'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        action: keep
        regex: locust-k8s-operator
      - source_labels: [__address__]
        action: replace
        regex: ([^:]+)(?::\d+)?
        replacement: $1:8080
        target_label: __address__
```

### ServiceMonitor (Prometheus Operator)

If using the Prometheus Operator, create a ServiceMonitor to automatically discover and scrape operator metrics:

=== "HTTP Metrics"

    ```yaml
    apiVersion: monitoring.coreos.com/v1
    kind: ServiceMonitor
    metadata:
      name: locust-operator-metrics
      namespace: locust-operator-system
      labels:
        app.kubernetes.io/name: locust-k8s-operator
    spec:
      selector:
        matchLabels:
          app.kubernetes.io/name: locust-k8s-operator
      endpoints:
        - port: metrics
          path: /metrics
          interval: 30s
    ```

    Ensure Helm values have:
    ```yaml
    metrics:
      enabled: true
      secure: false  # HTTP metrics
    ```

=== "HTTPS Metrics (Recommended)"

    ```yaml
    apiVersion: monitoring.coreos.com/v1
    kind: ServiceMonitor
    metadata:
      name: locust-operator-metrics
      namespace: locust-operator-system
      labels:
        app.kubernetes.io/name: locust-k8s-operator
    spec:
      selector:
        matchLabels:
          app.kubernetes.io/name: locust-k8s-operator
      endpoints:
        - port: metrics
          path: /metrics
          interval: 30s
          scheme: https
          bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
          tlsConfig:
            insecureSkipVerify: true  # For development only
    ```

    Ensure Helm values have:
    ```yaml
    metrics:
      enabled: true
      secure: true  # Enable HTTPS metrics (default: false)
    ```

!!! warning "Production TLS"
    For production, use cert-manager to manage TLS certificates instead of `insecureSkipVerify: true`. See the operator's `config/prometheus/` directory for examples.

### Operator Metrics Queries

Useful PromQL queries for operator monitoring:

```promql
# Reconciliation rate (per second)
rate(controller_runtime_reconcile_total[5m])

# Reconciliation error rate
rate(controller_runtime_reconcile_errors_total[5m])

# Average reconciliation duration
rate(controller_runtime_reconcile_time_seconds_sum[5m]) 
  / rate(controller_runtime_reconcile_time_seconds_count[5m])

# Current workqueue depth
workqueue_depth

# Queue processing rate
rate(workqueue_adds_total[5m])
```

---

## :material-chart-box-outline: Locust Test Metrics

!!! warning "Two Metrics Approaches - Choose One"
    The operator provides **two mutually exclusive** methods for collecting Locust test metrics:
    
    **1. Prometheus Exporter Sidecar** (default, v1 & v2 API)
    
    - Uses `containersol/locust_exporter` sidecar on port 9646
    - Exposes Prometheus-formatted metrics
    - Works with Prometheus scraping
    - Documented in this section below
    
    **2. Native OpenTelemetry** (v2 API only)
    
    - Locust exports directly via OTLP protocol
    - No sidecar container needed
    - Metrics sent to OTel Collector
    - See [OpenTelemetry section above](#opentelemetry-metrics-traces)
    
    **When OTel is enabled, the exporter sidecar is NOT deployed.** All Prometheus exporter documentation below only applies to **non-OTel mode**.

### Metrics Exporter Sidecar (Non-OTel Mode)

When OpenTelemetry is **not** enabled, the operator automatically injects a Prometheus metrics exporter sidecar into the Locust master pod. This exporter scrapes Locust's built-in stats endpoint and exposes metrics in Prometheus format.

**What the Operator Creates Automatically**:

1. **Metrics Exporter Sidecar Container**:
   - Image: `containersol/locust_exporter:v0.5.0`
   - Port: 9646
   - Path: `/metrics`

2. **Kubernetes Service**: `<test-name>-master`
   - Includes metrics port 9646
   - Provides stable DNS endpoint

3. **Pod Annotations** (for Prometheus auto-discovery):
   ```yaml
   prometheus.io/scrape: "true"
   prometheus.io/path: "/metrics"
   prometheus.io/port: "9646"
   ```

**No manual setup required** - the operator handles everything.

### Available Locust Metrics

The exporter provides these key metrics from Locust:

| Metric | Type | Description |
|--------|------|-------------|
| `locust_requests_total` | Counter | Total number of requests |
| `locust_requests_current_rps` | Gauge | Current requests per second |
| `locust_requests_current_fail_per_sec` | Gauge | Current failures per second |
| `locust_requests_avg_response_time` | Gauge | Average response time (ms) |
| `locust_requests_min_response_time` | Gauge | Minimum response time (ms) |
| `locust_requests_max_response_time` | Gauge | Maximum response time (ms) |
| `locust_requests_avg_content_length` | Gauge | Average response size (bytes) |
| `locust_users` | Gauge | Current number of simulated users |
| `locust_errors` | Counter | Total errors by type |

For the complete list, see the [locust_exporter documentation](https://github.com/ContainerSolutions/locust_exporter).

### Locust Metrics Queries

Useful PromQL queries for load test monitoring:

```promql
# Total request rate across all endpoints
sum(rate(locust_requests_total[1m]))

# Error rate
sum(rate(locust_errors[1m]))

# Average response time
avg(locust_requests_avg_response_time)

# Max response time (locust_exporter exposes gauges, not histograms)
max(locust_requests_max_response_time)

# Current active users
sum(locust_users)

# Request rate by endpoint
sum(rate(locust_requests_total[1m])) by (name, method)

# Error percentage
100 * sum(rate(locust_errors[1m])) 
  / sum(rate(locust_requests_total[1m]))
```

### Integration Examples

The metrics are automatically exposed by the operator-created Service and pod annotations. Simply configure your monitoring tools to discover them:

**Prometheus**: Configure Kubernetes service discovery to scrape pods with `prometheus.io/scrape: "true"` annotation. The operator adds these annotations automatically - no manual configuration of individual tests needed.

**Grafana**: Connect to your Prometheus datasource and create dashboards using the PromQL queries above. Import panels from existing [Locust dashboard examples](https://grafana.com/grafana/dashboards/?search=locust).

**NewRelic**: Deploy a Prometheus agent configured to scrape Kubernetes pods with `prometheus.io/scrape: true` and forward metrics to NewRelic. See [Issue #118](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/118) for production deployment patterns.

**DataDog**: Configure the DataDog agent's Prometheus integration to auto-discover and scrape pods with `prometheus.io/*` annotations. The DataDog agent automatically finds operator-created test pods.

!!! tip "Production Deployment"
    For large-scale deployments, see [Issue #118](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/118) which documents production patterns used with thousands of tests in NewRelic and DataDog environments.

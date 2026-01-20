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
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"
```

See [Advanced Topics - OpenTelemetry](advanced_topics.md#opentelemetry-integration) for detailed configuration options.

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
  jaeger:
    endpoint: jaeger-collector:14250
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [otlp]
      exporters: [prometheus]
    traces:
      receivers: [otlp]
      exporters: [jaeger]
```

!!! tip
    The Helm chart includes an optional OTel Collector deployment. Enable it with `otelCollector.enabled: true`.

---

## :material-robot-outline: Operator Metrics

The Go operator exposes controller-runtime metrics on port 8080:

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


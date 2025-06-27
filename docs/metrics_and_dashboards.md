---
title: Metrics & Dashboards
description: Information on Metrics & Dashboards.
---

# Metrics & Dashboards

The Locust Kubernetes Operator is designed with observability in mind, providing out-of-the-box support for Prometheus metrics. This allows you to gain deep insights into your performance tests and the operator's behavior.

## :material-export: Prometheus Metrics Exporter

By default, the operator deploys a [Prometheus metrics exporter](https://github.com/ContainerSolutions/locust_exporter) alongside each Locust master and worker pod. This exporter collects detailed metrics from the Locust instances and exposes them in a format that Prometheus can scrape.

### :material-key-variant: Key Metrics

Some of the key metrics you can monitor include:

-   `locust_requests_total`: The total number of requests made.
-   `locust_requests_failed_total`: The total number of failed requests.
-   `locust_response_time_seconds`: The response time of requests.
-   `locust_users`: The number of simulated users.

### :material-tune: Configuration

To enable Prometheus to scrape these metrics, you'll need to configure a scrape job in your `prometheus.yml` file. Here's an example configuration:

```yaml
scrape_configs:
  - job_name: 'locust'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
```

## :material-view-dashboard-outline: Grafana Dashboards

Once you have your metrics flowing into Prometheus, you can create powerful and informative dashboards in Grafana to visualize your test results. You can build panels to track key performance indicators (KPIs) such as response times, request rates, and error rates.

There are also community-built Grafana dashboards available for Locust that you can adapt for your needs.

## :material-robot-outline: Operator Metrics

In addition to the Locust-specific metrics, the operator itself exposes a set of metrics through Micronaut's metrics module. These metrics provide insights into the operator's health and performance, including JVM metrics, uptime, and more. You can find these metrics by scraping the operator's pod on the `/health` endpoint.


---
title: How-To Guides
description: Task-oriented recipes for specific goals
tags:
  - how-to
  - guides
  - recipes
---

# How-To Guides

Task-oriented recipes for specific goals. Each guide walks you through a complete solution from start to finish.

## Configuration

Set up and configure your load tests:

- **[Configure resource limits and requests](configuration/configure-resources.md)** — Control CPU and memory allocation for master and worker pods
- **[Use a private image registry](configuration/use-private-registry.md)** — Pull Locust images from private registries with authentication
- **[Mount volumes to test pods](configuration/mount-volumes.md)** — Attach data, certificates, or configuration files from various sources
- **[Configure Kafka and AWS MSK integration](configuration/configure-kafka.md)** — Set up authenticated Kafka access for event-driven testing
- **[Configure automatic cleanup with TTL](configuration/configure-ttl.md)** — Automatically remove finished jobs and pods after tests complete

## Observability

Monitor and measure test performance:

- **[Configure OpenTelemetry integration](../advanced_topics.md#opentelemetry-integration)** — Export traces and metrics using native OTel support
- **[Set up Prometheus monitoring](../metrics_and_dashboards.md)** — Collect and visualize test metrics with Prometheus and Grafana

## Scaling

Scale tests for high load and optimize resource placement:

- **[Scale worker replicas for high load](scaling/scale-workers.md)** — Size worker replicas based on simulated user count
- **[Use node affinity for dedicated test nodes](scaling/use-node-affinity.md)** — Target specific nodes using labels and affinity rules
- **[Configure tolerations for tainted nodes](scaling/configure-tolerations.md)** — Schedule pods on nodes with taints
- **[Use node selector for simple node targeting](scaling/use-node-selector.md)** — Target nodes using simple label matching

## Security

Secure your tests and manage credentials:

- **[Inject secrets into test pods](../advanced_topics.md#environment-secret-injection)** — Use Kubernetes secrets for API keys, tokens, and credentials
- **[Use image pull secrets](configuration/use-private-registry.md)** — Authenticate with private container registries

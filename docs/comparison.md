---
title: Comparison - Locust on Kubernetes
description: Compare the Locust Kubernetes Operator with alternatives for running Locust load tests on Kubernetes. Feature comparison, decision guide, and migration paths.
---

# Comparison: Locust on Kubernetes

When running Locust load tests on Kubernetes, you have three main approaches to choose from:

1. **Locust Kubernetes Operator** (this project) - Full lifecycle management via Custom Resource Definition (CRD)
2. **Official Locust Helm Chart** (deliveryhero/helm-charts or locustio/locust) - Helm-based deployment
3. **Manual Deployment** - Raw Kubernetes manifests (Deployments, Services, ConfigMaps)

This page helps you evaluate which approach fits your use case, with an objective feature comparison, decision guide, and migration paths.

## Feature Comparison

| Feature | Locust K8s Operator | Official Helm Chart | Manual Deployment |
|---------|:-------------------:|:-------------------:|:-----------------:|
| Declarative CRD API | ✓ | ✗ | ✗ |
| Automated lifecycle (create/cleanup) | ✓ | Partial (Helm) | ✗ |
| Immutable test guarantee | ✓ | ✗ | ✗ |
| Validation webhooks | ✓ | ✗ | ✗ |
| CI/CD integration (autoQuit) | ✓ | Manual config | Manual config |
| OpenTelemetry native | ✓ | ✗ | ✗ |
| Secret injection (envFrom) | ✓ | Manual config | Manual config |
| Volume mounting | ✓ | ✓ | ✓ |
| Horizontal worker scaling | ✓ (workerReplicas) | ✓ (values) | ✓ (replicas) |
| Resource governance | ✓ (operator defaults + CR override) | ✓ (values) | ✓ (resource specs) |
| Status monitoring | ✓ (conditions, phases) | ✗ | ✗ |
| Pod health detection | ✓ | ✗ | ✗ |
| Leader election (HA) | ✓ | N/A | N/A |
| Helm chart provided | ✓ | ✓ | ✗ |
| Custom test scripts | ConfigMap ref | ConfigMap mount | ConfigMap mount |
| Multi-test isolation | ✓ (per-CR namespace) | Manual | Manual |
| Setup complexity | Low (Helm install) | Low (Helm install) | High (many manifests) |

## Decision Guide

!!! tip "Choose the Locust K8s Operator when..."

    - Running Locust tests in CI/CD pipelines regularly
    - Need automated test lifecycle management (create, run, cleanup)
    - Want immutability guarantees (no mid-test changes)
    - Require OpenTelemetry observability
    - Multiple teams sharing a cluster need governance and isolation
    - Need pod health monitoring and status conditions
    - Want validation webhooks to catch configuration errors before deployment

!!! info "Choose the Official Helm Chart when..."

    - Running occasional ad-hoc load tests
    - Already managing everything through Helm values
    - Don't need CRD-based lifecycle management
    - Simpler setup is more important than automation features
    - Want a minimal footprint without custom controllers

!!! info "Choose Manual Deployment when..."

    - Learning Locust on Kubernetes for the first time
    - Need maximum flexibility and customization
    - Running a one-off test in a development environment
    - Want to understand the underlying Kubernetes primitives
    - Have specific requirements not covered by existing solutions

## Migration Paths

### From Manual Deployment to Operator

If you're currently using raw Kubernetes manifests (Deployments, Services, ConfigMaps), migrating to the operator is straightforward:

1. Keep your existing ConfigMap with test scripts
2. Create a LocustTest CR that references your ConfigMap
3. Deploy the CR - the operator handles the rest

**Example:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: my-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-existing-configmap  # Reference your existing ConfigMap
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
```

[:octicons-arrow-right-24: Get started with the operator](getting_started.md)

### From Helm Chart to Operator

If you're using the official Locust Helm chart, you can map your Helm values to LocustTest CR fields:

- Helm `image` → CR `spec.image`
- Helm `master.args` → CR `spec.master.command`
- Helm `worker.replicas` → CR `spec.worker.replicas`
- Helm `locust_locustfile_configmap` → CR `spec.testFiles.configMapRef`
- Helm `locust_lib_configmap` → CR `spec.testFiles.libConfigMapRef`

The operator provides additional capabilities like automated cleanup, validation webhooks, and OpenTelemetry integration that aren't available in the Helm chart.

[:octicons-arrow-right-24: See detailed field mapping in the migration guide](migration.md#field-mapping-reference)

### From v1 Operator to v2 Operator

If you're already using v1 of the Locust Kubernetes Operator, migration to v2 is seamless:

- **Backward compatibility**: v1 CRs continue to work via automatic conversion webhook
- **New features**: Access OpenTelemetry, secret injection, volume mounting, separate resource specs
- **Performance**: 75% smaller memory footprint, sub-second startup time

[:octicons-arrow-right-24: Complete v1-to-v2 migration guide](migration.md)

## Ready to Get Started?

The Locust Kubernetes Operator provides the most comprehensive solution for running Locust tests on Kubernetes, especially for teams running tests regularly in CI/CD pipelines or production environments.

[:octicons-arrow-right-24: Get started in 5 minutes](getting_started.md){ .md-button .md-button--primary }

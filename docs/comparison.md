---
title: Comparison - Locust on Kubernetes
description: Compare the Locust Kubernetes Operator with alternatives for running Locust load tests on Kubernetes. Includes official Locust operator, k6 operator, and manual deployment. Feature comparison, performance benchmarks, decision guide, and migration paths.
---

# Comparison: Locust on Kubernetes

When running Locust load tests on Kubernetes, you have four main approaches to choose from:

1. **Locust Kubernetes Operator** (this project) - Full lifecycle management via Custom Resource Definition (CRD)
2. **Official Locust Operator** (locustio/k8s-operator) - Locust team operator
3. **k6 Operator** (Grafana) - Distributed k6 testing on Kubernetes
4. **Manual Deployment** - Raw Kubernetes manifests (Deployments, Services, ConfigMaps)

This page helps you evaluate which approach fits your use case, with an objective feature comparison, performance benchmarks, decision guide, and migration paths.

## Feature Comparison

| Feature | This Operator | Official Operator | k6 Operator | Manual Deploy |
|---------|:-------------:|:-----------------:|:-----------:|:-------------:|
| Declarative CRD API | ✓ | ✓ | ✓ | ✗ |
| Automated lifecycle | ✓ | ✓ | ✓ | ✗ |
| Immutable test guarantee | ✓ | ✗ | ✗ | ✗ |
| Validation webhooks | ✓ | Not documented | ✓ | ✗ |
| CI/CD integration (autoQuit) | ✓ | Not documented | ✓ (cloud) | Manual |
| OpenTelemetry native | ✓ | ✗ | ✗ | ✗ |
| Secret injection (envFrom) | ✓ | Not documented | ✓ | Manual |
| Volume mounting | ✓ | ✓ (ConfigMap) | ✓ | ✓ |
| Horizontal worker scaling | ✓ | ✓ | ✓ | ✓ |
| Resource governance | ✓ (operator + CR) | Not documented | ✓ | ✓ |
| Status monitoring (conditions) | ✓ | Not documented | ✓ | ✗ |
| Pod health detection | ✓ | ✗ | ✗ | ✗ |
| Leader election (HA) | ✓ | Not documented | ✓ | N/A |
| Helm chart | ✓ | ✓ | ✓ | ✗ |
| API versions supported | v1 + v2 (conversion) | Single version | Single version | N/A |
| Documentation pages | 20+ | 1 | Extensive | N/A |

**Note:** "Not documented" indicates features that may exist but are not described in the official documentation. The Official Locust Operator is maintained by the Locust core team.

## Why Choose This Operator

!!! success "Battle-Tested Reliability"

    - **Battle-Tested** - Production use on AWS EKS since 2022
    - **Comprehensive documentation** - Comprehensive coverage of getting started, API reference, architecture, security, FAQ, and migration guides
    - **Go-native performance** - Sub-second startup time, 75 MB container image, 64 MB memory footprint
    - **Feature-rich capabilities** - OpenTelemetry integration, validation webhooks, pod health monitoring, immutable test guarantee
    - **Active development** - Continuous improvement with community feedback and contributions

## Performance Benchmarks

All metrics measured from production deployment on AWS EKS. Container images measured via `docker images`, memory usage via Kubernetes metrics-server, startup time as duration to pod Ready state.

### Container Image Size

| Metric | This Operator (Go) | Previous Version (Java) |
|--------|:------------------:|:-----------------------:|
| Image size | 75 MB | 325 MB |
| Reduction | 77% smaller | — |

**Source:** `docker images` output

### Runtime Performance

| Metric | This Operator (Go) | Previous Version (Java) |
|--------|:------------------:|:-----------------------:|
| Memory usage (idle) | 64 MB | 256 MB |
| Startup time | < 1 second | ~60 seconds |

**Source:** Kubernetes metrics-server and pod Ready state timing

**Methodology:** Measurements from production deployment on AWS EKS. Container images measured via `docker images`, memory via Kubernetes metrics-server, startup time as duration to pod Ready state.

**Note:** Performance data for the Official Locust Operator and k6 Operator is not published by their maintainers.

## Decision Guide

!!! tip "Choose This Operator when..."

    - Running Locust tests in CI/CD pipelines regularly
    - Need automated test lifecycle management (create, run, cleanup)
    - Want immutability guarantees (no mid-test changes)
    - Require OpenTelemetry observability
    - Multiple teams sharing a cluster need governance and isolation
    - Need pod health monitoring and status conditions
    - Want validation webhooks to catch configuration errors before deployment

!!! info "Choose k6 Operator when..."

    - Using k6 (not Locust) for load testing
    - Need Grafana Cloud integration for observability
    - Prefer k6 scripting language and ecosystem
    - Part of the Grafana observability stack

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

[:octicons-arrow-right-24: Get started with the operator](getting_started/index.md)

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

The Locust Kubernetes Operator provides comprehensive lifecycle management for running Locust tests on Kubernetes, with features designed for CI/CD pipelines and production environments.

[:octicons-arrow-right-24: Get started in 5 minutes](getting_started/index.md){ .md-button .md-button--primary }

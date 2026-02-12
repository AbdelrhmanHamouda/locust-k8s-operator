---
title: Configure automatic cleanup with TTL
description: Automatically remove finished jobs and pods after tests complete
tags:
  - configuration
  - cleanup
  - ttl
  - automation
---

# Configure automatic cleanup with TTL

Automatically clean up finished master and worker jobs and their pods after tests complete using Kubernetes TTL (time-to-live).

## Prerequisites

- Kubernetes 1.12+ (TTL controller support)
- Locust Kubernetes Operator installed

## What gets cleaned up

When TTL is configured:

- ✓ **Cleaned up:** Master and worker Jobs and their Pods
- ✗ **Kept:** LocustTest CR, ConfigMaps, Secrets, Services

This allows you to review test results (via `kubectl get locusttest`) while automatically removing resource-consuming pods.

## Set TTL via Helm values

Configure TTL for all tests during operator installation:

```yaml
# values.yaml
locustPods:
  ttlSecondsAfterFinished: 3600  # Clean up after 1 hour
```

Install or upgrade the operator:

```bash
helm upgrade --install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  -f values.yaml
```

All LocustTest resources will inherit this TTL value.

## Set TTL via CLI

Override TTL at installation time:

```bash
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  --set locustPods.ttlSecondsAfterFinished=7200  # 2 hours
```

## Common TTL values

| Value | Duration | Use case |
|-------|----------|----------|
| `0` | Immediate | CI/CD where you collect results before cleanup |
| `300` | 5 minutes | Quick tests where results are exported immediately |
| `3600` | 1 hour | Standard tests with manual result review |
| `7200` | 2 hours | Long tests with delayed result analysis |
| `86400` | 24 hours | Tests requiring extensive post-analysis |
| `""` (empty) | Never | Development or when using external cleanup |

## Disable TTL

To disable automatic cleanup:

```yaml
# values.yaml
locustPods:
  ttlSecondsAfterFinished: ""  # Empty string disables TTL
```

Or omit the field entirely:

```yaml
# values.yaml
locustPods:
  # ttlSecondsAfterFinished not set - no TTL
```

Without TTL, jobs and pods persist until manually deleted.

## Verify TTL configuration

Check that TTL is set on created jobs:

```bash
# Run a test
kubectl apply -f locusttest.yaml

# Check master job TTL
kubectl get job -l performance-test-pod-name=my-test-master -o yaml | grep ttlSecondsAfterFinished
```

Expected output:

```yaml
ttlSecondsAfterFinished: 3600
```

## Watch automatic cleanup

Monitor cleanup in action:

```bash
# List all jobs with timestamps
kubectl get jobs -o wide --watch

# After TTL expires, jobs transition to deleted
```

Verify cleanup occurred:

```bash
# Jobs should be gone after TTL
kubectl get jobs -l performance-test-name=my-test

# Pods should also be gone
kubectl get pods -l performance-test-name=my-test

# But CR still exists
kubectl get locusttest my-test
```

## Example: CI/CD with immediate cleanup

For CI/CD pipelines where you collect results before cleanup:

```yaml
# values.yaml
locustPods:
  ttlSecondsAfterFinished: 0  # Clean up immediately after completion
```

In your pipeline:

```bash
# Run the test
kubectl apply -f locusttest.yaml

# Wait for completion
kubectl wait --for=jsonpath='{.status.phase}'=Succeeded \
  locusttest/ci-test --timeout=10m

# Collect results BEFORE cleanup happens
kubectl logs job/ci-test-master > results.log

# Jobs and pods will be cleaned up within seconds
# CR persists for historical tracking
```

## Example: Development with manual cleanup

During development, disable TTL to inspect pods:

```yaml
# values.yaml
locustPods:
  ttlSecondsAfterFinished: ""  # Disable automatic cleanup
```

Clean up manually when done:

```bash
# Delete just the test
kubectl delete locusttest my-test

# Or delete all test resources
kubectl delete locusttest --all
```

## Backward compatibility

The operator maintains backward compatibility with the old configuration path:

```yaml
# Old path (still supported)
config:
  loadGenerationJobs:
    ttlSecondsAfterFinished: 3600

# New path (recommended)
locustPods:
  ttlSecondsAfterFinished: 3600
```

Helper functions in the Helm chart ensure both paths work. Use the new path for future configurations.

## How TTL works

Kubernetes TTL controller monitors finished jobs:

1. Test completes (phase: Succeeded or Failed)
2. Job transitions to finished state
3. TTL countdown starts
4. After TTL seconds, controller deletes job
5. Cascading deletion removes dependent pods

**Important:** TTL countdown starts when the job finishes, not when it starts.

## What's next

- **[Scale worker replicas](../scaling/scale-workers.md)** — Size tests appropriately to minimize wasted resources
- **[Configure resources](configure-resources.md)** — Set resource limits to prevent cluster exhaustion
- **[Configure OpenTelemetry](../observability/configure-opentelemetry.md)** — Export metrics before cleanup

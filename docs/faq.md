---
title: Frequently Asked Questions
description: Common questions and answers about the Locust Kubernetes Operator
tags:
  - faq
  - troubleshooting
  - guide
---

# Frequently Asked Questions

This page answers the most common questions about operating the Locust Kubernetes Operator in production. For step-by-step tutorials, see [Getting Started](getting_started/index.md). For advanced configuration, see [Advanced Topics](advanced_topics.md).

## Test Lifecycle

### Why can't I update a running test?

Tests are **immutable by design**. Once a LocustTest CR is created, the operator ignores all changes to the `spec` field and sets a `SpecDrifted` condition to indicate drift was detected.

This ensures predictable behavior — each test run uses exactly the configuration it was created with, with no mid-flight configuration changes. See [How Does It Work - Immutable Tests](how_does_it_work.md#immutable-tests) for the design rationale.

To change test parameters, use the delete-and-recreate pattern:

```bash
kubectl delete locusttest my-test
# Edit your YAML with desired changes
kubectl apply -f locusttest.yaml
```

### How do I change test parameters?

Delete the LocustTest CR, edit your YAML file with the desired changes, and recreate it:

```bash
kubectl delete locusttest my-test
# Edit locusttest.yaml (change image, replicas, commands, etc.)
kubectl apply -f locusttest.yaml
```

The operator will create new Jobs with the updated configuration. Previous test results remain in your monitoring system (if using OpenTelemetry or metrics export).

### What happens if I edit a LocustTest CR after creation?

The operator detects spec changes but **ignores** them. It sets a `SpecDrifted` condition on the CR to indicate the spec has been modified:

```bash
kubectl get locusttest my-test -o jsonpath='{.status.conditions[?(@.type=="SpecDrifted")]}'
```

The test continues running with its original configuration. To apply changes, delete and recreate the CR.

### How do I run the same test multiple times?

Delete and recreate the CR with the same YAML:

```bash
kubectl delete locusttest my-test
kubectl apply -f locusttest.yaml  # Same file
```

Or use unique names with a suffix to keep test history:

```bash
kubectl apply -f locusttest-run-01.yaml
# Later...
kubectl apply -f locusttest-run-02.yaml
```

## Scaling

### Can I scale workers during a running test?

No, due to immutability. The worker replica count (`worker.replicas`) is set at test creation time and cannot be changed while the test runs.

To run with different worker counts:

```bash
kubectl delete locusttest my-test
# Edit YAML to update worker.replicas
kubectl apply -f locusttest.yaml
```

Note: Locust's web UI shows real-time user distribution across connected workers regardless of the replica count.

### What's the maximum number of workers?

The CRD enforces a maximum of **500 workers** per LocustTest. This limit prevents accidental resource exhaustion.

For larger scales:

- Run multiple LocustTest CRs against the same target (each test independently generates load)
- Use fewer workers with more users per worker (adjust `--users` and `--spawn-rate` in `master.command`)

### How do I size worker resources?

Resource requirements depend on test complexity:

| Test Type | CPU per Worker | Memory per Worker | Notes |
|-----------|---------------|-------------------|-------|
| Light HTTP tests | 250m | 128Mi | Simple GET/POST requests |
| Medium complexity | 500m | 256Mi | JSON parsing, simple logic |
| Heavy tests | 1000m | 512Mi-1Gi | Complex business logic, large payloads |

Start conservative and observe resource usage via `kubectl top pods`. See [Advanced Topics - Resource Management](how-to-guides/configuration/configure-resources.md) for detailed sizing guidance.

!!! tip "Resource Precedence"
    The operator applies resources in order of specificity: (1) CR spec resources (highest), (2) Helm role-specific resources (`masterResources`/`workerResources`), (3) Helm unified resources (`locustPods.resources`).

## Debugging

### My test is stuck in Pending phase

Check in this order:

1. **Check operator events**: `kubectl describe locusttest <test-name>` — look for errors in the Events section
2. **Check pod status**: `kubectl get pods -l performance-test-name=<test-name>` — look for scheduling errors or image pull failures
3. **Check PodsHealthy condition**: `kubectl get locusttest <test-name> -o jsonpath='{.status.conditions[?(@.type=="PodsHealthy")]}'` — the operator reports pod issues here
4. **Check ConfigMap exists**: If using `testFiles.configMapRef`, ensure the ConfigMap exists: `kubectl get configmap <name>`

The operator has a 2-minute grace period before reporting pod failures, allowing time for image pulls and startup.

### My test shows Failed phase

Check the failure reason:

1. **Check conditions**: `kubectl describe locusttest <test-name>` — the Status section shows why it failed
2. **Check master logs**: `kubectl logs <test-name>-master-<hash>` — Locust errors appear here
3. **Common causes**:
    - **Locustfile syntax error**: Python errors in your test script
    - **Target host unreachable**: Network connectivity issues
    - **ConfigMap not found**: Missing test files
    - **Image pull failure**: Invalid image name or missing pull secrets

### Workers show 0/N connected

The `connectedWorkers` field is an approximation from `Job.Status.Active`. Workers need time to start, pull images, and connect to the master.

Check worker connectivity:

1. **Verify worker pods are running**: `kubectl get pods -l locust-role=worker`
2. **Verify master service exists**: `kubectl get svc <test-name>-master`
3. **Check worker logs**: `kubectl logs <test-name>-worker-<hash>` — workers should show "Connected to master"
4. **Verify network connectivity**: Workers connect to the master on port 5557

Workers typically connect within 30-60 seconds after pod startup.

### How do I access the Locust web UI?

Port-forward to the master service:

```bash
kubectl port-forward svc/<test-name>-master 8089:8089
```

Then visit [http://localhost:8089](http://localhost:8089) in your browser.

!!! note "Autostart Behavior"
    If `autostart: true` (default), the test starts automatically and the web UI shows the running test. Set `autostart: false` to control test start from the web UI.

### ConfigMap not found error

The operator detects missing ConfigMaps via pod health monitoring and reports the issue in the `PodsHealthy` condition.

You can create the ConfigMap **before or after** the LocustTest CR:

```bash
# Create ConfigMap from local files
kubectl create configmap my-test-scripts --from-file=test.py=./test.py

# If LocustTest already exists, the operator detects recovery automatically
```

The operator monitors pod status every 30 seconds and updates conditions when ConfigMaps become available.

## Migration

### Can I use v1 and v2 CRs at the same time?

Yes, with the conversion webhook enabled. The operator automatically converts v1 CRs to v2 internally, allowing both versions to coexist.

v1 CRs continue to work with their existing configuration. See [Migration Guide](migration.md) for conversion details.

### Do I need to recreate existing v1 tests?

No, existing v1 tests continue to work. However, new features (OpenTelemetry integration, environment variable injection, volume mounts) require the v2 API.

Migrate when you need v2-only features or when convenient. See [Migration Guide](migration.md) for the conversion process.

## Configuration

### How do I pass extra CLI arguments to Locust?

Use `master.extraArgs` and `worker.extraArgs` in the v2 API. These are appended after the command seed and operator-managed flags:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
spec:
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
    extraArgs:
      - "--loglevel"
      - "DEBUG"
  worker:
    command: "--locustfile /lotest/src/test.py"
    extraArgs:
      - "--loglevel"
      - "DEBUG"
    replicas: 5
```

!!! warning "Reserved Flags"
    The operator manages these flags automatically: `--master`, `--worker`, `--master-host`, `--master-port`, `--expect-workers`, `--autostart`, `--autoquit`. Do not set them manually.

### What resource precedence applies?

The operator applies resources in this order (first non-empty value wins):

1. **CR spec resources** (highest precedence): Set in `LocustTest.spec.master.resources` or `LocustTest.spec.worker.resources`
2. **Helm role-specific resources**: Set in `values.yaml` as `locustPods.masterResources` or `locustPods.workerResources`
3. **Helm unified resources**: Set in `values.yaml` as `locustPods.resources`

This allows global defaults with role-specific overrides and per-test customization.

## Observability

### Should I use OpenTelemetry or the metrics sidecar?

**Use OpenTelemetry for new deployments.** It provides traces and metrics without requiring a sidecar container, reducing resource overhead.

The metrics sidecar is maintained for legacy compatibility. Use it only if:

- Your monitoring stack doesn't support OTLP
- You have existing dashboards built on the CSV metrics format

See [Advanced Topics - OpenTelemetry Integration](how-to-guides/observability/configure-opentelemetry.md) for configuration details.

### How do I monitor test progress programmatically?

Use the LocustTest status conditions for automation:

```bash
# Check if test is ready
kubectl get locusttest my-test -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'

# Get current phase
kubectl get locusttest my-test -o jsonpath='{.status.phase}'

# Get worker count
kubectl get locusttest my-test -o jsonpath='{.status.connectedWorkers}/{.status.expectedWorkers}'
```

See [API Reference - Status Fields](api_reference.md#status-fields) for all available status information.

---
title: Configure resource limits and requests
description: Control CPU and memory allocation for Locust pods
tags:
  - configuration
  - resources
  - performance
---

# Configure resource limits and requests

Resource configuration ensures your load tests have the resources they need without consuming excessive cluster capacity.

## Prerequisites

- Locust Kubernetes Operator installed
- Basic understanding of Kubernetes resource requests and limits

## Set global defaults via Helm

Configure default resources for all tests during operator installation:

```yaml
# values.yaml
locustPods:
  resources:
    requests:
      cpu: "250m"        # Guaranteed CPU
      memory: "256Mi"    # Guaranteed memory
    limits:
      cpu: "1000m"       # Maximum CPU
      memory: "512Mi"    # Maximum memory
```

Install or upgrade the operator:

```bash
helm upgrade --install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  -f values.yaml
```

These defaults apply to all Locust pods unless overridden in individual CRs.

## Configure per-test resources

Override defaults for specific tests using the v2 API. Master and worker pods can have different resource configurations:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: resource-optimized-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
    resources:
      requests:
        memory: "256Mi"    # Master needs less memory
        cpu: "100m"        # Master is not CPU-intensive
      limits:
        memory: "512Mi"
        cpu: "500m"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 10
    resources:
      requests:
        memory: "512Mi"    # Workers need more memory for load generation
        cpu: "500m"        # Workers are CPU-intensive
      limits:
        memory: "1Gi"
        cpu: "1000m"
```

Apply the configuration:

```bash
kubectl apply -f locusttest-resources.yaml
```

## Disable CPU limits for performance tests

CPU limits can cause throttling in performance-sensitive tests. Disable them by omitting the CPU limit field:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: no-cpu-limit-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        # No CPU limit - allows maximum performance
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 10
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        # No CPU limit - workers can use all available CPU
```

**When to disable CPU limits:**

- High-throughput performance tests (>5000 RPS)
- Benchmarking scenarios where you need maximum performance
- Tests with bursty traffic patterns

**Risk:** Pods can consume all available CPU on the node, potentially affecting other workloads. Use with [node affinity](../scaling/use-node-affinity.md) to isolate tests on dedicated nodes.

## Resource sizing guidelines

**Master pod:**

- CPU: 100-500m (master coordinates, doesn't generate load)
- Memory: 256-512Mi (depends on test complexity and UI usage)
- Usually 1 replica

**Worker pod:**

- CPU: 500-1000m per worker (depends on test script complexity)
- Memory: 256-512Mi per worker (depends on data handling)
- Scale workers based on user count (see [Scale worker replicas](../scaling/scale-workers.md))

**Example sizing for 1000 users:**

```yaml
master:
  resources:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "512Mi"
      cpu: "500m"

worker:
  replicas: 20  # ~50 users per worker
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "1Gi"
      # CPU limit omitted for performance
```

## Verify resource configuration

Check actual resource specs on running pods:

```bash
# Get master pod name
MASTER_POD=$(kubectl get pod -l locust.io/role=master -o jsonpath='{.items[0].metadata.name}')

# Verify resource configuration
kubectl describe pod $MASTER_POD | grep -A 10 "Limits:\|Requests:"
```

Expected output:

```
Limits:
  memory:  512Mi
Requests:
  cpu:     100m
  memory:  256Mi
```

## Monitor resource usage

Check actual resource consumption:

```bash
# Real-time resource usage
kubectl top pod -l locust.io/test-id=resource-optimized-test

# Watch resource usage during test
kubectl top pod -l locust.io/test-id=resource-optimized-test --watch
```

If pods consistently hit memory limits, they'll be OOMKilled. If they hit CPU limits, they'll be throttled (slower performance).

## What's next

- **[Scale worker replicas](../scaling/scale-workers.md)** — Calculate worker count for high-load scenarios
- **[Use node affinity](../scaling/use-node-affinity.md)** — Run resource-intensive tests on dedicated nodes
- **[Configure tolerations](../scaling/configure-tolerations.md)** — Schedule tests on high-performance node pools

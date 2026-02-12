---
title: Scale worker replicas for high load
description: Size worker replicas based on simulated user count and throughput
tags:
  - scaling
  - workers
  - performance
---

# Scale worker replicas for high load

Calculate and configure worker replicas to handle your target user count and request throughput.

## Prerequisites

- Locust Kubernetes Operator installed
- Basic understanding of distributed load testing

## How worker replicas affect throughput

Each worker pod generates load independently. More workers = more throughput capacity.

**Key factors:**

- **User count:** Each worker can efficiently handle 50-100 simulated users (depends on test complexity)
- **Request rate:** CPU-intensive tests (complex parsing, encryption) need more workers
- **Memory usage:** Tests with large payloads or state need more memory per worker

## Calculate worker count

**Formula:**

```
workers = ceil(total_users / users_per_worker)
```

**Default rule of thumb:** 50 users per worker

**Examples:**

| Target Users | Users/Worker | Workers Needed |
|--------------|--------------|----------------|
| 100 | 50 | 2 |
| 500 | 50 | 10 |
| 1000 | 50 | 20 |
| 5000 | 50 | 100 |

**Adjust users per worker based on test complexity:**

- **Simple tests** (basic HTTP GET): 100 users/worker
- **Standard tests** (REST API with JSON): 50 users/worker
- **Complex tests** (heavy parsing, encryption, large payloads): 25 users/worker

## Configure worker replicas

Set worker count in your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: scaled-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: |
      --locustfile /lotest/src/test.py
      --host https://api.example.com
      --users 1000          # Total simulated users
      --spawn-rate 50       # Users to add per second
      --run-time 10m
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 20          # 1000 users / 50 users per worker = 20 workers
```

Apply the configuration:

```bash
kubectl apply -f locusttest-scaled.yaml
```

## Example: 1000 users test

Complete configuration for 1000 concurrent users:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: high-load-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: high-load-test-script
  master:
    command: |
      --locustfile /lotest/src/test.py
      --host https://api.example.com
      --users 1000
      --spawn-rate 50
      --run-time 15m
    resources:
      requests:
        memory: "256Mi"
        cpu: "200m"
      limits:
        memory: "512Mi"
        cpu: "500m"
  worker:
    replicas: 20  # 1000 users at 50 users/worker
    command: "--locustfile /lotest/src/test.py"
    resources:
      requests:
        memory: "512Mi"   # More memory for load generation
        cpu: "500m"
      limits:
        memory: "1Gi"
        # CPU limit omitted for maximum performance
```

## Resource implications

Each worker consumes cluster resources:

**Per-worker resource baseline:**

- CPU: 500m request, no limit (for performance)
- Memory: 512Mi-1Gi (depends on test data)

**Total resources for 20 workers:**

- CPU: 10 cores requested (20 workers × 500m)
- Memory: 10-20Gi (20 workers × 512Mi-1Gi)

**Planning checklist:**

- [ ] Cluster has enough capacity for all workers
- [ ] Consider using [node affinity](use-node-affinity.md) to target specific node pools
- [ ] Configure [resource limits](../configuration/configure-resources.md) appropriately
- [ ] Use [tolerations](configure-tolerations.md) if running on dedicated nodes

## Monitor connected workers

Verify that all workers connect successfully:

```bash
# Watch test status
kubectl get locusttest high-load-test -w

# Check status field
kubectl get locusttest high-load-test -o jsonpath='{.status}'
```

Look for:

```json
{
  "phase": "Running",
  "expectedWorkers": 20,
  "connectedWorkers": 20
}
```

**If `connectedWorkers` < `expectedWorkers`:**

```bash
# List worker pods
kubectl get pods -l locust.io/role=worker

# Check for pending or failed pods
kubectl get pods -l locust.io/role=worker | grep -v Running

# Describe problematic pods
kubectl describe pod <worker-pod-name>
```

Common issues:

- Insufficient cluster capacity (pending pods)
- Image pull failures
- Resource quota exceeded
- Node selector or affinity constraints not satisfied

## View worker pod distribution

Check which nodes are running workers:

```bash
kubectl get pods -l locust.io/role=worker -o wide
```

Output shows pod-to-node distribution:

```
NAME                        NODE            STATUS
high-load-test-worker-0     node-pool-1-a   Running
high-load-test-worker-1     node-pool-1-b   Running
high-load-test-worker-2     node-pool-1-c   Running
...
```

**Best practice:** Distribute workers across multiple nodes for resilience and better resource utilization.

## Scaling considerations

**Spawn rate:**

Match spawn rate to worker count and network capacity:

```
recommended_spawn_rate = workers × 5-10 users/second
```

For 20 workers: 100-200 users/second spawn rate is reasonable.

**Example:**

```yaml
master:
  command: |
    --users 1000
    --spawn-rate 100   # 20 workers × 5 users/sec/worker
```

Too high spawn rate overwhelms workers during ramp-up. Too low takes too long to reach target.

**Network bandwidth:**

High worker counts can saturate network:

- 20 workers × 100 RPS = 2000 total RPS
- At 10KB per request = 20MB/s bandwidth

Ensure cluster networking can handle aggregate throughput.

**Master capacity:**

Master coordinates all workers. Very high worker counts (>50) may require increased master resources:

```yaml
master:
  resources:
    requests:
      memory: "512Mi"  # Increased from 256Mi
      cpu: "500m"      # Increased from 200m
    limits:
      memory: "1Gi"
      cpu: "1000m"
```

## Dynamic scaling (manual)

Scale workers during test execution:

```bash
# Increase workers
kubectl scale locusttest high-load-test --replicas=30

# Wait for new workers to connect
kubectl get locusttest high-load-test -o jsonpath='{.status.connectedWorkers}'
```

**Note:** Locust handles worker connections dynamically. New workers join running tests automatically.

## What's next

- **[Configure resources](../configuration/configure-resources.md)** — Set appropriate CPU and memory for workers
- **[Use node affinity](use-node-affinity.md)** — Target high-performance nodes for workers
- **[Configure tolerations](configure-tolerations.md)** — Run workers on dedicated node pools

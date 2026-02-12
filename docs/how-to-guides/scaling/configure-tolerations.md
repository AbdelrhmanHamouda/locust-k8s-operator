---
title: Configure tolerations for tainted nodes
description: Schedule pods on nodes with taints
tags:
  - scaling
  - scheduling
  - tolerations
  - taints
---

# Configure tolerations for tainted nodes

Schedule Locust pods on tainted nodes using tolerations, enabling dedicated node pools and preventing other workloads from using test infrastructure.

## Prerequisites

- Locust Kubernetes Operator installed
- Access to taint cluster nodes

## When to use tolerations

**Common use cases:**

- **Dedicated node pools:** Reserve nodes exclusively for load testing
- **High-performance nodes:** Prevent regular workloads from consuming resources
- **Spot instances:** Allow tests on spot/preemptible nodes with taints
- **Specialized hardware:** Schedule on GPU or high-memory nodes

**How taints and tolerations work:**

1. **Taint nodes:** Mark nodes as special-purpose (e.g., "dedicated=load-testing:NoSchedule")
2. **Add tolerations:** Pods with matching tolerations can be scheduled on tainted nodes
3. **Result:** Only pods with tolerations use the tainted nodes

## Taint your nodes

Add taints to nodes you want to dedicate for testing:

```bash
# Taint a node for load testing
kubectl taint nodes node-1 dedicated=load-testing:NoSchedule

# Taint multiple nodes
kubectl taint nodes node-2 dedicated=load-testing:NoSchedule
kubectl taint nodes node-3 dedicated=load-testing:NoSchedule

# Verify taints
kubectl describe node node-1 | grep Taints
```

**Taint effects:**

| Effect | Behavior |
|--------|----------|
| `NoSchedule` | New pods without toleration won't be scheduled |
| `PreferNoSchedule` | Scheduler tries to avoid placing pods here (soft) |
| `NoExecute` | Existing pods without toleration are evicted |

**Recommendation:** Use `NoSchedule` for dedicated test nodes.

## Configure tolerations

Add `scheduling.tolerations` to your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: toleration-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 10
  scheduling:
    tolerations:
      - key: dedicated
        operator: Equal
        value: load-testing
        effect: NoSchedule
```

Apply the configuration:

```bash
kubectl apply -f locusttest-toleration.yaml
```

## Toleration operators

**Equal operator:** Exact match required

```yaml
tolerations:
  - key: dedicated
    operator: Equal
    value: load-testing  # Must match exactly
    effect: NoSchedule
```

Matches taint: `dedicated=load-testing:NoSchedule`

**Exists operator:** Key must exist, value doesn't matter

```yaml
tolerations:
  - key: dedicated
    operator: Exists  # Any value for key "dedicated"
    effect: NoSchedule
```

Matches taints: `dedicated=load-testing:NoSchedule`, `dedicated=anything:NoSchedule`

## Multiple tolerations

Tolerate multiple taints:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: multi-toleration-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 10
  scheduling:
    tolerations:
      - key: dedicated
        operator: Equal
        value: load-testing
        effect: NoSchedule
      - key: spot-instance
        operator: Exists
        effect: NoSchedule
      - key: high-performance
        operator: Equal
        value: "true"
        effect: PreferNoSchedule
```

Pods can be scheduled on nodes with any of these taints.

## Example: Dedicated node pool

Complete setup for dedicated load testing nodes:

**1. Taint the nodes:**

```bash
kubectl taint nodes node-pool-load-1 workload=load-testing:NoSchedule
kubectl taint nodes node-pool-load-2 workload=load-testing:NoSchedule
kubectl taint nodes node-pool-load-3 workload=load-testing:NoSchedule
```

**2. Label the nodes (for affinity):**

```bash
kubectl label nodes node-pool-load-1 workload-type=load-testing
kubectl label nodes node-pool-load-2 workload-type=load-testing
kubectl label nodes node-pool-load-3 workload-type=load-testing
```

**3. Create LocustTest with affinity + tolerations:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: dedicated-pool-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 15
  scheduling:
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type
                  operator: In
                  values:
                    - load-testing  # Target labeled nodes
    tolerations:
      - key: workload
        operator: Equal
        value: load-testing
        effect: NoSchedule  # Tolerate the taint
```

**Result:** Pods only run on dedicated nodes, and only these pods can use those nodes.

## Example: Spot instances

Run tests on cost-optimized spot/preemptible instances:

**1. Taint spot nodes:**

```bash
# Cloud providers often add this taint automatically
kubectl taint nodes spot-node-1 cloud.google.com/gke-preemptible=true:NoSchedule
```

**2. Tolerate spot node taints:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: spot-instance-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 20
  scheduling:
    tolerations:
      - key: cloud.google.com/gke-preemptible  # GKE spot instances
        operator: Exists
        effect: NoSchedule
      - key: eks.amazonaws.com/capacityType    # AWS spot instances
        operator: Equal
        value: SPOT
        effect: NoSchedule
```

## NoExecute effect

`NoExecute` evicts running pods without toleration:

```bash
# Taint with NoExecute
kubectl taint nodes node-1 maintenance=scheduled:NoExecute
```

Pods without toleration are immediately evicted. Use for:

- Scheduled node maintenance
- Emergency capacity reclaim
- Node pool draining

**Toleration with grace period:**

```yaml
tolerations:
  - key: maintenance
    operator: Equal
    value: scheduled
    effect: NoExecute
    tolerationSeconds: 300  # Pod survives 5 minutes, then evicted
```

## Verify tolerations and node placement

Check that pods are scheduled on tainted nodes:

```bash
# Show pod placement
kubectl get pods -l performance-test-name=toleration-test -o wide

# Check node taints
kubectl describe node <node-name> | grep Taints

# Verify pod tolerations
kubectl get pod <pod-name> -o jsonpath='{.spec.tolerations}'
```

## Troubleshoot scheduling failures

If pods remain `Pending`:

```bash
kubectl describe pod <pod-name> | grep -A 10 "Events:"
```

**Common issues:**

**Missing toleration:**

```
Warning  FailedScheduling  0/3 nodes are available: 3 node(s) had taint {dedicated: load-testing}
```

Fix: Add matching toleration to the CR.

**Typo in taint key or value:**

```bash
# Check actual taint
kubectl describe node node-1 | grep Taints

# Output: Taints: dedicated=load-testing:NoSchedule
```

Ensure toleration matches exactly (case-sensitive).

**Wrong effect:**

Taint: `dedicated=load-testing:NoSchedule`
Toleration: `effect: PreferNoSchedule` (mismatch)

Fix: Match the effect in toleration to the taint.

## Remove taints

When no longer needed:

```bash
# Remove specific taint
kubectl taint nodes node-1 dedicated=load-testing:NoSchedule-

# Note the trailing minus (-) to remove
```

## What's next

- **[Use node affinity](use-node-affinity.md)** — Target specific nodes (often used together)
- **[Use node selector](use-node-selector.md)** — Simpler alternative without taints
- **[Scale worker replicas](scale-workers.md)** — Calculate capacity for dedicated pools

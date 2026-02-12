---
title: Use node selector for simple node targeting
description: Target nodes using simple label matching
tags:
  - scaling
  - scheduling
  - node selector
---

# Use node selector for simple node targeting

Target specific nodes using simple label matching with node selector, the easiest way to control pod placement.

## Prerequisites

- Locust Kubernetes Operator installed
- Access to label cluster nodes

## When to use node selector

**Use node selector when:**

- You need simple label matching (key=value)
- All conditions are AND (all labels must match)
- You want the simplest configuration

**Use node affinity when:**

- You need OR logic (match any of multiple labels)
- You need soft preferences (preferred but not required)
- You need complex expressions (In, NotIn, Exists, DoesNotExist)

See [Use node affinity](use-node-affinity.md) for advanced scenarios.

## Label your nodes

Add labels to nodes:

```bash
# Label for SSD storage
kubectl label nodes node-1 disktype=ssd

# Label for performance environment
kubectl label nodes node-1 environment=performance

# Label multiple nodes
kubectl label nodes node-2 disktype=ssd environment=performance
kubectl label nodes node-3 disktype=ssd environment=performance

# Verify labels
kubectl get nodes --show-labels | grep disktype
```

## Configure node selector

Add `scheduling.nodeSelector` to your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: nodeselector-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  scheduling:
    nodeSelector:
      disktype: ssd  # Only schedule on nodes with this label
```

Apply the configuration:

```bash
kubectl apply -f locusttest-nodeselector.yaml
```

## Multiple labels (AND logic)

Require multiple labels on nodes:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: multi-label-selector
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
    nodeSelector:
      disktype: ssd              # Must have SSD
      environment: performance   # AND must be performance environment
```

Nodes must have **both** labels to be selected.

## Example: High-performance nodes

Target high-performance node pool:

**1. Label your high-performance nodes:**

```bash
kubectl label nodes perf-node-1 performance-tier=high
kubectl label nodes perf-node-2 performance-tier=high
kubectl label nodes perf-node-3 performance-tier=high
```

**2. Configure test to use labeled nodes:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: high-perf-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: performance-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 20
  scheduling:
    nodeSelector:
      performance-tier: high  # Only high-performance nodes
```

## Example: AWS instance type targeting

Target specific EC2 instance types:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: aws-instance-test
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
    nodeSelector:
      node.kubernetes.io/instance-type: c5.2xlarge  # Compute-optimized
```

**Note:** This only matches one instance type. For multiple types, use [node affinity](use-node-affinity.md) with `In` operator.

## Example: Zone-specific deployment

Keep tests in a specific availability zone:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: zone-specific-test
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
    nodeSelector:
      topology.kubernetes.io/zone: us-east-1a  # Only zone 1a
```

## Verify node placement

Check that pods are scheduled on the correct nodes:

```bash
# Show pod-to-node mapping
kubectl get pods -l performance-test-name=nodeselector-test -o wide

# Check labels on nodes where pods are running
NODE=$(kubectl get pod -l performance-test-pod-name=nodeselector-test-master -o jsonpath='{.items[0].spec.nodeName}')
kubectl get node $NODE --show-labels | grep disktype
```

Expected: All pods running on nodes with matching labels.

## Troubleshoot scheduling failures

If pods remain `Pending`:

```bash
kubectl describe pod <pod-name> | grep -A 10 "Events:"
```

**Common issue:**

```
Warning  FailedScheduling  0/5 nodes are available: 5 node(s) didn't match Pod's node affinity/selector
```

**Causes:**

1. **No nodes with matching labels:**

   ```bash
   # Check if any nodes have the label
   kubectl get nodes -l disktype=ssd
   ```

   Fix: Label at least one node.

2. **Typo in label key or value:**

   ```bash
   # Check actual labels
   kubectl get nodes --show-labels | grep disktype
   ```

   Ensure spelling and case match exactly.

3. **Insufficient capacity on labeled nodes:**

   ```bash
   # Check node resources
   kubectl describe node -l disktype=ssd | grep -A 5 "Allocated resources"
   ```

   Fix: Add more labeled nodes or reduce resource requests.

## Compare with node affinity

**Node selector:**

```yaml
scheduling:
  nodeSelector:
    disktype: ssd
```

**Equivalent node affinity:**

```yaml
scheduling:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
              - key: disktype
                operator: In
                values:
                  - ssd
```

Node selector is simpler. Node affinity is more powerful.

## Combine with other scheduling

Node selector works with tolerations:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: selector-toleration-test
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
    nodeSelector:
      disktype: ssd  # Simple label matching
    tolerations:
      - key: dedicated
        operator: Equal
        value: load-testing
        effect: NoSchedule  # Tolerate taint on SSD nodes
```

See [Configure tolerations](configure-tolerations.md) for details.

## What's next

- **[Use node affinity](use-node-affinity.md)** — Complex scheduling with OR logic and preferences
- **[Configure tolerations](configure-tolerations.md)** — Schedule on tainted nodes
- **[Scale worker replicas](scale-workers.md)** — Calculate capacity for labeled nodes

---
title: Use node affinity for dedicated test nodes
description: Target specific nodes using labels and affinity rules
tags:
  - scaling
  - scheduling
  - node affinity
---

# Use node affinity for dedicated test nodes

Schedule Locust pods on specific nodes using node affinity, enabling dedicated test infrastructure or zone isolation.

## Prerequisites

- Locust Kubernetes Operator installed
- Access to label cluster nodes

## When to use node affinity

**Common use cases:**

- **Dedicated nodes:** Run load tests on nodes reserved for testing
- **High-performance nodes:** Target nodes with faster CPUs or more memory
- **Zone isolation:** Keep tests in specific availability zones
- **Cost optimization:** Use spot instances or lower-cost node pools for testing

**Node affinity vs node selector:**

- **Node selector:** Simple label matching (use this for basic needs)
- **Node affinity:** Complex rules with OR logic, soft preferences, multiple conditions

Use node affinity when you need the flexibility. Use [node selector](use-node-selector.md) for simplicity.

## Label your nodes

Add labels to nodes where you want to run tests:

```bash
# Label nodes for load testing
kubectl label nodes node-1 workload-type=load-testing
kubectl label nodes node-2 workload-type=load-testing
kubectl label nodes node-3 workload-type=load-testing

# Verify labels
kubectl get nodes --show-labels | grep workload-type
```

**Example labels:**

```bash
# By workload type
kubectl label nodes node-1 workload-type=load-testing

# By performance tier
kubectl label nodes node-2 performance-tier=high

# By environment
kubectl label nodes node-3 environment=testing

# By instance type (AWS)
kubectl label nodes node-4 node.kubernetes.io/instance-type=c5.2xlarge
```

## Configure node affinity

Add `scheduling.affinity.nodeAffinity` to your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: affinity-test
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
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:  # Hard requirement
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type
                  operator: In
                  values:
                    - load-testing  # Only schedule on nodes with this label
```

Apply the configuration:

```bash
kubectl apply -f locusttest-affinity.yaml
```

## Multiple label requirements

Require multiple labels on nodes (AND logic):

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: multi-label-affinity
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
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type           # Must be load-testing
                  operator: In
                  values:
                    - load-testing
                - key: performance-tier        # AND must be high-performance
                  operator: In
                  values:
                    - high
                - key: environment             # AND must be in testing env
                  operator: In
                  values:
                    - testing
```

All conditions must be true for a node to be selected.

## Example: AWS instance type targeting

Target specific EC2 instance types:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: aws-instance-affinity
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
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: node.kubernetes.io/instance-type
                  operator: In
                  values:
                    - c5.2xlarge   # Compute-optimized instances
                    - c5.4xlarge
```

## Example: Zone isolation

Keep tests in specific availability zones:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: zone-affinity-test
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
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: topology.kubernetes.io/zone
                  operator: In
                  values:
                    - us-east-1a  # Only use nodes in zone 1a
```

## Verify node placement

Check that pods are scheduled on the correct nodes:

```bash
# Show pod-to-node mapping
kubectl get pods -l performance-test-name=affinity-test -o wide

# Check specific labels on nodes where pods are running
kubectl get nodes -l workload-type=load-testing
```

Expected output showing pods only on labeled nodes:

```
NAME                           NODE                          STATUS
affinity-test-master-abc123    node-1 (workload=load-test)   Running
affinity-test-worker-0         node-1 (workload=load-test)   Running
affinity-test-worker-1         node-2 (workload=load-test)   Running
affinity-test-worker-2         node-3 (workload=load-test)   Running
```

## Troubleshoot scheduling failures

If pods remain in `Pending` state:

```bash
# Check pod events
kubectl describe pod <pod-name> | grep -A 10 "Events:"
```

**Common issue:**

```
Warning  FailedScheduling  No nodes are available that match all of the following predicates: NodeAffinity (3)
```

**Causes:**

1. **No nodes with matching labels:**

   ```bash
   # Check labeled nodes exist
   kubectl get nodes -l workload-type=load-testing
   ```

   Fix: Label at least one node.

2. **Insufficient capacity on labeled nodes:**

   ```bash
   # Check node resources
   kubectl describe nodes -l workload-type=load-testing | grep -A 5 "Allocated resources"
   ```

   Fix: Add more nodes with the label or reduce resource requests.

3. **Typo in label key or value:**

   Verify label spelling matches exactly:

   ```bash
   kubectl get nodes --show-labels | grep workload
   ```

## Combine with tolerations

Often used together for dedicated node pools:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: affinity-toleration-test
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
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type
                  operator: In
                  values:
                    - load-testing
    tolerations:
      - key: dedicated
        operator: Equal
        value: load-testing
        effect: NoSchedule
```

See [Configure tolerations](configure-tolerations.md) for details.

## What's next

- **[Configure tolerations](configure-tolerations.md)** — Schedule on tainted nodes (often used together)
- **[Use node selector](use-node-selector.md)** — Simpler alternative for basic label matching
- **[Scale worker replicas](scale-workers.md)** — Calculate worker count for dedicated nodes

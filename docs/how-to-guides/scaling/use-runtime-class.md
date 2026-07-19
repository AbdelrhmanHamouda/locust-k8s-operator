---
title: Use a runtime class for sandboxed pods
description: Run Locust pods with an alternative container runtime such as gVisor or Kata Containers
tags:
  - scaling
  - scheduling
  - runtime class
  - security
---

# Use a runtime class for sandboxed pods

Run Locust master and worker pods with an alternative container runtime — such as gVisor or Kata Containers — using the standard Kubernetes `runtimeClassName` field.

## Prerequisites

- Locust Kubernetes Operator installed
- A [RuntimeClass](https://kubernetes.io/docs/concepts/containers/runtime-class/) configured in your cluster (e.g. gVisor's `runsc` handler on your nodes)

## When to use a runtime class

**Use a runtime class when:**

- Your security policy requires third-party or user-supplied code to run sandboxed (Locust workers execute arbitrary locustfile Python)
- You run on GKE Sandbox or another platform that exposes sandboxed runtimes via RuntimeClass
- You need kernel-level isolation between load-generation pods and the host

**Skip it when:**

- Your cluster has no RuntimeClass objects configured — pods referencing a missing class will never schedule
- You need maximum load-generation throughput from minimal hardware (see the performance note below)

## Verify a RuntimeClass exists

```bash
kubectl get runtimeclass
```

Example output on a cluster with gVisor:

```
NAME     HANDLER   AGE
gvisor   runsc     42d
```

## Configure per test (CR field)

Add `scheduling.runtimeClassName` to your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: sandboxed-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  scheduling:
    runtimeClassName: gvisor  # Master and worker pods run sandboxed
```

Apply the configuration:

```bash
kubectl apply -f locusttest-sandboxed.yaml
```

The field applies to both master and worker pods. The value must be a valid
[DNS-1123 subdomain](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names);
the CRD schema rejects malformed names at admission time.

## Configure an operator-wide default

If every test in the cluster must run sandboxed, set a default on the operator instead of editing each CR. Helm value:

```yaml
# values.yaml
locustPods:
  runtimeClassName: gvisor
```

This sets the `DEFAULT_RUNTIME_CLASS_NAME` environment variable on the operator, which applies to every generated master and worker pod.

**Precedence:** a `scheduling.runtimeClassName` set in a LocustTest CR always wins over the operator-wide default. If neither is set, pods use the cluster's default runtime.

## Sandbox the operator pod itself

To run the operator pod under the same runtime (common in environments where *all* third-party workloads must be sandboxed):

```yaml
# values.yaml (top level)
runtimeClassName: gvisor
```

This only affects the operator Deployment, not the generated Locust pods.

## Verify runtime placement

```bash
# Confirm the field landed on the pods
kubectl get pods -l performance-test-name=<test-name> \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.runtimeClassName}{"\n"}{end}'
```

Expected: every master/worker pod lists the configured class.

## Troubleshoot scheduling failures

If pods remain `Pending`:

```bash
kubectl describe pod <pod-name> | grep -A 10 "Events:"
```

**Common issue:**

```
Warning  FailedCreatePodSandBox  ... RuntimeClass "gvisor" not found
```

**Causes:**

1. **RuntimeClass object missing:**

   ```bash
   kubectl get runtimeclass gvisor
   ```

   Fix: install the runtime and create the RuntimeClass, or correct the name.

2. **Runtime handler not present on any node:** the RuntimeClass exists but no node runs the handler (e.g. `runsc` not installed). Use the RuntimeClass `scheduling` stanza or combine with [node selector](use-node-selector.md) to target nodes that have the runtime.

!!! warning "Performance overhead"

    Sandboxed runtimes intercept syscalls, and load generators are syscall- and
    network-heavy by nature. Expect measurably lower throughput per worker under
    gVisor compared to runc — plan to scale
    [worker replicas](scale-workers.md) up accordingly, and validate your target
    request rate before relying on results.

## What's next

- **[Use node selector](use-node-selector.md)** — Target nodes that have the runtime handler installed
- **[Configure tolerations](configure-tolerations.md)** — Schedule on dedicated sandboxed node pools
- **[Scale worker replicas](scale-workers.md)** — Compensate for sandbox overhead with more workers

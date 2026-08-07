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
[DNS-1123 subdomain](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names)
— or the empty string, which is the explicit opt-out described below. The CRD schema rejects
malformed names at admission time.

## Configure an operator-wide default

If every test in the cluster must run sandboxed, set a default on the operator instead of editing each CR. Helm value:

```yaml
# values.yaml
locustPods:
  runtimeClassName: gvisor
```

This sets the `DEFAULT_RUNTIME_CLASS_NAME` environment variable on the operator, which applies to every generated master and worker pod. The operator validates this value at startup and refuses to start if it is not a valid DNS-1123 subdomain — a typo fails loudly once rather than breaking every test run.

**Precedence:**

| CR `scheduling.runtimeClassName` | Result |
| --- | --- |
| Set to a name | That name is used; the operator-wide default is ignored. |
| Omitted | The operator-wide default is used, if one is configured. |
| Set to `""` (empty string) | Explicit opt-out — the pod runs on the cluster's default runtime even when an operator-wide default is configured. |

Opting a single test out of the operator-wide default:

```yaml
spec:
  scheduling:
    runtimeClassName: ""  # Ignore the operator default; use the cluster default runtime
```

!!! note "Resolved once, at creation"

    The runtime class is resolved when the operator creates the test's Jobs. Editing
    `scheduling.runtimeClassName` on an existing LocustTest has no effect — the operator does not
    support in-place updates. Delete and re-create the LocustTest to change it.

    The field is also v2-only: a LocustTest round-tripped through the `v1` API loses it, so an
    explicit `""` opt-out becomes "inherit the operator default" again. Use `locust.io/v2`.

!!! warning "The operator-wide default is a default, not an enforcement boundary"

    `locustPods.runtimeClassName` sets a *starting value*, not a guarantee. Anyone who can create a
    LocustTest can override it — with a different RuntimeClass, or with `""` to opt out entirely.
    That is deliberate: the operator defaults, it does not police.

    If your environment *requires* that every Locust pod runs sandboxed, enforce it where
    enforcement belongs — at admission, with a cluster policy engine such as
    [Kyverno](https://kyverno.io/) or a built-in
    [ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/),
    which can reject or mutate any pod that does not carry the required `runtimeClassName`
    regardless of what the CR asked for. Restricting who may create LocustTest resources via RBAC
    is a useful complement, not a substitute.

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

## Troubleshoot runtime failures

A pod that names a RuntimeClass the cluster cannot honour is still admitted and still scheduled — it
then enters the `Failed` phase rather than sitting in `Pending`. The Job retries until it hits its
backoff limit, and the LocustTest ends up `Failed`.

```bash
kubectl get pods -l performance-test-name=<test-name>
kubectl describe pod <pod-name> | grep -A 10 "Events:"
kubectl get events -n <namespace> --sort-by=.lastTimestamp | tail -20
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

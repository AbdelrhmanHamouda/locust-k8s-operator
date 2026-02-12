---
title: Configure pod security settings
description: Understand and configure security contexts, RBAC, and network policies for operator and test pods
tags:
  - security
  - pod security
  - rbac
  - network policy
  - hardening
---

# Configure pod security settings

The operator applies security settings to all test pods by default. This guide explains the default security posture, RBAC requirements, and network isolation options.

## Default security context

The operator automatically applies a security context to all Locust test pods (master and worker). This security context meets Kubernetes Pod Security Standards **"baseline"** profile requirements.

### Security settings applied

```yaml
# Applied to all test pods by default
securityContext:
  seccompProfile:
    type: RuntimeDefault                # Use runtime's default seccomp profile
```

**Why this default:**

- **seccompProfile: RuntimeDefault** — Uses the container runtime's default seccomp profile to restrict system calls.

!!! note "Non-root execution"
    The official Locust image (`locustio/locust`) runs as a non-root user by default (UID 1000), but the operator does not explicitly set `runAsNonRoot: true` on the pod security context. If you require enforced non-root execution, see the [restricted profile section](#pod-security-standards-compliance) below.

### Why NOT readOnlyRootFilesystem

The operator does NOT set `readOnlyRootFilesystem: true` because:

- Locust needs to write to `/tmp` for temporary files
- Python pip may need cache directories for plugin installation
- The locustfile may write temporary data during test execution

If your test doesn't require write access, you can customize the security context (see below).

## Customizing security context

Override the default security context per test if needed:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: hardened-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
    # Custom security context for master pod (if needed)
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
    # Custom security context for worker pod (if needed)
```

!!! note
    The v2 API does not currently expose securityContext at the CR level. The operator uses defaults from Helm values. To customize per-test, you would need to modify the operator's Helm chart values.

## RBAC best practices

### Operator RBAC

The operator's service account needs permissions to manage LocustTest resources and create test infrastructure.

**What the operator needs:**

| Resource | Verbs | Purpose |
|----------|-------|---------|
| `locusttests` | get, list, watch, update, patch | Watch CRs and reconcile state |
| `locusttests/status` | get, update, patch | Report test status |
| `locusttests/finalizers` | update | Manage deletion lifecycle |
| `configmaps` | get, list, watch | Read test files and library code |
| `secrets` | get, list, watch | Read credentials for env injection |
| `services` | get, list, watch, create, delete | Master service for worker communication |
| `pods` | get, list, watch | Monitor pod health for status reporting |
| `events` | create, patch | Report status changes and errors |
| `jobs` | get, list, watch, create, delete | Master and worker pods (immutable pattern) |

!!! note "Read-only Secret access"
    The operator **never creates or modifies** ConfigMaps or Secrets. It only reads them to populate environment variables and volume mounts in test pods.

**ClusterRole vs Role:**

The operator supports two RBAC modes (configured via Helm):

| Mode | Scope | Use case |
|------|-------|----------|
| ClusterRole (default) | All namespaces | Multi-tenant platform, centralized operator |
| Role | Single namespace | Security-sensitive environments, namespace isolation |

Configure in Helm values:

```yaml
# values.yaml
k8s:
  clusterRole:
    enabled: false  # Restrict to operator namespace only
```

### Test pod RBAC

Test pods run as non-root and do **not** get elevated privileges. By default:

- No service account token is mounted (unless you explicitly set `serviceAccountName`)
- No Kubernetes API access
- No elevated Linux capabilities

**If your test needs Kubernetes API access:**

1. Create a service account with minimal permissions:
   ```bash
   kubectl create serviceaccount locust-test-runner
   ```

2. Grant only required permissions:
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: Role
   metadata:
     name: locust-pod-reader
   rules:
     - apiGroups: [""]
       resources: ["pods"]
       verbs: ["get", "list"]
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
     name: locust-pod-reader-binding
   subjects:
     - kind: ServiceAccount
       name: locust-test-runner
   roleRef:
     kind: Role
     name: locust-pod-reader
     apiGroup: rbac.authorization.k8s.io
   ```

3. Reference in LocustTest CR (not yet supported in v2 API).

!!! warning "Least privilege"
    Only grant the minimum permissions your test needs. Avoid `cluster-admin` or broad wildcard permissions.

### User RBAC for test creators

Users who create and manage LocustTest CRs need different permissions than the operator.

**Minimal test creator role:**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: locusttest-creator
  namespace: performance-testing
rules:
  # Create and manage LocustTest CRs
  - apiGroups: ["locust.io"]
    resources: ["locusttests"]
    verbs: ["get", "list", "watch", "create", "delete"]

  # Create ConfigMaps for test files
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "create", "update", "delete"]

  # View pods for debugging
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list"]

  # View events for status monitoring
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["get", "list", "watch"]
```

**Verify user permissions:**

```bash
# Check if user can create LocustTest
kubectl auth can-i create locusttests --as jane.doe

# Check if user can read secrets (should be "no")
kubectl auth can-i get secrets --as jane.doe
```

## Network isolation

Use NetworkPolicies to restrict traffic to/from test pods.

### Allow only necessary traffic

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: locust-test-isolation
  namespace: performance-testing
spec:
  podSelector:
    matchLabels:
      performance-test-name: my-test    # Apply to specific test
  policyTypes:
    - Ingress
    - Egress

  ingress:
    # Allow communication between pods in the same test
    - from:
        - podSelector:
            matchLabels:
              performance-test-name: my-test
      ports:
        - port: 5557                    # Worker -> Master
          protocol: TCP

  egress:
    # Allow DNS resolution
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
      ports:
        - port: 53
          protocol: UDP

    # Allow traffic to target system under test
    - to:
        - podSelector: {}               # All pods (adjust as needed)
      ports:
        - port: 80
          protocol: TCP
        - port: 443
          protocol: TCP

    # Allow traffic to OTel Collector (if using OpenTelemetry)
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: monitoring
        - podSelector:
            matchLabels:
              app: otel-collector
      ports:
        - port: 4317                    # OTLP gRPC
          protocol: TCP
```

**What this policy allows:**

- **Ingress:** Only communication between pods in the same test (master ↔ workers)
- **Egress:** DNS, target system (HTTP/HTTPS), OTel Collector

**What this policy blocks:**

- Cross-test communication
- External egress except explicitly allowed
- Ingress from outside the test

### Verification

**Check if NetworkPolicy is active:**

```bash
kubectl get networkpolicy -n performance-testing
```

**Test connectivity from a worker pod:**

```bash
# Get a worker pod
POD=$(kubectl get pods -l performance-test-pod-name=my-test-worker -o jsonpath='{.items[0].metadata.name}')

# Test target system connectivity
kubectl exec $POD -- curl -I https://api.example.com

# Test master connectivity
kubectl exec $POD -- nc -zv my-test-master 5557

# Test blocked traffic (should timeout or fail)
kubectl exec $POD -- curl -I https://blocked-host.com --max-time 5
```

### NetworkPolicy best practices

1. **Start with allow-all, then restrict:** Test your application first, then add NetworkPolicies gradually.

2. **Allow DNS:** Always allow egress to `kube-system` namespace port 53 for DNS resolution.

3. **Test-specific policies:** Use `performance-test-name` label to isolate individual tests.

4. **Monitor denied traffic:** Use a CNI that logs dropped packets (Calico, Cilium) to identify blocked traffic.

5. **Document exceptions:** If you must allow broad egress, document why in the NetworkPolicy annotations.

## Verification

### Check pod security context

```bash
# Get pod security context
kubectl get pod -l performance-test-name=my-test -o jsonpath='{.items[0].spec.securityContext}' | jq .
```

**Expected output:**

```json
{
  "seccompProfile": {
    "type": "RuntimeDefault"
  }
}
```

### Verify non-root execution

```bash
# Check which user the pod runs as
POD=$(kubectl get pods -l performance-test-pod-name=my-test-master -o jsonpath='{.items[0].metadata.name}')
kubectl exec $POD -- id
```

**Expected output:**

```
uid=1000(locust) gid=1000(locust) groups=1000(locust)
```

If you see `uid=0(root)`, the pod is running as root (violation of security policy).

### Verify RBAC permissions

```bash
# Check operator service account permissions
kubectl auth can-i --list --as=system:serviceaccount:locust-operator-system:locust-k8s-operator

# Check if test pod has Kubernetes API access (should be "no" by default)
kubectl exec $POD -- curl -k https://kubernetes.default.svc
```

Expected: Connection refused or authentication error (test pods should NOT have API access by default).

## Pod Security Standards compliance

The operator's default security settings meet these Pod Security Standards profiles:

| Profile | Compliant | Notes |
|---------|-----------|-------|
| **Baseline** | ✅ Yes | Seccomp profile satisfies baseline requirements |
| **Restricted** | ⚠️ Partial | Missing: `runAsNonRoot`, `allowPrivilegeEscalation=false`, `capabilities drop ALL` |
| **Privileged** | ✅ Yes | No restrictions |

**To meet "restricted" profile:**

Users would need to add the following settings (via Helm values customization):

```yaml
securityContext:
  runAsNonRoot: true                     # Add this
  allowPrivilegeEscalation: false        # Add this
  seccompProfile:
    type: RuntimeDefault
  capabilities:                          # Add this
    drop:
      - ALL
```

Currently, the operator meets "baseline" by default. For "restricted" compliance, customize the operator's Helm values.

## Related guides

- [Inject secrets and configuration](inject-secrets.md) — Manage credentials for test pods
- [Security Best Practices](../../security.md) — Complete security guide (RBAC, secrets, external integrations)
- [API Reference](../../api_reference.md) — LocustTest CR specification

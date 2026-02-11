---
title: Security Best Practices
description: RBAC configuration, secret management, and security hardening for the Locust Kubernetes Operator
tags:
  - security
  - rbac
  - secrets
  - guide
---

# Security Best Practices

This guide covers security best practices for deploying and operating the Locust Kubernetes Operator in production environments. It provides practical examples for RBAC configuration, secret management, and security hardening.

## Operator RBAC Permissions

### What the Operator Needs

The operator follows least-privilege principles. It requires specific permissions to manage LocustTest resources and create load test infrastructure.

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
| `leases` | get, list, watch, create, update, patch | Leader election (only when HA enabled) |

!!! note "Read-Only Secret Access"
    The operator **never creates or modifies** ConfigMaps or Secrets. It only reads them to populate environment variables and volume mounts in test pods. Users manage Secret creation and rotation.

### Namespace-Scoped vs Cluster-Scoped

The operator supports two RBAC modes:

**ClusterRole** (`k8s.clusterRole.enabled: true`, default)

- Operator manages LocustTest CRs in **all namespaces**
- Use when multiple teams share one operator deployment
- Typical for platform teams managing centralized performance testing

**Role** (`k8s.clusterRole.enabled: false`)

- Operator limited to its **deployment namespace**
- Use for single-tenant deployments or strict namespace isolation
- Typical for security-sensitive environments

Configure the mode in Helm values:

```yaml
# values.yaml
k8s:
  clusterRole:
    enabled: false  # Restrict to operator namespace only
```

### User RBAC for Test Creators

Users who create and manage LocustTest CRs need different permissions than the operator itself. Here are minimal RBAC examples:

**Test Creator Role** (create and manage performance tests):

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
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: locusttest-creator-binding
  namespace: performance-testing
subjects:
  - kind: User
    name: jane.doe
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: locusttest-creator
  apiGroup: rbac.authorization.k8s.io
```

**Test Viewer Role** (read-only access for monitoring):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: locusttest-viewer
  namespace: performance-testing
rules:
  # View LocustTest CRs
  - apiGroups: ["locust.io"]
    resources: ["locusttests"]
    verbs: ["get", "list", "watch"]
  # View pods and logs
  - apiGroups: [""]
    resources: ["pods", "pods/log", "events"]
    verbs: ["get", "list", "watch"]
```

Users with this role can monitor test status and view logs but cannot create or modify tests.

## Secret Management

### Injecting Secrets into Tests

The operator provides three approaches for injecting secrets into Locust test pods:

| Method | Best For | Configuration |
|--------|----------|---------------|
| **Secret environment variables** (`env.secretRefs`) | API keys, tokens, passwords | Mounts all keys from a Secret as environment variables |
| **Secret file mounts** (`env.secretMounts`) | Certificates, key files, config files | Mounts Secret keys as files in the container filesystem |
| **Individual secret references** (`env.variables[].valueFrom.secretKeyRef`) | Specific keys from a secret | Fine-grained control over which keys to inject |

**Quick Example** (secret environment variables):

```yaml
apiVersion: locust.io/v2
kind: LocustTest
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  env:
    secretRefs:
      - name: api-credentials  # All keys become env vars
```

See [Advanced Topics - Environment Variables](advanced_topics.md#environment-variables) for detailed examples of all three approaches.

### Secret Rotation

Because tests are immutable, running tests continue to use the secret values they started with. Secret rotation requires recreating the test.

**Rotation Process:**

1. **Update the Secret** in Kubernetes with new credentials:
   ```bash
   kubectl create secret generic api-credentials \
     --from-literal=API_TOKEN=new-token-value \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Delete the LocustTest CR**:
   ```bash
   kubectl delete locusttest my-test
   ```

3. **Recreate the LocustTest CR** — new pods pick up updated secret values:
   ```bash
   kubectl apply -f locusttest.yaml
   ```

!!! tip "Scheduled Rotation"
    For automated secret rotation, integrate with external secrets management tools (see next section) that synchronize secrets on a schedule.

### External Secrets Integration

The operator works seamlessly with [External Secrets Operator](https://external-secrets.io/) for automatic secret synchronization from external secrets managers (AWS Secrets Manager, HashiCorp Vault, Google Secret Manager, Azure Key Vault, etc.).

**Example** (AWS Secrets Manager integration):

```yaml
# ExternalSecret syncs from AWS Secrets Manager to a K8s Secret
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: load-test-credentials
  namespace: performance-testing
spec:
  refreshInterval: 1h  # Sync every hour
  secretStoreRef:
    name: aws-secretsmanager
    kind: ClusterSecretStore
  target:
    name: load-test-credentials  # K8s Secret name
  data:
    - secretKey: API_TOKEN
      remoteRef:
        key: /perf-testing/api-token
    - secretKey: DB_PASSWORD
      remoteRef:
        key: /perf-testing/db-password
---
# LocustTest references the synced Secret
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: my-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  env:
    secretRefs:
      - name: load-test-credentials  # Uses synced Secret
```

!!! tip "Secret Source Agnostic"
    The operator doesn't care how Secrets are created. You can use External Secrets Operator, Sealed Secrets, Vault Agent, manual `kubectl create secret`, or any other method.

## Pod Security

### Operator Pod Security

The operator runs with a hardened security context by default, meeting Kubernetes Pod Security Standards **"restricted"** profile:

```yaml
# From values.yaml (default configuration)
securityContext:
  runAsNonRoot: true
  runAsUser: 65532  # Non-root user
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  seccompProfile:
    type: RuntimeDefault
```

These settings are enabled by default in the Helm chart. No additional configuration is required.

### Test Pod Security

Test pods run the user-provided Locust image. Security depends on the image you use.

**Recommendations:**

- Use the official `locustio/locust` image or build a hardened variant
- Avoid running test containers as root
- Set resource limits to prevent resource exhaustion:
  ```yaml
  master:
    resources:
      limits:
        cpu: 2000m
        memory: 2Gi
  worker:
    resources:
      limits:
        cpu: 1000m
        memory: 1Gi
  ```

Test pods inherit the default security context from Helm values (`locustPods.securityContext`). Override per-test if needed.

## Network Security

### Master-Worker Communication

Master and worker pods communicate internally within the cluster:

- **Port 5557**: Master listens for worker connections (internal only)
- **Port 8089**: Web UI on master pod (use port-forward for access)

For production use:

- **Do not expose port 8089 externally** — use `kubectl port-forward` for temporary access
- If using NetworkPolicies, ensure master and worker pods can communicate

### NetworkPolicy Example

Restrict pod communication to within the same test:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: locust-internal
  namespace: performance-testing
spec:
  podSelector:
    matchLabels:
      performance-test-name: my-test
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
        - port: 5557  # Worker -> Master
        - port: 8089  # Web UI (optional)
  egress:
    # Allow all egress (tests need to reach target systems)
    - {}
```

!!! note "Egress Requirements"
    Test pods need egress access to reach the target system under test. The example above allows unrestricted egress. Restrict further if required by your security policies.

### Service Mesh Compatibility

The operator is compatible with service mesh solutions (Istio, Linkerd). However:

- Master-worker communication on port 5557 must work within the mesh
- Ensure sidecar injection doesn't break pod startup (adjust readiness probes if needed)
- Test traffic to external targets may require egress configuration in the mesh

## Image Security

### Using Private Registries

If Locust images are in a private registry, configure image pull secrets:

**Helm Configuration:**

```yaml
# values.yaml
image:
  pullSecrets:
    - name: my-registry-secret
```

**Create the pull secret:**

```bash
kubectl create secret docker-registry my-registry-secret \
  --docker-server=registry.example.com \
  --docker-username=user \
  --docker-password=pass \
  --docker-email=user@example.com \
  -n performance-testing
```

### Image Scanning

Scan Locust images for vulnerabilities before use:

```bash
# Example with Trivy
trivy image locustio/locust:2.20.0
```

Build custom hardened images if the official image doesn't meet security requirements.

## Audit and Compliance

### Operator Audit Logging

Kubernetes audit logs capture all operator actions. Enable audit logging at the cluster level to track:

- LocustTest CR creation/deletion
- Job creation by the operator
- Secret access attempts

### Compliance Considerations

- **PCI-DSS**: Ensure Secrets are encrypted at rest (etcd encryption)
- **SOC 2**: Log all operator actions via audit logs
- **GDPR**: Avoid storing personal data in LocustTest CRs or test results

## Additional Resources

- [Getting Started](getting_started.md) — Initial setup and first test
- [Advanced Topics](advanced_topics.md) — Environment variables, volumes, resource management
- [API Reference](api_reference.md) — Complete CR specification
- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [External Secrets Operator](https://external-secrets.io/)

# Phase 10: Environment & Secret Injection

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Status:** ✅ Complete  
**Completed:** 2026-01-20  
**Dependencies:** Phase 7 (v2 API Types), Phase 9 (Status Subresource)

---

## Objective

Enable ConfigMap and Secret injection into Locust pods, allowing users to securely pass credentials, API keys, and configuration without hardcoding them in test files. This addresses Issue #149, a highly requested feature with strong community support.

---

## Requirements Reference

- **REQUIREMENTS.md §5.1.2:** Environment Injection
- **Issue #149:** Inject Kubernetes Secrets into Locust Worker Pods
- **issue-analysis/P1-High/issue-149-secrets-injection.md:** Detailed analysis

---

## Scope

### In Scope

- Process `spec.env.configMapRefs` → `envFrom` with ConfigMapEnvSource
- Process `spec.env.secretRefs` → `envFrom` with SecretEnvSource  
- Process `spec.env.variables` → individual `env` entries
- Process `spec.env.secretMounts` → Volume + VolumeMount for secret files
- Support prefix for ConfigMap/Secret env injection
- Validation webhook to prevent path conflicts with operator-managed paths
- Unit and integration tests for all injection types

### Out of Scope

- Vault integration (secrets should be synced via External Secrets Operator)
- Dynamic secret rotation detection
- Env injection into metrics exporter sidecar (future enhancement)

---

## Key Deliverables

| File | Description |
|------|-------------|
| `internal/resources/env.go` | Environment variable building functions |
| `internal/resources/env_test.go` | Unit tests for env builders |
| `internal/resources/job.go` | Updated to use env builders |
| `api/v2/locusttest_webhook.go` | Validation webhook for path conflicts |
| `api/v2/locusttest_webhook_test.go` | Webhook validation tests |

---

## Success Criteria

1. ConfigMap values available as environment variables in master/worker pods
2. Secret values available as environment variables in master/worker pods
3. Prefix support works correctly for namespaced env vars
4. Secrets mountable as files at user-specified paths
5. Validation rejects paths that conflict with operator-managed paths (`/lotest/src/`, `/opt/locust/lib`)
6. All tests pass with ≥80% coverage for new code
7. Backward compatible with existing deployments (empty `env` spec = no change)

---

## Example Usage

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: load-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "locust -f /lotest/src/locustfile.py"
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 5
  testFiles:
    configMapRef: locust-scripts
  env:
    # Inject all keys from a ConfigMap as env vars
    configMapRefs:
      - name: app-config
        prefix: "APP_"  # Results in APP_KEY1, APP_KEY2, etc.
    # Inject all keys from a Secret as env vars
    secretRefs:
      - name: api-credentials
        prefix: ""  # Results in API_TOKEN, API_SECRET, etc.
    # Individual environment variables
    variables:
      - name: TARGET_HOST
        value: "https://api.example.com"
      - name: LOG_LEVEL
        valueFrom:
          configMapKeyRef:
            name: app-config
            key: log-level
    # Mount secrets as files
    secretMounts:
      - name: tls-certs
        mountPath: /etc/locust/certs
        readOnly: true
```

---

## Quick Start

```bash
# After implementation, verify with:
make generate
make manifests
make build
make test

# Test env injection manually
kubectl apply -f config/samples/locust_v2_locusttest_with_env.yaml
kubectl exec -it <pod-name> -- env | grep APP_
kubectl exec -it <pod-name> -- ls /etc/locust/certs
```

---

## Related Documents

- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist
- [DESIGN.md](./DESIGN.md) - Technical design and code patterns

# Phase 13: Helm Chart Updates

**Effort:** 1 day  
**Priority:** P0 - Critical Path  
**Status:** Complete  
**Dependencies:** Phase 6 (Integration Tests)

---

## Objective

Rewrite the Helm chart with a **clean-slate design** optimized for the Go operator, then add backward compatibility shims for existing users. This approach produces a simpler, more maintainable chart (~30% fewer lines) while preserving upgrade paths.

---

## Requirements Reference

- **ROADMAP.md §Phase 13:** Helm Chart Updates
- **REQUIREMENTS.md §6.2:** Helm Chart

---

## Background

### Current State (Java Operator)

The existing Helm chart deploys a Java/Micronaut-based operator with:
- JVM-specific environment variables (METRICS_JVM_ENABLE, etc.)
- Micronaut health endpoints (`/health` on port 8080)
- Kafka integration for event streaming
- Higher default memory limits suited for JVM (~1Gi)
- `containersol/locust_exporter` sidecar configuration

### Target State (Go Operator)

The Go operator is significantly lighter and uses different patterns:
- No JVM configuration needed
- Controller-runtime health endpoints (`/healthz`, `/readyz` on port 8081)
- Leader election for HA deployments
- Much lower memory requirements (~64Mi default)
- Native OTel support replaces sidecar (optional)

---

## Scope

### In Scope

- **Rewrite** `values.yaml` with clean, flat structure
- **Rewrite** `deployment.yaml` for Go operator (~80 lines vs 155)
- Update `Chart.yaml` to v2.0.0
- Add backward compatibility helpers in `_helpers.tpl`
- Update RBAC for new API groups
- Add optional webhook/OTel Collector templates

### Out of Scope

- Migration tooling from v1 to v2 chart
- Kafka removal (keep via compat helpers)

---

## Key Deliverables

| File | Action | Description |
|------|--------|-------------|
| `Chart.yaml` | Modify | Version 2.0.0 |
| `values.yaml` | **Rewrite** | Clean structure (~100 lines vs 144) |
| `templates/deployment.yaml` | **Rewrite** | Clean spec (~80 lines vs 155) |
| `templates/_helpers.tpl` | Modify | Add backward compat helpers |
| `templates/serviceaccount-and-roles.yaml` | Modify | New RBAC rules |
| `templates/webhook.yaml` | Create | Optional webhook config |
| `templates/certificate.yaml` | Create | Optional cert-manager |
| `templates/otel-collector.yaml` | Create | Optional OTel Collector |

---

## Success Criteria

1. `helm lint` passes
2. Chart installs on Kind cluster
3. Operator pod becomes ready
4. Test CR creates Jobs/Service
5. Old value paths work via helpers (backward compat)
6. Optional features deploy when enabled

---

## Breaking Changes

| Change | Impact | Migration |
|--------|--------|-----------|
| Health endpoint path | Probes reference new paths | Auto-updated via chart |
| Health endpoint port | Port 8081 instead of 8080 | Auto-updated via chart |
| JVM options removed | No impact on users | N/A |
| Kafka deprecated | Users relying on Kafka | Keep optional, mark deprecated |
| Memory defaults | Lower defaults | Users can override |

---

## Clean values.yaml Structure

```yaml
# Operator config
image:
  repository: lotest/locust-k8s-operator
  tag: ""  # Defaults to appVersion
resources:
  limits: { cpu: 500m, memory: 128Mi }
  requests: { cpu: 10m, memory: 64Mi }
replicaCount: 1

# Feature toggles (top-level, flat)
leaderElection:
  enabled: true
metrics:
  enabled: false
  port: 8080
webhook:
  enabled: true
  port: 9443
  certManager:
    enabled: true

# Locust test pod config
locustPods:
  resources:
    requests: { cpu: 250m, memory: 128Mi }
    limits: { cpu: 1000m, memory: 1024Mi }
  affinityInjection: true
  tolerationsInjection: true
  metricsExporter:
    image: containersol/locust_exporter:v0.5.0
    port: 9646

# Optional OTel Collector
otelCollector:
  enabled: false
```

## Backward Compatibility

Old value paths still work via helper functions:

| Old Path | New Path |
|----------|----------|
| `config.loadGenerationPods.resource.cpuRequest` | `locustPods.resources.requests.cpu` |
| `config.loadGenerationPods.affinity.enableCrInjection` | `locustPods.affinityInjection` |
| `micronaut.*` | **Removed** (no Go equivalent) |
| `appPort` | **Removed** (fixed at 8081) |

---

## Quick Start

```bash
# After implementation, verify with:
helm lint charts/locust-k8s-operator
helm template test charts/locust-k8s-operator

# Dry-run install
helm install --dry-run test charts/locust-k8s-operator

# Actual install on test cluster
helm install locust-operator charts/locust-k8s-operator

# Verify operator is running
kubectl get pods -l app.kubernetes.io/name=locust-k8s-operator
kubectl logs -l app.kubernetes.io/name=locust-k8s-operator

# Test CRD and CR creation
kubectl apply -f config/samples/locust_v1_locusttest.yaml
kubectl get locusttests
```

---

## Related Documents

- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist
- [DESIGN.md](./DESIGN.md) - Technical design and template changes

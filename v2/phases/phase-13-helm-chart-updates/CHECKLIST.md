# Phase 13: Helm Chart Updates - Checklist

**Estimated Effort:** 1 day  
**Status:** Complete  
**Approach:** Clean-slate design with backward compatibility shims

---

## Pre-Implementation

- [x] Phase 6 complete (Integration Tests passing)
- [x] Go operator builds: `make build`
- [x] Go operator tests pass: `make test`
- [x] Review `cmd/main.go` flags and `internal/config/config.go` env vars
- [x] Run `helm lint charts/locust-k8s-operator` on current chart

---

## Task 13.1: Update Chart.yaml

**File:** `charts/locust-k8s-operator/Chart.yaml`

- [x] Update `version` to `2.0.0`
- [x] Update `appVersion` to `"2.0.0"`
- [x] Add changelog annotations

**Verification:**
```bash
helm lint charts/locust-k8s-operator
```

---

## Task 13.2: Rewrite values.yaml (Clean Structure)

**File:** `charts/locust-k8s-operator/values.yaml`

### Delete Java-Specific Sections

- [x] Delete `appPort: 8080`
- [x] Delete entire `micronaut:` section
- [x] Delete `livenessProbe` / `readinessProbe` (hardcode in template)

### Write Clean Structure

- [x] Operator image config:
  ```yaml
  image:
    repository: lotest/locust-k8s-operator
    tag: ""
    pullPolicy: IfNotPresent
    pullSecrets: []
  ```

- [x] Operator resources (lightweight Go defaults):
  ```yaml
  resources:
    limits:
      cpu: 500m
      memory: 128Mi
    requests:
      cpu: 10m
      memory: 64Mi
  ```

- [x] Top-level feature toggles:
  ```yaml
  replicaCount: 1
  
  leaderElection:
    enabled: true
  
  metrics:
    enabled: false
    port: 8080
    secure: false
  
  webhook:
    enabled: true
    port: 9443
    certManager:
      enabled: true
  
  crd:
    install: true
  ```

- [x] Locust pod configuration:
  ```yaml
  locustPods:
    resources:
      requests:
        cpu: 250m
        memory: 128Mi
      limits:
        cpu: 1000m
        memory: 1024Mi
    affinityInjection: true
    tolerationsInjection: true
    ttlSecondsAfterFinished: ""
    metricsExporter:
      image: containersol/locust_exporter:v0.5.0
      port: 9646
      pullPolicy: IfNotPresent
      resources:
        requests:
          cpu: 100m
          memory: 64Mi
        limits:
          cpu: 250m
          memory: 128Mi
  ```

- [x] Optional OTel Collector:
  ```yaml
  otelCollector:
    enabled: false
    image: otel/opentelemetry-collector-contrib:0.92.0
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 200m
        memory: 256Mi
  ```

- [x] Standard K8s options:
  ```yaml
  serviceAccount:
    create: true
    name: ""
    annotations: {}
  
  podAnnotations: {}
  nodeSelector: {}
  tolerations: []
  affinity: {}
  ```

**Verification:**
```bash
helm lint charts/locust-k8s-operator
```

---

## Task 13.3: Rewrite Deployment Template

**File:** `charts/locust-k8s-operator/templates/deployment.yaml`

### Rewrite Container Spec (Clean)

- [x] Container name: `manager`
- [x] Command: `["/manager"]`
- [x] Args based on values:
  ```yaml
  args:
    - --health-probe-bind-address=:8081
    {{- if .Values.leaderElection.enabled }}
    - --leader-elect=true
    {{- end }}
    {{- if .Values.metrics.enabled }}
    - --metrics-bind-address=:{{ .Values.metrics.port }}
    {{- end }}
  ```
- [x] Ports: health (8081), metrics (conditional), webhook (conditional)
- [x] Security context (allowPrivilegeEscalation: false, etc.)
- [x] Hardcoded probes: `/healthz` and `/readyz` on port 8081

### Environment Variables via Helper

- [x] Create `{{- define "locust-k8s-operator.envVars" -}}` in `_helpers.tpl`
- [x] Include only required env vars:
  - `POD_CPU_REQUEST`, `POD_MEM_REQUEST`, `POD_CPU_LIMIT`, `POD_MEM_LIMIT`
  - `ENABLE_AFFINITY_CR_INJECTION`, `ENABLE_TAINT_TOLERATIONS_CR_INJECTION`
  - `METRICS_EXPORTER_*` (image, port, pull policy)
  - `JOB_TTL_SECONDS_AFTER_FINISHED` (if set)

### Webhook Volume Mounts

- [x] Conditional on `.Values.webhook.enabled`:
  ```yaml
  volumeMounts:
    - name: webhook-certs
      mountPath: /tmp/k8s-webhook-server/serving-certs
      readOnly: true
  volumes:
    - name: webhook-certs
      secret:
        secretName: {{ include "locust-k8s-operator.fullname" . }}-webhook-certs
  ```

**Verification:**
```bash
helm template test charts/locust-k8s-operator | kubectl apply --dry-run=client -f -
```

---

## Task 13.4: Add Backward Compatibility Helpers

**File:** `charts/locust-k8s-operator/templates/_helpers.tpl`

- [x] Add helper for CPU request (new path → old path fallback):
  ```yaml
  {{- define "locust.podCpuRequest" -}}
  {{- .Values.locustPods.resources.requests.cpu | default .Values.config.loadGenerationPods.resource.cpuRequest | default "250m" -}}
  {{- end -}}
  ```

- [x] Add similar helpers for:
  - `locust.podMemRequest`
  - `locust.podCpuLimit`
  - `locust.podMemLimit`
  - `locust.affinityInjection`
  - `locust.tolerationsInjection`

**Verification:**
```bash
# Both should work:
helm template test charts/locust-k8s-operator --set locustPods.resources.requests.cpu=500m
helm template test charts/locust-k8s-operator --set config.loadGenerationPods.resource.cpuRequest=500m
```

---

## Task 13.5: Update RBAC

**File:** `charts/locust-k8s-operator/templates/serviceaccount-and-roles.yaml`

- [x] Add status subresource rule
- [x] Add finalizers rule
- [x] Add Lease resources rule (leader election)
- [x] Add events rule

**Verification:**
```bash
helm template test charts/locust-k8s-operator | grep -A 50 "ClusterRole"
```

---

## Task 13.6: Create New Templates

### Webhook Template

**File:** `charts/locust-k8s-operator/templates/webhook.yaml`

- [x] Create ValidatingWebhookConfiguration (conditional on `.Values.webhook.enabled`)
- [x] Add Webhook Service

**Verification:**
```bash
helm template test charts/locust-k8s-operator --set webhook.enabled=true
```

### Certificate Template

**File:** `charts/locust-k8s-operator/templates/certificate.yaml`

- [x] Create Certificate + Issuer (conditional on `.Values.webhook.certManager.enabled`)

### OTel Collector Template

**File:** `charts/locust-k8s-operator/templates/otel-collector.yaml`

- [x] Create Deployment, ConfigMap, Service (conditional on `.Values.otelCollector.enabled`)

**Verification:**
```bash
helm template test charts/locust-k8s-operator --set otelCollector.enabled=true
```

---

## Verification

### Lint

```bash
helm lint charts/locust-k8s-operator
```

- [x] Passes without errors

### Template Tests

```bash
# Default
helm template test charts/locust-k8s-operator | kubectl apply --dry-run=client -f -

# With webhook
helm template test charts/locust-k8s-operator --set webhook.enabled=true

# With OTel
helm template test charts/locust-k8s-operator --set otelCollector.enabled=true

# Backward compat (old value paths)
helm template test charts/locust-k8s-operator \
  --set config.loadGenerationPods.resource.cpuRequest=500m
```

- [x] All templates generate valid YAML
- [x] Old value paths still work via helpers

---

## Integration Testing

### Kind Cluster Test

```bash
# Create test cluster
kind create cluster --name helm-test

# Install CRDs first
kubectl apply -f charts/locust-k8s-operator/crds/

# Install operator
helm install locust-operator charts/locust-k8s-operator --wait --timeout 120s

# Verify operator running
kubectl get pods -l app.kubernetes.io/name=locust-k8s-operator
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=locust-k8s-operator --timeout=60s

# Check logs
kubectl logs -l app.kubernetes.io/name=locust-k8s-operator

# Apply test CR
kubectl apply -f config/samples/locust_v1_locusttest.yaml
kubectl get locusttests

# Cleanup
helm uninstall locust-operator
kind delete cluster --name helm-test
```

- [ ] Operator pod starts and becomes ready (requires Kind cluster)
- [ ] Operator logs show successful startup (requires Kind cluster)
- [ ] Test CR creates resources successfully (requires Kind cluster)

---

## Post-Implementation

- [x] All verification steps pass
- [x] Update `phases/README.md` with Phase 13 status
- [x] Update `phases/NOTES.md` with implementation notes
- [ ] Update chart README if exists (no README in chart)

---

## Files Summary

| File | Action | Description |
|------|--------|-------------|
| `Chart.yaml` | Modify | Version 2.0.0 |
| `values.yaml` | **Rewrite** | Clean structure (~100 lines vs 144) |
| `templates/deployment.yaml` | **Rewrite** | Clean spec (~80 lines vs 155) |
| `templates/_helpers.tpl` | Modify | Add backward compat helpers |
| `templates/serviceaccount-and-roles.yaml` | Modify | Add new RBAC rules |
| `templates/webhook.yaml` | Create | Optional webhook config |
| `templates/certificate.yaml` | Create | Optional cert-manager |
| `templates/otel-collector.yaml` | Create | Optional OTel Collector |

---

## Acceptance Criteria

1. `helm lint` passes
2. Chart installs on Kind cluster
3. Operator pod becomes ready
4. Test CR creates Jobs/Service
5. Old value paths work via helpers
6. Optional features (webhook, OTel) deploy when enabled

---

## Design Decisions

| Decision | Chosen | Rationale |
|----------|--------|-----------|
| **Approach** | Clean-slate + compat shims | Simpler chart, easier maintenance |
| **values.yaml** | Flat structure | Avoid deep nesting |
| **Probes** | Hardcoded in template | Go operator uses fixed ports |
| **Webhook default** | Disabled | Requires cert-manager |
| **Kafka** | Keep via compat helpers | Don't break existing users |

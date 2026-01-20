# Phase 13: Helm Chart Updates - Technical Design

**Version:** 1.1  
**Status:** Draft  
**Approach:** Clean-slate design with backward compatibility shims

---

## Design Philosophy

**Clean-first, compatibility second.** Design the minimal Helm chart the Go operator needs, then add compatibility mappings for existing users.

---

## Overview

The Go operator is fundamentally simpler than the Java operator. This design starts from what the Go operator actually requires (derived from `cmd/main.go` and `internal/config/config.go`), not from the existing Java chart.

---

## 1. What the Go Operator Actually Needs

### 1.1 From `cmd/main.go` - Container Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--health-probe-bind-address` | `:8081` | Health/ready probes |
| `--metrics-bind-address` | `0` (disabled) | Prometheus metrics |
| `--leader-elect` | `false` | HA leader election |
| `--metrics-secure` | `true` | HTTPS for metrics |
| `--webhook-cert-path` | `""` | Webhook TLS certs |

### 1.2 From `internal/config/config.go` - Environment Variables

**Required for Locust pods:**
```
POD_CPU_REQUEST, POD_MEM_REQUEST, POD_CPU_LIMIT, POD_MEM_LIMIT
POD_EPHEMERAL_REQUEST, POD_EPHEMERAL_LIMIT
```

**Metrics exporter sidecar:**
```
METRICS_EXPORTER_IMAGE, METRICS_EXPORTER_PORT, METRICS_EXPORTER_IMAGE_PULL_POLICY
METRICS_EXPORTER_CPU_REQUEST, METRICS_EXPORTER_MEM_REQUEST, ...
```

**Feature flags:**
```
ENABLE_AFFINITY_CR_INJECTION, ENABLE_TAINT_TOLERATIONS_CR_INJECTION
JOB_TTL_SECONDS_AFTER_FINISHED
```

**Kafka (optional/deprecated):**
```
KAFKA_BOOTSTRAP_SERVERS, KAFKA_SECURITY_ENABLED, ...
```

### 1.3 What the Java Chart Has That's NOT Needed

| Java Chart Section | Status |
|--------------------|--------|
| `appPort: 8080` | **Remove** - Go uses fixed ports |
| `micronaut.*` | **Remove** - JVM framework config |
| `livenessProbe.port: micronaut-port` | **Replace** - Use `:8081` |
| `METRICS_JVM_ENABLE`, etc. | **Remove** - Micronaut metrics |
| Complex nested Kafka structure | **Flatten** - Simpler structure |

---

## 2. Clean values.yaml Design

### 2.1 Target Structure (Minimal)

```yaml
# =============================================================================
# Locust K8s Operator - Go v2.0.0
# =============================================================================

# -- Operator image configuration
image:
  repository: lotest/locust-k8s-operator
  tag: ""  # Defaults to appVersion
  pullPolicy: IfNotPresent
  pullSecrets: []

# -- Operator pod resources (Go binary is lightweight)
resources:
  limits:
    cpu: 500m
    memory: 128Mi
  requests:
    cpu: 10m
    memory: 64Mi

# -- Replica count (use 1 with leader election for HA)
replicaCount: 2

# -- Leader election for HA deployments
leaderElection:
  enabled: true

# -- Metrics endpoint (Prometheus)
metrics:
  enabled: false  # Disabled by default, enable if scraping
  port: 8080
  secure: false

# -- Webhook configuration (for validation/conversion)
webhook:
  enabled: true
  port: 9443
  certManager:
    enabled: true  # Use cert-manager for TLS

# -- CRD installation
crd:
  install: true

# =============================================================================
# Locust Test Pod Configuration (what the operator creates)
# =============================================================================

locustPods:
  # -- Default resources for Locust containers
  resources:
    requests:
      cpu: 250m
      memory: 128Mi
    limits:
      cpu: 1000m
      memory: 1024Mi

  # -- Inject affinity/tolerations from CR spec
  affinityInjection: true
  tolerationsInjection: true

  # -- Job TTL after completion (empty = Kubernetes default)
  ttlSecondsAfterFinished: ""

  # -- Metrics exporter sidecar (for v1 API / non-OTel mode)
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

# =============================================================================
# Optional: OTel Collector (for v2 API OTel mode)
# =============================================================================

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

# =============================================================================
# Standard Kubernetes options
# =============================================================================

serviceAccount:
  create: true
  name: ""
  annotations: {}

podAnnotations: {}
nodeSelector: {}
tolerations: []
affinity: {}
```

### 2.2 Backward Compatibility Shims

To support existing users, add a `_helpers.tpl` function that maps old paths to new:

```yaml
# In _helpers.tpl - compatibility mappings
{{- define "locust.podCpuRequest" -}}
{{- .Values.locustPods.resources.requests.cpu | default .Values.config.loadGenerationPods.resource.cpuRequest | default "250m" -}}
{{- end -}}
```

**Old → New Path Mappings:**

| Old Path (Java Chart) | New Path (Go Chart) |
|-----------------------|---------------------|
| `config.loadGenerationPods.resource.cpuRequest` | `locustPods.resources.requests.cpu` |
| `config.loadGenerationPods.resource.memRequest` | `locustPods.resources.requests.memory` |
| `config.loadGenerationPods.affinity.enableCrInjection` | `locustPods.affinityInjection` |
| `config.loadGenerationPods.taintTolerations.enableCrInjection` | `locustPods.tolerationsInjection` |
| `config.loadGenerationJobs.ttlSecondsAfterFinished` | `locustPods.ttlSecondsAfterFinished` |
| `config.loadGenerationPods.metricsExporter.*` | `locustPods.metricsExporter.*` |
| `micronaut.*` | **Removed** (no equivalent) |
| `appPort` | **Removed** (fixed at 8081) |

### 2.3 Removed Sections (No Longer Needed)

```yaml
# DELETE entirely - Java/Micronaut specific
appPort: 8080  # DELETE

micronaut:  # DELETE entire section
  metrics:
    enabled: true
    web:
      enabled: true
    jvm:
      enabled: true
    # ... all of this
```

---

## 3. Clean Deployment Template

### 3.1 Minimal Container Spec

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "locust-k8s-operator.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "locust-k8s-operator.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "locust-k8s-operator.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "locust-k8s-operator.serviceAccountName" . }}
      securityContext:
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: manager
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          command: ["/manager"]
          args:
            - --health-probe-bind-address=:8081
            {{- if .Values.leaderElection.enabled }}
            - --leader-elect=true
            {{- end }}
            {{- if .Values.metrics.enabled }}
            - --metrics-bind-address=:{{ .Values.metrics.port }}
            {{- if .Values.metrics.secure }}
            - --metrics-secure=true
            {{- end }}
            {{- end }}
          ports:
            - name: health
              containerPort: 8081
            {{- if .Values.metrics.enabled }}
            - name: metrics
              containerPort: {{ .Values.metrics.port }}
            {{- end }}
            {{- if .Values.webhook.enabled }}
            - name: webhook
              containerPort: {{ .Values.webhook.port }}
            {{- end }}
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
            readOnlyRootFilesystem: true
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          env:
            {{- include "locust-k8s-operator.envVars" . | nindent 12 }}
          {{- if .Values.webhook.enabled }}
          volumeMounts:
            - name: webhook-certs
              mountPath: /tmp/k8s-webhook-server/serving-certs
              readOnly: true
          {{- end }}
      {{- if .Values.webhook.enabled }}
      volumes:
        - name: webhook-certs
          secret:
            secretName: {{ include "locust-k8s-operator.fullname" . }}-webhook-certs
      {{- end }}
```

### 3.2 Environment Variables Helper

```yaml
# In _helpers.tpl
{{- define "locust-k8s-operator.envVars" -}}
# Locust pod resources
- name: POD_CPU_REQUEST
  value: {{ include "locust.podCpuRequest" . | quote }}
- name: POD_MEM_REQUEST
  value: {{ include "locust.podMemRequest" . | quote }}
- name: POD_CPU_LIMIT
  value: {{ include "locust.podCpuLimit" . | quote }}
- name: POD_MEM_LIMIT
  value: {{ include "locust.podMemLimit" . | quote }}
# Feature flags
- name: ENABLE_AFFINITY_CR_INJECTION
  value: {{ include "locust.affinityInjection" . | quote }}
- name: ENABLE_TAINT_TOLERATIONS_CR_INJECTION
  value: {{ include "locust.tolerationsInjection" . | quote }}
# Metrics exporter
- name: METRICS_EXPORTER_IMAGE
  value: {{ .Values.locustPods.metricsExporter.image | quote }}
- name: METRICS_EXPORTER_PORT
  value: {{ .Values.locustPods.metricsExporter.port | quote }}
- name: METRICS_EXPORTER_IMAGE_PULL_POLICY
  value: {{ .Values.locustPods.metricsExporter.pullPolicy | quote }}
# TTL (optional)
{{- if .Values.locustPods.ttlSecondsAfterFinished }}
- name: JOB_TTL_SECONDS_AFTER_FINISHED
  value: {{ .Values.locustPods.ttlSecondsAfterFinished | quote }}
{{- end }}
{{- end -}}
```

### 3.3 Comparison: Before vs After

**Java Chart (155 lines, complex):**
- 50+ environment variables
- JVM-specific ports
- Kafka integration inline
- Micronaut health probes

**Go Chart (~80 lines, clean):**
- ~15 environment variables
- Fixed health port (8081)
- Conditional metrics/webhook ports
- Standard K8s security context

---

## 5. RBAC Updates

### 5.1 ClusterRole Rules

```yaml
# Current rules + new rules for v2 API
rules:
  # Core resources
  - apiGroups: [""]
    resources: ["configmaps", "secrets", "services", "events"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Jobs
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # LocustTest CRD - both v1 and v2
  - apiGroups: ["locust.io"]
    resources: ["locusttests"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Status subresource
  - apiGroups: ["locust.io"]
    resources: ["locusttests/status"]
    verbs: ["get", "update", "patch"]
  
  # Finalizers
  - apiGroups: ["locust.io"]
    resources: ["locusttests/finalizers"]
    verbs: ["update"]
  
  # Leader election
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  
  # Events
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "patch"]
```

---

## 4. New Template Files

### 4.1 Webhook Configuration

**File:** `templates/webhook.yaml`

```yaml
{{- if .Values.webhook.enabled }}
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-validating
  annotations:
    {{- if .Values.webhook.certManager.enabled }}
    cert-manager.io/inject-ca-from: {{ .Release.Namespace }}/{{ include "locust-k8s-operator.fullname" . }}-serving-cert
    {{- end }}
webhooks:
  - name: vlocusttest.kb.io
    admissionReviewVersions: ["v1"]
    clientConfig:
      service:
        name: {{ include "locust-k8s-operator.fullname" . }}-webhook
        namespace: {{ .Release.Namespace }}
        path: /validate-locust-io-v2-locusttest
    rules:
      - apiGroups: ["locust.io"]
        apiVersions: ["v1", "v2"]
        operations: ["CREATE", "UPDATE"]
        resources: ["locusttests"]
    sideEffects: None
    failurePolicy: Fail
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-webhook
spec:
  ports:
    - port: 443
      targetPort: webhook
  selector:
    {{- include "locust-k8s-operator.selectorLabels" . | nindent 4 }}
{{- end }}
```

### 4.2 Certificate (cert-manager)

**File:** `templates/certificate.yaml`

```yaml
{{- if and .Values.webhook.enabled .Values.webhook.certManager.enabled }}
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-serving-cert
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "locust-k8s-operator.labels" . | nindent 4 }}
spec:
  dnsNames:
    - {{ include "locust-k8s-operator.fullname" . }}-webhook.{{ .Release.Namespace }}.svc
    - {{ include "locust-k8s-operator.fullname" . }}-webhook.{{ .Release.Namespace }}.svc.cluster.local
  issuerRef:
    kind: Issuer
    name: {{ include "locust-k8s-operator.fullname" . }}-selfsigned-issuer
  secretName: {{ include "locust-k8s-operator.fullname" . }}-webhook-certs
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-selfsigned-issuer
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "locust-k8s-operator.labels" . | nindent 4 }}
spec:
  selfSigned: {}
{{- end }}
```

### 6.3 OTel Collector (Optional)

**File:** `templates/otel-collector.yaml`

```yaml
{{- if .Values.otelCollector.enabled }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-otel-collector
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "locust-k8s-operator.labels" . | nindent 4 }}
    app.kubernetes.io/component: otel-collector
spec:
  replicas: {{ .Values.otelCollector.replicas | default 1 }}
  selector:
    matchLabels:
      {{- include "locust-k8s-operator.selectorLabels" . | nindent 6 }}
      app.kubernetes.io/component: otel-collector
  template:
    metadata:
      labels:
        {{- include "locust-k8s-operator.selectorLabels" . | nindent 8 }}
        app.kubernetes.io/component: otel-collector
    spec:
      containers:
        - name: otel-collector
          image: {{ .Values.otelCollector.image }}
          args:
            - --config=/conf/otel-collector-config.yaml
          ports:
            - name: otlp-grpc
              containerPort: 4317
              protocol: TCP
            - name: otlp-http
              containerPort: 4318
              protocol: TCP
          resources:
            {{- toYaml .Values.otelCollector.resources | nindent 12 }}
          volumeMounts:
            - name: config
              mountPath: /conf
      volumes:
        - name: config
          configMap:
            name: {{ include "locust-k8s-operator.fullname" . }}-otel-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-otel-config
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "locust-k8s-operator.labels" . | nindent 4 }}
data:
  otel-collector-config.yaml: |
    {{- .Values.otelCollector.config | nindent 4 }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "locust-k8s-operator.fullname" . }}-otel-collector
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "locust-k8s-operator.labels" . | nindent 4 }}
    app.kubernetes.io/component: otel-collector
spec:
  ports:
    - name: otlp-grpc
      port: 4317
      targetPort: otlp-grpc
      protocol: TCP
    - name: otlp-http
      port: 4318
      targetPort: otlp-http
      protocol: TCP
  selector:
    {{- include "locust-k8s-operator.selectorLabels" . | nindent 4 }}
    app.kubernetes.io/component: otel-collector
{{- end }}
```

---

## 5. Backward Compatibility Strategy

### 5.1 Helper Functions for Old Value Paths

```yaml
# _helpers.tpl - compatibility shims
{{- define "locust.podCpuRequest" -}}
{{- .Values.locustPods.resources.requests.cpu | default .Values.config.loadGenerationPods.resource.cpuRequest | default "250m" -}}
{{- end -}}

{{- define "locust.podMemRequest" -}}
{{- .Values.locustPods.resources.requests.memory | default .Values.config.loadGenerationPods.resource.memRequest | default "128Mi" -}}
{{- end -}}

{{- define "locust.affinityInjection" -}}
{{- .Values.locustPods.affinityInjection | default .Values.config.loadGenerationPods.affinity.enableCrInjection | default true -}}
{{- end -}}

{{- define "locust.tolerationsInjection" -}}
{{- .Values.locustPods.tolerationsInjection | default .Values.config.loadGenerationPods.taintTolerations.enableCrInjection | default true -}}
{{- end -}}
```

### 5.2 What's Preserved vs Removed

| Category | Status | Notes |
|----------|--------|-------|
| `image.*` | **Preserved** | Same structure |
| `resources` | **Preserved** | Same structure |
| `serviceAccount.*` | **Preserved** | Same structure |
| `config.loadGenerationPods.*` | **Deprecated** | Mapped via helpers |
| `config.loadGenerationJobs.*` | **Deprecated** | Mapped via helpers |
| `micronaut.*` | **Removed** | No Go equivalent |
| `appPort` | **Removed** | Fixed at 8081 |
| `livenessProbe`/`readinessProbe` | **Removed** | Hardcoded in template |

---

## 6. File Structure

### 6.1 Final Chart Layout

```
charts/locust-k8s-operator/
├── Chart.yaml                        # v2.0.0
├── values.yaml                       # Clean structure (~100 lines)
├── crds/                             # Symlink to Go operator CRDs
├── templates/
│   ├── _helpers.tpl                  # Include compat shims
│   ├── deployment.yaml               # Rewritten (~80 lines)
│   ├── serviceaccount.yaml           # Simplified
│   ├── clusterrole.yaml              # Updated RBAC
│   ├── clusterrolebinding.yaml
│   ├── webhook.yaml                  # NEW (optional)
│   ├── certificate.yaml              # NEW (optional)
│   └── otel-collector.yaml           # NEW (optional)
└── README.md                         # Updated docs
```

### 6.2 Lines of Code Comparison

| File | Java Chart | Go Chart | Change |
|------|------------|----------|--------|
| `values.yaml` | 144 lines | ~100 lines | -30% |
| `deployment.yaml` | 155 lines | ~80 lines | -48% |
| Total templates | ~300 lines | ~200 lines | -33% |

---

## 7. Testing Strategy

```bash
# Lint
helm lint charts/locust-k8s-operator

# Template generation
helm template test charts/locust-k8s-operator

# With old values (backward compat)
helm template test charts/locust-k8s-operator \
  --set config.loadGenerationPods.resource.cpuRequest=500m

# With new values
helm template test charts/locust-k8s-operator \
  --set locustPods.resources.requests.cpu=500m

# Both should produce same output
```

---

## 8. References

- [controller-runtime Health Probes](https://book.kubebuilder.io/reference/probes.html)
- [cert-manager Webhook Certs](https://cert-manager.io/docs/concepts/webhook/)
- [OTel Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)

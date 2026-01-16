# Locust K8s Operator v2.0 - Requirements Specification

**Document Version:** 1.0  
**Created:** January 2026  
**Status:** Draft  
**Scope:** Go migration + feature enhancements

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Goals & Non-Goals](#2-goals--non-goals)
3. [Technical Requirements](#3-technical-requirements)
4. [API Specification](#4-api-specification)
5. [Feature Requirements](#5-feature-requirements)
6. [Infrastructure Requirements](#6-infrastructure-requirements)
7. [Quality Requirements](#7-quality-requirements)
8. [Constraints & Assumptions](#8-constraints--assumptions)
9. [Success Criteria](#9-success-criteria)
10. [Appendix: Issue Traceability](#10-appendix-issue-traceability)

---

## 1. Executive Summary

### 1.1 Overview

Migrate the Locust Kubernetes Operator from Java (Micronaut + JOSDK) to Go (Operator SDK), while enhancing the operator with new features based on user feedback and modern Locust capabilities.

### 1.2 Key Drivers

- **Ecosystem Alignment:** Go is the standard for Kubernetes operators
- **Resource Efficiency:** ~10x smaller images, ~50-70% less memory
- **User Requests:** Address high-priority feature requests from the community
- **Observability:** Native OpenTelemetry support, replacing legacy sidecar approach

### 1.3 High-Level Scope

| Category | Description |
|----------|-------------|
| **Must Have** | Go migration, OTel support, configmap and secrets injection, volume mounting, status subresource |
| **Should Have** | Separate master/worker resources, configurable commands, OLM distribution |
| **Out of Scope** | Major architectural changes to test execution model |

---

## 2. Goals & Non-Goals

### 2.1 Goals

1. **Complete Go Migration**
   - Rewrite operator using Operator SDK framework
   - Maintain functional parity with Java implementation
   - Improve resource efficiency and startup time

2. **Enhanced Observability**
   - Native OpenTelemetry integration for Locust metrics/traces
   - Remove dependency on legacy `locust_exporter` sidecar
   - Helm chart optionally deploys OTel Collector

3. **Improved Pod Customization**
   - Secrets injection via environment variables and file mounts
   - Arbitrary volume mounting support
   - Separate resource requests for master vs worker pods

4. **Better Operational Experience**
   - Status subresource with test phase and conditions
   - Configurable master command flags (`--autostart`, `--autoquit`)
   - Multi-architecture container images (amd64, arm64)

5. **Dual Distribution**
   - Helm chart (primary)
   - OLM bundle for OperatorHub

### 2.2 Non-Goals

- Changing the immutable test design (NO-OP on updates remains)
- Supporting non-Kubernetes environments
- Building a custom Locust distribution
- Real-time test control via operator (start/stop/pause)

---

## 3. Technical Requirements

### 3.1 Technology Stack

| Component | Requirement |
|-----------|-------------|
| **Language** | Go 1.22+ |
| **Framework** | Operator SDK (latest stable) |
| **Kubernetes Client** | controller-runtime / client-go |
| **Logging** | `slog` (stdlib) or `logr` (controller-runtime default) |
| **Metrics** | Prometheus via controller-runtime |
| **Configuration** | Environment variables with struct tags |

### 3.2 Kubernetes Compatibility

| Requirement | Specification |
|-------------|---------------|
| **Minimum K8s Version** | 1.27+ |
| **Target Platforms** | EKS (primary), GKE, AKS, vanilla K8s |
| **CRD API Version** | `apiextensions.k8s.io/v1` |
| **Admission Webhooks** | Conversion webhook for v1→v2 |

### 3.3 Resilience Requirements

The operator MUST be designed for resilience in dynamic environments:

| Requirement | Description |
|-------------|-------------|
| **Idempotent Reconciliation** | Safe to re-run reconcile at any point |
| **Pod Rescheduling Tolerance** | Handle operator pod restarts gracefully |
| **Spot Instance Compatibility** | Managed pods can be evicted; operator recovers state |
| **Leader Election** | HA deployment with proper leader election |
| **Rate Limiting** | Respect API server rate limits |

### 3.4 Project Structure

```
/
├── cmd/
│   └── manager/
│       └── main.go
├── api/
│   ├── v1/                          # Legacy API (conversion target)
│   │   └── locusttest_types.go
│   └── v2/                          # Primary API
│       ├── locusttest_types.go
│       ├── groupversion_info.go
│       └── zz_generated.deepcopy.go
├── internal/
│   ├── controller/
│   │   ├── locusttest_controller.go
│   │   └── locusttest_controller_test.go
│   ├── resources/
│   │   ├── job.go
│   │   ├── service.go
│   │   └── helpers.go
│   ├── webhook/
│   │   └── conversion.go            # v1 ↔ v2 conversion
│   └── config/
│       └── config.go
├── config/
│   ├── crd/
│   ├── rbac/
│   ├── manager/
│   └── webhook/
├── charts/
│   └── locust-k8s-operator/
└── test/
    └── e2e/
```

---

## 4. API Specification

### 4.1 API Versioning Strategy

| Version | Status | Description |
|---------|--------|-------------|
| `locust.io/v1` | Deprecated | Legacy API, supported via conversion webhook |
| `locust.io/v2` | Current | Primary API with new features |

**Conversion Webhook:** Required to support seamless v1→v2 migration. Users with existing v1 CRs will have them automatically converted.

### 4.2 LocustTest v2 Spec

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: example-test
spec:
  # ============================================
  # IMAGE CONFIGURATION
  # ============================================
  
  # Required: Container image for Locust
  image: "locustio/locust:2.43.1"
  
  # Optional: Image pull configuration
  imagePullPolicy: IfNotPresent    # Always | IfNotPresent | Never
  imagePullSecrets:
    - name: my-registry-secret

  # ============================================
  # MASTER CONFIGURATION (grouped)
  # ============================================
  
  master:
    # Required: Base command for master node
    # Operator appends: --master --master-port=5557 --expect-workers=N
    command: "locust -f /lotest/src/locustfile.py"
    
    # Optional: Resource requests/limits (Issue #246)
    resources:
      requests:
        cpu: "1000m"
        memory: "2Gi"
      limits:
        cpu: "2000m"
        memory: "4Gi"
    
    # Optional: Master-specific labels
    labels:
      team: platform
      role: coordinator
    
    # Optional: Master-specific annotations
    annotations:
      prometheus.io/scrape: "true"
    
    # Optional: Control --autostart flag (Issue #245)
    autostart: true              # default: true
    
    # Optional: Control --autoquit behavior (Issue #245)
    autoquit:
      enabled: true              # default: true
      timeout: 60                # seconds after test completion
    
    # Optional: Additional CLI arguments
    extraArgs:
      - "--only-summary"
      - "--enable-rebalancing"

  # ============================================
  # WORKER CONFIGURATION (grouped)
  # ============================================
  
  worker:
    # Required: Base command for worker nodes
    # Operator appends: --worker --master-host=<service> --master-port=5557
    command: "locust -f /lotest/src/locustfile.py"
    
    # Required: Number of worker pods
    # +kubebuilder:validation:Minimum=1
    # +kubebuilder:validation:Maximum=500
    replicas: 10
    
    # Optional: Resource requests/limits (Issue #246)
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
    
    # Optional: Worker-specific labels
    labels:
      team: platform
      role: load-generator
    
    # Optional: Worker-specific annotations
    annotations: {}
    
    # Optional: Additional CLI arguments
    extraArgs: []

  # ============================================
  # TEST FILES (renamed for clarity)
  # ============================================
  
  testFiles:
    # Optional: ConfigMap containing locustfile(s)
    # Mounted at /lotest/src/ by default
    configMapRef: "my-locust-tests"
    
    # Optional: ConfigMap containing library/helper files
    # Mounted at /opt/locust/lib by default
    libConfigMapRef: "my-locust-lib"
    
    # Optional: Custom mount paths (new)
    srcMountPath: "/lotest/src"
    libMountPath: "/opt/locust/lib"

  # ============================================
  # SCHEDULING (grouped)
  # ============================================
  
  scheduling:
    # Optional: Node affinity rules
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type
                  operator: In
                  values:
                    - performance-testing
    
    # Optional: Tolerations
    tolerations:
      - key: dedicated
        operator: Equal
        value: locust
        effect: NoSchedule
    
    # Optional: Node selector (new)
    nodeSelector:
      kubernetes.io/arch: amd64

  # ============================================
  # ENVIRONMENT INJECTION (Issue #149 - expanded)
  # ============================================
  
  env:
    # Inject all keys from ConfigMaps as env vars
    configMapRefs:
      - name: test-config
        prefix: "TEST_"           # Optional prefix
    
    # Inject all keys from Secrets as env vars
    secretRefs:
      - name: api-credentials
        prefix: ""
    
    # Inject specific values
    variables:
      - name: CUSTOM_VAR
        value: "static-value"
      - name: API_TOKEN
        valueFrom:
          secretKeyRef:
            name: api-credentials
            key: token
      - name: CONFIG_VALUE
        valueFrom:
          configMapKeyRef:
            name: test-config
            key: some-key
    
    # Mount secrets as files
    secretMounts:
      - name: tls-certs
        mountPath: /etc/locust/certs
        readOnly: true

  # ============================================
  # VOLUME MOUNTING (Issue #252)
  # ============================================
  
  volumes:
    # Define additional volumes
    - name: test-results
      persistentVolumeClaim:
        claimName: locust-results-pvc
    - name: shared-data
      emptyDir: {}
  
  volumeMounts:
    # Mount to both master and worker by default
    - name: test-results
      mountPath: /results
    # Or specify target
    - name: shared-data
      mountPath: /shared
      target: worker             # master | worker | both (default)

  # ============================================
  # OBSERVABILITY
  # ============================================
  
  observability:
    # Native OpenTelemetry configuration (recommended)
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: grpc             # grpc | http/protobuf
      insecure: true             # Skip TLS verification
      # Additional OTel SDK environment variables
      extraEnvVars:
        OTEL_RESOURCE_ATTRIBUTES: "service.name=locust-test"
        OTEL_TRACES_SAMPLER: "parentbased_traceidratio"
        OTEL_TRACES_SAMPLER_ARG: "0.1"
```

### 4.3 LocustTest v2 Status

```yaml
status:
  # Simple phase for quick status checks
  phase: Running              # Pending | Running | Succeeded | Failed
  
  # Worker connection tracking
  expectedWorkers: 10
  connectedWorkers: 10
  
  # Timestamps
  startTime: "2026-01-16T21:30:00Z"
  completionTime: null        # Set when test ends
  
  # Test result (set on completion)
  result:
    totalRequests: 150000
    failedRequests: 150
    averageResponseTime: 45   # milliseconds
  
  # Standard Kubernetes conditions
  conditions:
    - type: Ready
      status: "True"
      reason: AllResourcesCreated
      message: "Master and worker jobs are running"
      lastTransitionTime: "2026-01-16T21:30:05Z"
    
    - type: WorkersConnected
      status: "True"
      reason: AllWorkersConnected
      message: "10/10 workers connected to master"
      lastTransitionTime: "2026-01-16T21:30:30Z"
    
    - type: TestCompleted
      status: "False"
      reason: TestInProgress
      message: "Test is currently running"
      lastTransitionTime: "2026-01-16T21:30:00Z"
```

### 4.4 v1 to v2 Field Mapping

| v1 Field | v2 Field | Conversion Notes |
|----------|----------|------------------|
| `masterCommandSeed` | `master.command` | Renamed and grouped |
| `workerCommandSeed` | `worker.command` | Renamed and grouped |
| `workerReplicas` | `worker.replicas` | Renamed and grouped |
| `image` | `image` | Direct mapping |
| `configMap` | `testFiles.configMapRef` | Renamed for clarity |
| `libConfigMap` | `testFiles.libConfigMapRef` | Renamed for clarity |
| `labels.master` | `master.labels` | Moved under master |
| `labels.worker` | `worker.labels` | Moved under worker |
| `annotations.master` | `master.annotations` | Moved under master |
| `annotations.worker` | `worker.annotations` | Moved under worker |
| `affinity` | `scheduling.affinity` | Grouped under scheduling |
| `tolerations` | `scheduling.tolerations` | Grouped under scheduling |
| `imagePullPolicy` | `imagePullPolicy` | Direct mapping |
| `imagePullSecrets` | `imagePullSecrets` | Direct mapping |
| *(n/a)* | `master.resources` | Uses operator defaults |
| *(n/a)* | `worker.resources` | Uses operator defaults |
| *(n/a)* | `master.autostart` | Default: `true` |
| *(n/a)* | `master.autoquit` | Default: `{enabled: true, timeout: 60}` |
| *(n/a)* | `master.extraArgs` | Default: empty |
| *(n/a)* | `worker.extraArgs` | Default: empty |
| *(n/a)* | `testFiles.srcMountPath` | Default: `/lotest/src` |
| *(n/a)* | `testFiles.libMountPath` | Default: `/opt/locust/lib` |
| *(n/a)* | `scheduling.nodeSelector` | Default: empty |
| *(n/a)* | `env` | Default: empty |
| *(n/a)* | `volumes` | Default: empty |
| *(n/a)* | `volumeMounts` | Default: empty |
| *(n/a)* | `observability` | Default: OTel disabled |

---

## 5. Feature Requirements

### 5.1 Must Have (P1)

#### 5.1.1 Native OpenTelemetry Support

**Requirement:** Replace the `locust_exporter` sidecar with native Locust OpenTelemetry integration.

| Aspect | Specification |
|--------|---------------|
| **Locust Flag** | `--otel` added to master command when enabled |
| **Environment Variables** | Inject `OTEL_*` env vars based on spec |
| **Default State** | Enabled when `observability.openTelemetry.enabled: true` |
| **Helm Chart** | Optionally deploy OTel Collector (default: deploy) |
| **Backward Compat** | Users can still use external collectors |

**Helm OTel Collector Options:**
```yaml
# values.yaml
openTelemetry:
  # Option A: Just configure pods (user provides collector)
  collector:
    deploy: false
    externalEndpoint: "otel-collector.monitoring:4317"
  
  # Option B: Deploy collector with chart (default)
  collector:
    deploy: true
    image: otel/opentelemetry-collector:latest
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
    exporters:
      # Forward to user's backend
      otlp:
        endpoint: ""
      # Or use Prometheus
      prometheus:
        enabled: true
        port: 9090
```

**Acceptance Criteria:**
- [ ] Locust pods start with `--otel` flag when enabled
- [ ] OTel environment variables correctly injected
- [ ] Helm chart can deploy OTel Collector
- [ ] Metrics/traces flow to configured backend
- [ ] No `locust_exporter` sidecar when OTel enabled

#### 5.1.2 Environment Injection - ConfigMaps & Secrets (Issue #149)

**Requirement:** Allow users to inject ConfigMaps and Secrets into Locust pods.

| Aspect | Specification |
|--------|---------------|
| **ConfigMap Env Vars** | `env.configMapRefs` injects all keys from ConfigMaps |
| **Secret Env Vars** | `env.secretRefs` injects all keys from Secrets |
| **Individual Variables** | `env.variables` for specific key injection |
| **File Mounts** | `env.secretMounts` mounts secrets as files |
| **Scope** | Apply to both master and worker pods |
| **External Secrets** | Works with External Secrets Operator, Vault, etc. |

**Acceptance Criteria:**
- [ ] ConfigMap values available as env vars in pods
- [ ] Secrets available as env vars in pods
- [ ] Secrets mountable as files via `env.secretMounts`
- [ ] Works with external secret operators
- [ ] No credentials in test ConfigMaps required

#### 5.1.3 Volume Mounting (Issue #252)

**Requirement:** Allow users to mount arbitrary volumes to Locust pods.

| Aspect | Specification |
|--------|---------------|
| **Supported Types** | PVC, emptyDir, configMap, secret, projected, hostPath |
| **Target Selection** | Mount to master, worker, or both |
| **Conflict Prevention** | Validate no conflicts with operator-managed mounts |

**Acceptance Criteria:**
- [ ] Users can define volumes in CR spec
- [ ] Volumes mounted to correct pods
- [ ] No conflicts with `/lotest/src/` and `/opt/locust/lib`
- [ ] PVC mounting works for result collection

#### 5.1.4 Status Subresource

**Requirement:** Add status subresource to track test state.

| Aspect | Specification |
|--------|---------------|
| **Phase** | `Pending`, `Running`, `Succeeded`, `Failed` |
| **Worker Tracking** | `expectedWorkers`, `connectedWorkers` |
| **Timestamps** | `startTime`, `completionTime` |
| **Conditions** | `Ready`, `WorkersConnected`, `TestCompleted` |

**Acceptance Criteria:**
- [ ] Phase accurately reflects test state
- [ ] Worker count updated in real-time
- [ ] Conditions follow K8s conventions
- [ ] `kubectl get locusttests` shows phase column

#### 5.1.5 API v1 Conversion Webhook

**Requirement:** Support v1 CRs via conversion webhook.

| Aspect | Specification |
|--------|---------------|
| **Direction** | Bidirectional (v1 ↔ v2) |
| **Defaults** | v1→v2 uses sensible defaults for new fields |
| **Deprecation** | v1 marked deprecated, removed in v3.0 |

**Acceptance Criteria:**
- [ ] Existing v1 CRs work without modification
- [ ] v1 CRs converted to v2 on read
- [ ] v2 CRs can be stored as v1 (lossy for new fields)
- [ ] Clear deprecation warnings in logs

### 5.2 Should Have (P2)

#### 5.2.1 Separate Master/Worker Resources (Issue #246)

**Requirement:** Allow different resource requests/limits for master vs worker.

| Aspect | Specification |
|--------|---------------|
| **CR-Level** | `master.resources` and `worker.resources` |
| **Helm-Level** | Default values in Helm chart |
| **Precedence** | CR overrides Helm defaults |

#### 5.2.2 Configurable Master Commands (Issue #245)

**Requirement:** Make `--autostart` and `--autoquit` configurable.

| Aspect | Specification |
|--------|---------------|
| **Autostart** | Enable/disable via `master.autostart` |
| **Autoquit** | Enable/disable + timeout via `master.autoquit` |
| **Extra Args** | `master.extraArgs` and `worker.extraArgs` for additional arguments |

#### 5.2.3 OLM Bundle Distribution

**Requirement:** Generate OLM bundle for OperatorHub distribution.

| Aspect | Specification |
|--------|---------------|
| **Bundle Format** | OLM bundle format |
| **Channels** | `stable`, `alpha` |
| **CSV Generation** | Auto-generated from code annotations |

### 5.3 Nice to Have (P3)

#### 5.3.1 JSON/HTML Output Support

```yaml
spec:
  output:
    jsonFile:
      enabled: true
      path: "/results/output.json"
    htmlReport:
      enabled: true
      path: "/results/report.html"
```

#### 5.3.2 Processes Per Pod

```yaml
spec:
  worker:
    processesPerPod: 4  # Run multiple Locust processes per pod
```

#### 5.3.3 Web UI Configuration

```yaml
spec:
  webUI:
    basePath: "/locust"  # For ingress routing
```

---

## 6. Infrastructure Requirements

### 6.1 Container Images

| Requirement | Specification |
|-------------|---------------|
| **Architectures** | `linux/amd64`, `linux/arm64` |
| **Base Image** | `gcr.io/distroless/static:nonroot` or `scratch` |
| **Image Size Target** | < 50MB |
| **Registry** | Docker Hub, GitHub Container Registry |
| **Tagging** | SemVer (`v2.0.0`), `latest`, SHA |

### 6.2 Helm Chart

| Requirement | Specification |
|-------------|---------------|
| **Chart Version** | Follows operator version |
| **Values Schema** | JSON Schema validation |
| **OTel Collector** | Optional sub-chart or embedded |
| **RBAC** | Auto-generated, minimal permissions |
| **CRD Management** | Helm manages CRD lifecycle |

### 6.3 CI/CD Pipeline

| Stage | Tools |
|-------|-------|
| **Lint** | `golangci-lint`, `helm lint` |
| **Unit Tests** | `go test` |
| **Integration Tests** | `envtest` |
| **E2E Tests** | Kind cluster |
| **Build** | `ko` or multi-stage Docker |
| **Release** | GoReleaser, Helm Chart Releaser |

---

## 7. Quality Requirements

### 7.1 Testing Requirements

| Test Type | Tool | Coverage Target |
|-----------|------|-----------------|
| **Unit Tests** | `go test` + `testify` | 80% |
| **Controller Tests** | `envtest` | All reconcile paths |
| **E2E Tests** | Kind + Ginkgo | Critical user journeys |
| **Webhook Tests** | `envtest` | Conversion accuracy |

### 7.2 Code Quality

| Aspect | Requirement |
|--------|-------------|
| **Linting** | `golangci-lint` with strict config |
| **Formatting** | `gofmt` / `goimports` |
| **Documentation** | GoDoc for public APIs |
| **Commit Messages** | Conventional Commits |

### 7.3 Security

| Aspect | Requirement |
|--------|-------------|
| **RBAC** | Minimal required permissions |
| **Pod Security** | Run as non-root |
| **Image Scanning** | Trivy in CI |
| **Secrets** | Never logged, never in plain text |

---

## 8. Constraints & Assumptions

### 8.1 Constraints

| Constraint | Description |
|------------|-------------|
| **Timeline** | Best-effort, spare time development |
| **Breaking Changes** | Acceptable with conversion webhook support |
| **Backward Compat** | v1 API supported for 2 minor versions |
| **EKS Compatibility** | Must work on EKS without issues |

### 8.2 Assumptions

| Assumption | Rationale |
|------------|-----------|
| Users have OTel infrastructure | Modern observability standard |
| Kubernetes 1.27+ | Required for stable features |
| Users can update Helm values | Migration path for new features |
| Locust 2.40+ images | OTel support requires recent Locust |

### 8.3 Design Principles

| Principle | Application |
|-----------|-------------|
| **Immutable Tests** | NO-OP on CR updates (delete + recreate for changes) |
| **Idempotent Reconciliation** | Safe re-execution at any point |
| **Fail-Safe Defaults** | Missing config uses safe defaults |
| **Minimal RBAC** | Only request necessary permissions |

---

## 9. Success Criteria

### 9.1 Functional Success

- [ ] All v1 functionality works in Go implementation
- [ ] v1 CRs work via conversion webhook
- [ ] OTel metrics/traces flow correctly
- [ ] Secrets injection works with External Secrets Operator
- [ ] Volume mounting enables result collection
- [ ] Status reflects accurate test phase

### 9.2 Operational Success

- [ ] Operator starts in < 5 seconds
- [ ] Memory usage < 64MB under normal load
- [ ] Handles 100+ concurrent LocustTest CRs
- [ ] Survives operator pod restarts
- [ ] Works on EKS with spot instances

### 9.3 Distribution Success

- [ ] Helm chart installable via `helm install`
- [ ] OLM bundle published to OperatorHub
- [ ] Multi-arch images available
- [ ] Documentation updated for v2

---

## 10. Appendix: Issue Traceability

| Issue | Title | Priority | Requirement Section |
|-------|-------|----------|---------------------|
| #72 | Metrics exporter sidecar management | P1 | 5.1.1 (solved by OTel) |
| #149 | Secrets injection | P1 | 5.1.2 |
| #252 | Volume mounting | P1 | 5.1.3 |
| #246 | Separate master/worker resources | P2 | 5.2.1 |
| #245 | Configurable master commands | P2 | 5.2.2 |
| #50 | Automatic test resolution | P2 | 5.1.1 (duplicate of #72) |
| #118 | Metrics documentation | P3 | Documentation |
| #253 | Connection delay | P3 | Investigation needed |
| #254 | Metrics clarification | P3 | Documentation |

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01 | - | Initial requirements specification |

---

**Companion Documents:**
- `analysis/ASSESSMENT.md` - Technical viability assessment
- `analysis/TECHNICAL.md` - Detailed migration guide
- `analysis/LOCUST_FEATURES.md` - Locust feature analysis

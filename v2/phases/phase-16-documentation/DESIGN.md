# Phase 16: Documentation - Design Document

**Version:** 1.0  
**Status:** ✅ Completed

---

## 1. Documentation Architecture

### 1.1 Current Structure

```
docs/
├── index.md                    # Landing page
├── getting_started.md          # Quick start guide
├── features.md                 # Feature overview
├── helm_deploy.md              # Helm installation
├── advanced_topics.md          # Kafka, affinity, tolerations, resources
├── metrics_and_dashboards.md   # Prometheus, Grafana
├── how_does_it_work.md         # Architecture overview
├── contribute.md               # Contribution guide
├── local-development.md        # Dev setup
├── integration-testing.md      # Testing guide
├── pull-request-process.md     # PR workflow
├── license.md                  # License info
├── roadmap.md                  # Future plans
└── assets/                     # Images, styles
```

### 1.2 Proposed Structure (v2.0)

```
docs/
├── index.md                    # Updated with v2 highlights
├── getting_started.md          # Updated for v2 API
├── features.md                 # Updated with new features
├── helm_deploy.md              # Updated Helm values
├── advanced_topics.md          # Updated + new sections
├── metrics_and_dashboards.md   # Updated + OTel section
├── migration.md                # NEW: v1→v2 migration guide
├── api_reference.md            # NEW: Complete API documentation
├── how_does_it_work.md         # Updated for Go architecture
├── contribute.md               # Minor updates if needed
├── local-development.md        # Updated for Go development
├── integration-testing.md      # Updated for Go testing
├── pull-request-process.md     # No changes
├── license.md                  # No changes
├── roadmap.md                  # Updated with v2 completion
└── assets/                     # New images if needed
```

---

## 2. Content Guidelines

### 2.1 Writing Style

- **Concise:** Get to the point quickly
- **Actionable:** Focus on what users need to do
- **Example-driven:** Include working code/YAML examples
- **Consistent:** Use same terminology throughout
- **Versioned:** Clearly indicate v1 vs v2 when relevant

### 2.2 Terminology

| Term | Usage |
|------|-------|
| v1 API | `locust.io/v1` - legacy, deprecated |
| v2 API | `locust.io/v2` - current, recommended |
| Go operator | The new operator implementation |
| Java operator | The legacy operator (pre-v2.0) |
| LocustTest | The custom resource kind |
| CR | Custom Resource (instance of LocustTest) |

### 2.3 Code Example Standards

**YAML Examples:**
```yaml
# Always include:
# 1. Full apiVersion and kind
# 2. Meaningful metadata.name
# 3. Comments for non-obvious fields
# 4. Working, tested examples

apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: example-test  # Use descriptive names
spec:
  image: locustio/locust:2.20.0  # Always specify version
  master:
    command: "locust -f /lotest/src/locustfile.py --host https://example.com"
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 3  # Comment explaining purpose
```

**Shell Commands:**
```bash
# Always prefix with explanation
# Use realistic examples
kubectl apply -f locusttest.yaml
kubectl get locusttests
kubectl describe locusttest example-test
```

---

## 3. Page-by-Page Design

### 3.1 README.md

**Purpose:** Repository landing page, first impression for GitHub visitors

**Structure:**
```markdown
<logo and title>

<tagline>

<badges>

## 🎉 v2.0 - Go Rewrite

Brief announcement of major release with key benefits:
- Performance improvements
- New features summary
- Link to migration guide

## At a Glance

Existing content (keep)

## Documentation

Link to gh-pages

## Project Status

Update to reflect v2.0 release

## Contribute / License

Keep existing
```

**Design Decisions:**
- Keep visual consistency with current README
- Add prominent v2.0 section without being disruptive
- Link to detailed docs rather than duplicating content

---

### 3.2 docs/index.md (Landing Page)

**Purpose:** Documentation home, showcase capabilities

**Structure:**
```markdown
# Performance testing that simply works

<hero section - keep existing>

## 🚀 What's New in v2.0 (NEW SECTION)

Grid cards highlighting:
- Go rewrite benefits
- OpenTelemetry support
- Secret injection
- Volume mounting

## Build for cloud-native performance testing

<existing content>

## Designed for teams and organizations

<existing content>
```

**Design Decisions:**
- Insert v2.0 section after hero, before existing content
- Use Material for MkDocs grid cards for visual consistency
- Don't remove existing content - augment it

---

### 3.3 docs/migration.md (NEW)

**Purpose:** Guide existing v1 users to v2

**Structure:**
```markdown
# Migration Guide: v1 to v2

## Overview
- Why we rewrote in Go
- What changes for users
- Compatibility guarantees

## Before You Begin
- Prerequisites
- Backup recommendations

## Step 1: Update Helm Chart
- Commands to upgrade
- New values to consider

## Step 2: Verify Existing CRs
- Conversion webhook handles v1→v2
- How to verify conversion works

## Step 3: Migrate CRs to v2 Format (Recommended)
- Field mapping table
- Example transformations

## Step 4: Leverage New Features
- Quick overview of new capabilities
- Links to detailed docs

## Troubleshooting
- Common issues
- How to get help

## Rollback Procedure
- Steps to revert if needed
```

**Design Decisions:**
- Assume reader is existing v1 user
- Provide both "quick path" (just upgrade Helm) and "full path" (migrate CRs)
- Include complete field mapping reference
- Offer rollback procedure for safety

---

### 3.4 docs/api_reference.md (NEW)

**Purpose:** Complete API documentation

**Structure:**
```markdown
# API Reference

## Overview
- API versions
- Group: locust.io

## LocustTest v2 (Recommended)

### Spec Fields

#### Root Fields
| Field | Type | Required | Description |
|-------|------|----------|-------------|

#### MasterSpec
| Field | Type | Required | Description |
|-------|------|----------|-------------|

#### WorkerSpec
...

#### EnvConfig
...

### Status Fields
...

### Complete Example
...

## LocustTest v1 (Deprecated)

⚠️ v1 API is deprecated. Use v2 for new deployments.

[Link to migration guide]

### Spec Fields (v1)
Brief reference for legacy users
```

**Design Decisions:**
- v2 first, prominently featured
- v1 documented but clearly marked deprecated
- Table format for quick scanning
- Complete examples for each major section

---

### 3.5 docs/features.md Updates

**New Feature Cards to Add:**

```markdown
-   :material-eye-check: **Native OpenTelemetry Support**

    ---

    Export traces and metrics directly from Locust using native OpenTelemetry 
    integration. No sidecar required - configure endpoints, protocols, and 
    custom attributes directly in your CR.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#opentelemetry)

-   :material-key-variant: **Secret & ConfigMap Injection**

    ---

    Securely inject credentials, API keys, and configuration from Kubernetes 
    Secrets and ConfigMaps. Supports environment variables and file mounts 
    with automatic prefix handling.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#environment-injection)

-   :material-harddisk: **Flexible Volume Mounting**

    ---

    Mount test data, certificates, and configuration files from PersistentVolumes, 
    ConfigMaps, or Secrets. Target specific components (master, worker, or both) 
    with fine-grained control.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#volume-mounting)

-   :material-tune-vertical: **Separate Resource Specs**

    ---

    Configure resources, labels, and annotations independently for master and 
    worker pods. Optimize each component based on its specific needs.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#separate-resources)

-   :material-list-status: **Enhanced Status Tracking**

    ---

    Monitor test progress with rich status information including phase 
    (Pending, Running, Succeeded, Failed), Kubernetes conditions, and 
    worker connection status.

    [:octicons-arrow-right-24: Learn more](api_reference.md#status-fields)
```

---

### 3.6 docs/advanced_topics.md Updates

**New Sections:**

#### OpenTelemetry Section
```markdown
## :material-chart-timeline: OpenTelemetry Integration

Locust 2.x supports native OpenTelemetry for exporting traces and metrics.
The operator can configure this automatically.

### Enabling OpenTelemetry

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: otel-enabled-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "locust -f /lotest/src/test.py"
  worker:
    command: "locust -f /lotest/src/test.py"
    replicas: 5
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"  # or "http"
      extraEnvVars:
        - name: OTEL_SERVICE_NAME
          value: "my-load-test"
        - name: OTEL_RESOURCE_ATTRIBUTES
          value: "environment=staging,team=platform"
```

### OTel vs Metrics Sidecar

| Aspect | OpenTelemetry | Metrics Sidecar |
|--------|---------------|-----------------|
| Setup complexity | Low | Low |
| Traces | ✅ Yes | ❌ No |
| Metrics | ✅ Yes | ✅ Yes |
| Additional pods | None | None (sidecar) |
| Recommended for | New deployments | Legacy compatibility |
```

#### Environment Injection Section
```markdown
## :material-key: Environment & Secret Injection

Inject configuration and credentials into Locust pods without hardcoding.

### ConfigMap Environment Variables

```yaml
spec:
  env:
    configMapRefs:
      - name: app-config
        prefix: "APP_"  # Results in APP_KEY1, APP_KEY2, etc.
```

### Secret Environment Variables

```yaml
spec:
  env:
    secretRefs:
      - name: api-credentials
        prefix: ""  # No prefix, use key names directly
```

### Individual Variables

```yaml
spec:
  env:
    variables:
      - name: TARGET_HOST
        value: "https://api.example.com"
      - name: API_TOKEN
        valueFrom:
          secretKeyRef:
            name: api-secret
            key: token
```

### Secret File Mounts

```yaml
spec:
  env:
    secretMounts:
      - name: tls-certs
        mountPath: /etc/locust/certs
        readOnly: true
```

### Reserved Paths

The following paths are reserved and cannot be used for secret mounts:
- `/lotest/src/` - Test script mount point
- `/opt/locust/lib` - Library mount point
```

#### Volume Mounting Section
```markdown
## :material-folder-multiple: Volume Mounting

Mount arbitrary volumes to Locust pods for test data or configuration.

### Basic Volume Mount

```yaml
spec:
  volumes:
    - name: test-data
      persistentVolumeClaim:
        claimName: test-data-pvc
  volumeMounts:
    - name: test-data
      mountPath: /data
      target: both  # master, worker, or both
```

### Target Filtering

Control which pods receive the volume mount:

| Target | Master | Worker |
|--------|--------|--------|
| `master` | ✅ | ❌ |
| `worker` | ❌ | ✅ |
| `both` | ✅ | ✅ |
```

---

### 3.7 docs/helm_deploy.md Updates

**Changes:**

1. **Remove JVM-specific settings:**
   - `JAVA_OPTS`
   - JVM heap settings
   - Micronaut configuration

2. **Add Go operator settings:**
   - Lower memory defaults (~64Mi)
   - Webhook configuration
   - Leader election settings

3. **Update values table:**

```markdown
### Webhook Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `webhook.enabled` | Enable conversion webhook | `true` |
| `webhook.port` | Webhook server port | `9443` |
| `webhook.certDir` | Certificate directory | `/tmp/k8s-webhook-server/serving-certs` |

### Resource Defaults

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.memory` | Operator memory limit | `64Mi` |
| `resources.limits.cpu` | Operator CPU limit | `100m` |
| `resources.requests.memory` | Operator memory request | `32Mi` |
| `resources.requests.cpu` | Operator CPU request | `10m` |
```

---

### 3.8 docs/metrics_and_dashboards.md Updates

**Add OpenTelemetry section:**

```markdown
## :material-chart-timeline: OpenTelemetry Metrics & Traces

### Native OpenTelemetry Support

Locust 2.x includes native OpenTelemetry support, which the operator 
can configure automatically. This provides both metrics and distributed 
tracing without requiring the metrics exporter sidecar.

### Configuring OTel

See [Advanced Topics - OpenTelemetry](advanced_topics.md#opentelemetry) 
for detailed configuration options.

### OTel Collector Setup

For a complete observability setup, deploy an OTel Collector:

```yaml
# Example OTel Collector config
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  prometheus:
    endpoint: 0.0.0.0:8889
  jaeger:
    endpoint: jaeger-collector:14250

service:
  pipelines:
    metrics:
      receivers: [otlp]
      exporters: [prometheus]
    traces:
      receivers: [otlp]
      exporters: [jaeger]
```
```

**Update Operator Metrics section:**

```markdown
## :material-robot-outline: Operator Metrics

The Go operator exposes controller-runtime metrics on port 8080:

- `controller_runtime_reconcile_total` - Total reconciliations
- `controller_runtime_reconcile_errors_total` - Reconciliation errors
- `controller_runtime_reconcile_time_seconds` - Reconciliation duration
- `workqueue_depth` - Current queue depth
- `workqueue_adds_total` - Items added to queue

These metrics can be scraped by Prometheus using the standard 
`/metrics` endpoint on the operator pod.
```

---

## 4. Visual Assets

### 4.1 Required Updates

- [ ] Review existing images for Java references
- [ ] Consider adding v2 architecture diagram if significantly different
- [ ] No logo changes needed

### 4.2 Optional Enhancements

- Animated GIF showing status transitions (Pending→Running→Succeeded)
- Comparison chart (v1 vs v2 features)
- Architecture diagram update for OTel flow

---

## 5. mkdocs.yml Updates

```yaml
nav:
  - Home: index.md
  - Getting Started: getting_started.md
  - Features: features.md
  - Helm Deployment: helm_deploy.md
  - Advanced Topics: advanced_topics.md
  - Metrics & Dashboards: metrics_and_dashboards.md
  - API Reference: api_reference.md        # NEW
  - Migration Guide: migration.md          # NEW
  - How Does It Work: how_does_it_work.md
  - Roadmap: roadmap.md
  - Contributing:
    - How to Contribute: contribute.md
    - Local Development: local-development.md
    - Integration Testing: integration-testing.md
    - Pull Request Process: pull-request-process.md
  - License: license.md
```

---

## 6. Quality Checklist

### Before Merge

- [ ] All examples validate against v2 CRD
- [ ] No broken internal links
- [ ] No broken external links
- [ ] Consistent terminology throughout
- [ ] No Java/JVM references in operator context
- [ ] mkdocs build --strict passes
- [ ] Reviewed by at least one other person

### Post-Merge

- [ ] Verify gh-pages deployment
- [ ] Test all quick start commands
- [ ] Verify search functionality works
- [ ] Check mobile rendering

---

## 7. Maintenance Notes

### Future Documentation Updates

When adding new features, update:
1. `features.md` - Add feature card
2. `advanced_topics.md` - Add detailed section
3. `api_reference.md` - Document new fields
4. `CHANGELOG.md` - Note the change

### Version-Specific Content

Use admonitions for version-specific notes:

```markdown
!!! info "New in v2.0"
    This feature is only available in the v2 API.

!!! warning "Deprecated"
    This field is deprecated and will be removed in v3.0.
```

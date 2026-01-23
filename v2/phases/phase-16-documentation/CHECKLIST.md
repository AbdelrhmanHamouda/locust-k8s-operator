# Phase 16: Documentation - Checklist

**Estimated Effort:** 1 day  
**Status:** ✅ Completed

---

## Pre-Implementation

- [x] Phase 15 complete (E2E tests passing)
- [x] All v2 features implemented and tested (Phases 7-12)
- [x] Review current documentation structure in `docs/`
- [x] Review mkdocs.yml for navigation structure
- [x] Run `mkdocs serve` to verify current state
- [x] Gather all new feature specs from completed phases

---

## Task 16.1: Update README.md

**File:** `README.md`

### Changes Required

- [x] Add v2.0 release announcement banner/section
- [x] Add "What's New in v2.0" section highlighting:
  - [x] Go rewrite (from Java/Micronaut)
  - [x] Performance improvements (memory, startup time)
  - [x] New features list with brief descriptions
- [x] Add link to migration guide in docs
- [x] Keep existing look and feel (badges, structure)
- [x] Update any Java-specific references if present
- [x] Ensure all links are valid

### Content Structure

```markdown
## 🎉 v2.0 Release

The Locust Kubernetes Operator has been completely rewritten in Go!

### Why the Rewrite?
- **Smaller footprint**: ~64MB vs ~256MB memory
- **Faster startup**: <1s vs 3-5s
- **Better ecosystem alignment**: Native controller-runtime
- **Enhanced features**: OTel, secrets, volumes

[Migration Guide →](https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/)
```

**Verification:**
- [x] README renders correctly on GitHub
- [x] All badges still work
- [x] Links are valid

---

## Task 16.2: Update docs/getting_started.md

**File:** `docs/getting_started.md`

### Changes Required

- [x] Update API version examples from `locust.io/v1` to `locust.io/v2`
- [x] Add note about v1 backward compatibility
- [x] Update CR examples with v2 spec structure:
  - [x] `master.command` instead of `masterCommandSeed`
  - [x] `worker.command` instead of `workerCommandSeed`
  - [x] `worker.replicas` instead of `workerReplicas`
  - [x] `testFiles.configMapRef` instead of `configMap`
- [x] Add section on new v2 features (brief overview with links)
- [x] Update Helm installation notes if chart version changes
- [x] Verify all kubectl commands still work

### v2 CR Example

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: demo-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "locust -f /lotest/src/demo_test.py --host https://example.com"
    autostart: true
    autoquit: 10
  worker:
    command: "locust -f /lotest/src/demo_test.py"
    replicas: 3
  testFiles:
    configMapRef: demo-test-map
```

**Verification:**
- [x] Examples are valid v2 CRs
- [x] Commands work with new operator

---

## Task 16.3: Update docs/features.md

**File:** `docs/features.md`

### New Features to Add

- [x] **Native OpenTelemetry Support**
  - Direct tracing and metrics export
  - No sidecar required when enabled
  - Configurable endpoints and protocols
  
- [x] **Secret & ConfigMap Injection**
  - Environment variables from Secrets
  - Environment variables from ConfigMaps
  - Prefix support for namespacing
  - File mounting for credentials
  
- [x] **Flexible Volume Mounting**
  - Mount arbitrary volumes
  - Target filtering (master/worker/both)
  - Support for PVC, ConfigMap, Secret volumes
  
- [x] **Separate Resource Specifications**
  - Independent master/worker resource limits
  - Per-component labels and annotations
  - Granular control over scheduling
  
- [x] **Enhanced Status Tracking**
  - Phase indicators (Pending, Running, Succeeded, Failed)
  - Kubernetes conditions
  - Worker connection status

### Feature Card Template

```markdown
-   :material-eye-check: **Native OpenTelemetry Support**

    ---

    Export traces and metrics directly from Locust using native OpenTelemetry integration. 
    Configure exporters, endpoints, and protocols without sidecars.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#opentelemetry)
```

**Verification:**
- [x] All feature cards render correctly
- [x] Links to detailed sections work

---

## Task 16.4: Create docs/migration.md

**File:** `docs/migration.md` (NEW)

### Content Sections

- [x] **Overview**
  - What's changing and why
  - Backward compatibility guarantees
  
- [x] **API Changes**
  - Field mapping table (v1 → v2)
  - Deprecation warnings for v1
  - Conversion webhook behavior
  
- [x] **Breaking Changes**
  - List any breaking changes (if any)
  - Workarounds or migration steps
  
- [x] **Step-by-Step Migration**
  1. Update Helm chart
  2. Verify existing CRs still work (v1 conversion)
  3. Migrate CRs to v2 format (recommended)
  4. Test new features
  
- [x] **Field Mapping Reference**

| v1 Field | v2 Field | Notes |
|----------|----------|-------|
| `masterCommandSeed` | `master.command` | Renamed |
| `workerCommandSeed` | `worker.command` | Renamed |
| `workerReplicas` | `worker.replicas` | Moved |
| `configMap` | `testFiles.configMapRef` | Grouped |
| `libConfigMap` | `testFiles.libConfigMapRef` | Grouped |
| `labels.master` | `master.labels` | Restructured |
| `annotations.worker` | `worker.annotations` | Restructured |
| `affinity` | `scheduling.affinity` | Grouped |
| `tolerations` | `scheduling.tolerations` | Grouped |
| - | `env` | **New in v2** |
| - | `volumes` | **New in v2** |
| - | `observability` | **New in v2** |
| - | `status` | **New in v2** |

- [x] **Rollback Procedure**
  - How to revert to v1 operator if needed

**Verification:**
- [x] Migration steps are accurate
- [x] Field mapping is complete

---

## Task 16.5: Update API Reference

**File:** `docs/api_reference.md` (NEW or update existing)

### Content

- [x] Document v2 LocustTestSpec fields
  - [x] `image`, `imagePullPolicy`, `imagePullSecrets`
  - [x] `master` (MasterSpec)
  - [x] `worker` (WorkerSpec)
  - [x] `testFiles` (TestFilesConfig)
  - [x] `scheduling` (SchedulingConfig)
  - [x] `env` (EnvConfig)
  - [x] `volumes` and `volumeMounts`
  - [x] `observability` (ObservabilityConfig)
  
- [x] Document LocustTestStatus fields
  - [x] `phase`
  - [x] `conditions`
  - [x] `masterJob`, `workerJob`
  - [x] `observedGeneration`
  
- [x] Add v1 deprecation notice with link to migration guide
- [x] Include complete CR examples for common use cases

**Verification:**
- [x] All fields documented
- [x] Examples validate against CRD

---

## Task 16.6: Ensure docs/ Structure is Complete

**Verify Files:**

- [x] `docs/index.md` - Landing page with new features highlight
- [x] `docs/getting_started.md` - Updated quick start
- [x] `docs/features.md` - All features including new ones
- [x] `docs/helm_deploy.md` - Updated Helm values
- [x] `docs/advanced_topics.md` - New configuration options
- [x] `docs/metrics_and_dashboards.md` - OTel section added
- [x] `docs/migration.md` - New migration guide
- [x] `docs/api_reference.md` - v2 API documentation
- [x] `docs/contribute.md` - No changes needed (verify)
- [x] `docs/roadmap.md` - Update with v2 completion

**Navigation (mkdocs.yml):**

- [x] Add migration.md to nav
- [x] Add api_reference.md to nav
- [x] Verify nav order makes sense

---

## Task 16.7: Update Helm Values Documentation

**File:** `docs/helm_deploy.md`

### Changes Required

- [x] Update image repository (if changed)
- [x] Remove JVM-specific configurations:
  - [x] `JAVA_OPTS`
  - [x] JVM memory settings
  - [x] Micronaut-specific configs
- [x] Add Go operator configurations:
  - [x] Memory limits (~64Mi default)
  - [x] Leader election settings
  - [x] Webhook configuration
- [x] Update resource defaults table
- [x] Add OTel Collector deployment option (if applicable)
- [x] Update RBAC description if permissions changed

### New Values Table Additions

| Parameter | Description | Default |
|-----------|-------------|---------|
| `webhook.enabled` | Enable conversion webhook | `true` |
| `webhook.port` | Webhook server port | `9443` |
| `otel.collector.enabled` | Deploy OTel collector | `false` |
| `otel.collector.endpoint` | OTel collector endpoint | `""` |

**Verification:**
- [x] Values match actual chart
- [x] Examples work

---

## Task 16.8: Update Observability Documentation

**File:** `docs/metrics_and_dashboards.md`

### Changes Required

- [x] Add **OpenTelemetry** section:
  - [x] How to enable OTel in CR
  - [x] Configuring OTel endpoint
  - [x] Example OTel Collector setup
  - [x] Trace correlation
  
- [x] Update Prometheus metrics section:
  - [x] Sidecar vs OTel comparison
  - [x] When to use each approach
  
- [x] Add Go operator metrics:
  - [x] controller-runtime metrics
  - [x] Reconciliation metrics
  - [x] Remove Micronaut/JVM metrics references

### OTel Example

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: otel-enabled-test
spec:
  # ... other fields
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"
      extraEnvVars:
        - name: OTEL_SERVICE_NAME
          value: "locust-load-test"
```

**Verification:**
- [x] OTel example is valid
- [x] Instructions are accurate

---

## Task 16.9: Document New Feature Workflows

**File:** `docs/advanced_topics.md`

### New Sections to Add

- [x] **Environment & Secret Injection**
  - Full example with all options
  - Best practices for credentials
  - Integration with External Secrets Operator
  
- [x] **Volume Mounting**
  - Use cases (test data, certificates)
  - Target filtering explanation
  - Reserved path restrictions
  
- [x] **Separate Master/Worker Resources**
  - When to use different resources
  - Example configurations
  
- [x] **Status Monitoring**
  - Using `kubectl get locusttests`
  - Interpreting conditions
  - Scripting with status fields

**Verification:**
- [x] Examples are complete and valid
- [x] Best practices are accurate

---

## Task 16.10: Update docs/index.md Landing Page

**File:** `docs/index.md`

### Changes Required

- [x] Add v2.0 release banner/callout
- [x] Update hero section with new capabilities
- [x] Add "What's New" feature cards:
  - [x] Go rewrite benefits
  - [x] OpenTelemetry
  - [x] Secret injection
  - [x] Volume mounting
- [x] Add migration CTA for existing users
- [x] Update any statistics if applicable

### New Section

```markdown
## 🚀 What's New in v2.0

<div class="grid cards" markdown>

-   :material-language-go:{ .lg .middle } __Rewritten in Go__

    ---

    Faster startup, lower memory footprint, and better Kubernetes integration 
    with native controller-runtime.

-   :material-eye-check:{ .lg .middle } __OpenTelemetry Native__

    ---

    Export traces and metrics directly without sidecars using Locust's 
    native OTel support.

-   :material-key:{ .lg .middle } __Secret Injection__

    ---

    Securely inject credentials from Kubernetes Secrets into your load tests.

-   :material-harddisk:{ .lg .middle } __Volume Mounting__

    ---

    Mount test data, certificates, and configuration files from various sources.

</div>

[See all new features →](features.md) | [Migration guide →](migration.md)
```

**Verification:**
- [x] Page renders correctly
- [x] Links work

---

## Task 16.11: Update CHANGELOG.md

**File:** `CHANGELOG.md`

### v2.0.0 Entry

```markdown
## 2.0.0 (YYYY-MM-DD)

### 🎉 Major Release - Go Rewrite

The Locust Kubernetes Operator has been completely rewritten in Go using 
controller-runtime. This release brings significant improvements and new features.

### Breaking Changes

- Operator image changed from Java to Go
- Helm chart updated with new configuration options
- JVM-specific settings removed

### New Features

- **Go Migration**: Complete rewrite using Operator SDK and controller-runtime
- **v2 API**: New grouped API structure with enhanced fields
- **OpenTelemetry Support**: Native OTel integration (Phase 12)
- **Secret Injection**: ConfigMap and Secret env injection (Phase 10, Issue #149)
- **Volume Mounting**: Arbitrary volume support (Phase 11, Issue #252)
- **Status Subresource**: Phase tracking and conditions (Phase 9)
- **Conversion Webhook**: Automatic v1↔v2 conversion (Phase 8)
- **Separate Resources**: Independent master/worker resource specs (Phase 7)

### Improvements

- Reduced memory footprint (~64MB vs ~256MB)
- Faster startup time (<1s vs 3-5s)
- Improved test coverage (80%+ unit, envtest integration)
- Multi-architecture Docker images (amd64, arm64)

### Migration

See the [Migration Guide](https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/) 
for upgrading from v1.x.
```

**Verification:**
- [x] Changelog follows existing format
- [x] All features listed

---

## Task 16.12: Update Developer Documentation (Critical)

These files were completely outdated - referencing Java/Gradle/JUnit instead of Go/Make/envtest.

### 16.12.1: Rewrite docs/local-development.md

**File:** `docs/local-development.md`

- [x] Update prerequisites (Go 1.23+, Docker, kubectl, Kind/Minikube, Helm)
- [x] Update initial setup (clone, make tidy, install tools)
- [x] Add Common Development Commands section (make build, test, lint, ci)
- [x] Add Code Generation section (make manifests, generate)
- [x] Update local testing with Kind (recommended approach)
- [x] Update local testing with Minikube (alternative)
- [x] Add project structure diagram
- [x] Update documentation preview commands

### 16.12.2: Rewrite docs/integration-testing.md

**File:** `docs/integration-testing.md`

- [x] Rename to "Testing Guide" (covers all test types)
- [x] Add overview table (Unit, Integration, E2E)
- [x] Add test structure diagram
- [x] Document unit tests (Go testing, examples)
- [x] Document integration tests (envtest, Ginkgo)
- [x] Document E2E tests (Kind, Ginkgo)
- [x] Add coverage targets table
- [x] Add troubleshooting section
- [x] Add "Writing New Tests" section

### 16.12.3: Update docs/pull-request-process.md

**File:** `docs/pull-request-process.md`

- [x] Update test commands (Gradle→Make)
- [x] Add manifest generation step
- [x] Add PR requirements checklist
- [x] Add CI pipeline checks table
- [x] Add common PR scenarios section

### 16.12.4: Rewrite docs/how_does_it_work.md

**File:** `docs/how_does_it_work.md`

- [x] Add architecture overview with ASCII diagram
- [x] Document core components (Controller, Builders, Webhook)
- [x] Document workflow (creation, execution, cleanup)
- [x] Add key design decisions section
- [x] Add technology stack table

### 16.12.5: Update docs/contribute.md

**File:** `docs/contribute.md`

- [x] Add Technology Stack section
- [x] Update link text for testing guide
- [x] Add link to architecture doc

---

## Verification

### Documentation Build

```bash
# Install dependencies if needed
pip install mkdocs mkdocs-material

# Serve locally
mkdocs serve

# Build with strict mode (catches broken links)
mkdocs build --strict
```

- [x] No build errors
- [x] No broken links
- [x] All pages render correctly

### Content Review

- [x] All v2 API examples validate against CRD
- [x] No references to Java/JVM/Micronaut in operator context
- [x] Helm values match actual chart
- [x] Screenshots updated if any exist

### Link Verification

- [x] Internal links work
- [x] External links work
- [x] GitHub links are correct

---

## Post-Implementation

- [x] All verification steps pass
- [x] `mkdocs build --strict` succeeds
- [x] Update `phases/README.md` with Phase 16 status
- [x] Commit with message: `docs: update documentation for v2.0 release`

---

## Files Summary

### User-Facing Documentation

| File | Action | Est. Changes |
|------|--------|--------------|
| `README.md` | Modify | +30 lines |
| `docs/index.md` | Modify | +50 lines |
| `docs/getting_started.md` | Modify | ~100 lines changed |
| `docs/features.md` | Modify | +80 lines |
| `docs/migration.md` | Create | ~200 lines |
| `docs/api_reference.md` | Create | ~300 lines |
| `docs/helm_deploy.md` | Modify | ~50 lines changed |
| `docs/advanced_topics.md` | Modify | +150 lines |
| `docs/metrics_and_dashboards.md` | Modify | +80 lines |
| `docs/roadmap.md` | Modify | ~20 lines |
| `CHANGELOG.md` | Modify | +40 lines |
| `mkdocs.yml` | Modify | +5 lines |

### Developer Documentation (Critical - Java → Go Migration)

| File | Action | Est. Changes |
|------|--------|--------------|
| `docs/local-development.md` | **Rewrite** | ~240 lines (was Java/Gradle, now Go/Make) |
| `docs/integration-testing.md` | **Rewrite** | ~310 lines (was JUnit/Testcontainers, now envtest/Ginkgo) |
| `docs/pull-request-process.md` | Modify | ~70 lines changed (Gradle→Make commands) |
| `docs/how_does_it_work.md` | **Rewrite** | ~145 lines (added Go architecture details) |
| `docs/contribute.md` | Modify | +15 lines (added tech stack section) |

**Total Estimated:** ~1900 lines of documentation

---

## Quick Reference Commands

```bash
# Preview docs
mkdocs serve

# Build docs
mkdocs build

# Strict build (catches errors)
mkdocs build --strict

# Deploy to GitHub Pages (usually via CI)
mkdocs gh-deploy
```

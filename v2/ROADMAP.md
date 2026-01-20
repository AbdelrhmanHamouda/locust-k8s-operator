# Locust K8s Operator v2.0 - Go Migration Roadmap

**Document Version:** 1.0  
**Created:** January 2026  
**Status:** Active  
**Total Estimated Effort:** 13-18 days

---

## Overview

This roadmap defines the phased implementation plan for migrating the Locust Kubernetes Operator from Java (Micronaut + JOSDK) to Go (Operator SDK). Each phase is designed to be:

- **Small enough** to be implementable in 1-3 days
- **Testable** with clear acceptance criteria
- **Independent** where possible for parallel work
- **Mapped** to specific requirements from `REQUIREMENTS.md`

---

## Phase Summary

| Phase | Name | Effort | Dependencies | Priority |
|-------|------|--------|--------------|----------|
| 0 | Project Scaffolding | 0.5 day | None | P0 |
| 1 | v1 API Types (Parity) | 1 day | Phase 0 | P0 |
| 2 | Configuration System | 0.5 day | Phase 0 | P0 |
| 3 | Resource Builders | 2 days | Phase 1, 2 | P0 |
| 4 | Core Reconciler | 1.5 days | Phase 3 | P0 |
| 5 | Unit Tests | 2 days | Phase 4 | P0 |
| 6 | Integration Tests (envtest) | 2 days | Phase 5 | P0 |
| 7 | v2 API Types (New Features) | 1.5 days | Phase 4 | P1 |
| 8 | Conversion Webhook (v1↔v2) | 1 day | Phase 7 | P1 |
| 9 | Status Subresource | 1 day | Phase 7 | P1 |
| 10 | Environment & Secret Injection | 1 day | Phase 7 | P1 |
| 11 | Volume Mounting | 0.5 day | Phase 7 | P1 |
| 12 | OpenTelemetry Support | 1 day | Phase 7 | P1 |
| 13 | Helm Chart Updates | 1 day | Phase 6 | P0 |
| 14 | CI/CD Pipeline | 0.5 day | Phase 13 | P0 |
| 15 | E2E Tests (Kind) | 2 days | Phase 14 | P1 |
| 16 | Documentation | 1 day | Phase 15 | P1 |
| 17 | OLM Bundle (Optional) | 1 day | Phase 16 | P2 |

---

## Phase 0: Project Scaffolding

**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Requirements:** §3.1 Technology Stack, §3.4 Project Structure

### Objective
Initialize the Go operator project using Operator SDK with proper structure.

### Tasks

- [x] **0.1** Install Operator SDK CLI (v1.37+)
- [x] **0.2** Initialize project structure
  ```bash
  operator-sdk init \
    --domain locust.io \
    --repo github.com/AbdelrhmanHamouda/locust-k8s-operator \
    --plugins go/v4
  ```
- [x] **0.3** Create v1 API scaffold
  ```bash
  operator-sdk create api \
    --group locust \
    --version v1 \
    --kind LocustTest \
    --resource \
    --controller
  ```
- [x] **0.4** Configure `go.mod` with correct dependencies
- [x] **0.5** Verify project builds: `make build`
- [x] **0.6** Verify CRD generation: `make manifests`

### Acceptance Criteria
- [ ] Project compiles without errors
- [ ] `make manifests` generates CRD in `config/crd/bases/`
- [ ] `make test` runs (even with no tests)
- [ ] Directory structure matches §3.4

### Output Artifacts
- `go.mod`, `go.sum`
- `cmd/main.go`
- `api/v1/` directory
- `internal/controller/` directory
- `config/` directory with RBAC, CRD bases

---

## Phase 1: v1 API Types (Parity)

**Effort:** 1 day  
**Priority:** P0 - Critical Path  
**Requirements:** §4.4 v1 to v2 Field Mapping (v1 fields)  
**Reference:** `analysis/TECHNICAL.md` §5.1, `research/JAVA_TO_GO_MAPPING.md` §1

### Objective
Define Go types that exactly match the current Java v1 CRD for backward compatibility.

### Tasks

- [x] **1.1** Define `LocustTestSpec` struct with all v1 fields
  - `masterCommandSeed`, `workerCommandSeed`, `workerReplicas`
  - `image`, `imagePullPolicy`, `imagePullSecrets`
  - `configMap`, `libConfigMap`
  - `labels`, `annotations` (nested maps)
  - `affinity`, `tolerations`
- [x] **1.2** Define supporting types
  - `PodLabels`, `PodAnnotations`
  - `LocustTestAffinity`, `LocustTestNodeAffinity`
  - `LocustTestToleration`
- [x] **1.3** Add kubebuilder validation markers
  - `+kubebuilder:validation:Required` for required fields
  - `+kubebuilder:validation:Minimum=1`, `Maximum=500` for `workerReplicas`
  - `+kubebuilder:validation:Enum` for `imagePullPolicy`
- [x] **1.4** Add printer columns for `kubectl get`
- [x] **1.5** Run `make manifests` and verify CRD matches Java version
- [x] **1.6** Write type tests for JSON marshaling/unmarshaling

### Acceptance Criteria
- [ ] Generated CRD schema matches existing Java CRD
- [ ] Sample CR from `kube/sample-cr/` validates against new CRD
- [ ] `make generate` produces `zz_generated.deepcopy.go`

### Files to Create/Modify
- `api/v1/locusttest_types.go`
- `api/v1/groupversion_info.go`

---


## Phase 2: Configuration System

**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Requirements:** §3.1 Configuration  
**Reference:** `analysis/TECHNICAL.md` §5.2, `research/JAVA_TO_GO_MAPPING.md` §3

### Objective
Implement environment-based configuration matching Java `SysConfig.java`.

### Tasks

- [x] **2.1** Create `internal/config/config.go`
- [x] **2.2** Define `OperatorConfig` struct with fields:
  - `TTLSecondsAfterFinished *int32`
  - `PodCPURequest`, `PodMemRequest`, `PodCPULimit`, `PodMemLimit`
  - `MetricsExporterImage`, `MetricsExporterPort`
  - `EnableAffinityCRInjection`, `EnableTolerationsCRInjection`
- [x] **2.3** Implement `LoadConfig()` function reading from env vars
- [x] **2.4** Add helper functions: `getEnv`, `getEnvBool`, `getEnvInt32`, `getEnvInt32Ptr`
- [x] **2.5** Write unit tests for config loading

### Acceptance Criteria
- [x] Config loads with defaults when env vars not set
- [x] Config respects env var overrides
- [x] All Java `application.yml` properties have Go equivalents

### Files to Create
- `internal/config/config.go`
- `internal/config/config_test.go`

---

## Phase 3: Resource Builders

**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Requirements:** §3.3 Resilience (idempotent creation)  
**Reference:** `analysis/TECHNICAL.md` §5.3, `research/RESOURCE_MANAGEMENT.md`

### Objective
Implement Job and Service builders matching Java `ResourceCreationHelpers.java`.

### Tasks

#### Day 1: Core Builders
- [x] **3.1** Create `internal/resources/types.go`
  - Define `OperationalMode` (Master/Worker)
  - Define internal DTOs if needed
- [x] **3.2** Create `internal/resources/constants.go`
  - Port constants from Java `Constants.java`
  - Ports: 5557, 5558, 8089, 8080, 9646
  - Mount paths: `/lotest/src/`, `/opt/locust/lib`
- [x] **3.3** Create `internal/resources/labels.go`
  - `NodeName(crName, mode)` function
  - `BuildLabels(lt, mode)` function
  - `BuildAnnotations(lt, mode, cfg)` function

#### Day 2: Job & Service Builders
- [x] **3.4** Create `internal/resources/command.go`
  - `BuildMasterCommand()` - construct master CLI args
  - `BuildWorkerCommand()` - construct worker CLI args
- [x] **3.5** Create `internal/resources/job.go`
  - `BuildMasterJob(lt, cfg)` - returns `*batchv1.Job`
  - `BuildWorkerJob(lt, cfg)` - returns `*batchv1.Job`
  - `buildLocustContainer()` helper
  - `buildMetricsExporterContainer()` helper
  - `buildVolumes()`, `buildVolumeMounts()` helpers
- [x] **3.6** Create `internal/resources/service.go`
  - `BuildMasterService(lt, cfg)` - returns `*corev1.Service`
- [x] **3.7** Create `internal/resources/ports.go`
  - `MasterPorts()`, `WorkerPorts()` helpers

### Acceptance Criteria
- [x] `BuildMasterJob()` produces Job spec matching Java output
- [x] `BuildWorkerJob()` produces Job spec matching Java output
- [x] `BuildMasterService()` produces Service spec matching Java output
- [x] All resource builders are pure functions (no side effects)

### Files to Create
- `internal/resources/types.go`
- `internal/resources/constants.go`
- `internal/resources/labels.go`
- `internal/resources/command.go`
- `internal/resources/job.go`
- `internal/resources/service.go`
- `internal/resources/ports.go`

---

## Phase 4: Core Reconciler

**Effort:** 1.5 days  
**Priority:** P0 - Critical Path  
**Requirements:** §3.3 Resilience, §8.3 Design Principles (Immutable Tests)  
**Reference:** `analysis/TECHNICAL.md` §5.4, `research/OPERATOR_SDK_PATTERNS.md` §3

### Objective
Implement the reconciliation loop matching Java `LocustTestReconciler.java` behavior.

### Tasks

- [x] **4.1** Implement `Reconcile()` method in `internal/controller/locusttest_controller.go`
  - Fetch LocustTest CR
  - Handle not found (deleted)
  - **NO-OP on updates** (generation > 1) - matching Java behavior
- [x] **4.2** Implement `createResources()` helper
  - Build master Service, master Job, worker Job
  - Set owner references for automatic cleanup
  - Create resources with `IsAlreadyExists` handling
- [x] **4.3** Implement `SetupWithManager()` 
  - Watch `LocustTest` resources
  - Own `Job` and `Service` resources
  - Apply `GenerationChangedPredicate` filter
- [x] **4.4** Wire reconciler in `cmd/main.go`
  - Load config
  - Create reconciler with dependencies
  - Register with manager
- [x] **4.5** Add event recording for resource creation

### Acceptance Criteria
- [x] CR creation triggers Job and Service creation
- [x] CR updates are NO-OP (by design)
- [x] CR deletion triggers automatic cleanup via owner references
- [x] Operator logs match Java logging patterns
- [x] `make run` starts operator successfully

### Files to Modify
- `internal/controller/locusttest_controller.go`
- `cmd/main.go`

---

## Phase 5: Unit Tests

**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Requirements:** §7.1 Testing Requirements (80% coverage)  
**Reference:** `analysis/TECHNICAL.md` §4.2.1

### Objective
Port unit tests from Java to Go, achieving 80% code coverage.

### Tasks

#### Day 1: Resource Builder Tests
- [x] **5.1** Create `internal/resources/job_test.go`
  - Test `BuildMasterJob()` output structure
  - Test `BuildWorkerJob()` output structure
  - Test command construction
  - Test volume mounts
- [x] **5.2** Create `internal/resources/service_test.go`
  - Test `BuildMasterService()` output structure
  - Test port configuration
- [x] **5.3** Create `internal/resources/labels_test.go`
  - Test `NodeName()` edge cases
  - Test label merging
  - Test annotation building

#### Day 2: Controller & Config Tests
- [x] **5.4** Create `internal/config/config_test.go`
  - Test default values
  - Test env var overrides
  - Test optional/nullable fields
- [x] **5.5** Create `internal/controller/locusttest_controller_test.go`
  - Test reconcile on create
  - Test NO-OP on update
  - Test resource creation logic
- [x] **5.6** Create test fixtures in `internal/testdata/`
- [x] **5.7** Run coverage: `make test COVERAGE=true`

### Acceptance Criteria
- [x] All tests pass: `make test`
- [x] Coverage ≥ 80% for `internal/resources/` (97.7%)
- [x] Coverage ≥ 70% for `internal/controller/` (100%)

### Files to Create
- `internal/resources/*_test.go`
- `internal/config/config_test.go`
- `internal/controller/locusttest_controller_test.go`
- `internal/testdata/` fixtures

---

## Phase 6: Integration Tests (envtest)

**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Requirements:** §7.1 Testing Requirements (envtest)  
**Reference:** `CONTROLLER_RUNTIME_DEEP_DIVE.md`

### Objective
Implement controller integration tests using envtest framework.

### Tasks

- [x] **6.1** Set up envtest in `internal/controller/suite_test.go`
  - Configure test environment
  - Start API server
  - Register CRDs
- [x] **6.2** Implement create flow test
  - Create LocustTest CR
  - Verify Service created
  - Verify master Job created
  - Verify worker Job created
  - Verify owner references set
- [x] **6.3** Implement update NO-OP test
  - Create CR
  - Update CR spec
  - Verify no new resources created
- [x] **6.4** Implement delete flow test
  - Create CR
  - Delete CR
  - Verify CR deleted (note: envtest lacks GC controller)
- [x] **6.5** Implement error handling tests
  - Idempotent resource creation
  - Multi-namespace isolation
  - Rapid create/delete cycles

### Acceptance Criteria
- [x] All integration tests pass: `make test` (21 tests)
- [x] Tests run in CI without external dependencies
- [x] Test execution < 2 minutes (~31 seconds)

### Files to Create
- `internal/controller/suite_test.go`
- `internal/controller/integration_test.go`

---

## Phase 7: v2 API Types (New Features)

**Effort:** 1.5 days  
**Priority:** P1 - Must Have  
**Requirements:** §4.2 LocustTest v2 Spec, §5.1 Must Have Features  
**Reference:** `CRD_API_DESIGN.md` §2

### Objective
Define v2 API with grouped configuration and new feature fields.

### Tasks

- [x] **7.1** Create v2 API scaffold
  ```bash
  operator-sdk create api \
    --group locust \
    --version v2 \
    --kind LocustTest \
    --resource \
    --controller=false
  ```
- [x] **7.2** Define grouped specs in `api/v2/locusttest_types.go`
  - `MasterSpec` (command, resources, labels, annotations, autostart, autoquit, extraArgs)
  - `WorkerSpec` (command, replicas, resources, labels, annotations, extraArgs)
  - `TestFilesConfig` (configMapRef, libConfigMapRef, srcMountPath, libMountPath)
  - `SchedulingConfig` (affinity, tolerations, nodeSelector)
- [x] **7.3** Define new feature types
  - `EnvConfig` (configMapRefs, secretRefs, variables, secretMounts)
  - `VolumeMount` with target field (master/worker/both)
  - `ObservabilityConfig` with `OpenTelemetryConfig`
- [x] **7.4** Define `LocustTestStatus` with phase and conditions
- [x] **7.5** Add kubebuilder markers
  - `+kubebuilder:storageversion` on v2
  - Printer columns for phase, workers, age
  - `+kubebuilder:subresource:status`
- [x] **7.6** Generate manifests: `make manifests`

### Acceptance Criteria
- [x] v2 CRD generated with all new fields
- [x] v2 marked as storage version
- [x] Sample v2 CR validates against schema

### Files to Create
- `api/v2/locusttest_types.go`
- `api/v2/groupversion_info.go`

---

## Phase 8: Conversion Webhook (v1↔v2)

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Requirements:** §5.1.5 API v1 Conversion Webhook  
**Reference:** `CRD_API_DESIGN.md` §3

### Objective
Implement bidirectional conversion between v1 and v2 APIs.

### Tasks

- [x] **8.1** Mark v2 as Hub: `api/v2/locusttest_conversion.go`
  - Implement `Hub()` method
- [x] **8.2** Implement v1 Spoke conversion: `api/v1/locusttest_conversion.go`
  - `ConvertTo()` - v1 → v2
  - `ConvertFrom()` - v2 → v1 (lossy for new fields)
- [x] **8.3** Create webhook scaffold
  ```bash
  operator-sdk create webhook \
    --group locust \
    --version v1 \
    --kind LocustTest \
    --conversion
  ```
- [x] **8.4** Configure webhook in `config/webhook/`
- [x] **8.5** Add deprecation warning marker to v1
- [x] **8.6** Write conversion tests

### Acceptance Criteria
- [x] v1 CR can be read as v2
- [x] v2 CR can be stored and read as v1 (with data loss warning)
- [x] Conversion tests pass
- [x] Deprecation warning appears when using v1

### Files to Create
- `api/v1/locusttest_conversion.go`
- `api/v2/locusttest_conversion.go`
- `api/v1/locusttest_webhook.go`

---

## Phase 9: Status Subresource

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Requirements:** §5.1.4 Status Subresource  
**Reference:** `CRD_API_DESIGN.md` §5, `OPERATOR_SDK_PATTERNS.md` §5

### Objective
Implement status tracking with phase and conditions.

### Tasks

- [x] **9.1** Define condition constants in `api/v2/conditions.go`
  - `ConditionTypeReady`, `ConditionTypeWorkersConnected`, `ConditionTypeTestCompleted`
  - Reason constants
  - Phase constants (Pending, Running, Succeeded, Failed)
- [x] **9.2** Create status helpers in `internal/controller/status.go`
  - `updatePhase()`
  - `setCondition()`
  - `setReady()`
- [x] **9.3** Update reconciler to track status
  - Set phase to Pending on create
  - Set phase to Running when Jobs start
  - Update conditions based on Job status
- [x] **9.4** Add Job status watching
- [x] **9.5** Write status update tests

### Acceptance Criteria
- [x] `kubectl get locusttests` shows Phase column
- [x] Status reflects actual Job states
- [x] Conditions follow K8s conventions
- [x] Status updates don't trigger reconcile loops

### Files to Create
- `api/v2/conditions.go`
- `internal/controller/status.go`

---

## Phase 10: Environment & Secret Injection

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Requirements:** §5.1.2 Environment Injection (Issue #149)  
**Status:** ✅ Complete (2026-01-20)

### Objective
Enable ConfigMap and Secret injection into Locust pods.

### Tasks

- [x] **10.1** Update resource builders to handle `env` config
  - Process `configMapRefs` → `envFrom` with ConfigMapRef
  - Process `secretRefs` → `envFrom` with SecretRef
  - Process `variables` → individual `env` entries
- [x] **10.2** Implement secret file mounting
  - Process `secretMounts` → Volume + VolumeMount
- [x] **10.3** Add validation webhook
  - Ensure no conflicts with operator-managed paths
- [x] **10.4** Write tests for env injection

### Acceptance Criteria
- [x] ConfigMap values available as env vars in pods
- [x] Secret values available as env vars in pods
- [x] Secrets mountable as files
- [x] Path conflicts rejected at validation

### Files Created/Modified
- `internal/resources/env.go` (new - 120 LOC)
- `internal/resources/env_test.go` (new - 430 LOC)
- `api/v2/locusttest_webhook.go` (new - 140 LOC)
- `api/v2/locusttest_webhook_test.go` (new - 340 LOC)
- `internal/resources/job.go` (modified)
- `internal/resources/job_test.go` (modified - +7 tests)
- `config/samples/locust_v2_locusttest_with_env.yaml` (new)

---

## Phase 11: Volume Mounting

**Effort:** 0.5 day  
**Priority:** P1 - Must Have  
**Requirements:** §5.1.3 Volume Mounting (Issue #252)

### Objective
Enable arbitrary volume mounting to Locust pods.

### Tasks

- [ ] **11.1** Update Job builders to process `volumes` and `volumeMounts`
- [ ] **11.2** Implement target filtering (master/worker/both)
- [ ] **11.3** Add validation for reserved paths
  - Block `/lotest/src/` and `/opt/locust/lib` conflicts
- [ ] **11.4** Write volume mounting tests

### Acceptance Criteria
- [ ] User-defined volumes mounted to correct pods
- [ ] Target filtering works correctly
- [ ] Reserved path conflicts rejected

### Files to Modify
- `internal/resources/job.go`
- `internal/resources/volumes.go` (new)

---

## Phase 12: OpenTelemetry Support

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Requirements:** §5.1.1 Native OpenTelemetry Support  
**Reference:** `analysis/LOCUST_FEATURES.md` §1.1

### Objective
Replace metrics exporter sidecar with native Locust OTel integration.

### Tasks

- [ ] **12.1** Update command builders
  - Add `--otel` flag when OTel enabled
- [ ] **12.2** Implement OTel env var injection
  - `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`
  - `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_PROTOCOL`
  - Custom `extraEnvVars` from spec
- [ ] **12.3** Conditionally disable metrics sidecar
  - Skip sidecar when `observability.openTelemetry.enabled: true`
- [ ] **12.4** Update Service ports (remove metrics port when OTel)
- [ ] **12.5** Write OTel integration tests

### Acceptance Criteria
- [ ] `--otel` flag added when OTel enabled
- [ ] OTel env vars correctly injected
- [ ] Metrics sidecar not deployed when OTel enabled
- [ ] Backward compatible (sidecar still works when OTel disabled)

### Files to Modify
- `internal/resources/command.go`
- `internal/resources/job.go`
- `internal/resources/otel.go` (new)

---

## Phase 13: Helm Chart Updates

**Effort:** 1 day  
**Priority:** P0 - Critical Path  
**Requirements:** §6.2 Helm Chart

### Objective
Update Helm chart for Go operator deployment.

### Tasks

- [ ] **13.1** Update `values.yaml`
  - Change image to Go operator image
  - Remove JVM-specific configs (JAVA_OPTS, memory)
  - Add OTel Collector options
  - Reduce default memory limits (~64Mi)
- [ ] **13.2** Update `Chart.yaml`
  - Bump chart version
  - Update appVersion
- [ ] **13.3** Update deployment template
  - Remove JVM environment variables
  - Add Go operator env vars
- [ ] **13.4** Add OTel Collector deployment (optional)
- [ ] **13.5** Update RBAC for new permissions
- [ ] **13.6** Run `helm lint` and `helm template` tests

### Acceptance Criteria
- [ ] `helm lint` passes
- [ ] Chart installs on test cluster
- [ ] Operator runs with default values
- [ ] OTel Collector deployable via values

### Files to Modify
- `charts/locust-k8s-operator/values.yaml`
- `charts/locust-k8s-operator/Chart.yaml`
- `charts/locust-k8s-operator/templates/deployment.yaml`

---

## Phase 14: CI/CD Pipeline

**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Requirements:** §6.3 CI/CD Pipeline

### Objective
Update GitHub Actions for Go build and test.

### Tasks

- [ ] **14.1** Update `.github/workflows/ci.yaml`
  - Replace Gradle with Go build steps
  - Add `golangci-lint` step
  - Update test commands
  - Add coverage reporting
- [ ] **14.2** Update Docker build
  - Use `ko`
  - Multi-arch support (amd64, arm64)
- [ ] **14.3** Update release workflow
  - GoReleaser or `ko` for releases
  - Helm chart release
- [ ] **14.4** Add Makefile targets for CI

### Acceptance Criteria
- [ ] CI passes on PR
- [ ] Docker image builds and pushes
- [ ] Multi-arch images available
- [ ] Release workflow functional

### Files to Modify
- `.github/workflows/ci.yaml`
- `.github/workflows/release.yml` (new or update)
- `Dockerfile`
- `Makefile`

---

## Phase 15: E2E Tests (Kind)

**Effort:** 2 days  
**Priority:** P1 - Should Have  
**Requirements:** §7.1 Testing Requirements (E2E)

### Objective
Implement end-to-end tests using Kind cluster.

### Tasks

- [ ] **15.1** Set up E2E test framework in `test/e2e/`
- [ ] **15.2** Create Kind cluster configuration
- [ ] **15.3** Implement basic test flow
  - Install operator
  - Create LocustTest CR
  - Verify pods running
  - Verify test completion
  - Clean up
- [ ] **15.4** Implement v1 backward compatibility test
- [ ] **15.5** Implement OTel integration test
- [ ] **15.6** Add E2E to CI pipeline

### Acceptance Criteria
- [ ] E2E tests pass locally
- [ ] E2E tests pass in CI
- [ ] E2E tests contribute coverage to Codecov if applicable
- [ ] Tests complete in < 10 minutes

### Files to Create
- `test/e2e/e2e_test.go`
- `test/e2e/suite_test.go`
- `.github/workflows/e2e.yaml`

---

## Phase 16: Documentation

**Effort:** 1 day  
**Priority:** P1 - Should Have  
**Requirements:** §7.2 Code Quality (Documentation)

### Objective
Update all documentation for v2.0 release.

### Tasks

- [ ] **16.1** Update `README.md`
  - New installation instructions
  - v2 API examples
  - Migration guide from v1
- [ ] **16.2** Update `docs/getting_started.md`
- [ ] **16.3** Update `docs/features.md` with new features
  - OTel support
  - Secret injection
  - Volume mounting
  - Separate resources
- [ ] **16.4** Add `docs/migration.md` for v1→v2
- [ ] **16.5** Update API reference docs
- [ ] **16.6** Add CHANGELOG entry for v2.0

### Acceptance Criteria
- [ ] All docs reflect v2 API
- [ ] Migration guide complete
- [ ] Examples work with new operator

### Files to Modify
- `README.md`
- `docs/*.md`
- `CHANGELOG.md`

---

## Phase 17: OLM Bundle (Optional)

**Effort:** 1 day  
**Priority:** P2 - Nice to Have  
**Requirements:** §5.2.3 OLM Bundle Distribution

### Objective
Generate OLM bundle for OperatorHub distribution.

### Tasks

- [ ] **17.1** Generate bundle manifests
  ```bash
  operator-sdk generate bundle \
    --version 2.0.0 \
    --channels stable,alpha \
    --default-channel stable
  ```
- [ ] **17.2** Configure CSV (ClusterServiceVersion)
  - Add operator description
  - Add icon
  - Configure RBAC
- [ ] **17.3** Validate bundle
  ```bash
  operator-sdk bundle validate ./bundle
  ```
- [ ] **17.4** Test with `operator-sdk run bundle`
- [ ] **17.5** Add scorecard tests

### Acceptance Criteria
- [ ] Bundle validates
- [ ] Bundle installs via OLM
- [ ] Scorecard tests pass

### Files to Create
- `bundle/` directory
- `bundle/manifests/`
- `bundle/metadata/`

---

## Requirement Traceability

| Requirement | Section | Phases |
|-------------|---------|--------|
| Go Migration | §2.1 Goal 1 | 0-6, 13-14 |
| v1 API Parity | §4.4 | 1, 8 |
| v2 API New Features | §4.2 | 7 |
| Status Subresource | §5.1.4 | 9 |
| Conversion Webhook | §5.1.5 | 8 |
| OpenTelemetry Support | §5.1.1 | 12 |
| Secret Injection | §5.1.2 (Issue #149) | 10 |
| Volume Mounting | §5.1.3 (Issue #252) | 11 |
| Separate Resources | §5.2.1 (Issue #246) | 7 |
| Configurable Commands | §5.2.2 (Issue #245) | 7 |
| Helm Chart | §6.2 | 13 |
| CI/CD Pipeline | §6.3 | 14 |
| Unit Tests (80%) | §7.1 | 5 |
| Integration Tests | §7.1 | 6 |
| E2E Tests | §7.1 | 15 |
| Documentation | §7.2 | 16 |
| OLM Distribution | §5.2.3 | 17 |

---

## Critical Path

The minimum viable migration requires completing these phases in order:

```
Phase 0 → Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 13 → Phase 14
```

This delivers a **functionally equivalent Go operator** with v1 API compatibility.

New features (Phases 7-12) can be implemented in parallel after Phase 6 is complete.

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| API schema mismatch | Compare generated CRD with existing Java CRD in Phase 1 |
| Behavior differences | Comprehensive unit + integration tests in Phases 5-6 |
| Conversion data loss | Document v2-only fields, warn on v1 usage |
| CI/CD breakage | Test pipeline changes in feature branch first |
| Helm compatibility | Test chart upgrade from current version |

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01 | Initial roadmap |

---

**Companion Documents:**
- `REQUIREMENTS.md` - Feature requirements specification
- `analysis/ASSESSMENT.md` - Technical viability assessment
- `analysis/TECHNICAL.md` - Detailed technical analysis
- `research/` - Domain research findings

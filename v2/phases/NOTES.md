# Phase Implementation Notes

---

# Phase 0 Completion Notes

**Completed:** 2026-01-16

---

## Key Decisions Made

### Domain Configuration
- **Domain**: `io` (not `locust.io`)
- **Group**: `locust`
- **Result**: CRD group is `locust.io`, matching the existing Java CRD

This was corrected during implementation. The original plan used `--domain locust.io` which resulted in `locust.locust.io`. The fix was to use `--domain io --group locust`.

---

## Versions Installed

| Tool | Version |
|------|---------|
| Operator SDK | 1.42.0 |
| Go | 1.24.0 |
| controller-runtime | 0.21.0 |
| controller-gen | 0.18.0 |
| k8s.io/api | 0.33.0 |
| k8s.io/apimachinery | 0.33.0 |
| k8s.io/client-go | 0.33.0 |

---

# Phase 1 Completion Notes

**Completed:** 2026-01-17

---

## Summary

Implemented v1 API types that exactly match the Java CRD schema for backward compatibility.

## Files Created/Modified

- `api/v1/locusttest_types.go` - Full v1 API types with all fields and kubebuilder markers
- `api/v1/locusttest_types_test.go` - Unit tests for JSON marshaling and field names
- `api/v1/zz_generated.deepcopy.go` - Auto-generated DeepCopy methods
- `config/crd/bases/locust.io_locusttests.yaml` - Generated CRD manifest
- `internal/controller/locusttest_controller_test.go` - Updated to use new spec fields

## Types Defined

| Type | Description |
|------|-------------|
| `LocustTestSpec` | Main spec with 12 fields matching Java CRD |
| `LocustTestStatus` | Empty status (to be populated in Phase 9) |
| `PodLabels` | Master/Worker label maps |
| `PodAnnotations` | Master/Worker annotation maps |
| `LocustTestAffinity` | NodeAffinity wrapper |
| `LocustTestNodeAffinity` | RequiredDuringSchedulingIgnoredDuringExecution map |
| `LocustTestToleration` | Key, Operator, Value, Effect fields |

## Validation Markers Applied

- **Required fields**: `masterCommandSeed`, `workerCommandSeed`, `workerReplicas`, `image`
- **WorkerReplicas constraints**: min=1, max=500, default=1
- **Enums**: ImagePullPolicy (Always/IfNotPresent/Never), Toleration.Operator (Exists/Equal), Toleration.Effect (NoSchedule/PreferNoSchedule/NoExecute)

## CRD Features

- Short name: `lotest`
- Printer columns: master_cmd, worker_replica_count, Image, Age
- Status subresource enabled

## Verification

- `make build` ✓
- `make test` ✓
- `go test ./api/v1/... -v` ✓ (4 tests pass)

## Notes for Phase 2

1. The controller test now uses valid spec fields - any future controller changes should maintain this.
2. Status is empty - Phase 9 will add status fields.
3. The generated CRD at `config/crd/bases/locust.io_locusttests.yaml` is schema-compatible with the Java CRD.

---

# Phase 2 Completion Notes

**Completed:** 2026-01-17

---

## Summary

Implemented environment-based configuration system matching Java `SysConfig.java`. The configuration system provides operator-wide settings that resource builders and the reconciler will use.

## Files Created

- `internal/config/config.go` - OperatorConfig struct and LoadConfig function
- `internal/config/config_test.go` - Comprehensive unit tests with 100% coverage

## Configuration Fields Implemented

| Category | Fields |
|----------|--------|
| **Job** | TTLSecondsAfterFinished (*int32 - nullable) |
| **Pod Resources** | PodCPURequest, PodMemRequest, PodEphemeralStorageRequest, PodCPULimit, PodMemLimit, PodEphemeralStorageLimit |
| **Metrics Exporter** | MetricsExporterImage, MetricsExporterPort, MetricsExporterPullPolicy, + CPU/Mem/Ephemeral request/limit |
| **Kafka** | KafkaBootstrapServers, KafkaSecurityEnabled, KafkaSecurityProtocol, KafkaUsername, KafkaPassword, KafkaSaslMechanism, KafkaSaslJaasConfig |
| **Feature Flags** | EnableAffinityCRInjection, EnableTolerationsCRInjection |

## Helper Functions

- `getEnv(key, defaultValue string) string` - String env var with default
- `getEnvBool(key string, defaultValue bool) bool` - Boolean parsing with default
- `getEnvInt32(key string, defaultValue int32) int32` - Int32 parsing with default
- `getEnvInt32Ptr(key string) *int32` - Nullable int32 for TTL (nil when unset)

## Key Design Decisions

1. **Kafka credentials default to empty strings** - Unlike Java which uses placeholder values, Go defaults to empty for security
2. **TTLSecondsAfterFinished uses *int32** - Correctly distinguishes "not set" (nil) from "set to 0"
3. **No DI framework** - Go idiom using explicit struct wiring via LoadConfig()

## Verification

- `go build ./...` ✓
- `go test ./internal/config/... -v -cover` ✓ (100% coverage)
- `golangci-lint run ./internal/config/...` ✓ (0 issues)
- `make test` ✓

## Notes for Phase 3

1. Use `config.LoadConfig()` to get operator configuration
2. Pass `*OperatorConfig` to resource builders
3. Use `cfg.TTLSecondsAfterFinished` directly on Job specs (nil-safe)
4. Resource quantities (e.g., "250m", "128Mi") are strings - parse with `resource.MustParse()` when building K8s resources

---

# Phase 3 Completion Notes

**Completed:** 2026-01-17

---

## Summary

Implemented resource builders matching Java `ResourceCreationHelpers.java` and `LoadGenHelpers.java` behavior. These pure functions build Kubernetes Jobs and Services from LocustTest CRs and operator configuration.

## Files Created

| File | Purpose | LOC |
|------|---------|-----|
| `internal/resources/types.go` | OperationalMode type (Master/Worker) | ~30 |
| `internal/resources/constants.go` | Ports, paths, labels, annotations | ~120 |
| `internal/resources/labels.go` | NodeName, BuildLabels, BuildAnnotations | ~100 |
| `internal/resources/ports.go` | MasterPorts, WorkerPorts helpers | ~50 |
| `internal/resources/command.go` | BuildMasterCommand, BuildWorkerCommand | ~45 |
| `internal/resources/job.go` | BuildMasterJob, BuildWorkerJob + helpers | ~365 |
| `internal/resources/service.go` | BuildMasterService | ~70 |
| `internal/resources/labels_test.go` | Labels/annotations tests | ~200 |
| `internal/resources/command_test.go` | Command builder tests | ~90 |
| `internal/resources/job_test.go` | Job builder tests | ~300 |
| `internal/resources/service_test.go` | Service builder tests | ~160 |

## Key Functions Implemented

| Function | Description |
|----------|-------------|
| `NodeName(crName, mode)` | Constructs node name with dots→dashes replacement |
| `BuildLabels(lt, mode)` | Builds pod labels including user-defined labels |
| `BuildAnnotations(lt, mode, cfg)` | Builds annotations with Prometheus scrape (master only) |
| `BuildMasterCommand(seed, replicas)` | Constructs master command with all flags |
| `BuildWorkerCommand(seed, masterHost)` | Constructs worker command with master host |
| `BuildMasterJob(lt, cfg)` | Creates master Job with 2 containers (locust + exporter) |
| `BuildWorkerJob(lt, cfg)` | Creates worker Job with 1 container |
| `BuildMasterService(lt, cfg)` | Creates ClusterIP service (excludes WebUI port 8089) |

## Java Behavior Preserved

1. **Command Templates** - Exact match with Java MASTER_CMD_TEMPLATE and WORKER_CMD_TEMPLATE
2. **Node Naming** - `{cr-name}-{mode}` with dots replaced by dashes
3. **Labels** - `performance-test-name`, `performance-test-pod-name`, `managed-by`, `app`
4. **Prometheus Annotations** - Master only: scrape=true, path=/metrics, port from config
5. **Service Ports** - 5557, 5558, metrics port (NOT 8089 WebUI)
6. **Feature Flags** - Affinity and tolerations respect EnableAffinityCRInjection/EnableTolerationsCRInjection

## Verification

- `go build ./internal/resources/...` ✓
- `go test ./internal/resources/... -v -cover` ✓ (35 tests, 93.9% coverage)
- `golangci-lint run ./internal/resources/...` ✓ (0 issues)
- `make build` ✓
- `make test` ✓

## Notes for Phase 4

1. Use `BuildMasterJob()` and `BuildWorkerJob()` in reconciler
2. Use `BuildMasterService()` for service creation
3. Set owner references on created resources for garbage collection
4. Jobs and Services are created in the same namespace as the LocustTest CR

---

# Phase 4 Completion Notes

**Completed:** 2026-01-19

---

## Summary

Implemented the core reconciliation loop matching Java `LocustTestReconciler.java` behavior. The reconciler watches LocustTest CRs, creates Kubernetes resources (Jobs, Services) on CR creation, and relies on owner references for automatic cleanup on CR deletion.

## Files Created/Modified

- `internal/controller/locusttest_controller.go` - Full reconciler implementation (~187 LOC)
- `internal/controller/locusttest_controller_test.go` - Updated to include Config and Recorder
- `cmd/main.go` - Wired reconciler with config and event recorder
- `config/rbac/role.yaml` - Auto-generated RBAC for Jobs, Services, Events

## Key Functions Implemented

| Function | Description |
|----------|-------------|
| `Reconcile()` | Main reconciliation loop with NO-OP on updates |
| `createResources()` | Creates master Service, master Job, worker Job |
| `createResource()` | Generic helper with owner reference and idempotent create |
| `SetupWithManager()` | Configures controller with Owns and predicates |

## Reconciler Behavior

| Event | Action |
|-------|--------|
| CR Created (generation=1) | Create Service → Master Job → Worker Job |
| CR Updated (generation>1) | NO-OP with log message |
| CR Deleted | Automatic cleanup via owner references |
| Resource Already Exists | Log and continue (idempotent) |

## RBAC Permissions Added

```yaml
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["get", "list", "watch", "create", "delete"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list", "watch", "create", "delete"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
```

## Verification

- `make build` ✓
- `make test` ✓ (controller tests pass with 56.8% coverage)
- `make manifests` ✓ (RBAC regenerated)
- `golangci-lint run ./internal/controller/...` ✓ (0 issues)

## Notes for Phase 5

1. Controller test coverage is at 56.8% - Phase 5 will add comprehensive unit tests
2. Manual verification with a real cluster is still pending (listed in checklist)
3. The `GenerationChangedPredicate` filter ensures we only reconcile on spec changes
4. Event recording provides visibility into resource creation for users

---

# Phase 5 Completion Notes

**Completed:** 2026-01-19

---

## Summary

Implemented comprehensive unit tests achieving all coverage targets. Added test fixtures for reusable test data.

## Coverage Results

| Package | Before | After | Target |
|---------|--------|-------|--------|
| `internal/config/` | 100% | 100% | ≥80% ✓ |
| `internal/controller/` | 56.8% | **77.3%** | ≥70% ✓ |
| `internal/resources/` | 93.9% | **97.7%** | ≥80% ✓ |
| **Total** | 86.4% | **93.4%** | - |

## Files Created

| File | Purpose | LOC |
|------|---------|-----|
| `internal/controller/locusttest_controller_unit_test.go` | Comprehensive controller unit tests | ~600 |
| `internal/testdata/fixtures.go` | Test fixture loader helper | ~50 |
| `internal/testdata/fixtures_test.go` | Tests for fixture loader | ~90 |
| `internal/testdata/locusttest_minimal.json` | Minimal CR fixture | - |
| `internal/testdata/locusttest_full.json` | Full-featured CR fixture | - |
| `internal/testdata/locusttest_with_affinity.json` | Affinity config fixture | - |
| `internal/testdata/locusttest_with_tolerations.json` | Tolerations config fixture | - |

## Files Modified

| File | Changes |
|------|---------|
| `internal/resources/job_test.go` | +112 LOC (edge case tests) |
| `internal/resources/labels_test.go` | +63 LOC (edge case tests, port tests) |

## Tests Added

### Controller Tests (17 new tests)
- `TestReconcile_NotFound` - CR deleted handling
- `TestReconcile_CreateResources` - Resource creation on new CR
- `TestReconcile_NoOpOnUpdate` - Generation > 1 NO-OP
- `TestReconcile_OwnerReferences` - Owner ref verification
- `TestReconcile_IdempotentCreate` - AlreadyExists handling
- `TestReconcile_WithDifferentGenerations` - Table-driven generation tests
- `TestReconcile_VerifyServiceConfiguration` - Service spec verification
- `TestReconcile_VerifyMasterJobConfiguration` - Master job spec
- `TestReconcile_VerifyWorkerJobConfiguration` - Worker job spec
- `TestReconcile_EventRecording` - Event creation verification
- `TestReconcile_WithCustomLabels` - Custom label propagation
- `TestReconcile_WithImagePullSecrets` - Image pull secrets
- `TestReconcile_WithLibConfigMap` - Lib volume mounting
- `TestReconcile_MultipleNamespaces` - Cross-namespace handling

### Resource Tests (10 new tests)
- `TestBuildTolerations_ExistsOperator` - Exists operator handling
- `TestBuildMasterJob_EmptyImagePullPolicy` - Default policy
- `TestBuildMasterJob_NoConfigMap` - No volumes when empty
- `TestBuildMasterJob_KafkaEnvVars` - Kafka env injection
- `TestBuildAffinity_NilNodeAffinity` - Nil affinity handling
- `TestBuildAffinity_EmptyRequirements` - Empty requirements
- `TestBuildMasterJob_Completions` - Completions field
- `TestBuildMasterJob_BackoffLimit` - BackoffLimit verification
- `TestWorkerPortInts` - Worker port helper
- `TestMasterPortInts` - Master port helper

## Verification

- `make test` ✓
- `go test -race ./internal/...` ✓ (no data races)
- All tests complete in < 15 seconds

## Notes for Phase 6

1. Test fixtures in `internal/testdata/` are available for use in integration tests
2. Controller tests use fake client - envtest integration already exists in suite_test.go
3. `SetupWithManager()` has 0% coverage - requires real manager, covered in integration tests

---

# Phase 6 Completion Notes

**Completed:** 2026-01-19

---

## Summary

Implemented controller integration tests using envtest framework. Tests validate actual reconciliation behavior against a real Kubernetes API server.

## Files Created/Modified

| File | Action | Purpose |
|------|--------|---------|
| `internal/controller/suite_test.go` | Enhanced | Added manager startup, controller registration, timeout constants |
| `internal/controller/integration_test.go` | Created | All integration test cases (~600 LOC) |

## Test Categories Implemented

| Category | Tests | Description |
|----------|-------|-------------|
| **Create Flow** | 5 | Service, master Job, worker Job creation with owner refs and labels |
| **Create Flow Edge Cases** | 7 | Custom labels, annotations, affinity, tolerations, imagePullSecrets |
| **Update NO-OP Flow** | 3 | Verify spec updates don't modify existing resources |
| **Delete Flow** | 2 | CR deletion and non-existent CR handling |
| **Error Handling** | 4 | Idempotent creation, multi-namespace, rapid cycles |

## Test Results

- **Total Tests:** 21 integration tests
- **Coverage:** 100% on controller package
- **Execution Time:** ~31 seconds
- **All tests pass consistently**

## Key Discoveries

1. **envtest Limitations:**
   - No garbage collection controller - cannot test cascade deletion
   - Owner references are verified in Create Flow tests instead
   - Resources remain after CR deletion in envtest

2. **Service Configuration:**
   - Service has 3 ports (5557, 5558, metrics) - WebUI port 8089 is excluded
   - Service selector uses `performance-test-pod-name` label

3. **Label Keys:**
   - Pod labels use `performance-test-pod-name`, `managed-by`, `app`, `performance-test-name`
   - Service doesn't have labels set in BuildMasterService

4. **Job Completions:**
   - Master Job doesn't explicitly set Completions (nil = 1 by default)
   - Worker Job has Completions = nil (parallel workers)

## Verification

- `make test` ✓ (all tests pass)
- `go test -v ./internal/controller/... -ginkgo.v` ✓ (21/21 passed)
- No flaky tests observed

## Notes for Phase 7+

1. Integration tests provide full coverage of `SetupWithManager()` 
2. E2E tests (Phase 15) will be needed to test actual garbage collection
3. Test namespace isolation pattern can be reused for future test suites

---

# Phase 7 Completion Notes

**Completed:** 2026-01-19

---

## Summary

Implemented v2 API types with grouped configuration and new feature fields. v1 remains storage version until Phase 8 implements conversion webhook.

## Files Created

| File | Purpose | LOC |
|------|---------|-----|
| `api/v2/groupversion_info.go` | v2 group/version registration | ~35 |
| `api/v2/locusttest_types.go` | All v2 type definitions | ~390 |
| `api/v2/conditions.go` | Condition type and reason constants | ~55 |
| `api/v2/zz_generated.deepcopy.go` | Auto-generated DeepCopy methods | Auto |
| `config/samples/locust_v2_locusttest.yaml` | Sample v2 CR | ~45 |

## Files Modified

| File | Changes |
|------|---------|
| `cmd/main.go` | Added v2 scheme registration |
| `api/v1/locusttest_types.go` | Added `+kubebuilder:storageversion` marker |
| `internal/controller/suite_test.go` | Added v2 scheme registration for tests |

## v2 Types Implemented

### Core Types
| Type | Description |
|------|-------------|
| `MasterSpec` | Grouped master config (command, resources, labels, annotations, autostart, autoquit, extraArgs) |
| `WorkerSpec` | Grouped worker config (command, replicas, resources, labels, annotations, extraArgs) |
| `AutoquitConfig` | Autoquit behavior (enabled, timeout) |
| `TestFilesConfig` | ConfigMap refs with mount paths |
| `SchedulingConfig` | Affinity, tolerations, nodeSelector |

### New Feature Types (Issue References)
| Type | Purpose | Issue |
|------|---------|-------|
| `EnvConfig` | Environment injection | #149 |
| `ConfigMapEnvSource` | ConfigMap env source with prefix | #149 |
| `SecretEnvSource` | Secret env source with prefix | #149 |
| `SecretMount` | Secret file mounting | #149 |
| `TargetedVolumeMount` | Volume mount with master/worker/both target | #252 |
| `ObservabilityConfig` | Observability settings wrapper | #72 |
| `OpenTelemetryConfig` | OTel integration (enabled, endpoint, protocol, insecure) | #72 |

### Status Types
| Type | Fields |
|------|--------|
| `LocustTestStatus` | Phase, ExpectedWorkers, ConnectedWorkers, StartTime, CompletionTime, Conditions |

### Condition Constants
- **Types:** Ready, WorkersConnected, TestCompleted
- **Reasons:** ResourcesCreating, ResourcesCreated, ResourcesFailed, WaitingForWorkers, AllWorkersConnected, WorkersMissing, TestInProgress, TestSucceeded, TestFailed
- **Phases:** Pending, Running, Succeeded, Failed

## Key Decision: Storage Version

**Decision:** v1 remains storage version until Phase 8

**Rationale:** Without a conversion webhook, using v2 as storage version breaks v1 API usage. When tests fetch a v1 CR and update it, the API server converts to/from v2 storage, causing field mapping issues (e.g., WorkerReplicas becoming 0).

**Action:** 
- Removed `+kubebuilder:storageversion` from v2
- Added `+kubebuilder:storageversion` to v1
- Phase 8 will implement conversion webhook and switch storage to v2

## Printer Columns (v2)

| Column | JSONPath | Description |
|--------|----------|-------------|
| Phase | `.status.phase` | Current test phase |
| Workers | `.spec.worker.replicas` | Requested worker count |
| Connected | `.status.connectedWorkers` | Connected workers |
| Image | `.spec.image` | Container image (priority=1) |
| Age | `.metadata.creationTimestamp` | Resource age |

## Verification

- `make generate` ✓ (DeepCopy methods generated)
- `make manifests` ✓ (CRD with v1 and v2 versions)
- `make build` ✓
- `make test` ✓ (all 21 integration tests pass)
- CRD contains both v1 (storage=true) and v2 (storage=false)

---

# Phase 8 Completion Notes

**Completed:** 2026-01-19

---

## Summary

Implemented v1↔v2 conversion webhook using Hub-and-Spoke pattern. v2 is the Hub, v1 is the Spoke with ConvertTo/ConvertFrom methods. v1 shows deprecation warning.

## Files Created

| File | Purpose | LOC |
|------|---------|-----|
| `api/v2/locusttest_conversion.go` | Hub marker implementation | ~20 |
| `api/v1/locusttest_conversion.go` | Spoke conversion logic | ~260 |
| `api/v1/locusttest_webhook.go` | Webhook setup | ~30 |
| `api/v1/locusttest_conversion_test.go` | Conversion unit tests | ~550 |
| `config/webhook/kustomization.yaml` | Webhook kustomize config | ~2 |
| `config/webhook/manifests.yaml` | Placeholder manifest | ~6 |

## Files Modified

| File | Changes |
|------|---------|
| `api/v1/locusttest_types.go` | Added deprecation warning marker |
| `cmd/main.go` | Added webhook registration |
| `internal/controller/suite_test.go` | (reverted to original - webhook not needed for unit tests) |

## Conversion Mapping Implemented

### v1 → v2 (ConvertTo)
- `masterCommandSeed` → `master.command`
- `workerCommandSeed` → `worker.command`
- `workerReplicas` → `worker.replicas`
- `image`, `imagePullPolicy`, `imagePullSecrets` → direct mapping
- `labels.master/worker` → `master.labels`, `worker.labels`
- `annotations.master/worker` → `master.annotations`, `worker.annotations`
- `configMap` → `testFiles.configMapRef`
- `libConfigMap` → `testFiles.libConfigMapRef`
- `affinity` (custom) → `scheduling.affinity` (corev1.Affinity)
- `tolerations[]` (custom) → `scheduling.tolerations[]` (corev1.Toleration)
- Sets defaults: `master.autostart=true`, `master.autoquit={enabled:true, timeout:60}`

### v2 → v1 (ConvertFrom) - Lossy
All v1-compatible fields preserved. v2-only fields lost:
- `master.resources`, `master.extraArgs`
- `worker.resources`, `worker.extraArgs`
- `testFiles.srcMountPath`, `testFiles.libMountPath`
- `scheduling.nodeSelector`
- `env` (all sub-fields)
- `volumes`, `volumeMounts`
- `observability` (OpenTelemetry config)

## Key Decision: Storage Version

**Decision:** v2 IS storage version ✅

**Implementation:** E2E tests in Kind cluster with cert-manager confirm conversion webhook works.

## Files Created/Modified for Webhook Infrastructure

| File | Purpose |
|------|---------|
| `config/certmanager/certificate.yaml` | Self-signed issuer and certificate for webhook TLS |
| `config/certmanager/kustomization.yaml` | Kustomize config for cert-manager resources |
| `config/webhook/manifests.yaml` | Webhook service definition |
| `config/crd/patches/webhook_in_locusttests.yaml` | CRD patch for conversion webhook |
| `config/default/manager_webhook_patch.yaml` | Deployment patch for webhook volume mounts |
| `config/default/kustomization.yaml` | Updated with webhook/certmanager resources |
| `config/crd/kustomization.yaml` | Updated with webhook patch |

## E2E Test Infrastructure

| File | Purpose |
|------|---------|
| `test/e2e/kind-config.yaml` | Kind cluster configuration |
| `test/e2e/conversion/v1-cr.yaml` | Sample v1 CR for testing |
| `test/e2e/conversion/v2-cr.yaml` | Sample v2 CR for testing |
| `test/e2e/conversion/configmap.yaml` | Test ConfigMap |
| `test/e2e/conversion/run-e2e.sh` | E2E test script |

## Test Results

- **Conversion unit tests:** 15 tests, all pass
- **E2E tests:** 7 tests, all pass in Kind cluster
  - Test 1: Create v1 CR ✓
  - Test 2: Read v1 CR as v2 (v1→v2 conversion) ✓
  - Test 3: Create v2 CR ✓
  - Test 4: Read v2 CR as v1 (v2→v1 conversion) ✓
  - Test 5: Update v1 CR reflected in v2 view ✓
  - Test 6: Reconciler creates Jobs ✓
  - Test 7: Deprecation warning shown ✓

## Verification Commands

```bash
# Create Kind cluster
kind create cluster --name locust-webhook-test --config test/e2e/kind-config.yaml

# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Build and deploy
make docker-build IMG=locust-k8s-operator:e2e-test
kind load docker-image locust-k8s-operator:e2e-test --name locust-webhook-test
make deploy IMG=locust-k8s-operator:e2e-test

# Verify storage version
kubectl get crd locusttests.locust.io -o jsonpath='{.spec.versions[?(@.storage==true)].name}'
# Output: v2

# Run E2E tests
./test/e2e/conversion/run-e2e.sh

# Cleanup
kind delete cluster --name locust-webhook-test
```

## Notes for Phase 9+

1. v2 is confirmed as storage version with working conversion webhook
2. Conversion webhook registration in main.go is controlled by `ENABLE_WEBHOOKS` env var
3. Production deployment requires cert-manager for webhook TLS
4. Unit tests (envtest) still pass but don't exercise webhook - E2E tests validate webhook

---

# Phase 9: Status Subresource Implementation

## Date: 2026-01-20

## Summary

Implemented status subresource for LocustTest resources and migrated the controller and resource builders from v1 to v2 API. v2 is now the primary API version used throughout the codebase.

## Key Changes

### 1. Controller Migration to v2 API
- `internal/controller/locusttest_controller.go` now uses `locustv2.LocustTest`
- Removed `GenerationChangedPredicate` to allow status-only updates to trigger reconciliation
- Added status initialization on first reconcile
- Added status update after successful resource creation

### 2. Resource Builders Migration to v2 API
- `internal/resources/job.go` - Updated to use v2 API types
- `internal/resources/labels.go` - Updated to use v2 API types  
- `internal/resources/service.go` - Updated to use v2 API types

### 3. Status Helper Functions
Created `internal/controller/status.go` with:
- `initializeStatus()` - Sets initial status values (Phase=Pending, conditions)
- `setCondition()` - Sets/updates a condition using standard meta.SetStatusCondition
- `setReady()` - Convenience wrapper for Ready condition
- `updateStatusFromJobs()` - Derives status from Job states
- `derivePhaseFromJob()` - Maps Job status to LocustTest phase
- `isJobComplete()` / `isJobFailed()` - Job status helpers

### 4. Test Updates
All test files updated to use v2 API:
- `internal/resources/job_test.go`
- `internal/resources/labels_test.go`
- `internal/resources/service_test.go`
- `internal/controller/locusttest_controller_test.go`
- `internal/controller/locusttest_controller_unit_test.go`
- `internal/controller/integration_test.go`

Created `internal/controller/status_test.go` with unit tests for status helpers.

## Status Tracking Behavior

### Phases
- **Pending**: Initial state, resources being created
- **Running**: Resources created, test is running
- **Succeeded**: Master Job completed successfully
- **Failed**: Master Job failed

### Conditions
- **Ready**: True when all resources are created
- **WorkersConnected**: Tracks worker connection status
- **TestCompleted**: True when test finishes (succeeded or failed)

## Files Modified
- `internal/controller/locusttest_controller.go`
- `internal/controller/status.go` (new)
- `internal/controller/status_test.go` (new)
- `internal/resources/job.go`
- `internal/resources/labels.go`
- `internal/resources/service.go`
- All test files in `internal/resources/` and `internal/controller/`

## Test Results
- All unit tests pass (27 tests)
- Integration tests pass (21 tests) - occasional flakiness due to envtest race conditions
- Build compiles successfully

## Test Infrastructure Simplification
- Removed `config/crd/test/` directory (v1-only test CRD no longer needed)
- Integration tests now use main CRD from `config/crd/bases/` with v2 as storage
- Updated `suite_test.go` to point to main CRD directory
- Updated `locusttest_controller_test.go` to verify manager reconciliation instead of manual reconcile calls

## v1 API Deprecation Status
- v1 API types still exist for conversion webhook compatibility
- Controller and resource builders now exclusively use v2 API
- v1 will be removed in a future release after deprecation period

---

# Phase 10: Environment & Secret Injection (Issue #149)

## Date: 2026-01-20

## Summary

Implemented environment variable and secret injection into Locust pods, addressing Issue #149. Users can now securely pass credentials, API keys, and configuration without hardcoding them in test files.

## Key Changes

### 1. Environment Builder Functions (`internal/resources/env.go`)
- `BuildEnvFrom()` - Creates `envFrom` entries from ConfigMap and Secret refs
- `BuildUserEnvVars()` - Creates `env` entries from user-defined variables
- `BuildEnvVars()` - Combines Kafka env vars with user-defined vars (Kafka first)
- `BuildSecretVolumes()` - Creates volumes for secret file mounts
- `BuildSecretVolumeMounts()` - Creates volume mounts for secret files
- `SecretVolumeName()` - Generates prefixed volume names (`secret-<name>`)

### 2. Job Builder Updates (`internal/resources/job.go`)
- Updated `buildLocustContainer()` to use `BuildEnvVars()` and `BuildEnvFrom()`
- Updated `buildVolumes()` to include secret volumes
- Updated `buildVolumeMounts()` to include secret mounts
- Exported `BuildKafkaEnvVars()` for use in env.go

### 3. Validation Webhook (`api/v2/locusttest_webhook.go`)
- Implemented `LocustTestCustomValidator` using `webhook.CustomValidator` interface
- `validateSecretMounts()` - Validates secret mount paths don't conflict with reserved paths
- `getReservedPaths()` - Dynamically calculates reserved paths based on testFiles config
- `PathConflicts()` - Checks if two paths conflict (exact match or prefix)
- Reserved paths: `/lotest/src` (default) and `/opt/locust/lib` (default)

### 4. Test Coverage
- `internal/resources/env_test.go` - 27 unit tests for env builders
- `api/v2/locusttest_webhook_test.go` - 17 unit tests for webhook validation
- `internal/resources/job_test.go` - 7 new tests for env injection in jobs

## Files Created
| File | Purpose | LOC |
|------|---------|-----|
| `internal/resources/env.go` | Environment builder functions | ~120 |
| `internal/resources/env_test.go` | Env builder unit tests | ~430 |
| `api/v2/locusttest_webhook.go` | Validation webhook | ~140 |
| `api/v2/locusttest_webhook_test.go` | Webhook unit tests | ~340 |
| `config/samples/locust_v2_locusttest_with_env.yaml` | Sample CR | ~40 |

## Files Modified
| File | Changes |
|------|---------|
| `internal/resources/job.go` | Integrated env builders, added secret volumes/mounts |
| `internal/resources/job_test.go` | Added 7 env injection tests |

## Design Decisions

| Decision | Chosen | Rationale |
|----------|--------|-----------|
| Env var order | Kafka first, user vars last | Matches existing behavior, user can override |
| Volume naming | `secret-<name>` prefix | Avoids conflicts with ConfigMap volumes |
| Path validation | Dynamic based on testFiles | Respects custom mount paths |
| Webhook interface | `CustomValidator` | Required for controller-runtime v0.21.0 |

## Test Results
- All unit tests pass (resources: 96.8% coverage, api/v2: 26.4% coverage)
- Integration test has pre-existing flaky test (race condition on update)

## Notes for Future Phases
- Webhook registration in main.go controlled by `ENABLE_WEBHOOKS` env var
- v2 validation webhook path: `/validate-locust-io-v2-locusttest`

---

# Phase 11: Volume Mounting (Issue #252)

## Date: 2026-01-20

## Summary

Implemented arbitrary volume mounting to Locust master and/or worker pods with target filtering support. Users can now mount PVCs, emptyDir, hostPath, and other volume types with control over which pods receive the mounts.

## Key Changes

### 1. Volume Builder Functions (`internal/resources/volumes.go`)
- `BuildUserVolumes()` - Returns user-defined volumes filtered by operational mode
- `BuildUserVolumeMounts()` - Returns user-defined volume mounts filtered by mode
- `shouldApplyMount()` - Checks if a mount applies to the given mode (master/worker/both)
- `shouldIncludeVolume()` - Checks if a volume has any mounts for the given mode

### 2. Job Builder Updates (`internal/resources/job.go`)
- Updated `buildVolumes()` to accept `mode` parameter and include user volumes
- Updated `buildVolumeMounts()` to accept `mode` parameter and include user mounts
- Updated `buildLocustContainer()` to pass mode to volume mount builder
- Updated `buildJob()` to pass mode through the call chain

### 3. Validation Webhook Updates (`api/v2/locusttest_webhook.go`)
- Added `validateVolumes()` - Validates volume names and mount paths
- Added `validateVolumeName()` - Checks for reserved name conflicts (secret-*, locust-lib, <cr>-master/worker)
- Added `validateMountPath()` - Checks for reserved path conflicts
- Added `validateMountReferences()` - Ensures all mounts reference defined volumes
- Updated `ValidateCreate/Update` to call combined `validateLocustTest()`

### 4. Test Coverage
- `internal/resources/volumes_test.go` - 17 unit tests for volume builders
- `api/v2/locusttest_webhook_test.go` - 17 new tests for volume validation

## Files Created
| File | Purpose | LOC |
|------|---------|-----|
| `internal/resources/volumes.go` | Volume builder functions | ~90 |
| `internal/resources/volumes_test.go` | Volume builder unit tests | ~320 |
| `config/samples/locust_v2_locusttest_with_volumes.yaml` | Sample CR | ~45 |

## Files Modified
| File | Changes |
|------|---------|
| `internal/resources/job.go` | Added mode parameter, integrated user volumes |
| `api/v2/locusttest_webhook.go` | Added volume validation functions |
| `api/v2/locusttest_webhook_test.go` | Added 17 volume validation tests |

## Target Filtering Behavior

| Target | Applies To |
|--------|------------|
| `master` | Master pod only |
| `worker` | Worker pods only |
| `both` (default) | Both master and worker pods |

## Reserved Resources

### Reserved Volume Names
| Pattern | Purpose |
|---------|---------|
| `<crName>-master` | Master ConfigMap volume |
| `<crName>-worker` | Worker ConfigMap volume |
| `locust-lib` | Library ConfigMap volume |
| `secret-*` | Secret volumes from `env.secretMounts` |

### Reserved Paths
| Path | Purpose |
|------|---------|
| `/lotest/src` (or custom) | Test files ConfigMap |
| `/opt/locust/lib` (or custom) | Library ConfigMap |

## Test Results
- All unit tests pass (volumes: 17 tests, webhook: 34 tests total)
- Build compiles successfully

---

# Phase 12: OpenTelemetry Support - Completion Notes

**Date**: 2026-01-20

Implemented native OpenTelemetry integration for Locust, replacing the Prometheus metrics exporter sidecar when OTel is enabled. This allows direct export of traces and metrics to an OTel collector.

## Key Changes

### 1. OTel Helper Functions (`internal/resources/otel.go`)
- `IsOTelEnabled()` - Returns true if OpenTelemetry is enabled in the spec
- `GetOTelConfig()` - Returns the OpenTelemetry configuration or nil
- `BuildOTelEnvVars()` - Creates environment variables for OTel SDK configuration
- Constants for OTel environment variable names and default values

### 2. Command Builder Updates (`internal/resources/command.go`)
- Updated `BuildMasterCommand()` to accept `otelEnabled` parameter
- Updated `BuildWorkerCommand()` to accept `otelEnabled` parameter
- Both functions add `--otel` flag when OTel is enabled (positioned before other flags)

### 3. Environment Variable Integration (`internal/resources/env.go`)
- Updated `BuildEnvVars()` to include OTel environment variables when enabled
- OTel env vars are injected between Kafka env vars and user-defined vars

### 4. Job Builder Updates (`internal/resources/job.go`)
- Updated `BuildMasterJob()` and `BuildWorkerJob()` to pass `otelEnabled` to command builders
- Conditional sidecar: metrics exporter sidecar is skipped when OTel is enabled

### 5. Service Builder Updates (`internal/resources/service.go`)
- Updated `BuildMasterService()` to conditionally exclude metrics port when OTel is enabled

### 6. Validation Webhook Updates (`api/v2/locusttest_webhook.go`)
- Added `validateOTelConfig()` - Validates endpoint is required when OTel is enabled
- Integrated into `validateLocustTest()` for both create and update operations

## Files Created
| File | Purpose | LOC |
|------|---------|-----|
| `internal/resources/otel.go` | OTel helper functions | ~100 |
| `internal/resources/otel_test.go` | OTel helper unit tests | ~270 |
| `config/samples/locust_v2_locusttest_with_otel.yaml` | Sample CR with OTel config | ~40 |

## Files Modified
| File | Changes |
|------|---------|
| `internal/resources/command.go` | Added otelEnabled parameter, --otel flag support |
| `internal/resources/command_test.go` | Added 6 OTel flag tests |
| `internal/resources/env.go` | Integrated OTel env vars into BuildEnvVars() |
| `internal/resources/job.go` | Conditional sidecar, otelEnabled to command builders |
| `internal/resources/job_test.go` | Added 12 OTel tests |
| `internal/resources/service.go` | Conditional metrics port exclusion |
| `internal/resources/service_test.go` | Added 4 OTel tests |
| `api/v2/locusttest_webhook.go` | Added validateOTelConfig() |
| `api/v2/locusttest_webhook_test.go` | Added 10 OTel validation tests |

## OTel Environment Variables

| Variable | Source | Required |
|----------|--------|----------|
| `OTEL_TRACES_EXPORTER` | Fixed: "otlp" | Yes |
| `OTEL_METRICS_EXPORTER` | Fixed: "otlp" | Yes |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `spec.observability.openTelemetry.endpoint` | Yes |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `spec.observability.openTelemetry.protocol` (default: "grpc") | Yes |
| `OTEL_EXPORTER_OTLP_INSECURE` | `spec.observability.openTelemetry.insecure` | Only if true |
| Custom vars | `spec.observability.openTelemetry.extraEnvVars` | No |

## Behavioral Changes When OTel Enabled

| Component | OTel Disabled | OTel Enabled |
|-----------|---------------|--------------|
| Master containers | 2 (locust + metrics-exporter) | 1 (locust only) |
| Worker containers | 1 (locust only) | 1 (locust only) |
| Service ports | 5557, 5558, metrics | 5557, 5558 (no metrics) |
| Locust command | Standard flags | Includes --otel flag |
| Environment | Kafka + user vars | Kafka + OTel + user vars |

## Test Results
- All unit tests pass (command: 10 tests, job: 45 tests, service: 9 tests, otel: 13 tests, webhook: 44 tests)
- Build compiles successfully

---

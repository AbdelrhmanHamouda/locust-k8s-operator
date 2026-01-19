# Phase 7: v2 API Types - Checklist

**Estimated Effort:** 1.5 days  
**Status:** ✅ Complete

---

## Pre-Implementation

- [x] Phase 4 complete (Core Reconciler working)
- [x] Review `CRD_API_DESIGN.md` for complete type definitions
- [x] Review `REQUIREMENTS.md` §4.2 for v2 spec requirements
- [x] Verify existing v1 API in `api/v1/locusttest_types.go`
- [x] Run `make build && make test` to ensure clean starting state

---

## Day 1: API Scaffold & Core Types

### Task 7.1: Create v2 API Scaffold

```bash
operator-sdk create api \
  --group locust \
  --version v2 \
  --kind LocustTest \
  --resource \
  --controller=false
```

- [x] v2 directory created: `api/v2/`
- [x] `api/v2/groupversion_info.go` exists
- [x] `api/v2/locusttest_types.go` exists

---

### Task 7.2: Update `api/v2/groupversion_info.go`

**File:** `api/v2/groupversion_info.go`

- [x] Package doc comment has `+kubebuilder:object:generate=true`
- [x] Package doc comment has `+groupName=locust.io`
- [x] `GroupVersion` set to `{Group: "locust.io", Version: "v2"}`
- [x] `SchemeBuilder` and `AddToScheme` exported

---

### Task 7.3: Define Core Spec Types

**File:** `api/v2/locusttest_types.go`

#### MasterSpec
- [x] `Command` field (required)
- [x] `Resources` field (corev1.ResourceRequirements)
- [x] `Labels` field (map[string]string)
- [x] `Annotations` field (map[string]string)
- [x] `Autostart` field (*bool, default true)
- [x] `Autoquit` field (*AutoquitConfig)
- [x] `ExtraArgs` field ([]string)

#### AutoquitConfig
- [x] `Enabled` field (bool, default true)
- [x] `Timeout` field (int32, default 60, min 0)

#### WorkerSpec
- [x] `Command` field (required)
- [x] `Replicas` field (required, min 1, max 500)
- [x] `Resources` field (corev1.ResourceRequirements)
- [x] `Labels` field (map[string]string)
- [x] `Annotations` field (map[string]string)
- [x] `ExtraArgs` field ([]string)

#### TestFilesConfig
- [x] `ConfigMapRef` field
- [x] `LibConfigMapRef` field
- [x] `SrcMountPath` field (default "/lotest/src")
- [x] `LibMountPath` field (default "/opt/locust/lib")

#### SchedulingConfig
- [x] `Affinity` field (*corev1.Affinity)
- [x] `Tolerations` field ([]corev1.Toleration)
- [x] `NodeSelector` field (map[string]string)

---

## Day 2: New Feature Types & Validation

### Task 7.4: Define New Feature Types

**File:** `api/v2/locusttest_types.go`

#### EnvConfig (Issue #149)
- [x] `ConfigMapRefs` field ([]ConfigMapEnvSource)
- [x] `SecretRefs` field ([]SecretEnvSource)
- [x] `Variables` field ([]corev1.EnvVar)
- [x] `SecretMounts` field ([]SecretMount)

#### ConfigMapEnvSource
- [x] `Name` field (required)
- [x] `Prefix` field (optional)

#### SecretEnvSource
- [x] `Name` field (required)
- [x] `Prefix` field (optional)

#### SecretMount
- [x] `Name` field (required)
- [x] `MountPath` field (required)
- [x] `ReadOnly` field (default true)

#### TargetedVolumeMount (Issue #252)
- [x] Embedded `corev1.VolumeMount`
- [x] `Target` field (enum: master/worker/both, default both)

#### ObservabilityConfig (Issue #72)
- [x] `OpenTelemetry` field (*OpenTelemetryConfig)

#### OpenTelemetryConfig
- [x] `Enabled` field (bool, default false)
- [x] `Endpoint` field (string)
- [x] `Protocol` field (enum: grpc/http/protobuf, default grpc)
- [x] `Insecure` field (bool, default false)
- [x] `ExtraEnvVars` field (map[string]string)

---

### Task 7.5: Define LocustTestStatus

**File:** `api/v2/locusttest_types.go`

- [x] `Phase` field (enum: Pending/Running/Succeeded/Failed)
- [x] `ExpectedWorkers` field (int32)
- [x] `ConnectedWorkers` field (int32)
- [x] `StartTime` field (*metav1.Time)
- [x] `CompletionTime` field (*metav1.Time)
- [x] `Conditions` field ([]metav1.Condition with patchStrategy)

---

### Task 7.6: Define LocustTestSpec

**File:** `api/v2/locusttest_types.go`

- [x] `Image` field (required)
- [x] `ImagePullPolicy` field (enum, default IfNotPresent)
- [x] `ImagePullSecrets` field ([]corev1.LocalObjectReference)
- [x] `Master` field (MasterSpec, required)
- [x] `Worker` field (WorkerSpec, required)
- [x] `TestFiles` field (*TestFilesConfig)
- [x] `Scheduling` field (*SchedulingConfig)
- [x] `Env` field (*EnvConfig)
- [x] `Volumes` field ([]corev1.Volume)
- [x] `VolumeMounts` field ([]TargetedVolumeMount)
- [x] `Observability` field (*ObservabilityConfig)

---

### Task 7.7: Add Kubebuilder Markers

**File:** `api/v2/locusttest_types.go`

- [x] `+kubebuilder:object:root=true`
- [x] `+kubebuilder:subresource:status`
- [x] `+kubebuilder:resource:shortName=lotest`
- [x] `+kubebuilder:storageversion` (deferred to Phase 8 - v1 remains storage)
- [x] PrintColumn: Phase
- [x] PrintColumn: Workers (spec.worker.replicas)
- [x] PrintColumn: Connected (status.connectedWorkers)
- [x] PrintColumn: Image (priority=1)
- [x] PrintColumn: Age

---

### Task 7.8: Create Conditions Constants

**File:** `api/v2/conditions.go`

Condition Types:
- [x] `ConditionTypeReady`
- [x] `ConditionTypeWorkersConnected`
- [x] `ConditionTypeTestCompleted`

Ready Reasons:
- [x] `ReasonResourcesCreating`
- [x] `ReasonResourcesCreated`
- [x] `ReasonResourcesFailed`

WorkersConnected Reasons:
- [x] `ReasonWaitingForWorkers`
- [x] `ReasonAllWorkersConnected`
- [x] `ReasonWorkersMissing`

TestCompleted Reasons:
- [x] `ReasonTestInProgress`
- [x] `ReasonTestSucceeded`
- [x] `ReasonTestFailed`

Phase Constants:
- [x] `PhasePending`
- [x] `PhaseRunning`
- [x] `PhaseSucceeded`
- [x] `PhaseFailed`

---

### Task 7.9: Update cmd/main.go

**File:** `cmd/main.go`

- [x] Import `locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"`
- [x] Add `utilruntime.Must(locustv2.AddToScheme(scheme))`

---

### Task 7.10: Create Sample v2 CR

**File:** `config/samples/locust_v2_locusttest.yaml`

- [x] Valid v2 CR with all major sections
- [x] Master config with resources
- [x] Worker config with replicas and resources
- [x] TestFiles config
- [x] Scheduling config (nodeSelector)

---

## Verification

### Code Generation

```bash
make generate
make manifests
```

- [x] `api/v2/zz_generated.deepcopy.go` generated
- [x] `config/crd/bases/locust.io_locusttests.yaml` updated
- [x] CRD contains both v1 and v2 versions
- [x] v1 marked as `storageVersion: true` (v2 deferred to Phase 8)

### Build & Test

```bash
make build
make test
```

- [x] Project builds without errors
- [x] All existing tests still pass

### CRD Validation

```bash
kubectl apply --dry-run=client -f config/samples/locust_v2_locusttest.yaml
```

- [x] Sample v2 CR validates against schema
- [x] Required field validation works
- [x] Enum field validation works

### Printer Columns

After applying CR to cluster:

```bash
kubectl get locusttests
```

- [x] Phase column shows
- [x] Workers column shows spec.worker.replicas
- [x] Connected column shows status.connectedWorkers
- [x] Age column shows

---

## Post-Implementation

- [x] All verification steps pass
- [x] Update `phases/README.md` with Phase 7 status
- [x] Update `phases/NOTES.md` with any deviations
- [x] Document any decisions made during implementation

---

## Files Summary

| File | Action | Est. LOC |
|------|--------|----------|
| `api/v2/groupversion_info.go` | Create/Update | ~30 |
| `api/v2/locusttest_types.go` | Create | ~350 |
| `api/v2/conditions.go` | Create | ~50 |
| `api/v2/zz_generated.deepcopy.go` | Generated | Auto |
| `cmd/main.go` | Modify | +3 |
| `config/samples/locust_v2_locusttest.yaml` | Create | ~50 |

---

## Quick Reference Commands

```bash
# Create v2 API scaffold
operator-sdk create api --group locust --version v2 --kind LocustTest --resource --controller=false

# Generate DeepCopy
make generate

# Generate CRD manifests
make manifests

# Build project
make build

# Run tests
make test

# Validate sample CR
kubectl apply --dry-run=client -f config/samples/locust_v2_locusttest.yaml

# Check CRD versions
kubectl get crd locusttests.locust.io -o yaml | grep -A10 "versions:"

# View generated CRD
cat config/crd/bases/locust.io_locusttests.yaml | head -100
```

---

## Acceptance Criteria Summary

1. **v2 Types Defined:** All types from CRD_API_DESIGN.md §2 implemented
2. **Storage Version:** v2 marked as storage version
3. **Markers Applied:** All kubebuilder validation and printer markers present
4. **Code Generates:** `make generate && make manifests` succeeds
5. **Project Builds:** `make build` succeeds
6. **Tests Pass:** `make test` shows all tests passing
7. **Sample Validates:** v2 sample CR passes schema validation

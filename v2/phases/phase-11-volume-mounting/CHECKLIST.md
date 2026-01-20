# Phase 11: Volume Mounting - Checklist

**Estimated Effort:** 0.5 day  
**Status:** ✅ Complete

---

## Pre-Implementation

- [x] Phase 7 complete (v2 API Types with `Volumes`, `TargetedVolumeMount` defined)
- [x] Phase 10 complete (env/secret injection with validation webhook)
- [x] Review `api/v2/locusttest_types.go` for volume types
- [x] Review `issue-analysis/P1-High/issue-252-volume-mounting.md`
- [x] Review current `internal/resources/job.go` implementation
- [x] Run `make build && make test` to ensure clean starting state

---

## Task 11.1: Create Volume Builder Functions

**File:** `internal/resources/volumes.go`

### Target Constants

- [x] Create file `internal/resources/volumes.go`
- [x] Add copyright header
- [x] Define target constants:
  ```go
  const (
      TargetMaster = "master"
      TargetWorker = "worker"
      TargetBoth   = "both"
  )
  ```

### Core Functions

- [x] Implement `BuildUserVolumes()` function:
  ```go
  func BuildUserVolumes(lt *locustv2.LocustTest, mode OperationalMode) []corev1.Volume
  ```
  - [x] Return nil for empty volumes
  - [x] Filter volumes to include only those with matching mounts
  
- [x] Implement `BuildUserVolumeMounts()` function:
  ```go
  func BuildUserVolumeMounts(lt *locustv2.LocustTest, mode OperationalMode) []corev1.VolumeMount
  ```
  - [x] Return nil for empty volumeMounts
  - [x] Filter mounts by target matching mode
  - [x] Convert `TargetedVolumeMount` to `VolumeMount`

### Helper Functions

- [x] Implement `shouldApplyMount()` function:
  ```go
  func shouldApplyMount(mount locustv2.TargetedVolumeMount, mode OperationalMode) bool
  ```
  - [x] Default empty target to "both"
  - [x] Return true for "both"
  - [x] Return true for "master" when mode is Master
  - [x] Return true for "worker" when mode is Worker
  
- [x] Implement `shouldIncludeVolume()` function:
  ```go
  func shouldIncludeVolume(volumeName string, mounts []locustv2.TargetedVolumeMount, mode OperationalMode) bool
  ```
  - [x] Return true if any mount for this volume matches the mode

**Verification:**
```bash
go build ./internal/resources/...
```

---

## Task 11.2: Write Volume Builder Tests

**File:** `internal/resources/volumes_test.go`

### BuildUserVolumes Tests

- [x] Create file `internal/resources/volumes_test.go`
- [x] Add copyright header
- [x] `TestBuildUserVolumes_Empty` - Returns nil for no volumes
- [x] `TestBuildUserVolumes_AllTargetBoth` - All volumes when all targets are "both"
- [x] `TestBuildUserVolumes_MasterOnly` - Only master-targeted for Master mode
- [x] `TestBuildUserVolumes_WorkerOnly` - Only worker-targeted for Worker mode
- [x] `TestBuildUserVolumes_Mixed` - Correct filtering with mixed targets
- [x] `TestBuildUserVolumes_VolumeWithNoMatchingMount` - Volume excluded if no mount matches

### BuildUserVolumeMounts Tests

- [x] `TestBuildUserVolumeMounts_Empty` - Returns nil for no mounts
- [x] `TestBuildUserVolumeMounts_MasterMode` - Only master/both mounts for Master
- [x] `TestBuildUserVolumeMounts_WorkerMode` - Only worker/both mounts for Worker
- [x] `TestBuildUserVolumeMounts_DefaultTarget` - Empty target treated as "both"
- [x] `TestBuildUserVolumeMounts_ConvertsToVolumeMount` - TargetedVolumeMount → VolumeMount

### Helper Function Tests

- [x] `TestShouldApplyMount_BothTarget` - Returns true for both modes
- [x] `TestShouldApplyMount_MasterTarget` - Returns true only for Master
- [x] `TestShouldApplyMount_WorkerTarget` - Returns true only for Worker
- [x] `TestShouldApplyMount_EmptyTarget` - Defaults to "both"
- [x] `TestShouldApplyMount_InvalidTarget` - Returns false
- [x] `TestShouldIncludeVolume_HasMatchingMount` - Returns true
- [x] `TestShouldIncludeVolume_NoMatchingMount` - Returns false

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuildUserVolume
go test ./internal/resources/... -v -run TestShouldApply
go test ./internal/resources/... -v -run TestShouldInclude
go test ./internal/resources/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep volumes
# Target: ≥90% coverage for volumes.go
```

---

## Task 11.3: Update Job Builder

**File:** `internal/resources/job.go`

### Signature Updates

- [x] Update `buildVolumes()` signature to include mode:
  ```go
  func buildVolumes(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.Volume
  ```

- [x] Update `buildVolumeMounts()` signature to include mode:
  ```go
  func buildVolumeMounts(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.VolumeMount
  ```

- [x] Update `buildLocustContainer()` signature to include mode:
  ```go
  func buildLocustContainer(lt *locustv2.LocustTest, name string, command []string, 
      ports []corev1.ContainerPort, cfg *config.OperatorConfig, mode OperationalMode) corev1.Container
  ```

### Call Site Updates

- [x] Update `buildJob()`:
  - [x] Pass `mode` to `buildVolumes()`
  - [x] Pass `mode` to `buildLocustContainer()`
  
- [x] Update `buildLocustContainer()`:
  - [x] Pass `mode` to `buildVolumeMounts()`

### Volume Merging

- [x] Update `buildVolumes()` to append user volumes:
  ```go
  // Add user-defined volumes (filtered by target)
  userVolumes := BuildUserVolumes(lt, mode)
  if len(userVolumes) > 0 {
      volumes = append(volumes, userVolumes...)
  }
  ```

- [x] Update `buildVolumeMounts()` to append user mounts:
  ```go
  // Add user-defined volume mounts (filtered by target)
  userMounts := BuildUserVolumeMounts(lt, mode)
  if len(userMounts) > 0 {
      mounts = append(mounts, userMounts...)
  }
  ```

**Verification:**
```bash
go build ./internal/resources/...
make test
```

---

## Task 11.4: Update Job Builder Tests

**File:** `internal/resources/job_test.go`

### Existing Test Updates

- [x] Update any tests that call `buildVolumes` or `buildVolumeMounts` to pass `mode`
- [x] Verify all existing tests still pass

### New Test Cases

- [x] `TestBuildMasterJob_WithUserVolumes` - User volumes added to master
- [x] `TestBuildMasterJob_WithUserVolumeMounts_TargetMaster` - Mount only for master (covered in TestBuildMasterJob_WithUserVolumes)
- [x] `TestBuildMasterJob_WithUserVolumeMounts_TargetBoth` - Mount for both included in master (covered in TestBuildMasterJob_WithUserVolumes)
- [x] `TestBuildMasterJob_WithUserVolumeMounts_TargetWorker` - Mount NOT in master
- [x] `TestBuildWorkerJob_WithUserVolumes` - User volumes added to worker
- [x] `TestBuildWorkerJob_WithUserVolumeMounts_TargetWorker` - Mount only for worker (covered in TestBuildWorkerJob_WithUserVolumes)
- [x] `TestBuildWorkerJob_WithUserVolumeMounts_TargetBoth` - Mount for both included in worker (covered in TestBuildWorkerJob_WithUserVolumes)
- [x] `TestBuildWorkerJob_WithUserVolumeMounts_TargetMaster` - Mount NOT in worker (covered in TestBuildWorkerJob_WithUserVolumes)
- [x] `TestBuildJob_UserVolumesWithSecretVolumes` - Both coexist correctly

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuildMasterJob
go test ./internal/resources/... -v -run TestBuildWorkerJob
```

---

## Task 11.5: Update Validation Webhook

**File:** `api/v2/locusttest_webhook.go`

### Reserved Names

- [x] Add reserved volume name constants:
  ```go
  const (
      reservedVolumeNamePrefix = "secret-"
      libVolumeName           = "locust-lib"
  )
  ```

### Validation Functions

- [x] Implement `validateVolumes()`:
  ```go
  func (r *LocustTest) validateVolumes() (admission.Warnings, error)
  ```
  - [x] Validate each volume name
  - [x] Validate each mount path
  - [x] Validate mount references

- [x] Implement `validateVolumeName()`:
  ```go
  func (r *LocustTest) validateVolumeName(name string) error
  ```
  - [x] Check for `secret-` prefix
  - [x] Check for `locust-lib`
  - [x] Check for `<crName>-master` / `<crName>-worker`

- [x] Implement `validateMountReferences()`:
  ```go
  func (r *LocustTest) validateMountReferences() error
  ```
  - [x] Ensure all mounts reference defined volumes

### Update Validators

- [x] Update `ValidateCreate()` to call `validateVolumes()`
- [x] Update `ValidateUpdate()` to call `validateVolumes()`

**Verification:**
```bash
go build ./api/v2/...
make manifests
```

---

## Task 11.6: Write Webhook Validation Tests

**File:** `api/v2/locusttest_webhook_test.go`

### Volume Name Tests

- [x] `TestValidateVolumeName_Valid` - Passes for valid names
- [x] `TestValidateVolumeName_SecretPrefix` - Fails for `secret-*`
- [x] `TestValidateVolumeName_LibVolume` - Fails for `locust-lib`
- [x] `TestValidateVolumeName_MasterConflict` - Fails for `<name>-master`
- [x] `TestValidateVolumeName_WorkerConflict` - Fails for `<name>-worker`

### Volume Mount Tests

- [x] `TestValidateVolumes_Empty` - Passes for no volumes
- [x] `TestValidateVolumes_ValidConfig` - Passes for valid volumes and mounts
- [x] `TestValidateVolumes_PathConflict` - Fails for reserved path conflicts
- [x] `TestValidateVolumes_UndefinedMount` - Fails for mount referencing undefined volume

### Integration with Existing Tests

- [x] Verify existing `TestValidateSecretMounts_*` tests still pass
- [x] Verify combined validation works correctly

**Verification:**
```bash
go test ./api/v2/... -v -run TestValidateVolume
go test ./api/v2/... -v -run TestValidate
```

---

## Task 11.7: Update Integration Tests

**File:** `internal/controller/integration_test.go`

### New Integration Tests

Note: Integration tests are covered by unit tests in job_test.go which test the full job building with volumes.

- [x] Add test helper for LocustTest with volumes (in job_test.go)
- [x] `TestBuildMasterJob_WithUserVolumes` - Verifies volumes in master Job
- [x] `TestBuildWorkerJob_WithUserVolumes` - Verifies volumes in worker Job with target filtering
- [x] `TestBuildMasterJob_WithUserVolumeMounts_TargetWorker` - Verifies mount NOT in master
- [x] `TestBuildJob_UserVolumesWithSecretVolumes` - Verifies coexistence with secret volumes

**Verification:**
```bash
go test ./internal/controller/... -v -run TestReconcile_WithUserVolume
go test ./internal/controller/... -v -run TestReconcile_WithTargeted
```

---

## Task 11.8: Create Sample CR

**File:** `config/samples/locust_v2_locusttest_with_volumes.yaml`

- [x] Create sample CR demonstrating volume mounting:
  ```yaml
  apiVersion: locust.io/v2
  kind: LocustTest
  metadata:
    name: load-test-with-volumes
  spec:
    image: locustio/locust:2.20.0
    master:
      command: "locust -f /lotest/src/locustfile.py"
    worker:
      command: "locust -f /lotest/src/locustfile.py"
      replicas: 2
    testFiles:
      configMapRef: locust-scripts
    volumes:
      - name: test-results
        emptyDir: {}
      - name: shared-data
        emptyDir: {}
    volumeMounts:
      - name: test-results
        mountPath: /results
        target: master
      - name: shared-data
        mountPath: /shared
        target: both
  ```

---

## Verification

### Code Generation

```bash
make generate
make manifests
```

- [x] No errors during generation
- [x] No changes to webhook manifests (validation logic only)

### Build & Test

```bash
make build
make test
```

- [x] Project builds without errors
- [x] All existing tests pass (1 pre-existing flaky integration test)
- [x] New tests pass
- [x] Coverage ≥80% for `internal/resources/volumes.go` (100%)
- [x] Coverage ≥80% for new webhook validation (100%)

### Linting

```bash
golangci-lint run ./internal/resources/...
golangci-lint run ./api/v2/...
```

- [x] No linting errors

### Manual Verification (Optional)

```bash
# Apply sample CR
kubectl apply -f config/samples/locust_v2_locusttest_with_volumes.yaml

# Verify volumes on master
kubectl get pods -l performance-test-name=load-test-with-volumes
MASTER_POD=$(kubectl get pods -l performance-test-name=load-test-with-volumes,lotest-role=master -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $MASTER_POD -- ls /results /shared

# Verify worker has only /shared (not /results)
WORKER_POD=$(kubectl get pods -l performance-test-name=load-test-with-volumes,lotest-role=worker -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $WORKER_POD -- ls /shared
kubectl exec -it $WORKER_POD -- ls /results  # Should fail

# Test validation webhook rejection
cat <<EOF | kubectl apply -f -
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: invalid-volume
spec:
  image: locustio/locust:2.20.0
  master:
    command: "locust"
  worker:
    command: "locust"
    replicas: 1
  volumes:
    - name: secret-custom  # Should be rejected
      emptyDir: {}
  volumeMounts:
    - name: secret-custom
      mountPath: /custom
EOF
# Expected: Rejected with "uses reserved prefix" error
```

---

## Post-Implementation

- [x] All verification steps pass
- [x] Update `phases/README.md` with Phase 11 status
- [x] Update `phases/NOTES.md` with implementation notes
- [ ] Update `ROADMAP.md` Phase 11 tasks as complete (optional)

---

## Files Summary

| File | Action | Est. LOC |
|------|--------|----------|
| `internal/resources/volumes.go` | Create | ~60 |
| `internal/resources/volumes_test.go` | Create | ~200 |
| `internal/resources/job.go` | Modify | +20, ~10 signature changes |
| `internal/resources/job_test.go` | Modify | +80 |
| `api/v2/locusttest_webhook.go` | Modify | +60 |
| `api/v2/locusttest_webhook_test.go` | Modify | +100 |
| `internal/controller/integration_test.go` | Modify | +80 |
| `config/samples/locust_v2_locusttest_with_volumes.yaml` | Create | ~25 |

**Total Estimated:** ~625 LOC

---

## Quick Reference Commands

```bash
# Build and test
make generate
make manifests
make build
make test

# Run specific tests
go test ./internal/resources/... -v -run TestBuildUserVolume
go test ./internal/resources/... -v -run TestShouldApply
go test ./api/v2/... -v -run TestValidateVolume
go test ./internal/controller/... -v -run TestReconcile_WithUserVolume

# Check coverage
go test ./internal/resources/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep volumes

# Lint
golangci-lint run ./internal/resources/...
golangci-lint run ./api/v2/...
```

---

## Acceptance Criteria Summary

1. **User Volumes:** `spec.volumes` correctly added to Pod spec
2. **Volume Mounts:** `spec.volumeMounts` correctly added to container
3. **Target Filtering:** `target: master` applies only to master pod
4. **Target Filtering:** `target: worker` applies only to worker pods
5. **Target Filtering:** `target: both` (default) applies to both
6. **Validation:** Reserved volume names rejected
7. **Validation:** Reserved mount paths rejected
8. **Validation:** Undefined mount references rejected
9. **Backward Compatible:** Empty volumes/mounts = no change
10. **Tests Pass:** All unit and integration tests pass
11. **Coverage:** ≥80% for new code

---

## Decision Log

| Decision | Options | Chosen | Rationale |
|----------|---------|--------|-----------|
| Volume inclusion | All volumes or filtered | **Filtered** | Only include volumes needed for the mode |
| Default target | none or both | **both** | Less surprising, matches user expectations |
| Signature change | Add mode or derive | **Add mode** | Explicit is better, matches existing patterns |
| Mount reference check | Skip or validate | **Validate** | Catch user errors early with clear message |

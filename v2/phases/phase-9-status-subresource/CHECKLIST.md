# Phase 9: Status Subresource - Checklist

**Estimated Effort:** 1 day  
**Status:** ✅ Complete

---

## Pre-Implementation

- [x] Phase 7 complete (v2 API Types defined with `LocustTestStatus`)
- [x] Review `CRD_API_DESIGN.md` §5 for status patterns
- [x] Review `OPERATOR_SDK_PATTERNS.md` §5 for status management
- [x] Review `REQUIREMENTS.md` §5.1.4 for acceptance criteria
- [x] Verify v2 types have `+kubebuilder:subresource:status` marker
- [x] Run `make build && make test` to ensure clean starting state

---

## Task 9.1: Define Condition Constants

**File:** `api/v2/conditions.go`

### Condition Types

- [x] Create file `api/v2/conditions.go`
- [x] Add copyright header
- [x] Define `ConditionTypeReady` constant
- [x] Define `ConditionTypeWorkersConnected` constant  
- [x] Define `ConditionTypeTestCompleted` constant

### Condition Reasons

- [x] Define Ready condition reasons:
  - `ReasonResourcesCreating` - Resources are being created
  - `ReasonResourcesCreated` - All resources created successfully
  - `ReasonResourcesFailed` - Resource creation failed
- [x] Define WorkersConnected condition reasons:
  - `ReasonWaitingForWorkers` - Waiting for workers to connect
  - `ReasonAllWorkersConnected` - All expected workers connected
  - `ReasonWorkersMissing` - Some workers failed to connect
- [x] Define TestCompleted condition reasons:
  - `ReasonTestInProgress` - Test is currently running
  - `ReasonTestSucceeded` - Test completed successfully
  - `ReasonTestFailed` - Test failed

### Phase Constants

- [x] Define `PhasePending` - Initial state
- [x] Define `PhaseRunning` - Test is executing
- [x] Define `PhaseSucceeded` - Test completed successfully
- [x] Define `PhaseFailed` - Test failed

**Verification:**
```bash
go build ./api/v2/...
```

---

## Task 9.2: Create Status Helper Functions

**File:** `internal/controller/status.go`

### Core Functions

- [x] Create file `internal/controller/status.go`
- [x] Add copyright header
- [x] Implement `updatePhase()` helper:
  ```go
  func (r *LocustTestReconciler) updatePhase(ctx context.Context, lt *locustv2.LocustTest, phase string) error
  ```
- [x] Implement `setCondition()` helper:
  ```go
  func (r *LocustTestReconciler) setCondition(lt *locustv2.LocustTest, condType, reason, message string, status metav1.ConditionStatus)
  ```
- [x] Implement `setReady()` convenience helper:
  ```go
  func (r *LocustTestReconciler) setReady(lt *locustv2.LocustTest, ready bool, reason, message string)
  ```
- [x] Implement `updateStatusFromJobs()` to derive status from Job states:
  ```go
  func (r *LocustTestReconciler) updateStatusFromJobs(ctx context.Context, lt *locustv2.LocustTest) error
  ```

### Helper Utilities

- [x] Implement `getJobPhase()` to extract phase from Job status (named `derivePhaseFromJob`)
- [x] Implement `isJobComplete()` to check Job completion
- [x] Implement `isJobFailed()` to check Job failure

**Verification:**
```bash
go build ./internal/controller/...
```

---

## Task 9.3: Update Reconciler for Initial Status

**File:** `internal/controller/locusttest_controller.go`

### On CR Creation

- [x] After fetching CR, check if status needs initialization
- [x] Set initial phase to `Pending` if phase is empty
- [x] Set `ExpectedWorkers` from `spec.worker.replicas` (v2) or `spec.workerReplicas` (v1)
- [x] Set initial `Ready` condition to `False` with reason `ResourcesCreating`
- [x] Update status before creating resources

### After Resource Creation

- [x] Update phase to `Running` after successful resource creation
- [x] Set `Ready` condition to `True` with reason `ResourcesCreated`
- [x] Set `StartTime` to current time
- [x] Record event for status transition

**Code Pattern:**
```go
// Initialize status on first reconcile
if lt.Status.Phase == "" {
    lt.Status.Phase = locustv2.PhasePending
    lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas
    r.setReady(lt, false, locustv2.ReasonResourcesCreating, "Creating resources")
    if err := r.Status().Update(ctx, lt); err != nil {
        return ctrl.Result{}, err
    }
}
```

**Verification:**
```bash
go build ./...
make test
```

---

## Task 9.4: Add Job Status Watching

**File:** `internal/controller/locusttest_controller.go`

### Watch Configuration

- [x] Verify `Owns(&batchv1.Job{})` is in `SetupWithManager()`
- [x] Add predicate to filter Job status updates (not just creation)

### Status Update on Job Changes

- [x] Implement logic to update phase based on master Job status:
  - Job pending → Phase `Pending`
  - Job running → Phase `Running`
  - Job succeeded → Phase `Succeeded`
  - Job failed → Phase `Failed`
- [x] Update `TestCompleted` condition when master Job finishes
- [x] Set `CompletionTime` when master Job completes

**Code Pattern:**
```go
// In Reconcile, after checking if CR exists
if lt.Generation > 1 && lt.Status.Phase != "" {
    // This is a status-only update from Job changes
    return r.updateStatusFromJobs(ctx, lt)
}
```

**Note:** The `GenerationChangedPredicate` may need adjustment to allow status updates from Job changes. Consider using a custom predicate or removing it for status tracking.

---

## Task 9.5: Handle v1 API Status (Compatibility)

**File:** `internal/controller/locusttest_controller.go`

### v1 Considerations

- [x] Current controller uses v1 API (`locustv1.LocustTest`) - **Migrated to v2**
- [x] v1 does not have status subresource
- [x] Two options:
  1. **Option A:** Update controller to use v2 API directly (recommended) ✅ **CHOSEN**
  2. **Option B:** Status only works for v2 CRs
- [x] Document chosen approach in NOTES.md

### If Using v2 Controller

- [x] Update import to use `locustv2`
- [x] Update reconciler struct field types
- [x] Conversion webhook handles v1→v2 for storage
- [x] Status fields ignored when converting v2→v1

---

## Task 9.6: Write Status Tests

**File:** `internal/controller/status_test.go`

### Unit Tests

- [x] Create file `internal/controller/status_test.go`
- [x] `TestSetCondition_NewCondition` - Adding new condition (TestSetCondition)
- [x] `TestSetCondition_UpdateExisting` - Updating existing condition (TestSetCondition)
- [x] `TestSetReady_True` - Setting ready to true (TestSetReady)
- [x] `TestSetReady_False` - Setting ready to false (TestSetReady)
- [x] `TestGetJobPhase_Pending` - Job in pending state (TestDerivePhaseFromJob)
- [x] `TestGetJobPhase_Running` - Job with active pods (TestDerivePhaseFromJob)
- [x] `TestGetJobPhase_Succeeded` - Job completed successfully (TestDerivePhaseFromJob)
- [x] `TestGetJobPhase_Failed` - Job failed (TestDerivePhaseFromJob)

### Integration Tests

**File:** `internal/controller/integration_test.go` (update)

- [x] `TestReconcile_InitialStatus` - Status set on creation
- [x] `TestReconcile_StatusAfterResourceCreation` - Phase becomes Running
- [x] `TestStatus_DoesNotTriggerReconcile` - Status updates don't cause loops

**Verification:**
```bash
go test ./internal/controller/... -v -run TestStatus
go test ./internal/controller/... -v -run TestSetCondition
go test ./internal/controller/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep status
```

---

## Verification

### Code Generation

```bash
make generate
make manifests
```

- [x] No errors during generation
- [x] CRD contains status subresource configuration

### Build & Test

```bash
make build
make test
```

- [x] Project builds without errors
- [x] All existing tests pass
- [x] New status tests pass
- [x] Coverage ≥80% for `internal/controller/status.go`

### Manual Verification (Optional)

```bash
# Apply a test CR
kubectl apply -f config/samples/locust_v2_locusttest.yaml

# Check status
kubectl get locusttests
kubectl get locusttest <name> -o yaml | grep -A 20 "status:"

# Verify phase column
kubectl get locusttests -o wide
```

- [x] Phase column shows in `kubectl get`
- [x] Status reflects actual test state

---

## Post-Implementation

- [x] All verification steps pass
- [x] Update `phases/README.md` with Phase 9 status
- [x] Update `phases/NOTES.md` with any deviations or decisions
- [x] Commit with message: `feat: implement status subresource for LocustTest`

---

## Files Summary

| File | Action | Est. LOC |
|------|--------|----------|
| `api/v2/conditions.go` | Create | ~50 |
| `internal/controller/status.go` | Create | ~120 |
| `internal/controller/locusttest_controller.go` | Modify | +50 |
| `internal/controller/status_test.go` | Create | ~200 |
| `internal/controller/integration_test.go` | Modify | +50 |

**Total Estimated:** ~470 LOC

---

## Quick Reference Commands

```bash
# Build and test
make generate
make manifests
make build
make test

# Run specific tests
go test ./internal/controller/... -v -run TestStatus
go test ./api/v2/... -v

# Check coverage
go test ./internal/controller/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Verify CRD status subresource
cat config/crd/bases/locust.io_locusttests.yaml | grep -A 5 "subresources:"

# Check printer columns
cat config/crd/bases/locust.io_locusttests.yaml | grep -A 10 "additionalPrinterColumns:"
```

---

## Acceptance Criteria Summary

1. **Constants Defined:** Condition types, reasons, and phases in `api/v2/conditions.go`
2. **Helpers Implemented:** Status update functions in `internal/controller/status.go`
3. **Initial Status:** Phase set to `Pending` on CR creation
4. **Running Status:** Phase updated to `Running` after resource creation
5. **Completion Status:** Phase updated based on Job completion
6. **Conditions Set:** Ready, WorkersConnected, TestCompleted conditions updated
7. **No Loops:** Status updates don't trigger unnecessary reconciles
8. **Tests Pass:** All unit and integration tests pass
9. **kubectl Shows Phase:** `kubectl get locusttests` displays Phase column

---

## Decision Log

| Decision | Options | Chosen | Rationale |
|----------|---------|--------|-----------|
| Controller API version | v1 or v2 | **v2** | v2 has status, controller migrated to use v2 directly |
| Worker count source | Spec only or Locust API | **Spec only** | Locust API requires additional complexity |
| Predicate for Job updates | GenerationChanged or custom | **Removed predicate** | Allows Job status changes to trigger reconciles |

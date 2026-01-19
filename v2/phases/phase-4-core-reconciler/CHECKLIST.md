# Phase 4: Core Reconciler - Checklist

**Estimated Effort:** 1.5 days  
**Status:** ✅ Complete

---

## Pre-Implementation

- [x] Phase 3 complete (all resource builders exist in `internal/resources/`)
- [x] Reviewed Java `LocustTestReconciler.java` (128 LOC)
- [x] Reviewed Java `ResourceCreationManager.java`
- [x] Reviewed Go patterns in `research/OPERATOR_SDK_PATTERNS.md`
- [x] Understand current scaffolded controller in `internal/controller/locusttest_controller.go`

---

## Day 1: Core Reconciler Implementation

### Task 4.1: Update Reconciler Struct

**File:** `internal/controller/locusttest_controller.go`

- [x] Add `Config *config.OperatorConfig` field to `LocustTestReconciler`
- [x] Add `Recorder record.EventRecorder` field for event recording
- [x] Import required packages:
  - [x] `"k8s.io/client-go/tools/record"`
  - [x] `"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"`
  - [x] `"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/resources"`
  - [x] `batchv1 "k8s.io/api/batch/v1"`
  - [x] `corev1 "k8s.io/api/core/v1"`
  - [x] `apierrors "k8s.io/apimachinery/pkg/api/errors"`
  - [x] `"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"`

---

### Task 4.2: Add RBAC Markers

**File:** `internal/controller/locusttest_controller.go`

- [x] Add Job RBAC marker:
  ```go
  // +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;delete
  ```
- [x] Add Service RBAC marker:
  ```go
  // +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
  ```
- [x] Add Events RBAC marker:
  ```go
  // +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
  ```
- [x] Run `make manifests` to regenerate RBAC

---

### Task 4.3: Implement Reconcile() Method

**File:** `internal/controller/locusttest_controller.go`

Main reconcile function:
- [x] Get logger from context
- [x] Fetch LocustTest CR using `r.Get()`
- [x] Handle not found error (return success - already deleted)
- [x] Check `generation > 1` for NO-OP on updates
- [x] Log update message and return early if update
- [x] Log "LocustTest created" on initial creation
- [x] Call `r.createResources()` helper

Error handling:
- [x] Use `apierrors.IsNotFound(err)` for not found
- [x] Return errors to trigger backoff requeue

---

### Task 4.4: Implement createResources() Helper

**File:** `internal/controller/locusttest_controller.go`

- [x] Build master Service using `resources.BuildMasterService()`
- [x] Build master Job using `resources.BuildMasterJob()`
- [x] Build worker Job using `resources.BuildWorkerJob()`
- [x] Create master Service with owner reference
- [x] Create master Job with owner reference
- [x] Create worker Job with owner reference
- [x] Record events for each successful creation
- [x] Handle `IsAlreadyExists` errors gracefully

---

### Task 4.5: Implement createResource() Generic Helper

**File:** `internal/controller/locusttest_controller.go`

- [x] Create helper function for creating a resource with owner reference:
  ```go
  func (r *LocustTestReconciler) createResource(ctx context.Context, lt *locustv1.LocustTest, obj client.Object, kind string) error
  ```
- [x] Set controller reference using `controllerutil.SetControllerReference()`
- [x] Create resource using `r.Create()`
- [x] Handle `IsAlreadyExists` - log and continue
- [x] Record "Created" event on success
- [x] Log error and return on failure

---

## Day 1.5: Controller Setup & Main Wiring

### Task 4.6: Update SetupWithManager()

**File:** `internal/controller/locusttest_controller.go`

- [x] Add `Owns(&batchv1.Job{})` to watch owned Jobs
- [x] Add `Owns(&corev1.Service{})` to watch owned Services
- [x] Add `WithEventFilter(predicate.GenerationChangedPredicate{})` to filter status updates
- [x] Import `"sigs.k8s.io/controller-runtime/pkg/predicate"`

---

### Task 4.7: Wire Reconciler in main.go

**File:** `cmd/main.go`

- [x] Import `"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"`
- [x] Load config: `cfg := config.LoadConfig()`
- [x] Get event recorder: `mgr.GetEventRecorderFor("locusttest-controller")`
- [x] Update reconciler initialization with Config and Recorder fields
- [x] Verify `make build` succeeds

---

## Verification

### Build Verification

- [x] `go build ./internal/controller/...` succeeds
- [x] `make build` succeeds (full project)
- [x] `make manifests` generates updated RBAC
- [x] No compiler warnings

### Manual Verification

- [ ] `make run` starts operator without errors
- [ ] Create sample LocustTest CR
- [ ] Verify Service created with `kubectl get svc`
- [ ] Verify Jobs created with `kubectl get jobs`
- [ ] Verify owner references set on resources
- [ ] Delete LocustTest CR
- [ ] Verify all resources cleaned up automatically

### Log Verification

- [ ] "LocustTest created: '{name}'" appears on create
- [ ] "Update operations..." appears on spec update
- [ ] No errors in normal operation

---

## Post-Implementation

- [x] All verification steps pass
- [x] Code follows Go conventions (gofmt, golint)
- [x] Functions are documented with godoc comments
- [x] RBAC permissions are minimal (principle of least privilege)
- [ ] Update `phases/README.md` with Phase 4 status
- [x] Update `phases/NOTES.md` with any deviations or notes

---

## Files Modified Summary

| File | Status | Changes |
|------|--------|---------|
| `internal/controller/locusttest_controller.go` | [x] | ~187 LOC (full rewrite) |
| `internal/controller/locusttest_controller_test.go` | [x] | Added Config and Recorder |
| `cmd/main.go` | [x] | ~10 lines modified |
| `config/rbac/role.yaml` | [x] | Auto-generated |

---

## Test Commands

```bash
# Build and verify
make build
make manifests

# Run operator locally
make run

# In another terminal, create test CR
kubectl apply -f config/samples/locust_v1_locusttest.yaml

# Verify resources
kubectl get locusttests
kubectl get jobs
kubectl get svc

# Check owner references
kubectl get job <job-name> -o jsonpath='{.metadata.ownerReferences}'

# Delete and verify cleanup
kubectl delete locusttest <name>
kubectl get jobs  # Should be empty
kubectl get svc   # Should be empty
```

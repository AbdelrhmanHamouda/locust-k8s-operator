# Phase 9: Status Subresource - Technical Design

**Version:** 1.0  
**Status:** Draft

---

## Overview

This document details the technical design for implementing the status subresource for LocustTest resources. The status subresource enables tracking test lifecycle phases, worker connections, and standard Kubernetes conditions.

---

## 1. Architecture

### 1.1 Status Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        LocustTest CR                             │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ spec:                                                     │   │
│  │   worker.replicas: 10                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ status:                                                   │   │
│  │   phase: Running                                          │   │
│  │   expectedWorkers: 10                                     │   │
│  │   connectedWorkers: 10                                    │   │
│  │   startTime: "2026-01-19T22:00:00Z"                       │   │
│  │   conditions: [Ready, WorkersConnected, TestCompleted]    │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Reconciler                                  │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │ Fetch CR    │───▶│ Create/Get  │───▶│ Update      │         │
│  │             │    │ Jobs        │    │ Status      │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Jobs                                     │
│  ┌─────────────────────┐    ┌─────────────────────┐            │
│  │ Master Job          │    │ Worker Job          │            │
│  │ status:             │    │ status:             │            │
│  │   active: 1         │    │   active: 10        │            │
│  │   succeeded: 0      │    │   succeeded: 0      │            │
│  └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Phase State Machine

```
                    ┌───────────┐
                    │  (empty)  │
                    └─────┬─────┘
                          │ CR Created
                          ▼
                    ┌───────────┐
                    │  Pending  │
                    └─────┬─────┘
                          │ Jobs Created & Running
                          ▼
                    ┌───────────┐
              ┌─────│  Running  │─────┐
              │     └───────────┘     │
              │                       │
    Master Job│                       │Master Job
    Succeeded ▼                       ▼ Failed
        ┌───────────┐           ┌───────────┐
        │ Succeeded │           │  Failed   │
        └───────────┘           └───────────┘
```

---

## 2. API Design

### 2.1 Condition Constants

**File:** `api/v2/conditions.go`

```go
package v2

// Condition Types
const (
    // ConditionTypeReady indicates whether all resources are created and healthy.
    ConditionTypeReady = "Ready"

    // ConditionTypeWorkersConnected indicates whether workers have connected to master.
    // Note: Initial implementation uses spec.worker.replicas as expected count.
    // Future: Could integrate with Locust API for real-time count.
    ConditionTypeWorkersConnected = "WorkersConnected"

    // ConditionTypeTestCompleted indicates whether the test has finished execution.
    ConditionTypeTestCompleted = "TestCompleted"
)

// Reasons for ConditionTypeReady
const (
    ReasonResourcesCreating = "ResourcesCreating"
    ReasonResourcesCreated  = "ResourcesCreated"
    ReasonResourcesFailed   = "ResourcesFailed"
)

// Reasons for ConditionTypeWorkersConnected
const (
    ReasonWaitingForWorkers   = "WaitingForWorkers"
    ReasonAllWorkersConnected = "AllWorkersConnected"
    ReasonWorkersMissing      = "WorkersMissing"
)

// Reasons for ConditionTypeTestCompleted
const (
    ReasonTestInProgress = "TestInProgress"
    ReasonTestSucceeded  = "TestSucceeded"
    ReasonTestFailed     = "TestFailed"
)

// Phases represent the lifecycle state of a LocustTest
const (
    PhasePending   = "Pending"
    PhaseRunning   = "Running"
    PhaseSucceeded = "Succeeded"
    PhaseFailed    = "Failed"
)
```

### 2.2 Status Structure (Already Defined in v2)

The `LocustTestStatus` is already defined in `api/v2/locusttest_types.go`:

```go
type LocustTestStatus struct {
    Phase            string             `json:"phase,omitempty"`
    ExpectedWorkers  int32              `json:"expectedWorkers,omitempty"`
    ConnectedWorkers int32              `json:"connectedWorkers,omitempty"`
    StartTime        *metav1.Time       `json:"startTime,omitempty"`
    CompletionTime   *metav1.Time       `json:"completionTime,omitempty"`
    Conditions       []metav1.Condition `json:"conditions,omitempty"`
}
```

---

## 3. Implementation Details

### 3.1 Status Helper Functions

**File:** `internal/controller/status.go`

```go
package controller

import (
    "context"

    batchv1 "k8s.io/api/batch/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/meta"
    
    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// updatePhase updates the phase and persists to API server.
func (r *LocustTestReconciler) updatePhase(ctx context.Context, lt *locustv2.LocustTest, phase string) error {
    if lt.Status.Phase == phase {
        return nil // No change needed
    }
    lt.Status.Phase = phase
    return r.Status().Update(ctx, lt)
}

// setCondition sets a condition on the LocustTest status.
// Uses the standard meta.SetStatusCondition helper for proper handling.
func (r *LocustTestReconciler) setCondition(
    lt *locustv2.LocustTest,
    condType string,
    status metav1.ConditionStatus,
    reason string,
    message string,
) {
    meta.SetStatusCondition(&lt.Status.Conditions, metav1.Condition{
        Type:               condType,
        Status:             status,
        Reason:             reason,
        Message:            message,
        LastTransitionTime: metav1.Now(),
        ObservedGeneration: lt.Generation,
    })
}

// setReady is a convenience wrapper for setting the Ready condition.
func (r *LocustTestReconciler) setReady(lt *locustv2.LocustTest, ready bool, reason, message string) {
    status := metav1.ConditionFalse
    if ready {
        status = metav1.ConditionTrue
    }
    r.setCondition(lt, locustv2.ConditionTypeReady, status, reason, message)
}

// initializeStatus sets initial status values for a new LocustTest.
func (r *LocustTestReconciler) initializeStatus(lt *locustv2.LocustTest) {
    lt.Status.Phase = locustv2.PhasePending
    lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas
    lt.Status.ConnectedWorkers = 0
    
    r.setReady(lt, false, locustv2.ReasonResourcesCreating, "Creating resources")
    r.setCondition(lt, locustv2.ConditionTypeWorkersConnected, 
        metav1.ConditionFalse, locustv2.ReasonWaitingForWorkers, 
        "Waiting for workers to connect")
    r.setCondition(lt, locustv2.ConditionTypeTestCompleted,
        metav1.ConditionFalse, locustv2.ReasonTestInProgress,
        "Test has not started")
}

// updateStatusFromJobs derives status from the current state of owned Jobs.
func (r *LocustTestReconciler) updateStatusFromJobs(
    ctx context.Context, 
    lt *locustv2.LocustTest,
    masterJob *batchv1.Job,
    workerJob *batchv1.Job,
) error {
    // Determine phase from master Job status
    newPhase := derivePhaseFromJob(masterJob)
    
    // Update phase if changed
    if lt.Status.Phase != newPhase {
        lt.Status.Phase = newPhase
        
        // Set timestamps
        if newPhase == locustv2.PhaseRunning && lt.Status.StartTime == nil {
            now := metav1.Now()
            lt.Status.StartTime = &now
        }
        
        if newPhase == locustv2.PhaseSucceeded || newPhase == locustv2.PhaseFailed {
            now := metav1.Now()
            lt.Status.CompletionTime = &now
            
            // Update TestCompleted condition
            if newPhase == locustv2.PhaseSucceeded {
                r.setCondition(lt, locustv2.ConditionTypeTestCompleted,
                    metav1.ConditionTrue, locustv2.ReasonTestSucceeded,
                    "Test completed successfully")
            } else {
                r.setCondition(lt, locustv2.ConditionTypeTestCompleted,
                    metav1.ConditionTrue, locustv2.ReasonTestFailed,
                    "Test failed")
            }
        }
    }
    
    // Update worker connection status (approximation from worker Job)
    if workerJob != nil {
        lt.Status.ConnectedWorkers = workerJob.Status.Active
        
        if lt.Status.ConnectedWorkers >= lt.Status.ExpectedWorkers {
            r.setCondition(lt, locustv2.ConditionTypeWorkersConnected,
                metav1.ConditionTrue, locustv2.ReasonAllWorkersConnected,
                fmt.Sprintf("%d/%d workers connected", 
                    lt.Status.ConnectedWorkers, lt.Status.ExpectedWorkers))
        }
    }
    
    return r.Status().Update(ctx, lt)
}

// derivePhaseFromJob determines the LocustTest phase from Job status.
func derivePhaseFromJob(job *batchv1.Job) string {
    if job == nil {
        return locustv2.PhasePending
    }
    
    // Check for completion conditions
    for _, condition := range job.Status.Conditions {
        if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
            return locustv2.PhaseSucceeded
        }
        if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
            return locustv2.PhaseFailed
        }
    }
    
    // Check if running
    if job.Status.Active > 0 {
        return locustv2.PhaseRunning
    }
    
    return locustv2.PhasePending
}
```

### 3.2 Reconciler Updates

**File:** `internal/controller/locusttest_controller.go`

Key changes to the Reconcile function:

```go
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := logf.FromContext(ctx)

    // Fetch the LocustTest CR
    locustTest := &locustv2.LocustTest{}
    if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
        if apierrors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // Initialize status on first reconcile
    if locustTest.Status.Phase == "" {
        r.initializeStatus(locustTest)
        if err := r.Status().Update(ctx, locustTest); err != nil {
            log.Error(err, "Failed to initialize status")
            return ctrl.Result{}, err
        }
        // Re-fetch after status update
        if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
            return ctrl.Result{}, err
        }
    }

    // NO-OP on updates - matching Java behavior
    if locustTest.Generation > 1 {
        log.Info("LocustTest updated - NO-OP by design")
        // Still update status from Jobs if needed
        return r.reconcileStatus(ctx, locustTest)
    }

    // Create resources
    result, err := r.createResources(ctx, locustTest)
    if err != nil {
        // Update status to reflect failure
        r.setReady(locustTest, false, locustv2.ReasonResourcesFailed, err.Error())
        if statusErr := r.Status().Update(ctx, locustTest); statusErr != nil {
            log.Error(statusErr, "Failed to update status after error")
        }
        return result, err
    }

    // Update status after successful resource creation
    locustTest.Status.Phase = locustv2.PhaseRunning
    now := metav1.Now()
    locustTest.Status.StartTime = &now
    r.setReady(locustTest, true, locustv2.ReasonResourcesCreated, "All resources created")
    
    if err := r.Status().Update(ctx, locustTest); err != nil {
        log.Error(err, "Failed to update status after resource creation")
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}

// reconcileStatus updates status based on current Job states.
func (r *LocustTestReconciler) reconcileStatus(ctx context.Context, lt *locustv2.LocustTest) (ctrl.Result, error) {
    log := logf.FromContext(ctx)
    
    // Get master Job
    masterJob := &batchv1.Job{}
    masterJobName := resources.NodeName(lt.Name, resources.Master)
    if err := r.Get(ctx, client.ObjectKey{
        Namespace: lt.Namespace,
        Name:      masterJobName,
    }, masterJob); err != nil {
        if !apierrors.IsNotFound(err) {
            return ctrl.Result{}, err
        }
        masterJob = nil
    }
    
    // Get worker Job
    workerJob := &batchv1.Job{}
    workerJobName := resources.NodeName(lt.Name, resources.Worker)
    if err := r.Get(ctx, client.ObjectKey{
        Namespace: lt.Namespace,
        Name:      workerJobName,
    }, workerJob); err != nil {
        if !apierrors.IsNotFound(err) {
            return ctrl.Result{}, err
        }
        workerJob = nil
    }
    
    // Update status from Jobs
    if err := r.updateStatusFromJobs(ctx, lt, masterJob, workerJob); err != nil {
        log.Error(err, "Failed to update status from Jobs")
        return ctrl.Result{}, err
    }
    
    return ctrl.Result{}, nil
}
```

### 3.3 Controller Setup Updates

The `SetupWithManager` function needs adjustment to handle Job status updates:

```go
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&locustv2.LocustTest{}).
        Owns(&batchv1.Job{}).
        Owns(&corev1.Service{}).
        // Remove GenerationChangedPredicate to allow Job status triggers
        // Or use a custom predicate that allows both:
        WithEventFilter(predicate.Or(
            predicate.GenerationChangedPredicate{},
            predicate.ResourceVersionChangedPredicate{},
        )).
        Named("locusttest").
        Complete(r)
}
```

**Alternative:** Use `Watches` with custom handler for Jobs to trigger status updates without full reconcile.

---

## 4. API Version Considerations

### 4.1 Current State

- Controller currently uses v1 API (`locustv1.LocustTest`)
- v1 API does not have status fields
- v2 API has full status support

### 4.2 Recommended Approach

**Migrate controller to v2 API:**

1. Update imports from `locustv1` to `locustv2`
2. Update field accesses (e.g., `spec.workerReplicas` → `spec.worker.replicas`)
3. Conversion webhook handles v1 CRs transparently
4. Status updates only apply to v2 storage

**Benefits:**
- Clean implementation
- Full status support
- v1 users automatically get v2 features via conversion

**Migration Steps:**
1. Update controller imports
2. Update field references in reconciler
3. Update field references in resource builders
4. Run tests to verify conversion works

---

## 5. Testing Strategy

### 5.1 Unit Tests

| Test | Description |
|------|-------------|
| `TestSetCondition_New` | Adding a new condition |
| `TestSetCondition_Update` | Updating existing condition |
| `TestSetCondition_NoChange` | No update when condition unchanged |
| `TestDerivePhaseFromJob_Nil` | Returns Pending for nil Job |
| `TestDerivePhaseFromJob_Active` | Returns Running for active Job |
| `TestDerivePhaseFromJob_Complete` | Returns Succeeded for complete Job |
| `TestDerivePhaseFromJob_Failed` | Returns Failed for failed Job |
| `TestInitializeStatus` | Correct initial values set |

### 5.2 Integration Tests

| Test | Description |
|------|-------------|
| `TestReconcile_InitialStatus` | Status initialized on creation |
| `TestReconcile_StatusAfterCreate` | Phase becomes Running after resources |
| `TestReconcile_JobComplete` | Phase becomes Succeeded when Job completes |
| `TestReconcile_JobFailed` | Phase becomes Failed when Job fails |
| `TestStatus_NoReconcileLoop` | Status updates don't trigger loops |

---

## 6. Future Enhancements

### 6.1 Real-time Worker Count

Current implementation uses Job's `status.active` as approximation. Future enhancement could:

1. Query Locust master API for actual connected workers
2. Implement sidecar or exec probe for worker count
3. Use Locust's `/stats/workers` endpoint

### 6.2 Test Metrics in Status

Could add test result summary:

```go
type TestResult struct {
    TotalRequests       int64   `json:"totalRequests,omitempty"`
    FailedRequests      int64   `json:"failedRequests,omitempty"`
    AverageResponseTime float64 `json:"averageResponseTime,omitempty"`
}
```

### 6.3 Progress Tracking

For time-based tests, could track progress:

```go
type Progress struct {
    CurrentUsers int32  `json:"currentUsers,omitempty"`
    TargetUsers  int32  `json:"targetUsers,omitempty"`
    Duration     string `json:"duration,omitempty"`
}
```

---

## 7. References

- [Kubernetes API Conventions - Status](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)
- [controller-runtime Status Client](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#StatusWriter)
- [meta.SetStatusCondition](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/meta#SetStatusCondition)
- [Kubebuilder Status Subresource](https://book.kubebuilder.io/reference/generating-crd.html#status)

# Phase 4: Core Reconciler - Implementation Plan

**Effort:** 1.5 days  
**Priority:** P0 - Critical Path  
**Prerequisites:** Phase 3 (Resource Builders)  
**Requirements:** §3.3 Resilience, §8.3 Design Principles (Immutable Tests)

---

## Objective

Implement the reconciliation loop matching Java `LocustTestReconciler.java` behavior. The reconciler watches LocustTest CRs, creates Kubernetes resources (Jobs, Services) on CR creation, and relies on owner references for automatic cleanup on CR deletion.

---

## Day 1: Core Reconciler Implementation

### Task 4.1: Update Reconciler Struct

**File:** `internal/controller/locusttest_controller.go`

Update the reconciler struct to include configuration and event recorder:

```go
package controller

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/resources"
)

// LocustTestReconciler reconciles a LocustTest object
type LocustTestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Config   *config.OperatorConfig
	Recorder record.EventRecorder
}
```

---

### Task 4.2: Add RBAC Markers

Add RBAC markers before the Reconcile function to grant necessary permissions:

```go
// +kubebuilder:rbac:groups=locust.io,resources=locusttests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=locust.io,resources=locusttests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=locust.io,resources=locusttests/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
```

After adding markers, regenerate RBAC with:
```bash
make manifests
```

---

### Task 4.3: Implement Reconcile() Method

The main reconciliation function matching Java behavior:

```go
// Reconcile handles LocustTest CR events.
// On creation: Creates master Service, master Job, and worker Job.
// On update: NO-OP by design (tests are immutable).
// On deletion: Automatic cleanup via owner references.
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the LocustTest CR
	locustTest := &locustv1.LocustTest{}
	if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted - nothing to do (cleanup via owner references)
			log.V(1).Info("LocustTest not found, likely deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch LocustTest")
		return ctrl.Result{}, err
	}

	// NO-OP on updates - matching Java behavior
	// Generation > 1 means the spec has been modified after creation
	if locustTest.Generation > 1 {
		log.Info("LocustTest updated - NO-OP by design",
			"name", locustTest.Name,
			"namespace", locustTest.Namespace)
		log.Info("Update operations on LocustTest are not currently supported!",
			"name", locustTest.Name)
		return ctrl.Result{}, nil
	}

	// On initial creation
	log.Info("LocustTest created",
		"name", locustTest.Name,
		"namespace", locustTest.Namespace)

	// Log detailed CR information (debug level)
	log.V(1).Info("Custom resource information",
		"image", locustTest.Spec.Image,
		"masterCommand", locustTest.Spec.MasterCommandSeed,
		"workerCommand", locustTest.Spec.WorkerCommandSeed,
		"workerReplicas", locustTest.Spec.WorkerReplicas,
		"configMap", locustTest.Spec.ConfigMap)

	// Create resources
	return r.createResources(ctx, locustTest)
}
```

---

### Task 4.4: Implement createResources() Helper

Create all required Kubernetes resources with owner references:

```go
// createResources creates the master Service, master Job, and worker Job.
// Resources are created with owner references for automatic garbage collection.
func (r *LocustTestReconciler) createResources(ctx context.Context, lt *locustv1.LocustTest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Build resources using resource builders from Phase 3
	masterService := resources.BuildMasterService(lt, r.Config)
	masterJob := resources.BuildMasterJob(lt, r.Config)
	workerJob := resources.BuildWorkerJob(lt, r.Config)

	// Create master Service
	if err := r.createResource(ctx, lt, masterService, "Service"); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Master Service reconciled", "name", masterService.Name)

	// Create master Job
	if err := r.createResource(ctx, lt, masterJob, "Job"); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Master Job reconciled", "name", masterJob.Name)

	// Create worker Job
	if err := r.createResource(ctx, lt, workerJob, "Job"); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Worker Job reconciled", "name", workerJob.Name)

	log.Info("All resources created successfully",
		"locustTest", lt.Name,
		"masterService", masterService.Name,
		"masterJob", masterJob.Name,
		"workerJob", workerJob.Name)

	return ctrl.Result{}, nil
}
```

---

### Task 4.5: Implement createResource() Generic Helper

A helper function to create any resource with owner reference and proper error handling:

```go
// createResource creates a Kubernetes resource with owner reference set.
// If the resource already exists, it logs and returns success (idempotent).
func (r *LocustTestReconciler) createResource(ctx context.Context, lt *locustv1.LocustTest, obj client.Object, kind string) error {
	log := logf.FromContext(ctx)

	// Set owner reference for automatic garbage collection
	if err := controllerutil.SetControllerReference(lt, obj, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference",
			"kind", kind,
			"name", obj.GetName())
		return err
	}

	// Create the resource
	if err := r.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Resource already exists - this is fine (idempotent)
			log.V(1).Info("Resource already exists",
				"kind", kind,
				"name", obj.GetName())
			return nil
		}
		log.Error(err, "Failed to create resource",
			"kind", kind,
			"name", obj.GetName())
		return err
	}

	// Record event for successful creation
	r.Recorder.Event(lt, corev1.EventTypeNormal, "Created",
		fmt.Sprintf("Created %s %s", kind, obj.GetName()))

	log.Info("Created resource",
		"kind", kind,
		"name", obj.GetName(),
		"namespace", obj.GetNamespace())

	return nil
}
```

---

## Day 1.5: Controller Setup & Main Wiring

### Task 4.6: Update SetupWithManager()

Configure the controller to watch owned resources and filter events:

```go
// SetupWithManager sets up the controller with the Manager.
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&locustv1.LocustTest{}).
		Owns(&batchv1.Job{}).     // Watch owned Jobs for status updates
		Owns(&corev1.Service{}).  // Watch owned Services
		WithEventFilter(predicate.GenerationChangedPredicate{}). // Filter status-only updates
		Named("locusttest").
		Complete(r)
}
```

**Note:** The `GenerationChangedPredicate` ensures we only reconcile when the spec changes (generation incremented), not on status-only updates.

---

### Task 4.7: Wire Reconciler in main.go

**File:** `cmd/main.go`

Update the main function to load config and create the reconciler with all dependencies:

```go
// Add import
import (
	// ... existing imports ...
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

func main() {
	// ... existing setup code ...

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		// ... existing options ...
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Load operator configuration
	cfg := config.LoadConfig()
	setupLog.Info("Operator configuration loaded",
		"ttlSecondsAfterFinished", cfg.TTLSecondsAfterFinished,
		"metricsExporterImage", cfg.MetricsExporterImage,
		"affinityInjection", cfg.EnableAffinityCRInjection,
		"tolerationsInjection", cfg.EnableTolerationsCRInjection)

	// Setup LocustTest reconciler
	if err = (&controller.LocustTestReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Config:   cfg,
		Recorder: mgr.GetEventRecorderFor("locusttest-controller"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LocustTest")
		os.Exit(1)
	}

	// ... rest of main ...
}
```

---

## Complete Controller File

Here's the complete `internal/controller/locusttest_controller.go`:

```go
/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/resources"
)

// LocustTestReconciler reconciles a LocustTest object
type LocustTestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Config   *config.OperatorConfig
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=locust.io,resources=locusttests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=locust.io,resources=locusttests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=locust.io,resources=locusttests/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile handles LocustTest CR events.
// On creation: Creates master Service, master Job, and worker Job.
// On update: NO-OP by design (tests are immutable).
// On deletion: Automatic cleanup via owner references.
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the LocustTest CR
	locustTest := &locustv1.LocustTest{}
	if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted - nothing to do (cleanup via owner references)
			log.V(1).Info("LocustTest not found, likely deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch LocustTest")
		return ctrl.Result{}, err
	}

	// NO-OP on updates - matching Java behavior
	// Generation > 1 means the spec has been modified after creation
	if locustTest.Generation > 1 {
		log.Info("LocustTest updated - NO-OP by design",
			"name", locustTest.Name,
			"namespace", locustTest.Namespace)
		log.Info("Update operations on LocustTest are not currently supported!",
			"name", locustTest.Name)
		return ctrl.Result{}, nil
	}

	// On initial creation
	log.Info("LocustTest created",
		"name", locustTest.Name,
		"namespace", locustTest.Namespace)

	// Log detailed CR information (debug level)
	log.V(1).Info("Custom resource information",
		"image", locustTest.Spec.Image,
		"masterCommand", locustTest.Spec.MasterCommandSeed,
		"workerCommand", locustTest.Spec.WorkerCommandSeed,
		"workerReplicas", locustTest.Spec.WorkerReplicas,
		"configMap", locustTest.Spec.ConfigMap)

	// Create resources
	return r.createResources(ctx, locustTest)
}

// createResources creates the master Service, master Job, and worker Job.
// Resources are created with owner references for automatic garbage collection.
func (r *LocustTestReconciler) createResources(ctx context.Context, lt *locustv1.LocustTest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Build resources using resource builders from Phase 3
	masterService := resources.BuildMasterService(lt, r.Config)
	masterJob := resources.BuildMasterJob(lt, r.Config)
	workerJob := resources.BuildWorkerJob(lt, r.Config)

	// Create master Service
	if err := r.createResource(ctx, lt, masterService, "Service"); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Master Service reconciled", "name", masterService.Name)

	// Create master Job
	if err := r.createResource(ctx, lt, masterJob, "Job"); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Master Job reconciled", "name", masterJob.Name)

	// Create worker Job
	if err := r.createResource(ctx, lt, workerJob, "Job"); err != nil {
		return ctrl.Result{}, err
	}
	log.V(1).Info("Worker Job reconciled", "name", workerJob.Name)

	log.Info("All resources created successfully",
		"locustTest", lt.Name,
		"masterService", masterService.Name,
		"masterJob", masterJob.Name,
		"workerJob", workerJob.Name)

	return ctrl.Result{}, nil
}

// createResource creates a Kubernetes resource with owner reference set.
// If the resource already exists, it logs and returns success (idempotent).
func (r *LocustTestReconciler) createResource(ctx context.Context, lt *locustv1.LocustTest, obj client.Object, kind string) error {
	log := logf.FromContext(ctx)

	// Set owner reference for automatic garbage collection
	if err := controllerutil.SetControllerReference(lt, obj, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference",
			"kind", kind,
			"name", obj.GetName())
		return err
	}

	// Create the resource
	if err := r.Create(ctx, obj); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Resource already exists - this is fine (idempotent)
			log.V(1).Info("Resource already exists",
				"kind", kind,
				"name", obj.GetName())
			return nil
		}
		log.Error(err, "Failed to create resource",
			"kind", kind,
			"name", obj.GetName())
		return err
	}

	// Record event for successful creation
	r.Recorder.Event(lt, corev1.EventTypeNormal, "Created",
		fmt.Sprintf("Created %s %s", kind, obj.GetName()))

	log.Info("Created resource",
		"kind", kind,
		"name", obj.GetName(),
		"namespace", obj.GetNamespace())

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&locustv1.LocustTest{}).
		Owns(&batchv1.Job{}).     // Watch owned Jobs for status updates
		Owns(&corev1.Service{}).  // Watch owned Services
		WithEventFilter(predicate.GenerationChangedPredicate{}). // Filter status-only updates
		Named("locusttest").
		Complete(r)
}
```

---

## Verification Steps

### 1. Build Verification

```bash
# Compile the controller
go build ./internal/controller/...

# Full project build
make build

# Regenerate manifests (RBAC)
make manifests

# Verify RBAC was updated
cat config/rbac/role.yaml | grep -A5 "jobs"
cat config/rbac/role.yaml | grep -A5 "services"
```

### 2. Local Testing

```bash
# Terminal 1: Run operator
make run

# Terminal 2: Create sample CR
cat <<EOF | kubectl apply -f -
apiVersion: locust.io/v1
kind: LocustTest
metadata:
  name: sample-test
  namespace: default
spec:
  masterCommandSeed: "locust -f /lotest/src/test.py"
  workerCommandSeed: "locust -f /lotest/src/test.py"
  workerReplicas: 3
  image: locustio/locust:latest
  imagePullPolicy: Always
  configMap: locust-test-files
EOF

# Verify resources created
kubectl get locusttests
kubectl get jobs
kubectl get svc

# Check owner references
kubectl get job sample-test-master -o jsonpath='{.metadata.ownerReferences}' | jq

# Check events
kubectl describe locusttest sample-test | grep Events -A10

# Test update NO-OP
kubectl patch locusttest sample-test --type='json' -p='[{"op": "replace", "path": "/spec/workerReplicas", "value": 5}]'
# Operator should log "Update operations on LocustTest are not currently supported!"

# Delete and verify cleanup
kubectl delete locusttest sample-test
kubectl get jobs    # Should be empty
kubectl get svc     # Should be empty
```

### 3. Log Verification

Expected log output on create:
```
INFO    LocustTest created    {"name": "sample-test", "namespace": "default"}
INFO    Created resource      {"kind": "Service", "name": "sample-test-master", "namespace": "default"}
INFO    Created resource      {"kind": "Job", "name": "sample-test-master", "namespace": "default"}
INFO    Created resource      {"kind": "Job", "name": "sample-test-worker", "namespace": "default"}
INFO    All resources created successfully    {"locustTest": "sample-test", ...}
```

Expected log output on update:
```
INFO    LocustTest updated - NO-OP by design    {"name": "sample-test", "namespace": "default"}
INFO    Update operations on LocustTest are not currently supported!    {"name": "sample-test"}
```

---

## Edge Cases to Handle

| Scenario | Behavior |
|----------|----------|
| CR deleted before reconcile | `IsNotFound` - return success |
| Resource already exists | Log at debug level, continue |
| API server error | Return error, trigger backoff |
| Invalid CR spec | Validation should reject at admission (Phase 1 markers) |
| Namespace deletion | Owner references trigger cascade delete |

---

## Java Behavior Comparison

| Aspect | Java | Go |
|--------|------|-----|
| Update handling | `Generation > 1 → noUpdate()` | `Generation > 1 → return {}` |
| Resource creation order | Service → Master Job → Worker Job | Same |
| Cleanup | Explicit delete in `cleanup()` | Automatic via owner references |
| Logging | SLF4J with `{}` placeholders | Structured logging with key-value pairs |
| Event recording | Not implemented | Events recorded on creation |

---

## References

- [Java LocustTestReconciler](../../../src/main/java/com/locust/operator/controller/LocustTestReconciler.java)
- [OPERATOR_SDK_PATTERNS.md](../../research/OPERATOR_SDK_PATTERNS.md) - §3 Reconciliation Patterns
- [CONTROLLER_RUNTIME_DEEP_DIVE.md](../../research/CONTROLLER_RUNTIME_DEEP_DIVE.md)
- [controller-runtime pkg docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime)

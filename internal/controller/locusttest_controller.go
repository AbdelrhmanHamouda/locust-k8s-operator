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
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/resources"
)

const finalizerName = "locust.io/cleanup"

// LocustTestReconciler reconciles a LocustTest object
type LocustTestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Config   *config.OperatorConfig
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=locust.io,resources=locusttests,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=locust.io,resources=locusttests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=locust.io,resources=locusttests/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile handles LocustTest CR events.
// On creation: Creates master Service, master Job, and worker Job.
// On update: NO-OP by design (tests are immutable).
// On deletion: Finalizer emits log + Event, then cleanup via owner references.
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the LocustTest CR
	locustTest := &locustv2.LocustTest{}
	if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
		if apierrors.IsNotFound(err) {
			// CR deleted - nothing to do (cleanup via owner references)
			log.V(1).Info("LocustTest not found, likely deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to fetch LocustTest")
		return ctrl.Result{}, err
	}

	// Handle deletion: finalizer ensures visible logs and events
	if !locustTest.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(locustTest, finalizerName) {
			log.Info("LocustTest deleted, cleaning up resources via owner references",
				"name", locustTest.Name,
				"namespace", locustTest.Namespace)
			r.Recorder.Event(locustTest, corev1.EventTypeNormal, "Deleting",
				"LocustTest and owned resources being cleaned up")
			controllerutil.RemoveFinalizer(locustTest, finalizerName)
			if err := r.Update(ctx, locustTest); err != nil {
				log.Error(err, "Failed to remove finalizer")
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer on first reconcile if not present
	if !controllerutil.ContainsFinalizer(locustTest, finalizerName) {
		controllerutil.AddFinalizer(locustTest, finalizerName)
		if err := r.Update(ctx, locustTest); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, fmt.Errorf("failed to add finalizer: %w", err)
		}
		if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to re-fetch after adding finalizer: %w", err)
		}
	}

	// Initialize status on first reconcile
	if locustTest.Status.Phase == "" {
		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
				return err
			}
			r.initializeStatus(locustTest)
			return r.Status().Update(ctx, locustTest)
		}); err != nil {
			log.Error(err, "Failed to initialize status")
			return ctrl.Result{}, fmt.Errorf("failed to initialize status: %w", err)
		}
		// Re-fetch after status update to get the latest resource version
		if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to re-fetch LocustTest after status update: %w", err)
		}
	}

	// If resources already exist (Phase is Running or terminal), check Job status
	// This handles reconciles triggered by Job status changes
	if locustTest.Status.Phase == locustv2.PhaseRunning ||
		locustTest.Status.Phase == locustv2.PhaseSucceeded ||
		locustTest.Status.Phase == locustv2.PhaseFailed {
		return r.reconcileStatus(ctx, locustTest)
	}

	// Phase == Pending: create resources
	// Log informational message if this is a spec update (generation > 1)
	// The phase-based state machine handles this correctly — Pending always creates
	if locustTest.Generation > 1 {
		log.V(1).Info("Update operations on LocustTest are not supported by design",
			"name", locustTest.Name,
			"namespace", locustTest.Namespace)
	}

	// On initial creation or pending (Phase is Pending)
	log.Info("LocustTest created",
		"name", locustTest.Name,
		"namespace", locustTest.Namespace)

	// Log detailed CR information (debug level)
	var configMapRef string
	if locustTest.Spec.TestFiles != nil {
		configMapRef = locustTest.Spec.TestFiles.ConfigMapRef
	}
	log.V(1).Info("Custom resource information",
		"image", locustTest.Spec.Image,
		"masterCommand", locustTest.Spec.Master.Command,
		"workerCommand", locustTest.Spec.Worker.Command,
		"workerReplicas", locustTest.Spec.Worker.Replicas,
		"configMap", configMapRef)

	// Create resources
	return r.createResources(ctx, locustTest)
}

// createResources creates the master Service, master Job, and worker Job.
// Resources are created with owner references for automatic garbage collection.
func (r *LocustTestReconciler) createResources(ctx context.Context, lt *locustv2.LocustTest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Build resources using resource builders from Phase 3
	masterService := resources.BuildMasterService(lt, r.Config)
	masterJob := resources.BuildMasterJob(lt, r.Config, log)
	workerJob := resources.BuildWorkerJob(lt, r.Config, log)

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

	// Update status after successful resource creation (with conflict retry)
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		if err := r.Get(ctx, client.ObjectKeyFromObject(lt), lt); err != nil {
			return err
		}
		lt.Status.Phase = locustv2.PhaseRunning
		lt.Status.ObservedGeneration = lt.Generation
		if lt.Status.StartTime == nil {
			now := metav1.Now()
			lt.Status.StartTime = &now
		}
		r.setReady(lt, true, locustv2.ReasonResourcesCreated, "All resources created")
		return r.Status().Update(ctx, lt)
	}); err != nil {
		log.Error(err, "Failed to update status after resource creation")
		return ctrl.Result{}, fmt.Errorf("failed to update status after resource creation: %w", err)
	}

	return ctrl.Result{}, nil
}

// createResource creates a Kubernetes resource with owner reference set.
// If the resource already exists, it logs and returns success (idempotent).
func (r *LocustTestReconciler) createResource(ctx context.Context, lt *locustv2.LocustTest, obj client.Object, kind string) error {
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

// handleExternalResourceDeletion handles the case where a resource was externally deleted.
// It transitions the LocustTest to Pending phase to trigger recreation on the next reconcile.
// Returns (shouldRequeue, requeueAfter, error).
func (r *LocustTestReconciler) handleExternalResourceDeletion(
	ctx context.Context,
	lt *locustv2.LocustTest,
	resourceName, resourceKind string,
	obj client.Object,
) (bool, time.Duration, error) {
	log := logf.FromContext(ctx)

	// Try to fetch the resource
	if err := r.Get(ctx, client.ObjectKey{Name: resourceName, Namespace: lt.Namespace}, obj); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource was externally deleted — transition to Pending for recovery
			log.Info(fmt.Sprintf("%s externally deleted, transitioning to Pending for recovery", resourceKind),
				resourceKind, resourceName)
			r.Recorder.Event(lt, corev1.EventTypeWarning, "ResourceDeleted",
				fmt.Sprintf("%s %s was deleted externally, will attempt recreation", resourceKind, resourceName))

			// Reset to Pending to trigger resource recreation on next reconcile
			log.Info("Attempting to update status to Pending after external deletion",
				"currentPhase", lt.Status.Phase,
				"generation", lt.Generation,
				"observedGeneration", lt.Status.ObservedGeneration)
			if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				if err := r.Get(ctx, client.ObjectKeyFromObject(lt), lt); err != nil {
					log.Error(err, "Failed to re-fetch LocustTest during status update retry")
					return err
				}
				log.V(1).Info("Re-fetched LocustTest for status update",
					"resourceVersion", lt.ResourceVersion,
					"phase", lt.Status.Phase)
				lt.Status.Phase = locustv2.PhasePending
				lt.Status.ObservedGeneration = lt.Generation
				r.setReady(lt, false, locustv2.ReasonResourcesCreating, "Recreating externally deleted resources")
				return r.Status().Update(ctx, lt)
			}); err != nil {
				log.Error(err, fmt.Sprintf("Failed to update status after detecting %s deletion", resourceKind),
					resourceKind, resourceName,
					"retryAttempts", "exhausted")
				// Still requeue to retry the entire reconciliation
				return true, 5 * time.Second, nil
			}
			log.Info(fmt.Sprintf("Successfully updated status to Pending, will recreate %s", resourceKind),
				resourceKind, resourceName)
			return true, time.Second, nil
		}
		return false, 0, fmt.Errorf("failed to get %s: %w", resourceKind, err)
	}

	// Resource exists, no action needed
	return false, 0, nil
}

// checkResourcesExist verifies that all required resources (Service, Master Job, Worker Job) exist.
// If any resource is missing, it handles external deletion recovery.
// Returns (masterJob, workerJob, shouldRequeue, requeueAfter, error).
func (r *LocustTestReconciler) checkResourcesExist(
	ctx context.Context,
	lt *locustv2.LocustTest,
) (*batchv1.Job, *batchv1.Job, bool, time.Duration, error) {
	// Check for externally deleted Service
	masterServiceName := lt.Name + "-master"
	masterService := &corev1.Service{}
	if shouldRequeue, requeueAfter, err := r.handleExternalResourceDeletion(
		ctx, lt, masterServiceName, "Master Service", masterService,
	); err != nil {
		return nil, nil, false, 0, err
	} else if shouldRequeue {
		return nil, nil, true, requeueAfter, nil
	}

	// Check for externally deleted master Job
	masterJob := &batchv1.Job{}
	masterJobName := lt.Name + "-master"
	if shouldRequeue, requeueAfter, err := r.handleExternalResourceDeletion(
		ctx, lt, masterJobName, "Master Job", masterJob,
	); err != nil {
		return nil, nil, false, 0, err
	} else if shouldRequeue {
		return nil, nil, true, requeueAfter, nil
	}

	// Check for externally deleted worker Job
	workerJob := &batchv1.Job{}
	workerJobName := lt.Name + "-worker"
	if shouldRequeue, requeueAfter, err := r.handleExternalResourceDeletion(
		ctx, lt, workerJobName, "Worker Job", workerJob,
	); err != nil {
		return nil, nil, false, 0, err
	} else if shouldRequeue {
		return nil, nil, true, requeueAfter, nil
	}

	// All resources exist
	return masterJob, workerJob, false, 0, nil
}

// shouldSkipStatusUpdate checks if the LocustTest is in a terminal state where status updates should be skipped.
// Returns true if Phase is Succeeded or Failed (terminal states).
func shouldSkipStatusUpdate(lt *locustv2.LocustTest) bool {
	return lt.Status.Phase == locustv2.PhaseSucceeded || lt.Status.Phase == locustv2.PhaseFailed
}

// reconcileStatus updates the LocustTest status based on owned Job states.
// Called when resources already exist and we need to track Job completion.
func (r *LocustTestReconciler) reconcileStatus(ctx context.Context, lt *locustv2.LocustTest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Check that all resources exist, handle external deletion if needed
	masterJob, workerJob, shouldRequeue, requeueAfter, err := r.checkResourcesExist(ctx, lt)
	if err != nil {
		return ctrl.Result{}, err
	}
	if shouldRequeue {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	// Don't update if already in terminal state (unless resources are missing — handled above)
	if shouldSkipStatusUpdate(lt) {
		return ctrl.Result{}, nil
	}

	// Check pod health before updating status from Jobs
	podHealthStatus, requeueAfter := r.checkPodHealth(ctx, lt)

	// Update status from Jobs (pass pod health to update logic)
	if err := r.updateStatusFromJobs(ctx, lt, masterJob, workerJob, podHealthStatus); err != nil {
		log.Error(err, "Failed to update status from Jobs")
		return ctrl.Result{}, fmt.Errorf("failed to update status from Jobs: %w", err)
	}

	// Requeue if pods are in grace period
	if requeueAfter > 0 {
		return ctrl.Result{RequeueAfter: requeueAfter}, nil
	}

	return ctrl.Result{}, nil
}

// mapPodToLocustTest maps a Pod event to the owning LocustTest reconcile request.
// Pods are owned by Jobs (Pod→Job), and Jobs are owned by LocustTests (Job→LocustTest).
// This function traverses the two-level owner reference chain: Pod → Job → LocustTest.
func (r *LocustTestReconciler) mapPodToLocustTest(ctx context.Context, obj client.Object) []reconcile.Request {
	log := logf.FromContext(ctx)

	// Step 1: Find the owning Job from the pod's owner references
	var jobName string
	for _, ref := range obj.GetOwnerReferences() {
		if ref.Kind == "Job" && ref.APIVersion == "batch/v1" {
			jobName = ref.Name
			break
		}
	}
	if jobName == "" {
		return nil
	}

	// Step 2: Fetch the Job to find the owning LocustTest
	job := &batchv1.Job{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      jobName,
	}, job); err != nil {
		log.V(1).Info("Failed to fetch Job for pod mapping", "pod", obj.GetName(), "job", jobName, "error", err)
		return nil
	}

	// Step 3: Find the LocustTest owner from the Job's owner references
	for _, ref := range job.GetOwnerReferences() {
		if ref.Kind == "LocustTest" {
			log.V(1).Info("Mapped pod to LocustTest", "pod", obj.GetName(), "job", jobName, "locustTest", ref.Name)
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Namespace: obj.GetNamespace(),
						Name:      ref.Name,
					},
				},
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&locustv2.LocustTest{}).
		Owns(&batchv1.Job{}).    // Watch owned Jobs for status updates
		Owns(&corev1.Service{}). // Watch owned Services
		Watches(                 // Watch pods via custom mapping (pods are owned by Jobs, not LocustTest)
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.mapPodToLocustTest),
		).
		Named("locusttest").
		Complete(r)
}

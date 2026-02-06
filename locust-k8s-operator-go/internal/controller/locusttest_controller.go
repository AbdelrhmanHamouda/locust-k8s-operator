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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
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

	// Initialize status on first reconcile
	if locustTest.Status.Phase == "" {
		r.initializeStatus(locustTest)
		if err := r.Status().Update(ctx, locustTest); err != nil {
			log.Error(err, "Failed to initialize status")
			return ctrl.Result{}, err
		}
		// Re-fetch after status update to get the latest resource version
		if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
			return ctrl.Result{}, err
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

	// Refetch to get latest resource version before status update
	if err := r.Get(ctx, client.ObjectKeyFromObject(lt), lt); err != nil {
		log.Error(err, "Failed to refetch LocustTest before status update")
		return ctrl.Result{}, err
	}

	// Update status after successful resource creation
	lt.Status.Phase = locustv2.PhaseRunning
	now := metav1.Now()
	lt.Status.StartTime = &now
	r.setReady(lt, true, locustv2.ReasonResourcesCreated, "All resources created")

	if err := r.Status().Update(ctx, lt); err != nil {
		log.Error(err, "Failed to update status after resource creation")
		return ctrl.Result{}, err
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

// reconcileStatus updates the LocustTest status based on owned Job states.
// Called when resources already exist and we need to track Job completion.
func (r *LocustTestReconciler) reconcileStatus(ctx context.Context, lt *locustv2.LocustTest) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Don't update if already in terminal state
	if lt.Status.Phase == locustv2.PhaseSucceeded || lt.Status.Phase == locustv2.PhaseFailed {
		return ctrl.Result{}, nil
	}

	// Fetch master Job to determine status
	masterJob := &batchv1.Job{}
	masterJobName := lt.Name + "-master"
	if err := r.Get(ctx, client.ObjectKey{Name: masterJobName, Namespace: lt.Namespace}, masterJob); err != nil {
		if apierrors.IsNotFound(err) {
			// Job not found yet, requeue
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	// Fetch worker Job for worker count
	workerJob := &batchv1.Job{}
	workerJobName := lt.Name + "-worker"
	if err := r.Get(ctx, client.ObjectKey{Name: workerJobName, Namespace: lt.Namespace}, workerJob); err != nil {
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

// SetupWithManager sets up the controller with the Manager.
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&locustv2.LocustTest{}).
		Owns(&batchv1.Job{}).    // Watch owned Jobs for status updates
		Owns(&corev1.Service{}). // Watch owned Services
		Named("locusttest").
		Complete(r)
}

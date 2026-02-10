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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

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

// updateStatusFromJobs derives status from the current state of owned Jobs.
func (r *LocustTestReconciler) updateStatusFromJobs(
	ctx context.Context,
	lt *locustv2.LocustTest,
	masterJob *batchv1.Job,
	workerJob *batchv1.Job,
) error {
	// Determine phase from master Job status
	newPhase := derivePhaseFromJob(masterJob)

	// Update phase if changed and emit events
	if lt.Status.Phase != newPhase {
		oldPhase := lt.Status.Phase
		lt.Status.Phase = newPhase

		// Emit event for significant transitions (CORE-26)
		switch newPhase {
		case locustv2.PhaseRunning:
			r.Recorder.Event(lt, corev1.EventTypeNormal, "TestStarted", "Load test execution started")
		case locustv2.PhaseSucceeded:
			r.Recorder.Event(lt, corev1.EventTypeNormal, "TestCompleted", "Load test completed successfully")
		case locustv2.PhaseFailed:
			r.Recorder.Event(lt, corev1.EventTypeWarning, "TestFailed", "Load test execution failed")
		case locustv2.PhasePending:
			// No event for Pending - it's the initial state or recovery state
		}

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
				r.setReady(lt, false, locustv2.ReasonResourcesFailed, "Test failed")
			}
		}

		_ = oldPhase // Keep variable to document this is a phase transition
	}

	// Update worker connection status (approximation from worker Job)
	if workerJob != nil {
		lt.Status.ConnectedWorkers = workerJob.Status.Active

		if lt.Status.ConnectedWorkers >= lt.Status.ExpectedWorkers {
			r.setCondition(lt, locustv2.ConditionTypeWorkersConnected,
				metav1.ConditionTrue, locustv2.ReasonAllWorkersConnected,
				fmt.Sprintf("%d/%d workers connected",
					lt.Status.ConnectedWorkers, lt.Status.ExpectedWorkers))
		} else {
			r.setCondition(lt, locustv2.ConditionTypeWorkersConnected,
				metav1.ConditionFalse, locustv2.ReasonWorkersMissing,
				fmt.Sprintf("%d/%d workers connected",
					lt.Status.ConnectedWorkers, lt.Status.ExpectedWorkers))
		}
	}

	// Update ObservedGeneration (CORE-25)
	lt.Status.ObservedGeneration = lt.Generation

	if err := r.Status().Update(ctx, lt); err != nil {
		return fmt.Errorf("failed to update status from Jobs: %w", err)
	}
	return nil
}

// derivePhaseFromJob determines the LocustTest phase from Job status.
func derivePhaseFromJob(job *batchv1.Job) locustv2.Phase {
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

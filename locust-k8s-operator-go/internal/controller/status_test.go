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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

func TestInitializeStatus(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: locustv2.LocustTestSpec{
			Worker: locustv2.WorkerSpec{
				Replicas: 5,
			},
		},
	}

	reconciler := &LocustTestReconciler{}
	reconciler.initializeStatus(lt)

	assert.Equal(t, locustv2.PhasePending, lt.Status.Phase)
	assert.Equal(t, int32(5), lt.Status.ExpectedWorkers)
	assert.Equal(t, int32(0), lt.Status.ConnectedWorkers)
	assert.Nil(t, lt.Status.StartTime)
	assert.Nil(t, lt.Status.CompletionTime)

	// Verify conditions are set
	require.Len(t, lt.Status.Conditions, 3)

	// Check Ready condition
	readyCondition := findCondition(lt.Status.Conditions, locustv2.ConditionTypeReady)
	require.NotNil(t, readyCondition)
	assert.Equal(t, metav1.ConditionFalse, readyCondition.Status)
	assert.Equal(t, locustv2.ReasonResourcesCreating, readyCondition.Reason)

	// Check WorkersConnected condition
	workersCondition := findCondition(lt.Status.Conditions, locustv2.ConditionTypeWorkersConnected)
	require.NotNil(t, workersCondition)
	assert.Equal(t, metav1.ConditionFalse, workersCondition.Status)
	assert.Equal(t, locustv2.ReasonWaitingForWorkers, workersCondition.Reason)

	// Check TestCompleted condition
	completedCondition := findCondition(lt.Status.Conditions, locustv2.ConditionTypeTestCompleted)
	require.NotNil(t, completedCondition)
	assert.Equal(t, metav1.ConditionFalse, completedCondition.Status)
	assert.Equal(t, locustv2.ReasonTestInProgress, completedCondition.Reason)
}

func TestSetCondition(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
	}

	reconciler := &LocustTestReconciler{}

	// Set initial condition
	reconciler.setCondition(lt, locustv2.ConditionTypeReady,
		metav1.ConditionFalse, locustv2.ReasonResourcesCreating, "Creating resources")

	require.Len(t, lt.Status.Conditions, 1)
	assert.Equal(t, locustv2.ConditionTypeReady, lt.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, lt.Status.Conditions[0].Status)
	assert.Equal(t, locustv2.ReasonResourcesCreating, lt.Status.Conditions[0].Reason)
	assert.Equal(t, int64(1), lt.Status.Conditions[0].ObservedGeneration)

	// Update the same condition
	reconciler.setCondition(lt, locustv2.ConditionTypeReady,
		metav1.ConditionTrue, locustv2.ReasonResourcesCreated, "All resources created")

	require.Len(t, lt.Status.Conditions, 1) // Should still be 1, not 2
	assert.Equal(t, metav1.ConditionTrue, lt.Status.Conditions[0].Status)
	assert.Equal(t, locustv2.ReasonResourcesCreated, lt.Status.Conditions[0].Reason)
}

func TestSetReady(t *testing.T) {
	tests := []struct {
		name           string
		ready          bool
		reason         string
		message        string
		expectedStatus metav1.ConditionStatus
	}{
		{
			name:           "set ready true",
			ready:          true,
			reason:         locustv2.ReasonResourcesCreated,
			message:        "All resources created",
			expectedStatus: metav1.ConditionTrue,
		},
		{
			name:           "set ready false",
			ready:          false,
			reason:         locustv2.ReasonResourcesFailed,
			message:        "Failed to create resources",
			expectedStatus: metav1.ConditionFalse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := &locustv2.LocustTest{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "default",
					Generation: 1,
				},
			}

			reconciler := &LocustTestReconciler{}
			reconciler.setReady(lt, tt.ready, tt.reason, tt.message)

			condition := findCondition(lt.Status.Conditions, locustv2.ConditionTypeReady)
			require.NotNil(t, condition)
			assert.Equal(t, tt.expectedStatus, condition.Status)
			assert.Equal(t, tt.reason, condition.Reason)
			assert.Equal(t, tt.message, condition.Message)
		})
	}
}

func TestDerivePhaseFromJob(t *testing.T) {
	tests := []struct {
		name          string
		job           *batchv1.Job
		expectedPhase string
	}{
		{
			name:          "nil job returns Pending",
			job:           nil,
			expectedPhase: locustv2.PhasePending,
		},
		{
			name: "job with no conditions returns Pending",
			job: &batchv1.Job{
				Status: batchv1.JobStatus{},
			},
			expectedPhase: locustv2.PhasePending,
		},
		{
			name: "job with active pods returns Running",
			job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Active: 1,
				},
			},
			expectedPhase: locustv2.PhaseRunning,
		},
		{
			name: "completed job returns Succeeded",
			job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobComplete,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedPhase: locustv2.PhaseSucceeded,
		},
		{
			name: "failed job returns Failed",
			job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobFailed,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedPhase: locustv2.PhaseFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase := derivePhaseFromJob(tt.job)
			assert.Equal(t, tt.expectedPhase, phase)
		})
	}
}

// findCondition finds a condition by type in a slice of conditions.
func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

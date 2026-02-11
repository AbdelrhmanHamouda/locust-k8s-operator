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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// healthyPodStatus returns a default healthy PodHealthStatus for tests that don't test pod health.
func healthyPodStatus() PodHealthStatus {
	return PodHealthStatus{Healthy: true, Reason: locustv2.ReasonPodsHealthy, Message: "All pods are healthy"}
}

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
	require.Len(t, lt.Status.Conditions, 4)

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

	// Check PodsHealthy condition
	podsHealthyCondition := findCondition(lt.Status.Conditions, locustv2.ConditionTypePodsHealthy)
	require.NotNil(t, podsHealthyCondition)
	assert.Equal(t, metav1.ConditionTrue, podsHealthyCondition.Status)
	assert.Equal(t, locustv2.ReasonPodsStarting, podsHealthyCondition.Reason)
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
		expectedPhase locustv2.Phase
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

// TestUpdateStatusFromJobs_FullStateMachine tests all phase transitions.
func TestUpdateStatusFromJobs_FullStateMachine(t *testing.T) {
	tests := []struct {
		name              string
		initialPhase      locustv2.Phase
		masterJob         *batchv1.Job
		workerJob         *batchv1.Job
		expectedPhase     locustv2.Phase
		expectedStartTime bool
		expectedComplete  bool
		expectedWorkers   int32
	}{
		{
			name:              "no jobs - stays Pending",
			initialPhase:      locustv2.PhasePending,
			masterJob:         nil,
			workerJob:         nil,
			expectedPhase:     locustv2.PhasePending,
			expectedStartTime: false,
			expectedComplete:  false,
			expectedWorkers:   0,
		},
		{
			name:         "active master job - transitions to Running",
			initialPhase: locustv2.PhasePending,
			masterJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Active: 1,
				},
			},
			workerJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Active: 3,
				},
			},
			expectedPhase:     locustv2.PhaseRunning,
			expectedStartTime: true,
			expectedComplete:  false,
			expectedWorkers:   3,
		},
		{
			name:         "completed master job - transitions to Succeeded",
			initialPhase: locustv2.PhaseRunning,
			masterJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobComplete,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			workerJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Active: 3,
				},
			},
			expectedPhase:     locustv2.PhaseSucceeded,
			expectedStartTime: false, // already set
			expectedComplete:  true,
			expectedWorkers:   3,
		},
		{
			name:         "failed master job - transitions to Failed",
			initialPhase: locustv2.PhaseRunning,
			masterJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobFailed,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			workerJob:         nil,
			expectedPhase:     locustv2.PhaseFailed,
			expectedStartTime: false,
			expectedComplete:  true,
			expectedWorkers:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test LocustTest
			lt := &locustv2.LocustTest{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test",
					Namespace:  "default",
					Generation: 2,
				},
				Spec: locustv2.LocustTestSpec{
					Worker: locustv2.WorkerSpec{
						Replicas: 5,
					},
				},
				Status: locustv2.LocustTestStatus{
					Phase:           tt.initialPhase,
					ExpectedWorkers: 5,
				},
			}

			reconciler, _ := newTestReconciler(lt)
			ctx := context.Background()

			// Call updateStatusFromJobs
			err := reconciler.updateStatusFromJobs(ctx, lt, tt.masterJob, tt.workerJob, healthyPodStatus())
			require.NoError(t, err)

			// Verify phase
			assert.Equal(t, tt.expectedPhase, lt.Status.Phase)

			// Verify ObservedGeneration
			assert.Equal(t, int64(2), lt.Status.ObservedGeneration)

			// Verify StartTime
			if tt.expectedStartTime {
				assert.NotNil(t, lt.Status.StartTime)
			}

			// Verify CompletionTime and TestCompleted condition
			if tt.expectedComplete {
				assert.NotNil(t, lt.Status.CompletionTime)
				completedCond := findCondition(lt.Status.Conditions, locustv2.ConditionTypeTestCompleted)
				require.NotNil(t, completedCond)
				assert.Equal(t, metav1.ConditionTrue, completedCond.Status)
				if tt.expectedPhase == locustv2.PhaseSucceeded {
					assert.Equal(t, locustv2.ReasonTestSucceeded, completedCond.Reason)
				} else {
					assert.Equal(t, locustv2.ReasonTestFailed, completedCond.Reason)
				}
			}

			// Verify worker connection count
			assert.Equal(t, tt.expectedWorkers, lt.Status.ConnectedWorkers)
		})
	}
}

// TestUpdateStatusFromJobs_PhaseTransitionEvents verifies events are emitted on phase changes.
func TestUpdateStatusFromJobs_PhaseTransitionEvents(t *testing.T) {
	tests := []struct {
		name          string
		initialPhase  locustv2.Phase
		masterJob     *batchv1.Job
		expectedEvent string
	}{
		{
			name:         "Pending to Running emits TestStarted",
			initialPhase: locustv2.PhasePending,
			masterJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Active: 1,
				},
			},
			expectedEvent: "TestStarted",
		},
		{
			name:         "Running to Succeeded emits TestCompleted",
			initialPhase: locustv2.PhaseRunning,
			masterJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobComplete,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedEvent: "TestCompleted",
		},
		{
			name:         "Running to Failed emits TestFailed",
			initialPhase: locustv2.PhaseRunning,
			masterJob: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobFailed,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedEvent: "TestFailed",
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
				Spec: locustv2.LocustTestSpec{
					Worker: locustv2.WorkerSpec{
						Replicas: 3,
					},
				},
				Status: locustv2.LocustTestStatus{
					Phase:           tt.initialPhase,
					ExpectedWorkers: 3,
				},
			}

			reconciler, recorder := newTestReconciler(lt)
			ctx := context.Background()

			// Call updateStatusFromJobs
			err := reconciler.updateStatusFromJobs(ctx, lt, tt.masterJob, nil, healthyPodStatus())
			require.NoError(t, err)

			// Verify event was emitted
			select {
			case event := <-recorder.Events:
				assert.Contains(t, event, tt.expectedEvent)
			default:
				t.Errorf("Expected event %s but none was emitted", tt.expectedEvent)
			}
		})
	}
}

// TestUpdateStatusFromJobs_NoEventOnSamePhase verifies no event is emitted when phase doesn't change.
func TestUpdateStatusFromJobs_NoEventOnSamePhase(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: locustv2.LocustTestSpec{
			Worker: locustv2.WorkerSpec{
				Replicas: 3,
			},
		},
		Status: locustv2.LocustTestStatus{
			Phase:           locustv2.PhaseRunning,
			ExpectedWorkers: 3,
		},
	}

	masterJob := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1, // Still running
		},
	}

	reconciler, recorder := newTestReconciler(lt)
	ctx := context.Background()

	// Call updateStatusFromJobs
	err := reconciler.updateStatusFromJobs(ctx, lt, masterJob, nil, healthyPodStatus())
	require.NoError(t, err)

	// Verify no event was emitted
	select {
	case event := <-recorder.Events:
		t.Errorf("Expected no event but got: %s", event)
	default:
		// Correct - no event emitted
	}

	// Phase should still be Running
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase)
}

// TestUpdateStatusFromJobs_ObservedGeneration verifies ObservedGeneration is always updated.
func TestUpdateStatusFromJobs_ObservedGeneration(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 5, // Higher generation
		},
		Spec: locustv2.LocustTestSpec{
			Worker: locustv2.WorkerSpec{
				Replicas: 3,
			},
		},
		Status: locustv2.LocustTestStatus{
			Phase:              locustv2.PhaseRunning,
			ObservedGeneration: 3, // Lower observed generation
			ExpectedWorkers:    3,
		},
	}

	masterJob := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1,
		},
	}

	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	// Call updateStatusFromJobs
	err := reconciler.updateStatusFromJobs(ctx, lt, masterJob, nil, healthyPodStatus())
	require.NoError(t, err)

	// Verify ObservedGeneration was updated to match Generation
	assert.Equal(t, lt.Generation, lt.Status.ObservedGeneration)
}

// TestDerivePhaseFromJob_TypeSafety verifies typed Phase return.
func TestDerivePhaseFromJob_TypeSafety(t *testing.T) {
	tests := []struct {
		name          string
		job           *batchv1.Job
		expectedPhase locustv2.Phase
	}{
		{
			name:          "nil job",
			job:           nil,
			expectedPhase: locustv2.PhasePending,
		},
		{
			name: "both Complete and Failed conditions - Complete wins",
			job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:   batchv1.JobComplete,
							Status: corev1.ConditionTrue,
						},
						{
							Type:   batchv1.JobFailed,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			expectedPhase: locustv2.PhaseSucceeded,
		},
		{
			name: "active job",
			job: &batchv1.Job{
				Status: batchv1.JobStatus{
					Active: 1,
				},
			},
			expectedPhase: locustv2.PhaseRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phase := derivePhaseFromJob(tt.job)
			assert.Equal(t, tt.expectedPhase, phase)

			// Verify it's the typed Phase type
			var _ locustv2.Phase = phase //nolint:staticcheck
		})
	}
}

// TestUpdateStatusFromJobs_WorkersConnectedCondition verifies WorkersConnected condition.
func TestUpdateStatusFromJobs_WorkersConnectedCondition(t *testing.T) {
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
		Status: locustv2.LocustTestStatus{
			Phase:           locustv2.PhaseRunning,
			ExpectedWorkers: 5,
		},
	}

	masterJob := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1,
		},
	}

	workerJob := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 5, // All workers connected
		},
	}

	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	// Call updateStatusFromJobs
	err := reconciler.updateStatusFromJobs(ctx, lt, masterJob, workerJob, healthyPodStatus())
	require.NoError(t, err)

	// Verify WorkersConnected condition
	workersCond := findCondition(lt.Status.Conditions, locustv2.ConditionTypeWorkersConnected)
	require.NotNil(t, workersCond)
	assert.Equal(t, metav1.ConditionTrue, workersCond.Status)
	assert.Equal(t, locustv2.ReasonAllWorkersConnected, workersCond.Reason)
	assert.Contains(t, workersCond.Message, "5/5 workers connected")
}

// TestUpdateStatusFromJobs_SpecDriftedCondition verifies SpecDrifted condition is set when Generation > 1.
func TestUpdateStatusFromJobs_SpecDriftedCondition(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 3, // Simulating spec edits
		},
		Spec: locustv2.LocustTestSpec{
			Worker: locustv2.WorkerSpec{
				Replicas: 5,
			},
		},
		Status: locustv2.LocustTestStatus{
			Phase:              locustv2.PhaseRunning,
			ObservedGeneration: 1,
			ExpectedWorkers:    5,
		},
	}

	masterJob := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1,
		},
	}

	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	// Call updateStatusFromJobs
	err := reconciler.updateStatusFromJobs(ctx, lt, masterJob, nil, healthyPodStatus())
	require.NoError(t, err)

	// Verify SpecDrifted condition exists with ConditionTrue
	specDriftedCond := findCondition(lt.Status.Conditions, locustv2.ConditionTypeSpecDrifted)
	require.NotNil(t, specDriftedCond)
	assert.Equal(t, metav1.ConditionTrue, specDriftedCond.Status)
	assert.Equal(t, locustv2.ReasonSpecChangeIgnored, specDriftedCond.Reason)
	assert.Contains(t, specDriftedCond.Message, "Delete and recreate")
}

// TestUpdateStatusFromJobs_NoSpecDriftedOnGeneration1 verifies SpecDrifted condition is NOT set when Generation == 1.
func TestUpdateStatusFromJobs_NoSpecDriftedOnGeneration1(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1, // No spec edits
		},
		Spec: locustv2.LocustTestSpec{
			Worker: locustv2.WorkerSpec{
				Replicas: 5,
			},
		},
		Status: locustv2.LocustTestStatus{
			Phase:              locustv2.PhaseRunning,
			ObservedGeneration: 1,
			ExpectedWorkers:    5,
		},
	}

	masterJob := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1,
		},
	}

	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	// Call updateStatusFromJobs
	err := reconciler.updateStatusFromJobs(ctx, lt, masterJob, nil, healthyPodStatus())
	require.NoError(t, err)

	// Verify SpecDrifted condition does NOT exist
	specDriftedCond := findCondition(lt.Status.Conditions, locustv2.ConditionTypeSpecDrifted)
	assert.Nil(t, specDriftedCond)
}

// TestUpdateStatusFromJobs_RetryOnConflict verifies that updateStatusFromJobs retries on 409 Conflict.
func TestUpdateStatusFromJobs_RetryOnConflict(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: locustv2.LocustTestSpec{
			Worker: locustv2.WorkerSpec{
				Replicas: 3,
			},
		},
		Status: locustv2.LocustTestStatus{
			Phase:           locustv2.PhaseRunning,
			ExpectedWorkers: 3,
		},
	}

	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt).
		WithStatusSubresource(&locustv2.LocustTest{}).
		Build()
	recorder := record.NewFakeRecorder(10)

	cc := &conflictOnUpdateClient{
		Client:        fakeClient,
		conflictCount: 2, // fail twice, succeed on third attempt
	}

	reconciler := &LocustTestReconciler{
		Client:   cc,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	masterJob := &batchv1.Job{
		Status: batchv1.JobStatus{Active: 1},
	}

	ctx := context.Background()
	err := reconciler.updateStatusFromJobs(ctx, lt, masterJob, nil, healthyPodStatus())
	require.NoError(t, err)

	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase)
	assert.Equal(t, 3, cc.updateCalls, "Expected 2 conflicts + 1 successful update")
}

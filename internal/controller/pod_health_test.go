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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// --- analyzePodFailure tests ---

func TestAnalyzePodFailure(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			TestFiles: &locustv2.TestFilesConfig{
				ConfigMapRef: "my-configmap",
			},
		},
	}

	tests := []struct {
		name           string
		pod            *corev1.Pod
		expectedNil    bool
		expectedReason string
	}{
		{
			name: "healthy pod returns nil",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "healthy-pod"},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodScheduled, Status: corev1.ConditionTrue},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "locust",
							State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
						},
					},
				},
			},
			expectedNil: true,
		},
		{
			name: "scheduling failure returns ReasonPodSchedulingError",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "unschedulable-pod"},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:    corev1.PodScheduled,
							Status:  corev1.ConditionFalse,
							Message: "0/3 nodes are available: insufficient cpu",
						},
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodSchedulingError,
		},
		{
			name: "failed init container returns failure",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "init-fail-pod"},
				Status: corev1.PodStatus{
					InitContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "init",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 1,
									Reason:   "Error",
								},
							},
						},
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodInitError,
		},
		{
			name: "failed main container returns failure",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "crash-pod"},
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "locust",
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason:  "CrashLoopBackOff",
									Message: "back-off 5m0s restarting failed container",
								},
							},
						},
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodCrashLoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzePodFailure(tt.pod, lt)
			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedReason, result.FailureType)
				assert.Equal(t, tt.pod.Name, result.Name)
			}
		})
	}
}

// --- analyzeContainerStatus tests ---

func TestAnalyzeContainerStatus(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			TestFiles: &locustv2.TestFilesConfig{
				ConfigMapRef: "my-configmap",
			},
		},
	}

	tests := []struct {
		name            string
		status          corev1.ContainerStatus
		isInitContainer bool
		expectedNil     bool
		expectedReason  string
	}{
		{
			name: "CreateContainerConfigError returns ReasonPodConfigError",
			status: corev1.ContainerStatus{
				Name: "locust",
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{
						Reason:  "CreateContainerConfigError",
						Message: `configmap "my-configmap" not found`,
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodConfigError,
		},
		{
			name: "ImagePullBackOff returns ReasonPodImagePullError",
			status: corev1.ContainerStatus{
				Name: "locust",
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: "Back-off pulling image",
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodImagePullError,
		},
		{
			name: "ErrImagePull returns ReasonPodImagePullError",
			status: corev1.ContainerStatus{
				Name: "locust",
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{
						Reason:  "ErrImagePull",
						Message: "Failed to pull image",
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodImagePullError,
		},
		{
			name: "CrashLoopBackOff returns ReasonPodCrashLoop",
			status: corev1.ContainerStatus{
				Name: "locust",
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{
						Reason:  "CrashLoopBackOff",
						Message: "back-off 5m0s restarting failed container",
					},
				},
			},
			expectedNil:    false,
			expectedReason: locustv2.ReasonPodCrashLoop,
		},
		{
			name: "terminated init container with non-zero exit returns ReasonPodInitError",
			status: corev1.ContainerStatus{
				Name: "init-container",
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						ExitCode: 1,
						Reason:   "Error",
					},
				},
			},
			isInitContainer: true,
			expectedNil:     false,
			expectedReason:  locustv2.ReasonPodInitError,
		},
		{
			name: "terminated main container with non-zero exit returns ReasonPodCrashLoop",
			status: corev1.ContainerStatus{
				Name: "locust",
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						ExitCode: 137,
						Reason:   "OOMKilled",
					},
				},
			},
			isInitContainer: false,
			expectedNil:     false,
			expectedReason:  locustv2.ReasonPodCrashLoop,
		},
		{
			name: "terminated with exit code 0 returns nil",
			status: corev1.ContainerStatus{
				Name: "locust",
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						ExitCode: 0,
						Reason:   "Completed",
					},
				},
			},
			expectedNil: true,
		},
		{
			name: "running container returns nil",
			status: corev1.ContainerStatus{
				Name:  "locust",
				State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
			},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeContainerStatus("test-pod", tt.status, tt.isInitContainer, lt)
			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedReason, result.FailureType)
				assert.Equal(t, "test-pod", result.Name)
			}
		})
	}
}

// --- extractConfigMapError tests ---

func TestExtractConfigMapError(t *testing.T) {
	tests := []struct {
		name     string
		errorMsg string
		lt       *locustv2.LocustTest
		expected string
	}{
		{
			name:     "matching ConfigMap name from spec returns enhanced message",
			errorMsg: `configmap "my-configmap" not found`,
			lt: &locustv2.LocustTest{
				Spec: locustv2.LocustTestSpec{
					TestFiles: &locustv2.TestFilesConfig{
						ConfigMapRef: "my-configmap",
					},
				},
			},
			expected: `ConfigMap not found (expected: my-configmap). configmap "my-configmap" not found`,
		},
		{
			name:     "non-matching ConfigMap name returns original message",
			errorMsg: `configmap "other-configmap" not found`,
			lt: &locustv2.LocustTest{
				Spec: locustv2.LocustTestSpec{
					TestFiles: &locustv2.TestFilesConfig{
						ConfigMapRef: "my-configmap",
					},
				},
			},
			expected: `configmap "other-configmap" not found`,
		},
		{
			name:     "generic error with spec ConfigMap returns enhanced message",
			errorMsg: "some generic config error",
			lt: &locustv2.LocustTest{
				Spec: locustv2.LocustTestSpec{
					TestFiles: &locustv2.TestFilesConfig{
						ConfigMapRef: "my-configmap",
					},
				},
			},
			expected: "ConfigMap not found (expected: my-configmap). some generic config error",
		},
		{
			name:     "generic error without spec ConfigMap returns original",
			errorMsg: "some generic config error",
			lt: &locustv2.LocustTest{
				Spec: locustv2.LocustTestSpec{
					TestFiles: &locustv2.TestFilesConfig{},
				},
			},
			expected: "some generic config error",
		},
		{
			name:     "nil TestFiles returns original message",
			errorMsg: "some config error",
			lt: &locustv2.LocustTest{
				Spec: locustv2.LocustTestSpec{},
			},
			expected: "some config error",
		},
		{
			name:     "LibConfigMapRef used when ConfigMapRef empty",
			errorMsg: "some generic config error",
			lt: &locustv2.LocustTest{
				Spec: locustv2.LocustTestSpec{
					TestFiles: &locustv2.TestFilesConfig{
						LibConfigMapRef: "my-lib-configmap",
					},
				},
			},
			expected: "ConfigMap not found (expected: my-lib-configmap). some generic config error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractConfigMapError(tt.errorMsg, tt.lt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- buildFailureMessage tests ---

func TestBuildFailureMessage(t *testing.T) {
	tests := []struct {
		name             string
		failures         []PodFailureInfo
		expectedType     string
		expectedContains []string
	}{
		{
			name:         "empty failures returns healthy",
			failures:     []PodFailureInfo{},
			expectedType: locustv2.ReasonPodsHealthy,
			expectedContains: []string{
				"All pods are healthy",
			},
		},
		{
			name: "single failure returns formatted message",
			failures: []PodFailureInfo{
				{Name: "pod-1", FailureType: locustv2.ReasonPodCrashLoop, ErrorMessage: "back-off restarting"},
			},
			expectedType: locustv2.ReasonPodCrashLoop,
			expectedContains: []string{
				"pod-1",
				"1 pod(s) affected",
				"back-off restarting",
			},
		},
		{
			name: "multiple failures same type grouped",
			failures: []PodFailureInfo{
				{Name: "pod-1", FailureType: locustv2.ReasonPodImagePullError, ErrorMessage: "pull error"},
				{Name: "pod-2", FailureType: locustv2.ReasonPodImagePullError, ErrorMessage: "pull error"},
			},
			expectedType: locustv2.ReasonPodImagePullError,
			expectedContains: []string{
				"pod-1",
				"pod-2",
				"2 pod(s) affected",
			},
		},
		{
			name: "mixed failure types prioritized - ConfigError first",
			failures: []PodFailureInfo{
				{Name: "pod-1", FailureType: locustv2.ReasonPodCrashLoop, ErrorMessage: "crash"},
				{Name: "pod-2", FailureType: locustv2.ReasonPodConfigError, ErrorMessage: "config missing"},
			},
			expectedType: locustv2.ReasonPodConfigError,
			expectedContains: []string{
				"pod-2",
				"config missing",
				"Create the ConfigMap",
			},
		},
		{
			name: "ConfigError includes recovery hint",
			failures: []PodFailureInfo{
				{Name: "pod-1", FailureType: locustv2.ReasonPodConfigError, ErrorMessage: "ConfigMap not found"},
			},
			expectedType: locustv2.ReasonPodConfigError,
			expectedContains: []string{
				"Create the ConfigMap and the pods will restart automatically",
			},
		},
		{
			name: "mixed types - ImagePull beats Scheduling",
			failures: []PodFailureInfo{
				{Name: "pod-1", FailureType: locustv2.ReasonPodSchedulingError, ErrorMessage: "no nodes"},
				{Name: "pod-2", FailureType: locustv2.ReasonPodImagePullError, ErrorMessage: "image not found"},
			},
			expectedType: locustv2.ReasonPodImagePullError,
			expectedContains: []string{
				"pod-2",
				"image not found",
			},
		},
		{
			name: "mixed types - Scheduling beats CrashLoop",
			failures: []PodFailureInfo{
				{Name: "pod-1", FailureType: locustv2.ReasonPodCrashLoop, ErrorMessage: "crash"},
				{Name: "pod-2", FailureType: locustv2.ReasonPodSchedulingError, ErrorMessage: "no nodes"},
			},
			expectedType: locustv2.ReasonPodSchedulingError,
			expectedContains: []string{
				"pod-2",
				"no nodes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failureType, message := buildFailureMessage(tt.failures)
			assert.Equal(t, tt.expectedType, failureType)
			for _, substr := range tt.expectedContains {
				assert.Contains(t, message, substr)
			}
		})
	}
}

// --- findOldestPodCreationTime tests ---

func TestFindOldestPodCreationTime(t *testing.T) {
	now := time.Now()

	t.Run("empty list returns approximately now", func(t *testing.T) {
		result := findOldestPodCreationTime([]corev1.Pod{})
		assert.WithinDuration(t, now, result, 2*time.Second)
	})

	t.Run("single pod returns its creation time", func(t *testing.T) {
		creationTime := now.Add(-10 * time.Minute)
		pods := []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(creationTime)}},
		}
		result := findOldestPodCreationTime(pods)
		assert.Equal(t, creationTime, result)
	})

	t.Run("multiple pods returns oldest", func(t *testing.T) {
		oldest := now.Add(-30 * time.Minute)
		middle := now.Add(-15 * time.Minute)
		newest := now.Add(-5 * time.Minute)
		pods := []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(middle)}},
			{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(oldest)}},
			{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(newest)}},
		}
		result := findOldestPodCreationTime(pods)
		assert.Equal(t, oldest, result)
	})
}

// --- checkPodHealth tests (require fake client) ---

// listErrorClient wraps a client.Client and returns an error on List calls.
type listErrorClient struct {
	client.Client
	listError error
}

func (c *listErrorClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if c.listError != nil {
		return c.listError
	}
	return c.Client.List(ctx, list, opts...)
}

func TestCheckPodHealth_NoPods(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	status, requeueDuration := reconciler.checkPodHealth(ctx, lt)
	assert.True(t, status.Healthy)
	assert.Equal(t, locustv2.ReasonPodsStarting, status.Reason)
	assert.Equal(t, "Waiting for pods to be created", status.Message)
	assert.Equal(t, time.Duration(0), requeueDuration)
}

func TestCheckPodHealth_PodsInGracePeriod(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	// Create a pod that was just created (within grace period)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-master-abc",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(time.Now()),
			Labels: map[string]string{
				"performance-test-name": "test",
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "locust",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "back-off",
						},
					},
				},
			},
		},
	}

	reconciler, _ := newTestReconciler(lt, pod)
	ctx := context.Background()

	status, requeueDuration := reconciler.checkPodHealth(ctx, lt)
	assert.True(t, status.Healthy)
	assert.True(t, status.InGracePeriod)
	assert.Equal(t, locustv2.ReasonPodsStarting, status.Reason)
	assert.Greater(t, requeueDuration, time.Duration(0))
}

func TestCheckPodHealth_AllHealthyAfterGracePeriod(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	// Create a pod that was created well before the grace period
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-master-abc",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
			Labels: map[string]string{
				"performance-test-name": "test",
			},
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodScheduled, Status: corev1.ConditionTrue},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:  "locust",
					State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				},
			},
		},
	}

	reconciler, _ := newTestReconciler(lt, pod)
	ctx := context.Background()

	status, requeueDuration := reconciler.checkPodHealth(ctx, lt)
	assert.True(t, status.Healthy)
	assert.Equal(t, locustv2.ReasonPodsHealthy, status.Reason)
	assert.Equal(t, "All pods are healthy", status.Message)
	assert.Equal(t, time.Duration(0), requeueDuration)
}

func TestCheckPodHealth_FailedPodsAfterGracePeriod(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			TestFiles: &locustv2.TestFilesConfig{
				ConfigMapRef: "my-configmap",
			},
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-master-abc",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
			Labels: map[string]string{
				"performance-test-name": "test",
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "locust",
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "CreateContainerConfigError",
							Message: `configmap "my-configmap" not found`,
						},
					},
				},
			},
		},
	}

	reconciler, _ := newTestReconciler(lt, pod)
	ctx := context.Background()

	status, requeueDuration := reconciler.checkPodHealth(ctx, lt)
	assert.False(t, status.Healthy)
	assert.Equal(t, locustv2.ReasonPodConfigError, status.Reason)
	require.Len(t, status.FailedPods, 1)
	assert.Equal(t, "test-master-abc", status.FailedPods[0].Name)
	assert.Equal(t, time.Duration(0), requeueDuration)
}

func TestCheckPodHealth_ListError(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt).
		WithStatusSubresource(&locustv2.LocustTest{}).
		Build()

	errClient := &listErrorClient{
		Client:    fakeClient,
		listError: fmt.Errorf("connection refused"),
	}

	reconciler := &LocustTestReconciler{
		Client: errClient,
		Scheme: scheme,
		Config: newTestOperatorConfig(),
	}

	ctx := context.Background()

	status, requeueDuration := reconciler.checkPodHealth(ctx, lt)
	assert.True(t, status.Healthy, "should return healthy on list error to avoid blocking")
	assert.Equal(t, locustv2.ReasonPodsHealthy, status.Reason)
	assert.Equal(t, time.Duration(0), requeueDuration)
}

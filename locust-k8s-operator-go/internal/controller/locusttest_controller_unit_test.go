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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

// newTestScheme creates a scheme with all required types registered.
func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = locustv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	return scheme
}

// newTestReconciler creates a reconciler with a fake client for testing.
func newTestReconciler(objs ...client.Object) (*LocustTestReconciler, *record.FakeRecorder) {
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
	recorder := record.NewFakeRecorder(10)

	return &LocustTestReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}, recorder
}

// newTestOperatorConfig creates a test operator configuration.
func newTestOperatorConfig() *config.OperatorConfig {
	return &config.OperatorConfig{
		PodCPURequest:              "250m",
		PodMemRequest:              "128Mi",
		PodEphemeralStorageRequest: "30M",
		PodCPULimit:                "1000m",
		PodMemLimit:                "1024Mi",
		PodEphemeralStorageLimit:   "50M",

		MetricsExporterImage:                   "containersol/locust_exporter:v0.5.0",
		MetricsExporterPort:                    9646,
		MetricsExporterPullPolicy:              "Always",
		MetricsExporterCPURequest:              "250m",
		MetricsExporterMemRequest:              "128Mi",
		MetricsExporterEphemeralStorageRequest: "30M",
		MetricsExporterCPULimit:                "1000m",
		MetricsExporterMemLimit:                "1024Mi",
		MetricsExporterEphemeralStorageLimit:   "50M",

		KafkaBootstrapServers: "localhost:9092",
		KafkaSecurityEnabled:  false,
	}
}

// newTestLocustTestCR creates a test LocustTest CR.
func newTestLocustTestCR(name, namespace string) *locustv1.LocustTest {
	return &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
			UID:        "test-uid-12345",
		},
		Spec: locustv1.LocustTestSpec{
			MasterCommandSeed: "locust -f /lotest/src/test.py",
			WorkerCommandSeed: "locust -f /lotest/src/test.py",
			WorkerReplicas:    3,
			Image:             "locustio/locust:latest",
			ConfigMap:         "test-configmap",
		},
	}
}

func TestReconcile_NotFound(t *testing.T) {
	reconciler, _ := newTestReconciler()

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_CreateResources(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	reconciler, recorder := newTestReconciler(lt)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Verify Service created
	svc := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	assert.NoError(t, err)
	assert.Equal(t, "my-test-master", svc.Name)

	// Verify master Job created
	masterJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterJob)
	assert.NoError(t, err)

	// Verify worker Job created
	workerJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-worker",
		Namespace: "default",
	}, workerJob)
	assert.NoError(t, err)

	// Verify events recorded (3 resources created)
	eventCount := 0
	for {
		select {
		case event := <-recorder.Events:
			assert.Contains(t, event, "Created")
			eventCount++
		default:
			goto done
		}
	}
done:
	assert.Equal(t, 3, eventCount, "Expected 3 creation events")
}

func TestReconcile_NoOpOnUpdate(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	lt.Generation = 2 // Simulates an update
	reconciler, _ := newTestReconciler(lt)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Verify no resources created
	svc := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	assert.True(t, apierrors.IsNotFound(err), "Service should not be created on update")
}

func TestReconcile_OwnerReferences(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify owner reference on Service
	svc := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	require.NoError(t, err)

	require.Len(t, svc.OwnerReferences, 1)
	assert.Equal(t, "my-test", svc.OwnerReferences[0].Name)
	assert.Equal(t, "LocustTest", svc.OwnerReferences[0].Kind)

	// Verify owner reference on master Job
	masterJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterJob)
	require.NoError(t, err)
	require.Len(t, masterJob.OwnerReferences, 1)
	assert.Equal(t, "my-test", masterJob.OwnerReferences[0].Name)

	// Verify owner reference on worker Job
	workerJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-worker",
		Namespace: "default",
	}, workerJob)
	require.NoError(t, err)
	require.Len(t, workerJob.OwnerReferences, 1)
	assert.Equal(t, "my-test", workerJob.OwnerReferences[0].Name)
}

func TestReconcile_IdempotentCreate(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	reconciler, _ := newTestReconciler(lt)

	// First reconcile - creates resources
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Second reconcile with same generation should handle existing resources gracefully
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	assert.NoError(t, err, "Should not error even if resources exist")
}

func TestReconcile_WithDifferentGenerations(t *testing.T) {
	tests := []struct {
		name                  string
		generation            int64
		expectResourceCreated bool
	}{
		{
			name:                  "generation 1 creates resources",
			generation:            1,
			expectResourceCreated: true,
		},
		{
			name:                  "generation 2 is NO-OP",
			generation:            2,
			expectResourceCreated: false,
		},
		{
			name:                  "generation 10 is NO-OP",
			generation:            10,
			expectResourceCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := newTestLocustTestCR("test-gen", "default")
			lt.Generation = tt.generation
			reconciler, _ := newTestReconciler(lt)

			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-gen",
					Namespace: "default",
				},
			})
			require.NoError(t, err)

			svc := &corev1.Service{}
			err = reconciler.Get(context.Background(), types.NamespacedName{
				Name:      "test-gen-master",
				Namespace: "default",
			}, svc)

			if tt.expectResourceCreated {
				assert.NoError(t, err, "Service should be created")
			} else {
				assert.True(t, apierrors.IsNotFound(err), "Service should not be created")
			}
		})
	}
}

func TestReconcile_VerifyServiceConfiguration(t *testing.T) {
	lt := newTestLocustTestCR("svc-test", "default")
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "svc-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	svc := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "svc-test-master",
		Namespace: "default",
	}, svc)
	require.NoError(t, err)

	// Service type defaults to ClusterIP in K8s (empty string in fake client is OK)
	// The actual K8s API will default empty to ClusterIP

	// Verify service has correct selector
	assert.Equal(t, "svc-test-master", svc.Spec.Selector["performance-test-pod-name"])

	// Verify service has ports (3 ports: 5557, 5558, metrics)
	assert.Len(t, svc.Spec.Ports, 3)
}

func TestReconcile_VerifyMasterJobConfiguration(t *testing.T) {
	lt := newTestLocustTestCR("job-test", "default")
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "job-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	job := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "job-test-master",
		Namespace: "default",
	}, job)
	require.NoError(t, err)

	// Verify master parallelism is 1
	require.NotNil(t, job.Spec.Parallelism)
	assert.Equal(t, int32(1), *job.Spec.Parallelism)

	// Verify master has 2 containers (locust + metrics exporter)
	assert.Len(t, job.Spec.Template.Spec.Containers, 2)

	// Verify RestartPolicy is Never
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)
}

func TestReconcile_VerifyWorkerJobConfiguration(t *testing.T) {
	lt := newTestLocustTestCR("worker-test", "default")
	lt.Spec.WorkerReplicas = 5
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "worker-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	job := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "worker-test-worker",
		Namespace: "default",
	}, job)
	require.NoError(t, err)

	// Verify worker parallelism equals WorkerReplicas
	require.NotNil(t, job.Spec.Parallelism)
	assert.Equal(t, int32(5), *job.Spec.Parallelism)

	// Verify worker has 1 container (no metrics exporter)
	assert.Len(t, job.Spec.Template.Spec.Containers, 1)
}

func TestReconcile_EventRecording(t *testing.T) {
	lt := newTestLocustTestCR("event-test", "default")
	reconciler, recorder := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "event-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Collect all events
	var events []string
	for {
		select {
		case event := <-recorder.Events:
			events = append(events, event)
		default:
			goto done
		}
	}
done:

	// Should have 3 events: Service, Master Job, Worker Job
	assert.Len(t, events, 3)

	// Verify event content
	serviceEventFound := false
	masterJobEventFound := false
	workerJobEventFound := false

	for _, event := range events {
		if contains(event, "Service") && contains(event, "event-test-master") {
			serviceEventFound = true
		}
		if contains(event, "Job") && contains(event, "event-test-master") {
			masterJobEventFound = true
		}
		if contains(event, "Job") && contains(event, "event-test-worker") {
			workerJobEventFound = true
		}
	}

	assert.True(t, serviceEventFound, "Service creation event should be recorded")
	assert.True(t, masterJobEventFound, "Master Job creation event should be recorded")
	assert.True(t, workerJobEventFound, "Worker Job creation event should be recorded")
}

func TestReconcile_WithCustomLabels(t *testing.T) {
	lt := newTestLocustTestCR("label-test", "default")
	lt.Spec.Labels = &locustv1.PodLabels{
		Master: map[string]string{
			"custom-label": "master-value",
		},
		Worker: map[string]string{
			"custom-label": "worker-value",
		},
	}
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "label-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify master job has custom label
	masterJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "label-test-master",
		Namespace: "default",
	}, masterJob)
	require.NoError(t, err)
	assert.Equal(t, "master-value", masterJob.Spec.Template.Labels["custom-label"])

	// Verify worker job has custom label
	workerJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "label-test-worker",
		Namespace: "default",
	}, workerJob)
	require.NoError(t, err)
	assert.Equal(t, "worker-value", workerJob.Spec.Template.Labels["custom-label"])
}

func TestReconcile_WithImagePullSecrets(t *testing.T) {
	lt := newTestLocustTestCR("secret-test", "default")
	lt.Spec.ImagePullSecrets = []string{"my-registry-secret"}
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "secret-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	job := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "secret-test-master",
		Namespace: "default",
	}, job)
	require.NoError(t, err)

	require.Len(t, job.Spec.Template.Spec.ImagePullSecrets, 1)
	assert.Equal(t, "my-registry-secret", job.Spec.Template.Spec.ImagePullSecrets[0].Name)
}

func TestReconcile_WithLibConfigMap(t *testing.T) {
	lt := newTestLocustTestCR("lib-test", "default")
	lt.Spec.LibConfigMap = "locust-lib"
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "lib-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	job := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "lib-test-master",
		Namespace: "default",
	}, job)
	require.NoError(t, err)

	// Should have 2 volumes: configmap and lib
	assert.Len(t, job.Spec.Template.Spec.Volumes, 2)

	// Find lib volume (LibVolumeName constant is "lib")
	var libVolumeFound bool
	for _, v := range job.Spec.Template.Spec.Volumes {
		if v.Name == "lib" {
			libVolumeFound = true
			assert.Equal(t, "locust-lib", v.ConfigMap.Name)
		}
	}
	assert.True(t, libVolumeFound, "Lib volume should exist")
}

func TestReconcile_MultipleNamespaces(t *testing.T) {
	lt1 := newTestLocustTestCR("test1", "namespace-a")
	lt2 := newTestLocustTestCR("test2", "namespace-b")
	reconciler, _ := newTestReconciler(lt1, lt2)

	// Reconcile first CR
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test1",
			Namespace: "namespace-a",
		},
	})
	require.NoError(t, err)

	// Reconcile second CR
	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test2",
			Namespace: "namespace-b",
		},
	})
	require.NoError(t, err)

	// Verify resources in namespace-a
	svc1 := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "test1-master",
		Namespace: "namespace-a",
	}, svc1)
	assert.NoError(t, err)

	// Verify resources in namespace-b
	svc2 := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "test2-master",
		Namespace: "namespace-b",
	}, svc2)
	assert.NoError(t, err)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// errorClient is a fake client that returns errors for testing error paths.
type errorClient struct {
	client.Client
	getError    error
	createError error
}

func (e *errorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if e.getError != nil {
		return e.getError
	}
	return e.Client.Get(ctx, key, obj, opts...)
}

func (e *errorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if e.createError != nil {
		return e.createError
	}
	return e.Client.Create(ctx, obj, opts...)
}

func TestReconcile_GetError(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	recorder := record.NewFakeRecorder(10)

	// Wrap with error client that returns an error on Get
	errClient := &errorClient{
		Client:   fakeClient,
		getError: apierrors.NewInternalError(fmt.Errorf("internal server error")),
	}

	reconciler := &LocustTestReconciler{
		Client:   errClient,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test",
			Namespace: "default",
		},
	})

	assert.Error(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_CreateServiceError(t *testing.T) {
	lt := newTestLocustTestCR("error-test", "default")
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt).
		Build()
	recorder := record.NewFakeRecorder(10)

	// Wrap with error client that returns an error on Create
	errClient := &errorClient{
		Client:      fakeClient,
		createError: apierrors.NewInternalError(fmt.Errorf("failed to create")),
	}

	reconciler := &LocustTestReconciler{
		Client:   errClient,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "error-test",
			Namespace: "default",
		},
	})

	assert.Error(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

// sequentialErrorClient returns error only after N successful creates
type sequentialErrorClient struct {
	client.Client
	createCount      int
	errorAfterCreate int
	createError      error
}

func (s *sequentialErrorClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	s.createCount++
	if s.createCount > s.errorAfterCreate {
		return s.createError
	}
	return s.Client.Create(ctx, obj, opts...)
}

func TestReconcile_CreateMasterJobError(t *testing.T) {
	lt := newTestLocustTestCR("job-error-test", "default")
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt).
		Build()
	recorder := record.NewFakeRecorder(10)

	// Error after first create (Service succeeds, Master Job fails)
	errClient := &sequentialErrorClient{
		Client:           fakeClient,
		errorAfterCreate: 1,
		createError:      apierrors.NewInternalError(fmt.Errorf("failed to create job")),
	}

	reconciler := &LocustTestReconciler{
		Client:   errClient,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "job-error-test",
			Namespace: "default",
		},
	})

	assert.Error(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_CreateWorkerJobError(t *testing.T) {
	lt := newTestLocustTestCR("worker-error-test", "default")
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt).
		Build()
	recorder := record.NewFakeRecorder(10)

	// Error after second create (Service and Master Job succeed, Worker Job fails)
	errClient := &sequentialErrorClient{
		Client:           fakeClient,
		errorAfterCreate: 2,
		createError:      apierrors.NewInternalError(fmt.Errorf("failed to create worker job")),
	}

	reconciler := &LocustTestReconciler{
		Client:   errClient,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "worker-error-test",
			Namespace: "default",
		},
	})

	assert.Error(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestCreateResource_SetControllerReferenceError(t *testing.T) {
	// Create a LocustTest without a UID - this causes SetControllerReference to fail
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "no-uid-test",
			Namespace:  "default",
			Generation: 1,
			// UID intentionally not set
		},
		Spec: locustv1.LocustTestSpec{
			MasterCommandSeed: "locust -f /lotest/src/test.py",
			WorkerCommandSeed: "locust -f /lotest/src/test.py",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
			ConfigMap:         "test-configmap",
		},
	}

	// Use a scheme without LocustTest to cause SetControllerReference to fail
	badScheme := runtime.NewScheme()
	_ = batchv1.AddToScheme(badScheme)
	_ = corev1.AddToScheme(badScheme)
	// Note: NOT adding locustv1 to scheme

	fakeClient := fake.NewClientBuilder().
		WithScheme(badScheme).
		Build()
	recorder := record.NewFakeRecorder(10)

	reconciler := &LocustTestReconciler{
		Client:   fakeClient,
		Scheme:   badScheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
		},
	}

	err := reconciler.createResource(context.Background(), lt, svc, "Service")
	assert.Error(t, err)
}

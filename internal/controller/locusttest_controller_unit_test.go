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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

// newTestScheme creates a scheme with all required types registered.
func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = locustv2.AddToScheme(scheme)
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
		WithStatusSubresource(&locustv2.LocustTest{}).
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
func newTestLocustTestCR(name, namespace string) *locustv2.LocustTest {
	return &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
			UID:        "test-uid-12345",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
			TestFiles: &locustv2.TestFilesConfig{
				ConfigMapRef: "test-configmap",
			},
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

func TestReconcile_PendingPhaseCreatesResourcesRegardlessOfGeneration(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	lt.Generation = 2 // Simulates an update while still in Pending phase
	reconciler, _ := newTestReconciler(lt)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Verify resources ARE created (fixes operator-restart edge case)
	svc := &corev1.Service{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	assert.NoError(t, err, "Service should be created even with generation > 1 when phase is Pending")

	// Verify master Job created
	masterJob := &batchv1.Job{}
	err = reconciler.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterJob)
	assert.NoError(t, err, "Master Job should be created")
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
		name       string
		generation int64
	}{
		{
			name:       "generation 1 pending creates resources",
			generation: 1,
		},
		{
			name:       "generation 2 pending creates resources",
			generation: 2,
		},
		{
			name:       "generation 10 pending creates resources",
			generation: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := newTestLocustTestCR("test-gen", "default")
			lt.Generation = tt.generation
			// Phase defaults to empty string (Pending) from newTestLocustTestCR
			reconciler, _ := newTestReconciler(lt)

			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-gen",
					Namespace: "default",
				},
			})
			require.NoError(t, err)

			// All pending CRs create resources regardless of generation
			svc := &corev1.Service{}
			err = reconciler.Get(context.Background(), types.NamespacedName{
				Name:      "test-gen-master",
				Namespace: "default",
			}, svc)
			assert.NoError(t, err, "Service should be created for Pending phase regardless of generation")
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
	lt.Spec.Worker.Replicas = 5
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
		if strings.Contains(event, "Service") && strings.Contains(event, "event-test-master") {
			serviceEventFound = true
		}
		if strings.Contains(event, "Job") && strings.Contains(event, "event-test-master") {
			masterJobEventFound = true
		}
		if strings.Contains(event, "Job") && strings.Contains(event, "event-test-worker") {
			workerJobEventFound = true
		}
	}

	assert.True(t, serviceEventFound, "Service creation event should be recorded")
	assert.True(t, masterJobEventFound, "Master Job creation event should be recorded")
	assert.True(t, workerJobEventFound, "Worker Job creation event should be recorded")
}

func TestReconcile_WithCustomLabels(t *testing.T) {
	lt := newTestLocustTestCR("label-test", "default")
	lt.Spec.Master.Labels = map[string]string{
		"custom-label": "master-value",
	}
	lt.Spec.Worker.Labels = map[string]string{
		"custom-label": "worker-value",
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
	lt.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
		{Name: "my-registry-secret"},
	}
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
	lt.Spec.TestFiles.LibConfigMapRef = "locust-lib"
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

	// Find lib volume (LibVolumeName constant is "locust-lib")
	var libVolumeFound bool
	for _, v := range job.Spec.Template.Spec.Volumes {
		if v.Name == "locust-lib" {
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
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "no-uid-test",
			Namespace:  "default",
			Generation: 1,
			// UID intentionally not set
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 1,
			},
			TestFiles: &locustv2.TestFilesConfig{
				ConfigMapRef: "test-configmap",
			},
		},
	}

	// Use a scheme without LocustTest to cause SetControllerReference to fail
	badScheme := runtime.NewScheme()
	_ = batchv1.AddToScheme(badScheme)
	_ = corev1.AddToScheme(badScheme)
	// Note: NOT adding locustv2 to scheme

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

func TestReconcile_ExternalDeletion_MasterService(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	reconciler, recorder := newTestReconciler(lt)
	ctx := context.Background()

	// First reconcile - creates resources and transitions to Running
	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Drain creation events from first reconcile
	for i := 0; i < 3; i++ {
		select {
		case <-recorder.Events:
			// Drain creation events (Service, Master Job, Worker Job)
		default:
		}
	}

	// Refetch CR to get updated status
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase)

	// Manually delete the master Service
	masterService := &corev1.Service{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterService)
	require.NoError(t, err)
	err = reconciler.Delete(ctx, masterService)
	require.NoError(t, err)

	// Second reconcile - should detect deletion and reset to Pending
	result, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)
	assert.Greater(t, result.RequeueAfter, time.Duration(0), "Should requeue after detecting deletion")

	// Check that Warning event was recorded
	select {
	case event := <-recorder.Events:
		assert.Contains(t, event, "Warning")
		assert.Contains(t, event, "ResourceDeleted")
		assert.Contains(t, event, "Service")
		assert.Contains(t, event, "my-test-master")
	default:
		t.Fatal("Expected Warning event for external deletion")
	}

	// Refetch CR to check Phase was reset
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhasePending, lt.Status.Phase, "Phase should be reset to Pending")

	// Third reconcile - should recreate the Service (self-healing)
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify Service exists again
	masterService = &corev1.Service{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterService)
	assert.NoError(t, err, "Service should be recreated")

	// Refetch CR to check Phase transitioned back to Running
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase, "Phase should be Running after recreation")
}

func TestReconcile_ExternalDeletion_MasterJob(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	reconciler, recorder := newTestReconciler(lt)
	ctx := context.Background()

	// First reconcile - creates resources and transitions to Running
	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Drain creation events from first reconcile
	for i := 0; i < 3; i++ {
		select {
		case <-recorder.Events:
			// Drain creation events (Service, Master Job, Worker Job)
		default:
		}
	}

	// Refetch CR to get updated status
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase)

	// Manually delete the master Job
	masterJob := &batchv1.Job{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterJob)
	require.NoError(t, err)
	err = reconciler.Delete(ctx, masterJob)
	require.NoError(t, err)

	// Second reconcile - should detect deletion and reset to Pending
	result, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)
	assert.Greater(t, result.RequeueAfter, time.Duration(0), "Should requeue after detecting deletion")

	// Check that Warning event was recorded
	select {
	case event := <-recorder.Events:
		assert.Contains(t, event, "Warning")
		assert.Contains(t, event, "ResourceDeleted")
		assert.Contains(t, event, "Job")
		assert.Contains(t, event, "my-test-master")
	default:
		t.Fatal("Expected Warning event for external deletion")
	}

	// Refetch CR to check Phase was reset
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhasePending, lt.Status.Phase, "Phase should be reset to Pending")

	// Third reconcile - should recreate the Job (self-healing)
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify Job exists again
	masterJob = &batchv1.Job{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterJob)
	assert.NoError(t, err, "Master Job should be recreated")

	// Refetch CR to check Phase transitioned back to Running
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase, "Phase should be Running after recreation")
}

func TestReconcile_ExternalDeletion_WorkerJob(t *testing.T) {
	lt := newTestLocustTestCR("my-test", "default")
	reconciler, recorder := newTestReconciler(lt)
	ctx := context.Background()

	// First reconcile - creates resources and transitions to Running
	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Drain creation events from first reconcile
	for i := 0; i < 3; i++ {
		select {
		case <-recorder.Events:
			// Drain creation events (Service, Master Job, Worker Job)
		default:
		}
	}

	// Refetch CR to get updated status
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase)

	// Manually delete the worker Job
	workerJob := &batchv1.Job{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test-worker",
		Namespace: "default",
	}, workerJob)
	require.NoError(t, err)
	err = reconciler.Delete(ctx, workerJob)
	require.NoError(t, err)

	// Second reconcile - should detect deletion and reset to Pending
	result, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)
	assert.Greater(t, result.RequeueAfter, time.Duration(0), "Should requeue after detecting deletion")

	// Check that Warning event was recorded
	select {
	case event := <-recorder.Events:
		assert.Contains(t, event, "Warning")
		assert.Contains(t, event, "ResourceDeleted")
		assert.Contains(t, event, "Job")
		assert.Contains(t, event, "my-test-worker")
	default:
		t.Fatal("Expected Warning event for external deletion")
	}

	// Refetch CR to check Phase was reset
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhasePending, lt.Status.Phase, "Phase should be reset to Pending")

	// Third reconcile - should recreate the Job (self-healing)
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify Job exists again
	workerJob = &batchv1.Job{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test-worker",
		Namespace: "default",
	}, workerJob)
	assert.NoError(t, err, "Worker Job should be recreated")

	// Refetch CR to check Phase transitioned back to Running
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "my-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase, "Phase should be Running after recreation")
}

// conflictOnUpdateClient wraps a client.Client and returns 409 Conflict errors
// on the first N Status().Update() calls, then delegates to the real client.
type conflictOnUpdateClient struct {
	client.Client
	conflictCount int // number of conflicts to return before succeeding
	updateCalls   int // tracks total Status().Update() calls
}

func (c *conflictOnUpdateClient) Status() client.SubResourceWriter {
	return &conflictStatusWriter{
		SubResourceWriter: c.Client.Status(),
		parent:            c,
	}
}

type conflictStatusWriter struct {
	client.SubResourceWriter
	parent *conflictOnUpdateClient
}

func (w *conflictStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	w.parent.updateCalls++
	if w.parent.updateCalls <= w.parent.conflictCount {
		return apierrors.NewConflict(
			schema.GroupResource{Group: "locust.io", Resource: "locusttests"},
			obj.GetName(),
			fmt.Errorf("the object has been modified"),
		)
	}
	return w.SubResourceWriter.Update(ctx, obj, opts...)
}

func TestCreateResources_RetryOnConflict(t *testing.T) {
	lt := newTestLocustTestCR("conflict-test", "default")
	// Pre-set Phase to Pending so initializeStatus is skipped
	lt.Status.Phase = locustv2.PhasePending
	lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas

	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt).
		WithStatusSubresource(&locustv2.LocustTest{}).
		Build()
	recorder := record.NewFakeRecorder(10)

	cc := &conflictOnUpdateClient{
		Client:        fakeClient,
		conflictCount: 1, // fail once, succeed on second attempt
	}

	reconciler := &LocustTestReconciler{
		Client:   cc,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "conflict-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
	assert.Equal(t, 2, cc.updateCalls, "Expected 1 conflict + 1 successful update")

	// Verify status was correctly set despite the conflict
	err = cc.Get(context.Background(), types.NamespacedName{
		Name:      "conflict-test",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhaseRunning, lt.Status.Phase)
}

func TestReconcile_ExternalDeletion_RetryOnConflict(t *testing.T) {
	lt := newTestLocustTestCR("conflict-del", "default")
	// Pre-set to Running phase with resources "already created"
	lt.Status.Phase = locustv2.PhaseRunning
	lt.Status.ExpectedWorkers = lt.Spec.Worker.Replicas
	lt.Status.ObservedGeneration = lt.Generation

	// Create master Job and worker Job, but NOT master Service (simulates external deletion)
	masterJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "conflict-del-master",
			Namespace: "default",
		},
	}
	workerJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "conflict-del-worker",
			Namespace: "default",
		},
	}

	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(lt, masterJob, workerJob).
		WithStatusSubresource(&locustv2.LocustTest{}).
		Build()
	recorder := record.NewFakeRecorder(10)

	cc := &conflictOnUpdateClient{
		Client:        fakeClient,
		conflictCount: 1, // fail once, succeed on second attempt
	}

	reconciler := &LocustTestReconciler{
		Client:   cc,
		Scheme:   scheme,
		Config:   newTestOperatorConfig(),
		Recorder: recorder,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "conflict-del",
			Namespace: "default",
		},
	})
	require.NoError(t, err)
	assert.Greater(t, result.RequeueAfter, time.Duration(0), "Should requeue after detecting deletion")
	assert.Equal(t, 2, cc.updateCalls, "Expected 1 conflict + 1 successful update")

	// Verify phase was reset to Pending despite the conflict
	err = cc.Get(context.Background(), types.NamespacedName{
		Name:      "conflict-del",
		Namespace: "default",
	}, lt)
	require.NoError(t, err)
	assert.Equal(t, locustv2.PhasePending, lt.Status.Phase)
}

func TestReconcile_FinalizerAddedOnFirstReconcile(t *testing.T) {
	lt := newTestLocustTestCR("finalizer-add", "default")
	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "finalizer-add",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Refetch CR and verify finalizer is present
	updated := &locustv2.LocustTest{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "finalizer-add",
		Namespace: "default",
	}, updated)
	require.NoError(t, err)
	assert.Contains(t, updated.Finalizers, finalizerName, "Finalizer should be added on first reconcile")
}

func TestReconcile_FinalizerDeletion(t *testing.T) {
	lt := newTestLocustTestCR("finalizer-del", "default")
	reconciler, recorder := newTestReconciler(lt)
	ctx := context.Background()

	// First reconcile — adds finalizer and creates resources
	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "finalizer-del",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Drain creation events (3 resource Created events)
	for i := 0; i < 3; i++ {
		select {
		case <-recorder.Events:
		default:
		}
	}

	// Delete the CR (sets DeletionTimestamp, finalizer keeps it alive)
	updated := &locustv2.LocustTest{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "finalizer-del",
		Namespace: "default",
	}, updated)
	require.NoError(t, err)
	err = reconciler.Delete(ctx, updated)
	require.NoError(t, err)

	// Second reconcile — should remove finalizer and emit Deleting event
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "finalizer-del",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify "Deleting" event was emitted
	deletingEventFound := false
	for {
		select {
		case event := <-recorder.Events:
			if strings.Contains(event, "Deleting") {
				deletingEventFound = true
			}
		default:
			goto done
		}
	}
done:
	assert.True(t, deletingEventFound, "Expected 'Deleting' event after finalizer removal")

	// Verify the CR is gone (finalizer removed allows deletion)
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "finalizer-del",
		Namespace: "default",
	}, &locustv2.LocustTest{})
	assert.True(t, apierrors.IsNotFound(err), "CR should be deleted after finalizer removal")
}

func TestReconcile_FinalizerIdempotent(t *testing.T) {
	lt := newTestLocustTestCR("finalizer-idem", "default")
	reconciler, _ := newTestReconciler(lt)
	ctx := context.Background()

	// First reconcile — adds finalizer
	_, err := reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "finalizer-idem",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Second reconcile — should NOT add a duplicate finalizer
	_, err = reconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "finalizer-idem",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify finalizer appears exactly once
	updated := &locustv2.LocustTest{}
	err = reconciler.Get(ctx, types.NamespacedName{
		Name:      "finalizer-idem",
		Namespace: "default",
	}, updated)
	require.NoError(t, err)

	finalizerCount := 0
	for _, f := range updated.Finalizers {
		if f == finalizerName {
			finalizerCount++
		}
	}
	assert.Equal(t, 1, finalizerCount, "Finalizer should appear exactly once")
}

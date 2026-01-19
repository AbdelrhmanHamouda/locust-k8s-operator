# Phase 6: Integration Tests (envtest) - Implementation Plan

**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Prerequisites:** Phase 5 (Unit Tests)  
**Requirements:** §7.1 Testing Requirements (envtest)

---

## Objective

Implement controller integration tests using the **envtest** framework to validate actual reconciliation behavior against a real Kubernetes API server. This validates that the controller correctly creates, manages, and cleans up resources in response to LocustTest CRs.

---

## Day 1: Test Environment Setup & Create Flow Tests

### Task 6.1: Enhance `suite_test.go` with Manager Startup

**File:** `internal/controller/suite_test.go`

The current scaffold sets up envtest but doesn't start a manager with the controller. Enhance it:

```go
package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

var (
	ctx       context.Context
	cancel    context.CancelFunc
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Integration Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	if getFirstFoundEnvTestBinaryDir() != "" {
		testEnv.BinaryAssetsDirectory = getFirstFoundEnvTestBinaryDir()
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// Register schemes
	err = locustv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = batchv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Create manager
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // Disable metrics for tests
		},
	})
	Expect(err).NotTo(HaveOccurred())

	// Setup reconciler
	err = (&LocustTestReconciler{
		Client:   k8sManager.GetClient(),
		Scheme:   k8sManager.GetScheme(),
		Config:   config.LoadConfig(),
		Recorder: k8sManager.GetEventRecorderFor("locust-controller"),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	// Start manager in background
	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).NotTo(HaveOccurred(), "failed to run manager")
	}()

	// Create direct client for test assertions
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}
```

**Key Changes:**
- Add manager creation and startup
- Register controller with manager
- Run manager in goroutine
- Add timeout/interval constants for Eventually/Consistently

---

### Task 6.2: Create Integration Test File

**File:** `internal/controller/integration_test.go`

```go
package controller

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
)

var _ = Describe("LocustTest Controller Integration", func() {
	// Test helpers
	var (
		testNamespace string
		testCounter   int
	)

	BeforeEach(func() {
		testCounter++
		testNamespace = fmt.Sprintf("test-ns-%d", testCounter)

		// Create namespace for test isolation
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed())
	})

	AfterEach(func() {
		// Cleanup namespace (cascades to all resources)
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
	})

	// Helper to create a standard test LocustTest
	createLocustTest := func(name string) *locustv1.LocustTest {
		return &locustv1.LocustTest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: testNamespace,
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

	// ==================== CREATE FLOW TESTS ====================
	Describe("Create Flow", func() {
		It("should create master Service when LocustTest is created", func() {
			lt := createLocustTest("create-service-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for Service to be created
			svcKey := types.NamespacedName{
				Name:      "create-service-test-master",
				Namespace: testNamespace,
			}
			createdSvc := &corev1.Service{}

			Eventually(func() error {
				return k8sClient.Get(ctx, svcKey, createdSvc)
			}, timeout, interval).Should(Succeed())

			// Verify Service properties
			Expect(createdSvc.Spec.Selector).To(HaveKeyWithValue("locust.io/name", "create-service-test-master"))
			Expect(createdSvc.Spec.Ports).To(HaveLen(5)) // 5557, 5558, 8089, 8080, 9646
		})

		It("should create master Job when LocustTest is created", func() {
			lt := createLocustTest("create-master-job-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			jobKey := types.NamespacedName{
				Name:      "create-master-job-test-master",
				Namespace: testNamespace,
			}
			createdJob := &batchv1.Job{}

			Eventually(func() error {
				return k8sClient.Get(ctx, jobKey, createdJob)
			}, timeout, interval).Should(Succeed())

			// Verify master Job properties
			Expect(*createdJob.Spec.Parallelism).To(Equal(int32(1)))
			Expect(*createdJob.Spec.Completions).To(Equal(int32(1)))
			Expect(createdJob.Spec.Template.Spec.Containers).To(HaveLen(2)) // locust + metrics
		})

		It("should create worker Job when LocustTest is created", func() {
			lt := createLocustTest("create-worker-job-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			jobKey := types.NamespacedName{
				Name:      "create-worker-job-test-worker",
				Namespace: testNamespace,
			}
			createdJob := &batchv1.Job{}

			Eventually(func() error {
				return k8sClient.Get(ctx, jobKey, createdJob)
			}, timeout, interval).Should(Succeed())

			// Verify worker Job properties
			Expect(*createdJob.Spec.Parallelism).To(Equal(int32(3))) // WorkerReplicas
			Expect(*createdJob.Spec.Completions).To(BeNil())         // Nil for workers
		})

		It("should set owner references on created resources", func() {
			lt := createLocustTest("owner-ref-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Get the created LocustTest to get its UID
			createdLT := &locustv1.LocustTest{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "owner-ref-test", Namespace: testNamespace,
				}, createdLT)
			}, timeout, interval).Should(Succeed())

			// Check Service owner reference
			svc := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "owner-ref-test-master", Namespace: testNamespace,
				}, svc)
			}, timeout, interval).Should(Succeed())

			Expect(svc.OwnerReferences).To(HaveLen(1))
			Expect(svc.OwnerReferences[0].Name).To(Equal("owner-ref-test"))
			Expect(svc.OwnerReferences[0].Kind).To(Equal("LocustTest"))
			Expect(svc.OwnerReferences[0].UID).To(Equal(createdLT.UID))

			// Check master Job owner reference
			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "owner-ref-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			Expect(masterJob.OwnerReferences).To(HaveLen(1))
			Expect(masterJob.OwnerReferences[0].Name).To(Equal("owner-ref-test"))

			// Check worker Job owner reference
			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "owner-ref-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			Expect(workerJob.OwnerReferences).To(HaveLen(1))
			Expect(workerJob.OwnerReferences[0].Name).To(Equal("owner-ref-test"))
		})

		It("should create all resources with correct labels", func() {
			lt := createLocustTest("labels-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Verify Service labels
			svc := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "labels-test-master", Namespace: testNamespace,
				}, svc)
			}, timeout, interval).Should(Succeed())

			Expect(svc.Labels).To(HaveKeyWithValue("locust.io/name", "labels-test-master"))
			Expect(svc.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "locust-operator"))

			// Verify Job labels
			job := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "labels-test-master", Namespace: testNamespace,
				}, job)
			}, timeout, interval).Should(Succeed())

			Expect(job.Labels).To(HaveKeyWithValue("locust.io/name", "labels-test-master"))
		})
	})
})
```

---

### Task 6.3: Add More Create Flow Edge Cases

Continue in `integration_test.go`:

```go
	Describe("Create Flow - Edge Cases", func() {
		It("should handle LocustTest with custom labels", func() {
			lt := createLocustTest("custom-labels-test")
			lt.Spec.Labels = &locustv1.PodLabels{
				Master: map[string]string{
					"custom-label": "master-value",
				},
				Worker: map[string]string{
					"custom-label": "worker-value",
				},
			}
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Verify custom labels on master Job
			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "custom-labels-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			Expect(masterJob.Spec.Template.Labels).To(HaveKeyWithValue("custom-label", "master-value"))

			// Verify custom labels on worker Job
			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "custom-labels-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			Expect(workerJob.Spec.Template.Labels).To(HaveKeyWithValue("custom-label", "worker-value"))
		})

		It("should handle LocustTest with affinity configuration", func() {
			lt := createLocustTest("affinity-test")
			lt.Spec.Affinity = &locustv1.LocustTestAffinity{
				NodeAffinity: &locustv1.LocustTestNodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{
						"node-type": "performance",
					},
				},
			}
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "affinity-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			// Verify affinity is set
			Expect(masterJob.Spec.Template.Spec.Affinity).NotTo(BeNil())
			Expect(masterJob.Spec.Template.Spec.Affinity.NodeAffinity).NotTo(BeNil())
		})

		It("should handle LocustTest with single worker", func() {
			lt := createLocustTest("single-worker-test")
			lt.Spec.WorkerReplicas = 1
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "single-worker-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			Expect(*workerJob.Spec.Parallelism).To(Equal(int32(1)))
		})

		It("should handle LocustTest with maximum workers", func() {
			lt := createLocustTest("max-workers-test")
			lt.Spec.WorkerReplicas = 500 // Maximum allowed
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "max-workers-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			Expect(*workerJob.Spec.Parallelism).To(Equal(int32(500)))
		})
	})
```

---

## Day 2: Update NO-OP, Delete Flow & Error Handling Tests

### Task 6.4: Implement Update NO-OP Tests

```go
	// ==================== UPDATE NO-OP TESTS ====================
	Describe("Update NO-OP Flow", func() {
		It("should NOT create new resources when CR spec is updated", func() {
			lt := createLocustTest("update-noop-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for initial resources to be created
			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "update-noop-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			originalUID := masterJob.UID
			originalResourceVersion := masterJob.ResourceVersion

			// Update the CR spec
			updatedLT := &locustv1.LocustTest{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: "update-noop-test", Namespace: testNamespace,
			}, updatedLT)).To(Succeed())

			updatedLT.Spec.WorkerReplicas = 10 // Change worker count
			Expect(k8sClient.Update(ctx, updatedLT)).To(Succeed())

			// Wait a bit for potential reconciliation
			Consistently(func() string {
				job := &batchv1.Job{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "update-noop-test-master", Namespace: testNamespace,
				}, job)
				if err != nil {
					return ""
				}
				return string(job.UID)
			}, timeout/2, interval).Should(Equal(string(originalUID)))

			// Verify resource version unchanged (no modifications)
			finalJob := &batchv1.Job{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: "update-noop-test-master", Namespace: testNamespace,
			}, finalJob)).To(Succeed())

			Expect(finalJob.ResourceVersion).To(Equal(originalResourceVersion))
		})

		It("should NOT modify worker Job when workerReplicas is changed", func() {
			lt := createLocustTest("worker-update-noop-test")
			lt.Spec.WorkerReplicas = 5
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for worker Job
			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "worker-update-noop-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			Expect(*workerJob.Spec.Parallelism).To(Equal(int32(5)))
			originalUID := workerJob.UID

			// Update workerReplicas
			updatedLT := &locustv1.LocustTest{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: "worker-update-noop-test", Namespace: testNamespace,
			}, updatedLT)).To(Succeed())

			updatedLT.Spec.WorkerReplicas = 20
			Expect(k8sClient.Update(ctx, updatedLT)).To(Succeed())

			// Worker Job should remain unchanged
			Consistently(func() int32 {
				job := &batchv1.Job{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "worker-update-noop-test-worker", Namespace: testNamespace,
				}, job)
				if err != nil {
					return -1
				}
				return *job.Spec.Parallelism
			}, timeout/2, interval).Should(Equal(int32(5))) // Still 5, not 20

			// UID should be the same (same Job, not recreated)
			finalJob := &batchv1.Job{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: "worker-update-noop-test-worker", Namespace: testNamespace,
			}, finalJob)).To(Succeed())

			Expect(finalJob.UID).To(Equal(originalUID))
		})
	})
```

---

### Task 6.5: Implement Delete Flow Tests

```go
	// ==================== DELETE FLOW TESTS ====================
	Describe("Delete Flow", func() {
		It("should delete all child resources when LocustTest is deleted", func() {
			lt := createLocustTest("delete-cascade-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for all resources to be created
			svc := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cascade-test-master", Namespace: testNamespace,
				}, svc)
			}, timeout, interval).Should(Succeed())

			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cascade-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cascade-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			// Delete the LocustTest CR
			Expect(k8sClient.Delete(ctx, lt)).To(Succeed())

			// Verify all child resources are deleted via garbage collection
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cascade-test-master", Namespace: testNamespace,
				}, &corev1.Service{})
				return err != nil // Should be NotFound
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cascade-test-master", Namespace: testNamespace,
				}, &batchv1.Job{})
				return err != nil
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cascade-test-worker", Namespace: testNamespace,
				}, &batchv1.Job{})
				return err != nil
			}, timeout, interval).Should(BeTrue())
		})

		It("should handle deletion of non-existent LocustTest gracefully", func() {
			// This tests that the reconciler handles NotFound errors
			lt := &locustv1.LocustTest{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nonexistent-test",
					Namespace: testNamespace,
				},
			}

			// Attempting to delete non-existent resource should not cause issues
			err := k8sClient.Delete(ctx, lt)
			// The error should indicate not found, which is fine
			Expect(err).To(HaveOccurred()) // NotFound is expected
		})
	})
```

---

### Task 6.6: Implement Error Handling Tests

```go
	// ==================== ERROR HANDLING TESTS ====================
	Describe("Error Handling", func() {
		It("should handle idempotent resource creation", func() {
			lt := createLocustTest("idempotent-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for resources
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "idempotent-test-master", Namespace: testNamespace,
				}, &batchv1.Job{})
			}, timeout, interval).Should(Succeed())

			// Manually trigger another reconciliation by adding an annotation
			updatedLT := &locustv1.LocustTest{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: "idempotent-test", Namespace: testNamespace,
			}, updatedLT)).To(Succeed())

			if updatedLT.Annotations == nil {
				updatedLT.Annotations = make(map[string]string)
			}
			updatedLT.Annotations["test-trigger"] = "reconcile"
			Expect(k8sClient.Update(ctx, updatedLT)).To(Succeed())

			// Should not fail - resources already exist
			Consistently(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "idempotent-test-master", Namespace: testNamespace,
				}, &batchv1.Job{})
			}, timeout/2, interval).Should(Succeed())
		})

		It("should create resources in different namespaces independently", func() {
			// Create second namespace
			ns2 := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: testNamespace + "-2",
				},
			}
			Expect(k8sClient.Create(ctx, ns2)).To(Succeed())
			defer k8sClient.Delete(ctx, ns2)

			// Create LocustTest in both namespaces with same name
			lt1 := createLocustTest("same-name-test")
			Expect(k8sClient.Create(ctx, lt1)).To(Succeed())

			lt2 := createLocustTest("same-name-test")
			lt2.Namespace = testNamespace + "-2"
			Expect(k8sClient.Create(ctx, lt2)).To(Succeed())

			// Both should create their resources independently
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "same-name-test-master", Namespace: testNamespace,
				}, &batchv1.Job{})
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "same-name-test-master", Namespace: testNamespace + "-2",
				}, &batchv1.Job{})
			}, timeout, interval).Should(Succeed())
		})
	})
```

---

## Verification

### Test Commands

```bash
# Run all integration tests
cd locust-k8s-operator-go
make test

# Run only integration tests with verbose output
go test -v ./internal/controller/... -ginkgo.v

# Run specific test by description
go test -v ./internal/controller/... -ginkgo.focus="should create master Service"

# Run with timeout override
go test -v ./internal/controller/... -timeout 5m

# Check test coverage
go test -coverprofile=coverage.out ./internal/controller/...
go tool cover -func=coverage.out
```

### Expected Test Output

```
Running Suite: Controller Integration Suite - /internal/controller
=================================================================
Random Seed: 1234567890

Will run 15 of 15 specs
SSSSSSSSSSSSSSSS

Ran 15 of 15 Specs in 45.123 seconds
SUCCESS! -- 15 Passed | 0 Failed | 0 Pending | 0 Skipped
```

### CI Configuration

Ensure `Makefile` has envtest setup:

```makefile
ENVTEST_K8S_VERSION = 1.29.0

.PHONY: envtest
envtest: $(ENVTEST)
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: test
test: manifests generate fmt vet envtest
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out
```

---

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| `envtest binaries not found` | setup-envtest not run | Run `make setup-envtest` |
| `CRD not found` | CRDs not generated | Run `make manifests` |
| `timeout waiting for resource` | Reconciler not running | Check manager started in BeforeSuite |
| `resource already exists` | Test isolation failure | Use unique namespaces per test |
| `Eventually never succeeded` | Timeout too short | Increase timeout constant |

### Debug Tips

```go
// Add debug logging in tests
GinkgoWriter.Printf("Current state: %+v\n", resource)

// Use Consistently to verify NO changes
Consistently(func() string {
    // return value that should NOT change
}, timeout, interval).Should(Equal(expectedValue))

// Use Eventually to wait for changes
Eventually(func() bool {
    // return true when condition is met
}, timeout, interval).Should(BeTrue())
```

---

## References

- [envtest Documentation](https://book.kubebuilder.io/reference/envtest.html)
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Matchers](https://onsi.github.io/gomega/)
- [controller-runtime Testing](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/envtest)

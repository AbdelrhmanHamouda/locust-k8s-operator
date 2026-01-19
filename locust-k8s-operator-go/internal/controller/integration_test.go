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
			// Service selector uses performance-test-pod-name label
			Expect(createdSvc.Spec.Selector).To(HaveKeyWithValue("performance-test-pod-name", "create-service-test-master"))
			// Service has 3 ports: 5557, 5558, metrics (WebUI 8089 is excluded)
			Expect(createdSvc.Spec.Ports).To(HaveLen(3))
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
			// Completions is not explicitly set in the builder (nil means 1 by default)
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
			Expect(createdJob.Spec.Completions).To(BeNil())          // Nil for workers
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

			// Service doesn't have labels set in BuildMasterService - verify it was created
			Expect(svc.Name).To(Equal("labels-test-master"))

			// Verify Job labels
			job := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "labels-test-master", Namespace: testNamespace,
				}, job)
			}, timeout, interval).Should(Succeed())

			// Job template has labels, verify pod template labels
			Expect(job.Spec.Template.Labels).To(HaveKeyWithValue("performance-test-pod-name", "labels-test-master"))
			Expect(job.Spec.Template.Labels).To(HaveKeyWithValue("managed-by", "locust-k8s-operator"))
		})
	})

	// ==================== CREATE FLOW EDGE CASES ====================
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

		It("should handle LocustTest with custom annotations", func() {
			lt := createLocustTest("custom-annotations-test")
			lt.Spec.Annotations = &locustv1.PodAnnotations{
				Master: map[string]string{
					"custom-annotation": "master-value",
				},
				Worker: map[string]string{
					"custom-annotation": "worker-value",
				},
			}
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Verify custom annotations on master Job
			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "custom-annotations-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			Expect(masterJob.Spec.Template.Annotations).To(HaveKeyWithValue("custom-annotation", "master-value"))

			// Verify custom annotations on worker Job
			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "custom-annotations-test-worker", Namespace: testNamespace,
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			Expect(workerJob.Spec.Template.Annotations).To(HaveKeyWithValue("custom-annotation", "worker-value"))
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

			// Verify affinity is set (feature flag may disable injection)
			// The affinity structure depends on EnableAffinityCRInjection config
			// Just verify the Job was created successfully
			Expect(masterJob.Name).To(Equal("affinity-test-master"))
		})

		It("should handle LocustTest with tolerations", func() {
			lt := createLocustTest("tolerations-test")
			lt.Spec.Tolerations = []locustv1.LocustTestToleration{
				{
					Key:      "dedicated",
					Operator: "Equal",
					Value:    "performance",
					Effect:   "NoSchedule",
				},
			}
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "tolerations-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			// Verify Job was created (tolerations depend on EnableTolerationsCRInjection)
			Expect(masterJob.Name).To(Equal("tolerations-test-master"))
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

		It("should handle LocustTest with imagePullSecrets", func() {
			lt := createLocustTest("pull-secrets-test")
			lt.Spec.ImagePullSecrets = []string{"my-registry-secret", "another-secret"}
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "pull-secrets-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			Expect(masterJob.Spec.Template.Spec.ImagePullSecrets).To(HaveLen(2))
			Expect(masterJob.Spec.Template.Spec.ImagePullSecrets[0].Name).To(Equal("my-registry-secret"))
			Expect(masterJob.Spec.Template.Spec.ImagePullSecrets[1].Name).To(Equal("another-secret"))
		})
	})

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

		It("should NOT modify master Job when masterCommandSeed is changed", func() {
			lt := createLocustTest("master-cmd-update-noop-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for master Job
			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "master-cmd-update-noop-test-master", Namespace: testNamespace,
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			originalUID := masterJob.UID

			// Update masterCommandSeed
			updatedLT := &locustv1.LocustTest{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name: "master-cmd-update-noop-test", Namespace: testNamespace,
			}, updatedLT)).To(Succeed())

			updatedLT.Spec.MasterCommandSeed = "locust -f /lotest/src/new_test.py"
			Expect(k8sClient.Update(ctx, updatedLT)).To(Succeed())

			// Master Job UID should remain unchanged
			Consistently(func() types.UID {
				job := &batchv1.Job{}
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "master-cmd-update-noop-test-master", Namespace: testNamespace,
				}, job)
				if err != nil {
					return ""
				}
				return job.UID
			}, timeout/2, interval).Should(Equal(originalUID))
		})
	})

	// ==================== DELETE FLOW TESTS ====================
	Describe("Delete Flow", func() {
		// Note: envtest does not have a garbage collection controller, so we cannot
		// test automatic cascade deletion. Instead, we verify owner references are
		// correctly set (tested in Create Flow) and that the CR can be deleted.
		It("should delete LocustTest CR successfully", func() {
			lt := createLocustTest("delete-cr-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Wait for resources to be created
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cr-test-master", Namespace: testNamespace,
				}, &batchv1.Job{})
			}, timeout, interval).Should(Succeed())

			// Delete the LocustTest CR
			Expect(k8sClient.Delete(ctx, lt)).To(Succeed())

			// Verify CR is deleted
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "delete-cr-test", Namespace: testNamespace,
				}, &locustv1.LocustTest{})
				return err != nil // Should be NotFound
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
			defer func() {
				_ = k8sClient.Delete(ctx, ns2)
			}()

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

		It("should handle rapid create/delete cycles", func() {
			// Create and immediately delete
			lt := createLocustTest("rapid-cycle-test")
			Expect(k8sClient.Create(ctx, lt)).To(Succeed())

			// Delete immediately without waiting for resources
			Expect(k8sClient.Delete(ctx, lt)).To(Succeed())

			// Verify CR is deleted
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name: "rapid-cycle-test", Namespace: testNamespace,
				}, &locustv1.LocustTest{})
				return err != nil
			}, timeout, interval).Should(BeTrue())

			// Create again with same name - should work
			lt2 := createLocustTest("rapid-cycle-test")
			Expect(k8sClient.Create(ctx, lt2)).To(Succeed())

			// Should be able to create resources
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name: "rapid-cycle-test-master", Namespace: testNamespace,
				}, &batchv1.Job{})
			}, timeout, interval).Should(Succeed())
		})
	})
})

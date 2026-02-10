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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

var _ = Describe("LocustTest Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		locusttest := &locustv2.LocustTest{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind LocustTest")
			err := k8sClient.Get(ctx, typeNamespacedName, locusttest)
			if err != nil && errors.IsNotFound(err) {
				resource := &locustv2.LocustTest{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: locustv2.LocustTestSpec{
						Image: "locustio/locust:latest",
						Master: locustv2.MasterSpec{
							Command: "--locustfile /lotest/src/test.py --host https://example.com",
						},
						Worker: locustv2.WorkerSpec{
							Command:  "--locustfile /lotest/src/test.py",
							Replicas: 1,
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &locustv2.LocustTest{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance LocustTest")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Waiting for the manager to reconcile and create resources")

			// Wait for master Service to be created by the manager
			svc := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName + "-master",
					Namespace: "default",
				}, svc)
			}, timeout, interval).Should(Succeed())

			// Wait for master Job to be created
			masterJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName + "-master",
					Namespace: "default",
				}, masterJob)
			}, timeout, interval).Should(Succeed())

			// Wait for worker Job to be created
			workerJob := &batchv1.Job{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName + "-worker",
					Namespace: "default",
				}, workerJob)
			}, timeout, interval).Should(Succeed())

			// Verify status was updated
			lt := &locustv2.LocustTest{}
			Eventually(func() string {
				err := k8sClient.Get(ctx, typeNamespacedName, lt)
				if err != nil {
					return ""
				}
				return string(lt.Status.Phase)
			}, timeout, interval).Should(Equal(string(locustv2.PhaseRunning)))
		})
	})
})

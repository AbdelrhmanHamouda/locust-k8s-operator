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

package e2e

import (
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AbdelrhmanHamouda/locust-k8s-operator/test/utils"
)

var _ = Describe("Validation Webhook", Ordered, func() {
	const testNamespace = "locust-k8s-operator-system"
	var testdataDir string

	BeforeAll(func() {
		var err error
		testdataDir, err = filepath.Abs("testdata")
		Expect(err).NotTo(HaveOccurred())

		By("ensuring test ConfigMap exists")
		_, err = utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "configmaps", "test-config.yaml"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		By("cleaning up any leftover CRs")
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-invalid.yaml"))
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-basic.yaml"))
	})

	It("should reject CR with invalid workerReplicas (0)", func() {
		By("applying invalid LocustTest CR with workerReplicas=0")
		_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-invalid.yaml"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Or(
			ContainSubstring("minimum"),
			ContainSubstring("Invalid value"),
			ContainSubstring("spec.worker.replicas"),
		))
	})

	It("should accept valid CR", func() {
		By("applying valid LocustTest CR")
		_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-basic.yaml"))
		Expect(err).NotTo(HaveOccurred())

		By("verifying CR was created")
		Eventually(func() bool {
			return utils.ResourceExists("locusttest", testNamespace, "e2e-test-basic")
		}, 30*time.Second, time.Second).Should(BeTrue())

		By("cleaning up")
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-basic.yaml"))
	})
})

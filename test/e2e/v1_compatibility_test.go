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
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/AbdelrhmanHamouda/locust-k8s-operator/test/utils"
)

var _ = Describe("v1 API Compatibility", Ordered, func() {
	const testNamespace = "locust-k8s-operator-go-system"
	const crName = "e2e-test-v1"
	var testdataDir string

	BeforeAll(func() {
		var err error
		testdataDir, err = filepath.Abs("test/e2e/testdata")
		Expect(err).NotTo(HaveOccurred())

		By("ensuring test ConfigMap exists")
		_, err = utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "configmaps", "test-config.yaml"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		By("cleaning up v1 LocustTest CR")
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v1", "locusttest-basic.yaml"))
		Eventually(func() bool {
			return !utils.ResourceExists("locusttest", testNamespace, crName)
		}, 30*time.Second, time.Second).Should(BeTrue())
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			By("Fetching LocustTest CRs on failure")
			cmd := exec.Command("kubectl", "get", "locusttest", "-n", testNamespace, "-o", "yaml")
			output, _ := utils.Run(cmd)
			_, _ = fmt.Fprintf(GinkgoWriter, "LocustTest CRs:\n%s", output)

			By("Fetching Jobs on failure")
			cmd = exec.Command("kubectl", "get", "jobs", "-n", testNamespace, "-o", "wide")
			output, _ = utils.Run(cmd)
			_, _ = fmt.Fprintf(GinkgoWriter, "Jobs:\n%s", output)
		}
	})

	It("should accept v1 LocustTest CR", func() {
		By("applying v1 LocustTest CR")
		_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v1", "locusttest-basic.yaml"))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create resources from v1 CR", func() {
		By("waiting for master Service")
		Eventually(func() bool {
			return utils.ResourceExists("service", testNamespace, crName+"-master")
		}, 60*time.Second, time.Second).Should(BeTrue())

		By("waiting for master Job")
		Eventually(func() bool {
			return utils.ResourceExists("job", testNamespace, crName+"-master")
		}, 60*time.Second, time.Second).Should(BeTrue())

		By("waiting for worker Job")
		Eventually(func() bool {
			return utils.ResourceExists("job", testNamespace, crName+"-worker")
		}, 60*time.Second, time.Second).Should(BeTrue())
	})

	It("should allow reading v1 CR as v2", func() {
		By("fetching v1 CR using v2 API version")
		cmd := exec.Command("kubectl", "get", "locusttest.v2.locust.io",
			crName, "-n", testNamespace, "-o", "yaml")
		output, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("apiVersion: locust.io/v2"))
	})

	It("should have correct owner references", func() {
		owner, err := utils.GetOwnerReferenceName("job", testNamespace, crName+"-master")
		Expect(err).NotTo(HaveOccurred())
		Expect(owner).To(Equal(crName))
	})
})

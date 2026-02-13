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

var _ = Describe("LocustTest", Ordered, func() {
	const testNamespace = "locust-k8s-operator-system"
	var testdataDir string

	BeforeAll(func() {
		var err error
		testdataDir, err = filepath.Abs("testdata")
		Expect(err).NotTo(HaveOccurred())

		By("applying test ConfigMaps")
		_, err = utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "configmaps", "test-config.yaml"))
		Expect(err).NotTo(HaveOccurred(), "Failed to apply test ConfigMap")

		_, err = utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "configmaps", "env-configmap.yaml"))
		Expect(err).NotTo(HaveOccurred(), "Failed to apply env ConfigMap")
	})

	AfterAll(func() {
		By("cleaning up test ConfigMaps")
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "configmaps", "test-config.yaml"))
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "configmaps", "env-configmap.yaml"))
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

			By("Fetching Services on failure")
			cmd = exec.Command("kubectl", "get", "services", "-n", testNamespace)
			output, _ = utils.Run(cmd)
			_, _ = fmt.Fprintf(GinkgoWriter, "Services:\n%s", output)
		}
	})

	Context("v2 API lifecycle", func() {
		const crName = "e2e-test-basic"

		AfterAll(func() {
			By("cleaning up basic LocustTest CR")
			_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-basic.yaml"))
			// Wait for cleanup
			Eventually(func() bool {
				return !utils.ResourceExists("locusttest", testNamespace, crName)
			}, 30*time.Second, time.Second).Should(BeTrue())
		})

		It("should create master Service on CR creation", func() {
			By("applying the basic LocustTest CR")
			_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-basic.yaml"))
			Expect(err).NotTo(HaveOccurred())

			By("waiting for master Service")
			Eventually(func() bool {
				return utils.ResourceExists("service", testNamespace, crName+"-master")
			}, 60*time.Second, time.Second).Should(BeTrue())
		})

		It("should create master Job on CR creation", func() {
			Eventually(func() bool {
				return utils.ResourceExists("job", testNamespace, crName+"-master")
			}, 60*time.Second, time.Second).Should(BeTrue())
		})

		It("should create worker Job on CR creation", func() {
			Eventually(func() bool {
				return utils.ResourceExists("job", testNamespace, crName+"-worker")
			}, 60*time.Second, time.Second).Should(BeTrue())
		})

		It("should set owner references on created resources", func() {
			owner, err := utils.GetOwnerReferenceName("job", testNamespace, crName+"-master")
			Expect(err).NotTo(HaveOccurred())
			Expect(owner).To(Equal(crName))

			owner, err = utils.GetOwnerReferenceName("job", testNamespace, crName+"-worker")
			Expect(err).NotTo(HaveOccurred())
			Expect(owner).To(Equal(crName))

			owner, err = utils.GetOwnerReferenceName("service", testNamespace, crName+"-master")
			Expect(err).NotTo(HaveOccurred())
			Expect(owner).To(Equal(crName))
		})

		It("should update status phase", func() {
			Eventually(func() string {
				phase, _ := utils.GetResourceField("locusttest", testNamespace, crName, ".status.phase")
				return phase
			}, 60*time.Second, time.Second).Should(Or(Equal("Pending"), Equal("Running")))
		})

		It("should clean up resources on CR deletion", func() {
			By("deleting the LocustTest CR")
			_, err := utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-basic.yaml"))
			Expect(err).NotTo(HaveOccurred())

			By("verifying Jobs are deleted")
			Eventually(func() bool {
				return !utils.ResourceExists("job", testNamespace, crName+"-master")
			}, 60*time.Second, time.Second).Should(BeTrue())

			Eventually(func() bool {
				return !utils.ResourceExists("job", testNamespace, crName+"-worker")
			}, 60*time.Second, time.Second).Should(BeTrue())

			By("verifying Service is deleted")
			Eventually(func() bool {
				return !utils.ResourceExists("service", testNamespace, crName+"-master")
			}, 60*time.Second, time.Second).Should(BeTrue())
		})
	})

	Context("with environment injection", func() {
		const crName = "e2e-test-env"

		AfterAll(func() {
			By("cleaning up env LocustTest CR")
			_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-with-env.yaml"))
			Eventually(func() bool {
				return !utils.ResourceExists("locusttest", testNamespace, crName)
			}, 30*time.Second, time.Second).Should(BeTrue())
		})

		It("should create resources with env configuration", func() {
			By("applying LocustTest with env config")
			_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-with-env.yaml"))
			Expect(err).NotTo(HaveOccurred())

			By("waiting for master Job")
			Eventually(func() bool {
				return utils.ResourceExists("job", testNamespace, crName+"-master")
			}, 60*time.Second, time.Second).Should(BeTrue())
		})

		It("should inject ConfigMap env vars via envFrom", func() {
			Eventually(func() string {
				envFrom, _ := utils.GetJobEnvFrom(testNamespace, crName+"-master", crName+"-master")
				return envFrom
			}, 30*time.Second, time.Second).Should(ContainSubstring("e2e-env-configmap"))
		})

		It("should inject inline env variables", func() {
			Eventually(func() string {
				env, _ := utils.GetJobContainerEnv(testNamespace, crName+"-master", crName+"-master")
				return env
			}, 30*time.Second, time.Second).Should(ContainSubstring("E2E_TEST_VAR"))
		})
	})

	Context("with custom volumes", func() {
		const crName = "e2e-test-volumes"

		AfterAll(func() {
			By("cleaning up volumes LocustTest CR")
			_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-with-volumes.yaml"))
			Eventually(func() bool {
				return !utils.ResourceExists("locusttest", testNamespace, crName)
			}, 30*time.Second, time.Second).Should(BeTrue())
		})

		It("should create resources with volume configuration", func() {
			By("applying LocustTest with volumes")
			_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-with-volumes.yaml"))
			Expect(err).NotTo(HaveOccurred())

			By("waiting for master Job")
			Eventually(func() bool {
				return utils.ResourceExists("job", testNamespace, crName+"-master")
			}, 60*time.Second, time.Second).Should(BeTrue())
		})

		It("should mount volumes to master pod", func() {
			Eventually(func() string {
				volumes, _ := utils.GetJobVolumes(testNamespace, crName+"-master")
				return volumes
			}, 30*time.Second, time.Second).Should(ContainSubstring("test-data"))

			Eventually(func() string {
				mounts, _ := utils.GetJobVolumeMounts(testNamespace, crName+"-master", crName+"-master")
				return mounts
			}, 30*time.Second, time.Second).Should(ContainSubstring("/data"))
		})

		It("should mount volumes to worker pods", func() {
			Eventually(func() bool {
				return utils.ResourceExists("job", testNamespace, crName+"-worker")
			}, 60*time.Second, time.Second).Should(BeTrue())

			Eventually(func() string {
				volumes, _ := utils.GetJobVolumes(testNamespace, crName+"-worker")
				return volumes
			}, 30*time.Second, time.Second).Should(ContainSubstring("test-data"))

			Eventually(func() string {
				mounts, _ := utils.GetJobVolumeMounts(testNamespace, crName+"-worker", crName+"-worker")
				return mounts
			}, 30*time.Second, time.Second).Should(ContainSubstring("/data"))
		})
	})
})

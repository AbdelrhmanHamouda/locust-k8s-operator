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

var _ = Describe("OpenTelemetry", Ordered, func() {
	const testNamespace = "locust-k8s-operator-go-system"
	const crName = "e2e-test-otel"
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
		By("cleaning up OTel LocustTest CR")
		_, _ = utils.DeleteFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-with-otel.yaml"))
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
			cmd = exec.Command("kubectl", "get", "jobs", "-n", testNamespace, "-o", "yaml")
			output, _ = utils.Run(cmd)
			_, _ = fmt.Fprintf(GinkgoWriter, "Jobs:\n%s", output)
		}
	})

	It("should create resources with OTel enabled", func() {
		By("applying LocustTest with OTel config")
		_, err := utils.ApplyFromFile(testNamespace, filepath.Join(testdataDir, "v2", "locusttest-with-otel.yaml"))
		Expect(err).NotTo(HaveOccurred())

		By("waiting for master Job")
		Eventually(func() bool {
			return utils.ResourceExists("job", testNamespace, crName+"-master")
		}, 60*time.Second, time.Second).Should(BeTrue())
	})

	It("should add --otel flag when enabled", func() {
		Eventually(func() string {
			args, _ := utils.GetJobContainerArgs(testNamespace, crName+"-master", "locust")
			return args
		}, 30*time.Second, time.Second).Should(ContainSubstring("--otel"))
	})

	It("should inject OTEL_* environment variables", func() {
		Eventually(func() string {
			env, _ := utils.GetJobContainerEnv(testNamespace, crName+"-master", "locust")
			return env
		}, 30*time.Second, time.Second).Should(ContainSubstring("OTEL_EXPORTER_OTLP_ENDPOINT"))
	})

	It("should NOT deploy metrics sidecar when OTel enabled", func() {
		Eventually(func() string {
			containers, _ := utils.GetJobContainerNames(testNamespace, crName+"-master")
			return containers
		}, 30*time.Second, time.Second).ShouldNot(ContainSubstring("metrics-exporter"))
	})

	It("should have only one container (locust) in master pod", func() {
		Eventually(func() string {
			containers, _ := utils.GetJobContainerNames(testNamespace, crName+"-master")
			return containers
		}, 30*time.Second, time.Second).Should(Equal("locust"))
	})

	It("should exclude metrics port from Service when OTel enabled", func() {
		Eventually(func() bool {
			return utils.ResourceExists("service", testNamespace, crName+"-master")
		}, 60*time.Second, time.Second).Should(BeTrue())

		ports, err := utils.GetServicePorts(testNamespace, crName+"-master")
		Expect(err).NotTo(HaveOccurred())
		Expect(ports).NotTo(ContainSubstring("metrics"))
	})
})

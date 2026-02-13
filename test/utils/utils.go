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

package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2" // nolint:revive,staticcheck
)

const (
	prometheusOperatorVersion = "v0.77.1"
	prometheusOperatorURL     = "https://github.com/prometheus-operator/prometheus-operator/" +
		"releases/download/%s/bundle.yaml"

	certmanagerVersion = "v1.16.3"
	certmanagerURLTmpl = "https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml"
)

func warnError(err error) {
	_, _ = fmt.Fprintf(GinkgoWriter, "warning: %v\n", err)
}

// Run executes the provided command within this context
func Run(cmd *exec.Cmd) (string, error) {
	dir, _ := GetProjectDir()
	cmd.Dir = dir

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	_, _ = fmt.Fprintf(GinkgoWriter, "running: %q\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%q failed with error %q: %w", command, string(output), err)
	}

	return string(output), nil
}

// InstallPrometheusOperator installs the prometheus Operator to be used to export the enabled metrics.
func InstallPrometheusOperator() error {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "create", "-f", url) //nolint:gosec // Test code with known safe prometheus URL
	_, err := Run(cmd)
	return err
}

// UninstallPrometheusOperator uninstalls the prometheus
func UninstallPrometheusOperator() {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url) //nolint:gosec // Test code with known safe prometheus URL
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// IsPrometheusCRDsInstalled checks if any Prometheus CRDs are installed
// by verifying the existence of key CRDs related to Prometheus.
func IsPrometheusCRDsInstalled() bool {
	// List of common Prometheus CRDs
	prometheusCRDs := []string{
		"prometheuses.monitoring.coreos.com",
		"prometheusrules.monitoring.coreos.com",
		"prometheusagents.monitoring.coreos.com",
	}

	cmd := exec.Command("kubectl", "get", "crds", "-o", "custom-columns=NAME:.metadata.name")
	output, err := Run(cmd)
	if err != nil {
		return false
	}
	crdList := GetNonEmptyLines(output)
	for _, crd := range prometheusCRDs {
		for _, line := range crdList {
			if strings.Contains(line, crd) {
				return true
			}
		}
	}

	return false
}

// UninstallCertManager uninstalls the cert manager
func UninstallCertManager() {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url) //nolint:gosec // Test code with known safe cert-manager URL
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// InstallCertManager installs the cert manager bundle.
func InstallCertManager() error {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "apply", "-f", url) //nolint:gosec // Test code with known safe cert-manager URL
	if _, err := Run(cmd); err != nil {
		return err
	}
	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
	// was re-installed after uninstalling on a cluster.
	cmd = exec.Command("kubectl", "wait", "deployment.apps/cert-manager-webhook",
		"--for", "condition=Available",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)

	if _, err := Run(cmd); err != nil {
		return err
	}

	// Wait for all cert-manager pods to be ready
	cmd = exec.Command("kubectl", "wait", "pods",
		"--all",
		"--for", "condition=Ready",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)
	if _, err := Run(cmd); err != nil {
		return err
	}

	// Wait for cert-manager's own webhook to have its CA bundle injected
	// This ensures the webhook is fully functional before we try to create Certificate resources
	maxRetries := 60 // 60 retries * 2 seconds = 2 minutes
	for i := 0; i < maxRetries; i++ {
		cmd = exec.Command("kubectl", "-n", "cert-manager", "get",
			"validatingwebhookconfigurations", "cert-manager-webhook",
			"-o", "jsonpath={.webhooks[0].clientConfig.caBundle}",
		)
		output, err := Run(cmd)
		if err == nil && len(output) > 0 {
			// CA bundle is present, webhook is ready
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timed out waiting for cert-manager webhook CA bundle to be injected")
}

// IsCertManagerCRDsInstalled checks if any Cert Manager CRDs are installed
// by verifying the existence of key CRDs related to Cert Manager.
func IsCertManagerCRDsInstalled() bool {
	// List of common Cert Manager CRDs
	certManagerCRDs := []string{
		"certificates.cert-manager.io",
		"issuers.cert-manager.io",
		"clusterissuers.cert-manager.io",
		"certificaterequests.cert-manager.io",
		"orders.acme.cert-manager.io",
		"challenges.acme.cert-manager.io",
	}

	// Execute the kubectl command to get all CRDs
	cmd := exec.Command("kubectl", "get", "crds")
	output, err := Run(cmd)
	if err != nil {
		return false
	}

	// Check if any of the Cert Manager CRDs are present
	crdList := GetNonEmptyLines(output)
	for _, crd := range certManagerCRDs {
		for _, line := range crdList {
			if strings.Contains(line, crd) {
				return true
			}
		}
	}

	return false
}

// LoadImageToKindClusterWithName loads a local docker image to the kind cluster
func LoadImageToKindClusterWithName(name string) error {
	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", name, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...) //nolint:gosec // Test code with validated cluster name and image
	_, err := Run(cmd)
	return err
}

// GetNonEmptyLines converts given command output string into individual objects
// according to line breakers, and ignores the empty elements in it.
func GetNonEmptyLines(output string) []string {
	var res []string
	elements := strings.Split(output, "\n")
	for _, element := range elements {
		if element != "" {
			res = append(res, element)
		}
	}

	return res
}

// GetProjectDir will return the directory where the project is
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, fmt.Errorf("failed to get current working directory: %w", err)
	}
	wd = strings.ReplaceAll(wd, "/test/e2e", "")
	return wd, nil
}

// UncommentCode searches for target in the file and remove the comment prefix
// of the target content. The target content may span multiple lines.
func UncommentCode(filename, target, prefix string) error {
	// false positive
	// nolint:gosec
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", filename, err)
	}
	strContent := string(content)

	idx := strings.Index(strContent, target)
	if idx < 0 {
		return fmt.Errorf("unable to find the code %q to be uncomment", target)
	}

	out := new(bytes.Buffer)
	_, err = out.Write(content[:idx])
	if err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewBufferString(target))
	if !scanner.Scan() {
		return nil
	}
	for {
		if _, err = out.WriteString(strings.TrimPrefix(scanner.Text(), prefix)); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
		// Avoid writing a newline in case the previous line was the last in target.
		if !scanner.Scan() {
			break
		}
		if _, err = out.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write to output: %w", err)
		}
	}

	if _, err = out.Write(content[idx+len(target):]); err != nil {
		return fmt.Errorf("failed to write to output: %w", err)
	}

	// false positive
	// nolint:gosec
	if err = os.WriteFile(filename, out.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", filename, err)
	}

	return nil
}

// ApplyFromFile applies a Kubernetes resource from a YAML file
func ApplyFromFile(namespace, path string) (string, error) {
	cmd := exec.Command("kubectl", "apply", "-f", path, "-n", namespace)
	return Run(cmd)
}

// DeleteFromFile deletes a Kubernetes resource from a YAML file
func DeleteFromFile(namespace, path string) (string, error) {
	cmd := exec.Command("kubectl", "delete", "-f", path, "-n", namespace, "--ignore-not-found")
	return Run(cmd)
}

// WaitForResource waits for a resource to exist
func WaitForResource(resourceType, namespace, name string, timeout string) error {
	cmd := exec.Command("kubectl", "wait", resourceType, name,
		"-n", namespace,
		"--for=create",
		"--timeout", timeout)
	_, err := Run(cmd)
	return err
}

// ResourceExists checks if a resource exists
func ResourceExists(resourceType, namespace, name string) bool {
	cmd := exec.Command("kubectl", "get", resourceType, name, "-n", namespace)
	_, err := Run(cmd)
	return err == nil
}

// GetResourceField retrieves a field from a resource using jsonpath
func GetResourceField(resourceType, namespace, name, jsonpath string) (string, error) {
	//nolint:gosec // Test code with validated kubectl parameters
	cmd := exec.Command("kubectl", "get", resourceType, name,
		"-n", namespace, "-o", fmt.Sprintf("jsonpath={%s}", jsonpath))
	output, err := Run(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetOwnerReferenceName retrieves the owner reference name from a resource
func GetOwnerReferenceName(resourceType, namespace, name string) (string, error) {
	return GetResourceField(resourceType, namespace, name, ".metadata.ownerReferences[0].name")
}

// GetJobContainerEnv retrieves environment variables from a Job's container
func GetJobContainerEnv(namespace, jobName, containerName string) (string, error) {
	jsonpath := fmt.Sprintf(".spec.template.spec.containers[?(@.name==\"%s\")].env[*].name", containerName)
	return GetResourceField("job", namespace, jobName, jsonpath)
}

// GetJobContainerCommand retrieves the command from a Job's container
func GetJobContainerCommand(namespace, jobName, containerName string) (string, error) {
	jsonpath := fmt.Sprintf(".spec.template.spec.containers[?(@.name==\"%s\")].command", containerName)
	return GetResourceField("job", namespace, jobName, jsonpath)
}

// GetJobContainerArgs retrieves the args from a Job's container
func GetJobContainerArgs(namespace, jobName, containerName string) (string, error) {
	jsonpath := fmt.Sprintf(".spec.template.spec.containers[?(@.name==\"%s\")].args", containerName)
	return GetResourceField("job", namespace, jobName, jsonpath)
}

// GetJobContainerNames retrieves all container names from a Job
func GetJobContainerNames(namespace, jobName string) (string, error) {
	return GetResourceField("job", namespace, jobName, ".spec.template.spec.containers[*].name")
}

// GetServicePorts retrieves port names from a Service
func GetServicePorts(namespace, serviceName string) (string, error) {
	return GetResourceField("service", namespace, serviceName, ".spec.ports[*].name")
}

// GetJobEnvFrom retrieves envFrom configuration from a Job's container
func GetJobEnvFrom(namespace, jobName, containerName string) (string, error) {
	jsonpath := fmt.Sprintf(".spec.template.spec.containers[?(@.name==\"%s\")].envFrom", containerName)
	return GetResourceField("job", namespace, jobName, jsonpath)
}

// GetJobVolumes retrieves volume names from a Job
func GetJobVolumes(namespace, jobName string) (string, error) {
	return GetResourceField("job", namespace, jobName, ".spec.template.spec.volumes[*].name")
}

// GetJobVolumeMounts retrieves volume mount paths from a Job's container
func GetJobVolumeMounts(namespace, jobName, containerName string) (string, error) {
	jsonpath := fmt.Sprintf(".spec.template.spec.containers[?(@.name==\"%s\")].volumeMounts[*].mountPath", containerName)
	return GetResourceField("job", namespace, jobName, jsonpath)
}

// WaitForControllerReady waits for the controller-manager deployment to be ready
func WaitForControllerReady(namespace string, timeout string) error {
	_, _ = fmt.Fprintf(GinkgoWriter, "Waiting for controller-manager deployment to be ready...\n")
	cmd := exec.Command("kubectl", "wait", "deployment",
		"-l", "control-plane=controller-manager",
		"-n", namespace,
		"--for=condition=Available",
		"--timeout", timeout)
	_, err := Run(cmd)
	return err
}

// WaitForWebhookReady waits for the webhook service endpoint to be ready
func WaitForWebhookReady(namespace, serviceName string, timeout string) error {
	_, _ = fmt.Fprintf(GinkgoWriter, "Waiting for webhook service endpoint to be ready...\n")
	cmd := exec.Command("kubectl", "wait", "endpoints", serviceName,
		"-n", namespace,
		"--for=jsonpath={.subsets[0].addresses[0].ip}",
		"--timeout", timeout)
	_, err := Run(cmd)
	return err
}

// WaitForCertificateReady waits for the serving certificate to be ready
func WaitForCertificateReady(namespace, certName string, timeout string) error {
	_, _ = fmt.Fprintf(GinkgoWriter, "Waiting for certificate %s to be ready...\n", certName)
	cmd := exec.Command("kubectl", "wait", "certificate", certName,
		"-n", namespace,
		"--for=condition=Ready",
		"--timeout", timeout)
	_, err := Run(cmd)
	return err
}

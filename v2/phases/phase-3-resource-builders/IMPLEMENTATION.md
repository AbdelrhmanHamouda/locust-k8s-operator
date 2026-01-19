# Phase 3: Resource Builders - Implementation Plan

**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Prerequisites:** Phase 1 (v1 API Types), Phase 2 (Configuration System)  
**Requirements:** §3.3 Resilience (idempotent creation)

---

## Objective

Implement Job and Service builders matching Java `ResourceCreationHelpers.java` behavior. These builders are pure functions that take a LocustTest CR and operator configuration, returning fully-formed Kubernetes resources ready for creation.

---

## Day 1: Core Builders

### Task 3.1: Create Types and Constants

#### 3.1.1 Create `internal/resources/types.go`

```go
/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
...
*/

package resources

// OperationalMode represents the mode of a Locust node (master or worker).
type OperationalMode string

const (
	// Master represents the Locust master node that coordinates workers and serves the web UI.
	Master OperationalMode = "master"
	// Worker represents a Locust worker node that generates load.
	Worker OperationalMode = "worker"
)

// String returns the string representation of the operational mode.
func (m OperationalMode) String() string {
	return string(m)
}
```

#### 3.1.2 Create `internal/resources/constants.go`

Port this from Java `Constants.java`:

```go
package resources

// Port constants
const (
	// MasterPort is the port for master-worker communication.
	MasterPort = 5557
	// MasterBindPort is the secondary port for master-worker communication.
	MasterBindPort = 5558
	// WebUIPort is the port for the Locust web interface.
	WebUIPort = 8089
	// WorkerPort is the port exposed by worker containers.
	WorkerPort = 8080
	// DefaultMetricsExporterPort is the default port for the Prometheus metrics exporter.
	DefaultMetricsExporterPort = 9646
)

// Mount path constants
const (
	// DefaultMountPath is where the test ConfigMap is mounted.
	DefaultMountPath = "/lotest/src/"
	// LibMountPath is where the library ConfigMap is mounted.
	LibMountPath = "/opt/locust/lib"
)

// Label constants
const (
	// LabelTestName identifies the LocustTest CR name.
	LabelTestName = "performance-test-name"
	// LabelPodName identifies the specific pod (master or worker).
	LabelPodName = "performance-test-pod-name"
	// LabelManagedBy indicates the operator managing the resource.
	LabelManagedBy = "managed-by"
	// ManagedByValue is the value for the managed-by label.
	ManagedByValue = "locust-k8s-operator"
	// LabelApp is a standard app label.
	LabelApp = "app"
)

// Annotation constants for Prometheus scraping
const (
	// AnnotationPrometheusScrape enables Prometheus scraping.
	AnnotationPrometheusScrape = "prometheus.io/scrape"
	// AnnotationPrometheusPath specifies the metrics endpoint path.
	AnnotationPrometheusPath = "prometheus.io/path"
	// AnnotationPrometheusPort specifies the metrics port.
	AnnotationPrometheusPort = "prometheus.io/port"
	// MetricsEndpointPath is the path for the metrics endpoint.
	MetricsEndpointPath = "/metrics"
)

// Job constants
const (
	// BackoffLimit is the number of retries before marking a Job as failed.
	BackoffLimit = 0
	// RestartPolicyNever ensures pods are not restarted.
	RestartPolicyNever = "Never"
	// MasterReplicaCount is always 1 for the master node.
	MasterReplicaCount = 1
)

// Container constants
const (
	// LocustContainerName is the name of the main Locust container.
	LocustContainerName = "locust"
	// MetricsExporterContainerName is the name of the metrics sidecar container.
	MetricsExporterContainerName = "locust-metrics-exporter"
	// LibVolumeName is the name of the library volume.
	LibVolumeName = "lib"
)

// Environment variable names for metrics exporter
const (
	// ExporterURIEnvVar is the environment variable for the Locust URI.
	ExporterURIEnvVar = "LOCUST_EXPORTER_URI"
	// ExporterPortEnvVar is the environment variable for the exporter listen address.
	ExporterPortEnvVar = "LOCUST_EXPORTER_WEB_LISTEN_ADDRESS"
)

// Kafka environment variable names (passed through to containers)
const (
	EnvKafkaBootstrapServers   = "KAFKA_BOOTSTRAP_SERVERS"
	EnvKafkaSecurityEnabled    = "KAFKA_SECURITY_ENABLED"
	EnvKafkaSecurityProtocol   = "KAFKA_SECURITY_PROTOCOL_CONFIG"
	EnvKafkaSaslMechanism      = "KAFKA_SASL_MECHANISM"
	EnvKafkaSaslJaasConfig     = "KAFKA_SASL_JAAS_CONFIG"
	EnvKafkaUsername           = "KAFKA_USERNAME"
	EnvKafkaPassword           = "KAFKA_PASSWORD"
)

// Service constants
const (
	// ProtocolTCP is the TCP protocol for service ports.
	ProtocolTCP = "TCP"
	// PortNamePrefix is the prefix for port names.
	PortNamePrefix = "port"
	// MetricsPortName is the name for the metrics service port.
	MetricsPortName = "prometheus-metrics"
)

// Node affinity constants
const (
	// DefaultNodeMatchExpressionOperator is the default operator for node affinity expressions.
	DefaultNodeMatchExpressionOperator = "In"
)
```

---

### Task 3.2: Create Labels and Naming Helpers

**File:** `internal/resources/labels.go`

```go
package resources

import (
	"fmt"
	"strconv"
	"strings"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

// NodeName constructs the name for a master or worker node.
// Format: "{cr-name}-{mode}" with dots replaced by dashes.
// Example: "team-a.load-test" -> "team-a-load-test-master"
func NodeName(crName string, mode OperationalMode) string {
	sanitized := strings.ReplaceAll(crName, ".", "-")
	return fmt.Sprintf("%s-%s", sanitized, mode)
}

// BuildLabels constructs the labels for a pod template.
// These labels are used for pod selection and identification.
func BuildLabels(lt *locustv1.LocustTest, mode OperationalMode) map[string]string {
	nodeName := NodeName(lt.Name, mode)
	testName := strings.ReplaceAll(lt.Name, ".", "-")

	labels := map[string]string{
		LabelApp:       nodeName,
		LabelTestName:  testName,
		LabelPodName:   nodeName,
		LabelManagedBy: ManagedByValue,
	}

	// Merge user-defined labels from CR spec
	userLabels := getUserLabels(lt, mode)
	for k, v := range userLabels {
		labels[k] = v
	}

	return labels
}

// getUserLabels extracts user-defined labels for the specified mode from the CR spec.
func getUserLabels(lt *locustv1.LocustTest, mode OperationalMode) map[string]string {
	if lt.Spec.Labels == nil {
		return nil
	}

	if mode == Master {
		return lt.Spec.Labels.Master
	}
	return lt.Spec.Labels.Worker
}

// BuildAnnotations constructs the annotations for a pod template.
// Master pods get Prometheus scrape annotations for the metrics exporter.
func BuildAnnotations(lt *locustv1.LocustTest, mode OperationalMode, cfg *config.OperatorConfig) map[string]string {
	annotations := make(map[string]string)

	// Add Prometheus scrape annotations for master node only
	if mode == Master {
		annotations[AnnotationPrometheusScrape] = "true"
		annotations[AnnotationPrometheusPath] = MetricsEndpointPath
		annotations[AnnotationPrometheusPort] = strconv.Itoa(int(cfg.MetricsExporterPort))
	}

	// Merge user-defined annotations from CR spec
	userAnnotations := getUserAnnotations(lt, mode)
	for k, v := range userAnnotations {
		annotations[k] = v
	}

	return annotations
}

// getUserAnnotations extracts user-defined annotations for the specified mode from the CR spec.
func getUserAnnotations(lt *locustv1.LocustTest, mode OperationalMode) map[string]string {
	if lt.Spec.Annotations == nil {
		return nil
	}

	if mode == Master {
		return lt.Spec.Annotations.Master
	}
	return lt.Spec.Annotations.Worker
}
```

---

### Task 3.3: Create Port Helpers

**File:** `internal/resources/ports.go`

```go
package resources

import corev1 "k8s.io/api/core/v1"

// MasterPorts returns the container ports for a master node.
// Ports: 5557 (master), 5558 (bind), 8089 (web UI)
func MasterPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{ContainerPort: MasterPort, Protocol: corev1.ProtocolTCP},
		{ContainerPort: MasterBindPort, Protocol: corev1.ProtocolTCP},
		{ContainerPort: WebUIPort, Protocol: corev1.ProtocolTCP},
	}
}

// WorkerPorts returns the container ports for a worker node.
// Ports: 8080 (worker)
func WorkerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{ContainerPort: WorkerPort, Protocol: corev1.ProtocolTCP},
	}
}

// MasterPortInts returns the port numbers for master nodes as integers.
// Used for Service port configuration.
func MasterPortInts() []int32 {
	return []int32{MasterPort, MasterBindPort, WebUIPort}
}

// WorkerPortInts returns the port numbers for worker nodes as integers.
func WorkerPortInts() []int32 {
	return []int32{WorkerPort}
}
```

---

### Task 3.4: Create Command Builders

**File:** `internal/resources/command.go`

Reference Java templates:
```java
// Master: "%s --master --master-port=%d --expect-workers=%d --autostart --autoquit 60 --enable-rebalancing --only-summary"
// Worker: "%s --worker --master-port=%d --master-host=%s"
```

```go
package resources

import (
	"fmt"
	"strings"
)

// BuildMasterCommand constructs the command arguments for a master node.
// The command follows the template:
// "{seed} --master --master-port=5557 --expect-workers={N} --autostart --autoquit 60 --enable-rebalancing --only-summary"
func BuildMasterCommand(commandSeed string, workerReplicas int32) []string {
	// Build the full command string
	cmd := fmt.Sprintf("%s --master --master-port=%d --expect-workers=%d --autostart --autoquit 60 --enable-rebalancing --only-summary",
		commandSeed,
		MasterPort,
		workerReplicas,
	)

	// Split on whitespace to create argument list (matches Java split(" "))
	return strings.Fields(cmd)
}

// BuildWorkerCommand constructs the command arguments for a worker node.
// The command follows the template:
// "{seed} --worker --master-port=5557 --master-host={master-host}"
func BuildWorkerCommand(commandSeed string, masterHost string) []string {
	// Build the full command string
	cmd := fmt.Sprintf("%s --worker --master-port=%d --master-host=%s",
		commandSeed,
		MasterPort,
		masterHost,
	)

	// Split on whitespace to create argument list
	return strings.Fields(cmd)
}
```

---

## Day 2: Job and Service Builders

### Task 3.5: Create Job Builder

**File:** `internal/resources/job.go`

This is the largest file - port from `ResourceCreationHelpers.java`:

```go
package resources

import (
	"fmt"
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

// BuildMasterJob creates a Kubernetes Job for the Locust master node.
// The master coordinates workers, serves the web UI, and runs the metrics exporter sidecar.
func BuildMasterJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
	name := NodeName(lt.Name, Master)
	command := BuildMasterCommand(lt.Spec.MasterCommandSeed, lt.Spec.WorkerReplicas)

	// Build container list (locust + metrics exporter)
	containers := []corev1.Container{
		buildLocustContainer(lt, name, command, MasterPorts(), cfg),
		buildMetricsExporterContainer(cfg),
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: lt.Namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: cfg.TTLSecondsAfterFinished,
			Parallelism:             ptr(int32(MasterReplicaCount)),
			BackoffLimit:            ptr(int32(BackoffLimit)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      BuildLabels(lt, Master),
					Annotations: BuildAnnotations(lt, Master, cfg),
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					ImagePullSecrets: buildImagePullSecrets(lt),
					Containers:       containers,
					Volumes:          buildVolumes(lt, name),
					Affinity:         buildAffinity(lt, cfg),
					Tolerations:      buildTolerations(lt, cfg),
				},
			},
		},
	}
}

// BuildWorkerJob creates a Kubernetes Job for Locust worker nodes.
// Workers connect to the master and generate load based on the test configuration.
func BuildWorkerJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
	name := NodeName(lt.Name, Worker)
	masterHost := NodeName(lt.Name, Master)
	command := BuildWorkerCommand(lt.Spec.WorkerCommandSeed, masterHost)

	containers := []corev1.Container{
		buildLocustContainer(lt, name, command, WorkerPorts(), cfg),
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: lt.Namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: cfg.TTLSecondsAfterFinished,
			Parallelism:             ptr(lt.Spec.WorkerReplicas),
			BackoffLimit:            ptr(int32(BackoffLimit)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      BuildLabels(lt, Worker),
					Annotations: BuildAnnotations(lt, Worker, cfg),
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					ImagePullSecrets: buildImagePullSecrets(lt),
					Containers:       containers,
					Volumes:          buildVolumes(lt, name),
					Affinity:         buildAffinity(lt, cfg),
					Tolerations:      buildTolerations(lt, cfg),
				},
			},
		},
	}
}

// buildLocustContainer creates the main Locust container spec.
func buildLocustContainer(lt *locustv1.LocustTest, name string, command []string, ports []corev1.ContainerPort, cfg *config.OperatorConfig) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           lt.Spec.Image,
		ImagePullPolicy: corev1.PullPolicy(lt.Spec.ImagePullPolicy),
		Args:            command,
		Ports:           ports,
		VolumeMounts:    buildVolumeMounts(lt, name),
		Env:             buildKafkaEnvVars(cfg),
		Resources:       buildResourceRequirements(cfg, false),
	}
}

// buildMetricsExporterContainer creates the Prometheus metrics exporter sidecar container.
// Reference: https://github.com/ContainerSolutions/locust_exporter
func buildMetricsExporterContainer(cfg *config.OperatorConfig) corev1.Container {
	return corev1.Container{
		Name:            MetricsExporterContainerName,
		Image:           cfg.MetricsExporterImage,
		ImagePullPolicy: corev1.PullPolicy(cfg.MetricsExporterPullPolicy),
		Ports: []corev1.ContainerPort{
			{ContainerPort: cfg.MetricsExporterPort, Protocol: corev1.ProtocolTCP},
		},
		Env: []corev1.EnvVar{
			{Name: ExporterURIEnvVar, Value: fmt.Sprintf("http://localhost:%d", WebUIPort)},
			{Name: ExporterPortEnvVar, Value: fmt.Sprintf(":%d", cfg.MetricsExporterPort)},
		},
		Resources: buildResourceRequirements(cfg, true),
	}
}

// buildImagePullSecrets converts the list of secret names to LocalObjectReferences.
func buildImagePullSecrets(lt *locustv1.LocustTest) []corev1.LocalObjectReference {
	if len(lt.Spec.ImagePullSecrets) == 0 {
		return nil
	}

	secrets := make([]corev1.LocalObjectReference, len(lt.Spec.ImagePullSecrets))
	for i, name := range lt.Spec.ImagePullSecrets {
		secrets[i] = corev1.LocalObjectReference{Name: name}
	}
	return secrets
}

// buildVolumes creates the volume list for ConfigMap mounts.
func buildVolumes(lt *locustv1.LocustTest, nodeName string) []corev1.Volume {
	var volumes []corev1.Volume

	// Main test ConfigMap volume
	if lt.Spec.ConfigMap != "" {
		volumes = append(volumes, corev1.Volume{
			Name: nodeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: lt.Spec.ConfigMap},
				},
			},
		})
	}

	// Library ConfigMap volume
	if lt.Spec.LibConfigMap != "" {
		volumes = append(volumes, corev1.Volume{
			Name: LibVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: lt.Spec.LibConfigMap},
				},
			},
		})
	}

	return volumes
}

// buildVolumeMounts creates the volume mount list for the Locust container.
func buildVolumeMounts(lt *locustv1.LocustTest, nodeName string) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	// Main test ConfigMap mount
	if lt.Spec.ConfigMap != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      nodeName,
			MountPath: DefaultMountPath,
			ReadOnly:  false,
		})
	}

	// Library ConfigMap mount
	if lt.Spec.LibConfigMap != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      LibVolumeName,
			MountPath: LibMountPath,
			ReadOnly:  false,
		})
	}

	return mounts
}

// buildKafkaEnvVars creates the Kafka-related environment variables from operator config.
func buildKafkaEnvVars(cfg *config.OperatorConfig) []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: EnvKafkaBootstrapServers, Value: cfg.KafkaBootstrapServers},
		{Name: EnvKafkaSecurityEnabled, Value: strconv.FormatBool(cfg.KafkaSecurityEnabled)},
		{Name: EnvKafkaSecurityProtocol, Value: cfg.KafkaSecurityProtocol},
		{Name: EnvKafkaSaslMechanism, Value: cfg.KafkaSaslMechanism},
		{Name: EnvKafkaSaslJaasConfig, Value: cfg.KafkaSaslJaasConfig},
		{Name: EnvKafkaUsername, Value: cfg.KafkaUsername},
		{Name: EnvKafkaPassword, Value: cfg.KafkaPassword},
	}
}

// buildResourceRequirements creates the resource requests and limits for a container.
func buildResourceRequirements(cfg *config.OperatorConfig, isMetricsExporter bool) corev1.ResourceRequirements {
	var cpuReq, memReq, ephReq, cpuLim, memLim, ephLim string

	if isMetricsExporter {
		cpuReq = cfg.MetricsExporterCPURequest
		memReq = cfg.MetricsExporterMemRequest
		ephReq = cfg.MetricsExporterEphemeralStorageRequest
		cpuLim = cfg.MetricsExporterCPULimit
		memLim = cfg.MetricsExporterMemLimit
		ephLim = cfg.MetricsExporterEphemeralStorageLimit
	} else {
		cpuReq = cfg.PodCPURequest
		memReq = cfg.PodMemRequest
		ephReq = cfg.PodEphemeralStorageRequest
		cpuLim = cfg.PodCPULimit
		memLim = cfg.PodMemLimit
		ephLim = cfg.PodEphemeralStorageLimit
	}

	return corev1.ResourceRequirements{
		Requests: buildResourceList(cpuReq, memReq, ephReq),
		Limits:   buildResourceList(cpuLim, memLim, ephLim),
	}
}

// buildResourceList creates a ResourceList from string quantities.
func buildResourceList(cpu, memory, ephemeral string) corev1.ResourceList {
	resources := corev1.ResourceList{}

	if cpu != "" {
		resources[corev1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		resources[corev1.ResourceMemory] = resource.MustParse(memory)
	}
	if ephemeral != "" {
		resources[corev1.ResourceEphemeralStorage] = resource.MustParse(ephemeral)
	}

	return resources
}

// buildAffinity constructs pod affinity rules from the CR spec.
// Only applied when EnableAffinityCRInjection is true in operator config.
func buildAffinity(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *corev1.Affinity {
	if !cfg.EnableAffinityCRInjection || lt.Spec.Affinity == nil {
		return nil
	}

	affinity := &corev1.Affinity{}

	if lt.Spec.Affinity.NodeAffinity != nil && lt.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		nodeSelector := buildNodeSelector(lt.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
		affinity.NodeAffinity = &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: nodeSelector,
		}
	}

	return affinity
}

// buildNodeSelector converts the CR's node affinity map to a Kubernetes NodeSelector.
func buildNodeSelector(requirements map[string][]string) *corev1.NodeSelector {
	if len(requirements) == 0 {
		return nil
	}

	var matchExpressions []corev1.NodeSelectorRequirement
	for key, values := range requirements {
		matchExpressions = append(matchExpressions, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   values,
		})
	}

	return &corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{MatchExpressions: matchExpressions},
		},
	}
}

// buildTolerations constructs pod tolerations from the CR spec.
// Only applied when EnableTolerationsCRInjection is true in operator config.
func buildTolerations(lt *locustv1.LocustTest, cfg *config.OperatorConfig) []corev1.Toleration {
	if !cfg.EnableTolerationsCRInjection || len(lt.Spec.Tolerations) == 0 {
		return nil
	}

	tolerations := make([]corev1.Toleration, len(lt.Spec.Tolerations))
	for i, t := range lt.Spec.Tolerations {
		toleration := corev1.Toleration{
			Key:      t.Key,
			Operator: corev1.TolerationOperator(t.Operator),
			Effect:   corev1.TaintEffect(t.Effect),
		}

		// Only set Value if operator is "Equal"
		if t.Operator == "Equal" {
			toleration.Value = t.Value
		}

		tolerations[i] = toleration
	}

	return tolerations
}

// ptr is a helper function to create a pointer to a value.
func ptr[T any](v T) *T {
	return &v
}
```

---

### Task 3.6: Create Service Builder

**File:** `internal/resources/service.go`

```go
package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

// BuildMasterService creates a ClusterIP Service for the Locust master.
// This service exposes:
// - Port 5557: Master-worker communication
// - Port 5558: Master bind port
// - Port 9646: Prometheus metrics (via exporter sidecar)
// Note: Port 8089 (Web UI) is intentionally NOT exposed via this service.
// Users should port-forward or create an Ingress separately for web UI access.
func BuildMasterService(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *corev1.Service {
	name := NodeName(lt.Name, Master)

	// Build service ports (excluding web UI port, matching Java behavior)
	ports := []corev1.ServicePort{
		{
			Name:       PortNamePrefix + "5557",
			Port:       MasterPort,
			TargetPort: intstr.FromInt(MasterPort),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       PortNamePrefix + "5558",
			Port:       MasterBindPort,
			TargetPort: intstr.FromInt(MasterBindPort),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       MetricsPortName,
			Port:       cfg.MetricsExporterPort,
			TargetPort: intstr.FromInt(int(cfg.MetricsExporterPort)),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: lt.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				LabelPodName: name,
			},
			Ports: ports,
		},
	}
}
```

---

## Unit Testing Strategy

### Task 3.7: Create Test Files

Create test files alongside each source file:

#### `internal/resources/labels_test.go`

```go
package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

func TestNodeName(t *testing.T) {
	tests := []struct {
		name     string
		crName   string
		mode     OperationalMode
		expected string
	}{
		{"simple name master", "my-test", Master, "my-test-master"},
		{"simple name worker", "my-test", Worker, "my-test-worker"},
		{"dotted name master", "team-a.load-test", Master, "team-a-load-test-master"},
		{"dotted name worker", "team-a.load-test", Worker, "team-a-load-test-worker"},
		{"multiple dots", "a.b.c.test", Master, "a-b-c-test-master"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NodeName(tt.crName, tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildLabels(t *testing.T) {
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test"},
	}

	labels := BuildLabels(lt, Master)

	assert.Equal(t, "my-test-master", labels[LabelApp])
	assert.Equal(t, "my-test", labels[LabelTestName])
	assert.Equal(t, "my-test-master", labels[LabelPodName])
	assert.Equal(t, ManagedByValue, labels[LabelManagedBy])
}

func TestBuildLabels_WithUserLabels(t *testing.T) {
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test"},
		Spec: locustv1.LocustTestSpec{
			Labels: &locustv1.PodLabels{
				Master: map[string]string{"env": "prod", "team": "platform"},
			},
		},
	}

	labels := BuildLabels(lt, Master)

	assert.Equal(t, "prod", labels["env"])
	assert.Equal(t, "platform", labels["team"])
	assert.Equal(t, ManagedByValue, labels[LabelManagedBy]) // system labels preserved
}

func TestBuildAnnotations_Master(t *testing.T) {
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test"},
	}
	cfg := &config.OperatorConfig{MetricsExporterPort: 9646}

	annotations := BuildAnnotations(lt, Master, cfg)

	assert.Equal(t, "true", annotations[AnnotationPrometheusScrape])
	assert.Equal(t, MetricsEndpointPath, annotations[AnnotationPrometheusPath])
	assert.Equal(t, "9646", annotations[AnnotationPrometheusPort])
}

func TestBuildAnnotations_Worker(t *testing.T) {
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test"},
	}
	cfg := &config.OperatorConfig{MetricsExporterPort: 9646}

	annotations := BuildAnnotations(lt, Worker, cfg)

	// Workers should NOT have Prometheus annotations
	assert.Empty(t, annotations[AnnotationPrometheusScrape])
}
```

#### `internal/resources/command_test.go`

```go
package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMasterCommand(t *testing.T) {
	seed := "locust -f /lotest/src/test.py"
	workerReplicas := int32(5)

	cmd := BuildMasterCommand(seed, workerReplicas)

	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--master")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--expect-workers=5")
	assert.Contains(t, cmd, "--autostart")
	assert.Contains(t, cmd, "--autoquit")
	assert.Contains(t, cmd, "--enable-rebalancing")
	assert.Contains(t, cmd, "--only-summary")
}

func TestBuildWorkerCommand(t *testing.T) {
	seed := "locust -f /lotest/src/test.py"
	masterHost := "my-test-master"

	cmd := BuildWorkerCommand(seed, masterHost)

	assert.Contains(t, cmd, "locust")
	assert.Contains(t, cmd, "-f")
	assert.Contains(t, cmd, "/lotest/src/test.py")
	assert.Contains(t, cmd, "--worker")
	assert.Contains(t, cmd, "--master-port=5557")
	assert.Contains(t, cmd, "--master-host=my-test-master")
}

func TestBuildMasterCommand_SplitsCorrectly(t *testing.T) {
	// Ensure whitespace handling matches Java split(" ")
	seed := "locust   -f   /lotest/src/test.py" // multiple spaces
	cmd := BuildMasterCommand(seed, 3)

	// strings.Fields should collapse multiple spaces
	assert.Equal(t, "locust", cmd[0])
	assert.Equal(t, "-f", cmd[1])
	assert.Equal(t, "/lotest/src/test.py", cmd[2])
}
```

#### `internal/resources/job_test.go`

```go
package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

func newTestLocustTest() *locustv1.LocustTest {
	return &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv1.LocustTestSpec{
			MasterCommandSeed: "locust -f /lotest/src/test.py",
			WorkerCommandSeed: "locust -f /lotest/src/test.py",
			WorkerReplicas:    5,
			Image:             "locustio/locust:latest",
			ImagePullPolicy:   "Always",
			ConfigMap:         "test-configmap",
		},
	}
}

func newTestConfig() *config.OperatorConfig {
	return &config.OperatorConfig{
		PodCPURequest:              "250m",
		PodMemRequest:              "128Mi",
		PodCPULimit:                "1000m",
		PodMemLimit:                "1024Mi",
		MetricsExporterImage:       "containersol/locust_exporter:v0.5.0",
		MetricsExporterPort:        9646,
		MetricsExporterPullPolicy:  "Always",
		MetricsExporterCPURequest:  "250m",
		MetricsExporterMemRequest:  "128Mi",
		MetricsExporterCPULimit:    "1000m",
		MetricsExporterMemLimit:    "1024Mi",
		EnableAffinityCRInjection:  false,
		EnableTolerationsCRInjection: false,
	}
}

func TestBuildMasterJob(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Verify metadata
	assert.Equal(t, "my-test-master", job.Name)
	assert.Equal(t, "default", job.Namespace)

	// Verify job spec
	assert.Equal(t, int32(1), *job.Spec.Parallelism)
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit)

	// Verify pod template
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)

	// Verify containers (locust + metrics exporter)
	require.Len(t, job.Spec.Template.Spec.Containers, 2)
	assert.Equal(t, "my-test-master", job.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, MetricsExporterContainerName, job.Spec.Template.Spec.Containers[1].Name)

	// Verify volumes
	require.Len(t, job.Spec.Template.Spec.Volumes, 1)
	assert.Equal(t, "my-test-master", job.Spec.Template.Spec.Volumes[0].Name)
}

func TestBuildWorkerJob(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	// Verify metadata
	assert.Equal(t, "my-test-worker", job.Name)
	assert.Equal(t, "default", job.Namespace)

	// Verify job spec - workers should use WorkerReplicas for parallelism
	assert.Equal(t, int32(5), *job.Spec.Parallelism)
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit)

	// Verify containers (locust only, no metrics exporter)
	require.Len(t, job.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "my-test-worker", job.Spec.Template.Spec.Containers[0].Name)
}

func TestBuildMasterJob_WithTTL(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	ttl := int32(300)
	cfg.TTLSecondsAfterFinished = &ttl

	job := BuildMasterJob(lt, cfg)

	require.NotNil(t, job.Spec.TTLSecondsAfterFinished)
	assert.Equal(t, int32(300), *job.Spec.TTLSecondsAfterFinished)
}

func TestBuildMasterJob_WithImagePullSecrets(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.ImagePullSecrets = []string{"secret1", "secret2"}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	require.Len(t, job.Spec.Template.Spec.ImagePullSecrets, 2)
	assert.Equal(t, "secret1", job.Spec.Template.Spec.ImagePullSecrets[0].Name)
	assert.Equal(t, "secret2", job.Spec.Template.Spec.ImagePullSecrets[1].Name)
}

func TestBuildMasterJob_WithLibConfigMap(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.LibConfigMap = "lib-configmap"
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Should have 2 volumes: main configmap + lib configmap
	require.Len(t, job.Spec.Template.Spec.Volumes, 2)
	assert.Equal(t, LibVolumeName, job.Spec.Template.Spec.Volumes[1].Name)

	// Container should have 2 volume mounts
	require.Len(t, job.Spec.Template.Spec.Containers[0].VolumeMounts, 2)
	assert.Equal(t, DefaultMountPath, job.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, LibMountPath, job.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath)
}
```

#### `internal/resources/service_test.go`

```go
package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

func TestBuildMasterService(t *testing.T) {
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
	}
	cfg := &config.OperatorConfig{MetricsExporterPort: 9646}

	svc := BuildMasterService(lt, cfg)

	// Verify metadata
	assert.Equal(t, "my-test-master", svc.Name)
	assert.Equal(t, "default", svc.Namespace)

	// Verify spec
	assert.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type)

	// Verify selector
	assert.Equal(t, "my-test-master", svc.Spec.Selector[LabelPodName])

	// Verify ports (5557, 5558, 9646 - no 8089 web UI)
	require.Len(t, svc.Spec.Ports, 3)

	portNumbers := make(map[int32]bool)
	for _, p := range svc.Spec.Ports {
		portNumbers[p.Port] = true
	}

	assert.True(t, portNumbers[MasterPort], "should have master port 5557")
	assert.True(t, portNumbers[MasterBindPort], "should have bind port 5558")
	assert.True(t, portNumbers[9646], "should have metrics port 9646")
	assert.False(t, portNumbers[WebUIPort], "should NOT have web UI port 8089")
}

func TestBuildMasterService_CustomMetricsPort(t *testing.T) {
	lt := &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
	}
	cfg := &config.OperatorConfig{MetricsExporterPort: 9999}

	svc := BuildMasterService(lt, cfg)

	// Find metrics port
	var metricsPort *corev1.ServicePort
	for _, p := range svc.Spec.Ports {
		if p.Name == MetricsPortName {
			metricsPort = &p
			break
		}
	}

	require.NotNil(t, metricsPort)
	assert.Equal(t, int32(9999), metricsPort.Port)
}
```

---

## Verification Commands

After implementation, run these commands to verify:

```bash
cd locust-k8s-operator-go

# Ensure code compiles
make build

# Run unit tests
go test ./internal/resources/... -v

# Check test coverage
go test ./internal/resources/... -cover

# Run linter
golangci-lint run ./internal/resources/...
```

---

## Acceptance Criteria Verification

| Criteria | Verification Method |
|----------|---------------------|
| `BuildMasterJob()` produces correct Job | `TestBuildMasterJob` passes |
| `BuildWorkerJob()` produces correct Job | `TestBuildWorkerJob` passes |
| `BuildMasterService()` produces correct Service | `TestBuildMasterService` passes |
| Pure functions (no side effects) | No `Client` usage, no I/O in builders |
| Commands match Java templates | `TestBuildMasterCommand`, `TestBuildWorkerCommand` |
| Labels match Java output | `TestBuildLabels` |
| Prometheus annotations correct | `TestBuildAnnotations_Master` |

---

## Notes for Phase 4

The resource builders will be used in the reconciler like this:

```go
func (r *LocustTestReconciler) createResources(ctx context.Context, lt *locustv1.LocustTest) error {
    // Build resources
    masterService := resources.BuildMasterService(lt, r.Config)
    masterJob := resources.BuildMasterJob(lt, r.Config)
    workerJob := resources.BuildWorkerJob(lt, r.Config)

    // Set owner references
    if err := controllerutil.SetControllerReference(lt, masterService, r.Scheme); err != nil {
        return err
    }
    // ... set owner refs for jobs

    // Create resources
    if err := r.Create(ctx, masterService); err != nil && !apierrors.IsAlreadyExists(err) {
        return err
    }
    // ... create jobs
}
```

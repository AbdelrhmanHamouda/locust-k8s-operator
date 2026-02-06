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

package resources

import (
	"fmt"
	"strconv"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildMasterJob creates a Kubernetes Job for the Locust master node.
func BuildMasterJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig, logger logr.Logger) *batchv1.Job {
	nodeName := NodeName(lt.Name, Master)
	otelEnabled := IsOTelEnabled(lt)
	command := BuildMasterCommand(&lt.Spec.Master, lt.Spec.Worker.Replicas, otelEnabled, logger)

	return buildJob(lt, cfg, Master, nodeName, command)
}

// BuildWorkerJob creates a Kubernetes Job for the Locust worker nodes.
func BuildWorkerJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig, logger logr.Logger) *batchv1.Job {
	nodeName := NodeName(lt.Name, Worker)
	masterHost := NodeName(lt.Name, Master)
	otelEnabled := IsOTelEnabled(lt)
	command := BuildWorkerCommand(lt.Spec.Worker.Command, masterHost, otelEnabled, lt.Spec.Worker.ExtraArgs, logger)

	return buildJob(lt, cfg, Worker, nodeName, command)
}

// buildJob is the internal function that constructs a Job for either master or worker.
func buildJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig, mode OperationalMode, nodeName string, command []string) *batchv1.Job {
	labels := BuildLabels(lt, mode)
	annotations := BuildAnnotations(lt, mode, cfg)

	// Determine parallelism based on mode
	var parallelism int32
	if mode == Master {
		parallelism = MasterReplicaCount
	} else {
		parallelism = lt.Spec.Worker.Replicas
	}

	// Determine ports based on mode
	var ports []corev1.ContainerPort
	if mode == Master {
		ports = MasterPorts()
	} else {
		ports = WorkerPorts()
	}

	// Build containers
	containers := []corev1.Container{
		buildLocustContainer(lt, nodeName, command, ports, cfg, mode),
	}

	// Master gets the metrics exporter sidecar ONLY if OTel is disabled
	if mode == Master && !IsOTelEnabled(lt) {
		containers = append(containers, buildMetricsExporterContainer(cfg))
	}

	backoffLimit := int32(BackoffLimit)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: lt.Namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: cfg.TTLSecondsAfterFinished,
			Parallelism:             &parallelism,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:    corev1.RestartPolicyNever,
					ImagePullSecrets: buildImagePullSecrets(lt),
					Containers:       containers,
					Volumes:          buildVolumes(lt, nodeName, mode),
					Affinity:         buildAffinity(lt, cfg),
					Tolerations:      buildTolerations(lt, cfg),
					NodeSelector:     buildNodeSelector(lt),
				},
			},
		},
	}

	return job
}

// buildLocustContainer creates the main Locust container.
func buildLocustContainer(lt *locustv2.LocustTest, name string, command []string, ports []corev1.ContainerPort, cfg *config.OperatorConfig, mode OperationalMode) corev1.Container {
	container := corev1.Container{
		Name:            name,
		Image:           lt.Spec.Image,
		ImagePullPolicy: lt.Spec.ImagePullPolicy,
		Args:            command,
		Ports:           ports,
		Resources:       buildResourceRequirementsWithPrecedence(lt, cfg, mode),
		Env:             BuildEnvVars(lt, cfg),
		EnvFrom:         BuildEnvFrom(lt),
		VolumeMounts:    buildVolumeMounts(lt, name, mode),
	}

	// Default to IfNotPresent if not specified
	if container.ImagePullPolicy == "" {
		container.ImagePullPolicy = corev1.PullIfNotPresent
	}

	return container
}

// buildMetricsExporterContainer creates the Prometheus metrics exporter sidecar container.
func buildMetricsExporterContainer(cfg *config.OperatorConfig) corev1.Container {
	return corev1.Container{
		Name:            MetricsExporterContainerName,
		Image:           cfg.MetricsExporterImage,
		ImagePullPolicy: corev1.PullPolicy(cfg.MetricsExporterPullPolicy),
		Ports: []corev1.ContainerPort{
			{ContainerPort: cfg.MetricsExporterPort},
		},
		Resources: buildResourceRequirements(cfg, true),
		Env: []corev1.EnvVar{
			{
				Name:  ExporterURIEnvVar,
				Value: fmt.Sprintf("http://localhost:%d", WebUIPort),
			},
			{
				Name:  ExporterPortEnvVar,
				Value: fmt.Sprintf(":%d", cfg.MetricsExporterPort),
			},
		},
	}
}

// buildImagePullSecrets creates LocalObjectReferences for image pull secrets.
func buildImagePullSecrets(lt *locustv2.LocustTest) []corev1.LocalObjectReference {
	return lt.Spec.ImagePullSecrets
}

// buildVolumes creates the volumes for ConfigMap, LibConfigMap, Secrets, and user volumes.
func buildVolumes(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.Volume {
	var volumes []corev1.Volume

	// Get ConfigMap refs from v2 TestFiles config
	var configMapRef, libConfigMapRef string
	if lt.Spec.TestFiles != nil {
		configMapRef = lt.Spec.TestFiles.ConfigMapRef
		libConfigMapRef = lt.Spec.TestFiles.LibConfigMapRef
	}

	if configMapRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: nodeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapRef,
					},
				},
			},
		})
	}

	if libConfigMapRef != "" {
		volumes = append(volumes, corev1.Volume{
			Name: LibVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: libConfigMapRef,
					},
				},
			},
		})
	}

	// Add secret volumes from env.secretMounts
	secretVolumes := BuildSecretVolumes(lt)
	if len(secretVolumes) > 0 {
		volumes = append(volumes, secretVolumes...)
	}

	// Add user-defined volumes (filtered by target)
	userVolumes := BuildUserVolumes(lt, mode)
	if len(userVolumes) > 0 {
		volumes = append(volumes, userVolumes...)
	}

	return volumes
}

// buildVolumeMounts creates the volume mounts for ConfigMap, LibConfigMap, Secrets, and user mounts.
func buildVolumeMounts(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	// Get ConfigMap refs and mount paths from v2 TestFiles config
	var configMapRef, libConfigMapRef string
	srcMountPath := DefaultMountPath
	libMountPath := LibMountPath
	if lt.Spec.TestFiles != nil {
		configMapRef = lt.Spec.TestFiles.ConfigMapRef
		libConfigMapRef = lt.Spec.TestFiles.LibConfigMapRef
		if lt.Spec.TestFiles.SrcMountPath != "" {
			srcMountPath = lt.Spec.TestFiles.SrcMountPath
		}
		if lt.Spec.TestFiles.LibMountPath != "" {
			libMountPath = lt.Spec.TestFiles.LibMountPath
		}
	}

	if configMapRef != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      nodeName,
			MountPath: srcMountPath,
			ReadOnly:  false,
		})
	}

	if libConfigMapRef != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      LibVolumeName,
			MountPath: libMountPath,
			ReadOnly:  false,
		})
	}

	// Add secret mounts from env.secretMounts
	secretMounts := BuildSecretVolumeMounts(lt)
	if len(secretMounts) > 0 {
		mounts = append(mounts, secretMounts...)
	}

	// Add user-defined volume mounts (filtered by target)
	userMounts := BuildUserVolumeMounts(lt, mode)
	if len(userMounts) > 0 {
		mounts = append(mounts, userMounts...)
	}

	return mounts
}

// BuildKafkaEnvVars creates the Kafka environment variables for the Locust container.
func BuildKafkaEnvVars(cfg *config.OperatorConfig) []corev1.EnvVar {
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

// buildResourceRequirements creates resource requirements for containers.
// isMetricsExporter determines whether to use metrics exporter or locust container resources.
func buildResourceRequirements(cfg *config.OperatorConfig, isMetricsExporter bool) corev1.ResourceRequirements {
	var requests, limits corev1.ResourceList

	if isMetricsExporter {
		requests = buildResourceList(
			cfg.MetricsExporterCPURequest,
			cfg.MetricsExporterMemRequest,
			cfg.MetricsExporterEphemeralStorageRequest,
		)
		limits = buildResourceList(
			cfg.MetricsExporterCPULimit,
			cfg.MetricsExporterMemLimit,
			cfg.MetricsExporterEphemeralStorageLimit,
		)
	} else {
		requests = buildResourceList(
			cfg.PodCPURequest,
			cfg.PodMemRequest,
			cfg.PodEphemeralStorageRequest,
		)
		limits = buildResourceList(
			cfg.PodCPULimit,
			cfg.PodMemLimit,
			cfg.PodEphemeralStorageLimit,
		)
	}

	return corev1.ResourceRequirements{
		Requests: requests,
		Limits:   limits,
	}
}

// buildResourceRequirementsWithPrecedence implements resource precedence chain:
// Level 1: CR-level resources (complete override, same as native K8s)
// Level 2: Role-specific operator config (from Helm masterResources/workerResources)
// Level 3: Unified operator defaults (from Helm resources)
// CR resources are a COMPLETE OVERRIDE (not partial merge).
// Role-specific resources use FIELD-LEVEL FALLBACK to unified defaults.
func buildResourceRequirementsWithPrecedence(
	lt *locustv2.LocustTest,
	cfg *config.OperatorConfig,
	mode OperationalMode,
) corev1.ResourceRequirements {
	// Level 1: CR-level resources (highest precedence)
	// CR resources are a COMPLETE OVERRIDE (not partial merge) — same as native K8s
	var crResources *corev1.ResourceRequirements
	if mode == Master {
		crResources = &lt.Spec.Master.Resources
	} else if mode == Worker {
		crResources = &lt.Spec.Worker.Resources
	}

	if hasResourcesSpecified(crResources) {
		return *crResources // Complete override, return as-is
	}

	// Level 2: Role-specific operator config (from Helm masterResources/workerResources)
	// Each field independently falls through to Level 3 if empty.
	// This builds a resource spec where non-empty role-specific fields override
	// unified fields, and empty role-specific fields fall through to unified.
	var cpuReq, memReq, ephReq, cpuLim, memLim, ephLim string
	if mode == Master {
		cpuReq = cfg.MasterCPURequest
		memReq = cfg.MasterMemRequest
		ephReq = cfg.MasterEphemeralStorageRequest
		cpuLim = cfg.MasterCPULimit
		memLim = cfg.MasterMemLimit
		ephLim = cfg.MasterEphemeralStorageLimit
	} else {
		cpuReq = cfg.WorkerCPURequest
		memReq = cfg.WorkerMemRequest
		ephReq = cfg.WorkerEphemeralStorageRequest
		cpuLim = cfg.WorkerCPULimit
		memLim = cfg.WorkerMemLimit
		ephLim = cfg.WorkerEphemeralStorageLimit
	}

	// Field-level fallback: empty role-specific → unified
	if cpuReq == "" {
		cpuReq = cfg.PodCPURequest
	}
	if memReq == "" {
		memReq = cfg.PodMemRequest
	}
	if ephReq == "" {
		ephReq = cfg.PodEphemeralStorageRequest
	}
	if cpuLim == "" {
		cpuLim = cfg.PodCPULimit
	}
	if memLim == "" {
		memLim = cfg.PodMemLimit
	}
	if ephLim == "" {
		ephLim = cfg.PodEphemeralStorageLimit
	}

	return corev1.ResourceRequirements{
		Requests: buildResourceList(cpuReq, memReq, ephReq),
		Limits:   buildResourceList(cpuLim, memLim, ephLim),
	}
}

// hasResourcesSpecified checks if ResourceRequirements has any non-empty fields.
// This distinguishes "user set resources to empty" vs "user didn't set resources at all".
func hasResourcesSpecified(r *corev1.ResourceRequirements) bool {
	if r == nil {
		return false
	}
	return len(r.Requests) > 0 || len(r.Limits) > 0
}

// buildResourceList creates a ResourceList from CPU, memory, and ephemeral storage strings.
// Empty strings are skipped (not added to the resource list).
// Safe parsing is used (errors ignored) because values are pre-validated at operator startup.
func buildResourceList(cpu, memory, ephemeral string) corev1.ResourceList {
	resources := corev1.ResourceList{}

	if cpu != "" {
		// Safe: Already validated at startup in LoadConfig
		q, _ := resource.ParseQuantity(cpu)
		resources[corev1.ResourceCPU] = q
	}
	if memory != "" {
		q, _ := resource.ParseQuantity(memory)
		resources[corev1.ResourceMemory] = q
	}
	if ephemeral != "" {
		q, _ := resource.ParseQuantity(ephemeral)
		resources[corev1.ResourceEphemeralStorage] = q
	}

	return resources
}

// buildAffinity creates the pod affinity configuration from the CR spec.
// Returns nil if affinity injection is disabled or no affinity is specified.
func buildAffinity(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *corev1.Affinity {
	if !cfg.EnableAffinityCRInjection {
		return nil
	}

	if lt.Spec.Scheduling == nil || lt.Spec.Scheduling.Affinity == nil {
		return nil
	}

	// v2 uses standard corev1.Affinity directly
	return lt.Spec.Scheduling.Affinity
}

// buildTolerations creates pod tolerations from the CR spec.
// Returns nil if toleration injection is disabled or no tolerations are specified.
func buildTolerations(lt *locustv2.LocustTest, cfg *config.OperatorConfig) []corev1.Toleration {
	if !cfg.EnableTolerationsCRInjection {
		return nil
	}

	if lt.Spec.Scheduling == nil || lt.Spec.Scheduling.Tolerations == nil {
		return nil
	}

	// v2 uses standard corev1.Toleration directly
	return lt.Spec.Scheduling.Tolerations
}

// buildNodeSelector creates pod node selector from the CR spec.
// Returns nil if no node selector is specified.
func buildNodeSelector(lt *locustv2.LocustTest) map[string]string {
	if lt.Spec.Scheduling == nil || len(lt.Spec.Scheduling.NodeSelector) == 0 {
		return nil
	}

	return lt.Spec.Scheduling.NodeSelector
}

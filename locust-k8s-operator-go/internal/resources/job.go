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

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildMasterJob creates a Kubernetes Job for the Locust master node.
func BuildMasterJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
	nodeName := NodeName(lt.Name, Master)
	command := BuildMasterCommand(lt.Spec.MasterCommandSeed, lt.Spec.WorkerReplicas)

	return buildJob(lt, cfg, Master, nodeName, command)
}

// BuildWorkerJob creates a Kubernetes Job for the Locust worker nodes.
func BuildWorkerJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
	nodeName := NodeName(lt.Name, Worker)
	masterHost := NodeName(lt.Name, Master)
	command := BuildWorkerCommand(lt.Spec.WorkerCommandSeed, masterHost)

	return buildJob(lt, cfg, Worker, nodeName, command)
}

// buildJob is the internal function that constructs a Job for either master or worker.
func buildJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig, mode OperationalMode, nodeName string, command []string) *batchv1.Job {
	labels := BuildLabels(lt, mode)
	annotations := BuildAnnotations(lt, mode, cfg)

	// Determine parallelism based on mode
	var parallelism int32
	if mode == Master {
		parallelism = MasterReplicaCount
	} else {
		parallelism = lt.Spec.WorkerReplicas
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
		buildLocustContainer(lt, nodeName, command, ports, cfg),
	}

	// Master gets the metrics exporter sidecar
	if mode == Master {
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
					Volumes:          buildVolumes(lt, nodeName),
					Affinity:         buildAffinity(lt, cfg),
					Tolerations:      buildTolerations(lt, cfg),
				},
			},
		},
	}

	return job
}

// buildLocustContainer creates the main Locust container.
func buildLocustContainer(lt *locustv1.LocustTest, name string, command []string, ports []corev1.ContainerPort, cfg *config.OperatorConfig) corev1.Container {
	container := corev1.Container{
		Name:            name,
		Image:           lt.Spec.Image,
		ImagePullPolicy: corev1.PullPolicy(lt.Spec.ImagePullPolicy),
		Args:            command,
		Ports:           ports,
		Resources:       buildResourceRequirements(cfg, false),
		Env:             buildKafkaEnvVars(cfg),
		VolumeMounts:    buildVolumeMounts(lt, name),
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
func buildImagePullSecrets(lt *locustv1.LocustTest) []corev1.LocalObjectReference {
	if lt.Spec.ImagePullSecrets == nil {
		return nil
	}

	refs := make([]corev1.LocalObjectReference, len(lt.Spec.ImagePullSecrets))
	for i, secretName := range lt.Spec.ImagePullSecrets {
		refs[i] = corev1.LocalObjectReference{Name: secretName}
	}
	return refs
}

// buildVolumes creates the volumes for ConfigMap and LibConfigMap.
func buildVolumes(lt *locustv1.LocustTest, nodeName string) []corev1.Volume {
	var volumes []corev1.Volume

	if lt.Spec.ConfigMap != "" {
		volumes = append(volumes, corev1.Volume{
			Name: nodeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: lt.Spec.ConfigMap,
					},
				},
			},
		})
	}

	if lt.Spec.LibConfigMap != "" {
		volumes = append(volumes, corev1.Volume{
			Name: LibVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: lt.Spec.LibConfigMap,
					},
				},
			},
		})
	}

	return volumes
}

// buildVolumeMounts creates the volume mounts for ConfigMap and LibConfigMap.
func buildVolumeMounts(lt *locustv1.LocustTest, nodeName string) []corev1.VolumeMount {
	var mounts []corev1.VolumeMount

	if lt.Spec.ConfigMap != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      nodeName,
			MountPath: DefaultMountPath,
			ReadOnly:  false,
		})
	}

	if lt.Spec.LibConfigMap != "" {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      LibVolumeName,
			MountPath: LibMountPath,
			ReadOnly:  false,
		})
	}

	return mounts
}

// buildKafkaEnvVars creates the Kafka environment variables for the Locust container.
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

// buildResourceList creates a ResourceList from CPU, memory, and ephemeral storage strings.
// Empty strings are skipped (not added to the resource list).
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

// buildAffinity creates the pod affinity configuration from the CR spec.
// Returns nil if affinity injection is disabled or no affinity is specified.
func buildAffinity(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *corev1.Affinity {
	if !cfg.EnableAffinityCRInjection {
		return nil
	}

	if lt.Spec.Affinity == nil || lt.Spec.Affinity.NodeAffinity == nil {
		return nil
	}

	nodeSelector := buildNodeSelector(lt.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	if nodeSelector == nil {
		return nil
	}

	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: nodeSelector,
		},
	}
}

// buildNodeSelector creates a NodeSelector from the affinity requirements map.
func buildNodeSelector(requirements map[string]string) *corev1.NodeSelector {
	if len(requirements) == 0 {
		return nil
	}

	matchExpressions := make([]corev1.NodeSelectorRequirement, 0, len(requirements))
	for key, value := range requirements {
		matchExpressions = append(matchExpressions, corev1.NodeSelectorRequirement{
			Key:      key,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{value},
		})
	}

	return &corev1.NodeSelector{
		NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{
				MatchExpressions: matchExpressions,
			},
		},
	}
}

// buildTolerations creates pod tolerations from the CR spec.
// Returns nil if toleration injection is disabled or no tolerations are specified.
func buildTolerations(lt *locustv1.LocustTest, cfg *config.OperatorConfig) []corev1.Toleration {
	if !cfg.EnableTolerationsCRInjection {
		return nil
	}

	if lt.Spec.Tolerations == nil {
		return nil
	}

	tolerations := make([]corev1.Toleration, len(lt.Spec.Tolerations))
	for i, t := range lt.Spec.Tolerations {
		toleration := corev1.Toleration{
			Key:      t.Key,
			Operator: corev1.TolerationOperator(t.Operator),
			Effect:   corev1.TaintEffect(t.Effect),
		}

		// Only set Value if Operator is Equal
		if t.Operator == "Equal" {
			toleration.Value = t.Value
		}

		tolerations[i] = toleration
	}

	return tolerations
}

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
	"testing"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const secretTLSCertsVolumeName = "secret-tls-certs"

func newTestLocustTest() *locustv2.LocustTest {
	return &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image:           "locustio/locust:latest",
			ImagePullPolicy: corev1.PullAlways,
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
			TestFiles: &locustv2.TestFilesConfig{
				ConfigMapRef: "my-test-configmap",
			},
		},
	}
}

func newTestConfig() *config.OperatorConfig {
	return &config.OperatorConfig{
		PodCPURequest:              "250m",
		PodMemRequest:              "128Mi",
		PodEphemeralStorageRequest: "30M",
		PodCPULimit:                "1000m",
		PodMemLimit:                "1024Mi",
		PodEphemeralStorageLimit:   "50M",

		MetricsExporterImage:                   "containersol/locust_exporter:v0.5.0",
		MetricsExporterPort:                    9646,
		MetricsExporterPullPolicy:              "Always",
		MetricsExporterCPURequest:              "250m",
		MetricsExporterMemRequest:              "128Mi",
		MetricsExporterEphemeralStorageRequest: "30M",
		MetricsExporterCPULimit:                "1000m",
		MetricsExporterMemLimit:                "1024Mi",
		MetricsExporterEphemeralStorageLimit:   "50M",

		KafkaBootstrapServers: "localhost:9092",
		KafkaSecurityEnabled:  false,
		KafkaSecurityProtocol: "SASL_PLAINTEXT",
		KafkaSaslMechanism:    "SCRAM-SHA-512",
	}
}

func TestBuildMasterJob(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	require.NotNil(t, job)
	assert.Equal(t, "my-test-master", job.Name)
	assert.Equal(t, "default", job.Namespace)
}

func TestBuildMasterJob_Metadata(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	assert.Equal(t, "my-test-master", job.Name)
	assert.Equal(t, "default", job.Namespace)
}

func TestBuildMasterJob_Parallelism(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	require.NotNil(t, job.Spec.Parallelism)
	assert.Equal(t, int32(1), *job.Spec.Parallelism, "Master parallelism should always be 1")
}

func TestBuildMasterJob_Containers(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	containers := job.Spec.Template.Spec.Containers
	assert.Len(t, containers, 2, "Master should have 2 containers (locust + metrics exporter)")

	// Find container names
	containerNames := make([]string, len(containers))
	for i, c := range containers {
		containerNames[i] = c.Name
	}
	assert.Contains(t, containerNames, "my-test-master")
	assert.Contains(t, containerNames, MetricsExporterContainerName)
}

func TestBuildMasterJob_WithTTL(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	ttl := int32(3600)
	cfg.TTLSecondsAfterFinished = &ttl

	job := BuildMasterJob(lt, cfg)

	require.NotNil(t, job.Spec.TTLSecondsAfterFinished)
	assert.Equal(t, int32(3600), *job.Spec.TTLSecondsAfterFinished)
}

func TestBuildMasterJob_WithImagePullSecrets(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
		{Name: "my-registry-secret"},
		{Name: "another-secret"},
	}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	secrets := job.Spec.Template.Spec.ImagePullSecrets
	assert.Len(t, secrets, 2)
	assert.Equal(t, "my-registry-secret", secrets[0].Name)
	assert.Equal(t, "another-secret", secrets[1].Name)
}

func TestBuildMasterJob_WithLibConfigMap(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.TestFiles.LibConfigMapRef = "my-lib-configmap"
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	volumes := job.Spec.Template.Spec.Volumes
	assert.Len(t, volumes, 2, "Should have 2 volumes (configmap + lib)")

	// Check lib volume exists
	var libVolumeFound bool
	for _, v := range volumes {
		if v.Name == LibVolumeName {
			libVolumeFound = true
			assert.Equal(t, "my-lib-configmap", v.ConfigMap.Name)
		}
	}
	assert.True(t, libVolumeFound, "Lib volume should exist")

	// Check volume mounts
	container := job.Spec.Template.Spec.Containers[0]
	var libMountFound bool
	for _, m := range container.VolumeMounts {
		if m.Name == LibVolumeName {
			libMountFound = true
			assert.Equal(t, LibMountPath, m.MountPath)
		}
	}
	assert.True(t, libMountFound, "Lib volume mount should exist")
}

func TestBuildMasterJob_Labels(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	labels := job.Spec.Template.Labels
	assert.Equal(t, "my-test", labels[LabelApp])
	assert.Equal(t, "my-test-master", labels[LabelPodName])
	assert.Equal(t, ManagedByValue, labels[LabelManagedBy])
}

func TestBuildMasterJob_Annotations(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	annotations := job.Spec.Template.Annotations
	assert.Equal(t, "true", annotations[AnnotationPrometheusScrape])
	assert.Equal(t, MetricsEndpointPath, annotations[AnnotationPrometheusPath])
	assert.Equal(t, "9646", annotations[AnnotationPrometheusPort])
}

func TestBuildWorkerJob(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	require.NotNil(t, job)
	assert.Equal(t, "my-test-worker", job.Name)
	assert.Equal(t, "default", job.Namespace)
}

func TestBuildWorkerJob_Parallelism(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Worker.Replicas = 5
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	require.NotNil(t, job.Spec.Parallelism)
	assert.Equal(t, int32(5), *job.Spec.Parallelism, "Worker parallelism should equal Worker.Replicas")
}

func TestBuildWorkerJob_Containers(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	containers := job.Spec.Template.Spec.Containers
	assert.Len(t, containers, 1, "Worker should have 1 container only")
	assert.Equal(t, "my-test-worker", containers[0].Name)
}

func TestBuildWorkerJob_NoPrometheusAnnotations(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	annotations := job.Spec.Template.Annotations
	assert.Empty(t, annotations[AnnotationPrometheusScrape])
	assert.Empty(t, annotations[AnnotationPrometheusPath])
	assert.Empty(t, annotations[AnnotationPrometheusPort])
}

func TestBuildResourceRequirements(t *testing.T) {
	cfg := newTestConfig()

	resources := buildResourceRequirements(cfg, false)

	assert.Equal(t, "250m", resources.Requests.Cpu().String())
	assert.Equal(t, "128Mi", resources.Requests.Memory().String())
	assert.Equal(t, "1", resources.Limits.Cpu().String())
	assert.Equal(t, "1Gi", resources.Limits.Memory().String())
}

func TestBuildResourceRequirements_MetricsExporter(t *testing.T) {
	cfg := newTestConfig()

	resources := buildResourceRequirements(cfg, true)

	assert.Equal(t, "250m", resources.Requests.Cpu().String())
	assert.Equal(t, "128Mi", resources.Requests.Memory().String())
}

func TestBuildAffinity_Disabled(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = &locustv2.SchedulingConfig{
		Affinity: &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "node-type",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"performance"},
								},
							},
						},
					},
				},
			},
		},
	}
	cfg := newTestConfig()
	cfg.EnableAffinityCRInjection = false

	job := BuildMasterJob(lt, cfg)

	assert.Nil(t, job.Spec.Template.Spec.Affinity, "Affinity should be nil when feature flag is disabled")
}

func TestBuildAffinity_Enabled(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = &locustv2.SchedulingConfig{
		Affinity: &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "node-type",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"performance"},
								},
							},
						},
					},
				},
			},
		},
	}
	cfg := newTestConfig()
	cfg.EnableAffinityCRInjection = true

	job := BuildMasterJob(lt, cfg)

	require.NotNil(t, job.Spec.Template.Spec.Affinity)
	require.NotNil(t, job.Spec.Template.Spec.Affinity.NodeAffinity)
	require.NotNil(t, job.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)

	terms := job.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	require.Len(t, terms, 1)
	require.Len(t, terms[0].MatchExpressions, 1)
	assert.Equal(t, "node-type", terms[0].MatchExpressions[0].Key)
	assert.Equal(t, corev1.NodeSelectorOpIn, terms[0].MatchExpressions[0].Operator)
	assert.Contains(t, terms[0].MatchExpressions[0].Values, "performance")
}

func TestBuildTolerations_Disabled(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = &locustv2.SchedulingConfig{
		Tolerations: []corev1.Toleration{
			{
				Key:      "dedicated",
				Operator: corev1.TolerationOpEqual,
				Value:    "performance",
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
	}
	cfg := newTestConfig()
	cfg.EnableTolerationsCRInjection = false

	job := BuildMasterJob(lt, cfg)

	assert.Nil(t, job.Spec.Template.Spec.Tolerations, "Tolerations should be nil when feature flag is disabled")
}

func TestBuildTolerations_Enabled(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = &locustv2.SchedulingConfig{
		Tolerations: []corev1.Toleration{
			{
				Key:      "dedicated",
				Operator: corev1.TolerationOpEqual,
				Value:    "performance",
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
	}
	cfg := newTestConfig()
	cfg.EnableTolerationsCRInjection = true

	job := BuildMasterJob(lt, cfg)

	require.Len(t, job.Spec.Template.Spec.Tolerations, 1)
	assert.Equal(t, "dedicated", job.Spec.Template.Spec.Tolerations[0].Key)
	assert.Equal(t, corev1.TolerationOpEqual, job.Spec.Template.Spec.Tolerations[0].Operator)
	assert.Equal(t, "performance", job.Spec.Template.Spec.Tolerations[0].Value)
	assert.Equal(t, corev1.TaintEffectNoSchedule, job.Spec.Template.Spec.Tolerations[0].Effect)
}

func TestBuildTolerations_ExistsOperator(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = &locustv2.SchedulingConfig{
		Tolerations: []corev1.Toleration{
			{
				Key:      "node.kubernetes.io/not-ready",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoExecute,
			},
		},
	}
	cfg := newTestConfig()
	cfg.EnableTolerationsCRInjection = true

	job := BuildMasterJob(lt, cfg)

	require.Len(t, job.Spec.Template.Spec.Tolerations, 1)
	assert.Equal(t, corev1.TolerationOpExists, job.Spec.Template.Spec.Tolerations[0].Operator)
	assert.Empty(t, job.Spec.Template.Spec.Tolerations[0].Value, "Value should be empty for Exists operator")
}

func TestBuildMasterJob_EmptyImagePullPolicy(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.ImagePullPolicy = "" // Empty should default to IfNotPresent
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]
	assert.Equal(t, corev1.PullIfNotPresent, container.ImagePullPolicy)
}

func TestBuildMasterJob_NoConfigMap(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.TestFiles = nil // No test files config
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	assert.Empty(t, job.Spec.Template.Spec.Volumes)
	assert.Empty(t, job.Spec.Template.Spec.Containers[0].VolumeMounts)
}

func TestBuildMasterJob_KafkaEnvVars(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	cfg.KafkaSecurityEnabled = true
	cfg.KafkaBootstrapServers = "kafka.example.com:9092"
	cfg.KafkaSecurityProtocol = "SASL_SSL"
	cfg.KafkaUsername = "user"
	cfg.KafkaPassword = "secret"

	job := BuildMasterJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]
	envMap := make(map[string]string)
	for _, env := range container.Env {
		envMap[env.Name] = env.Value
	}

	assert.Equal(t, "kafka.example.com:9092", envMap["KAFKA_BOOTSTRAP_SERVERS"])
	assert.Equal(t, "SASL_SSL", envMap["KAFKA_SECURITY_PROTOCOL_CONFIG"])
	assert.Equal(t, "user", envMap["KAFKA_USERNAME"])
	assert.Equal(t, "secret", envMap["KAFKA_PASSWORD"])
}

func TestBuildAffinity_NilScheduling(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = nil
	cfg := newTestConfig()
	cfg.EnableAffinityCRInjection = true

	job := BuildMasterJob(lt, cfg)

	assert.Nil(t, job.Spec.Template.Spec.Affinity)
}

func TestBuildAffinity_NilAffinity(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Scheduling = &locustv2.SchedulingConfig{
		Affinity: nil,
	}
	cfg := newTestConfig()
	cfg.EnableAffinityCRInjection = true

	job := BuildMasterJob(lt, cfg)

	assert.Nil(t, job.Spec.Template.Spec.Affinity)
}

func TestBuildMasterJob_Completions(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Master job should not have Completions set (nil means run to completion)
	assert.Nil(t, job.Spec.Completions)
}

func TestBuildMasterJob_BackoffLimit(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	require.NotNil(t, job.Spec.BackoffLimit)
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit)
}

func TestBuildMasterJob_WithEnvConfigMapRef(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Env = &locustv2.EnvConfig{
		ConfigMapRefs: []locustv2.ConfigMapEnvSource{
			{Name: "app-config", Prefix: "APP_"},
		},
	}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]
	require.Len(t, container.EnvFrom, 1)
	assert.NotNil(t, container.EnvFrom[0].ConfigMapRef)
	assert.Equal(t, "app-config", container.EnvFrom[0].ConfigMapRef.Name)
	assert.Equal(t, "APP_", container.EnvFrom[0].Prefix)
}

func TestBuildMasterJob_WithEnvSecretRef(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Env = &locustv2.EnvConfig{
		SecretRefs: []locustv2.SecretEnvSource{
			{Name: "api-credentials"},
		},
	}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]
	require.Len(t, container.EnvFrom, 1)
	assert.NotNil(t, container.EnvFrom[0].SecretRef)
	assert.Equal(t, "api-credentials", container.EnvFrom[0].SecretRef.Name)
}

func TestBuildMasterJob_WithEnvVariables(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Env = &locustv2.EnvConfig{
		Variables: []corev1.EnvVar{
			{Name: "TARGET_HOST", Value: "https://example.com"},
			{Name: "LOG_LEVEL", Value: "DEBUG"},
		},
	}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]
	envMap := make(map[string]string)
	for _, env := range container.Env {
		envMap[env.Name] = env.Value
	}

	// User vars should be present
	assert.Equal(t, "https://example.com", envMap["TARGET_HOST"])
	assert.Equal(t, "DEBUG", envMap["LOG_LEVEL"])

	// Kafka vars should still be present
	assert.Contains(t, envMap, "KAFKA_BOOTSTRAP_SERVERS")
}

func TestBuildMasterJob_WithSecretMount(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Env = &locustv2.EnvConfig{
		SecretMounts: []locustv2.SecretMount{
			{Name: "tls-certs", MountPath: "/etc/locust/certs", ReadOnly: true},
		},
	}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Check volume exists
	var secretVolumeFound bool
	for _, v := range job.Spec.Template.Spec.Volumes {
		if v.Name == secretTLSCertsVolumeName {
			secretVolumeFound = true
			assert.NotNil(t, v.Secret)
			assert.Equal(t, "tls-certs", v.Secret.SecretName)
		}
	}
	assert.True(t, secretVolumeFound, "Secret volume should exist")

	// Check volume mount exists
	container := job.Spec.Template.Spec.Containers[0]
	var secretMountFound bool
	for _, m := range container.VolumeMounts {
		if m.Name == secretTLSCertsVolumeName {
			secretMountFound = true
			assert.Equal(t, "/etc/locust/certs", m.MountPath)
			assert.True(t, m.ReadOnly)
		}
	}
	assert.True(t, secretMountFound, "Secret volume mount should exist")
}

func TestBuildMasterJob_EnvCombinesKafkaAndUser(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Env = &locustv2.EnvConfig{
		Variables: []corev1.EnvVar{
			{Name: "USER_VAR", Value: "user-value"},
		},
	}
	cfg := newTestConfig()
	cfg.KafkaBootstrapServers = "kafka:9092"

	job := BuildMasterJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]

	// Should have 7 Kafka vars + 1 user var = 8 total
	assert.Len(t, container.Env, 8)

	// Kafka vars come first
	assert.Equal(t, "KAFKA_BOOTSTRAP_SERVERS", container.Env[0].Name)
	assert.Equal(t, "kafka:9092", container.Env[0].Value)

	// User var comes last
	assert.Equal(t, "USER_VAR", container.Env[7].Name)
	assert.Equal(t, "user-value", container.Env[7].Value)
}

func TestBuildWorkerJob_WithEnvConfig(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Env = &locustv2.EnvConfig{
		ConfigMapRefs: []locustv2.ConfigMapEnvSource{
			{Name: "app-config"},
		},
		Variables: []corev1.EnvVar{
			{Name: "TARGET_HOST", Value: "https://example.com"},
		},
		SecretMounts: []locustv2.SecretMount{
			{Name: "tls-certs", MountPath: "/etc/certs"},
		},
	}
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	container := job.Spec.Template.Spec.Containers[0]

	// EnvFrom should have ConfigMapRef
	require.Len(t, container.EnvFrom, 1)
	assert.Equal(t, "app-config", container.EnvFrom[0].ConfigMapRef.Name)

	// Env should have Kafka + user vars
	envMap := make(map[string]string)
	for _, env := range container.Env {
		envMap[env.Name] = env.Value
	}
	assert.Equal(t, "https://example.com", envMap["TARGET_HOST"])
	assert.Contains(t, envMap, "KAFKA_BOOTSTRAP_SERVERS")

	// Secret mount should exist
	var secretMountFound bool
	for _, m := range container.VolumeMounts {
		if m.Name == secretTLSCertsVolumeName {
			secretMountFound = true
		}
	}
	assert.True(t, secretMountFound)
}

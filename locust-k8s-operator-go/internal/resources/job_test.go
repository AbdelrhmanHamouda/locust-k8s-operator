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

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			WorkerReplicas:    3,
			Image:             "locustio/locust:latest",
			ImagePullPolicy:   "Always",
			ConfigMap:         "my-test-configmap",
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
	lt.Spec.ImagePullSecrets = []string{"my-registry-secret", "another-secret"}
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	secrets := job.Spec.Template.Spec.ImagePullSecrets
	assert.Len(t, secrets, 2)
	assert.Equal(t, "my-registry-secret", secrets[0].Name)
	assert.Equal(t, "another-secret", secrets[1].Name)
}

func TestBuildMasterJob_WithLibConfigMap(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.LibConfigMap = "my-lib-configmap"
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
	lt.Spec.WorkerReplicas = 5
	cfg := newTestConfig()

	job := BuildWorkerJob(lt, cfg)

	require.NotNil(t, job.Spec.Parallelism)
	assert.Equal(t, int32(5), *job.Spec.Parallelism, "Worker parallelism should equal WorkerReplicas")
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
	lt.Spec.Affinity = &locustv1.LocustTestAffinity{
		NodeAffinity: &locustv1.LocustTestNodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{
				"node-type": "performance",
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
	lt.Spec.Affinity = &locustv1.LocustTestAffinity{
		NodeAffinity: &locustv1.LocustTestNodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{
				"node-type": "performance",
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
	lt.Spec.Tolerations = []locustv1.LocustTestToleration{
		{
			Key:      "dedicated",
			Operator: "Equal",
			Value:    "performance",
			Effect:   "NoSchedule",
		},
	}
	cfg := newTestConfig()
	cfg.EnableTolerationsCRInjection = false

	job := BuildMasterJob(lt, cfg)

	assert.Nil(t, job.Spec.Template.Spec.Tolerations, "Tolerations should be nil when feature flag is disabled")
}

func TestBuildTolerations_Enabled(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.Tolerations = []locustv1.LocustTestToleration{
		{
			Key:      "dedicated",
			Operator: "Equal",
			Value:    "performance",
			Effect:   "NoSchedule",
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

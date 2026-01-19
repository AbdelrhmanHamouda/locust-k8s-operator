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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

func TestConvertTo_FullSpec(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-locust",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust -f /lotest/src/locustfile.py",
			WorkerCommandSeed: "locust -f /lotest/src/locustfile.py",
			WorkerReplicas:    10,
			Image:             "locustio/locust:2.43.1",
			ImagePullPolicy:   "IfNotPresent",
			ImagePullSecrets:  []string{"my-registry-secret"},
			ConfigMap:         "locust-tests",
			LibConfigMap:      "locust-lib",
			Labels: &PodLabels{
				Master: map[string]string{"app": "locust-master"},
				Worker: map[string]string{"app": "locust-worker"},
			},
			Annotations: &PodAnnotations{
				Master: map[string]string{"prometheus.io/scrape": "true"},
			},
			Tolerations: []LocustTestToleration{
				{Key: "dedicated", Operator: "Equal", Value: "locust", Effect: "NoSchedule"},
			},
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	// Verify metadata
	assert.Equal(t, "test-locust", dst.Name)
	assert.Equal(t, "default", dst.Namespace)

	// Verify image config
	assert.Equal(t, "locustio/locust:2.43.1", dst.Spec.Image)
	assert.Equal(t, corev1.PullIfNotPresent, dst.Spec.ImagePullPolicy)
	require.Len(t, dst.Spec.ImagePullSecrets, 1)
	assert.Equal(t, "my-registry-secret", dst.Spec.ImagePullSecrets[0].Name)

	// Verify master config
	assert.Equal(t, "locust -f /lotest/src/locustfile.py", dst.Spec.Master.Command)
	assert.True(t, *dst.Spec.Master.Autostart)
	require.NotNil(t, dst.Spec.Master.Autoquit)
	assert.True(t, dst.Spec.Master.Autoquit.Enabled)
	assert.Equal(t, int32(60), dst.Spec.Master.Autoquit.Timeout)
	assert.Equal(t, "locust-master", dst.Spec.Master.Labels["app"])
	assert.Equal(t, "true", dst.Spec.Master.Annotations["prometheus.io/scrape"])

	// Verify worker config
	assert.Equal(t, "locust -f /lotest/src/locustfile.py", dst.Spec.Worker.Command)
	assert.Equal(t, int32(10), dst.Spec.Worker.Replicas)
	assert.Equal(t, "locust-worker", dst.Spec.Worker.Labels["app"])

	// Verify test files config
	require.NotNil(t, dst.Spec.TestFiles)
	assert.Equal(t, "locust-tests", dst.Spec.TestFiles.ConfigMapRef)
	assert.Equal(t, "locust-lib", dst.Spec.TestFiles.LibConfigMapRef)

	// Verify scheduling config
	require.NotNil(t, dst.Spec.Scheduling)
	require.Len(t, dst.Spec.Scheduling.Tolerations, 1)
	assert.Equal(t, "dedicated", dst.Spec.Scheduling.Tolerations[0].Key)
}

func TestConvertFrom_FullSpec(t *testing.T) {
	src := &v2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-locust",
			Namespace: "default",
		},
		Spec: v2.LocustTestSpec{
			Image:           "locustio/locust:2.43.1",
			ImagePullPolicy: corev1.PullAlways,
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "my-secret"},
			},
			Master: v2.MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
				Labels:  map[string]string{"tier": "master"},
			},
			Worker: v2.WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 5,
				Labels:   map[string]string{"tier": "worker"},
			},
			TestFiles: &v2.TestFilesConfig{
				ConfigMapRef:    "tests-cm",
				LibConfigMapRef: "lib-cm",
			},
		},
	}

	dst := &LocustTest{}
	err := dst.ConvertFrom(src)
	require.NoError(t, err)

	// Verify metadata
	assert.Equal(t, "test-locust", dst.Name)

	// Verify flat fields
	assert.Equal(t, "locust -f /lotest/src/locustfile.py", dst.Spec.MasterCommandSeed)
	assert.Equal(t, "locust -f /lotest/src/locustfile.py", dst.Spec.WorkerCommandSeed)
	assert.Equal(t, int32(5), dst.Spec.WorkerReplicas)
	assert.Equal(t, "locustio/locust:2.43.1", dst.Spec.Image)
	assert.Equal(t, "Always", dst.Spec.ImagePullPolicy)

	// Verify labels
	require.NotNil(t, dst.Spec.Labels)
	assert.Equal(t, "master", dst.Spec.Labels.Master["tier"])
	assert.Equal(t, "worker", dst.Spec.Labels.Worker["tier"])

	// Verify test files
	assert.Equal(t, "tests-cm", dst.Spec.ConfigMap)
	assert.Equal(t, "lib-cm", dst.Spec.LibConfigMap)
}

func TestRoundTrip_V1ToV2ToV1(t *testing.T) {
	original := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "roundtrip-test",
			Namespace: "test-ns",
		},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust -f /lotest/src/test.py",
			WorkerCommandSeed: "locust -f /lotest/src/test.py",
			WorkerReplicas:    3,
			Image:             "locustio/locust:latest",
			ConfigMap:         "my-tests",
		},
	}

	// Convert v1 -> v2
	hub := &v2.LocustTest{}
	err := original.ConvertTo(hub)
	require.NoError(t, err)

	// Convert v2 -> v1
	result := &LocustTest{}
	err = result.ConvertFrom(hub)
	require.NoError(t, err)

	// Verify round-trip preserved v1 fields
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.Namespace, result.Namespace)
	assert.Equal(t, original.Spec.MasterCommandSeed, result.Spec.MasterCommandSeed)
	assert.Equal(t, original.Spec.WorkerCommandSeed, result.Spec.WorkerCommandSeed)
	assert.Equal(t, original.Spec.WorkerReplicas, result.Spec.WorkerReplicas)
	assert.Equal(t, original.Spec.Image, result.Spec.Image)
	assert.Equal(t, original.Spec.ConfigMap, result.Spec.ConfigMap)
}

func TestConvertTo_MinimalSpec(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "minimal-test",
		},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust",
			WorkerCommandSeed: "locust",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	// Verify required fields
	assert.Equal(t, "locust", dst.Spec.Master.Command)
	assert.Equal(t, "locust", dst.Spec.Worker.Command)
	assert.Equal(t, int32(1), dst.Spec.Worker.Replicas)

	// Verify optional fields are nil/empty
	assert.Nil(t, dst.Spec.TestFiles)
	assert.Nil(t, dst.Spec.Scheduling)
	assert.Nil(t, dst.Spec.Env)
	assert.Nil(t, dst.Spec.Observability)
}

func TestAffinityConversion(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "affinity-test"},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust",
			WorkerCommandSeed: "locust",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
			Affinity: &LocustTestAffinity{
				NodeAffinity: &LocustTestNodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{
						"node-type": "high-cpu",
					},
				},
			},
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	require.NotNil(t, dst.Spec.Scheduling)
	require.NotNil(t, dst.Spec.Scheduling.Affinity)
	require.NotNil(t, dst.Spec.Scheduling.Affinity.NodeAffinity)

	// Convert back
	result := &LocustTest{}
	err = result.ConvertFrom(dst)
	require.NoError(t, err)

	require.NotNil(t, result.Spec.Affinity)
	require.NotNil(t, result.Spec.Affinity.NodeAffinity)
	assert.Equal(t, "high-cpu", result.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution["node-type"])
}

func TestTolerationsConversion(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "tolerations-test"},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust",
			WorkerCommandSeed: "locust",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
			Tolerations: []LocustTestToleration{
				{Key: "dedicated", Operator: "Equal", Value: "locust", Effect: "NoSchedule"},
				{Key: "gpu", Operator: "Exists", Effect: "NoExecute"},
			},
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	require.NotNil(t, dst.Spec.Scheduling)
	require.Len(t, dst.Spec.Scheduling.Tolerations, 2)
	assert.Equal(t, "dedicated", dst.Spec.Scheduling.Tolerations[0].Key)
	assert.Equal(t, corev1.TolerationOpEqual, dst.Spec.Scheduling.Tolerations[0].Operator)
	assert.Equal(t, "locust", dst.Spec.Scheduling.Tolerations[0].Value)
	assert.Equal(t, corev1.TaintEffectNoSchedule, dst.Spec.Scheduling.Tolerations[0].Effect)

	// Convert back
	result := &LocustTest{}
	err = result.ConvertFrom(dst)
	require.NoError(t, err)

	require.Len(t, result.Spec.Tolerations, 2)
	assert.Equal(t, "dedicated", result.Spec.Tolerations[0].Key)
	assert.Equal(t, "Equal", result.Spec.Tolerations[0].Operator)
	assert.Equal(t, "locust", result.Spec.Tolerations[0].Value)
	assert.Equal(t, "NoSchedule", result.Spec.Tolerations[0].Effect)
}

func TestImagePullSecretsConversion(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "secrets-test"},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust",
			WorkerCommandSeed: "locust",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
			ImagePullSecrets:  []string{"secret1", "secret2", "secret3"},
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	require.Len(t, dst.Spec.ImagePullSecrets, 3)
	assert.Equal(t, "secret1", dst.Spec.ImagePullSecrets[0].Name)
	assert.Equal(t, "secret2", dst.Spec.ImagePullSecrets[1].Name)
	assert.Equal(t, "secret3", dst.Spec.ImagePullSecrets[2].Name)

	// Convert back
	result := &LocustTest{}
	err = result.ConvertFrom(dst)
	require.NoError(t, err)

	require.Len(t, result.Spec.ImagePullSecrets, 3)
	assert.Equal(t, "secret1", result.Spec.ImagePullSecrets[0])
	assert.Equal(t, "secret2", result.Spec.ImagePullSecrets[1])
	assert.Equal(t, "secret3", result.Spec.ImagePullSecrets[2])
}

func TestConvertTo_EmptyOptionalFields(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "empty-optional"},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust",
			WorkerCommandSeed: "locust",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
			Labels:            nil,
			Annotations:       nil,
			Affinity:          nil,
			Tolerations:       nil,
			ImagePullSecrets:  nil,
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	assert.Empty(t, dst.Spec.Master.Labels)
	assert.Empty(t, dst.Spec.Master.Annotations)
	assert.Empty(t, dst.Spec.Worker.Labels)
	assert.Empty(t, dst.Spec.Worker.Annotations)
	assert.Nil(t, dst.Spec.Scheduling)
	assert.Nil(t, dst.Spec.ImagePullSecrets)
}

func TestConvertFrom_EmptyOptionalFields(t *testing.T) {
	src := &v2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "empty-optional"},
		Spec: v2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: v2.MasterSpec{
				Command: "locust",
			},
			Worker: v2.WorkerSpec{
				Command:  "locust",
				Replicas: 1,
			},
		},
	}

	dst := &LocustTest{}
	err := dst.ConvertFrom(src)
	require.NoError(t, err)

	assert.Nil(t, dst.Spec.Labels)
	assert.Nil(t, dst.Spec.Annotations)
	assert.Nil(t, dst.Spec.Affinity)
	assert.Empty(t, dst.Spec.Tolerations)
	assert.Empty(t, dst.Spec.ImagePullSecrets)
	assert.Empty(t, dst.Spec.ConfigMap)
	assert.Empty(t, dst.Spec.LibConfigMap)
}

func TestConvertFrom_V2OnlyFieldsLost(t *testing.T) {
	src := &v2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "v2-only-fields"},
		Spec: v2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: v2.MasterSpec{
				Command:   "locust",
				ExtraArgs: []string{"--headless"},
			},
			Worker: v2.WorkerSpec{
				Command:   "locust",
				Replicas:  3,
				ExtraArgs: []string{"--processes", "4"},
			},
			TestFiles: &v2.TestFilesConfig{
				ConfigMapRef: "tests",
				SrcMountPath: "/custom/path",
				LibMountPath: "/custom/lib",
			},
			Scheduling: &v2.SchedulingConfig{
				NodeSelector: map[string]string{"zone": "us-west"},
			},
			Env: &v2.EnvConfig{
				Variables: []corev1.EnvVar{
					{Name: "DEBUG", Value: "true"},
				},
			},
			Observability: &v2.ObservabilityConfig{
				OpenTelemetry: &v2.OpenTelemetryConfig{
					Enabled:  true,
					Endpoint: "otel:4317",
				},
			},
		},
	}

	dst := &LocustTest{}
	err := dst.ConvertFrom(src)
	require.NoError(t, err)

	// Verify base fields are preserved
	assert.Equal(t, "locust", dst.Spec.MasterCommandSeed)
	assert.Equal(t, "locust", dst.Spec.WorkerCommandSeed)
	assert.Equal(t, int32(3), dst.Spec.WorkerReplicas)
	assert.Equal(t, "tests", dst.Spec.ConfigMap)

	// v2-only fields are lost - we can't verify they're nil in v1
	// because v1 doesn't have these fields at all
	// This test mainly ensures conversion doesn't error on v2-only fields
}

func TestConvertTo_LabelsAndAnnotations(t *testing.T) {
	src := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "labels-annotations-test"},
		Spec: LocustTestSpec{
			MasterCommandSeed: "locust",
			WorkerCommandSeed: "locust",
			WorkerReplicas:    1,
			Image:             "locustio/locust:latest",
			Labels: &PodLabels{
				Master: map[string]string{"env": "test", "component": "master"},
				Worker: map[string]string{"env": "test", "component": "worker"},
			},
			Annotations: &PodAnnotations{
				Master: map[string]string{"note": "master-annotation"},
				Worker: map[string]string{"note": "worker-annotation"},
			},
		},
	}

	dst := &v2.LocustTest{}
	err := src.ConvertTo(dst)
	require.NoError(t, err)

	assert.Equal(t, "test", dst.Spec.Master.Labels["env"])
	assert.Equal(t, "master", dst.Spec.Master.Labels["component"])
	assert.Equal(t, "master-annotation", dst.Spec.Master.Annotations["note"])

	assert.Equal(t, "test", dst.Spec.Worker.Labels["env"])
	assert.Equal(t, "worker", dst.Spec.Worker.Labels["component"])
	assert.Equal(t, "worker-annotation", dst.Spec.Worker.Annotations["note"])
}

func TestConvertFrom_LabelsAndAnnotations(t *testing.T) {
	src := &v2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "labels-annotations-test"},
		Spec: v2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: v2.MasterSpec{
				Command:     "locust",
				Labels:      map[string]string{"role": "master"},
				Annotations: map[string]string{"desc": "master-desc"},
			},
			Worker: v2.WorkerSpec{
				Command:     "locust",
				Replicas:    1,
				Labels:      map[string]string{"role": "worker"},
				Annotations: map[string]string{"desc": "worker-desc"},
			},
		},
	}

	dst := &LocustTest{}
	err := dst.ConvertFrom(src)
	require.NoError(t, err)

	require.NotNil(t, dst.Spec.Labels)
	assert.Equal(t, "master", dst.Spec.Labels.Master["role"])
	assert.Equal(t, "worker", dst.Spec.Labels.Worker["role"])

	require.NotNil(t, dst.Spec.Annotations)
	assert.Equal(t, "master-desc", dst.Spec.Annotations.Master["desc"])
	assert.Equal(t, "worker-desc", dst.Spec.Annotations.Worker["desc"])
}

func TestNilAffinityConversion(t *testing.T) {
	// Test nil affinity in v1
	result := convertAffinityToV2(nil)
	assert.Nil(t, result)

	// Test nil affinity in v2
	result2 := convertAffinityToV1(nil)
	assert.Nil(t, result2)

	// Test affinity with nil NodeAffinity
	result3 := convertAffinityToV2(&LocustTestAffinity{NodeAffinity: nil})
	assert.Nil(t, result3)

	// Test affinity with empty requirements
	result4 := convertAffinityToV2(&LocustTestAffinity{
		NodeAffinity: &LocustTestNodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{},
		},
	})
	assert.Nil(t, result4)
}

func TestEmptyTolerationsConversion(t *testing.T) {
	// Test empty tolerations in v1
	result := convertTolerationsToV2(nil)
	assert.Nil(t, result)

	result2 := convertTolerationsToV2([]LocustTestToleration{})
	assert.Nil(t, result2)

	// Test empty tolerations in v2
	result3 := convertTolerationsToV1(nil)
	assert.Nil(t, result3)

	result4 := convertTolerationsToV1([]corev1.Toleration{})
	assert.Nil(t, result4)
}

func TestEmptyImagePullSecretsConversion(t *testing.T) {
	// Test empty secrets in v1
	result := convertImagePullSecretsToV2(nil)
	assert.Nil(t, result)

	result2 := convertImagePullSecretsToV2([]string{})
	assert.Nil(t, result2)

	// Test empty secrets in v2
	result3 := convertImagePullSecretsToV1(nil)
	assert.Nil(t, result3)

	result4 := convertImagePullSecretsToV1([]corev1.LocalObjectReference{})
	assert.Nil(t, result4)
}

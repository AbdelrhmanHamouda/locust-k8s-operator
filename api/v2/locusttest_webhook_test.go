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

package v2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPathConflicts_ExactMatch(t *testing.T) {
	assert.True(t, PathConflicts("/lotest/src", "/lotest/src"))
	assert.True(t, PathConflicts("/opt/locust/lib", "/opt/locust/lib"))
}

func TestPathConflicts_Subpath(t *testing.T) {
	// /foo conflicts with /foo/bar because /foo is a prefix
	assert.True(t, PathConflicts("/lotest/src", "/lotest/src/secrets"))
	assert.True(t, PathConflicts("/lotest/src/secrets", "/lotest/src"))

	// Deeper nesting
	assert.True(t, PathConflicts("/opt/locust/lib", "/opt/locust/lib/utils"))
	assert.True(t, PathConflicts("/opt/locust/lib/utils", "/opt/locust/lib"))
}

func TestPathConflicts_NoConflict(t *testing.T) {
	// Completely different paths
	assert.False(t, PathConflicts("/lotest/src", "/etc/certs"))
	assert.False(t, PathConflicts("/opt/locust/lib", "/var/secrets"))

	// Similar prefix but not a subpath
	assert.False(t, PathConflicts("/lotest/src", "/lotest/src2"))
	assert.False(t, PathConflicts("/lotest/src2", "/lotest/src"))
}

func TestPathConflicts_TrailingSlash(t *testing.T) {
	// Trailing slashes should be normalized
	assert.True(t, PathConflicts("/lotest/src/", "/lotest/src"))
	assert.True(t, PathConflicts("/lotest/src", "/lotest/src/"))
	assert.True(t, PathConflicts("/lotest/src/", "/lotest/src/"))
}

func TestValidateSecretMounts_NilEnv(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: nil,
		},
	}

	err := validateSecretMounts(lt)
	assert.NoError(t, err)
}

func TestValidateSecretMounts_EmptyMounts(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{},
			},
		},
	}

	err := validateSecretMounts(lt)
	assert.NoError(t, err)
}

func TestValidateSecretMounts_ValidPath(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "tls-certs", MountPath: "/etc/locust/certs"},
				},
			},
		},
	}

	err := validateSecretMounts(lt)
	assert.NoError(t, err)
}

func TestValidateSecretMounts_ConflictDefault(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src"},
				},
			},
		},
	}

	err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
	assert.Contains(t, err.Error(), "/lotest/src")
}

func TestValidateSecretMounts_ConflictLib(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/opt/locust/lib"},
				},
			},
		},
	}

	err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
	assert.Contains(t, err.Error(), "/opt/locust/lib")
}

func TestValidateSecretMounts_ConflictSubpath(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src/secrets"},
				},
			},
		},
	}

	err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
}

func TestValidateSecretMounts_CustomTestFilesPath(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef: "my-scripts",
				SrcMountPath: "/custom/src",
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "secret", MountPath: "/custom/src/secrets"},
				},
			},
		},
	}

	err := validateSecretMounts(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "/custom/src")
}

func TestValidateSecretMounts_CustomPathAllowsDefault(t *testing.T) {
	// When using custom paths, the default paths should be allowed
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef: "my-scripts",
				SrcMountPath: "/custom/src",
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					// This would conflict with default but we're using custom
					{Name: "secret", MountPath: "/lotest/src"},
				},
			},
		},
	}

	err := validateSecretMounts(lt)
	// Should pass because we're using custom path, not default
	assert.NoError(t, err)
}

func TestGetReservedPaths_NoTestFiles(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, DefaultSrcMountPath)
	assert.Contains(t, paths, DefaultLibMountPath)
}

func TestGetReservedPaths_WithConfigMapRef(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef: "my-scripts",
			},
		},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, DefaultSrcMountPath)
	assert.Len(t, paths, 1)
}

func TestGetReservedPaths_WithBothConfigMaps(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef:    "my-scripts",
				LibConfigMapRef: "my-lib",
			},
		},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, DefaultSrcMountPath)
	assert.Contains(t, paths, DefaultLibMountPath)
	assert.Len(t, paths, 2)
}

func TestGetReservedPaths_CustomPaths(t *testing.T) {
	lt := &LocustTest{
		Spec: LocustTestSpec{
			TestFiles: &TestFilesConfig{
				ConfigMapRef:    "my-scripts",
				LibConfigMapRef: "my-lib",
				SrcMountPath:    "/custom/src",
				LibMountPath:    "/custom/lib",
			},
		},
	}

	paths := getReservedPaths(lt)
	assert.Contains(t, paths, "/custom/src")
	assert.Contains(t, paths, "/custom/lib")
	assert.NotContains(t, paths, DefaultSrcMountPath)
	assert.NotContains(t, paths, DefaultLibMountPath)
}

func TestValidateCreate(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "valid-secret", MountPath: "/etc/certs"},
				},
			},
		},
	}

	warnings, err := validator.ValidateCreate(context.Background(), lt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_Invalid(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src"},
				},
			},
		},
	}

	warnings, err := validator.ValidateCreate(context.Background(), lt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
	assert.Nil(t, warnings)
}

func TestValidateUpdate(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	oldLt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
		},
	}

	newLt := oldLt.DeepCopy()
	newLt.Spec.Env = &EnvConfig{
		SecretMounts: []SecretMount{
			{Name: "valid-secret", MountPath: "/etc/certs"},
		},
	}

	warnings, err := validator.ValidateUpdate(context.Background(), oldLt, newLt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateDelete(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}

	warnings, err := validator.ValidateDelete(context.Background(), lt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_WrongType(t *testing.T) {
	validator := &LocustTestCustomValidator{}

	// Pass wrong type
	warnings, err := validator.ValidateCreate(context.Background(), &LocustTestList{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected LocustTest")
	assert.Nil(t, warnings)
}

// ============================================
// Volume Validation Tests
// ============================================

func TestValidateVolumes_Empty(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Volumes:      nil,
			VolumeMounts: nil,
		},
	}

	err := validateVolumes(lt)
	assert.NoError(t, err)
}

func TestValidateVolumes_ValidConfig(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "test-results"},
				{Name: "shared-data"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "test-results", MountPath: "/results"}, Target: "master"},
				{VolumeMount: corev1.VolumeMount{Name: "shared-data", MountPath: "/shared"}, Target: "both"},
			},
		},
	}

	err := validateVolumes(lt)
	assert.NoError(t, err)
}

func TestValidateVolumeName_Valid(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	assert.NoError(t, validateVolumeName(lt, "test-results"))
	assert.NoError(t, validateVolumeName(lt, "shared-data"))
	assert.NoError(t, validateVolumeName(lt, "custom-volume"))
}

func TestValidateVolumeName_SecretPrefix(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	err := validateVolumeName(lt, "secret-custom")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uses reserved prefix")
	assert.Contains(t, err.Error(), "secret-")
}

func TestValidateVolumeName_LibVolume(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	err := validateVolumeName(lt, "locust-lib")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is reserved by the operator")
}

func TestValidateVolumeName_MasterConflict(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	err := validateVolumeName(lt, "my-test-master")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with operator-generated name")
}

func TestValidateVolumeName_WorkerConflict(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	err := validateVolumeName(lt, "my-test-worker")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with operator-generated name")
}

func TestValidateVolumes_PathConflict(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "bad-volume"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "bad-volume", MountPath: "/lotest/src"}, Target: "both"},
			},
		},
	}

	err := validateVolumes(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with reserved path")
}

func TestValidateVolumes_UndefinedMount(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "defined-volume"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "undefined-volume", MountPath: "/data"}, Target: "both"},
			},
		},
	}

	err := validateVolumes(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "references undefined volume")
}

func TestValidateMountReferences_Valid(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "vol1"},
				{Name: "vol2"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "vol1", MountPath: "/data1"}},
				{VolumeMount: corev1.VolumeMount{Name: "vol2", MountPath: "/data2"}},
			},
		},
	}

	err := validateMountReferences(lt)
	assert.NoError(t, err)
}

func TestValidateMountReferences_Invalid(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "vol1"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "vol1", MountPath: "/data1"}},
				{VolumeMount: corev1.VolumeMount{Name: "missing", MountPath: "/data2"}},
			},
		},
	}

	err := validateMountReferences(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestValidateLocustTest_CombinedValidation(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "valid-secret", MountPath: "/etc/certs"},
				},
			},
			Volumes: []corev1.Volume{
				{Name: "test-results"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "test-results", MountPath: "/results"}, Target: "master"},
			},
		},
	}

	warnings, err := validateLocustTest(lt)
	assert.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateLocustTest_SecretMountFailsFirst(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Env: &EnvConfig{
				SecretMounts: []SecretMount{
					{Name: "bad-secret", MountPath: "/lotest/src"},
				},
			},
			Volumes: []corev1.Volume{
				{Name: "secret-bad"}, // Also invalid but secret mount fails first
			},
		},
	}

	_, err := validateLocustTest(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secretMount path")
}

func TestValidateCreate_WithVolumes(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Volumes: []corev1.Volume{
				{Name: "test-results"},
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "test-results", MountPath: "/results"}, Target: "master"},
			},
		},
	}

	warnings, err := validator.ValidateCreate(context.Background(), lt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_WithInvalidVolumeName(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Volumes: []corev1.Volume{
				{Name: "secret-custom"}, // Invalid: uses reserved prefix
			},
			VolumeMounts: []TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "secret-custom", MountPath: "/custom"}},
			},
		},
	}

	_, err := validator.ValidateCreate(context.Background(), lt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uses reserved prefix")
}

// ============================================
// OTel Validation Tests
// ============================================

func TestValidateOTelConfig_NoObservability(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: nil,
		},
	}

	err := validateOTelConfig(lt)
	assert.NoError(t, err)
}

func TestValidateOTelConfig_NoOpenTelemetry(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: &ObservabilityConfig{
				OpenTelemetry: nil,
			},
		},
	}

	err := validateOTelConfig(lt)
	assert.NoError(t, err)
}

func TestValidateOTelConfig_Disabled(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled: false,
				},
			},
		},
	}

	err := validateOTelConfig(lt)
	assert.NoError(t, err)
}

func TestValidateOTelConfig_EnabledWithEndpoint(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled:  true,
					Endpoint: "otel-collector:4317",
				},
			},
		},
	}

	err := validateOTelConfig(lt)
	assert.NoError(t, err)
}

func TestValidateOTelConfig_EnabledNoEndpoint(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled:  true,
					Endpoint: "", // Missing endpoint
				},
			},
		},
	}

	err := validateOTelConfig(lt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required when OpenTelemetry is enabled")
}

func TestValidateOTelConfig_ValidProtocolGRPC(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled:  true,
					Endpoint: "otel-collector:4317",
					Protocol: "grpc",
				},
			},
		},
	}

	err := validateOTelConfig(lt)
	assert.NoError(t, err)
}

func TestValidateOTelConfig_ValidProtocolHTTP(t *testing.T) {
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled:  true,
					Endpoint: "otel-collector:4318",
					Protocol: "http/protobuf",
				},
			},
		},
	}

	err := validateOTelConfig(lt)
	assert.NoError(t, err)
}

func TestValidateCreate_WithOTelEnabled(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.32.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled:  true,
					Endpoint: "otel-collector:4317",
				},
			},
		},
	}

	warnings, err := validator.ValidateCreate(context.Background(), lt)
	require.NoError(t, err)
	assert.Nil(t, warnings)
}

func TestValidateCreate_WithOTelEnabledNoEndpoint(t *testing.T) {
	validator := &LocustTestCustomValidator{}
	lt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.32.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
			Observability: &ObservabilityConfig{
				OpenTelemetry: &OpenTelemetryConfig{
					Enabled: true,
					// No endpoint - should fail
				},
			},
		},
	}

	_, err := validator.ValidateCreate(context.Background(), lt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")
}

// ============================================
// Update Validation Tests
// ============================================

func TestValidateUpdate_Invalid(t *testing.T) {
	validator := &LocustTestCustomValidator{}

	oldLt := &LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 1,
			},
		},
	}

	t.Run("InvalidSecretMountPath", func(t *testing.T) {
		newLt := oldLt.DeepCopy()
		newLt.Spec.Env = &EnvConfig{
			SecretMounts: []SecretMount{
				{Name: "bad-secret", MountPath: "/lotest/src"},
			},
		}

		_, err := validator.ValidateUpdate(context.Background(), oldLt, newLt)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "conflicts with reserved path")
	})

	t.Run("InvalidVolumeName", func(t *testing.T) {
		newLt := oldLt.DeepCopy()
		newLt.Spec.Volumes = []corev1.Volume{
			{Name: "secret-my-volume"}, // Uses reserved prefix
		}
		newLt.Spec.VolumeMounts = []TargetedVolumeMount{
			{VolumeMount: corev1.VolumeMount{Name: "secret-my-volume", MountPath: "/custom"}},
		}

		_, err := validator.ValidateUpdate(context.Background(), oldLt, newLt)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "uses reserved prefix")
	})

	t.Run("OTelEnabledNoEndpoint", func(t *testing.T) {
		newLt := oldLt.DeepCopy()
		newLt.Spec.Observability = &ObservabilityConfig{
			OpenTelemetry: &OpenTelemetryConfig{
				Enabled:  true,
				Endpoint: "", // Missing endpoint
			},
		}

		_, err := validator.ValidateUpdate(context.Background(), oldLt, newLt)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint is required")
	})
}

// ============================================
// Boundary Tests
// ============================================

func TestValidateCreate_LongCRName(t *testing.T) {
	validator := &LocustTestCustomValidator{}

	t.Run("NameTooLong", func(t *testing.T) {
		// Create a name that would exceed 63 chars when "-worker" is added
		// 57 chars + "-worker" (7 chars) = 64 chars > 63 limit
		longName := "a123456789-123456789-123456789-123456789-123456789-123456"
		require.Equal(t, 57, len(longName), "Test name should be 57 chars")

		lt := &LocustTest{
			ObjectMeta: metav1.ObjectMeta{Name: longName, Namespace: "default"},
			Spec: LocustTestSpec{
				Image: "locustio/locust:2.20.0",
				Master: MasterSpec{
					Command: "locust -f /lotest/src/locustfile.py",
				},
				Worker: WorkerSpec{
					Command:  "locust -f /lotest/src/locustfile.py",
					Replicas: 1,
				},
			},
		}

		_, err := validator.ValidateCreate(context.Background(), lt)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
		assert.Contains(t, err.Error(), "63 characters")
	})

	t.Run("NameAtLimit", func(t *testing.T) {
		// 56 chars + "-worker" (7 chars) = 63 chars (exactly at limit)
		maxName := "a123456789-123456789-123456789-123456789-123456789-12345"
		require.Equal(t, 56, len(maxName), "Test name should be 56 chars")

		lt := &LocustTest{
			ObjectMeta: metav1.ObjectMeta{Name: maxName, Namespace: "default"},
			Spec: LocustTestSpec{
				Image: "locustio/locust:2.20.0",
				Master: MasterSpec{
					Command: "locust -f /lotest/src/locustfile.py",
				},
				Worker: WorkerSpec{
					Command:  "locust -f /lotest/src/locustfile.py",
					Replicas: 1,
				},
			},
		}

		_, err := validator.ValidateCreate(context.Background(), lt)
		require.NoError(t, err)
	})
}

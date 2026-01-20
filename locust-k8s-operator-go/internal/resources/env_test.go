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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildEnvFrom_NilEnvConfig(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: nil,
		},
	}

	result := BuildEnvFrom(lt)
	assert.Nil(t, result)
}

func TestBuildEnvFrom_EmptyEnvConfig(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{},
		},
	}

	result := BuildEnvFrom(lt)
	assert.Nil(t, result)
}

func TestBuildEnvFrom_ConfigMapRefs(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				ConfigMapRefs: []locustv2.ConfigMapEnvSource{
					{Name: "app-config"},
				},
			},
		},
	}

	result := BuildEnvFrom(lt)

	assert.Len(t, result, 1)
	assert.NotNil(t, result[0].ConfigMapRef)
	assert.Nil(t, result[0].SecretRef)
	assert.Equal(t, "app-config", result[0].ConfigMapRef.Name)
	assert.Empty(t, result[0].Prefix)
}

func TestBuildEnvFrom_ConfigMapRefs_WithPrefix(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				ConfigMapRefs: []locustv2.ConfigMapEnvSource{
					{Name: "app-config", Prefix: "APP_"},
				},
			},
		},
	}

	result := BuildEnvFrom(lt)

	assert.Len(t, result, 1)
	assert.Equal(t, "APP_", result[0].Prefix)
	assert.Equal(t, "app-config", result[0].ConfigMapRef.Name)
}

func TestBuildEnvFrom_SecretRefs(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretRefs: []locustv2.SecretEnvSource{
					{Name: "api-credentials"},
				},
			},
		},
	}

	result := BuildEnvFrom(lt)

	assert.Len(t, result, 1)
	assert.Nil(t, result[0].ConfigMapRef)
	assert.NotNil(t, result[0].SecretRef)
	assert.Equal(t, "api-credentials", result[0].SecretRef.Name)
	assert.Empty(t, result[0].Prefix)
}

func TestBuildEnvFrom_SecretRefs_WithPrefix(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretRefs: []locustv2.SecretEnvSource{
					{Name: "api-credentials", Prefix: "SECRET_"},
				},
			},
		},
	}

	result := BuildEnvFrom(lt)

	assert.Len(t, result, 1)
	assert.Equal(t, "SECRET_", result[0].Prefix)
	assert.Equal(t, "api-credentials", result[0].SecretRef.Name)
}

func TestBuildEnvFrom_Multiple(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				ConfigMapRefs: []locustv2.ConfigMapEnvSource{
					{Name: "config1", Prefix: "CFG1_"},
					{Name: "config2", Prefix: "CFG2_"},
				},
				SecretRefs: []locustv2.SecretEnvSource{
					{Name: "secret1"},
					{Name: "secret2", Prefix: "SEC_"},
				},
			},
		},
	}

	result := BuildEnvFrom(lt)

	assert.Len(t, result, 4)

	// ConfigMaps come first
	assert.Equal(t, "config1", result[0].ConfigMapRef.Name)
	assert.Equal(t, "CFG1_", result[0].Prefix)
	assert.Equal(t, "config2", result[1].ConfigMapRef.Name)
	assert.Equal(t, "CFG2_", result[1].Prefix)

	// Secrets come after
	assert.Equal(t, "secret1", result[2].SecretRef.Name)
	assert.Empty(t, result[2].Prefix)
	assert.Equal(t, "secret2", result[3].SecretRef.Name)
	assert.Equal(t, "SEC_", result[3].Prefix)
}

func TestBuildUserEnvVars_Nil(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: nil,
		},
	}

	result := BuildUserEnvVars(lt)
	assert.Nil(t, result)
}

func TestBuildUserEnvVars_Empty(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				Variables: []corev1.EnvVar{},
			},
		},
	}

	result := BuildUserEnvVars(lt)
	assert.Nil(t, result)
}

func TestBuildUserEnvVars_DirectValues(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				Variables: []corev1.EnvVar{
					{Name: "TARGET_HOST", Value: "https://example.com"},
					{Name: "LOG_LEVEL", Value: "DEBUG"},
				},
			},
		},
	}

	result := BuildUserEnvVars(lt)

	assert.Len(t, result, 2)
	assert.Equal(t, "TARGET_HOST", result[0].Name)
	assert.Equal(t, "https://example.com", result[0].Value)
	assert.Equal(t, "LOG_LEVEL", result[1].Name)
	assert.Equal(t, "DEBUG", result[1].Value)
}

func TestBuildUserEnvVars_ValueFrom(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				Variables: []corev1.EnvVar{
					{
						Name: "API_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "api-secret",
								},
								Key: "key",
							},
						},
					},
				},
			},
		},
	}

	result := BuildUserEnvVars(lt)

	assert.Len(t, result, 1)
	assert.Equal(t, "API_KEY", result[0].Name)
	assert.NotNil(t, result[0].ValueFrom)
	assert.NotNil(t, result[0].ValueFrom.SecretKeyRef)
	assert.Equal(t, "api-secret", result[0].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, "key", result[0].ValueFrom.SecretKeyRef.Key)
}

func TestBuildUserEnvVars_ReturnsCopy(t *testing.T) {
	original := []corev1.EnvVar{
		{Name: "KEY", Value: "value"},
	}
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				Variables: original,
			},
		},
	}

	result := BuildUserEnvVars(lt)

	// Modify the result
	result[0].Value = "modified"

	// Original should be unchanged
	assert.Equal(t, "value", original[0].Value)
}

func TestBuildEnvVars_OnlyKafka(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: nil,
		},
	}
	cfg := &config.OperatorConfig{
		KafkaBootstrapServers: "kafka:9092",
		KafkaSecurityEnabled:  false,
	}

	result := BuildEnvVars(lt, cfg)

	// Should have 7 Kafka env vars
	assert.Len(t, result, 7)
	assert.Equal(t, EnvKafkaBootstrapServers, result[0].Name)
	assert.Equal(t, "kafka:9092", result[0].Value)
}

func TestBuildEnvVars_Combined(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				Variables: []corev1.EnvVar{
					{Name: "USER_VAR", Value: "user-value"},
				},
			},
		},
	}
	cfg := &config.OperatorConfig{
		KafkaBootstrapServers: "kafka:9092",
	}

	result := BuildEnvVars(lt, cfg)

	// 7 Kafka vars + 1 user var
	assert.Len(t, result, 8)

	// Kafka vars come first
	assert.Equal(t, EnvKafkaBootstrapServers, result[0].Name)

	// User var comes last
	assert.Equal(t, "USER_VAR", result[7].Name)
	assert.Equal(t, "user-value", result[7].Value)
}

func TestBuildSecretVolumes_Nil(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: nil,
		},
	}

	result := BuildSecretVolumes(lt)
	assert.Nil(t, result)
}

func TestBuildSecretVolumes_Empty(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretMounts: []locustv2.SecretMount{},
			},
		},
	}

	result := BuildSecretVolumes(lt)
	assert.Nil(t, result)
}

func TestBuildSecretVolumes_Single(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretMounts: []locustv2.SecretMount{
					{Name: "tls-certs", MountPath: "/etc/certs"},
				},
			},
		},
	}

	result := BuildSecretVolumes(lt)

	assert.Len(t, result, 1)
	assert.Equal(t, "secret-tls-certs", result[0].Name)
	assert.NotNil(t, result[0].Secret)
	assert.Equal(t, "tls-certs", result[0].Secret.SecretName)
}

func TestBuildSecretVolumes_Multiple(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretMounts: []locustv2.SecretMount{
					{Name: "tls-certs", MountPath: "/etc/certs"},
					{Name: "ssh-keys", MountPath: "/root/.ssh"},
				},
			},
		},
	}

	result := BuildSecretVolumes(lt)

	assert.Len(t, result, 2)
	assert.Equal(t, "secret-tls-certs", result[0].Name)
	assert.Equal(t, "tls-certs", result[0].Secret.SecretName)
	assert.Equal(t, "secret-ssh-keys", result[1].Name)
	assert.Equal(t, "ssh-keys", result[1].Secret.SecretName)
}

func TestBuildSecretVolumeMounts_Nil(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: nil,
		},
	}

	result := BuildSecretVolumeMounts(lt)
	assert.Nil(t, result)
}

func TestBuildSecretVolumeMounts_Empty(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretMounts: []locustv2.SecretMount{},
			},
		},
	}

	result := BuildSecretVolumeMounts(lt)
	assert.Nil(t, result)
}

func TestBuildSecretVolumeMounts_Single(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretMounts: []locustv2.SecretMount{
					{Name: "tls-certs", MountPath: "/etc/certs", ReadOnly: true},
				},
			},
		},
	}

	result := BuildSecretVolumeMounts(lt)

	assert.Len(t, result, 1)
	assert.Equal(t, "secret-tls-certs", result[0].Name)
	assert.Equal(t, "/etc/certs", result[0].MountPath)
	assert.True(t, result[0].ReadOnly)
}

func TestBuildSecretVolumeMounts_ReadOnly(t *testing.T) {
	lt := &locustv2.LocustTest{
		Spec: locustv2.LocustTestSpec{
			Env: &locustv2.EnvConfig{
				SecretMounts: []locustv2.SecretMount{
					{Name: "secret1", MountPath: "/path1", ReadOnly: true},
					{Name: "secret2", MountPath: "/path2", ReadOnly: false},
				},
			},
		},
	}

	result := BuildSecretVolumeMounts(lt)

	assert.Len(t, result, 2)
	assert.True(t, result[0].ReadOnly)
	assert.False(t, result[1].ReadOnly)
}

func TestSecretVolumeName(t *testing.T) {
	tests := []struct {
		secretName string
		expected   string
	}{
		{"tls-certs", "secret-tls-certs"},
		{"api-keys", "secret-api-keys"},
		{"my.secret", "secret-my.secret"},
	}

	for _, tt := range tests {
		t.Run(tt.secretName, func(t *testing.T) {
			result := SecretVolumeName(tt.secretName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildEnvFrom_IntegrationWithFullSpec(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-load",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:2.20.0",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 5,
			},
			Env: &locustv2.EnvConfig{
				ConfigMapRefs: []locustv2.ConfigMapEnvSource{
					{Name: "app-config", Prefix: "APP_"},
				},
				SecretRefs: []locustv2.SecretEnvSource{
					{Name: "api-credentials"},
				},
				Variables: []corev1.EnvVar{
					{Name: "TARGET_HOST", Value: "https://api.example.com"},
				},
				SecretMounts: []locustv2.SecretMount{
					{Name: "tls-certs", MountPath: "/etc/locust/certs", ReadOnly: true},
				},
			},
		},
	}

	// Test all builders work together
	envFrom := BuildEnvFrom(lt)
	assert.Len(t, envFrom, 2)

	userVars := BuildUserEnvVars(lt)
	assert.Len(t, userVars, 1)

	volumes := BuildSecretVolumes(lt)
	assert.Len(t, volumes, 1)

	mounts := BuildSecretVolumeMounts(lt)
	assert.Len(t, mounts, 1)
}

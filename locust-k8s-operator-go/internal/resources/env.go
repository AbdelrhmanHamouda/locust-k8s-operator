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
	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
	corev1 "k8s.io/api/core/v1"
)

// BuildEnvFrom creates EnvFromSource entries from ConfigMap and Secret refs.
// Returns envFrom slice for container spec.
func BuildEnvFrom(lt *locustv2.LocustTest) []corev1.EnvFromSource {
	if lt.Spec.Env == nil {
		return nil
	}

	envFrom := make([]corev1.EnvFromSource, 0, len(lt.Spec.Env.ConfigMapRefs)+len(lt.Spec.Env.SecretRefs))

	// Process ConfigMapRefs
	for _, cmRef := range lt.Spec.Env.ConfigMapRefs {
		envFrom = append(envFrom, corev1.EnvFromSource{
			Prefix: cmRef.Prefix,
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cmRef.Name,
				},
			},
		})
	}

	// Process SecretRefs
	for _, secretRef := range lt.Spec.Env.SecretRefs {
		envFrom = append(envFrom, corev1.EnvFromSource{
			Prefix: secretRef.Prefix,
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretRef.Name,
				},
			},
		})
	}

	return envFrom
}

// BuildUserEnvVars creates EnvVar entries from the variables list.
// These are appended to the existing Kafka env vars.
func BuildUserEnvVars(lt *locustv2.LocustTest) []corev1.EnvVar {
	if lt.Spec.Env == nil || len(lt.Spec.Env.Variables) == 0 {
		return nil
	}

	// Return a copy to avoid mutating the original
	result := make([]corev1.EnvVar, len(lt.Spec.Env.Variables))
	copy(result, lt.Spec.Env.Variables)
	return result
}

// BuildEnvVars combines Kafka env vars with user-defined env vars.
func BuildEnvVars(lt *locustv2.LocustTest, cfg *config.OperatorConfig) []corev1.EnvVar {
	// Start with Kafka env vars (existing behavior)
	envVars := BuildKafkaEnvVars(cfg)

	// Append user-defined variables
	userVars := BuildUserEnvVars(lt)
	if len(userVars) > 0 {
		envVars = append(envVars, userVars...)
	}

	return envVars
}

// BuildSecretVolumes creates Volume entries for secret mounts.
func BuildSecretVolumes(lt *locustv2.LocustTest) []corev1.Volume {
	if lt.Spec.Env == nil || len(lt.Spec.Env.SecretMounts) == 0 {
		return nil
	}

	volumes := make([]corev1.Volume, 0, len(lt.Spec.Env.SecretMounts))
	for _, sm := range lt.Spec.Env.SecretMounts {
		volumes = append(volumes, corev1.Volume{
			Name: SecretVolumeName(sm.Name),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: sm.Name,
				},
			},
		})
	}
	return volumes
}

// BuildSecretVolumeMounts creates VolumeMount entries for secret mounts.
func BuildSecretVolumeMounts(lt *locustv2.LocustTest) []corev1.VolumeMount {
	if lt.Spec.Env == nil || len(lt.Spec.Env.SecretMounts) == 0 {
		return nil
	}

	mounts := make([]corev1.VolumeMount, 0, len(lt.Spec.Env.SecretMounts))
	for _, sm := range lt.Spec.Env.SecretMounts {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      SecretVolumeName(sm.Name),
			MountPath: sm.MountPath,
			ReadOnly:  sm.ReadOnly,
		})
	}
	return mounts
}

// SecretVolumeName generates a unique volume name for a secret mount.
func SecretVolumeName(secretName string) string {
	return "secret-" + secretName
}

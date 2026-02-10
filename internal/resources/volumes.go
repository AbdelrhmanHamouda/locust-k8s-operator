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
	corev1 "k8s.io/api/core/v1"
)

// Target constants for volume mount filtering.
const (
	TargetMaster = "master"
	TargetWorker = "worker"
	TargetBoth   = "both"
)

// BuildUserVolumes returns user-defined volumes filtered for the given mode.
// Only volumes that have at least one mount targeting this mode are included.
func BuildUserVolumes(lt *locustv2.LocustTest, mode OperationalMode) []corev1.Volume {
	if len(lt.Spec.Volumes) == 0 {
		return nil
	}

	var volumes []corev1.Volume
	for _, vol := range lt.Spec.Volumes {
		if shouldIncludeVolume(vol.Name, lt.Spec.VolumeMounts, mode) {
			volumes = append(volumes, vol)
		}
	}
	return volumes
}

// BuildUserVolumeMounts returns user-defined volume mounts filtered for the given mode.
func BuildUserVolumeMounts(lt *locustv2.LocustTest, mode OperationalMode) []corev1.VolumeMount {
	if len(lt.Spec.VolumeMounts) == 0 {
		return nil
	}

	var mounts []corev1.VolumeMount
	for _, tvm := range lt.Spec.VolumeMounts {
		if shouldApplyMount(tvm, mode) {
			// Convert TargetedVolumeMount to VolumeMount (strip Target field)
			mounts = append(mounts, tvm.VolumeMount)
		}
	}
	return mounts
}

// shouldApplyMount checks if a mount applies to the given operational mode.
func shouldApplyMount(mount locustv2.TargetedVolumeMount, mode OperationalMode) bool {
	target := mount.Target
	if target == "" {
		target = TargetBoth
	}

	switch target {
	case TargetBoth:
		return true
	case TargetMaster:
		return mode == Master
	case TargetWorker:
		return mode == Worker
	default:
		return false
	}
}

// shouldIncludeVolume checks if a volume has any mounts for the given mode.
func shouldIncludeVolume(volumeName string, mounts []locustv2.TargetedVolumeMount, mode OperationalMode) bool {
	for _, mount := range mounts {
		if mount.Name == volumeName && shouldApplyMount(mount, mode) {
			return true
		}
	}
	return false
}

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
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildUserVolumes_Empty(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			Volumes: nil,
		},
	}

	result := BuildUserVolumes(lt, Master)
	assert.Nil(t, result)

	result = BuildUserVolumes(lt, Worker)
	assert.Nil(t, result)
}

func TestBuildUserVolumes_AllTargetBoth(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "vol1", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "vol2", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "vol1", MountPath: "/data1"}, Target: "both"},
				{VolumeMount: corev1.VolumeMount{Name: "vol2", MountPath: "/data2"}, Target: "both"},
			},
		},
	}

	// Both volumes should be included for master
	result := BuildUserVolumes(lt, Master)
	assert.Len(t, result, 2)
	assert.Equal(t, "vol1", result[0].Name)
	assert.Equal(t, "vol2", result[1].Name)

	// Both volumes should be included for worker
	result = BuildUserVolumes(lt, Worker)
	assert.Len(t, result, 2)
}

func TestBuildUserVolumes_MasterOnly(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "results", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "results", MountPath: "/results"}, Target: "master"},
			},
		},
	}

	// Volume should be included for master
	result := BuildUserVolumes(lt, Master)
	assert.Len(t, result, 1)
	assert.Equal(t, "results", result[0].Name)

	// Volume should NOT be included for worker
	result = BuildUserVolumes(lt, Worker)
	assert.Nil(t, result)
}

func TestBuildUserVolumes_WorkerOnly(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "certs", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "certs", MountPath: "/certs"}, Target: "worker"},
			},
		},
	}

	// Volume should NOT be included for master
	result := BuildUserVolumes(lt, Master)
	assert.Nil(t, result)

	// Volume should be included for worker
	result = BuildUserVolumes(lt, Worker)
	assert.Len(t, result, 1)
	assert.Equal(t, "certs", result[0].Name)
}

func TestBuildUserVolumes_Mixed(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "results", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "shared", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "certs", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "results", MountPath: "/results"}, Target: "master"},
				{VolumeMount: corev1.VolumeMount{Name: "shared", MountPath: "/shared"}, Target: "both"},
				{VolumeMount: corev1.VolumeMount{Name: "certs", MountPath: "/certs"}, Target: "worker"},
			},
		},
	}

	// Master should get results and shared
	result := BuildUserVolumes(lt, Master)
	assert.Len(t, result, 2)
	names := []string{result[0].Name, result[1].Name}
	assert.Contains(t, names, "results")
	assert.Contains(t, names, "shared")

	// Worker should get shared and certs
	result = BuildUserVolumes(lt, Worker)
	assert.Len(t, result, 2)
	names = []string{result[0].Name, result[1].Name}
	assert.Contains(t, names, "shared")
	assert.Contains(t, names, "certs")
}

func TestBuildUserVolumes_VolumeWithNoMatchingMount(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			Volumes: []corev1.Volume{
				{Name: "orphan", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "used", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "used", MountPath: "/used"}, Target: "both"},
			},
		},
	}

	// Only "used" should be included, "orphan" has no mount
	result := BuildUserVolumes(lt, Master)
	assert.Len(t, result, 1)
	assert.Equal(t, "used", result[0].Name)
}

func TestBuildUserVolumeMounts_Empty(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			VolumeMounts: nil,
		},
	}

	result := BuildUserVolumeMounts(lt, Master)
	assert.Nil(t, result)

	result = BuildUserVolumeMounts(lt, Worker)
	assert.Nil(t, result)
}

func TestBuildUserVolumeMounts_MasterMode(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "results", MountPath: "/results"}, Target: "master"},
				{VolumeMount: corev1.VolumeMount{Name: "shared", MountPath: "/shared"}, Target: "both"},
				{VolumeMount: corev1.VolumeMount{Name: "certs", MountPath: "/certs"}, Target: "worker"},
			},
		},
	}

	result := BuildUserVolumeMounts(lt, Master)
	assert.Len(t, result, 2)
	paths := []string{result[0].MountPath, result[1].MountPath}
	assert.Contains(t, paths, "/results")
	assert.Contains(t, paths, "/shared")
}

func TestBuildUserVolumeMounts_WorkerMode(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "results", MountPath: "/results"}, Target: "master"},
				{VolumeMount: corev1.VolumeMount{Name: "shared", MountPath: "/shared"}, Target: "both"},
				{VolumeMount: corev1.VolumeMount{Name: "certs", MountPath: "/certs"}, Target: "worker"},
			},
		},
	}

	result := BuildUserVolumeMounts(lt, Worker)
	assert.Len(t, result, 2)
	paths := []string{result[0].MountPath, result[1].MountPath}
	assert.Contains(t, paths, "/shared")
	assert.Contains(t, paths, "/certs")
}

func TestBuildUserVolumeMounts_DefaultTarget(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{VolumeMount: corev1.VolumeMount{Name: "data", MountPath: "/data"}, Target: ""}, // Empty = both
			},
		},
	}

	// Should be included for both modes
	result := BuildUserVolumeMounts(lt, Master)
	assert.Len(t, result, 1)
	assert.Equal(t, "/data", result[0].MountPath)

	result = BuildUserVolumeMounts(lt, Worker)
	assert.Len(t, result, 1)
	assert.Equal(t, "/data", result[0].MountPath)
}

func TestBuildUserVolumeMounts_ConvertsToVolumeMount(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: locustv2.LocustTestSpec{
			VolumeMounts: []locustv2.TargetedVolumeMount{
				{
					VolumeMount: corev1.VolumeMount{
						Name:      "data",
						MountPath: "/data",
						ReadOnly:  true,
						SubPath:   "subdir",
					},
					Target: "both",
				},
			},
		},
	}

	result := BuildUserVolumeMounts(lt, Master)
	assert.Len(t, result, 1)
	assert.Equal(t, "data", result[0].Name)
	assert.Equal(t, "/data", result[0].MountPath)
	assert.True(t, result[0].ReadOnly)
	assert.Equal(t, "subdir", result[0].SubPath)
}

func TestShouldApplyMount_BothTarget(t *testing.T) {
	mount := locustv2.TargetedVolumeMount{
		VolumeMount: corev1.VolumeMount{Name: "test", MountPath: "/test"},
		Target:      "both",
	}

	assert.True(t, shouldApplyMount(mount, Master))
	assert.True(t, shouldApplyMount(mount, Worker))
}

func TestShouldApplyMount_MasterTarget(t *testing.T) {
	mount := locustv2.TargetedVolumeMount{
		VolumeMount: corev1.VolumeMount{Name: "test", MountPath: "/test"},
		Target:      "master",
	}

	assert.True(t, shouldApplyMount(mount, Master))
	assert.False(t, shouldApplyMount(mount, Worker))
}

func TestShouldApplyMount_WorkerTarget(t *testing.T) {
	mount := locustv2.TargetedVolumeMount{
		VolumeMount: corev1.VolumeMount{Name: "test", MountPath: "/test"},
		Target:      "worker",
	}

	assert.False(t, shouldApplyMount(mount, Master))
	assert.True(t, shouldApplyMount(mount, Worker))
}

func TestShouldApplyMount_EmptyTarget(t *testing.T) {
	mount := locustv2.TargetedVolumeMount{
		VolumeMount: corev1.VolumeMount{Name: "test", MountPath: "/test"},
		Target:      "", // Empty defaults to "both"
	}

	assert.True(t, shouldApplyMount(mount, Master))
	assert.True(t, shouldApplyMount(mount, Worker))
}

func TestShouldApplyMount_InvalidTarget(t *testing.T) {
	mount := locustv2.TargetedVolumeMount{
		VolumeMount: corev1.VolumeMount{Name: "test", MountPath: "/test"},
		Target:      "invalid",
	}

	assert.False(t, shouldApplyMount(mount, Master))
	assert.False(t, shouldApplyMount(mount, Worker))
}

func TestShouldIncludeVolume_HasMatchingMount(t *testing.T) {
	mounts := []locustv2.TargetedVolumeMount{
		{VolumeMount: corev1.VolumeMount{Name: "vol1", MountPath: "/vol1"}, Target: "master"},
		{VolumeMount: corev1.VolumeMount{Name: "vol2", MountPath: "/vol2"}, Target: "both"},
	}

	// vol1 matches master
	assert.True(t, shouldIncludeVolume("vol1", mounts, Master))
	assert.False(t, shouldIncludeVolume("vol1", mounts, Worker))

	// vol2 matches both
	assert.True(t, shouldIncludeVolume("vol2", mounts, Master))
	assert.True(t, shouldIncludeVolume("vol2", mounts, Worker))
}

func TestShouldIncludeVolume_NoMatchingMount(t *testing.T) {
	mounts := []locustv2.TargetedVolumeMount{
		{VolumeMount: corev1.VolumeMount{Name: "vol1", MountPath: "/vol1"}, Target: "master"},
	}

	// vol2 has no mount
	assert.False(t, shouldIncludeVolume("vol2", mounts, Master))
	assert.False(t, shouldIncludeVolume("vol2", mounts, Worker))
}

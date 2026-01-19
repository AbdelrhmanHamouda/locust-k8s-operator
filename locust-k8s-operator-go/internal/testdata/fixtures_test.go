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

package testdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLocustTest_Minimal(t *testing.T) {
	lt, err := LoadLocustTest("locusttest_minimal.json")
	require.NoError(t, err)
	require.NotNil(t, lt)

	assert.Equal(t, "minimal-test", lt.Name)
	assert.Equal(t, "default", lt.Namespace)
	assert.Equal(t, int32(1), lt.Spec.WorkerReplicas)
	assert.Equal(t, "locustio/locust:latest", lt.Spec.Image)
}

func TestLoadLocustTest_Full(t *testing.T) {
	lt, err := LoadLocustTest("locusttest_full.json")
	require.NoError(t, err)
	require.NotNil(t, lt)

	assert.Equal(t, "full-featured-test", lt.Name)
	assert.Equal(t, "load-testing", lt.Namespace)
	assert.Equal(t, int32(10), lt.Spec.WorkerReplicas)
	assert.Equal(t, "locustio/locust:2.20.0", lt.Spec.Image)
	assert.Equal(t, "IfNotPresent", lt.Spec.ImagePullPolicy)
	assert.Len(t, lt.Spec.ImagePullSecrets, 1)
	assert.Equal(t, "locust-scripts", lt.Spec.ConfigMap)
	assert.Equal(t, "locust-lib", lt.Spec.LibConfigMap)
	require.NotNil(t, lt.Spec.Labels)
	assert.Equal(t, "platform", lt.Spec.Labels.Master["team"])
}

func TestLoadLocustTest_WithAffinity(t *testing.T) {
	lt, err := LoadLocustTest("locusttest_with_affinity.json")
	require.NoError(t, err)
	require.NotNil(t, lt)

	assert.Equal(t, "affinity-test", lt.Name)
	require.NotNil(t, lt.Spec.Affinity)
	require.NotNil(t, lt.Spec.Affinity.NodeAffinity)
	assert.Equal(t, "performance", lt.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution["node-type"])
}

func TestLoadLocustTest_WithTolerations(t *testing.T) {
	lt, err := LoadLocustTest("locusttest_with_tolerations.json")
	require.NoError(t, err)
	require.NotNil(t, lt)

	assert.Equal(t, "tolerations-test", lt.Name)
	require.Len(t, lt.Spec.Tolerations, 2)
	assert.Equal(t, "dedicated", lt.Spec.Tolerations[0].Key)
	assert.Equal(t, "Equal", lt.Spec.Tolerations[0].Operator)
}

func TestLoadLocustTest_NotFound(t *testing.T) {
	lt, err := LoadLocustTest("nonexistent.json")
	assert.Error(t, err)
	assert.Nil(t, lt)
}

func TestMustLoadLocustTest_Success(t *testing.T) {
	lt := MustLoadLocustTest("locusttest_minimal.json")
	assert.NotNil(t, lt)
	assert.Equal(t, "minimal-test", lt.Name)
}

func TestMustLoadLocustTest_Panics(t *testing.T) {
	assert.Panics(t, func() {
		MustLoadLocustTest("nonexistent.json")
	})
}

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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocustTestSpec_JSONRoundTrip(t *testing.T) {
	spec := LocustTestSpec{
		MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
		WorkerCommandSeed: "--locustfile /lotest/src/test.py",
		WorkerReplicas:    3,
		Image:             "locustio/locust:latest",
		ImagePullPolicy:   "Always",
		ConfigMap:         "test-config",
		Labels: &PodLabels{
			Master: map[string]string{"app": "locust-master"},
			Worker: map[string]string{"app": "locust-worker"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(spec)
	require.NoError(t, err)

	// Unmarshal back
	var decoded LocustTestSpec
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify roundtrip
	assert.Equal(t, spec.MasterCommandSeed, decoded.MasterCommandSeed)
	assert.Equal(t, spec.WorkerReplicas, decoded.WorkerReplicas)
	assert.Equal(t, spec.Labels.Master["app"], decoded.Labels.Master["app"])
}

func TestLocustTestSpec_JSONFieldNames(t *testing.T) {
	spec := LocustTestSpec{
		MasterCommandSeed: "test",
		WorkerCommandSeed: "test",
		WorkerReplicas:    1,
		Image:             "test",
	}

	data, err := json.Marshal(spec)
	require.NoError(t, err)

	// Verify camelCase JSON field names
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"masterCommandSeed"`)
	assert.Contains(t, jsonStr, `"workerCommandSeed"`)
	assert.Contains(t, jsonStr, `"workerReplicas"`)
}

func TestLocustTestSpec_AllFields(t *testing.T) {
	spec := LocustTestSpec{
		MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
		WorkerCommandSeed: "--locustfile /lotest/src/test.py",
		WorkerReplicas:    5,
		Image:             "locustio/locust:2.15.1",
		ImagePullPolicy:   "IfNotPresent",
		ImagePullSecrets:  []string{"my-registry-secret"},
		ConfigMap:         "locust-test-config",
		LibConfigMap:      "locust-lib-config",
		Labels: &PodLabels{
			Master: map[string]string{"role": "master", "team": "platform"},
			Worker: map[string]string{"role": "worker", "team": "platform"},
		},
		Annotations: &PodAnnotations{
			Master: map[string]string{"prometheus.io/scrape": "true"},
			Worker: map[string]string{"prometheus.io/scrape": "true"},
		},
		Affinity: &LocustTestAffinity{
			NodeAffinity: &LocustTestNodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{
					"node-type": "compute",
				},
			},
		},
		Tolerations: []LocustTestToleration{
			{
				Key:      "dedicated",
				Operator: "Equal",
				Value:    "locust",
				Effect:   "NoSchedule",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(spec)
	require.NoError(t, err)

	// Unmarshal back
	var decoded LocustTestSpec
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, spec.MasterCommandSeed, decoded.MasterCommandSeed)
	assert.Equal(t, spec.WorkerCommandSeed, decoded.WorkerCommandSeed)
	assert.Equal(t, spec.WorkerReplicas, decoded.WorkerReplicas)
	assert.Equal(t, spec.Image, decoded.Image)
	assert.Equal(t, spec.ImagePullPolicy, decoded.ImagePullPolicy)
	assert.Equal(t, spec.ImagePullSecrets, decoded.ImagePullSecrets)
	assert.Equal(t, spec.ConfigMap, decoded.ConfigMap)
	assert.Equal(t, spec.LibConfigMap, decoded.LibConfigMap)
	assert.Equal(t, spec.Labels.Master, decoded.Labels.Master)
	assert.Equal(t, spec.Labels.Worker, decoded.Labels.Worker)
	assert.Equal(t, spec.Annotations.Master, decoded.Annotations.Master)
	assert.Equal(t, spec.Annotations.Worker, decoded.Annotations.Worker)
	assert.Equal(t, spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
		decoded.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	assert.Len(t, decoded.Tolerations, 1)
	assert.Equal(t, spec.Tolerations[0].Key, decoded.Tolerations[0].Key)
	assert.Equal(t, spec.Tolerations[0].Operator, decoded.Tolerations[0].Operator)
	assert.Equal(t, spec.Tolerations[0].Value, decoded.Tolerations[0].Value)
	assert.Equal(t, spec.Tolerations[0].Effect, decoded.Tolerations[0].Effect)
}

func TestLocustTestSpec_OmitEmptyFields(t *testing.T) {
	// Only required fields
	spec := LocustTestSpec{
		MasterCommandSeed: "test",
		WorkerCommandSeed: "test",
		WorkerReplicas:    1,
		Image:             "test",
	}

	data, err := json.Marshal(spec)
	require.NoError(t, err)

	jsonStr := string(data)

	// Optional fields should not be present
	assert.NotContains(t, jsonStr, `"imagePullPolicy"`)
	assert.NotContains(t, jsonStr, `"imagePullSecrets"`)
	assert.NotContains(t, jsonStr, `"configMap"`)
	assert.NotContains(t, jsonStr, `"libConfigMap"`)
	assert.NotContains(t, jsonStr, `"labels"`)
	assert.NotContains(t, jsonStr, `"annotations"`)
	assert.NotContains(t, jsonStr, `"affinity"`)
	assert.NotContains(t, jsonStr, `"tolerations"`)
}

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

// v1 is the deprecated spoke version, but the API server still serves and
// decodes it, so its JSON tags are a frozen wire contract: renaming one
// silently drops the field off existing CRs on the way into the conversion
// webhook. This test pins the serialized key names and the omitempty
// behaviour that keeps minimal v1 manifests free of empty optional fields.
func TestLocustTestSpec_WireFormat(t *testing.T) {
	required := LocustTestSpec{
		MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
		WorkerCommandSeed: "--locustfile /lotest/src/test.py",
		WorkerReplicas:    3,
		Image:             "locustio/locust:latest",
	}

	optionalKeys := []string{
		"imagePullPolicy",
		"imagePullSecrets",
		"configMap",
		"libConfigMap",
		"labels",
		"annotations",
		"affinity",
		"tolerations",
	}

	t.Run("required fields always serialize under their documented keys", func(t *testing.T) {
		data, err := json.Marshal(required)
		require.NoError(t, err)

		jsonStr := string(data)
		for _, key := range []string{"masterCommandSeed", "workerCommandSeed", "workerReplicas", "image"} {
			assert.Contains(t, jsonStr, `"`+key+`"`, "required field %q must always be serialized", key)
		}
	})

	t.Run("unset optional fields are omitted", func(t *testing.T) {
		data, err := json.Marshal(required)
		require.NoError(t, err)

		jsonStr := string(data)
		for _, key := range optionalKeys {
			assert.NotContains(t, jsonStr, `"`+key+`"`, "unset optional field %q must be omitted", key)
		}
	})

	t.Run("set optional fields serialize under their documented keys", func(t *testing.T) {
		full := required
		full.ImagePullPolicy = "IfNotPresent"
		full.ImagePullSecrets = []string{"my-registry-secret"}
		full.ConfigMap = "locust-test-config"
		full.LibConfigMap = "locust-lib-config"
		full.Labels = &PodLabels{
			Master: map[string]string{"role": "master"},
			Worker: map[string]string{"role": "worker"},
		}
		full.Annotations = &PodAnnotations{
			Master: map[string]string{"prometheus.io/scrape": "true"},
			Worker: map[string]string{"prometheus.io/scrape": "true"},
		}
		full.Affinity = &LocustTestAffinity{
			NodeAffinity: &LocustTestNodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: map[string]string{
					"node-type": "compute",
				},
			},
		}
		full.Tolerations = []LocustTestToleration{
			{Key: "dedicated", Operator: "Equal", Value: "locust", Effect: "NoSchedule"},
		}

		data, err := json.Marshal(full)
		require.NoError(t, err)

		jsonStr := string(data)
		for _, key := range optionalKeys {
			assert.Contains(t, jsonStr, `"`+key+`"`, "set optional field %q must be serialized", key)
		}

		// Nested keys carry the same contract; a rename here loses affinity and
		// toleration data on every stored v1 CR.
		for _, key := range []string{
			"nodeAffinity",
			"requiredDuringSchedulingIgnoredDuringExecution",
			"master",
			"worker",
			"key",
			"operator",
			"value",
			"effect",
		} {
			assert.Contains(t, jsonStr, `"`+key+`"`, "nested field %q must be serialized", key)
		}
	})
}

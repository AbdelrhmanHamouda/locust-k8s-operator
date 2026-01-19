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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeName(t *testing.T) {
	tests := []struct {
		name     string
		crName   string
		mode     OperationalMode
		expected string
	}{
		{
			name:     "basic master name",
			crName:   "my-test",
			mode:     Master,
			expected: "my-test-master",
		},
		{
			name:     "basic worker name",
			crName:   "my-test",
			mode:     Worker,
			expected: "my-test-worker",
		},
		{
			name:     "dots replaced with dashes",
			crName:   "team-a.load-test",
			mode:     Master,
			expected: "team-a-load-test-master",
		},
		{
			name:     "multiple dots replaced",
			crName:   "a.b.c.test",
			mode:     Worker,
			expected: "a-b-c-test-worker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NodeName(tt.crName, tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildLabels(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
		},
	}

	labels := BuildLabels(lt, Master)

	assert.Equal(t, "my-test", labels[LabelApp])
	assert.Equal(t, "my-test-master", labels[LabelPodName])
	assert.Equal(t, ManagedByValue, labels[LabelManagedBy])
	assert.Equal(t, "my-test", labels[LabelTestName])
}

func TestBuildLabels_WithUserLabels(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
				Labels: map[string]string{
					"custom-label": "master-value",
					"team":         "platform",
				},
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
				Labels: map[string]string{
					"custom-label": "worker-value",
				},
			},
		},
	}

	masterLabels := BuildLabels(lt, Master)
	assert.Equal(t, "master-value", masterLabels["custom-label"])
	assert.Equal(t, "platform", masterLabels["team"])

	workerLabels := BuildLabels(lt, Worker)
	assert.Equal(t, "worker-value", workerLabels["custom-label"])
	assert.Empty(t, workerLabels["team"])
}

func TestBuildAnnotations_Master(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
		},
	}

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	annotations := BuildAnnotations(lt, Master, cfg)

	assert.Equal(t, "true", annotations[AnnotationPrometheusScrape])
	assert.Equal(t, MetricsEndpointPath, annotations[AnnotationPrometheusPath])
	assert.Equal(t, "9646", annotations[AnnotationPrometheusPort])
}

func TestBuildAnnotations_Worker(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
		},
	}

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	annotations := BuildAnnotations(lt, Worker, cfg)

	// Worker should NOT have Prometheus annotations
	assert.Empty(t, annotations[AnnotationPrometheusScrape])
	assert.Empty(t, annotations[AnnotationPrometheusPath])
	assert.Empty(t, annotations[AnnotationPrometheusPort])
}

func TestBuildAnnotations_WithUserAnnotations(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
				Annotations: map[string]string{
					"custom-annotation": "master-value",
				},
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
				Annotations: map[string]string{
					"custom-annotation": "worker-value",
				},
			},
		},
	}

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	masterAnnotations := BuildAnnotations(lt, Master, cfg)
	assert.Equal(t, "master-value", masterAnnotations["custom-annotation"])
	// Should still have Prometheus annotations
	assert.Equal(t, "true", masterAnnotations[AnnotationPrometheusScrape])

	workerAnnotations := BuildAnnotations(lt, Worker, cfg)
	assert.Equal(t, "worker-value", workerAnnotations["custom-annotation"])
}

func TestBuildLabels_NilLabelsSpec(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/test.py",
				Labels:  nil,
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
		},
	}

	labels := BuildLabels(lt, Master)

	// Should have base labels even when user labels are nil
	assert.Equal(t, "my-test", labels[LabelApp])
	assert.Equal(t, ManagedByValue, labels[LabelManagedBy])
}

func TestBuildAnnotations_NilAnnotationsSpec(t *testing.T) {
	lt := &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-test",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:latest",
			Master: locustv2.MasterSpec{
				Command:     "locust -f /lotest/src/test.py",
				Annotations: nil,
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/test.py",
				Replicas: 3,
			},
		},
	}

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	annotations := BuildAnnotations(lt, Master, cfg)

	// Should still have Prometheus annotations for master
	assert.Equal(t, "true", annotations[AnnotationPrometheusScrape])
}

func TestWorkerPortInts(t *testing.T) {
	ports := WorkerPortInts()

	assert.Contains(t, ports, int32(WorkerPort))
	assert.Len(t, ports, 1)
}

func TestMasterPortInts(t *testing.T) {
	ports := MasterPortInts()

	assert.Contains(t, ports, int32(MasterPort))
	assert.Contains(t, ports, int32(MasterBindPort))
	assert.Contains(t, ports, int32(WebUIPort))
	assert.Len(t, ports, 3)
}

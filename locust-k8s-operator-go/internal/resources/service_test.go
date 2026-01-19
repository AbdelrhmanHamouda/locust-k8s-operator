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
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestLocustTestForService() *locustv2.LocustTest {
	return &locustv2.LocustTest{
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
}

func TestBuildMasterService(t *testing.T) {
	lt := newTestLocustTestForService()

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	svc := BuildMasterService(lt, cfg)

	require.NotNil(t, svc)
	assert.Equal(t, "my-test-master", svc.Name)
	assert.Equal(t, "default", svc.Namespace)
}

func TestBuildMasterService_Ports(t *testing.T) {
	lt := newTestLocustTestForService()

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	svc := BuildMasterService(lt, cfg)

	// Should have 3 ports: 5557, 5558, and metrics (9646)
	// WebUI port 8089 should NOT be included
	assert.Len(t, svc.Spec.Ports, 3)

	portNumbers := make([]int32, len(svc.Spec.Ports))
	for i, p := range svc.Spec.Ports {
		portNumbers[i] = p.Port
	}

	assert.Contains(t, portNumbers, int32(MasterPort))
	assert.Contains(t, portNumbers, int32(MasterBindPort))
	assert.Contains(t, portNumbers, int32(9646))
}

func TestBuildMasterService_NoWebUIPort(t *testing.T) {
	lt := newTestLocustTestForService()

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	svc := BuildMasterService(lt, cfg)

	// WebUI port 8089 should NOT be exposed
	for _, p := range svc.Spec.Ports {
		assert.NotEqual(t, int32(WebUIPort), p.Port, "WebUI port 8089 should NOT be exposed via service")
	}
}

func TestBuildMasterService_CustomMetricsPort(t *testing.T) {
	lt := newTestLocustTestForService()

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9999,
	}

	svc := BuildMasterService(lt, cfg)

	// Find the metrics port
	var metricsPort *corev1.ServicePort
	for i := range svc.Spec.Ports {
		if svc.Spec.Ports[i].Name == MetricsPortName {
			metricsPort = &svc.Spec.Ports[i]
			break
		}
	}

	require.NotNil(t, metricsPort, "Metrics port should exist")
	assert.Equal(t, int32(9999), metricsPort.Port)
}

func TestBuildMasterService_Selector(t *testing.T) {
	lt := newTestLocustTestForService()

	cfg := &config.OperatorConfig{
		MetricsExporterPort: 9646,
	}

	svc := BuildMasterService(lt, cfg)

	assert.Equal(t, "my-test-master", svc.Spec.Selector[LabelPodName])
}

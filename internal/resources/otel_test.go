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

// Helper to create a minimal LocustTest for OTel testing
func newOTelTestLocustTest() *locustv2.LocustTest {
	return &locustv2.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lt",
			Namespace: "default",
		},
		Spec: locustv2.LocustTestSpec{
			Image: "locustio/locust:2.32.0",
			Master: locustv2.MasterSpec{
				Command: "locust -f /lotest/src/locustfile.py",
			},
			Worker: locustv2.WorkerSpec{
				Command:  "locust -f /lotest/src/locustfile.py",
				Replicas: 3,
			},
		},
	}
}

// ===== IsOTelEnabled Tests =====

func TestIsOTelEnabled_NilObservability(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = nil

	assert.False(t, IsOTelEnabled(lt))
}

func TestIsOTelEnabled_NilOpenTelemetry(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: nil,
	}

	assert.False(t, IsOTelEnabled(lt))
}

func TestIsOTelEnabled_Disabled(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled: false,
		},
	}

	assert.False(t, IsOTelEnabled(lt))
}

func TestIsOTelEnabled_Enabled(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4317",
		},
	}

	assert.True(t, IsOTelEnabled(lt))
}

// ===== GetOTelConfig Tests =====

func TestGetOTelConfig_NilObservability(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = nil

	assert.Nil(t, GetOTelConfig(lt))
}

func TestGetOTelConfig_NilOpenTelemetry(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: nil,
	}

	assert.Nil(t, GetOTelConfig(lt))
}

func TestGetOTelConfig_HasConfig(t *testing.T) {
	lt := newOTelTestLocustTest()
	expectedConfig := &locustv2.OpenTelemetryConfig{
		Enabled:  true,
		Endpoint: "otel-collector:4317",
		Protocol: "grpc",
	}
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: expectedConfig,
	}

	result := GetOTelConfig(lt)
	assert.NotNil(t, result)
	assert.Equal(t, expectedConfig, result)
}

// ===== BuildOTelEnvVars Tests =====

func TestBuildOTelEnvVars_Disabled(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled: false,
		},
	}

	envVars := BuildOTelEnvVars(lt)
	assert.Nil(t, envVars)
}

func TestBuildOTelEnvVars_NilObservability(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = nil

	envVars := BuildOTelEnvVars(lt)
	assert.Nil(t, envVars)
}

func TestBuildOTelEnvVars_EnabledMinimal(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4317",
		},
	}

	envVars := BuildOTelEnvVars(lt)
	assert.NotNil(t, envVars)

	// Should have: traces exporter, metrics exporter, endpoint, protocol (default grpc)
	assert.Len(t, envVars, 4)

	envMap := envVarsToMap(envVars)
	assert.Equal(t, "otlp", envMap[EnvOTelTracesExporter])
	assert.Equal(t, "otlp", envMap[EnvOTelMetricsExporter])
	assert.Equal(t, "otel-collector:4317", envMap[EnvOTelExporterEndpoint])
	assert.Equal(t, "grpc", envMap[EnvOTelExporterProtocol])
}

func TestBuildOTelEnvVars_FullConfig(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector.monitoring:4317",
			Protocol: "grpc",
			Insecure: true,
		},
	}

	envVars := BuildOTelEnvVars(lt)
	assert.NotNil(t, envVars)

	// Should have: traces exporter, metrics exporter, endpoint, protocol, insecure
	assert.Len(t, envVars, 5)

	envMap := envVarsToMap(envVars)
	assert.Equal(t, "otlp", envMap[EnvOTelTracesExporter])
	assert.Equal(t, "otlp", envMap[EnvOTelMetricsExporter])
	assert.Equal(t, "otel-collector.monitoring:4317", envMap[EnvOTelExporterEndpoint])
	assert.Equal(t, "grpc", envMap[EnvOTelExporterProtocol])
	assert.Equal(t, "true", envMap[EnvOTelExporterInsecure])
}

func TestBuildOTelEnvVars_ExtraEnvVars(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4317",
			ExtraEnvVars: map[string]string{
				"OTEL_RESOURCE_ATTRIBUTES": "service.name=locust-load-test",
				"OTEL_TRACES_SAMPLER":      "parentbased_traceidratio",
				"OTEL_TRACES_SAMPLER_ARG":  "0.1",
			},
		},
	}

	envVars := BuildOTelEnvVars(lt)
	assert.NotNil(t, envVars)

	// Should have: traces exporter, metrics exporter, endpoint, protocol (default) + 3 extra
	assert.Len(t, envVars, 7)

	envMap := envVarsToMap(envVars)
	assert.Equal(t, "service.name=locust-load-test", envMap["OTEL_RESOURCE_ATTRIBUTES"])
	assert.Equal(t, "parentbased_traceidratio", envMap["OTEL_TRACES_SAMPLER"])
	assert.Equal(t, "0.1", envMap["OTEL_TRACES_SAMPLER_ARG"])
}

func TestBuildOTelEnvVars_DefaultProtocol(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4317",
			Protocol: "", // Empty - should default to grpc
		},
	}

	envVars := BuildOTelEnvVars(lt)
	envMap := envVarsToMap(envVars)
	assert.Equal(t, "grpc", envMap[EnvOTelExporterProtocol])
}

func TestBuildOTelEnvVars_HTTPProtocol(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4318",
			Protocol: "http/protobuf",
		},
	}

	envVars := BuildOTelEnvVars(lt)
	envMap := envVarsToMap(envVars)
	assert.Equal(t, "http/protobuf", envMap[EnvOTelExporterProtocol])
}

func TestBuildOTelEnvVars_InsecureFalse(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4317",
			Insecure: false,
		},
	}

	envVars := BuildOTelEnvVars(lt)
	envMap := envVarsToMap(envVars)

	// Insecure env var should NOT be present when false
	_, exists := envMap[EnvOTelExporterInsecure]
	assert.False(t, exists, "OTEL_EXPORTER_OTLP_INSECURE should not be set when Insecure is false")
}

func TestBuildOTelEnvVars_ExtraEnvVarsSortedOrder(t *testing.T) {
	lt := newOTelTestLocustTest()
	lt.Spec.Observability = &locustv2.ObservabilityConfig{
		OpenTelemetry: &locustv2.OpenTelemetryConfig{
			Enabled:  true,
			Endpoint: "otel-collector:4317",
			ExtraEnvVars: map[string]string{
				"ZEBRA_VAR": "z",
				"APPLE_VAR": "a",
				"MANGO_VAR": "m",
			},
		},
	}

	envVars := BuildOTelEnvVars(lt)

	// Extra env vars should appear after the core ones, in sorted order
	// Core: OTEL_TRACES_EXPORTER, OTEL_METRICS_EXPORTER, OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_PROTOCOL
	// Extra (sorted): APPLE_VAR, MANGO_VAR, ZEBRA_VAR
	assert.Len(t, envVars, 7)
	assert.Equal(t, "APPLE_VAR", envVars[4].Name)
	assert.Equal(t, "MANGO_VAR", envVars[5].Name)
	assert.Equal(t, "ZEBRA_VAR", envVars[6].Name)
}

// Helper to convert env vars slice to map for easier assertions
func envVarsToMap(envVars []corev1.EnvVar) map[string]string {
	result := make(map[string]string)
	for _, ev := range envVars {
		result[ev.Name] = ev.Value
	}
	return result
}

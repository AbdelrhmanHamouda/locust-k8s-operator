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
	"sort"
	"strconv"

	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
	corev1 "k8s.io/api/core/v1"
)

// OTel environment variable names
const (
	EnvOTelTracesExporter   = "OTEL_TRACES_EXPORTER"
	EnvOTelMetricsExporter  = "OTEL_METRICS_EXPORTER"
	EnvOTelExporterEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	EnvOTelExporterProtocol = "OTEL_EXPORTER_OTLP_PROTOCOL"
	EnvOTelExporterInsecure = "OTEL_EXPORTER_OTLP_INSECURE"
)

// Default OTel values
const (
	OTelExporterOTLP = "otlp"
	OTelProtocolGRPC = "grpc"
	OTelProtocolHTTP = "http/protobuf"
)

// IsOTelEnabled returns true if OpenTelemetry is enabled in the spec.
func IsOTelEnabled(lt *locustv2.LocustTest) bool {
	if lt.Spec.Observability == nil {
		return false
	}
	if lt.Spec.Observability.OpenTelemetry == nil {
		return false
	}
	return lt.Spec.Observability.OpenTelemetry.Enabled
}

// GetOTelConfig returns the OpenTelemetry configuration, or nil if not configured.
func GetOTelConfig(lt *locustv2.LocustTest) *locustv2.OpenTelemetryConfig {
	if lt.Spec.Observability == nil {
		return nil
	}
	return lt.Spec.Observability.OpenTelemetry
}

// BuildOTelEnvVars creates environment variables for OpenTelemetry configuration.
// Returns nil if OTel is not enabled.
func BuildOTelEnvVars(lt *locustv2.LocustTest) []corev1.EnvVar {
	if !IsOTelEnabled(lt) {
		return nil
	}

	otelCfg := GetOTelConfig(lt)
	if otelCfg == nil {
		return nil
	}

	var envVars []corev1.EnvVar

	// Core OTel exporter configuration
	envVars = append(envVars,
		corev1.EnvVar{Name: EnvOTelTracesExporter, Value: OTelExporterOTLP},
		corev1.EnvVar{Name: EnvOTelMetricsExporter, Value: OTelExporterOTLP},
	)

	// Endpoint (required when enabled - validated by webhook)
	if otelCfg.Endpoint != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvOTelExporterEndpoint,
			Value: otelCfg.Endpoint,
		})
	}

	// Protocol (default: grpc)
	protocol := otelCfg.Protocol
	if protocol == "" {
		protocol = OTelProtocolGRPC
	}
	envVars = append(envVars, corev1.EnvVar{
		Name:  EnvOTelExporterProtocol,
		Value: protocol,
	})

	// Insecure flag (only set if true)
	if otelCfg.Insecure {
		envVars = append(envVars, corev1.EnvVar{
			Name:  EnvOTelExporterInsecure,
			Value: strconv.FormatBool(true),
		})
	}

	// Extra environment variables from spec (sorted for deterministic order)
	if len(otelCfg.ExtraEnvVars) > 0 {
		keys := make([]string, 0, len(otelCfg.ExtraEnvVars))
		for key := range otelCfg.ExtraEnvVars {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: otelCfg.ExtraEnvVars[key],
			})
		}
	}

	return envVars
}

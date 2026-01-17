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

package config

import (
	"os"
	"strconv"
)

// OperatorConfig holds all operator configuration loaded from environment variables.
type OperatorConfig struct {
	// Job configuration
	// TTLSecondsAfterFinished specifies how long a Job should exist after completion.
	// nil means use Kubernetes default (don't set the field).
	TTLSecondsAfterFinished *int32

	// Pod resource configuration for Locust containers
	PodCPURequest              string
	PodMemRequest              string
	PodEphemeralStorageRequest string
	PodCPULimit                string
	PodMemLimit                string
	PodEphemeralStorageLimit   string

	// Metrics exporter sidecar configuration
	MetricsExporterImage                   string
	MetricsExporterPort                    int32
	MetricsExporterPullPolicy              string
	MetricsExporterCPURequest              string
	MetricsExporterMemRequest              string
	MetricsExporterEphemeralStorageRequest string
	MetricsExporterCPULimit                string
	MetricsExporterMemLimit                string
	MetricsExporterEphemeralStorageLimit   string

	// Kafka configuration for optional Kafka integration
	KafkaBootstrapServers string
	KafkaSecurityEnabled  bool
	KafkaSecurityProtocol string
	KafkaUsername         string
	KafkaPassword         string
	KafkaSaslMechanism    string
	KafkaSaslJaasConfig   string

	// Feature flags
	// EnableAffinityCRInjection enables injecting affinity rules from CR spec
	EnableAffinityCRInjection bool
	// EnableTolerationsCRInjection enables injecting tolerations from CR spec
	EnableTolerationsCRInjection bool
}

// LoadConfig loads operator configuration from environment variables.
// Default values match those in the Java operator's application.yml.
func LoadConfig() *OperatorConfig {
	return &OperatorConfig{
		// Job configuration
		TTLSecondsAfterFinished: getEnvInt32Ptr("JOB_TTL_SECONDS_AFTER_FINISHED"),

		// Pod resource configuration
		PodCPURequest:              getEnv("POD_CPU_REQUEST", "250m"),
		PodMemRequest:              getEnv("POD_MEM_REQUEST", "128Mi"),
		PodEphemeralStorageRequest: getEnv("POD_EPHEMERAL_REQUEST", "30M"),
		PodCPULimit:                getEnv("POD_CPU_LIMIT", "1000m"),
		PodMemLimit:                getEnv("POD_MEM_LIMIT", "1024Mi"),
		PodEphemeralStorageLimit:   getEnv("POD_EPHEMERAL_LIMIT", "50M"),

		// Metrics exporter configuration
		MetricsExporterImage:                   getEnv("METRICS_EXPORTER_IMAGE", "containersol/locust_exporter:v0.5.0"),
		MetricsExporterPort:                    getEnvInt32("METRICS_EXPORTER_PORT", 9646),
		MetricsExporterPullPolicy:              getEnv("METRICS_EXPORTER_IMAGE_PULL_POLICY", "Always"),
		MetricsExporterCPURequest:              getEnv("METRICS_EXPORTER_CPU_REQUEST", "250m"),
		MetricsExporterMemRequest:              getEnv("METRICS_EXPORTER_MEM_REQUEST", "128Mi"),
		MetricsExporterEphemeralStorageRequest: getEnv("METRICS_EXPORTER_EPHEMERAL_REQUEST", "30M"),
		MetricsExporterCPULimit:                getEnv("METRICS_EXPORTER_CPU_LIMIT", "1000m"),
		MetricsExporterMemLimit:                getEnv("METRICS_EXPORTER_MEM_LIMIT", "1024Mi"),
		MetricsExporterEphemeralStorageLimit:   getEnv("METRICS_EXPORTER_EPHEMERAL_LIMIT", "50M"),

		// Kafka configuration
		KafkaBootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaSecurityEnabled:  getEnvBool("KAFKA_SECURITY_ENABLED", false),
		KafkaSecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL_CONFIG", "SASL_PLAINTEXT"),
		KafkaUsername:         getEnv("KAFKA_USERNAME", ""),
		KafkaPassword:         getEnv("KAFKA_PASSWORD", ""),
		KafkaSaslMechanism:    getEnv("KAFKA_SASL_MECHANISM", "SCRAM-SHA-512"),
		KafkaSaslJaasConfig:   getEnv("KAFKA_SASL_JAAS_CONFIG", ""),

		// Feature flags
		EnableAffinityCRInjection:    getEnvBool("ENABLE_AFFINITY_CR_INJECTION", false),
		EnableTolerationsCRInjection: getEnvBool("ENABLE_TAINT_TOLERATIONS_CR_INJECTION", false),
	}
}

// getEnv returns the value of an environment variable or a default value if not set.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// getEnvBool returns the boolean value of an environment variable or a default value if not set.
func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}

// getEnvInt32 returns the int32 value of an environment variable or a default value if not set.
func getEnvInt32(key string, defaultValue int32) int32 {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return defaultValue
		}
		return int32(i)
	}
	return defaultValue
}

// getEnvInt32Ptr returns a pointer to an int32 value of an environment variable, or nil if not set.
// This is used for optional fields where nil indicates "not configured" vs 0.
func getEnvInt32Ptr(key string) *int32 {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil
		}
		val := int32(i)
		return &val
	}
	return nil
}

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
	"fmt"
	"os"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
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

	// Role-specific pod resource configuration for master/worker Locust containers.
	// Empty string means "use unified Pod* resource values above".
	// Set via Helm masterResources/workerResources values.
	MasterCPURequest              string
	MasterMemRequest              string
	MasterEphemeralStorageRequest string
	MasterCPULimit                string
	MasterMemLimit                string
	MasterEphemeralStorageLimit   string
	WorkerCPURequest              string
	WorkerMemRequest              string
	WorkerEphemeralStorageRequest string
	WorkerCPULimit                string
	WorkerMemLimit                string
	WorkerEphemeralStorageLimit   string

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
// Returns error if any resource values are invalid Kubernetes quantities.
func LoadConfig() (*OperatorConfig, error) {
	cfg := &OperatorConfig{
		// Job configuration
		TTLSecondsAfterFinished: getEnvInt32Ptr("JOB_TTL_SECONDS_AFTER_FINISHED"),

		// Pod resource configuration
		PodCPURequest:              getEnv("POD_CPU_REQUEST", "250m"),
		PodMemRequest:              getEnv("POD_MEM_REQUEST", "128Mi"),
		PodEphemeralStorageRequest: getEnv("POD_EPHEMERAL_REQUEST", "30M"),
		PodCPULimit:                getEnv("POD_CPU_LIMIT", "1000m"),
		PodMemLimit:                getEnv("POD_MEM_LIMIT", "1024Mi"),
		PodEphemeralStorageLimit:   getEnv("POD_EPHEMERAL_LIMIT", "50M"),

		// Role-specific pod resources (empty = use unified Pod* values above)
		MasterCPURequest:              getEnv("MASTER_POD_CPU_REQUEST", ""),
		MasterMemRequest:              getEnv("MASTER_POD_MEM_REQUEST", ""),
		MasterEphemeralStorageRequest: getEnv("MASTER_POD_EPHEMERAL_REQUEST", ""),
		MasterCPULimit:                getEnv("MASTER_POD_CPU_LIMIT", ""),
		MasterMemLimit:                getEnv("MASTER_POD_MEM_LIMIT", ""),
		MasterEphemeralStorageLimit:   getEnv("MASTER_POD_EPHEMERAL_LIMIT", ""),
		WorkerCPURequest:              getEnv("WORKER_POD_CPU_REQUEST", ""),
		WorkerMemRequest:              getEnv("WORKER_POD_MEM_REQUEST", ""),
		WorkerEphemeralStorageRequest: getEnv("WORKER_POD_EPHEMERAL_REQUEST", ""),
		WorkerCPULimit:                getEnv("WORKER_POD_CPU_LIMIT", ""),
		WorkerMemLimit:                getEnv("WORKER_POD_MEM_LIMIT", ""),
		WorkerEphemeralStorageLimit:   getEnv("WORKER_POD_EPHEMERAL_LIMIT", ""),

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

	// Validate all resource quantities at startup
	if err := validateResourceQuantities(cfg); err != nil {
		return nil, fmt.Errorf("invalid operator configuration: %w", err)
	}

	return cfg, nil
}

// validateResourceQuantities validates all resource quantity strings in config.
func validateResourceQuantities(cfg *OperatorConfig) error {
	quantities := map[string]string{
		"POD_CPU_REQUEST":                    cfg.PodCPURequest,
		"POD_MEM_REQUEST":                    cfg.PodMemRequest,
		"POD_EPHEMERAL_REQUEST":              cfg.PodEphemeralStorageRequest,
		"POD_CPU_LIMIT":                      cfg.PodCPULimit,
		"POD_MEM_LIMIT":                      cfg.PodMemLimit,
		"POD_EPHEMERAL_LIMIT":                cfg.PodEphemeralStorageLimit,
		"MASTER_POD_CPU_REQUEST":             cfg.MasterCPURequest,
		"MASTER_POD_MEM_REQUEST":             cfg.MasterMemRequest,
		"MASTER_POD_EPHEMERAL_REQUEST":       cfg.MasterEphemeralStorageRequest,
		"MASTER_POD_CPU_LIMIT":               cfg.MasterCPULimit,
		"MASTER_POD_MEM_LIMIT":               cfg.MasterMemLimit,
		"MASTER_POD_EPHEMERAL_LIMIT":         cfg.MasterEphemeralStorageLimit,
		"WORKER_POD_CPU_REQUEST":             cfg.WorkerCPURequest,
		"WORKER_POD_MEM_REQUEST":             cfg.WorkerMemRequest,
		"WORKER_POD_EPHEMERAL_REQUEST":       cfg.WorkerEphemeralStorageRequest,
		"WORKER_POD_CPU_LIMIT":               cfg.WorkerCPULimit,
		"WORKER_POD_MEM_LIMIT":               cfg.WorkerMemLimit,
		"WORKER_POD_EPHEMERAL_LIMIT":         cfg.WorkerEphemeralStorageLimit,
		"METRICS_EXPORTER_CPU_REQUEST":       cfg.MetricsExporterCPURequest,
		"METRICS_EXPORTER_MEM_REQUEST":       cfg.MetricsExporterMemRequest,
		"METRICS_EXPORTER_EPHEMERAL_REQUEST": cfg.MetricsExporterEphemeralStorageRequest,
		"METRICS_EXPORTER_CPU_LIMIT":         cfg.MetricsExporterCPULimit,
		"METRICS_EXPORTER_MEM_LIMIT":         cfg.MetricsExporterMemLimit,
		"METRICS_EXPORTER_EPHEMERAL_LIMIT":   cfg.MetricsExporterEphemeralStorageLimit,
	}

	var errs []error
	for name, value := range quantities {
		if value == "" {
			continue // Empty string means "not set", which is valid
		}
		if _, err := resource.ParseQuantity(value); err != nil {
			errs = append(errs, fmt.Errorf("invalid value for %s: %q is not a valid Kubernetes quantity", name, value))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("resource validation failed:\n%s", formatErrors(errs))
	}

	return nil
}

// formatErrors formats a slice of errors for display.
func formatErrors(errs []error) string {
	msgs := make([]string, 0, len(errs))
	for _, err := range errs {
		msgs = append(msgs, "  - "+err.Error())
	}
	return strings.Join(msgs, "\n")
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

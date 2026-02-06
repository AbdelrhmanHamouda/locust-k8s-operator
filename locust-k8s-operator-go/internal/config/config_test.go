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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear any existing env vars that might interfere
	envVars := []string{
		"JOB_TTL_SECONDS_AFTER_FINISHED",
		"POD_CPU_REQUEST",
		"POD_MEM_REQUEST",
		"POD_EPHEMERAL_REQUEST",
		"POD_CPU_LIMIT",
		"POD_MEM_LIMIT",
		"POD_EPHEMERAL_LIMIT",
		"METRICS_EXPORTER_IMAGE",
		"METRICS_EXPORTER_PORT",
		"METRICS_EXPORTER_IMAGE_PULL_POLICY",
		"METRICS_EXPORTER_CPU_REQUEST",
		"METRICS_EXPORTER_MEM_REQUEST",
		"METRICS_EXPORTER_EPHEMERAL_REQUEST",
		"METRICS_EXPORTER_CPU_LIMIT",
		"METRICS_EXPORTER_MEM_LIMIT",
		"METRICS_EXPORTER_EPHEMERAL_LIMIT",
		"KAFKA_BOOTSTRAP_SERVERS",
		"KAFKA_SECURITY_ENABLED",
		"KAFKA_SECURITY_PROTOCOL_CONFIG",
		"KAFKA_USERNAME",
		"KAFKA_PASSWORD",
		"KAFKA_SASL_MECHANISM",
		"KAFKA_SASL_JAAS_CONFIG",
		"ENABLE_AFFINITY_CR_INJECTION",
		"ENABLE_TAINT_TOLERATIONS_CR_INJECTION",
	}
	for _, env := range envVars {
		_ = os.Unsetenv(env)
	}

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Job configuration
	assert.Nil(t, cfg.TTLSecondsAfterFinished, "TTL should be nil when not set")

	// Pod resources - match Java application.yml defaults
	assert.Equal(t, "250m", cfg.PodCPURequest)
	assert.Equal(t, "128Mi", cfg.PodMemRequest)
	assert.Equal(t, "30M", cfg.PodEphemeralStorageRequest)
	assert.Equal(t, "1000m", cfg.PodCPULimit)
	assert.Equal(t, "1024Mi", cfg.PodMemLimit)
	assert.Equal(t, "50M", cfg.PodEphemeralStorageLimit)

	// Metrics exporter - match Java application.yml defaults
	assert.Equal(t, "containersol/locust_exporter:v0.5.0", cfg.MetricsExporterImage)
	assert.Equal(t, int32(9646), cfg.MetricsExporterPort)
	assert.Equal(t, "Always", cfg.MetricsExporterPullPolicy)
	assert.Equal(t, "250m", cfg.MetricsExporterCPURequest)
	assert.Equal(t, "128Mi", cfg.MetricsExporterMemRequest)
	assert.Equal(t, "30M", cfg.MetricsExporterEphemeralStorageRequest)
	assert.Equal(t, "1000m", cfg.MetricsExporterCPULimit)
	assert.Equal(t, "1024Mi", cfg.MetricsExporterMemLimit)
	assert.Equal(t, "50M", cfg.MetricsExporterEphemeralStorageLimit)

	// Kafka configuration - match Java application.yml defaults
	assert.Equal(t, "localhost:9092", cfg.KafkaBootstrapServers)
	assert.False(t, cfg.KafkaSecurityEnabled)
	assert.Equal(t, "SASL_PLAINTEXT", cfg.KafkaSecurityProtocol)
	assert.Equal(t, "", cfg.KafkaUsername) // Empty for security
	assert.Equal(t, "", cfg.KafkaPassword) // Empty for security
	assert.Equal(t, "SCRAM-SHA-512", cfg.KafkaSaslMechanism)
	assert.Equal(t, "", cfg.KafkaSaslJaasConfig) // Empty for security

	// Feature flags
	assert.False(t, cfg.EnableAffinityCRInjection)
	assert.False(t, cfg.EnableTolerationsCRInjection)
}

func TestLoadConfig_EnvironmentOverrides(t *testing.T) {
	// Set up test env vars
	t.Setenv("JOB_TTL_SECONDS_AFTER_FINISHED", "300")
	t.Setenv("POD_CPU_REQUEST", "500m")
	t.Setenv("POD_MEM_REQUEST", "256Mi")
	t.Setenv("POD_EPHEMERAL_REQUEST", "100M")
	t.Setenv("POD_CPU_LIMIT", "2000m")
	t.Setenv("POD_MEM_LIMIT", "2048Mi")
	t.Setenv("POD_EPHEMERAL_LIMIT", "200M")
	t.Setenv("METRICS_EXPORTER_IMAGE", "custom/exporter:v1.0.0")
	t.Setenv("METRICS_EXPORTER_PORT", "9000")
	t.Setenv("METRICS_EXPORTER_IMAGE_PULL_POLICY", "IfNotPresent")
	t.Setenv("ENABLE_AFFINITY_CR_INJECTION", "true")
	t.Setenv("ENABLE_TAINT_TOLERATIONS_CR_INJECTION", "true")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify overrides
	require.NotNil(t, cfg.TTLSecondsAfterFinished)
	assert.Equal(t, int32(300), *cfg.TTLSecondsAfterFinished)
	assert.Equal(t, "500m", cfg.PodCPURequest)
	assert.Equal(t, "256Mi", cfg.PodMemRequest)
	assert.Equal(t, "100M", cfg.PodEphemeralStorageRequest)
	assert.Equal(t, "2000m", cfg.PodCPULimit)
	assert.Equal(t, "2048Mi", cfg.PodMemLimit)
	assert.Equal(t, "200M", cfg.PodEphemeralStorageLimit)
	assert.Equal(t, "custom/exporter:v1.0.0", cfg.MetricsExporterImage)
	assert.Equal(t, int32(9000), cfg.MetricsExporterPort)
	assert.Equal(t, "IfNotPresent", cfg.MetricsExporterPullPolicy)
	assert.True(t, cfg.EnableAffinityCRInjection)
	assert.True(t, cfg.EnableTolerationsCRInjection)
}

func TestLoadConfig_TTLSecondsAfterFinished_ZeroValue(t *testing.T) {
	t.Setenv("JOB_TTL_SECONDS_AFTER_FINISHED", "0")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// TTL of 0 should be a valid value (immediate cleanup)
	require.NotNil(t, cfg.TTLSecondsAfterFinished)
	assert.Equal(t, int32(0), *cfg.TTLSecondsAfterFinished)
}

func TestLoadConfig_KafkaConfiguration(t *testing.T) {
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka.example.com:9092")
	t.Setenv("KAFKA_SECURITY_ENABLED", "true")
	t.Setenv("KAFKA_SECURITY_PROTOCOL_CONFIG", "SASL_SSL")
	t.Setenv("KAFKA_USERNAME", "user")
	t.Setenv("KAFKA_PASSWORD", "secret")
	t.Setenv("KAFKA_SASL_MECHANISM", "PLAIN")
	t.Setenv("KAFKA_SASL_JAAS_CONFIG", "org.apache.kafka.common.security.plain.PlainLoginModule required;")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "kafka.example.com:9092", cfg.KafkaBootstrapServers)
	assert.True(t, cfg.KafkaSecurityEnabled)
	assert.Equal(t, "SASL_SSL", cfg.KafkaSecurityProtocol)
	assert.Equal(t, "user", cfg.KafkaUsername)
	assert.Equal(t, "secret", cfg.KafkaPassword)
	assert.Equal(t, "PLAIN", cfg.KafkaSaslMechanism)
	assert.Equal(t, "org.apache.kafka.common.security.plain.PlainLoginModule required;", cfg.KafkaSaslJaasConfig)
}

func TestLoadConfig_MetricsExporterConfiguration(t *testing.T) {
	t.Setenv("METRICS_EXPORTER_IMAGE", "myregistry/locust-exporter:v2.0.0")
	t.Setenv("METRICS_EXPORTER_PORT", "8080")
	t.Setenv("METRICS_EXPORTER_IMAGE_PULL_POLICY", "Never")
	t.Setenv("METRICS_EXPORTER_CPU_REQUEST", "100m")
	t.Setenv("METRICS_EXPORTER_MEM_REQUEST", "64Mi")
	t.Setenv("METRICS_EXPORTER_EPHEMERAL_REQUEST", "10M")
	t.Setenv("METRICS_EXPORTER_CPU_LIMIT", "500m")
	t.Setenv("METRICS_EXPORTER_MEM_LIMIT", "512Mi")
	t.Setenv("METRICS_EXPORTER_EPHEMERAL_LIMIT", "100M")

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "myregistry/locust-exporter:v2.0.0", cfg.MetricsExporterImage)
	assert.Equal(t, int32(8080), cfg.MetricsExporterPort)
	assert.Equal(t, "Never", cfg.MetricsExporterPullPolicy)
	assert.Equal(t, "100m", cfg.MetricsExporterCPURequest)
	assert.Equal(t, "64Mi", cfg.MetricsExporterMemRequest)
	assert.Equal(t, "10M", cfg.MetricsExporterEphemeralStorageRequest)
	assert.Equal(t, "500m", cfg.MetricsExporterCPULimit)
	assert.Equal(t, "512Mi", cfg.MetricsExporterMemLimit)
	assert.Equal(t, "100M", cfg.MetricsExporterEphemeralStorageLimit)
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "value")

	assert.Equal(t, "value", getEnv("TEST_VAR", "default"))
	assert.Equal(t, "default", getEnv("NONEXISTENT_VAR", "default"))
}

func TestGetEnv_EmptyValue(t *testing.T) {
	// Empty string should return default
	t.Setenv("TEST_EMPTY", "")
	assert.Equal(t, "default", getEnv("TEST_EMPTY", "default"))
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue bool
		expected     bool
	}{
		{"true string", "true", true, false, true},
		{"false string", "false", true, true, false},
		{"1 value", "1", true, false, true},
		{"0 value", "0", true, true, false},
		{"TRUE uppercase", "TRUE", true, false, true},
		{"FALSE uppercase", "FALSE", true, true, false},
		{"invalid value returns default true", "invalid", true, true, true},
		{"invalid value returns default false", "invalid", true, false, false},
		{"unset uses default true", "", false, true, true},
		{"unset uses default false", "", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("TEST_BOOL", tt.envValue)
			} else {
				_ = os.Unsetenv("TEST_BOOL")
			}
			result := getEnvBool("TEST_BOOL", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt32(t *testing.T) {
	t.Run("valid integer", func(t *testing.T) {
		t.Setenv("TEST_INT", "42")
		assert.Equal(t, int32(42), getEnvInt32("TEST_INT", 0))
	})

	t.Run("unset returns default", func(t *testing.T) {
		_ = os.Unsetenv("TEST_INT")
		assert.Equal(t, int32(100), getEnvInt32("TEST_INT", 100))
	})

	t.Run("invalid value returns default", func(t *testing.T) {
		t.Setenv("TEST_INT", "invalid")
		assert.Equal(t, int32(100), getEnvInt32("TEST_INT", 100))
	})

	t.Run("negative value", func(t *testing.T) {
		t.Setenv("TEST_INT", "-10")
		assert.Equal(t, int32(-10), getEnvInt32("TEST_INT", 0))
	})

	t.Run("zero value", func(t *testing.T) {
		t.Setenv("TEST_INT", "0")
		assert.Equal(t, int32(0), getEnvInt32("TEST_INT", 100))
	})
}

func TestGetEnvInt32Ptr(t *testing.T) {
	t.Run("not set returns nil", func(t *testing.T) {
		_ = os.Unsetenv("TEST_PTR")
		result := getEnvInt32Ptr("TEST_PTR")
		assert.Nil(t, result)
	})

	t.Run("valid value returns pointer", func(t *testing.T) {
		t.Setenv("TEST_PTR", "42")
		result := getEnvInt32Ptr("TEST_PTR")
		require.NotNil(t, result)
		assert.Equal(t, int32(42), *result)
	})

	t.Run("zero value returns pointer to zero", func(t *testing.T) {
		t.Setenv("TEST_PTR", "0")
		result := getEnvInt32Ptr("TEST_PTR")
		require.NotNil(t, result)
		assert.Equal(t, int32(0), *result)
	})

	t.Run("invalid value returns nil", func(t *testing.T) {
		t.Setenv("TEST_PTR", "invalid")
		result := getEnvInt32Ptr("TEST_PTR")
		assert.Nil(t, result)
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		t.Setenv("TEST_PTR", "")
		result := getEnvInt32Ptr("TEST_PTR")
		assert.Nil(t, result)
	})

	t.Run("negative value returns pointer", func(t *testing.T) {
		t.Setenv("TEST_PTR", "-5")
		result := getEnvInt32Ptr("TEST_PTR")
		require.NotNil(t, result)
		assert.Equal(t, int32(-5), *result)
	})
}

// TestLoadConfig_ValidResourceQuantities tests that valid resource quantities pass validation
func TestLoadConfig_ValidResourceQuantities(t *testing.T) {
	t.Setenv("POD_CPU_REQUEST", "500m")
	t.Setenv("POD_MEM_REQUEST", "256Mi")
	t.Setenv("POD_CPU_LIMIT", "2000m")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "500m", cfg.PodCPURequest)
	assert.Equal(t, "256Mi", cfg.PodMemRequest)
	assert.Equal(t, "2000m", cfg.PodCPULimit)
}

// TestLoadConfig_InvalidResourceQuantity tests that invalid resource quantities return error
func TestLoadConfig_InvalidResourceQuantity(t *testing.T) {
	t.Setenv("POD_CPU_REQUEST", "garbage")

	cfg, err := LoadConfig()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "POD_CPU_REQUEST")
	assert.Contains(t, err.Error(), "garbage")
}

// TestLoadConfig_MultipleInvalidResources tests that multiple invalid values are all reported
func TestLoadConfig_MultipleInvalidResources(t *testing.T) {
	t.Setenv("POD_CPU_REQUEST", "invalid-cpu")
	t.Setenv("POD_MEM_LIMIT", "bad-memory")
	t.Setenv("METRICS_EXPORTER_CPU_REQUEST", "wrong")

	cfg, err := LoadConfig()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "POD_CPU_REQUEST")
	assert.Contains(t, err.Error(), "POD_MEM_LIMIT")
	assert.Contains(t, err.Error(), "METRICS_EXPORTER_CPU_REQUEST")
}

// TestLoadConfig_EmptyResourceStrings tests that empty strings are treated as optional (not validated)
func TestLoadConfig_EmptyResourceStrings(t *testing.T) {
	// Clear all resource env vars to get empty defaults if any
	envVars := []string{
		"POD_CPU_REQUEST",
		"POD_MEM_REQUEST",
		"POD_EPHEMERAL_REQUEST",
		"POD_CPU_LIMIT",
		"POD_MEM_LIMIT",
		"POD_EPHEMERAL_LIMIT",
		"MASTER_POD_CPU_REQUEST",
		"MASTER_POD_MEM_REQUEST",
		"WORKER_POD_CPU_REQUEST",
		"WORKER_POD_MEM_REQUEST",
	}
	for _, env := range envVars {
		_ = os.Unsetenv(env)
	}

	// Set master/worker role-specific to empty explicitly
	t.Setenv("MASTER_POD_CPU_REQUEST", "")
	t.Setenv("WORKER_POD_CPU_REQUEST", "")

	cfg, err := LoadConfig()

	// Should succeed - empty strings are valid (mean "not set")
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

// TestLoadConfig_RoleSpecificResourceDefaults tests that role-specific resource fields default to empty
func TestLoadConfig_RoleSpecificResourceDefaults(t *testing.T) {
	// Clear all MASTER_POD_* and WORKER_POD_* env vars
	envVars := []string{
		"MASTER_POD_CPU_REQUEST",
		"MASTER_POD_MEM_REQUEST",
		"MASTER_POD_EPHEMERAL_REQUEST",
		"MASTER_POD_CPU_LIMIT",
		"MASTER_POD_MEM_LIMIT",
		"MASTER_POD_EPHEMERAL_LIMIT",
		"WORKER_POD_CPU_REQUEST",
		"WORKER_POD_MEM_REQUEST",
		"WORKER_POD_EPHEMERAL_REQUEST",
		"WORKER_POD_CPU_LIMIT",
		"WORKER_POD_MEM_LIMIT",
		"WORKER_POD_EPHEMERAL_LIMIT",
	}
	for _, env := range envVars {
		_ = os.Unsetenv(env)
	}

	cfg, err := LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Assert all 12 role-specific fields default to empty string
	assert.Equal(t, "", cfg.MasterCPURequest)
	assert.Equal(t, "", cfg.MasterMemRequest)
	assert.Equal(t, "", cfg.MasterEphemeralStorageRequest)
	assert.Equal(t, "", cfg.MasterCPULimit)
	assert.Equal(t, "", cfg.MasterMemLimit)
	assert.Equal(t, "", cfg.MasterEphemeralStorageLimit)
	assert.Equal(t, "", cfg.WorkerCPURequest)
	assert.Equal(t, "", cfg.WorkerMemRequest)
	assert.Equal(t, "", cfg.WorkerEphemeralStorageRequest)
	assert.Equal(t, "", cfg.WorkerCPULimit)
	assert.Equal(t, "", cfg.WorkerMemLimit)
	assert.Equal(t, "", cfg.WorkerEphemeralStorageLimit)
}

// TestLoadConfig_RoleSpecificResourceOverrides tests that role-specific env vars are loaded correctly
func TestLoadConfig_RoleSpecificResourceOverrides(t *testing.T) {
	// Set some role-specific env vars
	t.Setenv("MASTER_POD_CPU_REQUEST", "500m")
	t.Setenv("MASTER_POD_MEM_REQUEST", "512Mi")
	t.Setenv("WORKER_POD_CPU_LIMIT", "2000m")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Assert fields populated correctly
	assert.Equal(t, "500m", cfg.MasterCPURequest)
	assert.Equal(t, "512Mi", cfg.MasterMemRequest)
	assert.Equal(t, "2000m", cfg.WorkerCPULimit)

	// Assert unset fields remain empty string
	assert.Equal(t, "", cfg.MasterEphemeralStorageRequest)
	assert.Equal(t, "", cfg.MasterCPULimit)
	assert.Equal(t, "", cfg.WorkerCPURequest)
}

// TestLoadConfig_InvalidRoleSpecificResource tests that invalid role-specific quantities return error
func TestLoadConfig_InvalidRoleSpecificResource(t *testing.T) {
	t.Setenv("MASTER_POD_CPU_REQUEST", "garbage")

	cfg, err := LoadConfig()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "MASTER_POD_CPU_REQUEST")
}

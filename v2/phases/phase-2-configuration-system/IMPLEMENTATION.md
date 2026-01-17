# Phase 2: Configuration System - Implementation Plan

**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Prerequisites:** Phase 0 complete  
**Requirements:** §3.1 Configuration

---

## Objective

Implement environment-based configuration matching Java `SysConfig.java`. The configuration system provides operator-wide settings that resource builders and the reconciler use to configure generated Kubernetes resources.

---

## Reference: Java SysConfig Fields

From `/src/main/java/com/locust/operator/controller/config/SysConfig.java`:

```java
@Singleton
public class SysConfig {
    // Kafka configuration
    private String kafkaBootstrapServers;
    private boolean kafkaSecurityEnabled;
    private String kafkaSecurityProtocol;
    private String kafkaUsername;
    private String kafkaUserPassword;
    private String kafkaSaslMechanism;
    private String kafkaSaslJaasConfig;

    // Job configuration (nullable!)
    private Object ttlSecondsAfterFinished;

    // Pod resources
    private String podCpuRequest;
    private String podMemRequest;
    private String podEphemeralStorageRequest;
    private String podCpuLimit;
    private String podMemLimit;
    private String podEphemeralStorageLimit;

    // Metrics exporter
    private String metricsExporterImage;
    private Integer metricsExporterPort;
    private String metricsExporterPullPolicy;
    private String metricsExporterCpuRequest;
    private String metricsExporterMemRequest;
    private String metricsExporterEphemeralStorageRequest;
    private String metricsExporterCpuLimit;
    private String metricsExporterMemLimit;
    private String metricsExporterEphemeralStorageLimit;

    // Feature flags
    private boolean affinityCrInjectionEnabled;
    private boolean tolerationsCrInjectionEnabled;
}
```

---

## Reference: Java Default Values

From `/src/main/resources/application.yml`:

```yaml
config:
  load-generation-jobs:
    ttl-seconds-after-finished: ${JOB_TTL_SECONDS_AFTER_FINISHED:}  # null default!
  load-generation-pods:
    affinity:
      enableCrInjection: ${ENABLE_AFFINITY_CR_INJECTION:false}
    taintTolerations:
      enableCrInjection: ${ENABLE_TAINT_TOLERATIONS_CR_INJECTION:false}
    resource:
      cpu-request: ${POD_CPU_REQUEST:`250m`}
      mem-request: ${POD_MEM_REQUEST:`128Mi`}
      ephemeralStorage-request: ${POD_EPHEMERAL_REQUEST:`30M`}
      cpu-limit: ${POD_CPU_LIMIT:`1000m`}
      mem-limit: ${POD_MEM_LIMIT:`1024Mi`}
      ephemeralStorage-limit: ${POD_EPHEMERAL_LIMIT:`50M`}
    metricsExporter:
      image: ${METRICS_EXPORTER_IMAGE:`containersol/locust_exporter:v0.5.0`}
      port: ${METRICS_EXPORTER_PORT:`9646`}
      pullPolicy: ${METRICS_EXPORTER_IMAGE_PULL_POLICY:`Always`}
      resource:
        cpu-request: ${METRICS_EXPORTER_CPU_REQUEST:`250m`}
        mem-request: ${METRICS_EXPORTER_MEM_REQUEST:`128Mi`}
        ephemeralStorage-request: ${METRICS_EXPORTER_EPHEMERAL_REQUEST:`30M`}
        cpu-limit: ${METRICS_EXPORTER_CPU_LIMIT:`1000m`}
        mem-limit: ${METRICS_EXPORTER_MEM_LIMIT:`1024Mi`}
        ephemeralStorage-limit: ${METRICS_EXPORTER_EPHEMERAL_LIMIT:`50M`}
    kafka:
      bootstrap-servers: ${KAFKA_BOOTSTRAP_SERVERS:`localhost:9092`}
      security:
        enabled: ${KAFKA_SECURITY_ENABLED:`false`}
        protocol: ${KAFKA_SECURITY_PROTOCOL_CONFIG:`SASL_PLAINTEXT`}
        username: ${KAFKA_USERNAME:`localKafkaUser`}
        password: ${KAFKA_PASSWORD:`localKafkaPassword`}
      sasl:
        mechanism: ${KAFKA_SASL_MECHANISM:`SCRAM-SHA-512`}
        jaas:
          config: ${KAFKA_SASL_JAAS_CONFIG:`placeholder`}
```

---

## Tasks

### Task 2.1: Create Config Package Directory

```bash
cd locust-k8s-operator-go
mkdir -p internal/config
```

---

### Task 2.2: Implement Helper Functions

**File:** `internal/config/config.go`

Create helper functions for environment variable parsing:

```go
package config

import (
	"os"
	"strconv"
)

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
```

**Key Points:**
- `getEnvInt32Ptr` returns `nil` when env var is not set (critical for TTL behavior)
- Error handling returns default/nil rather than panicking
- Matches the Java `getTtlSecondsAfterFinished()` null behavior

---

### Task 2.3: Define OperatorConfig Struct

Add the config struct with all fields organized by category:

```go
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
```

---

### Task 2.4: Implement LoadConfig Function

```go
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
```

**Note:** Kafka username/password defaults are empty strings in Go (not placeholder values) for security.

---

### Task 2.5: Write Unit Tests

**File:** `internal/config/config_test.go`

```go
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
		"ENABLE_AFFINITY_CR_INJECTION",
		"ENABLE_TAINT_TOLERATIONS_CR_INJECTION",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	cfg := LoadConfig()

	// Job configuration
	assert.Nil(t, cfg.TTLSecondsAfterFinished, "TTL should be nil when not set")

	// Pod resources
	assert.Equal(t, "250m", cfg.PodCPURequest)
	assert.Equal(t, "128Mi", cfg.PodMemRequest)
	assert.Equal(t, "30M", cfg.PodEphemeralStorageRequest)
	assert.Equal(t, "1000m", cfg.PodCPULimit)
	assert.Equal(t, "1024Mi", cfg.PodMemLimit)
	assert.Equal(t, "50M", cfg.PodEphemeralStorageLimit)

	// Metrics exporter
	assert.Equal(t, "containersol/locust_exporter:v0.5.0", cfg.MetricsExporterImage)
	assert.Equal(t, int32(9646), cfg.MetricsExporterPort)
	assert.Equal(t, "Always", cfg.MetricsExporterPullPolicy)

	// Feature flags
	assert.False(t, cfg.EnableAffinityCRInjection)
	assert.False(t, cfg.EnableTolerationsCRInjection)
}

func TestLoadConfig_EnvironmentOverrides(t *testing.T) {
	// Set up test env vars
	t.Setenv("JOB_TTL_SECONDS_AFTER_FINISHED", "300")
	t.Setenv("POD_CPU_REQUEST", "500m")
	t.Setenv("POD_MEM_REQUEST", "256Mi")
	t.Setenv("METRICS_EXPORTER_IMAGE", "custom/exporter:v1.0.0")
	t.Setenv("METRICS_EXPORTER_PORT", "9000")
	t.Setenv("ENABLE_AFFINITY_CR_INJECTION", "true")
	t.Setenv("ENABLE_TAINT_TOLERATIONS_CR_INJECTION", "true")

	cfg := LoadConfig()

	// Verify overrides
	require.NotNil(t, cfg.TTLSecondsAfterFinished)
	assert.Equal(t, int32(300), *cfg.TTLSecondsAfterFinished)
	assert.Equal(t, "500m", cfg.PodCPURequest)
	assert.Equal(t, "256Mi", cfg.PodMemRequest)
	assert.Equal(t, "custom/exporter:v1.0.0", cfg.MetricsExporterImage)
	assert.Equal(t, int32(9000), cfg.MetricsExporterPort)
	assert.True(t, cfg.EnableAffinityCRInjection)
	assert.True(t, cfg.EnableTolerationsCRInjection)
}

func TestLoadConfig_TTLSecondsAfterFinished_ZeroValue(t *testing.T) {
	t.Setenv("JOB_TTL_SECONDS_AFTER_FINISHED", "0")

	cfg := LoadConfig()

	// TTL of 0 should be a valid value (immediate cleanup)
	require.NotNil(t, cfg.TTLSecondsAfterFinished)
	assert.Equal(t, int32(0), *cfg.TTLSecondsAfterFinished)
}

func TestLoadConfig_KafkaConfiguration(t *testing.T) {
	t.Setenv("KAFKA_BOOTSTRAP_SERVERS", "kafka.example.com:9092")
	t.Setenv("KAFKA_SECURITY_ENABLED", "true")
	t.Setenv("KAFKA_USERNAME", "user")
	t.Setenv("KAFKA_PASSWORD", "secret")

	cfg := LoadConfig()

	assert.Equal(t, "kafka.example.com:9092", cfg.KafkaBootstrapServers)
	assert.True(t, cfg.KafkaSecurityEnabled)
	assert.Equal(t, "user", cfg.KafkaUsername)
	assert.Equal(t, "secret", cfg.KafkaPassword)
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "value")

	assert.Equal(t, "value", getEnv("TEST_VAR", "default"))
	assert.Equal(t, "default", getEnv("NONEXISTENT_VAR", "default"))
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"true string", "true", false, true},
		{"false string", "false", true, false},
		{"1 value", "1", false, true},
		{"0 value", "0", true, false},
		{"invalid value", "invalid", true, true}, // returns default on error
		{"empty uses default true", "", true, true},
		{"empty uses default false", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv("TEST_BOOL", tt.envValue)
			} else {
				os.Unsetenv("TEST_BOOL")
			}
			result := getEnvBool("TEST_BOOL", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt32(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	assert.Equal(t, int32(42), getEnvInt32("TEST_INT", 0))

	os.Unsetenv("TEST_INT")
	assert.Equal(t, int32(100), getEnvInt32("TEST_INT", 100))

	t.Setenv("TEST_INT", "invalid")
	assert.Equal(t, int32(100), getEnvInt32("TEST_INT", 100)) // returns default on error
}

func TestGetEnvInt32Ptr(t *testing.T) {
	// Not set - should return nil
	os.Unsetenv("TEST_PTR")
	result := getEnvInt32Ptr("TEST_PTR")
	assert.Nil(t, result)

	// Set to valid value
	t.Setenv("TEST_PTR", "42")
	result = getEnvInt32Ptr("TEST_PTR")
	require.NotNil(t, result)
	assert.Equal(t, int32(42), *result)

	// Set to invalid value - should return nil
	t.Setenv("TEST_PTR", "invalid")
	result = getEnvInt32Ptr("TEST_PTR")
	assert.Nil(t, result)
}
```

---

## Complete config.go

Here's the complete file:

```go
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
```

---

## Acceptance Criteria Verification

| Criteria | Verification | Expected |
|----------|--------------|----------|
| Config loads with defaults | `TestLoadConfig_DefaultValues` | All defaults match `application.yml` |
| Config respects overrides | `TestLoadConfig_EnvironmentOverrides` | Env vars override defaults |
| TTL nil when not set | `TestLoadConfig_DefaultValues` | `TTLSecondsAfterFinished` is `nil` |
| TTL 0 is valid | `TestLoadConfig_TTLSecondsAfterFinished_ZeroValue` | Pointer to `0`, not `nil` |
| All fields present | Struct definition | Matches Java `SysConfig.java` |

---

## Usage Example

In `cmd/main.go` (Phase 4):

```go
func main() {
    // Load operator configuration
    cfg := config.LoadConfig()
    
    // Pass to reconciler
    reconciler := &controller.LocustTestReconciler{
        Client:   mgr.GetClient(),
        Scheme:   mgr.GetScheme(),
        Config:   cfg,
        Recorder: mgr.GetEventRecorderFor("locust-controller"),
    }
}
```

In resource builders (Phase 3):

```go
func BuildMasterJob(lt *v1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
    job := &batchv1.Job{
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: cfg.TTLSecondsAfterFinished,
            // ...
        },
    }
    return job
}
```

---

## Troubleshooting

### Issue: TTL is always nil
**Symptom:** `TTLSecondsAfterFinished` is nil even when env var is set
**Solution:** Check env var name is exactly `JOB_TTL_SECONDS_AFTER_FINISHED`, check value is a valid integer

### Issue: Boolean parsing fails
**Symptom:** Feature flags always return default
**Solution:** Use "true"/"false" or "1"/"0" as env var values

### Issue: Resource quantities invalid
**Symptom:** Kubernetes rejects pod specs
**Solution:** Ensure values like "250m", "128Mi" are valid Kubernetes resource quantities

---

## Notes for Next Phase

Phase 3 (Resource Builders) will use `OperatorConfig` to:
1. Set `TTLSecondsAfterFinished` on Jobs
2. Configure resource requests/limits on containers
3. Build metrics exporter sidecar container
4. Inject Kafka environment variables when enabled
5. Apply affinity/tolerations based on feature flags

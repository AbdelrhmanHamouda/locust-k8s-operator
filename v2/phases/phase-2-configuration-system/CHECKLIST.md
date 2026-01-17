# Phase 2: Configuration System - Checklist

**Estimated Effort:** 0.5 day  
**Status:** Complete

---

## Pre-Implementation

- [ ] Phase 0 complete (project scaffolds exist)
- [ ] `internal/config/` directory does not exist yet
- [ ] Java `SysConfig.java` reviewed for all config fields
- [ ] Java `application.yml` reviewed for all default values

---

## Task 2.1: Create Config Directory Structure

- [ ] Create `internal/config/` directory
- [ ] Create `internal/config/config.go` file
- [ ] Create `internal/config/config_test.go` file

---

## Task 2.2: Define OperatorConfig Struct

Job configuration:
- [ ] `TTLSecondsAfterFinished *int32` (nullable)

Pod resource configuration:
- [ ] `PodCPURequest string` (default: "250m")
- [ ] `PodMemRequest string` (default: "128Mi")
- [ ] `PodEphemeralStorageRequest string` (default: "30M")
- [ ] `PodCPULimit string` (default: "1000m")
- [ ] `PodMemLimit string` (default: "1024Mi")
- [ ] `PodEphemeralStorageLimit string` (default: "50M")

Metrics exporter configuration:
- [ ] `MetricsExporterImage string` (default: "containersol/locust_exporter:v0.5.0")
- [ ] `MetricsExporterPort int32` (default: 9646)
- [ ] `MetricsExporterPullPolicy string` (default: "Always")
- [ ] `MetricsExporterCPURequest string` (default: "250m")
- [ ] `MetricsExporterMemRequest string` (default: "128Mi")
- [ ] `MetricsExporterEphemeralStorageRequest string` (default: "30M")
- [ ] `MetricsExporterCPULimit string` (default: "1000m")
- [ ] `MetricsExporterMemLimit string` (default: "1024Mi")
- [ ] `MetricsExporterEphemeralStorageLimit string` (default: "50M")

Kafka configuration:
- [ ] `KafkaBootstrapServers string` (default: "localhost:9092")
- [ ] `KafkaSecurityEnabled bool` (default: false)
- [ ] `KafkaSecurityProtocol string` (default: "SASL_PLAINTEXT")
- [ ] `KafkaUsername string` (default: "")
- [ ] `KafkaPassword string` (default: "")
- [ ] `KafkaSaslMechanism string` (default: "SCRAM-SHA-512")
- [ ] `KafkaSaslJaasConfig string` (default: "")

Feature flags:
- [ ] `EnableAffinityCRInjection bool` (default: false)
- [ ] `EnableTolerationsCRInjection bool` (default: false)

---

## Task 2.3: Implement Helper Functions

- [ ] `getEnv(key, defaultValue string) string`
- [ ] `getEnvBool(key string, defaultValue bool) bool`
- [ ] `getEnvInt32(key string, defaultValue int32) int32`
- [ ] `getEnvInt32Ptr(key string) *int32`

---

## Task 2.4: Implement LoadConfig Function

- [ ] Create `LoadConfig() *OperatorConfig` function
- [ ] Load all fields from environment variables
- [ ] Apply default values for unset variables
- [ ] Handle nullable `TTLSecondsAfterFinished` correctly

---

## Task 2.5: Environment Variable Naming Convention

Map Java property names to Go environment variable names:

| Java Property | Environment Variable |
|---------------|---------------------|
| `config.load-generation-jobs.ttl-seconds-after-finished` | `JOB_TTL_SECONDS_AFTER_FINISHED` |
| `config.load-generation-pods.resource.cpu-request` | `POD_CPU_REQUEST` |
| `config.load-generation-pods.resource.mem-request` | `POD_MEM_REQUEST` |
| `config.load-generation-pods.resource.ephemeralStorage-request` | `POD_EPHEMERAL_REQUEST` |
| `config.load-generation-pods.resource.cpu-limit` | `POD_CPU_LIMIT` |
| `config.load-generation-pods.resource.mem-limit` | `POD_MEM_LIMIT` |
| `config.load-generation-pods.resource.ephemeralStorage-limit` | `POD_EPHEMERAL_LIMIT` |
| `config.load-generation-pods.metricsExporter.image` | `METRICS_EXPORTER_IMAGE` |
| `config.load-generation-pods.metricsExporter.port` | `METRICS_EXPORTER_PORT` |
| `config.load-generation-pods.metricsExporter.pullPolicy` | `METRICS_EXPORTER_IMAGE_PULL_POLICY` |
| `config.load-generation-pods.metricsExporter.resource.cpu-request` | `METRICS_EXPORTER_CPU_REQUEST` |
| `config.load-generation-pods.metricsExporter.resource.mem-request` | `METRICS_EXPORTER_MEM_REQUEST` |
| `config.load-generation-pods.metricsExporter.resource.ephemeralStorage-request` | `METRICS_EXPORTER_EPHEMERAL_REQUEST` |
| `config.load-generation-pods.metricsExporter.resource.cpu-limit` | `METRICS_EXPORTER_CPU_LIMIT` |
| `config.load-generation-pods.metricsExporter.resource.mem-limit` | `METRICS_EXPORTER_MEM_LIMIT` |
| `config.load-generation-pods.metricsExporter.resource.ephemeralStorage-limit` | `METRICS_EXPORTER_EPHEMERAL_LIMIT` |
| `config.load-generation-pods.kafka.bootstrap-servers` | `KAFKA_BOOTSTRAP_SERVERS` |
| `config.load-generation-pods.kafka.security.enabled` | `KAFKA_SECURITY_ENABLED` |
| `config.load-generation-pods.kafka.security.protocol` | `KAFKA_SECURITY_PROTOCOL_CONFIG` |
| `config.load-generation-pods.kafka.security.username` | `KAFKA_USERNAME` |
| `config.load-generation-pods.kafka.security.password` | `KAFKA_PASSWORD` |
| `config.load-generation-pods.kafka.sasl.mechanism` | `KAFKA_SASL_MECHANISM` |
| `config.load-generation-pods.kafka.sasl.jaas.config` | `KAFKA_SASL_JAAS_CONFIG` |
| `config.load-generation-pods.affinity.enableCrInjection` | `ENABLE_AFFINITY_CR_INJECTION` |
| `config.load-generation-pods.taintTolerations.enableCrInjection` | `ENABLE_TAINT_TOLERATIONS_CR_INJECTION` |

---

## Task 2.6: Write Unit Tests

Default value tests:
- [ ] Test `LoadConfig()` returns defaults when no env vars set
- [ ] Verify each default value matches Java `application.yml`

Override tests:
- [ ] Test each field can be overridden via env var
- [ ] Test `TTLSecondsAfterFinished` is nil when not set
- [ ] Test `TTLSecondsAfterFinished` is pointer to value when set
- [ ] Test boolean parsing (true/false/1/0)
- [ ] Test int32 parsing

Edge cases:
- [ ] Test empty string env var behavior
- [ ] Test invalid int value behavior (graceful handling)

---

## Task 2.7: Verify Build

- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/config/...` passes
- [ ] No linting errors: `golangci-lint run`

---

## Post-Implementation Verification

- [ ] All config fields documented with comments
- [ ] Default values match Java `application.yml`
- [ ] Tests achieve >90% coverage for config package
- [ ] Config can be imported by controller package

---

## Completion

- [ ] Update `phases/NOTES.md` with Phase 2 completion notes
- [ ] Document any deviations from plan
- [ ] Note any issues discovered for future phases

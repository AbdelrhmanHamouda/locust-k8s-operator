# Phase Implementation Notes

---

# Phase 0 Completion Notes

**Completed:** 2026-01-16

---

## Key Decisions Made

### Domain Configuration
- **Domain**: `io` (not `locust.io`)
- **Group**: `locust`
- **Result**: CRD group is `locust.io`, matching the existing Java CRD

This was corrected during implementation. The original plan used `--domain locust.io` which resulted in `locust.locust.io`. The fix was to use `--domain io --group locust`.

---

## Versions Installed

| Tool | Version |
|------|---------|
| Operator SDK | 1.42.0 |
| Go | 1.24.0 |
| controller-runtime | 0.21.0 |
| controller-gen | 0.18.0 |
| k8s.io/api | 0.33.0 |
| k8s.io/apimachinery | 0.33.0 |
| k8s.io/client-go | 0.33.0 |

---

# Phase 1 Completion Notes

**Completed:** 2026-01-17

---

## Summary

Implemented v1 API types that exactly match the Java CRD schema for backward compatibility.

## Files Created/Modified

- `api/v1/locusttest_types.go` - Full v1 API types with all fields and kubebuilder markers
- `api/v1/locusttest_types_test.go` - Unit tests for JSON marshaling and field names
- `api/v1/zz_generated.deepcopy.go` - Auto-generated DeepCopy methods
- `config/crd/bases/locust.io_locusttests.yaml` - Generated CRD manifest
- `internal/controller/locusttest_controller_test.go` - Updated to use new spec fields

## Types Defined

| Type | Description |
|------|-------------|
| `LocustTestSpec` | Main spec with 12 fields matching Java CRD |
| `LocustTestStatus` | Empty status (to be populated in Phase 9) |
| `PodLabels` | Master/Worker label maps |
| `PodAnnotations` | Master/Worker annotation maps |
| `LocustTestAffinity` | NodeAffinity wrapper |
| `LocustTestNodeAffinity` | RequiredDuringSchedulingIgnoredDuringExecution map |
| `LocustTestToleration` | Key, Operator, Value, Effect fields |

## Validation Markers Applied

- **Required fields**: `masterCommandSeed`, `workerCommandSeed`, `workerReplicas`, `image`
- **WorkerReplicas constraints**: min=1, max=500, default=1
- **Enums**: ImagePullPolicy (Always/IfNotPresent/Never), Toleration.Operator (Exists/Equal), Toleration.Effect (NoSchedule/PreferNoSchedule/NoExecute)

## CRD Features

- Short name: `lotest`
- Printer columns: master_cmd, worker_replica_count, Image, Age
- Status subresource enabled

## Verification

- `make build` ✓
- `make test` ✓
- `go test ./api/v1/... -v` ✓ (4 tests pass)

## Notes for Phase 2

1. The controller test now uses valid spec fields - any future controller changes should maintain this.
2. Status is empty - Phase 9 will add status fields.
3. The generated CRD at `config/crd/bases/locust.io_locusttests.yaml` is schema-compatible with the Java CRD.

---

# Phase 2 Completion Notes

**Completed:** 2026-01-17

---

## Summary

Implemented environment-based configuration system matching Java `SysConfig.java`. The configuration system provides operator-wide settings that resource builders and the reconciler will use.

## Files Created

- `internal/config/config.go` - OperatorConfig struct and LoadConfig function
- `internal/config/config_test.go` - Comprehensive unit tests with 100% coverage

## Configuration Fields Implemented

| Category | Fields |
|----------|--------|
| **Job** | TTLSecondsAfterFinished (*int32 - nullable) |
| **Pod Resources** | PodCPURequest, PodMemRequest, PodEphemeralStorageRequest, PodCPULimit, PodMemLimit, PodEphemeralStorageLimit |
| **Metrics Exporter** | MetricsExporterImage, MetricsExporterPort, MetricsExporterPullPolicy, + CPU/Mem/Ephemeral request/limit |
| **Kafka** | KafkaBootstrapServers, KafkaSecurityEnabled, KafkaSecurityProtocol, KafkaUsername, KafkaPassword, KafkaSaslMechanism, KafkaSaslJaasConfig |
| **Feature Flags** | EnableAffinityCRInjection, EnableTolerationsCRInjection |

## Helper Functions

- `getEnv(key, defaultValue string) string` - String env var with default
- `getEnvBool(key string, defaultValue bool) bool` - Boolean parsing with default
- `getEnvInt32(key string, defaultValue int32) int32` - Int32 parsing with default
- `getEnvInt32Ptr(key string) *int32` - Nullable int32 for TTL (nil when unset)

## Key Design Decisions

1. **Kafka credentials default to empty strings** - Unlike Java which uses placeholder values, Go defaults to empty for security
2. **TTLSecondsAfterFinished uses *int32** - Correctly distinguishes "not set" (nil) from "set to 0"
3. **No DI framework** - Go idiom using explicit struct wiring via LoadConfig()

## Verification

- `go build ./...` ✓
- `go test ./internal/config/... -v -cover` ✓ (100% coverage)
- `golangci-lint run ./internal/config/...` ✓ (0 issues)
- `make test` ✓

## Notes for Phase 3

1. Use `config.LoadConfig()` to get operator configuration
2. Pass `*OperatorConfig` to resource builders
3. Use `cfg.TTLSecondsAfterFinished` directly on Job specs (nil-safe)
4. Resource quantities (e.g., "250m", "128Mi") are strings - parse with `resource.MustParse()` when building K8s resources

# Phase 12: OpenTelemetry Support - Checklist

**Estimated Effort:** 1 day  
**Status:** Not Started

---

## Pre-Implementation

- [ ] Phase 7 complete (v2 API Types with `ObservabilityConfig`, `OpenTelemetryConfig` defined)
- [ ] Review `api/v2/locusttest_types.go` for OTel types
- [ ] Review `analysis/LOCUST_FEATURES.md` §1.1 for Locust OTel details
- [ ] Review current `internal/resources/command.go` implementation
- [ ] Review current `internal/resources/job.go` implementation
- [ ] Run `make build && make test` to ensure clean starting state

---

## Task 12.1: Create OTel Helper Functions

**File:** `internal/resources/otel.go`

### Constants

- [ ] Create file `internal/resources/otel.go`
- [ ] Add copyright header
- [ ] Define OTel environment variable constants:
  ```go
  const (
      EnvOTelTracesExporter     = "OTEL_TRACES_EXPORTER"
      EnvOTelMetricsExporter    = "OTEL_METRICS_EXPORTER"
      EnvOTelExporterEndpoint   = "OTEL_EXPORTER_OTLP_ENDPOINT"
      EnvOTelExporterProtocol   = "OTEL_EXPORTER_OTLP_PROTOCOL"
      EnvOTelExporterInsecure   = "OTEL_EXPORTER_OTLP_INSECURE"
  )
  ```
- [ ] Define default OTel values:
  ```go
  const (
      OTelExporterOTLP = "otlp"
      OTelProtocolGRPC = "grpc"
  )
  ```

### Core Functions

- [ ] Implement `IsOTelEnabled()` function:
  ```go
  func IsOTelEnabled(lt *locustv2.LocustTest) bool
  ```
  - [ ] Return false if `observability` is nil
  - [ ] Return false if `openTelemetry` is nil
  - [ ] Return the value of `Enabled`

- [ ] Implement `GetOTelConfig()` function:
  ```go
  func GetOTelConfig(lt *locustv2.LocustTest) *locustv2.OpenTelemetryConfig
  ```
  - [ ] Return nil if not configured
  - [ ] Return the OpenTelemetry config

- [ ] Implement `BuildOTelEnvVars()` function:
  ```go
  func BuildOTelEnvVars(lt *locustv2.LocustTest) []corev1.EnvVar
  ```
  - [ ] Return nil if OTel is not enabled
  - [ ] Add `OTEL_TRACES_EXPORTER=otlp`
  - [ ] Add `OTEL_METRICS_EXPORTER=otlp`
  - [ ] Add `OTEL_EXPORTER_OTLP_ENDPOINT` from spec
  - [ ] Add `OTEL_EXPORTER_OTLP_PROTOCOL` (default: grpc)
  - [ ] Add `OTEL_EXPORTER_OTLP_INSECURE` if true
  - [ ] Add extra env vars from `extraEnvVars` map

**Verification:**
```bash
go build ./internal/resources/...
```

---

## Task 12.2: Write OTel Helper Tests

**File:** `internal/resources/otel_test.go`

### IsOTelEnabled Tests

- [ ] Create file `internal/resources/otel_test.go`
- [ ] Add copyright header
- [ ] `TestIsOTelEnabled_NilObservability` - Returns false when observability is nil
- [ ] `TestIsOTelEnabled_NilOpenTelemetry` - Returns false when openTelemetry is nil
- [ ] `TestIsOTelEnabled_Disabled` - Returns false when enabled is false
- [ ] `TestIsOTelEnabled_Enabled` - Returns true when enabled is true

### GetOTelConfig Tests

- [ ] `TestGetOTelConfig_NilObservability` - Returns nil
- [ ] `TestGetOTelConfig_NilOpenTelemetry` - Returns nil
- [ ] `TestGetOTelConfig_HasConfig` - Returns the config

### BuildOTelEnvVars Tests

- [ ] `TestBuildOTelEnvVars_Disabled` - Returns nil when OTel disabled
- [ ] `TestBuildOTelEnvVars_EnabledMinimal` - Core env vars with endpoint
- [ ] `TestBuildOTelEnvVars_FullConfig` - All env vars including protocol, insecure
- [ ] `TestBuildOTelEnvVars_ExtraEnvVars` - Extra env vars merged correctly
- [ ] `TestBuildOTelEnvVars_DefaultProtocol` - Defaults to grpc when not specified
- [ ] `TestBuildOTelEnvVars_HTTPProtocol` - Uses http/protobuf when specified
- [ ] `TestBuildOTelEnvVars_InsecureFalse` - Does not add insecure env var when false

**Verification:**
```bash
go test ./internal/resources/... -v -run TestIsOTelEnabled
go test ./internal/resources/... -v -run TestGetOTelConfig
go test ./internal/resources/... -v -run TestBuildOTelEnvVars
go test ./internal/resources/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep otel
# Target: ≥90% coverage for otel.go
```

---

## Task 12.3: Update Command Builders

**File:** `internal/resources/command.go`

### Signature Updates

- [ ] Update `BuildMasterCommand()` signature to include OTel flag:
  ```go
  func BuildMasterCommand(commandSeed string, workerReplicas int32, otelEnabled bool) []string
  ```

- [ ] Update `BuildWorkerCommand()` signature to include OTel flag:
  ```go
  func BuildWorkerCommand(commandSeed string, masterHost string, otelEnabled bool) []string
  ```

### Implementation Updates

- [ ] Update `BuildMasterCommand()`:
  - [ ] Add `--otel` flag after command seed when `otelEnabled` is true
  - [ ] Maintain existing flag order for other arguments

- [ ] Update `BuildWorkerCommand()`:
  - [ ] Add `--otel` flag after command seed when `otelEnabled` is true
  - [ ] Maintain existing flag order for other arguments

**Verification:**
```bash
go build ./internal/resources/...
```

---

## Task 12.4: Update Command Builder Tests

**File:** `internal/resources/command_test.go`

### Existing Test Updates

- [ ] Update existing `TestBuildMasterCommand` tests to pass `otelEnabled: false`
- [ ] Update existing `TestBuildWorkerCommand` tests to pass `otelEnabled: false`
- [ ] Verify all existing tests still pass

### New Test Cases

- [ ] `TestBuildMasterCommand_OTelDisabled` - No --otel flag in output
- [ ] `TestBuildMasterCommand_OTelEnabled` - --otel flag present after seed
- [ ] `TestBuildMasterCommand_OTelFlagPosition` - --otel appears before --master
- [ ] `TestBuildWorkerCommand_OTelDisabled` - No --otel flag in output
- [ ] `TestBuildWorkerCommand_OTelEnabled` - --otel flag present after seed
- [ ] `TestBuildWorkerCommand_OTelFlagPosition` - --otel appears before --worker

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuildMasterCommand
go test ./internal/resources/... -v -run TestBuildWorkerCommand
```

---

## Task 12.5: Update Job Builder

**File:** `internal/resources/job.go`

### Call Site Updates

- [ ] Update `BuildMasterJob()`:
  - [ ] Get OTel enabled status: `otelEnabled := IsOTelEnabled(lt)`
  - [ ] Pass `otelEnabled` to `BuildMasterCommand()`

- [ ] Update `BuildWorkerJob()`:
  - [ ] Get OTel enabled status: `otelEnabled := IsOTelEnabled(lt)`
  - [ ] Pass `otelEnabled` to `BuildWorkerCommand()`

### Conditional Sidecar

- [ ] Update `buildJob()` to conditionally skip metrics sidecar:
  ```go
  // Master gets the metrics exporter sidecar ONLY if OTel is disabled
  if mode == Master && !IsOTelEnabled(lt) {
      containers = append(containers, buildMetricsExporterContainer(cfg))
  }
  ```

### Environment Variable Injection

- [ ] Update `BuildEnvVars()` (in `env.go`) to include OTel env vars:
  ```go
  // Add OTel environment variables if enabled
  otelEnvVars := BuildOTelEnvVars(lt)
  if len(otelEnvVars) > 0 {
      envVars = append(envVars, otelEnvVars...)
  }
  ```

**Verification:**
```bash
go build ./internal/resources/...
make test
```

---

## Task 12.6: Update Service Builder

**File:** `internal/resources/service.go`

### Conditional Metrics Port

- [ ] Update `BuildMasterService()` to conditionally exclude metrics port:
  ```go
  // Add metrics port ONLY if OTel is disabled (sidecar will be deployed)
  if !IsOTelEnabled(lt) {
      servicePorts = append(servicePorts, corev1.ServicePort{
          Name:     MetricsPortName,
          Protocol: corev1.ProtocolTCP,
          Port:     cfg.MetricsExporterPort,
      })
  }
  ```

**Verification:**
```bash
go build ./internal/resources/...
```

---

## Task 12.7: Update Service Builder Tests

**File:** `internal/resources/service_test.go`

### Service Port Tests

- [ ] `TestBuildMasterService_OTelDisabled_HasMetricsPort` - Metrics port present when OTel disabled
- [ ] `TestBuildMasterService_OTelEnabled_NoMetricsPort` - Metrics port absent when OTel enabled
- [ ] `TestBuildMasterService_NoObservability_HasMetricsPort` - Metrics port present when no config

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuildMasterService_OTel
```

---

## Task 12.8: Update Job Builder Tests

**File:** `internal/resources/job_test.go`

### Sidecar Tests

- [ ] `TestBuildMasterJob_OTelDisabled_HasSidecar` - Metrics sidecar present when OTel disabled
- [ ] `TestBuildMasterJob_OTelEnabled_NoSidecar` - No sidecar when OTel enabled
- [ ] `TestBuildMasterJob_NoObservability_HasSidecar` - Sidecar present when no observability config
- [ ] `TestBuildWorkerJob_OTelEnabled_NoSidecar` - Workers never have sidecar (verify no change)

### Environment Variable Tests

- [ ] `TestBuildMasterJob_OTelEnabled_HasEnvVars` - OTel env vars in master container
- [ ] `TestBuildWorkerJob_OTelEnabled_HasEnvVars` - OTel env vars in worker containers
- [ ] `TestBuildJob_OTelEnabled_EnvVarValues` - Verify correct env var values

### Command Flag Tests

- [ ] `TestBuildMasterJob_OTelEnabled_CommandHasFlag` - Command includes --otel flag
- [ ] `TestBuildWorkerJob_OTelEnabled_CommandHasFlag` - Command includes --otel flag

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuildMasterJob_OTel
go test ./internal/resources/... -v -run TestBuildWorkerJob_OTel
```

---

## Task 12.9: Update Validation Webhook

**File:** `api/v2/locusttest_webhook.go`

### Validation Function

- [ ] Implement `validateOTelConfig()`:
  ```go
  func (r *LocustTest) validateOTelConfig() (admission.Warnings, error)
  ```
  - [ ] Return nil if observability is nil
  - [ ] Return nil if openTelemetry is nil
  - [ ] Return error if `enabled: true` but `endpoint` is empty
  - [ ] Optionally validate protocol value (already has kubebuilder enum)

### Update Validators

- [ ] Update `ValidateCreate()` to call `validateOTelConfig()`
- [ ] Update `ValidateUpdate()` to call `validateOTelConfig()`

**Verification:**
```bash
go build ./api/v2/...
make manifests
```

---

## Task 12.10: Write Webhook Validation Tests

**File:** `api/v2/locusttest_webhook_test.go`

### OTel Validation Tests

- [ ] `TestValidateOTelConfig_NoObservability` - Passes when observability is nil
- [ ] `TestValidateOTelConfig_NoOpenTelemetry` - Passes when openTelemetry is nil
- [ ] `TestValidateOTelConfig_Disabled` - Passes when OTel disabled
- [ ] `TestValidateOTelConfig_EnabledWithEndpoint` - Passes when enabled with endpoint
- [ ] `TestValidateOTelConfig_EnabledNoEndpoint` - Fails when enabled without endpoint
- [ ] `TestValidateOTelConfig_ValidProtocolGRPC` - Passes for grpc protocol
- [ ] `TestValidateOTelConfig_ValidProtocolHTTP` - Passes for http/protobuf protocol

### Integration with Existing Tests

- [ ] Verify existing webhook tests still pass
- [ ] Verify combined validation works correctly

**Verification:**
```bash
go test ./api/v2/... -v -run TestValidateOTelConfig
go test ./api/v2/... -v -run TestValidate
```

---

## Task 12.11: Create Sample CR

**File:** `config/samples/locust_v2_locusttest_with_otel.yaml`

- [ ] Create sample CR demonstrating OTel configuration:
  ```yaml
  apiVersion: locust.io/v2
  kind: LocustTest
  metadata:
    name: load-test-with-otel
    labels:
      app.kubernetes.io/name: locusttest
      app.kubernetes.io/instance: load-test-with-otel
  spec:
    image: locustio/locust:2.32.0
    master:
      command: "locust -f /lotest/src/locustfile.py"
    worker:
      command: "locust -f /lotest/src/locustfile.py"
      replicas: 2
    testFiles:
      configMapRef: locust-scripts
    observability:
      openTelemetry:
        enabled: true
        endpoint: "otel-collector.monitoring:4317"
        protocol: "grpc"
        insecure: true
        extraEnvVars:
          OTEL_RESOURCE_ATTRIBUTES: "service.name=locust-load-test"
  ```

---

## Verification

### Code Generation

```bash
make generate
make manifests
```

- [ ] No errors during generation
- [ ] CRD updated with any new validation markers

### Build & Test

```bash
make build
make test
```

- [ ] Project builds without errors
- [ ] All existing tests pass
- [ ] New tests pass
- [ ] Coverage ≥80% for `internal/resources/otel.go`
- [ ] Coverage ≥80% for new webhook validation

### Linting

```bash
golangci-lint run ./internal/resources/...
golangci-lint run ./api/v2/...
```

- [ ] No linting errors

### Manual Verification (Optional)

```bash
# Apply sample CR
kubectl apply -f config/samples/locust_v2_locusttest_with_otel.yaml

# Verify --otel flag in master command
kubectl get pods -l performance-test-name=load-test-with-otel
MASTER_POD=$(kubectl get pods -l performance-test-name=load-test-with-otel,lotest-role=master -o jsonpath='{.items[0].metadata.name}')
kubectl get pod $MASTER_POD -o jsonpath='{.spec.containers[0].args}' | grep -- --otel
# Expected: --otel flag present

# Verify OTel env vars
kubectl exec -it $MASTER_POD -- env | grep OTEL
# Expected: OTEL_TRACES_EXPORTER, OTEL_METRICS_EXPORTER, OTEL_EXPORTER_OTLP_ENDPOINT, etc.

# Verify NO metrics exporter sidecar
kubectl get pod $MASTER_POD -o jsonpath='{.spec.containers[*].name}'
# Expected: only "<name>-master" (no "locust-exporter")

# Test validation webhook rejection
cat <<EOF | kubectl apply -f -
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: invalid-otel
spec:
  image: locustio/locust:2.32.0
  master:
    command: "locust"
  worker:
    command: "locust"
    replicas: 1
  observability:
    openTelemetry:
      enabled: true
      # endpoint missing - should fail
EOF
# Expected: Rejected with "endpoint is required when OpenTelemetry is enabled" error

# Test backward compatibility (no OTel config)
cat <<EOF | kubectl apply -f -
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: legacy-test
spec:
  image: locustio/locust:2.32.0
  master:
    command: "locust -f /lotest/src/locustfile.py"
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 1
  testFiles:
    configMapRef: locust-scripts
EOF
# Expected: Sidecar present, no --otel flag

# Cleanup
kubectl delete locusttest load-test-with-otel legacy-test --ignore-not-found
```

---

## Post-Implementation

- [ ] All verification steps pass
- [ ] Update `phases/README.md` with Phase 12 status
- [ ] Update `phases/NOTES.md` with implementation notes
- [ ] Update `ROADMAP.md` Phase 12 tasks as complete (optional)

---

## Files Summary

| File | Action | Est. LOC |
|------|--------|----------|
| `internal/resources/otel.go` | Create | ~80 |
| `internal/resources/otel_test.go` | Create | ~200 |
| `internal/resources/command.go` | Modify | ~15 |
| `internal/resources/command_test.go` | Modify | +60 |
| `internal/resources/job.go` | Modify | ~10 |
| `internal/resources/job_test.go` | Modify | +100 |
| `internal/resources/service.go` | Modify | ~10 |
| `internal/resources/service_test.go` | Modify | +40 |
| `internal/resources/env.go` | Modify | +5 |
| `api/v2/locusttest_webhook.go` | Modify | +30 |
| `api/v2/locusttest_webhook_test.go` | Modify | +60 |
| `config/samples/locust_v2_locusttest_with_otel.yaml` | Create | ~25 |

**Total Estimated:** ~635 LOC

---

## Quick Reference Commands

```bash
# Build and test
make generate
make manifests
make build
make test

# Run specific tests
go test ./internal/resources/... -v -run TestIsOTelEnabled
go test ./internal/resources/... -v -run TestBuildOTelEnvVars
go test ./internal/resources/... -v -run TestBuildMasterCommand_OTel
go test ./internal/resources/... -v -run TestBuildMasterJob_OTel
go test ./api/v2/... -v -run TestValidateOTelConfig

# Check coverage
go test ./internal/resources/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep otel

# Lint
golangci-lint run ./internal/resources/...
golangci-lint run ./api/v2/...
```

---

## Acceptance Criteria Summary

1. **OTel Flag:** `--otel` flag added to master and worker commands when enabled
2. **Environment Variables:** Core OTel env vars injected when enabled
3. **Extra Env Vars:** User-specified extra env vars included
4. **Sidecar Conditional:** Metrics sidecar skipped when OTel enabled
5. **Service Ports:** Metrics port excluded from Service when OTel enabled
6. **Backward Compatible:** Sidecar and metrics port present when OTel disabled/not configured
7. **Validation:** Missing endpoint rejected when OTel enabled
8. **Tests Pass:** All unit and integration tests pass
9. **Coverage:** ≥80% for new code

---

## Decision Log

| Decision | Options | Chosen | Rationale |
|----------|---------|--------|-----------|
| Sidecar behavior | Remove sidecar vs keep both | **Remove sidecar** | Simplify pod, avoid duplicate metrics |
| Default protocol | None vs grpc | **grpc** | Match OTel SDK defaults |
| Insecure env var | Always set vs only when true | **Only when true** | Reduce env var noise |
| Extra env vars | New field vs reuse env.variables | **Dedicated field** | Clearer OTel-specific config |
| Endpoint validation | Always vs only when enabled | **Only when enabled** | Allow partial config for disabled state |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Image lacks OTel extras | Medium | High | Document requirement clearly |
| OTel collector not deployed | Medium | Medium | User responsibility, document setup |
| Env var conflicts | Low | Medium | OTel uses standard `OTEL_*` prefix |
| Command parsing issues | Low | Low | Test with real Locust image |

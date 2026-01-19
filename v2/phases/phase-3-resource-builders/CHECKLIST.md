# Phase 3: Resource Builders - Checklist

**Estimated Effort:** 2 days  
**Status:** ✅ Complete (2026-01-17)

---

## Pre-Implementation

- [x] Phase 1 complete (v1 API types exist in `api/v1/locusttest_types.go`)
- [x] Phase 2 complete (`internal/config/config.go` exists with `OperatorConfig`)
- [x] Reviewed Java `ResourceCreationHelpers.java` (588 LOC)
- [x] Reviewed Java `LoadGenHelpers.java` (398 LOC)
- [x] Reviewed Java `Constants.java` (89 LOC)
- [x] Created `internal/resources/` directory

---

## Day 1: Core Builders

### Task 3.1: Create Types File

**File:** `internal/resources/types.go`

- [x] Define `OperationalMode` type as string
- [x] Define `Master` constant ("master")
- [x] Define `Worker` constant ("worker")
- [x] Add `String()` method for `OperationalMode`
- [x] File compiles: `go build ./internal/resources/...`

---

### Task 3.2: Create Constants File

**File:** `internal/resources/constants.go`

Port constants:
- [x] `MasterPort = 5557`
- [x] `MasterBindPort = 5558`
- [x] `WebUIPort = 8089`
- [x] `WorkerPort = 8080`
- [x] `DefaultMetricsExporterPort = 9646`

Mount path constants:
- [x] `DefaultMountPath = "/lotest/src/"`
- [x] `LibMountPath = "/opt/locust/lib"`

Label constants:
- [x] `LabelTestName = "performance-test-name"`
- [x] `LabelPodName = "performance-test-pod-name"`
- [x] `LabelManagedBy = "managed-by"`
- [x] `ManagedByValue = "locust-k8s-operator"`
- [x] `LabelApp = "app"`

Prometheus annotation constants:
- [x] `AnnotationPrometheusScrape = "prometheus.io/scrape"`
- [x] `AnnotationPrometheusPath = "prometheus.io/path"`
- [x] `AnnotationPrometheusPort = "prometheus.io/port"`
- [x] `MetricsEndpointPath = "/metrics"`

Job constants:
- [x] `BackoffLimit = 0`
- [x] `RestartPolicyNever = "Never"`
- [x] `MasterReplicaCount = 1`

Container constants:
- [x] `LocustContainerName = "locust"`
- [x] `MetricsExporterContainerName = "locust-metrics-exporter"`
- [x] `LibVolumeName = "lib"`

Exporter env var constants:
- [x] `ExporterURIEnvVar = "LOCUST_EXPORTER_URI"`
- [x] `ExporterPortEnvVar = "LOCUST_EXPORTER_WEB_LISTEN_ADDRESS"`

Kafka env var constants:
- [x] `EnvKafkaBootstrapServers`
- [x] `EnvKafkaSecurityEnabled`
- [x] `EnvKafkaSecurityProtocol`
- [x] `EnvKafkaSaslMechanism`
- [x] `EnvKafkaSaslJaasConfig`
- [x] `EnvKafkaUsername`
- [x] `EnvKafkaPassword`

Service constants:
- [x] `ProtocolTCP = "TCP"`
- [x] `PortNamePrefix = "port"`
- [x] `MetricsPortName = "prometheus-metrics"`

---

### Task 3.3: Create Labels Helper

**File:** `internal/resources/labels.go`

Functions:
- [x] `NodeName(crName string, mode OperationalMode) string`
  - [x] Replaces dots with dashes
  - [x] Format: `{cr-name}-{mode}`
- [x] `BuildLabels(lt *locustv1.LocustTest, mode OperationalMode) map[string]string`
  - [x] Includes `app`, `performance-test-name`, `performance-test-pod-name`, `managed-by`
  - [x] Merges user-defined labels from CR spec
- [x] `getUserLabels(lt, mode)` helper (unexported)
- [x] `BuildAnnotations(lt *locustv1.LocustTest, mode OperationalMode, cfg *config.OperatorConfig) map[string]string`
  - [x] Master: includes Prometheus scrape annotations
  - [x] Worker: no Prometheus annotations
  - [x] Merges user-defined annotations from CR spec
- [x] `getUserAnnotations(lt, mode)` helper (unexported)

---

### Task 3.4: Create Ports Helper

**File:** `internal/resources/ports.go`

Functions:
- [x] `MasterPorts() []corev1.ContainerPort`
  - [x] Returns ports 5557, 5558, 8089
- [x] `WorkerPorts() []corev1.ContainerPort`
  - [x] Returns port 8080
- [x] `MasterPortInts() []int32` (optional, for service)
- [x] `WorkerPortInts() []int32` (optional)

---

### Task 3.5: Create Command Builder

**File:** `internal/resources/command.go`

Functions:
- [x] `BuildMasterCommand(commandSeed string, workerReplicas int32) []string`
  - [x] Appends `--master --master-port=5557`
  - [x] Appends `--expect-workers={N}`
  - [x] Appends `--autostart --autoquit 60`
  - [x] Appends `--enable-rebalancing --only-summary`
  - [x] Uses `strings.Fields()` to split
- [x] `BuildWorkerCommand(commandSeed string, masterHost string) []string`
  - [x] Appends `--worker`
  - [x] Appends `--master-port=5557`
  - [x] Appends `--master-host={host}`
  - [x] Uses `strings.Fields()` to split

---

## Day 2: Job & Service Builders

### Task 3.6: Create Job Builder

**File:** `internal/resources/job.go`

Main functions:
- [x] `BuildMasterJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job`
- [x] `BuildWorkerJob(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *batchv1.Job`

Helper functions (unexported):
- [x] `buildLocustContainer(lt, name, command, ports, cfg) corev1.Container`
- [x] `buildMetricsExporterContainer(cfg) corev1.Container`
- [x] `buildImagePullSecrets(lt) []corev1.LocalObjectReference`
- [x] `buildVolumes(lt, nodeName) []corev1.Volume`
- [x] `buildVolumeMounts(lt, nodeName) []corev1.VolumeMount`
- [x] `buildKafkaEnvVars(cfg) []corev1.EnvVar`
- [x] `buildResourceRequirements(cfg, isMetricsExporter) corev1.ResourceRequirements`
- [x] `buildResourceList(cpu, memory, ephemeral) corev1.ResourceList`
- [x] `buildAffinity(lt, cfg) *corev1.Affinity`
- [x] `buildNodeSelector(requirements) *corev1.NodeSelector`
- [x] `buildTolerations(lt, cfg) []corev1.Toleration`
- [x] `ptr[T any](v T) *T` helper for pointers (removed - unused)

Job structure verification:
- [x] Master Job has `Parallelism=1`
- [x] Worker Job has `Parallelism=WorkerReplicas`
- [x] Both have `BackoffLimit=0`
- [x] Both have `RestartPolicy=Never`
- [x] Both respect `TTLSecondsAfterFinished` from config
- [x] Master has 2 containers (locust + metrics exporter)
- [x] Worker has 1 container (locust only)
- [x] Volumes created for ConfigMap and LibConfigMap
- [x] Volume mounts at correct paths

---

### Task 3.7: Create Service Builder

**File:** `internal/resources/service.go`

Functions:
- [x] `BuildMasterService(lt *locustv1.LocustTest, cfg *config.OperatorConfig) *corev1.Service`

Service structure verification:
- [x] Service type is `ClusterIP`
- [x] Selector uses `LabelPodName` matching master node name
- [x] Exposes port 5557 (master)
- [x] Exposes port 5558 (bind)
- [x] Exposes metrics port from config (default 9646)
- [x] Does NOT expose port 8089 (web UI)

---

## Unit Tests

### Task 3.8: Labels Tests

**File:** `internal/resources/labels_test.go`

- [x] `TestNodeName` - basic naming
- [x] `TestNodeName` - dots replaced with dashes
- [x] `TestBuildLabels` - required labels present
- [x] `TestBuildLabels_WithUserLabels` - user labels merged
- [x] `TestBuildAnnotations_Master` - Prometheus annotations present
- [x] `TestBuildAnnotations_Worker` - no Prometheus annotations
- [x] `TestBuildAnnotations_WithUserAnnotations` - user annotations merged

---

### Task 3.9: Command Tests

**File:** `internal/resources/command_test.go`

- [x] `TestBuildMasterCommand` - all flags present
- [x] `TestBuildMasterCommand_SplitsCorrectly` - whitespace handling
- [x] `TestBuildWorkerCommand` - all flags present
- [x] `TestBuildWorkerCommand_MasterHostCorrect` - master host in command

---

### Task 3.10: Job Tests

**File:** `internal/resources/job_test.go`

Master job tests:
- [x] `TestBuildMasterJob` - basic structure
- [x] `TestBuildMasterJob_Metadata` - name, namespace
- [x] `TestBuildMasterJob_Parallelism` - always 1
- [x] `TestBuildMasterJob_Containers` - 2 containers
- [x] `TestBuildMasterJob_WithTTL` - TTL set from config
- [x] `TestBuildMasterJob_WithImagePullSecrets` - secrets mounted
- [x] `TestBuildMasterJob_WithLibConfigMap` - lib volume mounted
- [x] `TestBuildMasterJob_Labels` - correct labels
- [x] `TestBuildMasterJob_Annotations` - Prometheus annotations

Worker job tests:
- [x] `TestBuildWorkerJob` - basic structure
- [x] `TestBuildWorkerJob_Parallelism` - equals WorkerReplicas
- [x] `TestBuildWorkerJob_Containers` - 1 container only
- [x] `TestBuildWorkerJob_NoPrometheusAnnotations` - no scrape annotations

Resource tests:
- [x] `TestBuildResourceRequirements` - locust container resources
- [x] `TestBuildResourceRequirements_MetricsExporter` - exporter resources

Affinity/Tolerations tests:
- [x] `TestBuildAffinity_Disabled` - nil when feature flag off
- [x] `TestBuildAffinity_Enabled` - builds from CR spec
- [x] `TestBuildTolerations_Disabled` - nil when feature flag off
- [x] `TestBuildTolerations_Enabled` - builds from CR spec

---

### Task 3.11: Service Tests

**File:** `internal/resources/service_test.go`

- [x] `TestBuildMasterService` - basic structure
- [x] `TestBuildMasterService_Ports` - correct ports exposed
- [x] `TestBuildMasterService_NoWebUIPort` - port 8089 NOT exposed
- [x] `TestBuildMasterService_CustomMetricsPort` - respects config
- [x] `TestBuildMasterService_Selector` - correct pod selector

---

## Verification

### Build Verification

- [x] `go build ./internal/resources/...` succeeds
- [x] `make build` succeeds (full project)
- [x] No compiler warnings

### Test Verification

- [x] `go test ./internal/resources/... -v` passes (35 tests)
- [x] `go test ./internal/resources/... -cover` shows >80% coverage (93.9%)
- [x] `golangci-lint run ./internal/resources/...` passes (0 issues)

### Manual Verification

- [x] Compare generated Job YAML with Java-generated Job (visual inspection)
- [x] Compare generated Service YAML with Java-generated Service
- [x] Command strings match expected format

---

## Post-Implementation

- [x] All tests pass
- [x] Code follows Go conventions (gofmt, golint)
- [x] Functions are documented with godoc comments
- [x] No hardcoded values that should be constants
- [x] Update `phases/README.md` with Phase 3 status
- [x] Update `phases/NOTES.md` with any deviations or notes

---

## Files Created Summary

| File | Status | LOC |
|------|--------|-----|
| `internal/resources/types.go` | [x] | ~30 |
| `internal/resources/constants.go` | [x] | ~120 |
| `internal/resources/labels.go` | [x] | ~100 |
| `internal/resources/ports.go` | [x] | ~50 |
| `internal/resources/command.go` | [x] | ~45 |
| `internal/resources/job.go` | [x] | ~365 |
| `internal/resources/service.go` | [x] | ~70 |
| `internal/resources/labels_test.go` | [x] | ~200 |
| `internal/resources/command_test.go` | [x] | ~90 |
| `internal/resources/job_test.go` | [x] | ~300 |
| `internal/resources/service_test.go` | [x] | ~160 |
| **Total** | | **~1530** |

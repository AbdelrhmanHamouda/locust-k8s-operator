# Phase 10: Environment & Secret Injection - Checklist

**Estimated Effort:** 1 day  
**Status:** ✅ Complete  
**Completed:** 2026-01-20

---

## Pre-Implementation

- [ ] Phase 7 complete (v2 API Types with `EnvConfig` defined)
- [ ] Phase 9 complete (controller uses v2 API)
- [ ] Review `api/v2/locusttest_types.go` for EnvConfig structure
- [ ] Review `issue-analysis/P1-High/issue-149-secrets-injection.md`
- [ ] Review current `internal/resources/job.go` implementation
- [ ] Run `make build && make test` to ensure clean starting state

---

## Task 10.1: Create Environment Builder Functions

**File:** `internal/resources/env.go`

### Core Functions

- [ ] Create file `internal/resources/env.go`
- [ ] Add copyright header
- [ ] Implement `BuildEnvFrom()` function:
  ```go
  func BuildEnvFrom(lt *locustv2.LocustTest) []corev1.EnvFromSource
  ```
  - [ ] Return nil for nil EnvConfig
  - [ ] Process `configMapRefs` → `ConfigMapEnvSource` with prefix
  - [ ] Process `secretRefs` → `SecretEnvSource` with prefix
- [ ] Implement `BuildUserEnvVars()` function:
  ```go
  func BuildUserEnvVars(lt *locustv2.LocustTest) []corev1.EnvVar
  ```
  - [ ] Return nil for nil/empty variables
  - [ ] Return copy of variables slice
- [ ] Implement `BuildEnvVars()` function:
  ```go
  func BuildEnvVars(lt *locustv2.LocustTest, cfg *config.OperatorConfig) []corev1.EnvVar
  ```
  - [ ] Start with Kafka env vars
  - [ ] Append user-defined variables

### Secret Volume Functions

- [ ] Implement `BuildSecretVolumes()` function:
  ```go
  func BuildSecretVolumes(lt *locustv2.LocustTest) []corev1.Volume
  ```
  - [ ] Return nil for nil/empty secretMounts
  - [ ] Create volume with `SecretVolumeSource` for each mount
- [ ] Implement `BuildSecretVolumeMounts()` function:
  ```go
  func BuildSecretVolumeMounts(lt *locustv2.LocustTest) []corev1.VolumeMount
  ```
  - [ ] Return nil for nil/empty secretMounts
  - [ ] Create mount with correct name, path, and readOnly
- [ ] Implement `secretVolumeName()` helper:
  ```go
  func secretVolumeName(secretName string) string
  ```
  - [ ] Return `"secret-" + secretName`

**Verification:**
```bash
go build ./internal/resources/...
```

---

## Task 10.2: Write Environment Builder Tests

**File:** `internal/resources/env_test.go`

### EnvFrom Tests

- [ ] Create file `internal/resources/env_test.go`
- [ ] Add copyright header
- [ ] `TestBuildEnvFrom_NilEnvConfig` - Returns nil for nil env
- [ ] `TestBuildEnvFrom_EmptyEnvConfig` - Returns nil for empty config
- [ ] `TestBuildEnvFrom_ConfigMapRefs` - ConfigMapEnvSource created correctly
- [ ] `TestBuildEnvFrom_ConfigMapRefs_WithPrefix` - Prefix applied
- [ ] `TestBuildEnvFrom_SecretRefs` - SecretEnvSource created correctly
- [ ] `TestBuildEnvFrom_SecretRefs_WithPrefix` - Prefix applied
- [ ] `TestBuildEnvFrom_Multiple` - Multiple refs combined

### EnvVars Tests

- [ ] `TestBuildUserEnvVars_Nil` - Returns nil for nil env
- [ ] `TestBuildUserEnvVars_Empty` - Returns nil for empty variables
- [ ] `TestBuildUserEnvVars_DirectValues` - Values set correctly
- [ ] `TestBuildUserEnvVars_ValueFrom` - ValueFrom preserved
- [ ] `TestBuildEnvVars_OnlyKafka` - Returns only Kafka vars when no user vars
- [ ] `TestBuildEnvVars_Combined` - Kafka + user vars combined

### Secret Volume Tests

- [ ] `TestBuildSecretVolumes_Nil` - Returns nil for nil env
- [ ] `TestBuildSecretVolumes_Empty` - Returns nil for empty mounts
- [ ] `TestBuildSecretVolumes_Single` - Single volume created
- [ ] `TestBuildSecretVolumes_Multiple` - Multiple volumes created
- [ ] `TestBuildSecretVolumeMounts_Nil` - Returns nil for nil env
- [ ] `TestBuildSecretVolumeMounts_Empty` - Returns nil for empty mounts
- [ ] `TestBuildSecretVolumeMounts_Single` - Single mount created
- [ ] `TestBuildSecretVolumeMounts_ReadOnly` - ReadOnly honored
- [ ] `TestSecretVolumeName` - Correct name generation

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuild
go test ./internal/resources/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep env
# Target: ≥90% coverage for env.go
```

---

## Task 10.3: Update Job Builder

**File:** `internal/resources/job.go`

### Container Updates

- [ ] Update `buildLocustContainer()`:
  - [ ] Change `Env: buildKafkaEnvVars(cfg)` → `Env: BuildEnvVars(lt, cfg)`
  - [ ] Add `EnvFrom: BuildEnvFrom(lt)` field

### Volume Updates

- [ ] Update `buildVolumes()`:
  - [ ] After ConfigMap volumes, append `BuildSecretVolumes(lt)` results
  
- [ ] Update `buildVolumeMounts()`:
  - [ ] After ConfigMap mounts, append `BuildSecretVolumeMounts(lt)` results

### Move Function (Optional Refactor)

- [ ] Consider moving `buildKafkaEnvVars()` to `env.go` for consistency
  - Keep it exported if moved: `BuildKafkaEnvVars()`
  - Update references in job.go

**Verification:**
```bash
go build ./internal/resources/...
make test
```

---

## Task 10.4: Update Job Builder Tests

**File:** `internal/resources/job_test.go`

### New Test Cases

- [ ] `TestBuildMasterJob_WithEnvConfigMapRef` - envFrom contains ConfigMapRef
- [ ] `TestBuildMasterJob_WithEnvSecretRef` - envFrom contains SecretRef
- [ ] `TestBuildMasterJob_WithEnvVariables` - env contains user variables
- [ ] `TestBuildMasterJob_WithSecretMount` - volume and volumeMount created
- [ ] `TestBuildMasterJob_EnvCombinesKafkaAndUser` - Both sources present
- [ ] `TestBuildWorkerJob_WithEnvConfig` - Worker job also gets env injection

### Existing Test Verification

- [ ] Run all existing job tests to ensure no regressions
- [ ] Verify Kafka env vars still present in all tests

**Verification:**
```bash
go test ./internal/resources/... -v -run TestBuildMasterJob
go test ./internal/resources/... -v -run TestBuildWorkerJob
```

---

## Task 10.5: Implement Validation Webhook

**File:** `api/v2/locusttest_webhook.go`

### Webhook Setup

- [ ] Create file `api/v2/locusttest_webhook.go`
- [ ] Add copyright header
- [ ] Define `reservedPaths` slice with defaults
- [ ] Implement `SetupWebhookWithManager()` function
- [ ] Add kubebuilder webhook marker:
  ```go
  // +kubebuilder:webhook:path=/validate-locust-io-v2-locusttest,mutating=false,failurePolicy=fail,sideEffects=None,groups=locust.io,resources=locusttests,verbs=create;update,versions=v2,name=vlocusttest.kb.io,admissionReviewVersions=v1
  ```

### Validator Interface

- [ ] Implement `ValidateCreate()` - Call `validateSecretMounts()`
- [ ] Implement `ValidateUpdate()` - Call `validateSecretMounts()`
- [ ] Implement `ValidateDelete()` - Return nil (no validation needed)

### Validation Logic

- [ ] Implement `validateSecretMounts()`:
  - [ ] Return nil for nil env or empty secretMounts
  - [ ] Check each mount path against reserved paths
  - [ ] Return error with clear message if conflict found
- [ ] Implement `pathConflicts()` helper:
  - [ ] Normalize paths (remove trailing slashes)
  - [ ] Check if either path is prefix of the other
- [ ] Implement `getReservedPaths()` (dynamic based on testFiles):
  - [ ] Use default paths if testFiles not set
  - [ ] Use custom paths from testFiles if set
  - [ ] Only include paths that are actually in use

**Verification:**
```bash
go build ./api/v2/...
make manifests  # Generates webhook manifests
```

---

## Task 10.6: Write Webhook Tests

**File:** `api/v2/locusttest_webhook_test.go`

### Unit Tests

- [ ] Create file `api/v2/locusttest_webhook_test.go`
- [ ] Add copyright header
- [ ] `TestPathConflicts_ExactMatch` - Same path conflicts
- [ ] `TestPathConflicts_Subpath` - /foo conflicts with /foo/bar
- [ ] `TestPathConflicts_NoConflict` - /foo and /bar don't conflict
- [ ] `TestPathConflicts_TrailingSlash` - Handles trailing slashes

### Validation Tests

- [ ] `TestValidateSecretMounts_NilEnv` - Passes
- [ ] `TestValidateSecretMounts_EmptyMounts` - Passes
- [ ] `TestValidateSecretMounts_ValidPath` - Passes for /custom/path
- [ ] `TestValidateSecretMounts_ConflictDefault` - Fails for /lotest/src
- [ ] `TestValidateSecretMounts_ConflictLib` - Fails for /opt/locust/lib
- [ ] `TestValidateSecretMounts_ConflictSubpath` - Fails for /lotest/src/secrets
- [ ] `TestValidateSecretMounts_CustomTestFilesPath` - Uses custom path for validation

**Verification:**
```bash
go test ./api/v2/... -v -run TestValidate
go test ./api/v2/... -v -run TestPathConflicts
```

---

## Task 10.7: Update Integration Tests

**File:** `internal/controller/integration_test.go`

### New Integration Tests

- [ ] Add test helper to create LocustTest with EnvConfig
- [ ] `TestReconcile_WithEnvConfigMapRef`:
  - [ ] Create LocustTest with configMapRefs
  - [ ] Verify Job container has envFrom with ConfigMapRef
- [ ] `TestReconcile_WithEnvSecretRef`:
  - [ ] Create LocustTest with secretRefs
  - [ ] Verify Job container has envFrom with SecretRef
- [ ] `TestReconcile_WithEnvVariables`:
  - [ ] Create LocustTest with variables
  - [ ] Verify Job container has env entries
  - [ ] Verify Kafka vars still present
- [ ] `TestReconcile_WithSecretMount`:
  - [ ] Create LocustTest with secretMounts
  - [ ] Verify Job has secret volume
  - [ ] Verify container has volume mount

**Verification:**
```bash
go test ./internal/controller/... -v -run TestReconcile_WithEnv
```

---

## Task 10.8: Create Sample CR

**File:** `config/samples/locust_v2_locusttest_with_env.yaml`

- [ ] Create sample CR demonstrating env injection:
  ```yaml
  apiVersion: locust.io/v2
  kind: LocustTest
  metadata:
    name: load-test-with-env
  spec:
    image: locustio/locust:2.20.0
    master:
      command: "locust -f /lotest/src/locustfile.py"
    worker:
      command: "locust -f /lotest/src/locustfile.py"
      replicas: 2
    testFiles:
      configMapRef: locust-scripts
    env:
      configMapRefs:
        - name: app-config
          prefix: "APP_"
      secretRefs:
        - name: api-credentials
      variables:
        - name: TARGET_HOST
          value: "https://api.example.com"
      secretMounts:
        - name: tls-certs
          mountPath: /etc/locust/certs
          readOnly: true
  ```

---

## Verification

### Code Generation

```bash
make generate
make manifests
```

- [ ] No errors during generation
- [ ] Webhook manifests generated in `config/webhook/`

### Build & Test

```bash
make build
make test
```

- [ ] Project builds without errors
- [ ] All existing tests pass
- [ ] New tests pass
- [ ] Coverage ≥80% for `internal/resources/env.go`
- [ ] Coverage ≥80% for webhook validation functions

### Linting

```bash
golangci-lint run ./internal/resources/...
golangci-lint run ./api/v2/...
```

- [ ] No linting errors

### Manual Verification (Optional)

```bash
# Create test resources
kubectl create configmap app-config --from-literal=KEY1=value1
kubectl create secret generic api-credentials --from-literal=API_TOKEN=secret123

# Apply sample CR
kubectl apply -f config/samples/locust_v2_locusttest_with_env.yaml

# Verify env injection
kubectl get pods -l performance-test-name=load-test-with-env
kubectl exec -it <pod-name> -- env | grep -E "APP_|API_TOKEN|TARGET_HOST"

# Verify secret mount
kubectl exec -it <pod-name> -- ls /etc/locust/certs

# Test validation webhook (should fail)
cat <<EOF | kubectl apply -f -
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: invalid-mount
spec:
  image: locustio/locust:2.20.0
  master:
    command: "locust -f /lotest/src/locustfile.py"
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 1
  env:
    secretMounts:
      - name: bad-secret
        mountPath: /lotest/src/secrets  # Should be rejected
EOF
```

---

## Post-Implementation

- [ ] All verification steps pass
- [ ] Update `phases/README.md` with Phase 10 status
- [ ] Update `phases/NOTES.md` with implementation notes
- [ ] Commit with message: `feat: implement environment and secret injection (Issue #149)`

---

## Files Summary

| File | Action | Est. LOC |
|------|--------|----------|
| `internal/resources/env.go` | Create | ~100 |
| `internal/resources/env_test.go` | Create | ~300 |
| `internal/resources/job.go` | Modify | +15 |
| `internal/resources/job_test.go` | Modify | +100 |
| `api/v2/locusttest_webhook.go` | Create | ~80 |
| `api/v2/locusttest_webhook_test.go` | Create | ~150 |
| `internal/controller/integration_test.go` | Modify | +100 |
| `config/samples/locust_v2_locusttest_with_env.yaml` | Create | ~30 |

**Total Estimated:** ~875 LOC

---

## Quick Reference Commands

```bash
# Build and test
make generate
make manifests
make build
make test

# Run specific tests
go test ./internal/resources/... -v -run TestBuildEnv
go test ./api/v2/... -v -run TestValidate
go test ./internal/controller/... -v -run TestReconcile_WithEnv

# Check coverage
go test ./internal/resources/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep env

# Lint
golangci-lint run ./internal/resources/...
golangci-lint run ./api/v2/...
```

---

## Acceptance Criteria Summary

1. **ConfigMap Injection:** `configMapRefs` correctly create `envFrom` entries
2. **Secret Injection:** `secretRefs` correctly create `envFrom` entries
3. **Prefix Support:** Prefix field correctly applied to env vars
4. **Variables:** Individual `variables` added to container env
5. **Secret Mounts:** `secretMounts` create correct volumes and mounts
6. **Path Validation:** Reserved path conflicts rejected by webhook
7. **Backward Compatible:** Empty env spec = no change in behavior
8. **Kafka Preserved:** Kafka env vars still injected alongside user vars
9. **Tests Pass:** All unit and integration tests pass
10. **Coverage:** ≥80% for new code

---

## Decision Log

| Decision | Options | Chosen | Rationale |
|----------|---------|--------|-----------|
| Env order | User first or Kafka first | **Kafka first** | Matches existing behavior, user can override |
| Volume naming | secretName or prefixed | **Prefixed** | `secret-<name>` avoids conflicts with configmap volumes |
| Path validation | Static or dynamic | **Dynamic** | Respects custom testFiles paths |
| Webhook scope | v2 only or both | **v2 only** | v1 doesn't have env field |

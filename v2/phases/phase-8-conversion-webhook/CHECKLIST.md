# Phase 8: Conversion Webhook - Checklist

**Estimated Effort:** 1.5 days  
**Status:** ✅ Complete

---

## Critical Requirement

> **envtest does NOT run conversion webhooks.** v1 cannot remain as storage version.
> This phase requires E2E testing in Kind cluster to properly validate v2 as storage version.

---

## Pre-Implementation

- [ ] Phase 7 complete (v2 API Types defined)
- [ ] Review `CRD_API_DESIGN.md` §3 for conversion patterns
- [ ] Review `REQUIREMENTS.md` §5.1.5 for conversion requirements
- [ ] Verify existing v1 and v2 APIs in `api/v1/` and `api/v2/`
- [ ] Run `make build && make test` to ensure clean starting state

---

## Task 8.1: Mark v2 as Hub

**File:** `api/v2/locusttest_conversion.go`

- [ ] Create file `api/v2/locusttest_conversion.go`
- [ ] Implement `Hub()` method on `*LocustTest`
- [ ] Add copyright header

**Verification:**
```bash
go build ./api/v2/...
```

---

## Task 8.2: Implement v1 Spoke Conversion

**File:** `api/v1/locusttest_conversion.go`

### ConvertTo (v1 → v2)

- [ ] Metadata (ObjectMeta) copied
- [ ] `image` → `image`
- [ ] `imagePullPolicy` → `imagePullPolicy` (string to PullPolicy)
- [ ] `imagePullSecrets[]` (strings) → `imagePullSecrets[]` (LocalObjectReference)
- [ ] `masterCommandSeed` → `master.command`
- [ ] Set `master.autostart` = true (default)
- [ ] Set `master.autoquit` = {enabled: true, timeout: 60} (default)
- [ ] `labels["master"]` → `master.labels`
- [ ] `annotations["master"]` → `master.annotations`
- [ ] `workerCommandSeed` → `worker.command`
- [ ] `workerReplicas` → `worker.replicas`
- [ ] `labels["worker"]` → `worker.labels`
- [ ] `annotations["worker"]` → `worker.annotations`
- [ ] `configMap` → `testFiles.configMapRef`
- [ ] `libConfigMap` → `testFiles.libConfigMapRef`
- [ ] `affinity` (custom) → `scheduling.affinity` (corev1.Affinity)
- [ ] `tolerations[]` (custom) → `scheduling.tolerations[]` (corev1.Toleration)

### ConvertFrom (v2 → v1) - Lossy

- [ ] Metadata (ObjectMeta) copied
- [ ] `image` → `image`
- [ ] `imagePullPolicy` → `imagePullPolicy` (PullPolicy to string)
- [ ] `imagePullSecrets[]` → `imagePullSecrets[]` (LocalObjectReference to strings)
- [ ] `master.command` → `masterCommandSeed`
- [ ] `master.labels` → `labels["master"]`
- [ ] `master.annotations` → `annotations["master"]`
- [ ] `worker.command` → `workerCommandSeed`
- [ ] `worker.replicas` → `workerReplicas`
- [ ] `worker.labels` → `labels["worker"]`
- [ ] `worker.annotations` → `annotations["worker"]`
- [ ] `testFiles.configMapRef` → `configMap`
- [ ] `testFiles.libConfigMapRef` → `libConfigMap`
- [ ] `scheduling.affinity` → `affinity` (lossy)
- [ ] `scheduling.tolerations[]` → `tolerations[]`

### Helper Functions

- [ ] `convertImagePullSecretsToV2()`
- [ ] `convertImagePullSecretsToV1()`
- [ ] `convertAffinityToV2()`
- [ ] `convertAffinityToV1()`
- [ ] `convertTolerationsToV2()`
- [ ] `convertTolerationsToV1()`

**Verification:**
```bash
go build ./api/v1/...
```

---

## Task 8.3: Create Webhook Scaffold

```bash
operator-sdk create webhook \
  --group locust \
  --version v1 \
  --kind LocustTest \
  --conversion
```

- [ ] Command executed successfully
- [ ] `api/v1/locusttest_webhook.go` created/updated
- [ ] `config/webhook/` directory updated
- [ ] Webhook manifests generated

**File:** `api/v1/locusttest_webhook.go`

- [ ] `SetupWebhookWithManager()` implemented
- [ ] Logger configured

---

## Task 8.4: Add Deprecation Warning to v1

**File:** `api/v1/locusttest_types.go`

- [ ] Add marker: `+kubebuilder:deprecatedversion:warning="locust.io/v1 LocustTest is deprecated, migrate to locust.io/v2"`
- [ ] Remove `+kubebuilder:storageversion` if present

---

## Task 8.5: Update v2 Storage Version

**File:** `api/v2/locusttest_types.go`

- [ ] Verify `+kubebuilder:storageversion` marker present on `LocustTest` struct

---

## Task 8.6: Register Webhook with Manager

**File:** `cmd/main.go`

- [ ] Import `locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"`
- [ ] Add webhook registration block:
  ```go
  if os.Getenv("ENABLE_WEBHOOKS") != "false" {
      if err = (&locustv1.LocustTest{}).SetupWebhookWithManager(mgr); err != nil {
          // ...
      }
  }
  ```

---

## Task 8.7: Write Conversion Tests

**File:** `api/v1/locusttest_conversion_test.go`

- [ ] `TestConvertTo_FullSpec` - all v1 fields convert correctly
- [ ] `TestConvertFrom_FullSpec` - v2 fields convert to v1
- [ ] `TestRoundTrip_V1ToV2ToV1` - v1 data preserved after round-trip
- [ ] `TestConvertTo_MinimalSpec` - minimal v1 spec converts
- [ ] `TestAffinityConversion` - affinity round-trip works
- [ ] `TestTolerationsConversion` - tolerations round-trip works
- [ ] `TestImagePullSecretsConversion` - secrets format converts

**Verification:**
```bash
go test ./api/v1/... -v -run TestConvert
```

---

## Verification

### Code Generation

```bash
make generate
make manifests
```

- [ ] `api/v2/zz_generated.deepcopy.go` regenerated
- [ ] `config/crd/bases/locust.io_locusttests.yaml` updated
- [ ] CRD contains both v1 and v2 versions

### CRD Verification

```bash
cat config/crd/bases/locust.io_locusttests.yaml | grep -B2 -A2 "storage:"
```

- [ ] v2 has `storage: true`
- [ ] v1 has `storage: false`

### Deprecation Warning

```bash
cat config/crd/bases/locust.io_locusttests.yaml | grep -A2 "deprecated"
```

- [ ] v1 shows deprecation warning

### Conversion Webhook

```bash
cat config/crd/bases/locust.io_locusttests.yaml | grep -A10 "conversion:"
```

- [ ] Conversion strategy is `Webhook`
- [ ] Webhook path is `/convert`

### Build & Test

```bash
make build
make test
```

- [ ] Project builds without errors
- [ ] All existing tests still pass
- [ ] New conversion tests pass

---

## Post-Implementation

- [ ] All verification steps pass
- [ ] Update `phases/README.md` with Phase 8 status
- [ ] Update `phases/NOTES.md` with any deviations or decisions
- [ ] Document any v2-only fields that are lost in conversion

---

## Files Summary

| File | Action | Est. LOC |
|------|--------|----------|
| `api/v2/locusttest_conversion.go` | Create | ~15 |
| `api/v1/locusttest_conversion.go` | Create | ~200 |
| `api/v1/locusttest_webhook.go` | Create | ~25 |
| `api/v1/locusttest_types.go` | Modify | +1 (marker) |
| `api/v2/locusttest_types.go` | Verify | +0 (verify marker) |
| `cmd/main.go` | Modify | +10 |
| `api/v1/locusttest_conversion_test.go` | Create | ~200 |
| `config/webhook/` | Generated | Auto |
| `config/crd/bases/locust.io_locusttests.yaml` | Generated | Auto |

---

## Quick Reference Commands

```bash
# Create webhook scaffold
operator-sdk create webhook --group locust --version v1 --kind LocustTest --conversion

# Generate DeepCopy
make generate

# Generate CRD and webhook manifests
make manifests

# Build project
make build

# Run all tests
make test

# Run conversion tests only
go test ./api/v1/... -v -run TestConvert

# Check CRD versions
cat config/crd/bases/locust.io_locusttests.yaml | grep -A30 "versions:"

# Check conversion webhook config
cat config/crd/bases/locust.io_locusttests.yaml | grep -A10 "conversion:"

# Run locally with webhooks (requires cert-manager)
ENABLE_WEBHOOKS=true make run
```

---

## Acceptance Criteria Summary

1. **Hub Implemented:** v2 has `Hub()` method
2. **Spoke Implemented:** v1 has `ConvertTo()` and `ConvertFrom()`
3. **Storage Version:** v2 marked as storage version in CRD
4. **Deprecation Warning:** v1 shows warning when used
5. **Tests Pass:** All conversion tests pass
6. **Build Succeeds:** `make build` completes without errors
7. **Manifests Generated:** Webhook configuration in `config/webhook/`

---

## Lossy Fields Reference

The following v2-only fields are **NOT preserved** when converting v2 → v1:

| v2 Field | Notes |
|----------|-------|
| `master.resources` | Resource requests/limits |
| `master.extraArgs` | Additional CLI arguments |
| `worker.resources` | Resource requests/limits |
| `worker.extraArgs` | Additional CLI arguments |
| `testFiles.srcMountPath` | Custom mount path |
| `testFiles.libMountPath` | Custom mount path |
| `scheduling.nodeSelector` | Node selector labels |
| `env.configMapRefs` | Environment from ConfigMaps |
| `env.secretRefs` | Environment from Secrets |
| `env.variables` | Individual env vars |
| `env.secretMounts` | Secret file mounts |
| `volumes` | Volume definitions |
| `volumeMounts` | Volume mounts with target |
| `observability.openTelemetry` | OTel configuration |
| `status.*` | All status fields |

---

## Task 8.8: E2E Conversion Webhook Testing (Kind)

> **This section is REQUIRED to complete Phase 8.**

### Setup Kind Cluster

- [ ] Kind installed: `brew install kind`
- [ ] Create Kind config: `test/e2e/kind-config.yaml`
- [ ] Create cluster: `kind create cluster --name locust-webhook-test --config test/e2e/kind-config.yaml`

### Install cert-manager

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
kubectl wait --for=condition=Available deployment/cert-manager-webhook -n cert-manager --timeout=120s
```

- [ ] cert-manager installed and ready

### Build & Deploy Operator

```bash
make docker-build IMG=locust-k8s-operator:e2e-test
kind load docker-image locust-k8s-operator:e2e-test --name locust-webhook-test
make install
make deploy IMG=locust-k8s-operator:e2e-test
```

- [ ] Operator image built
- [ ] Image loaded into Kind
- [ ] CRDs installed with v2 as storage version
- [ ] Operator deployed and running

### Verify Storage Version

```bash
kubectl get crd locusttests.locust.io -o jsonpath='{.spec.versions[?(@.storage==true)].name}'
```

- [ ] Output shows `v2`

### Create E2E Test Files

- [ ] `test/e2e/conversion/v1-cr.yaml` created
- [ ] `test/e2e/conversion/v2-cr.yaml` created
- [ ] `test/e2e/conversion/configmap.yaml` created
- [ ] `test/e2e/conversion/run-e2e.sh` created and executable

### Run E2E Tests

```bash
./test/e2e/conversion/run-e2e.sh
```

- [ ] Test 1: Create v1 CR → stored as v2 ✓
- [ ] Test 2: Read v1 CR as v2 → conversion works ✓
- [ ] Test 3: Create v2 CR directly ✓
- [ ] Test 4: Read v2 CR as v1 → conversion works ✓
- [ ] Test 5: Update v1 CR → reflected in v2 view ✓
- [ ] Test 6: Jobs created by reconciler ✓

### Cleanup

```bash
kind delete cluster --name locust-webhook-test
```

- [ ] Kind cluster deleted

---

## Final Acceptance Criteria

1. **Hub Implemented:** v2 has `Hub()` method
2. **Spoke Implemented:** v1 has `ConvertTo()` and `ConvertFrom()`
3. **Storage Version:** v2 is storage version (verified in CRD)
4. **Deprecation Warning:** v1 shows warning when used
5. **Unit Tests Pass:** All conversion tests pass
6. **E2E Tests Pass:** Conversion webhook works in Kind cluster
7. **Build Succeeds:** `make build` completes without errors
8. **Manifests Generated:** Webhook configuration in `config/webhook/`

---

## E2E Quick Reference Commands

```bash
# Create Kind cluster
kind create cluster --name locust-webhook-test --config test/e2e/kind-config.yaml

# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Build and load image
make docker-build IMG=locust-k8s-operator:e2e-test
kind load docker-image locust-k8s-operator:e2e-test --name locust-webhook-test

# Deploy
make install
make deploy IMG=locust-k8s-operator:e2e-test

# Verify storage version
kubectl get crd locusttests.locust.io -o jsonpath='{.spec.versions[?(@.storage==true)].name}'

# Run E2E tests
./test/e2e/conversion/run-e2e.sh

# Cleanup
kind delete cluster --name locust-webhook-test
```

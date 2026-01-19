# Phase 8: Conversion Webhook (v1↔v2)

**Status:** ✅ Complete  
**Effort:** 1.5 days  
**Priority:** P1 - Must Have  
**Dependencies:** Phase 7 (v2 API Types)

---

## ⚠️ Critical: E2E Testing Required

> **envtest does NOT run conversion webhooks.** Unit tests alone cannot validate v2 as storage version.
> 
> This phase requires E2E testing in a Kind cluster with cert-manager to properly validate the conversion webhook works and v2 is the actual storage version.

---

## Overview

Implement bidirectional conversion between v1 and v2 APIs using the Hub-and-Spoke pattern. This enables backward compatibility for existing v1 CRs while making v2 the storage version.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Current State

Phase 7 delivered v2 API types with grouped configuration:
- `api/v2/locusttest_types.go` with all v2 types
- `api/v2/conditions.go` with condition/phase constants
- v1 currently remains as storage version (to be changed in this phase)

**Phase 8 Goal:** Enable seamless v1↔v2 conversion via webhook with v2 as storage version.

## Hub-and-Spoke Pattern

```
                    ┌─────────────────┐
                    │   v2 (Hub)      │
                    │ Storage Version │
                    └────────┬────────┘
                             │
                    ConvertTo/ConvertFrom
                             │
                    ┌────────┴────────┐
                    │   v1 (Spoke)    │
                    │   Deprecated    │
                    └─────────────────┘
```

- **Hub (v2):** The canonical representation. Only implements `Hub()` marker method.
- **Spoke (v1):** Implements `ConvertTo()` and `ConvertFrom()` to convert to/from Hub.

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Hub Version** | v2 | v2 is the canonical representation with all features |
| **Storage Version** | v2 | New features require v2 storage |
| **Lossy Conversion** | v2→v1 loses data | v1 cannot represent v2-only fields |
| **Deprecation Warning** | v1 CRDs | Warn users to migrate to v2 |

## Conversion Mapping

### v1 → v2 (ConvertTo)

| v1 Field | v2 Field |
|----------|----------|
| `masterCommandSeed` | `master.command` |
| `workerCommandSeed` | `worker.command` |
| `workerReplicas` | `worker.replicas` |
| `image` | `image` |
| `imagePullPolicy` | `imagePullPolicy` |
| `imagePullSecrets[]` (strings) | `imagePullSecrets[]` (LocalObjectReference) |
| `configMap` | `testFiles.configMapRef` |
| `libConfigMap` | `testFiles.libConfigMapRef` |
| `labels.master` | `master.labels` |
| `labels.worker` | `worker.labels` |
| `annotations.master` | `master.annotations` |
| `annotations.worker` | `worker.annotations` |
| `affinity` (custom type) | `scheduling.affinity` (corev1.Affinity) |
| `tolerations[]` (custom type) | `scheduling.tolerations[]` (corev1.Toleration) |

### v2 → v1 (ConvertFrom) - Lossy

Fields **lost** when converting v2 → v1:
- `master.resources`, `worker.resources`
- `master.autostart`, `master.autoquit`, `master.extraArgs`
- `worker.extraArgs`
- `testFiles.srcMountPath`, `testFiles.libMountPath`
- `scheduling.nodeSelector`
- `env` (configMapRefs, secretRefs, variables, secretMounts)
- `volumes`, `volumeMounts`
- `observability` (OpenTelemetry config)
- `status` (v1 has no status subresource)

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `api/v2/locusttest_conversion.go` | Create | Hub marker implementation |
| `api/v1/locusttest_conversion.go` | Create | Spoke conversion logic |
| `api/v1/locusttest_webhook.go` | Create | Webhook setup for v1 |
| `api/v1/locusttest_types.go` | Modify | Add deprecation marker |
| `cmd/main.go` | Modify | Register webhook with manager |
| `config/webhook/` | Generate | Webhook configuration |
| `config/default/manager_webhook_patch.yaml` | Generate | Webhook deployment patch |
| `config/default/webhookcainjection_patch.yaml` | Generate | CA injection for webhook |

## Acceptance Criteria

1. **Hub Implementation:**
   - v2 implements `Hub()` marker method
   - v2 marked with `+kubebuilder:storageversion`

2. **Spoke Conversion:**
   - v1 `ConvertTo()` correctly maps all v1 fields to v2
   - v1 `ConvertFrom()` maps v2 fields back to v1 (with documented loss)

3. **Deprecation Warning:**
   - v1 CRD shows deprecation warning when used
   - Warning message guides users to migrate to v2

4. **Conversion Tests:**
   - Round-trip test: v1 → v2 → v1 preserves v1 data
   - v2-only fields preserved in v2 storage

5. **Webhook Configuration:**
   - Webhook manifests generated in `config/webhook/`
   - Webhook starts with operator

## Non-Goals (Deferred)

| Feature | Phase |
|---------|-------|
| Status update logic | Phase 9 |
| Environment injection logic | Phase 10 |
| Volume mounting logic | Phase 11 |
| OpenTelemetry logic | Phase 12 |

## Commands

```bash
# Create webhook scaffold (conversion only, no validation yet)
operator-sdk create webhook \
  --group locust \
  --version v1 \
  --kind LocustTest \
  --conversion

# Generate webhook manifests
make manifests

# Build and test
make build
make test

# Test webhook locally (requires cert-manager)
make run ENABLE_WEBHOOKS=true
```

## Testing Strategy

### Unit Tests
- `api/v1/locusttest_conversion_test.go`
  - Test `ConvertTo()` with full v1 spec
  - Test `ConvertFrom()` with full v2 spec
  - Test round-trip preservation

### E2E Tests (Kind Cluster) - REQUIRED

> Unit tests verify conversion logic. E2E tests verify the **webhook actually works** in a real cluster.

**Setup:**
1. Create Kind cluster with webhook support
2. Install cert-manager for TLS
3. Build and deploy operator image
4. Verify v2 is storage version

**Test Cases:**
1. Create v1 CR → verify stored as v2, readable as v2
2. Create v2 CR → verify readable as v1
3. Update v1 CR → verify changes reflected in v2 view
4. Verify reconciler creates Jobs from converted resources

**Files:**
- `test/e2e/kind-config.yaml` - Kind cluster configuration
- `test/e2e/conversion/v1-cr.yaml` - Sample v1 CR for testing
- `test/e2e/conversion/v2-cr.yaml` - Sample v2 CR for testing
- `test/e2e/conversion/run-e2e.sh` - E2E test script

## Completion Checklist

- [x] Hub implementation (v2)
- [x] Spoke conversion (v1)
- [x] Unit tests pass
- [x] Webhook scaffold generated
- [ ] **v2 confirmed as storage version in cluster**
- [ ] **E2E tests pass in Kind cluster**
- [ ] Update `phases/NOTES.md` with completion status

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 8 definition (lines 413-452)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §5.1.5 API v1 Conversion Webhook
- [CRD_API_DESIGN.md](../../research/CRD_API_DESIGN.md) - §3 Conversion Webhooks
- [Kubebuilder Conversion](https://book.kubebuilder.io/multiversion-tutorial/conversion.html)
- [controller-runtime conversion](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/conversion)

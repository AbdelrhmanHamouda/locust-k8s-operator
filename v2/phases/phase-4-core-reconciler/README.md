# Phase 4: Core Reconciler

**Status:** ✅ Complete  
**Completed:** 2026-01-19  
**Effort:** 1.5 days  
**Priority:** P0 - Critical Path  
**Dependencies:** Phase 3 (Resource Builders)

---

## Overview

Implement the reconciliation loop matching Java `LocustTestReconciler.java` behavior. This phase wires together the resource builders from Phase 3 into a complete reconciler that watches LocustTest CRs, creates Kubernetes resources (Jobs, Services), and manages their lifecycle through owner references.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **NO-OP on Updates** | Skip reconcile if `generation > 1` | Match Java behavior - tests are immutable once created |
| **Owner References** | Use `controllerutil.SetControllerReference()` | Automatic garbage collection on CR deletion |
| **Create-Only** | Use `Create()` with `IsAlreadyExists` check | No updates to existing resources (immutable tests) |
| **Event Recording** | Use `record.EventRecorder` | User visibility into resource creation events |
| **Generation Predicate** | Apply `GenerationChangedPredicate` | Filter out status-only updates |

## Source of Truth

- **Java Reconciler:** `/src/main/java/com/locust/operator/controller/LocustTestReconciler.java`
- **Java Resource Creation Manager:** `/src/main/java/com/locust/operator/controller/utils/resource/manage/ResourceCreationManager.java`
- **Go Patterns:** `/v2/research/OPERATOR_SDK_PATTERNS.md`

## Files to Create/Modify

| File | Action | Purpose | Est. LOC |
|------|--------|---------|----------|
| `internal/controller/locusttest_controller.go` | Modify | Main reconciliation logic | ~150 |
| `cmd/main.go` | Modify | Wire reconciler with config | ~20 changes |

## Dependencies

### Upstream (Required Before Starting)
- **Phase 1:** `api/v1/locusttest_types.go` - CR types
- **Phase 2:** `internal/config/config.go` - Operator configuration
- **Phase 3:** `internal/resources/*.go` - Resource builders

### Downstream (Depends on This Phase)
- **Phase 5:** Unit Tests
- **Phase 6:** Integration Tests (envtest)

## Acceptance Criteria

1. **CR creation triggers resource creation:**
   - Master Service created with correct ports and selector
   - Master Job created with 2 containers (locust + metrics exporter)
   - Worker Job created with parallelism matching `workerReplicas`

2. **CR updates are NO-OP:**
   - Generation > 1 logs a message and returns without changes
   - Existing resources are not modified

3. **CR deletion triggers automatic cleanup:**
   - Owner references set on all created resources
   - Kubernetes garbage collection removes Jobs and Services

4. **Operator logs match Java patterns:**
   - "LocustTest created: '{name}'" on create
   - "Update operations on {name} are not currently supported!" on update

5. **Operator starts successfully:**
   - `make run` starts operator without errors
   - Watches LocustTest resources
   - Owns Job and Service resources

## Java Behavior to Preserve

### Reconcile Flow
```java
// On update >> NOOP
if (resource.getMetadata().getGeneration() > 1) {
    log.info("LocustTest updated: {} in namespace: {}.", name, namespace);
    log.info("Update operations on {} are not currently supported!", crdName);
    return UpdateControl.noUpdate();
}

// On add
log.info("LocustTest created: '{}'", name);

// Deploy resources in order:
// 1. Master Service
// 2. Master Job
// 3. Worker Job
```

### Cleanup Flow
In Go, cleanup is handled automatically via owner references - no explicit deletion code needed.

## RBAC Requirements

The controller needs permissions to manage Jobs and Services:

```yaml
# +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;delete
# +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete
# +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
```

## Error Handling Strategy

| Error Type | Action |
|------------|--------|
| CR Not Found | Return success (already deleted) |
| Resource Already Exists | Log and continue (idempotent) |
| API Conflict | Requeue immediately |
| Other Errors | Return error (triggers backoff) |

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 4 definition (lines 225-266)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §3.3 Resilience, §8.3 Design Principles
- [analysis/TECHNICAL.md](../../analysis/TECHNICAL.md) - §5.4 Reconciler
- [research/OPERATOR_SDK_PATTERNS.md](../../research/OPERATOR_SDK_PATTERNS.md) - §3 Reconciliation Patterns

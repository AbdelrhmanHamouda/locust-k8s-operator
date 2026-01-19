# Phase 3: Resource Builders

**Status:** ✅ Complete  
**Completed:** 2026-01-17  
**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Dependencies:** Phase 1 (v1 API Types), Phase 2 (Configuration System)

---

## Overview

Implement Job and Service builders matching Java `ResourceCreationHelpers.java` behavior. This phase creates the pure functions that build Kubernetes resources (Jobs, Services) from LocustTest CRs and operator configuration. These builders are the foundation for the reconciler in Phase 4.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Pure Functions** | Stateless builder functions | Easier testing, no side effects |
| **Struct Literals** | Go struct initialization over builders | Idiomatic Go, less verbose than Java builders |
| **Pointer Helpers** | `ptr.To[T]()` generic helper | Required for optional Kubernetes fields |
| **Separate Files** | One file per resource type | Clear separation of concerns |
| **Command Splitting** | `strings.Fields()` | Match Java `split(" ")` behavior |

## Source of Truth

- **Java Resource Helpers:** `/src/main/java/com/locust/operator/controller/utils/resource/manage/ResourceCreationHelpers.java`
- **Java Load Gen Helpers:** `/src/main/java/com/locust/operator/controller/utils/LoadGenHelpers.java`
- **Java Constants:** `/src/main/java/com/locust/operator/controller/utils/Constants.java`

## Files to Create

| File | Purpose | Est. LOC |
|------|---------|----------|
| `internal/resources/types.go` | OperationalMode type, constants | ~30 |
| `internal/resources/constants.go` | Ports, paths, labels | ~50 |
| `internal/resources/labels.go` | NodeName, BuildLabels, BuildAnnotations | ~80 |
| `internal/resources/command.go` | BuildMasterCommand, BuildWorkerCommand | ~60 |
| `internal/resources/job.go` | BuildMasterJob, BuildWorkerJob, helpers | ~200 |
| `internal/resources/service.go` | BuildMasterService | ~60 |
| `internal/resources/ports.go` | MasterPorts, WorkerPorts helpers | ~30 |

## Dependencies

- **Upstream:** 
  - Phase 1 (v1 API Types) - `api/v1/locusttest_types.go`
  - Phase 2 (Configuration System) - `internal/config/config.go`
- **Downstream:** Phase 4 (Core Reconciler)

## Acceptance Criteria

1. `BuildMasterJob()` produces Job spec matching Java output
2. `BuildWorkerJob()` produces Job spec matching Java output  
3. `BuildMasterService()` produces Service spec matching Java output
4. All resource builders are pure functions (no side effects)
5. Unit tests verify resource structure matches Java-generated resources
6. Commands are correctly split into string slices

## Key Java Behaviors to Preserve

### Command Templates
```
Master: "{seed} --master --master-port=5557 --expect-workers={N} --autostart --autoquit 60 --enable-rebalancing --only-summary"
Worker: "{seed} --worker --master-port=5557 --master-host={master-name}"
```

### Node Naming
```
Format: "{cr-name}-{mode}" with dots replaced by dashes
Example: "team-a.load-test" -> "team-a-load-test-master"
```

### Labels Applied to All Resources
```yaml
performance-test-name: "{cr-name}"
performance-test-pod-name: "{node-name}"
managed-by: "locust-k8s-operator"
```

### Prometheus Annotations (Master only)
```yaml
prometheus.io/scrape: "true"
prometheus.io/path: "/metrics"
prometheus.io/port: "9646"
```

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 3 definition (lines 168-222)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §3.3 Resilience (idempotent creation)
- [analysis/TECHNICAL.md](../../analysis/TECHNICAL.md) - §5.3 Resource Builders
- [research/RESOURCE_MANAGEMENT.md](../../research/RESOURCE_MANAGEMENT.md) - Go resource building patterns

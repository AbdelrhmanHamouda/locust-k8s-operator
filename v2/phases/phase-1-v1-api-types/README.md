# Phase 1: v1 API Types (Parity)

**Status:** Pending  
**Effort:** 1 day  
**Priority:** P0 - Critical Path  
**Dependency:** Phase 0 (Complete)

---

## Overview

Define Go types that exactly match the current Java v1 CRD for backward compatibility. This phase ensures existing LocustTest CRs continue to work without modification when migrating to the Go operator.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Field Names** | camelCase in Go, snake_case JSON tags | Match existing CRD JSON schema |
| **Nested Types** | Custom structs for labels/annotations/affinity | Match Java CRD structure exactly |
| **Validation** | Kubebuilder markers | Generate OpenAPI schema matching Java CRD |
| **Short Name** | `lotest` | Match existing kubectl shorthand |

## Source of Truth

- **Java CRD:** `/kube/crd/locust-test-crd.yaml`
- **Sample CR:** `/kube/sample-cr/locust-test-cr.yaml`

## Dependencies

- **Upstream:** Phase 0 (project scaffolding)
- **Downstream:** Phase 3 (Resource Builders), Phase 8 (Conversion Webhook)

## Acceptance Criteria

1. Generated CRD schema matches existing Java CRD
2. Sample CR from `kube/sample-cr/` validates against new CRD
3. `make generate` produces `zz_generated.deepcopy.go`
4. `kubectl get lotest` shows printer columns: master_cmd, worker_replica_count, Image, Age

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 1 definition (lines 93-131)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §4.4 v1 to v2 Field Mapping
- [analysis/TECHNICAL.md](../../analysis/TECHNICAL.md) - §5.1 API Types
- [research/JAVA_TO_GO_MAPPING.md](../../research/JAVA_TO_GO_MAPPING.md) - §1 Type Mapping
- [NOTES.md](../NOTES.md) - Phase 0 notes with v1 field requirements

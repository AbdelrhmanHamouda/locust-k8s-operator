# Phase 9: Status Subresource

**Effort:** 1 day  
**Priority:** P1 - Must Have  
**Status:** ✅ Complete  
**Dependencies:** Phase 7 (v2 API Types)

---

## Objective

Implement status tracking for LocustTest resources with phase, conditions, and worker connection tracking. This enables users to monitor test lifecycle through `kubectl` and enables automation based on test state.

---

## Requirements Reference

- **REQUIREMENTS.md §5.1.4:** Status Subresource
- **CRD_API_DESIGN.md §5:** Status Subresource Design
- **OPERATOR_SDK_PATTERNS.md §5:** Status Management

---

## Scope

### In Scope

- Define condition type and reason constants (`api/v2/conditions.go`)
- Create status helper functions (`internal/controller/status.go`)
- Update reconciler to set initial status on CR creation
- Watch Job status changes to update phase
- Update conditions based on Job lifecycle
- Printer columns for `kubectl get locusttests`

### Out of Scope

- Real-time worker connection count (requires Locust API integration)
- Test result metrics (future enhancement)
- Webhook-based status validation

---

## Key Deliverables

| File | Description |
|------|-------------|
| `api/v2/conditions.go` | Condition type/reason constants |
| `internal/controller/status.go` | Status update helper functions |
| `internal/controller/locusttest_controller.go` | Updated reconciler with status tracking |
| `internal/controller/status_test.go` | Unit tests for status helpers |

---

## Success Criteria

1. `kubectl get locusttests` shows Phase column with accurate state
2. Status reflects actual Job states (Pending → Running → Succeeded/Failed)
3. Conditions follow Kubernetes conventions (Ready, WorkersConnected, TestCompleted)
4. Status updates don't trigger unnecessary reconcile loops
5. All tests pass with ≥80% coverage for new code

---

## Quick Start

```bash
# After implementation, verify with:
make generate
make manifests
make test

# Check printer columns
kubectl get locusttests
# NAME          PHASE     WORKERS   CONNECTED   AGE
# demo-test     Running   10        10          5m
```

---

## Related Documents

- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist
- [DESIGN.md](./DESIGN.md) - Technical design and code patterns

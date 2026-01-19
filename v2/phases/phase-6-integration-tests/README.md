# Phase 6: Integration Tests (envtest)

**Status:** ✅ Complete  
**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Dependencies:** Phase 5 (Unit Tests)

---

## Overview

Implement controller integration tests using the **envtest** framework. Unlike unit tests that use fake clients, integration tests run against a real (lightweight) Kubernetes API server to validate actual controller behavior including watches, owner references, and garbage collection.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Current State

The existing `suite_test.go` provides the envtest scaffold:
- Basic test environment configuration
- CRD loading from `config/crd/bases/`
- k8sClient initialization
- Binary asset discovery for IDE compatibility

**Phase 6 Goal:** Build on this foundation to test full reconciliation flows.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Test Framework** | Ginkgo + Gomega | Standard for controller-runtime, async-friendly |
| **Test Environment** | envtest | Real API server without full cluster overhead |
| **Controller Startup** | Manager in BeforeSuite | Single manager for all tests |
| **Test Isolation** | Unique namespaces per test | Prevent resource conflicts |
| **Timeout Strategy** | Eventually/Consistently | Handle async reconciliation |

## Test Categories

### 1. Create Flow Tests
Verify complete resource creation when a new LocustTest CR is applied:
- Master Service created with correct ports and selectors
- Master Job created with correct spec
- Worker Job created with correct replicas
- Owner references set on all child resources
- Events recorded for resource creation

### 2. Update NO-OP Tests
Verify the operator's immutable test design:
- Updates to CR spec do not trigger new resources
- Existing resources remain unchanged after spec modification
- Generation > 1 results in early return from reconciler

### 3. Delete Flow Tests
Verify garbage collection through owner references:
- CR deletion triggers cascade delete of child resources
- Service, master Job, and worker Job are all cleaned up
- No orphaned resources remain

### 4. Error Handling Tests
Verify graceful handling of error conditions:
- Invalid CR spec rejected by webhook (if enabled)
- Missing ConfigMap handled gracefully
- API server errors trigger appropriate requeue

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/controller/suite_test.go` | Enhance | Add manager startup, controller registration |
| `internal/controller/integration_test.go` | Create | All integration test cases |

## Acceptance Criteria

1. **Test Coverage:**
   - Create flow fully tested
   - Update NO-OP behavior verified
   - Delete cascade verified
   - Error conditions tested

2. **Test Quality:**
   - All tests pass: `make test`
   - Tests run without external dependencies
   - Test execution < 2 minutes
   - No flaky tests

3. **CI Compatibility:**
   - Tests work in GitHub Actions
   - envtest binaries auto-downloaded

## Test Commands

```bash
# Run all tests (including integration)
make test

# Run only controller integration tests
go test -v ./internal/controller/... -ginkgo.focus="Integration"

# Run with verbose Ginkgo output
go test -v ./internal/controller/... -ginkgo.v

# Setup envtest binaries
make setup-envtest

# Run specific test by name
go test -v ./internal/controller/... -ginkgo.focus="should create master Service"
```

## envtest Key Concepts

### What envtest Provides
- **kube-apiserver**: Real API server for CRUD operations
- **etcd**: Backing store for API server
- **No kubelet/scheduler**: Resources are never actually scheduled

### What envtest Does NOT Provide
- Pod scheduling/execution
- Job completion status changes
- Network connectivity

### Test Implications
- Cannot test actual Locust execution
- Must verify resource specs, not runtime behavior
- Job status transitions require manual simulation

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Test Environment                         │
│  ┌─────────────────┐  ┌─────────────────────────────────┐  │
│  │   envtest       │  │         Test Code               │  │
│  │  ┌───────────┐  │  │  ┌───────────────────────────┐  │  │
│  │  │ API Server│◄─┼──┼──│   k8sClient (direct)      │  │  │
│  │  │           │  │  │  └───────────────────────────┘  │  │
│  │  │   etcd    │  │  │                                 │  │
│  │  └───────────┘  │  │  ┌───────────────────────────┐  │  │
│  │       ▲         │  │  │   Manager                  │  │  │
│  │       │         │  │  │   └── LocustTestReconciler│  │  │
│  │       │ Watch   │  │  │          │                 │  │  │
│  │       └─────────┼──┼──┼──────────┘                 │  │  │
│  └─────────────────┘  │  └───────────────────────────┘  │  │
└─────────────────────────────────────────────────────────────┘
```

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 6 definition (lines 319-361)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §7.1 Testing Requirements
- [CONTROLLER_RUNTIME_DEEP_DIVE.md](../../research/CONTROLLER_RUNTIME_DEEP_DIVE.md) - Controller patterns
- [envtest Documentation](https://book.kubebuilder.io/reference/envtest.html)
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)

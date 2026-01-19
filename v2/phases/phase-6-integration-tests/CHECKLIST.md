# Phase 6: Integration Tests (envtest) - Checklist

**Estimated Effort:** 2 days  
**Status:** ✅ Complete

---

## Pre-Implementation

- [ ] Phase 5 complete (unit tests passing with coverage targets)
- [ ] Review existing `internal/controller/suite_test.go`
- [ ] Run `make test` to verify current tests pass
- [ ] Run `make setup-envtest` to download envtest binaries
- [ ] Verify CRDs generated: `make manifests`
- [ ] Review Ginkgo/Gomega patterns for async testing

---

## Day 1: Test Environment Setup & Create Flow Tests

### Task 6.1: Enhance `internal/controller/suite_test.go`

**File:** `internal/controller/suite_test.go`

- [ ] Add batchv1 and corev1 scheme registration
- [ ] Create ctrl.Manager with test configuration
- [ ] Disable metrics server (`BindAddress: "0"`)
- [ ] Register LocustTestReconciler with manager
- [ ] Start manager in goroutine with GinkgoRecover
- [ ] Add timeout/interval constants for Eventually/Consistently
- [ ] Verify manager starts successfully

---

### Task 6.2: Create `internal/controller/integration_test.go`

**File:** `internal/controller/integration_test.go`

Test helpers:
- [ ] Add testNamespace variable with unique per-test names
- [ ] Add BeforeEach to create test namespace
- [ ] Add AfterEach to delete test namespace
- [ ] Create `createLocustTest()` helper function

Create Flow Tests:
- [ ] Test: should create master Service when LocustTest is created
- [ ] Test: should create master Job when LocustTest is created
- [ ] Test: should create worker Job when LocustTest is created
- [ ] Test: should set owner references on created resources
- [ ] Test: should create all resources with correct labels

---

### Task 6.3: Add Create Flow Edge Cases

**File:** `internal/controller/integration_test.go`

- [ ] Test: should handle LocustTest with custom labels
- [ ] Test: should handle LocustTest with custom annotations
- [ ] Test: should handle LocustTest with affinity configuration
- [ ] Test: should handle LocustTest with tolerations
- [ ] Test: should handle LocustTest with single worker
- [ ] Test: should handle LocustTest with maximum workers (500)
- [ ] Test: should handle LocustTest with imagePullSecrets

---

## Day 2: Update NO-OP, Delete Flow & Error Handling Tests

### Task 6.4: Implement Update NO-OP Tests

**File:** `internal/controller/integration_test.go`

- [ ] Test: should NOT create new resources when CR spec is updated
- [ ] Test: should NOT modify worker Job when workerReplicas is changed
- [ ] Test: should NOT modify master Job when masterCommandSeed is changed
- [ ] Test: resource UIDs should remain unchanged after CR update
- [ ] Test: resource ResourceVersions should remain unchanged after CR update

---

### Task 6.5: Implement Delete Flow Tests

**File:** `internal/controller/integration_test.go`

- [ ] Test: should delete all child resources when LocustTest is deleted
- [ ] Test: Service should be deleted via owner reference GC
- [ ] Test: Master Job should be deleted via owner reference GC
- [ ] Test: Worker Job should be deleted via owner reference GC
- [ ] Test: should handle deletion of non-existent LocustTest gracefully

---

### Task 6.6: Implement Error Handling Tests

**File:** `internal/controller/integration_test.go`

- [ ] Test: should handle idempotent resource creation (resources exist)
- [ ] Test: should create resources in different namespaces independently
- [ ] Test: should handle rapid create/delete cycles
- [ ] Test: should not fail when reconciling already-deleted CR

---

## Verification

### Test Execution

```bash
# Run all tests
make test

# Run only integration tests
go test -v ./internal/controller/... -ginkgo.v

# Run specific test by name
go test -v ./internal/controller/... -ginkgo.focus="Create Flow"
```

- [ ] All integration tests pass: `make test`
- [ ] No flaky tests (run 3x to verify)
- [ ] Test execution < 2 minutes total
- [ ] Tests work without external dependencies

### CI Verification

- [ ] Tests pass in GitHub Actions
- [ ] envtest binaries auto-downloaded in CI
- [ ] No manual setup required for CI

---

## Post-Implementation

- [ ] All verification steps pass
- [ ] Update `phases/README.md` with Phase 6 status
- [ ] Update `phases/NOTES.md` with any deviations or discoveries
- [ ] Document any known limitations of envtest

---

## Files Summary

| File | Status | Changes |
|------|--------|---------|
| `internal/controller/suite_test.go` | [ ] | Enhanced with manager startup (~50 LOC) |
| `internal/controller/integration_test.go` | [ ] | New file (~400 LOC) |

---

## Quick Reference Commands

```bash
# Setup envtest binaries (required first time)
make setup-envtest

# Generate CRDs (required if types changed)
make manifests

# Run all tests with verbose output
go test -v ./internal/... -ginkgo.v

# Run with specific timeout
go test -v ./internal/controller/... -timeout 5m

# Run only Create Flow tests
go test -v ./internal/controller/... -ginkgo.focus="Create Flow"

# Run only Delete Flow tests
go test -v ./internal/controller/... -ginkgo.focus="Delete Flow"

# Run only Update NO-OP tests
go test -v ./internal/controller/... -ginkgo.focus="Update NO-OP"

# Run only Error Handling tests
go test -v ./internal/controller/... -ginkgo.focus="Error Handling"

# Skip integration tests (run only unit tests)
go test -v ./internal/... -ginkgo.skip="Integration"

# List all test descriptions
go test -v ./internal/controller/... -ginkgo.dry-run

# Debug: show envtest binary location
$(LOCALBIN)/setup-envtest use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path
```

---

## Test Patterns Reference

### Eventually - Wait for async condition

```go
Eventually(func() error {
    return k8sClient.Get(ctx, key, &obj)
}, timeout, interval).Should(Succeed())
```

### Consistently - Verify NO change over time

```go
Consistently(func() string {
    k8sClient.Get(ctx, key, &obj)
    return string(obj.UID)
}, timeout, interval).Should(Equal(originalUID))
```

### BeforeEach/AfterEach - Test isolation

```go
BeforeEach(func() {
    ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: uniqueName}}
    Expect(k8sClient.Create(ctx, ns)).To(Succeed())
})

AfterEach(func() {
    Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
})
```

---

## Known Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No kubelet | Jobs never "run" | Verify specs only, not runtime |
| No scheduler | Pods stay Pending | Verify Pod specs, not status |
| No network | Services don't route | Verify Service specs only |
| Single etcd | No HA testing | Use E2E tests (Phase 15) for HA |

---

## Acceptance Criteria Summary

1. **Create Flow:** All 3 resources (Service, Master Job, Worker Job) created with correct specs
2. **Owner References:** All child resources have owner reference to LocustTest CR
3. **Update NO-OP:** Spec changes do NOT modify existing resources
4. **Delete Cascade:** Deleting CR triggers automatic cleanup of all children
5. **Error Handling:** Controller handles edge cases gracefully
6. **Test Quality:** No flaky tests, < 2 min execution, CI-compatible

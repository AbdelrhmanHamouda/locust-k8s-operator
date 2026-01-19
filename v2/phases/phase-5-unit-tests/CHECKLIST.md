# Phase 5: Unit Tests - Checklist

**Estimated Effort:** 2 days  
**Status:** ✅ Completed

---

## Pre-Implementation

- [ ] Phase 4 complete (reconciler implemented and building)
- [ ] Review existing test files in `locust-k8s-operator-go/internal/`
- [ ] Run `make test` to establish baseline
- [ ] Run coverage report to identify gaps: `go test -coverprofile=coverage.out ./internal/...`
- [ ] Review Java test patterns in `/src/test/java/`

---

## Day 1: Resource Builder Tests

### Task 5.1: Enhance `internal/resources/job_test.go`

**File:** `internal/resources/job_test.go`

- [ ] Add test for `BuildMasterJob()` with all optional fields populated
- [ ] Add test for `BuildWorkerJob()` with all optional fields populated
- [ ] Add test for Kafka environment variable injection
- [ ] Add test for custom port configurations
- [ ] Add test for nil/empty ConfigMap handling
- [ ] Add test for RestartPolicy verification
- [ ] Add test for BackoffLimit verification
- [ ] Add test for Completions field verification
- [ ] Add edge case: very long CR name (truncation behavior)
- [ ] Add edge case: special characters in CR name

---

### Task 5.2: Enhance `internal/resources/service_test.go`

**File:** `internal/resources/service_test.go`

- [ ] Add test for all master ports (5557, 5558, 8089, 8080, 9646)
- [ ] Add test for Service selector labels
- [ ] Add test for Service type (ClusterIP)
- [ ] Add test for custom metrics port configuration
- [ ] Add test for port naming conventions

---

### Task 5.3: Enhance `internal/resources/command_test.go`

**File:** `internal/resources/command_test.go`

- [ ] Add test for master command with all flags
- [ ] Add test for worker command with master host
- [ ] Add test for command argument ordering
- [ ] Add test for special characters in command seed
- [ ] Add test for empty command seed handling
- [ ] Add test for command string splitting behavior

---

### Task 5.4: Enhance `internal/resources/labels_test.go`

**File:** `internal/resources/labels_test.go`

- [ ] Add test for empty user labels
- [ ] Add test for label key conflicts (user vs operator)
- [ ] Add test for annotation key conflicts
- [ ] Add test for nil Labels/Annotations in spec
- [ ] Add test for very long label values

---

### Task 5.5: Create Test Fixtures

**Directory:** `internal/testdata/`

- [ ] Create `internal/testdata/` directory
- [ ] Create `locusttest_minimal.json` - minimal valid LocustTest CR
- [ ] Create `locusttest_full.json` - CR with all optional fields
- [ ] Create `locusttest_with_affinity.json` - CR with affinity config
- [ ] Create `locusttest_with_tolerations.json` - CR with tolerations
- [ ] Create `expected_master_job.json` - expected master Job output
- [ ] Create `expected_worker_job.json` - expected worker Job output
- [ ] Create `expected_master_service.json` - expected Service output

---

## Day 2: Controller & Config Tests

### Task 5.6: Rewrite `internal/controller/locusttest_controller_test.go`

**File:** `internal/controller/locusttest_controller_test.go`

Unit tests (using fake client):
- [ ] Test `Reconcile()` returns success when CR not found (deleted)
- [ ] Test `Reconcile()` creates resources on new CR (generation=1)
- [ ] Test `Reconcile()` is NO-OP on update (generation>1)
- [ ] Test `createResource()` handles IsAlreadyExists gracefully
- [ ] Test `createResource()` sets owner reference correctly
- [ ] Test event recording on resource creation

Error handling tests:
- [ ] Test `Reconcile()` handles API fetch error
- [ ] Test `createResource()` handles creation error
- [ ] Test `createResource()` handles SetControllerReference error

---

### Task 5.7: Review `internal/config/config_test.go`

**File:** `internal/config/config_test.go`

- [ ] Verify all env var combinations tested
- [ ] Add test for partial config (some vars set, some default)
- [ ] Add test for very large TTL values
- [ ] Add test for float-like string to int conversion failure
- [ ] Verify test isolation (env vars cleaned between tests)

---

### Task 5.8: Add Helper Test Utilities

**File:** `internal/testutil/testutil.go` (new, optional)

- [ ] Create `NewTestLocustTest()` helper (reusable across packages)
- [ ] Create `NewTestConfig()` helper
- [ ] Create `AssertJobEquals()` helper for deep comparison
- [ ] Create `LoadTestFixture()` helper for JSON fixtures

---

## Verification

### Coverage Check

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./internal/...

# Check coverage by package
go tool cover -func=coverage.out | grep -E "^(github.com.*internal/resources/|github.com.*internal/controller/|github.com.*internal/config/|total:)"
```

- [ ] `internal/resources/` coverage ≥ 80%
- [ ] `internal/controller/` coverage ≥ 70%
- [ ] `internal/config/` coverage ≥ 80%

### Test Quality

- [ ] All tests pass: `make test`
- [ ] No test flakiness (run 3x)
- [ ] Tests complete in < 30 seconds
- [ ] No race conditions: `go test -race ./internal/...`
- [ ] Tests don't require external dependencies

### Code Quality

- [ ] Tests follow table-driven pattern where appropriate
- [ ] Test names are descriptive (`TestBuildMasterJob_WithTTL`)
- [ ] Edge cases documented in test names
- [ ] No hardcoded magic values (use constants)
- [ ] Assertions use `assert`/`require` appropriately

---

## Post-Implementation

- [ ] All verification steps pass
- [ ] Coverage report generated and saved
- [ ] Update `phases/README.md` with Phase 5 status
- [ ] Update `phases/NOTES.md` with any deviations
- [ ] Document any discovered issues or edge cases

---

## Files Modified Summary

| File | Status | Changes |
|------|--------|---------|
| `internal/resources/job_test.go` | [ ] | +~50 LOC (new tests) |
| `internal/resources/service_test.go` | [ ] | +~30 LOC (new tests) |
| `internal/resources/command_test.go` | [ ] | +~40 LOC (new tests) |
| `internal/resources/labels_test.go` | [ ] | +~30 LOC (new tests) |
| `internal/controller/locusttest_controller_test.go` | [ ] | Rewrite (~200 LOC) |
| `internal/testdata/*.json` | [ ] | New fixtures |

---

## Quick Reference Commands

```bash
# Run all tests with verbose output
go test -v ./internal/...

# Run specific test by name
go test -v -run TestBuildMasterJob ./internal/resources/

# Run tests with coverage for single package
go test -cover ./internal/resources/

# Generate coverage HTML report
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# List uncovered lines
go tool cover -func=coverage.out | grep -v "100.0%"

# Run with race detector
go test -race ./internal/...

# Run benchmarks (if any)
go test -bench=. ./internal/resources/
```

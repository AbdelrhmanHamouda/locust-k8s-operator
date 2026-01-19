# Phase 5: Unit Tests

**Status:** 🔲 Not Started  
**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Dependencies:** Phase 4 (Core Reconciler)

---

## Overview

Port and enhance unit tests from Java to Go, achieving **80% code coverage** for `internal/resources/` and **70% coverage** for `internal/controller/`. This phase validates that the Go implementation produces identical behavior to the Java version through comprehensive testing.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Current Test Status

Existing test files in the Go codebase:

| File | Lines | Coverage Status |
|------|-------|-----------------|
| `internal/resources/job_test.go` | ~350 | Comprehensive |
| `internal/resources/service_test.go` | ~130 | Good |
| `internal/resources/labels_test.go` | ~208 | Good |
| `internal/resources/command_test.go` | ~75 | Basic |
| `internal/config/config_test.go` | ~298 | Comprehensive |
| `internal/controller/locusttest_controller_test.go` | ~94 | Minimal (Ginkgo scaffold) |

**Phase 5 Goal:** Review, enhance, and fill gaps to meet coverage targets.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Test Framework** | `testing` + `testify` for resources, Ginkgo for controller | Match existing patterns, Ginkgo for envtest integration |
| **Table-Driven Tests** | Use `[]struct{}` test tables | Idiomatic Go, easy to add cases |
| **Test Fixtures** | JSON files in `internal/testdata/` | Reusable, version-controlled expected outputs |
| **Mock Strategy** | Fake client from controller-runtime | Standard for K8s controller testing |
| **Coverage Tool** | `go test -cover` | Built-in, CI-compatible |

## Source of Truth

### Java Tests to Port
- `/src/test/java/com/locust/operator/controller/` - Controller tests
- Integration tests in `/src/integrationTest/` - Behavior reference

### Go Test Patterns
- Existing tests in `internal/resources/*_test.go`
- Ginkgo patterns in `internal/controller/suite_test.go`

## Test Categories

### 1. Resource Builder Tests (Day 1)

| Test File | Focus Areas |
|-----------|-------------|
| `job_test.go` | Job structure, containers, volumes, command args, affinity, tolerations |
| `service_test.go` | Service ports, selectors, type configuration |
| `labels_test.go` | NodeName generation, label merging, annotation building |
| `command_test.go` | Master/worker command construction, argument ordering |

### 2. Controller & Config Tests (Day 2)

| Test File | Focus Areas |
|-----------|-------------|
| `config_test.go` | Default values, env overrides, type conversion |
| `locusttest_controller_test.go` | Reconcile on create, NO-OP on update, resource creation |

## Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/resources/job_test.go` | Enhance | Add edge cases, volume mount tests |
| `internal/resources/service_test.go` | Enhance | Add port validation tests |
| `internal/resources/command_test.go` | Enhance | Add comprehensive argument tests |
| `internal/controller/locusttest_controller_test.go` | Rewrite | Full reconciliation flow tests |
| `internal/testdata/` | Create | JSON fixtures for expected outputs |
| `internal/testdata/locusttest_minimal.json` | Create | Minimal valid CR fixture |
| `internal/testdata/locusttest_full.json` | Create | Full-featured CR fixture |

## Acceptance Criteria

1. **Coverage Targets:**
   - `internal/resources/` ≥ 80%
   - `internal/controller/` ≥ 70%
   - `internal/config/` ≥ 80%

2. **Test Quality:**
   - All tests pass: `make test`
   - No flaky tests
   - Tests run in < 30 seconds (excluding envtest)

3. **Behavior Verification:**
   - Generated Job specs match Java output structure
   - Generated Service specs match Java output structure
   - Command construction matches Java behavior

## Test Commands

```bash
# Run all tests
make test

# Run with coverage
go test -coverprofile=coverage.out ./internal/...
go tool cover -func=coverage.out

# Run specific package tests
go test -v ./internal/resources/...
go test -v ./internal/controller/...
go test -v ./internal/config/...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run with race detection
go test -race ./internal/...
```

## Coverage Gaps to Address

Based on existing tests, these areas need additional coverage:

### `internal/resources/`
- [ ] Port configuration edge cases (custom ports)
- [ ] Volume mount path validation
- [ ] Empty/nil field handling
- [ ] Kafka environment variable injection
- [ ] Resource requirement parsing edge cases

### `internal/controller/`
- [ ] Reconcile with missing ConfigMap
- [ ] Reconcile with invalid spec
- [ ] Event recording verification
- [ ] Owner reference verification
- [ ] IsAlreadyExists handling path

### `internal/config/`
- [ ] Partial environment variable sets
- [ ] Invalid format handling

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 5 definition (lines 268-316)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §7.1 Testing Requirements (80% coverage)
- [analysis/TECHNICAL.md](../../analysis/TECHNICAL.md) - §4.2.1 Unit Test Patterns
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

# Phase 15: E2E Tests (Kind)

**Effort:** 2 days  
**Priority:** P1 - Should Have  
**Status:** Complete  
**Dependencies:** Phase 14 (CI/CD Pipeline)

---

## Objective

Implement comprehensive end-to-end tests using Kind cluster to validate the full operator lifecycle, including LocustTest CR reconciliation, v1 backward compatibility, and OpenTelemetry integration.

---

## Requirements Reference

- **ROADMAP.md §Phase 15:** E2E Tests (Kind)
- **REQUIREMENTS.md §7.1:** Testing Requirements (E2E)

---

## Background

### Current State

Basic E2E test infrastructure exists:
- `test/e2e/e2e_suite_test.go` - Suite setup with CertManager installation
- `test/e2e/e2e_test.go` - Basic operator deployment and metrics tests
- `.github/workflows/go-test-e2e.yml` - E2E workflow in CI
- Makefile targets: `test-e2e`, `setup-test-e2e`, `cleanup-test-e2e`

### What's Missing

The existing E2E tests only verify:
1. Operator pod is running
2. Metrics endpoint is accessible

**Not yet tested:**
- LocustTest CR creation and reconciliation
- Master/Worker Job creation verification
- Service creation verification
- v1 API backward compatibility
- v2 API with new features (env injection, volumes)
- OpenTelemetry configuration
- Test completion lifecycle
- Status updates and conditions

---

## Scope

### In Scope

- **LocustTest CR lifecycle tests** - Create, verify resources, delete
- **v1 backward compatibility tests** - Ensure v1 CRs still work
- **v2 feature tests** - Environment injection, volume mounting
- **OpenTelemetry integration tests** - OTel flag and env vars
- **Status verification tests** - Phase transitions, conditions
- **Negative tests** - Invalid CRs, validation webhook

### Out of Scope

- Performance/load testing
- Multi-cluster tests
- Actual Locust load test execution (we verify pod creation, not test results)

---

## Key Deliverables

| File | Action | Description |
|------|--------|-------------|
| `test/e2e/e2e_test.go` | **Extend** | Add LocustTest CR lifecycle tests |
| `test/e2e/locusttest_e2e_test.go` | **Create** | Dedicated LocustTest E2E tests |
| `test/e2e/v1_compatibility_test.go` | **Create** | v1 API backward compatibility |
| `test/e2e/otel_e2e_test.go` | **Create** | OpenTelemetry integration tests |
| `test/e2e/testdata/` | **Create** | Sample CRs for E2E tests |
| `.github/workflows/go-test-e2e.yml` | **Enhance** | Add coverage reporting |

---

## Success Criteria

1. E2E tests pass locally with `make test-e2e`
2. E2E tests pass in CI pipeline
3. Tests complete in < 10 minutes
4. Coverage reported to Codecov (if applicable)
5. All critical operator paths covered:
   - CR create → Jobs/Service created
   - CR delete → Resources cleaned up
   - v1 CR → Works via conversion
   - Invalid CR → Rejected by webhook

---

## Test Categories

### 1. Core Lifecycle Tests
- Create LocustTest v2 CR
- Verify master Service created
- Verify master Job created
- Verify worker Job created
- Verify owner references set
- Delete CR and verify cleanup

### 2. v1 Backward Compatibility
- Create LocustTest v1 CR
- Verify conversion webhook works
- Verify resources created correctly
- Verify deprecation warning in events

### 3. Environment Injection Tests
- Create CR with ConfigMap refs
- Create CR with Secret refs
- Create CR with inline env vars
- Verify env vars in pod specs

### 4. Volume Mounting Tests
- Create CR with custom volumes
- Verify volumes mounted correctly
- Verify target filtering (master/worker/both)

### 5. OpenTelemetry Tests
- Create CR with OTel enabled
- Verify `--otel` flag in command
- Verify OTel env vars injected
- Verify metrics sidecar NOT deployed

### 6. Validation Tests
- Submit invalid CR (missing required fields)
- Submit CR with reserved path conflict
- Verify webhook rejects invalid CRs

### 7. Status Tests
- Verify phase transitions (Pending → Running)
- Verify conditions are set correctly

---

## Quick Start

```bash
# Run E2E tests locally (requires Kind)
cd locust-k8s-operator-go
make test-e2e

# Run specific test
go test ./test/e2e/ -v -ginkgo.focus="LocustTest lifecycle"

# Run with existing cluster (skip Kind setup)
KIND_CLUSTER=my-cluster go test ./test/e2e/ -v
```

---

## Related Documents

- [DESIGN.md](./DESIGN.md) - Technical design and test architecture
- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist

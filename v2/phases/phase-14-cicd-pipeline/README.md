# Phase 14: CI/CD Pipeline

**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Status:** Complete  
**Dependencies:** Phase 13 (Helm Chart Updates)

---

## Objective

Update GitHub Actions workflows to build, test, and release the Go operator, replacing Gradle/JVM-based pipelines with Go tooling. Add multi-arch Docker image support and modernize the release process.

---

## Requirements Reference

- **ROADMAP.md §Phase 14:** CI/CD Pipeline
- **REQUIREMENTS.md §6.3:** CI/CD Pipeline

---

## Background

### Current State (Java Operator)

The existing CI/CD uses Java/Gradle tooling:
- **ci.yaml:** Gradle build, JaCoCo coverage, Helm chart lint/test
- **release.yaml:** Jib Docker build, Helm chart release
- JDK 21 setup, Gradle wrapper validation
- Single-arch image builds

### Target State (Go Operator)

The Go operator needs different tooling:
- Go build/test with coverage
- `golangci-lint` for code quality
- Multi-arch Docker images (amd64, arm64)
- `ko` for image building (Go-native, fast, no Dockerfile needed)
- Existing Makefile targets integration

### Existing Go Workflows (Partial)

Some Go workflows already exist but are scoped to `locust-k8s-operator-go/`:
- `go-lint.yml` - golangci-lint
- `go-test.yml` - Unit tests
- `go-test-e2e.yml` - E2E tests with Kind

These will be **replaced** by rewritten main workflows.

---

## Scope

### In Scope

- **Rewrite** `ci.yaml` entirely for Go (remove Java)
- **Rewrite** `release.yaml` for Go Docker builds
- **Delete** all Java-specific workflows
- Add coverage reporting for Go
- Multi-arch image support
- Update Makefile targets for CI

### Out of Scope

- OLM bundle CI (Phase 17)
- Complex release automation (GoReleaser)

---

## Key Deliverables

| File | Action | Description |
|------|--------|-------------|
| `.github/workflows/ci.yaml` | **Rewrite** | Go build, lint, test, coverage |
| `.github/workflows/release.yaml` | **Rewrite** | ko build, multi-arch |
| `.github/workflows/go-lint.yml` | **Delete** | Replaced by ci.yaml |
| `.github/workflows/go-test.yml` | **Delete** | Replaced by ci.yaml |
| `.github/workflows/integration-test.yml` | **Delete** | Java-specific |
| `locust-k8s-operator-go/Makefile` | Modify | Add CI-specific targets |

---

## Success Criteria

1. CI passes on PR with Go code changes
2. Docker image builds and pushes on release
3. Multi-arch images available (amd64, arm64)
4. Coverage reported to Codecov
5. Helm chart lint/test still works

---

## Strategy: Consolidate vs Replace

**Decision:** Complete replacement of Java CI/CD with Go tooling. Clean slate.

| Approach | Pros | Cons |
|----------|------|------|
| **Replace immediately (chosen)** | Clean slate, no legacy debt | N/A - Java operator deprecated |
| Consolidate | Gradual migration | Unnecessary complexity |
| Parallel pipelines | Safe | Maintenance burden |

### Implementation

1. **Rewrite** `ci.yaml` for Go (remove all Java/Gradle steps)
2. **Rewrite** `release.yaml` for Go Docker builds
3. **Delete** redundant `go-*.yml` files
4. **Delete** `integration-test.yml` (Java-specific)

---

## Quick Start

```bash
# After implementation, verify locally:
cd locust-k8s-operator-go

# Lint
make lint

# Test with coverage
make test

# Build multi-arch image locally
make docker-buildx IMG=lotest/locust-k8s-operator:dev

# Verify CI config
act -j build-go  # Using act for local CI testing
```

---

## Related Documents

- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist
- [DESIGN.md](./DESIGN.md) - Technical design and workflow structure

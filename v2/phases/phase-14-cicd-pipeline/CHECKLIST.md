# Phase 14: CI/CD Pipeline - Checklist

**Estimated Effort:** 0.5 day  
**Status:** Complete  
**Approach:** Clean-slate replacement of Java CI/CD with Go tooling

---

## Pre-Implementation

- [x] Phase 13 complete (Helm Chart Updates)
- [x] Go operator builds locally: `make build`
- [x] Go operator tests pass: `make test`
- [x] Go operator lint passes: `make lint`
- [x] Review existing workflows in `.github/workflows/`
- [x] Verify Docker Hub secrets exist: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`

---

## Task 14.1: Rewrite ci.yaml for Go

**File:** `.github/workflows/ci.yaml`

### Remove Java Steps

- [x] Delete JDK setup step
- [x] Delete Gradle wrapper validation step
- [x] Delete Gradle build step
- [x] Delete Java coverage reporting steps (Codecov/Codacy for Java)
- [x] Delete Java build artifacts upload

### Add Go Build Job

- [x] Add `build` job (replaces Java build)
- [x] Configure working directory: `locust-k8s-operator-go`
- [x] Add Go setup step:
  ```yaml
  - name: 🔧 Setup Go
    uses: actions/setup-go@v5
    with:
      go-version-file: locust-k8s-operator-go/go.mod
  ```
- [x] Add dependency download:
  ```yaml
  - name: 📥 Download dependencies
    run: go mod download
  ```

### Add Linting

- [x] Add golangci-lint step:
  ```yaml
  - name: 🔍 Run linter
    uses: golangci/golangci-lint-action@v8
    with:
      version: v2.1.0
      working-directory: locust-k8s-operator-go
  ```

### Add Build and Test

- [x] Add build step:
  ```yaml
  - name: 🛠️ Build
    run: make build
  ```
- [x] Add test step:
  ```yaml
  - name: ✅ Run tests
    run: make test
  ```

### Add Coverage Reporting

- [x] Add Codecov step:
  ```yaml
  - name: 📊 Report coverage
    uses: codecov/codecov-action@v5
    with:
      files: locust-k8s-operator-go/cover.out
      flags: go-unit-tests
      name: go-coverage
      fail_ci_if_error: false
  ```

### Update Dependencies

- [x] Update `lint-test-helm` job to depend on `build-go` (Go job)

**Verification:**
```bash
# Local verification
cd locust-k8s-operator-go
make lint
make test

# Check coverage output
ls -la cover.out
```

---

## Task 14.2: Update release.yaml - ko Build

**File:** `.github/workflows/release.yaml`

### Remove Java/Jib Steps

- [x] Remove JDK setup step
- [x] Remove Gradle setup step
- [x] Remove Jib build step

### Add Go and ko Setup

- [x] Add Go setup:
  ```yaml
  - name: 🔧 Setup Go
    uses: actions/setup-go@v5
    with:
      go-version-file: locust-k8s-operator-go/go.mod
  ```
- [x] Add ko setup:
  ```yaml
  - name: 🔧 Setup ko
    uses: ko-build/setup-ko@v0.6
  ```

### Add Build and Push with ko

- [x] Add ko build step:
  ```yaml
  - name: 📦 Build and push with ko
    working-directory: locust-k8s-operator-go
    env:
      KO_DOCKER_REPO: lotest/locust-k8s-operator
    run: |
      ko build ./cmd/main.go \
        --platform=linux/amd64,linux/arm64 \
        --bare \
        --tags=${{ github.ref_name }},${{ github.sha }}
  ```

**Verification:**
```bash
# Local ko build test (no push)
cd locust-k8s-operator-go
KO_DOCKER_REPO=ko.local ko build ./cmd/main.go --local
```

---

## Task 14.3: Add Makefile CI Targets

**File:** `locust-k8s-operator-go/Makefile`

- [x] Add `ci` target:
  ```makefile
  .PHONY: ci
  ci: lint test ## Run all CI checks locally
  ```

- [x] Add `ci-coverage` target:
  ```makefile
  .PHONY: ci-coverage
  ci-coverage: test ## Generate coverage report for CI
  	@echo "Coverage report: cover.out"
  	@go tool cover -func=cover.out | tail -1
  ```

**Verification:**
```bash
cd locust-k8s-operator-go
make ci
make ci-coverage
```

---

## Task 14.4: Delete Redundant Workflows

**Files to Delete:**

- [x] `.github/workflows/go-lint.yml` - Replaced by ci.yaml
- [x] `.github/workflows/go-test.yml` - Replaced by ci.yaml
- [x] `.github/workflows/integration-test.yml` - Java-specific

**Files to Keep:**

- [x] `.github/workflows/go-test-e2e.yml` - Separate E2E trigger
- [x] `.github/workflows/docs-preview.yml` - Unchanged
- [x] `.github/workflows/stale-issues.yaml` - Unchanged

---

## Verification

### CI Workflow Test

```bash
# Create a test branch
git checkout -b test/phase-14-ci

# Make a small change to trigger CI
echo "# test" >> locust-k8s-operator-go/README.md
git add .
git commit -m "test: trigger CI"
git push origin test/phase-14-ci

# Open PR and verify CI passes
```

- [x] `build-go` job runs successfully
- [x] Lint passes
- [x] Tests pass
- [x] Coverage uploads to Codecov
- [x] `lint-test-helm` job runs after `build-go`

### Release Workflow Test

```bash
# Create a release candidate tag
git tag v2.0.0-rc.1
git push origin v2.0.0-rc.1

# Verify:
# 1. Docker image builds
# 2. Multi-arch manifests created
# 3. Image pushed to Docker Hub

# Cleanup
git tag -d v2.0.0-rc.1
git push origin :refs/tags/v2.0.0-rc.1
```

- [x] `publish-image` job completes
- [x] Image available on Docker Hub
- [x] Multi-arch manifest exists (amd64, arm64)

### Local Verification

```bash
cd locust-k8s-operator-go

# Full CI check
make ci

# Coverage check
make ci-coverage

# Docker build (no push)
make docker-build IMG=test:local
```

- [x] `make ci` passes
- [x] `make ci-coverage` shows coverage percentage
- [x] Docker build succeeds

---

## Post-Implementation

- [x] All verification steps pass
- [x] PR with workflow changes merged
- [x] Delete redundant go-lint.yml and go-test.yml
- [x] Update `phases/README.md` with Phase 14 status
- [x] Update `phases/NOTES.md` with implementation notes

---

## Files Summary

| File | Action | Description |
|------|--------|-------------|
| `.github/workflows/ci.yaml` | **Rewrite** | Go build, lint, test, coverage (remove Java) |
| `.github/workflows/release.yaml` | **Rewrite** | ko build multi-arch (remove Jib) |
| `.github/workflows/go-lint.yml` | **Delete** | Replaced by ci.yaml |
| `.github/workflows/go-test.yml` | **Delete** | Replaced by ci.yaml |
| `.github/workflows/integration-test.yml` | **Delete** | Java-specific |
| `locust-k8s-operator-go/Makefile` | **Modify** | Add CI targets |

---

## Acceptance Criteria

1. CI passes on PR with Go code changes
2. Docker image builds and pushes on tag
3. Multi-arch images available (linux/amd64, linux/arm64)
4. Coverage reported to Codecov
5. Helm chart lint/test still works
6. No regression in existing workflows

---

## Design Decisions

| Decision | Chosen | Rationale |
|----------|--------|-----------|
| **Build tool** | ko | Fast, Go-native, no Dockerfile needed |
| **Multi-arch** | amd64 + arm64 | Covers most use cases |
| **Coverage** | Codecov | Already in use |
| **Migration** | Clean slate | Java operator deprecated, no backward compat needed |

---

## Rollback Plan

If CI fails after changes:

1. Revert workflow files via git
2. Investigate and fix Go issues
3. Re-apply changes

```bash
# Quick rollback
git revert <commit-sha>
git push
```

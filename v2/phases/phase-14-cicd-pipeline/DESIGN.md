# Phase 14: CI/CD Pipeline - Technical Design

**Version:** 1.0  
**Status:** Draft  
**Approach:** Clean-slate replacement of Java CI/CD with Go tooling

---

## Design Philosophy

**Clean slate.** Completely replace Java/Gradle workflows with Go tooling. No gradual migration—the Java operator is deprecated.

---

## 1. Current Workflow Analysis

### 1.1 Existing Workflows

| Workflow | Purpose | Action |
|----------|---------|--------|
| `ci.yaml` | Java build, Helm lint, docs test | **Rewrite** - Go only |
| `release.yaml` | Jib Docker, Helm release, docs | **Rewrite** - ko build |
| `go-lint.yml` | golangci-lint | **Delete** - Replaced |
| `go-test.yml` | Go unit tests | **Delete** - Replaced |
| `go-test-e2e.yml` | E2E with Kind | **Keep** - Separate trigger |
| `integration-test.yml` | Java integration | **Delete** - Java-specific |
| `docs-preview.yml` | Docs preview | **Keep** - Unchanged |
| `stale-issues.yaml` | Issue management | **Keep** - Unchanged |

### 1.2 Current ci.yaml Structure

```yaml
jobs:
  build:              # Java/Gradle build + coverage
  lint-test-helm:     # Helm chart testing
  docs-test:          # MkDocs build
```

### 1.3 Current release.yaml Structure

```yaml
jobs:
  Publish-image:      # Jib Docker build + push
  helm-chart-release: # Helm chart release
  docs-release:       # MkDocs deploy
```

---

## 2. Target Workflow Design

### 2.1 New ci.yaml Structure

```yaml
name: 🤖 CI Pipeline

on:
  push:
    branches: [master, main]
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]

permissions: read-all

jobs:
  # ============================================
  # Go Operator Build & Test
  # ============================================
  build-go:
    name: 🏗️ Build Go Operator
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: locust-k8s-operator-go
    steps:
      - name: 📂 Checkout repo
        uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
          repository: ${{ github.event.pull_request.head.repo.full_name }}

      - name: 🔧 Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: locust-k8s-operator-go/go.mod

      - name: 📥 Download dependencies
        run: go mod download

      - name: 🔍 Run linter
        uses: golangci/golangci-lint-action@v8
        with:
          version: v2.1.0
          working-directory: locust-k8s-operator-go

      - name: 🛠️ Build
        run: make build

      - name: ✅ Run tests
        run: make test

      - name: 📊 Report coverage
        uses: codecov/codecov-action@v5
        with:
          files: locust-k8s-operator-go/cover.out
          flags: go-unit-tests
          name: go-coverage
          fail_ci_if_error: false

  # ============================================
  # Helm Chart Lint & Test
  # ============================================
  lint-test-helm:
    name: 🌊 Lint & Test chart
    runs-on: ubuntu-latest
    needs: [build-go]  # Changed from Java build
    steps:
      # ... (existing Helm chart testing steps unchanged)

  # ============================================
  # Documentation Test
  # ============================================
  docs-test:
    name: 📚 Test documentation
    runs-on: ubuntu-latest
    needs: [lint-test-helm]
    steps:
      # ... (existing docs testing steps unchanged)

  # Java build job removed - Go operator only
```

### 2.2 New release.yaml Structure

```yaml
name: 🚀 Publish image & helm

on:
  push:
    tags:
      - "*"

jobs:
  # ============================================
  # Go Operator Docker Image
  # ============================================
  publish-image:
    name: 🐳 Publish image
    runs-on: ubuntu-latest
    env:
      DOCKER_IMAGE: lotest/locust-k8s-operator
    steps:
      - name: 📂 Checkout repo
        uses: actions/checkout@v4

      - name: 🔧 Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: locust-k8s-operator-go/go.mod

      - name: 🔧 Setup ko
        uses: ko-build/setup-ko@v0.6

      - name: 🔐 Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: 📦 Build and push with ko
        working-directory: locust-k8s-operator-go
        env:
          KO_DOCKER_REPO: ${{ env.DOCKER_IMAGE }}
        run: |
          ko build ./cmd/main.go \
            --platform=linux/amd64,linux/arm64 \
            --bare \
            --tags=${{ github.ref_name }},${{ github.sha }}

  # ============================================
  # Helm Chart Release (Unchanged)
  # ============================================
  helm-chart-release:
    # ... (existing steps unchanged)

  # ============================================
  # Documentation Release (Unchanged)
  # ============================================
  docs-release:
    # ... (existing steps unchanged)
```

---

## 3. Multi-Architecture Support

### 3.1 Platform Matrix

| Platform | Support | Notes |
|----------|---------|-------|
| `linux/amd64` | ✅ Required | Primary platform |
| `linux/arm64` | ✅ Required | ARM servers, Apple Silicon |
| `linux/s390x` | ❌ Optional | Mainframe, low demand |
| `linux/ppc64le` | ❌ Optional | Power, low demand |

### 3.2 Build Approaches

| Approach | Pros | Cons | Decision |
|----------|------|------|----------|
| **ko** | Fast, no Dockerfile, Go-native | Less flexible | ✅ Chosen |
| Docker Buildx | Full control | Slower, more setup | Not needed |
| GoReleaser | Full automation | Complex setup | Overkill for now |

### 3.3 ko Configuration

```yaml
- name: 🔧 Setup ko
  uses: ko-build/setup-ko@v0.6

- name: 📦 Build and push
  env:
    KO_DOCKER_REPO: lotest/locust-k8s-operator
  run: |
    ko build ./cmd/main.go \
      --platform=linux/amd64,linux/arm64 \
      --bare \
      --tags=${{ github.ref_name }}
```

**Benefits of ko:**
- No Dockerfile needed (uses Go build directly)
- Automatic distroless base image
- Built-in SBOM generation
- Faster builds (no Docker daemon)
- Native multi-arch support

---

## 4. Coverage Reporting

### 4.1 Go Coverage Output

The Makefile `test` target produces `cover.out`:

```makefile
test: manifests generate fmt vet setup-envtest generate-test-crds
	KUBEBUILDER_ASSETS="..." go test $$(go list ./... | grep -v /e2e) -coverprofile cover.out
```

### 4.2 Codecov Configuration

```yaml
- name: 📊 Report coverage
  uses: codecov/codecov-action@v5
  with:
    files: locust-k8s-operator-go/cover.out
    flags: go-unit-tests
    name: go-coverage
    fail_ci_if_error: false
```

### 4.3 Coverage Threshold

Maintain minimum coverage via codecov.yml (if exists) or accept current baseline.

---

## 5. Makefile Updates

### 5.1 New CI Targets

Add to `locust-k8s-operator-go/Makefile`:

```makefile
##@ CI

.PHONY: ci
ci: lint test ## Run all CI checks locally

.PHONY: ci-coverage
ci-coverage: test ## Generate coverage report for CI
	@echo "Coverage report: cover.out"
	@go tool cover -func=cover.out | tail -1

.PHONY: docker-build-ci
docker-build-ci: ## Build Docker image for CI (no push)
	$(CONTAINER_TOOL) build -t ${IMG} ./
```

### 5.2 Existing Targets Used in CI

| Target | Purpose | Used In |
|--------|---------|---------|
| `make build` | Build binary | ci.yaml |
| `make test` | Run tests + coverage | ci.yaml |
| `make lint` | Run golangci-lint | ci.yaml |
| `make docker-buildx` | Multi-arch build | release.yaml (or Buildx action) |
| `make manifests` | Generate CRDs | Pre-release |

---

## 6. Trigger Conditions

### 6.1 CI Triggers

```yaml
on:
  push:
    branches: [master, main]
  pull_request:
    types: [opened, synchronize, reopened, ready_for_review]
```

### 6.2 Release Triggers

```yaml
on:
  push:
    tags:
      - "v*"  # Semantic version tags
```

### 6.3 Path Filters (Optional)

Optionally filter CI runs to relevant paths:

```yaml
on:
  push:
    paths:
      - 'locust-k8s-operator-go/**'
      - 'charts/**'
      - '.github/workflows/ci.yaml'
```

---

## 7. Secret Requirements

### 7.1 Required Secrets

| Secret | Purpose | Already Exists |
|--------|---------|----------------|
| `DOCKERHUB_USERNAME` | Docker Hub login | ✅ Yes |
| `DOCKERHUB_TOKEN` | Docker Hub token | ✅ Yes |
| `GITHUB_TOKEN` | Helm/docs release | ✅ Auto |
| `CODACY_PROJECT_TOKEN` | Codacy coverage | ✅ Yes (optional) |

### 7.2 No New Secrets Required

All required secrets already exist in the repository.

---

## 8. Implementation Steps

### Single Phase - Clean Replacement

1. **Rewrite** `ci.yaml` with Go build/lint/test (remove all Java steps)
2. **Rewrite** `release.yaml` with Docker Buildx (remove Jib)
3. **Delete** `go-lint.yml` and `go-test.yml` (replaced by ci.yaml)
4. **Delete** `integration-test.yml` (Java-specific)

---

## 9. Testing the Workflows

### 9.1 Local Testing with act

```bash
# Install act
brew install act  # macOS

# Test CI workflow
act -j build-go -P ubuntu-latest=catthehacker/ubuntu:act-latest

# Test with secrets
act -j build-go --secret-file .secrets
```

### 9.2 PR Testing

Create a draft PR with workflow changes to test without affecting main.

### 9.3 Release Testing

1. Create a test tag: `v2.0.0-rc.1`
2. Verify image builds and pushes
3. Delete test release/tag

---

## 10. Rollback Plan

If Go CI fails:

1. Revert workflow files via git
2. Investigate and fix Go issues
3. Re-apply changes

---

## 11. References

- [Docker Buildx Action](https://github.com/docker/build-push-action)
- [Codecov GitHub Action](https://github.com/codecov/codecov-action)
- [golangci-lint Action](https://github.com/golangci/golangci-lint-action)
- [Kubernetes Operator SDK CI](https://sdk.operatorframework.io/docs/building-operators/golang/testing/)

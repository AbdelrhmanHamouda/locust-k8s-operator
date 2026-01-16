# Phase 0: Project Scaffolding - Implementation Plan

**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Prerequisites:** None  
**Requirements:** §3.1 Technology Stack, §3.4 Project Structure

---

## Objective

Initialize the Go operator project using Operator SDK with proper structure, ensuring it aligns with the project requirements and matches the target architecture defined in `analysis/TECHNICAL.md`.

---

## Prerequisites Checklist

Before starting, ensure the following tools are installed:

| Tool | Minimum Version | Verification Command |
|------|-----------------|---------------------|
| Go | 1.22+ | `go version` |
| Operator SDK | 1.37+ | `operator-sdk version` |
| Make | Any | `make --version` |
| Docker/Podman | Latest | `docker version` or `podman version` |
| kubectl | 1.27+ | `kubectl version --client` |

### Installation Commands (if needed)

```bash
# Go (macOS)
brew install go

# Operator SDK (macOS)
brew install operator-sdk

# Or download directly
export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.37.0
curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
chmod +x operator-sdk_${OS}_${ARCH}
sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk
```

---

## Tasks

### Task 0.1: Create Go Operator Directory

Create a dedicated directory for the Go operator within the repository.

```bash
mkdir -p locust-k8s-operator-go
cd locust-k8s-operator-go
```

**Rationale:** Keep Go implementation separate from Java during migration. Can be merged later or kept as parallel directory.

---

### Task 0.2: Initialize Operator SDK Project

```bash
operator-sdk init \
  --domain locust.io \
  --repo github.com/AbdelrhmanHamouda/locust-k8s-operator \
  --plugins go/v4
```

**Expected Output:**
- `go.mod` with module path `github.com/AbdelrhmanHamouda/locust-k8s-operator`
- `go.sum` with dependencies
- `cmd/main.go` - operator entrypoint
- `Makefile` with standard targets
- `Dockerfile` for building operator image
- `config/` directory with RBAC, manager, and default configurations
- `PROJECT` file with operator-sdk metadata
- `.gitignore` file

**Verification:**
```bash
cat PROJECT
# Should show domain: locust.io and repo path
```

---

### Task 0.3: Create v1 API Scaffold

```bash
operator-sdk create api \
  --group locust \
  --version v1 \
  --kind LocustTest \
  --resource \
  --controller
```

**Expected Output:**
- `api/v1/locusttest_types.go` - CRD type definitions (skeleton)
- `api/v1/groupversion_info.go` - API group registration
- `api/v1/zz_generated.deepcopy.go` - Auto-generated deep copy methods
- `internal/controller/locusttest_controller.go` - Controller skeleton
- `internal/controller/suite_test.go` - Test suite setup
- `config/crd/bases/` - CRD YAML (after `make manifests`)
- `config/samples/locust_v1_locusttest.yaml` - Sample CR
- `config/rbac/` - Updated RBAC rules

**Verification:**
```bash
ls -la api/v1/
ls -la internal/controller/
```

---

### Task 0.4: Configure go.mod Dependencies

Ensure required dependencies are present with correct versions:

```bash
# Update dependencies
go mod tidy

# Verify key dependencies
go list -m all | grep -E "(controller-runtime|client-go|apimachinery)"
```

**Required Dependencies:**
| Package | Purpose |
|---------|---------|
| `sigs.k8s.io/controller-runtime` | Controller framework |
| `k8s.io/api` | Kubernetes API types |
| `k8s.io/apimachinery` | Kubernetes meta types |
| `k8s.io/client-go` | Kubernetes client |

---

### Task 0.5: Verify Project Builds

```bash
# Build the operator binary
make build

# Verify binary created
ls -la bin/manager
```

**Expected:** Binary compiles without errors.

---

### Task 0.6: Verify CRD Generation

```bash
# Generate manifests (CRD, RBAC)
make manifests

# Verify CRD created
cat config/crd/bases/locust.io_locusttests.yaml
```

**Expected:** CRD YAML generated with:
- `apiVersion: apiextensions.k8s.io/v1`
- `kind: CustomResourceDefinition`
- `group: locust.io`
- `names.kind: LocustTest`
- `versions[0].name: v1`

---

### Task 0.7: Verify Test Framework

```bash
# Run tests (will pass with empty test suite)
make test
```

**Expected:** Test suite runs, even with no custom tests.

---

### Task 0.8: Create Project-Specific Directory Structure

Create additional directories required by the project structure (§3.4):

```bash
# Create internal packages (will be populated in later phases)
mkdir -p internal/config
mkdir -p internal/resources
mkdir -p test/e2e

# Create placeholder files
touch internal/config/.gitkeep
touch internal/resources/.gitkeep
touch test/e2e/.gitkeep
```

---

### Task 0.9: Update Makefile (Optional Enhancements)

Add helpful targets to the generated Makefile:

```makefile
# Add to Makefile

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Run go fmt
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy
```

---

## Acceptance Criteria

| Criteria | Verification Command | Expected Result |
|----------|---------------------|-----------------|
| Project compiles | `make build` | Binary at `bin/manager` |
| CRD generates | `make manifests` | CRD at `config/crd/bases/locust.io_locusttests.yaml` |
| Tests run | `make test` | Exit code 0 |
| Directory structure matches §3.4 | `tree -L 3` | Structure as defined |
| go.mod correct | `head go.mod` | Module: `github.com/AbdelrhmanHamouda/locust-k8s-operator` |

---

## Output Artifacts

After Phase 0 completion, the following files/directories will exist:

```
locust-k8s-operator-go/
├── api/
│   └── v1/
│       ├── groupversion_info.go
│       ├── locusttest_types.go
│       └── zz_generated.deepcopy.go
├── cmd/
│   └── main.go
├── config/
│   ├── crd/
│   │   ├── bases/
│   │   │   └── locust.io_locusttests.yaml
│   │   └── kustomization.yaml
│   ├── default/
│   │   └── kustomization.yaml
│   ├── manager/
│   │   ├── kustomization.yaml
│   │   └── manager.yaml
│   ├── rbac/
│   │   ├── kustomization.yaml
│   │   ├── leader_election_role.yaml
│   │   ├── leader_election_role_binding.yaml
│   │   ├── locusttest_editor_role.yaml
│   │   ├── locusttest_viewer_role.yaml
│   │   ├── role.yaml
│   │   ├── role_binding.yaml
│   │   └── service_account.yaml
│   └── samples/
│       ├── kustomization.yaml
│       └── locust_v1_locusttest.yaml
├── internal/
│   ├── config/
│   │   └── .gitkeep
│   ├── controller/
│   │   ├── locusttest_controller.go
│   │   └── suite_test.go
│   └── resources/
│       └── .gitkeep
├── test/
│   └── e2e/
│       └── .gitkeep
├── .dockerignore
├── .gitignore
├── Dockerfile
├── Makefile
├── PROJECT
├── go.mod
└── go.sum
```

---

## Notes for Next Phase

Phase 1 (v1 API Types) will:
1. Populate `api/v1/locusttest_types.go` with all v1 fields matching Java CRD
2. Add kubebuilder validation markers
3. Run `make manifests` to regenerate CRD with full schema
4. Verify CRD matches existing Java CRD

---

## Troubleshooting

### Issue: `operator-sdk: command not found`
**Solution:** Install Operator SDK using the commands in Prerequisites section.

### Issue: `go: module requires Go 1.22`
**Solution:** Upgrade Go version: `brew upgrade go` or download from golang.org.

### Issue: `make manifests` fails with controller-gen error
**Solution:** Run `make controller-gen` first, or ensure Go bin is in PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: Test fails with missing envtest binaries
**Solution:** Run `make envtest` to download test binaries:
```bash
make envtest
```

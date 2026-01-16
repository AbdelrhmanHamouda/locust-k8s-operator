# Phase 0: Project Scaffolding - Checklist

Quick reference checklist for Phase 0 tasks.

---

## Prerequisites

- [ ] Go 1.22+ installed (`go version`)
- [ ] Operator SDK 1.37+ installed (`operator-sdk version`)
- [ ] Make installed (`make --version`)
- [ ] Docker/Podman available (`docker version`)

---

## Tasks

- [ ] **0.1** Create `locust-k8s-operator-go/` directory
- [ ] **0.2** Run `operator-sdk init` with correct domain and repo
- [ ] **0.3** Run `operator-sdk create api` for LocustTest v1
- [ ] **0.4** Run `go mod tidy` and verify dependencies
- [ ] **0.5** Verify build: `make build`
- [ ] **0.6** Verify CRD generation: `make manifests`
- [ ] **0.7** Verify tests run: `make test`
- [ ] **0.8** Create additional directories (`internal/config`, `internal/resources`, `test/e2e`)
- [ ] **0.9** (Optional) Add lint/fmt targets to Makefile

---

## Acceptance Criteria

- [ ] `make build` succeeds (binary at `bin/manager`)
- [ ] `make manifests` generates CRD at `config/crd/bases/locust.io_locusttests.yaml`
- [ ] `make test` runs without errors
- [ ] Directory structure matches §3.4 from REQUIREMENTS.md
- [ ] `go.mod` has module `github.com/AbdelrhmanHamouda/locust-k8s-operator`

---

## Quick Commands

```bash
# Full setup sequence
mkdir -p locust-k8s-operator-go && cd locust-k8s-operator-go

operator-sdk init \
  --domain locust.io \
  --repo github.com/AbdelrhmanHamouda/locust-k8s-operator \
  --plugins go/v4

operator-sdk create api \
  --group locust \
  --version v1 \
  --kind LocustTest \
  --resource \
  --controller

go mod tidy

make build
make manifests
make test

mkdir -p internal/config internal/resources test/e2e
```

---

## Definition of Done

Phase 0 is complete when:
1. All checklist items are marked complete
2. All acceptance criteria pass
3. Output artifacts exist as documented in IMPLEMENTATION.md

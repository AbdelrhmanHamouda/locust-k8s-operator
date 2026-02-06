# Locust K8s Operator — Go Rewrite Polish

## What This Is

A comprehensive fix pass on the Go rewrite of the Locust Kubernetes Operator (PR #274, `feat/go` -> `master`). The operator manages distributed Locust load tests on Kubernetes via a LocustTest CRD. Eight expert reviewers identified 9 critical, ~25 major, and 50+ minor issues. This project addresses every single one, moves the Go code from `locust-k8s-operator-go/` to repo root, and removes the legacy Java operator code — making the PR merge-ready and production-perfect.

## Core Value

Every API field the operator exposes must actually work, and the operator must never crash or silently ignore user configuration.

## Requirements

### Validated

- ✓ Kubernetes operator using controller-runtime v0.21.0 — existing
- ✓ LocustTest CRD with v2 (storage) and v1 (deprecated) API versions — existing
- ✓ Conversion webhooks for v1 <-> v2 — existing
- ✓ Validation webhooks for CR creation/update — existing
- ✓ Master/worker Job-based test execution — existing
- ✓ Status subresource with phase tracking (Pending/Running/Succeeded/Failed) — existing
- ✓ Owner references for automatic garbage collection — existing
- ✓ OpenTelemetry support for native Locust metrics export — existing
- ✓ Volume mounting with target filtering — existing
- ✓ Environment variable and secret injection — existing
- ✓ Helm chart with webhook support and security hardening — existing
- ✓ Comprehensive test pyramid (~150 unit, ~20 integration, ~30 E2E) — existing
- ✓ CI/CD pipelines for build, test, and release — existing
- ✓ Multi-arch container images via ko — existing
- ✓ Documentation with migration guide, API reference, and developer docs — existing

### Active

- [ ] Wire `extraArgs` into master/worker command builders (C1)
- [ ] Wire CR-level `resources` with fallback to operator config (C2)
- [ ] Replace `resource.MustParse()` with safe parsing and error handling (C3)
- [ ] Unify volume name constants across packages (C4)
- [ ] Fix `ko build` path in release pipeline (C5)
- [ ] Remove fragile `Generation > 1` reconciliation guard (C6)
- [ ] Fix Helm CRD conversion webhook architecture (C7)
- [ ] Verify/fix `readOnlyRootFilesystem` with webhook cert writes (C8)
- [ ] Keep `failurePolicy: Fail` but ensure operator availability during upgrades (C9)
- [ ] Remove dead code: `isJobComplete()`/`isJobFailed()` (M1)
- [ ] Fix mount path trailing slash inconsistency (M2)
- [ ] Fix error message typo + duplicate logging in main (M3)
- [ ] Align `ko` action versions across workflows (M4)
- [ ] Add `timeout-minutes` to all CI/CD jobs (M5)
- [ ] Fix `ct.yaml` timeout (600s -> reasonable value) (M6)
- [ ] Address chart-releaser fork supply chain risk (M7)
- [ ] Add `permissions:` block to `go-test-e2e.yml` (M8)
- [ ] Pin Kind download with checksum verification (M9)
- [ ] Replace `go mod tidy` mutation with verification check (M10)
- [ ] Remove nested `.github/workflows/` under Go subdir (M11)
- [ ] Fix `imagePullSecrets` format in v2 doc examples (M12)
- [ ] Add missing `image` field in getting started example (M13)
- [ ] Fix status example with non-existent fields (M14)
- [ ] Fix operator CPU limit mismatch in migration docs (M15)
- [ ] Fix wrong working directory in dev docs (M16)
- [ ] Fix namespace inconsistency across docs (M17)
- [ ] Add CRD conversion webhook config to Helm chart (M18)
- [ ] Tighten Helm RBAC (remove unnecessary ConfigMap/Secret CRUD) (M19)
- [ ] Add recovery when owned resources are externally deleted (M20)
- [ ] Fix backward compatibility helpers in Helm templates (M21)
- [ ] Remove or wire `crd.install` value (M22)
- [ ] Fix `replicaCount` default (2 -> 1 for backward compat) (M23)
- [ ] Fix CRD symlink for `helm package` in CI (M24)
- [ ] Suppress Prometheus scrape annotations when OTel enabled (M25)
- [ ] Fix E2E conversion script label selectors (M26)
- [ ] Fix `os.Chdir()` global side effect in test utils (M27)
- [ ] All 56 minor issues from reviewer feedback (see pr_review.md Minor Issues section)
- [ ] Move Go code from `locust-k8s-operator-go/` to repo root
- [ ] Remove legacy Java operator code

### Out of Scope

- New features beyond what reviewers identified — this is a polish pass, not feature work
- Kubernetes version upgrades or dependency bumps beyond what's needed for fixes
- Redesigning the operator architecture — it's already approved by reviewers
- Python locust test files or example load test scripts

## Context

- **Origin**: Complete rewrite from Java (Micronaut + JOSDK) to Go (controller-runtime)
- **PR #274**: `feat/go` -> `master`, 238 files changed, +56,350 / -959 lines
- **Review team**: 8 expert reviewers (Architecture, Go, GitHub Actions, Documentation, Kubernetes, Locust, Helm, Testing)
- **Aggregate verdict**: REQUEST CHANGES — architecturally sound, targeted fixes needed
- **Existing test coverage**: ~150 unit, ~20 integration, ~30 E2E scenarios
- **Codebase location**: Currently in `locust-k8s-operator-go/` subdirectory, moving to repo root
- **Individual reviews**: Available in `reviews/` directory for detailed context per domain
- **Compiled review**: `pr_review.md` has the full prioritized issue list

## Constraints

- **Branch**: All work on `feat/go` branch — PR must stay mergeable to `master`
- **No breaking API changes**: v2 CRD API surface is established; wire existing fields, don't redesign
- **Webhook policy**: Keep `failurePolicy: Fail` per maintainer decision — ensure operator HA instead
- **Backward compatibility**: v1 API must continue to work through conversion webhooks
- **Test pyramid**: Maintain existing test coverage; add tests for new behavior (wired fields, error handling)
- **Review alignment**: Fixes must match reviewer recommendations — individual review files in `reviews/` are the source of truth

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Wire `extraArgs` and `resources` (not remove) | Users expect API fields to work; removing would be a breaking change | — Pending |
| Keep `failurePolicy: Fail` for webhooks | Strict validation is more important than availability during operator downtime | — Pending |
| Move Go code to repo root | Operator is the primary codebase now; subdirectory adds unnecessary complexity | — Pending |
| Remove Java code | Go rewrite replaces it entirely; keeping it adds confusion | — Pending |
| All review items in one pass | Comprehensive fix before merge; no tech debt carried forward | — Pending |

---
*Last updated: 2026-02-06 after initialization*

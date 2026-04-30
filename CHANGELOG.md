## 2.2.2 (unreleased)

### Fix

- **operator**: fix CrashLoopBackOff on default Helm install (#317). The
  webhook readiness check added in 2.2.0 unconditionally called
  `mgr.GetWebhookServer()`, which has the side effect of starting the
  webhook server even when webhooks were disabled — and the chart
  deliberately mounts no TLS certs in that path. The readiness gate is now
  registered only when `--enable-webhooks=true`, and webhook setup is
  driven by an explicit flag instead of an implicit env-var gate.
- **operator**: bound the webhook cert wait. Previously the operator could
  spin forever if cert files never appeared; it now fails fast after
  `--webhook-cert-wait-timeout` (default 2m) with a clear error message
  pointing at cert-manager configuration.

### Feat

- **operator**: add `--enable-webhooks` (default false) and
  `--webhook-cert-wait-timeout` (default 2m) flags. The Helm chart now
  passes `--enable-webhooks={{ .Values.webhook.enabled }}` from
  `deployment.yaml`, replacing the previous `ENABLE_WEBHOOKS` env var.
- **operator**: the binary continues to honour `ENABLE_WEBHOOKS` as a
  deprecated alias for one release and logs a deprecation warning when it
  is set. Slated for removal in 2.3.0.
- **ci**: add three regression-prevention jobs to `.github/workflows/ci.yaml`:
  `lint-test-helm` (now ungated; runs `ct install` against
  `charts/locust-k8s-operator/ci/{default,webhook-enabled}-values.yaml`
  with cert-manager preinstalled); `helm-default-smoke` (every PR — fresh
  default-values install asserting no restarts); and `helm-upgrade-smoke`
  (helm install v2.1.1 → upgrade to fix-branch). The default-values
  scenario is the one that would have caught #317.
- **tests**: add unit tests in `cmd/main_test.go` enforcing the
  invariant that `addHealthChecks` does NOT register a `webhook` readyz
  check when webhooks are disabled — the structural guarantee against #317
  recurring.

## 2.2.1 (2026-04-27)

### Fix

- **ci**: repin chart-releaser-action to v1.5.0 to unbreak publish

## 2.2.0 (2026-04-27)

### Feat

- add configurable security context for OpenShift compatibility (#304)

### Fix

- **security**: bump Go directive to 1.26.2 to clear stdlib CVEs (#310)

## 2.1.1 (2026-02-16)

### Fix

- **security**: resolve GitHub code scanning alert #1 (#291)
- **docs**: correct Codacy badge URLs for code quality and coverage (#280)

## 2.1.0 (2026-02-14)

### Feat

- **ci**: add Codacy coverage reporting to CI pipeline

### Fix

- add Artifact Hub metadata (#279)
- **ci**: remove excessive SHA tag from Docker image releases (#278)

## 2.0.0 (2026-02-13)

### Feat

- GO rewrite  (#274)

## 1.1.1 (2025-07-04)

## 1.1.0 (2025-07-03)

### Feat

- add support for Locust Lib ConfigMap (#243)

## 1.0.0 (2025-06-27)

### Feat

- disable resource limits for workers when config value is blank (#237)

## 0.11.0 (2024-10-26)

### Feat

- add metadata.namespace for deployment (#223)

## 0.10.0 (2024-09-16)

### Feat

- Support cluster role

## 0.9.1 (2024-07-04)

### Fix

- Grant job patch permission (#206)
- revert micronaut version (#205)
- Github actions failing to build (#204)

## 0.9.0 (2024-07-03)

### Feat

- implement pull secrets in helm chart  (#192)
- implement pull secrets in helm chart

Co-authored-by: Marcial White <marcial.white+gitlab@wizards.com>
- migrate resource creation from `createOrReplace` to `serverSideApply` (#165)

### Fix

- Stop Github actions from triggering twice in PRs (#161)

## 0.8.0 (2023-08-25)

### Feat

- **#132**: Fully configure the Metrics Exporter based on HELM (or default) configuration (#134)

### Fix

- **#126**: Migrate to JOSDK v4.4.1 (#127)

## 0.7.0 (2023-04-22)

### Feat

- **#89**: add support for pulling the Locust image from private registries (#98) by @jachinte

## 0.6.0 (2023-04-21)

### Feat

- **#13**: Add `managed-by` label to generated resources (#104)
- **#52**: add a TTL period to deployed jobs (#97) by @jachinte

## 0.5.0 (2023-01-27)

### Feat

- **#78**: Support HELM control to enable/disable injecting Affinity & Taint tolerations information from Custom Resources (#84)
- **#78**: Support adding taint tolerations to pods (#83)
- **#78**: Support adding node affinity through the custom resource (#81)

## 0.4.0 (2022-12-02)

### Feat

- **69**: Allow configmap volume mounts to be writable (#70)

## 0.3.0 (2022-11-30)

### Feat

- **65**: Allow for "-" to be part of the metadata name (#66)

## 0.2.3 (2022-11-30)

### Fix

- **63**: Correctly apply k8s service selector (#64)
- **63**: Correctly apply k8s service selector

## 0.2.2 (2022-11-29)

### Fix

- **58**: update the chart image tag to not override the app version (#62)
- **60**: update containersol/locust_exporter version to v0.5.0 (#61)

## 0.2.1 (2022-11-03)

### Fix

- **#53**: Restore release workflow permissions

## 0.2.0 (2022-11-03)

### Feat

- **#32**: add labels and annotations to master and worker pods (#45)

## 0.1.0 (2022-10-08)

### Feat

- **10**: Support HELM
- **10**: Support helm deployment for the operator
- support loadtest from configMap  (#15)
- Add codacy coverage reporter
- Operator reacts to `onAdd`, `onDelete` & noop `onUpdate` events + tessts
- CRD design

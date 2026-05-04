# Migration from Java to Go

## Overview

The Locust Kubernetes Operator has been completely rewritten from Java to Go. This represents a full architectural transition, not a simple port. While the core functionality remains the same, the implementation is fundamentally different.

## What Changed

### Language and Framework
- **Before**: Java with Micronaut framework and Java Operator SDK
- **After**: Go with Operator SDK / controller-runtime (Kubernetes standard operator framework)

### Project Structure
- **Before**: Java source in `src/`, Gradle build files, Maven dependencies
- **After**: Go source at repository root (`cmd/`, `api/`, `internal/`), Go modules

### Performance Characteristics
- **Memory footprint**: ~256MB (Java) → ~64MB (Go)
- **Startup time**: ~60s (Java) → <1s (Go)
- **Binary size**: ~100MB (Java + JVM) → ~30MB (Go static binary)

## What Stayed the Same

### API Compatibility
The `LocustTest` Custom Resource Definition (CRD) maintains full backward compatibility:
- Both v1 and v2 API versions are supported
- Existing manifests continue to work without modification
- Helm chart values remain compatible (with new optional features added)

### Behavior
The operator provides the same functionality:
- Creates master and worker Pods with Locust
- Manages Services for master UI and headless communication
- Handles ConfigMap-based Locust script injection
- Supports Secret and environment variable configuration

### Deployment
The Helm chart remains at `charts/locust-k8s-operator/` with the same installation process.

## Finding the Old Java Code

The Java operator source code has been preserved in the `archive/java-operator-v1` branch for reference.

To access it:

```bash
git fetch origin archive/java-operator-v1
git checkout archive/java-operator-v1
```

This branch contains the complete Java codebase as it existed before the Go rewrite. It is maintained for historical reference only and will not receive further updates.

## Key Differences for Developers

### Testing
- **Before**: JUnit 5, Mockito, Testcontainers
- **After**: Go's testing package, envtest for controller tests, Kind for E2E tests

### Build System
- **Before**: Gradle with multi-stage Dockerfile
- **After**: Make with standard Go build commands, multi-arch builds via BuildKit

### Dependencies
- **Before**: Maven Central packages
- **After**: Go modules from Go package ecosystem

### CI/CD
The CI/CD pipelines have been updated to use Go tooling:
- `go build`, `go test`, `go vet`, `golangci-lint` instead of Gradle tasks
- Multi-platform Docker builds (amd64 + arm64)
- Helm chart testing remains unchanged

## Migration for Users

Most users will not need to make any changes. The Go operator is a drop-in replacement for the Java operator:

1. Update the operator deployment via Helm (same chart, new appVersion)
2. Existing `LocustTest` resources continue to function
3. Review new features in v2.0 (OpenTelemetry, enhanced volumes, separate resources per role)

For detailed migration guidance, see the [Migration Guide](https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/) in the documentation.

## v2.2 → v2.2.2

### Recovering from a crashlooping v2.2.0 / v2.2.1 install

If your operator pod is currently in `CrashLoopBackOff` with this error:

```
ERROR setup problem running manager {"error": "open /tmp/k8s-webhook-server/serving-certs/tls.crt: no such file or directory"}
```

— this is [issue #317](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/317).
A `helm upgrade` to v2.2.2 fixes it:

```bash
helm repo update
helm upgrade locust-operator locust-k8s-operator/locust-k8s-operator \
  --version 2.2.2 \
  --namespace locust-system \
  --reuse-values
```

If `helm upgrade --wait` times out because the old pod is not Ready, force a
new ReplicaSet rollout: `kubectl -n locust-system rollout restart deploy/locust-operator`.

### Webhook configuration: env var → flag

The operator binary previously gated webhook registration on the
`ENABLE_WEBHOOKS` environment variable, which the Helm chart set from
`.Values.webhook.enabled`. As of v2.2.2 this is replaced by an explicit
`--enable-webhooks` command-line flag. The chart's `deployment.yaml` now
passes `--enable-webhooks={{ .Values.webhook.enabled }}` directly.

**For chart users**: no action required. `helm upgrade` rolls out the new
flag automatically.

**For users running the binary directly or with custom Kustomize / ArgoCD
overlays that set `ENABLE_WEBHOOKS`**: the env var is still honoured for one
release as a deprecated alias and the binary logs a deprecation warning each
time the binary starts. Migrate to `--enable-webhooks=true|false` before
v2.3.0, when the env var will be removed. Empty values (`ENABLE_WEBHOOKS=""`)
are now ignored — the prior `envVal != "false"` check would have silently
enabled webhooks on a typo. Valid values are those accepted by Go's
`strconv.ParseBool`: `1`/`t`/`T`/`TRUE`/`true`/`True` and the matching false
forms. Anything else is logged and ignored.

When both are set, the explicit flag wins.

### Breaking: `--enable-webhooks` default is now `false`

The binary default flipped from on (the master `os.Getenv("ENABLE_WEBHOOKS") != "false"`
check defaulted to enabled when the env var was unset) to off (the new flag
default is `false`). The Helm chart was already explicit, so chart users see
no behaviour change.

This affects you only if you run the binary **directly** (no chart) or via
**kustomize without the webhook overlay** — your install loses admission
webhooks unless you pass `--enable-webhooks=true` (or set the env var).

### Breaking: `--webhook-cert-path` is now mandatory when webhooks are on

If you pass `--enable-webhooks=true` (or `ENABLE_WEBHOOKS=true`) without
`--webhook-cert-path`, the binary now exits at startup with:

```
--webhook-cert-path is required when --enable-webhooks=true
```

Master silently fell back to controller-runtime's default temp dir, which is
the regression vector behind issue #317. The chart already passes
`--webhook-cert-path=/tmp/k8s-webhook-server/serving-certs` whenever
`webhook.enabled=true`, so chart users are unaffected.

### Security posture: admission webhook is off by default

The v2 ValidatingWebhook enforces invariants the reconciler does **not**
re-check, including:

- Reserved secret-mount path conflicts (`/lotest/src`, `/opt/locust/lib`) —
  prevents a CR author from shadowing operator-injected paths.
- Reserved volume name prefixes (`secret-*`, `locust-lib`, `<cr>-master`,
  `<cr>-worker`) — prevents overriding operator-managed volume mounts.
- DNS-label length on the CR name.
- OpenTelemetry config validation.

With `webhook.enabled=false` (the default), CR authors can submit resources
that bypass these checks. If you rely on multi-tenant CR-create RBAC, enable
the webhook and provide TLS certs (see
[Webhook Configuration](https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm_deploy/#webhook-configuration-optional)).

### New flag: `--webhook-cert-wait-timeout`

When webhooks are enabled, the operator polls for the cert files mounted by
cert-manager. Previously this poll was unbounded — a misconfigured
cert-manager would cause the operator to hang silently with no health probe
registered yet. The new `--webhook-cert-wait-timeout` flag (default `2m`)
bounds the wait and exits with a clear error message if certs never appear.
Set to `0` to wait indefinitely (not recommended; a misconfigured cert path
will hang the pod forever).

### Switching `webhook.enabled` from `true` to `false` (CRD downgrade)

The CRD's `spec.conversion` block is rendered only when
`webhook.enabled=true`. If you upgrade an installation from `webhook.enabled=true`
to `webhook.enabled=false`, Helm removes that block on the next upgrade. Any
`LocustTest` objects whose **stored** apiVersion is still `v1` then become
unreadable because the apiserver no longer has a conversion route to `v2`.

**Before flipping `webhook.enabled` to `false`:**

```bash
# 1. Inspect the CRD's storedVersions — these are the on-disk encodings the
#    apiserver still maintains. If "v1" is listed, real v1-encoded objects
#    exist in etcd and would become unreadable after the flip.
kubectl get crd locusttests.locust.io \
  -o jsonpath='{.status.storedVersions}'

# 2. If "v1" is in the output, force every existing object to be re-stored
#    at v2 by reading and re-applying it (apiserver writes back at the
#    storage version):
kubectl get locusttests.locust.io -A -o yaml \
  | kubectl replace --force -f -

# 3. Then drop "v1" from storedVersions (kubectl can't do this directly;
#    use a strategic-merge patch on /status):
kubectl patch crd locusttests.locust.io --subresource status --type=merge \
  -p '{"status":{"storedVersions":["v2"]}}'

# 4. Re-check storedVersions — should print ["v2"] only.
kubectl get crd locusttests.locust.io \
  -o jsonpath='{.status.storedVersions}'
```

Only after `storedVersions` contains `v2` alone is it safe to flip
`webhook.enabled` to `false`.

If you only ever applied LocustTest resources at `apiVersion: locust.io/v2`
**and** `kubectl get crd locusttests.locust.io -o jsonpath='{.status.storedVersions}'`
already prints `["v2"]`, this caveat does not apply to you.

## Rationale

The rewrite to Go was motivated by:

1. **Performance**: Lower memory usage and faster startup align with Kubernetes ecosystem expectations
2. **Ecosystem alignment**: controller-runtime is the de facto standard for Kubernetes operators
3. **Maintainability**: Simpler deployment (static binary), broader contributor pool familiar with Go
4. **Cloud-native fit**: Go is the lingua franca of cloud-native tooling

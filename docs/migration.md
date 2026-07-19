---
title: Upgrade Notes
description: Per-release upgrade and migration notes for the Locust Kubernetes Operator
tags:
  - migration
  - upgrade
  - v2
  - guide
---

# Upgrade Notes

Per-release upgrade and migration notes, newest first.

- [v2.2 → v2.2.2 (webhook default flip)](#v22-v222-webhook-default-flip) — `ENABLE_WEBHOOKS` deprecation, `--enable-webhooks` flag, mandatory `--webhook-cert-path`
- [v1 → v2 (Java/Go rewrite)](#v1-v2-javago-rewrite) — full rewrite walkthrough, including the CRD storedVersions caveat for downgrades

---

## v2.2 → v2.2.2 (webhook default flip)

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
re-check. With `webhook.enabled=false` (the default), CR authors can submit
resources that bypass those checks. See [Security → Admission Webhook](./security.md#admission-webhook-defense-in-depth)
for the invariants enforced and when to enable the webhook.

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

---

## v1 → v2 (Java/Go rewrite)

This guide helps existing users of the Locust Kubernetes Operator upgrade from v1 to v2. The v2 release is a complete rewrite in Go, bringing significant performance improvements and new features.


### Overview

#### Why We Rewrote in Go

The v2 operator was rewritten from Java to Go for several key reasons:

| Aspect | Java (v1) | Go (v2) |
|--------|-----------|---------|
| **Memory footprint** | ~256MB | ~64MB |
| **Startup time** | ~60 seconds | <1 second |
| **Framework** | Java Operator SDK | Operator SDK / controller-runtime |
| **Ecosystem alignment** | Minority | Majority of K8s operators |

#### What Changes for Users

- **API Version:** New `locust.io/v2` API with grouped configuration
- **Backward Compatibility:** v1 CRs continue to work via automatic conversion
- **New Features:** OpenTelemetry, secret injection, volume mounting, separate resource specs
- **Helm Chart:** Updated values structure (backward compatible)

#### Compatibility Guarantees

- **v1 API:** Fully supported via conversion webhook (deprecated, will be removed in v3)
- **Existing CRs:** Work without modification
- **Helm Values:** Backward compatibility shims for common settings

---

### Before You Begin

#### Prerequisites

- Kubernetes 1.29+
- Helm 3.x
- cert-manager v1.14+ (required for conversion webhook)

#### Backup Recommendations

Before upgrading, back up your existing resources:

```bash
# Export all LocustTest CRs
kubectl get locusttests -A -o yaml > locusttests-backup.yaml

# Export operator Helm values
helm get values locust-operator -n <namespace> > values-backup.yaml
```

!!! danger "Critical: Webhook Required for v1 API Compatibility"
    If you have existing v1 `LocustTest` CRs, the conversion webhook is **required** for them to continue working after upgrading to v2. Without it, v1 CRs will fail CRD schema validation.

    You **must**:

    1. Install [cert-manager](https://cert-manager.io/docs/installation/) before upgrading
    2. Enable the webhook during upgrade: `--set webhook.enabled=true`
    3. Verify the webhook is running after upgrade

    If you only use v2 CRs (or are starting fresh), the webhook is optional.

---

### Step 1: Update Helm Chart

#### Upgrade Command

```bash
# Update Helm repository
helm repo update locust-k8s-operator

# Upgrade to v2 (with webhook for v1 CR compatibility)
helm upgrade locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  --version 2.0.0 \
  --set webhook.enabled=true

# If you don't need v1 API compatibility, you can omit --set webhook.enabled=true
```

!!! note "CRD Upgrade"
    Helm automatically upgrades the CRD when using `helm upgrade`. The v2 CRD includes conversion webhook configuration when webhooks are enabled, allowing the API server to convert between v1 and v2 formats transparently.

#### New Helm Values

The v2 chart introduces a cleaner structure. Key changes:

| Old Path (v1) | New Path (v2) | Notes |
|---------------|---------------|-------|
| `config.loadGenerationPods.resource.cpuRequest` | `locustPods.resources.requests.cpu` | Backward compatible |
| `config.loadGenerationPods.resource.memLimit` | `locustPods.resources.limits.memory` | Backward compatible |
| `config.loadGenerationPods.affinity.enableCrInjection` | `locustPods.affinityInjection` | Backward compatible |
| `micronaut.*` | N/A | Removed (Java-specific) |
| `appPort` | N/A | Fixed at 8081 |
| N/A | `webhook.enabled` | New: Enable conversion webhook |
| N/A | `leaderElection.enabled` | New: Enable leader election |

#### Operator Resource Defaults

The Go operator controller requires significantly fewer resources than the Java version:

```yaml
resources:
  limits:
    memory: 256Mi
    cpu: 500m
  requests:
    memory: 64Mi
    cpu: 10m
```

---

### Step 2: Verify Existing CRs

The conversion webhook automatically converts v1 CRs to v2 format when stored. Verify your existing CRs work:

```bash
# List all LocustTests
kubectl get locusttests -A

# Check a specific CR
kubectl describe locusttest <name>
```

#### Verify Conversion

You can read a v1 CR as v2 to verify conversion:

```bash
# Read as v2 (even if created as v1)
kubectl get locusttest <name> -o yaml | grep "apiVersion:"
# Should show: apiVersion: locust.io/v2
```

!!! warning "Deprecation Warning"
    When using the v1 API, you'll see a deprecation warning in kubectl output. This is expected and indicates the conversion webhook is working.

---

### Step 3: Migrate CRs to v2 Format (Recommended)

While v1 CRs continue to work, migrating to v2 format is recommended to access new features.

#### Field Mapping Reference

| v1 Field | v2 Field | Notes |
|----------|----------|-------|
| `masterCommandSeed` | `master.command` | Direct mapping |
| `workerCommandSeed` | `worker.command` | Direct mapping |
| `workerReplicas` | `worker.replicas` | Direct mapping |
| `image` | `image` | No change |
| `imagePullPolicy` | `imagePullPolicy` | No change |
| `imagePullSecrets` | `imagePullSecrets` | No change |
| `configMap` | `testFiles.configMapRef` | Grouped under testFiles |
| `libConfigMap` | `testFiles.libConfigMapRef` | Grouped under testFiles |
| `labels.master` | `master.labels` | Grouped under master |
| `labels.worker` | `worker.labels` | Grouped under worker |
| `annotations.master` | `master.annotations` | Grouped under master |
| `annotations.worker` | `worker.annotations` | Grouped under worker |
| `affinity.nodeAffinity` | `scheduling.affinity` | Uses native K8s Affinity ⚠️[^1] |
| `tolerations` | `scheduling.tolerations` | Uses native K8s Tolerations |
| N/A | `master.resources` | New: Separate resource specs for master |
| N/A | `worker.resources` | New: Separate resource specs for worker |
| N/A | `master.extraArgs` | New: Additional CLI arguments for master |
| N/A | `worker.extraArgs` | New: Additional CLI arguments for worker |
| N/A | `master.autostart` | Auto-added during conversion (default: true) |
| N/A | `master.autoquit` | Auto-added during conversion (enabled: true, timeout: 60s) |

[^1]: **Affinity Conversion Note**: When converting v2 → v1, complex affinity rules may be simplified. Only `NodeSelectorOpIn` operators are preserved, and only the first value from multi-value expressions is kept. Pod affinity/anti-affinity and preferred scheduling rules are not preserved in v1.

#### Example Transformation

=== "v1 Format (Deprecated)"

    ```yaml
    apiVersion: locust.io/v1
    kind: LocustTest
    metadata:
      name: example-test
    spec:
      image: locustio/locust:2.43.3
      masterCommandSeed: --locustfile /lotest/src/test.py --host https://example.com
      workerCommandSeed: --locustfile /lotest/src/test.py
      workerReplicas: 5
      configMap: test-scripts
      labels:
        master:
          team: platform
        worker:
          team: platform
    ```

=== "v2 Format"

    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: example-test
    spec:
      image: locustio/locust:2.43.3
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
        labels:
          team: platform
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 5
        labels:
          team: platform
      testFiles:
        configMapRef: test-scripts
    ```

#### Lossy Conversion Details

!!! warning "V2-Only Fields Not Preserved in V1"
    When reading v2 CRs as v1 (or during rollback to v1), the following v2-exclusive fields **will be lost**:

    **Master/Worker Configuration:**
    
    - `master.resources` - Separate resource specs for master pod
    - `worker.resources` - Separate resource specs for worker pod
    - `master.extraArgs` - Additional CLI arguments for master
    - `worker.extraArgs` - Additional CLI arguments for worker
    - `master.autostart` - Autostart configuration
    - `master.autoquit` - Autoquit configuration

    **Test Files:**
    
    - `testFiles.srcMountPath` - Custom mount path for test files
    - `testFiles.libMountPath` - Custom mount path for library files

    **Scheduling:**
    
    - `scheduling.nodeSelector` - Node selector (v1 only supports nodeAffinity)
    - Complex affinity rules (see warning above)

    **Environment & Secrets:**
    
    - `env.configMapRefs` - ConfigMap environment injection
    - `env.secretRefs` - Secret environment injection
    - `env.variables` - Individual environment variables
    - `env.secretMounts` - Secret file mounts

    **Volumes:**
    
    - `volumes` - Volume definitions
    - `volumeMounts` - Volume mounts with target selection

    **Observability:**
    
    - `observability.openTelemetry` - OpenTelemetry configuration

    **Status:**
    
    - All `status` subresource fields (v1 has no status implementation)

    **Recommendation**: Before rolling back from v2 to v1, backup your v2 CRs to preserve this configuration.

---

### Step 4: Leverage New Features

After migrating to v2, you can use new features:

#### OpenTelemetry Support

```yaml
spec:
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"
```

[:octicons-arrow-right-24: Learn more about OpenTelemetry](how-to-guides/observability/configure-opentelemetry.md)

#### Secret & ConfigMap Injection

```yaml
spec:
  env:
    secretRefs:
      - name: api-credentials
        prefix: "API_"
    configMapRefs:
      - name: app-config
    variables:
      - name: TARGET_HOST
        value: "https://api.example.com"
```

[:octicons-arrow-right-24: Learn more about Environment Injection](how-to-guides/security/inject-secrets.md)

#### Volume Mounting

```yaml
spec:
  volumes:
    - name: test-data
      persistentVolumeClaim:
        claimName: test-data-pvc
  volumeMounts:
    - name: test-data
      mountPath: /data
      target: both  # master, worker, or both
```

[:octicons-arrow-right-24: Learn more about Volume Mounting](how-to-guides/configuration/mount-volumes.md)

#### Separate Resource Specs

```yaml
spec:
  master:
    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
  worker:
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
```

[:octicons-arrow-right-24: Learn more about Separate Resources](api_reference.md)

---

### Troubleshooting

#### Common Issues

##### Conversion Webhook Not Working

**Symptom:** v1 CRs fail with schema validation errors

**Solution:** Ensure cert-manager is installed and the webhook is enabled:

```bash
# Check cert-manager
kubectl get pods -n cert-manager

# Enable webhook in Helm
helm upgrade locust-operator locust-k8s-operator/locust-k8s-operator \
  --set webhook.enabled=true
```

##### Resources Not Created

**Symptom:** LocustTest CR created but no Jobs/Services appear

**Solution:** Check operator logs:

```bash
kubectl logs -n locust-system -l app.kubernetes.io/name=locust-k8s-operator
```

##### Status Not Updating

**Symptom:** LocustTest status remains empty

**Solution:** Verify RBAC permissions include `locusttests/status`:

```bash
kubectl auth can-i update locusttests/status --as=system:serviceaccount:locust-system:locust-operator
```

#### How to Get Help

- [GitHub Issues](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues)

---

### Rollback Procedure

If you need to revert to v1:

```bash
# Rollback Helm release
helm rollback locust-operator <previous-revision> -n locust-system

# Or reinstall v1
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --version 1.1.1 \
  -f values-backup.yaml
```

!!! note
    After rollback, v2-specific fields in CRs will be lost. Ensure you have backups of any v2-only configurations.

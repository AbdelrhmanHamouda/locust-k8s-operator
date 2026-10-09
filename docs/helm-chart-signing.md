---
title: Helm Chart Signing
description: How released Helm charts are signed, and the one-time maintainer setup for the signing key.
tags:
  - helm
  - release
  - security
---

# Helm Chart Signing

Released charts are signed with a dedicated GPG key. The release workflow runs
`helm package --sign`, which writes a provenance file
(`locust-k8s-operator-<version>.tgz.prov`) next to the chart archive.
chart-releaser uploads both to the GitHub release, so the `.prov` is served at
the chart's download URL plus `.prov`.

That's what Artifact Hub checks: when it indexes a chart version, it fetches
`<chart url>.prov` and marks the version as **Signed** if the file contains a PGP
signature. The `artifacthub.io/signKey` annotation in `Chart.yaml` tells
Artifact Hub which key to show users for verification. See the Artifact Hub
[annotations reference](https://artifacthub.io/docs/topics/annotations/helm/)
and Helm's [provenance docs](https://helm.sh/docs/topics/provenance/).

Signing lives in `hack/package-helm-chart.sh`, called from the
`helm-chart-release` job in `.github/workflows/release.yaml`. The job runs in
the `release` GitHub Environment, which holds the signing secrets.

The script fails closed once signing is set up:

- Before setup (no `artifacthub.io/signKey` annotation in `Chart.yaml` and no
  `docs/helm-signing-key.asc`), a missing `HELM_SIGNING_KEY` secret logs a
  warning and the chart is packaged unsigned.
- After setup (either of those exists), a missing `HELM_SIGNING_KEY` fails the
  release instead of publishing an unsigned chart.
- If the `signKey` fingerprint doesn't match the signing key, the job fails
  before packaging.
- After signing, `helm verify` checks the chart against
  `docs/helm-signing-key.asc`, the same key users download, so a signing key
  that isn't in the published file fails the release.

## One-time setup (maintainers)

Run these from the repository root on a machine with GnuPG 2.x and the `gh`
CLI logged in.

### 1. Generate a dedicated signing key

Use RSA: Helm's OpenPGP library doesn't reliably handle the newer ed25519
defaults. GnuPG prompts for a passphrase; pick a strong one and store it in your
password manager.

```bash
KEY_UID="Locust K8s Operator Helm Charts <you@example.com>"
gpg --quick-generate-key "$KEY_UID" rsa4096 sign never

FPR=$(gpg --with-colons --list-secret-keys "$KEY_UID" | awk -F: '$1 == "fpr" { print $10; exit }')
echo "$FPR"
```

GnuPG also writes a revocation certificate to
`~/.gnupg/openpgp-revocs.d/$FPR.rev`. Back that up together with an offline
copy of the private key (`gpg --armor --export-secret-keys "$FPR"`).

### 2. Create the `release` environment

The `helm-chart-release` job runs in a GitHub Environment named `release`, so
the signing secrets are only exposed to jobs that run in that environment. In
the repository, go to **Settings → Environments → New environment** and name it
`release`. Under **Deployment branches and tags**, choose **Selected branches
and tags** and add a rule with ref type **Tag** and the pattern `*`. Releases
are triggered by tag pushes, so this keeps branch workflows (including pull
requests) away from the key.

### 3. Add the environment secrets

The private key is piped straight into `gh`, so it never lands in a file. The
second command prompts for the passphrase, so it stays out of your shell
history.

```bash
gpg --armor --export-secret-keys "$FPR" \
  | gh secret set HELM_SIGNING_KEY --env release --repo AbdelrhmanHamouda/locust-k8s-operator
gh secret set HELM_SIGNING_KEY_PASSPHRASE --env release --repo AbdelrhmanHamouda/locust-k8s-operator
```

Don't also create repository-level secrets with these names: those are
readable by every workflow, which defeats the environment scoping.

### 4. Publish the public key

The docs site copies everything under `docs/` to GitHub Pages, so this file is
served at
<https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm-signing-key.asc>
after the next release.

```bash
gpg --armor --export "$FPR" > docs/helm-signing-key.asc
```

### 5. Add the signKey annotation

Add this under `annotations:` in `charts/locust-k8s-operator/Chart.yaml`,
replacing the fingerprint with the value of `$FPR`:

```yaml
  artifacthub.io/signKey: |
    fingerprint: 0123456789ABCDEF0123456789ABCDEF01234567
    url: https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm-signing-key.asc
```

Commit `docs/helm-signing-key.asc` and the `Chart.yaml` change, and merge them
only after the secrets from step 3 are in place: from then on, a release
without the key fails. The release job also fails if the annotation's
fingerprint doesn't match the signing key, or if the signing key isn't in
`docs/helm-signing-key.asc`.

### 6. Check the first signed release

Verify with the published key, the same way users do:

```bash
VERSION=x.y.z
curl -fsSLO "https://github.com/AbdelrhmanHamouda/locust-k8s-operator/releases/download/locust-k8s-operator-$VERSION/locust-k8s-operator-$VERSION.tgz"
curl -fsSLO "https://github.com/AbdelrhmanHamouda/locust-k8s-operator/releases/download/locust-k8s-operator-$VERSION/locust-k8s-operator-$VERSION.tgz.prov"
curl -fsSL https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm-signing-key.asc \
  | gpg --dearmor > /tmp/locust-helm-pubring.gpg
helm verify "locust-k8s-operator-$VERSION.tgz" --keyring /tmp/locust-helm-pubring.gpg
```

Artifact Hub picks up the change on its next scan of the repository (usually
within 30 minutes). Older chart versions stay unsigned; only versions released
after the secrets are in place get a `.prov` file.

## Rotating the key

Generate a new key, replace both secrets in the `release` environment, and
update the `signKey` fingerprint. Add the new public key to
`docs/helm-signing-key.asc` rather than replacing the old one, so charts signed
with the old key can still be verified:

```bash
gpg --armor --export "$OLD_FPR" "$NEW_FPR" > docs/helm-signing-key.asc
```

`signKey` holds a single fingerprint, so Artifact Hub only shows the current
key. Older versions stay verifiable because the `.asc` keeps both keys.

Update the annotation and the `.asc` in the same change, before the first
release signed with the new key. The release job fails if the fingerprint
doesn't match the signing key or the signing key isn't in the `.asc`.

## Testing signing locally

`hack/package-helm-chart.sh` is the exact script the release job runs. To try
it with a throwaway key that never touches your real keyring, point it at a
copy of the chart and a throwaway public key; otherwise, once signing is set
up, the checks against the real `signKey` annotation and
`docs/helm-signing-key.asc` fail with a throwaway key. Keep `GNUPGHOME` short:
gpg-agent puts its socket there, and long paths (like macOS `$TMPDIR`) exceed
the Unix socket path limit.

```bash
export GNUPGHOME=$(mktemp -d /tmp/gpg.XXXXXX)
TEST_DIR=$(mktemp -d /tmp/chart.XXXXXX)
gpg --batch --pinentry-mode loopback --passphrase test --quick-generate-key "Throwaway <test@example.com>" rsa3072 sign never
cp -R charts/locust-k8s-operator "$TEST_DIR/chart"
gpg --armor --export > "$TEST_DIR/key.asc"
TEST_FPR=$(gpg --with-colons --list-secret-keys | awk -F: '$1 == "fpr" { print $10; exit }')
# Point the copy's signKey annotation (if any) at the throwaway key.
sed -i.bak "s/fingerprint: .*/fingerprint: $TEST_FPR/" "$TEST_DIR/chart/Chart.yaml"
HELM_SIGNING_KEY=$(gpg --batch --pinentry-mode loopback --passphrase test --armor --export-secret-keys) \
HELM_SIGNING_KEY_PASSPHRASE=test VERSION=0.0.0-test \
CHART_DIR="$TEST_DIR/chart" SIGNING_PUBKEY_FILE="$TEST_DIR/key.asc" \
PACKAGE_DIR="$TEST_DIR/pkg" hack/package-helm-chart.sh
gpgconf --kill all
rm -rf "$GNUPGHOME" "$TEST_DIR"
unset GNUPGHOME
```

The script imports the key into its own temporary keyring, signs, runs
`helm verify` against `SIGNING_PUBKEY_FILE`, and deletes its temporary files
on exit. If the copy has no `signKey` annotation, it warns about that and still
signs.

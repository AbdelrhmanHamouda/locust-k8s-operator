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
`helm-chart-release` job in `.github/workflows/release.yaml`. If the
`HELM_SIGNING_KEY` secret isn't set, the script logs a warning and packages
the chart unsigned, so a release never fails because of missing secrets.

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

### 2. Add the repository secrets

The private key is piped straight into `gh`, so it never lands in a file. The
second command prompts for the passphrase, so it stays out of your shell
history.

```bash
gpg --armor --export-secret-keys "$FPR" \
  | gh secret set HELM_SIGNING_KEY --repo AbdelrhmanHamouda/locust-k8s-operator
gh secret set HELM_SIGNING_KEY_PASSPHRASE --repo AbdelrhmanHamouda/locust-k8s-operator
```

### 3. Publish the public key

The docs site copies everything under `docs/` to GitHub Pages, so this file is
served at
<https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm-signing-key.asc>
after the next release.

```bash
gpg --armor --export "$FPR" > docs/helm-signing-key.asc
```

### 4. Add the signKey annotation

Add this under `annotations:` in `charts/locust-k8s-operator/Chart.yaml`,
replacing the fingerprint with the value of `$FPR`:

```yaml
  artifacthub.io/signKey: |
    fingerprint: 0123456789ABCDEF0123456789ABCDEF01234567
    url: https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm-signing-key.asc
```

Commit `docs/helm-signing-key.asc` and the `Chart.yaml` change, and merge them
before tagging the next release. The release job warns if the annotation is
missing or its fingerprint doesn't match the signing key.

### 5. Check the first signed release

```bash
VERSION=x.y.z
curl -fsSLO "https://github.com/AbdelrhmanHamouda/locust-k8s-operator/releases/download/locust-k8s-operator-$VERSION/locust-k8s-operator-$VERSION.tgz"
curl -fsSLO "https://github.com/AbdelrhmanHamouda/locust-k8s-operator/releases/download/locust-k8s-operator-$VERSION/locust-k8s-operator-$VERSION.tgz.prov"
gpg --export "$FPR" > /tmp/locust-helm-pubring.gpg
helm verify "locust-k8s-operator-$VERSION.tgz" --keyring /tmp/locust-helm-pubring.gpg
```

Artifact Hub picks up the change on its next scan of the repository (usually
within 30 minutes). Older chart versions stay unsigned; only versions released
after the secrets are in place get a `.prov` file.

## Rotating the key

Generate a new key, replace both secrets, and update the `signKey`
fingerprint. Append the new public key to `docs/helm-signing-key.asc` rather
than replacing the old one, so charts signed with the old key can still be
verified.

## Testing signing locally

`hack/package-helm-chart.sh` is the exact script the release job runs. To try
it with a throwaway key that never touches your real keyring:

```bash
export GNUPGHOME=$(mktemp -d)
gpg --batch --pinentry-mode loopback --passphrase test --quick-generate-key "Throwaway <test@example.com>" rsa3072 sign never
HELM_SIGNING_KEY=$(gpg --batch --pinentry-mode loopback --passphrase test --armor --export-secret-keys) \
HELM_SIGNING_KEY_PASSPHRASE=test VERSION=0.0.0-test \
PACKAGE_DIR=$(mktemp -d) hack/package-helm-chart.sh
gpgconf --kill all
rm -rf "$GNUPGHOME"
unset GNUPGHOME
```

The script imports the key into its own temporary keyring, signs, runs
`helm verify` on the result, and deletes its temporary files on exit.

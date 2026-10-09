#!/usr/bin/env bash
# Packages the Helm chart for release and, when a signing key is available,
# signs it so a provenance file (<chart>.tgz.prov) is written next to the
# archive. chart-releaser uploads that .prov to the GitHub release, which is
# what Artifact Hub and `helm pull --verify` look for.
#
# Used by .github/workflows/release.yaml. See docs/helm-chart-signing.md.
#
# Signing fails closed once it's set up: if Chart.yaml has an
# artifacthub.io/signKey annotation or the published public key exists, a
# missing HELM_SIGNING_KEY is an error rather than an unsigned release.
#
# Environment:
#   VERSION                      chart version and appVersion (required)
#   HELM_SIGNING_KEY             ASCII-armored private key (required once signing is set up)
#   HELM_SIGNING_KEY_PASSPHRASE  passphrase for that key (optional for an unprotected key)
#   CHART_DIR                    default: charts/locust-k8s-operator
#   PACKAGE_DIR                  default: .cr-release-packages
#   SIGNING_PUBKEY_FILE          published public key; default: docs/helm-signing-key.asc
set -euo pipefail

: "${VERSION:?VERSION must be set}"
CHART_DIR="${CHART_DIR:-charts/locust-k8s-operator}"
PACKAGE_DIR="${PACKAGE_DIR:-.cr-release-packages}"
SIGNING_PUBKEY_FILE="${SIGNING_PUBKEY_FILE:-docs/helm-signing-key.asc}"

warn() {
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    echo "::warning title=Helm chart signing::$1"
  else
    echo "WARNING: $1" >&2
  fi
}

fail() {
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    echo "::error title=Helm chart signing::$1"
  else
    echo "ERROR: $1" >&2
  fi
  exit 1
}

# The artifacthub.io/signKey annotation is what Artifact Hub shows users as the
# key to verify against. Read its fingerprint, normalized to gpg's format
# (no quotes or spaces, uppercase).
has_sign_key_annotation=false
if grep -Eq '^[[:space:]]*artifacthub\.io/signKey:' "$CHART_DIR/Chart.yaml"; then
  has_sign_key_annotation=true
fi
annotated_fingerprint="$(awk '
  /^[[:space:]]*artifacthub\.io\/signKey:/ { in_key = 1; next }
  in_key && /fingerprint:/ { sub(/^[^:]*fingerprint:/, ""); print; exit }
  in_key && /^  [^ ]/ { exit }
' "$CHART_DIR/Chart.yaml" | tr -d "\"'[:space:]" | tr '[:lower:]' '[:upper:]')"

rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

package_args=("$CHART_DIR" --app-version "$VERSION" --version "$VERSION" --destination "$PACKAGE_DIR")

if [[ -z "${HELM_SIGNING_KEY:-}" ]]; then
  if [[ "$has_sign_key_annotation" == true || -f "$SIGNING_PUBKEY_FILE" ]]; then
    fail "HELM_SIGNING_KEY is not set, but signing is set up (artifacthub.io/signKey in $CHART_DIR/Chart.yaml or $SIGNING_PUBKEY_FILE exists). Refusing to publish an unsigned chart. See docs/helm-chart-signing.md."
  fi
  warn "HELM_SIGNING_KEY is not set; packaging the chart without a provenance file. See docs/helm-chart-signing.md."
  helm package "${package_args[@]}"
  exit 0
fi

# Keep the path short: gpg-agent puts its socket in GNUPGHOME, and macOS
# $TMPDIR paths can push it past the Unix socket path limit.
workdir="$(mktemp -d "${RUNNER_TEMP:-/tmp}/helm-sign.XXXXXX")"
export GNUPGHOME="$workdir/gnupg"
cleanup() {
  gpgconf --kill all >/dev/null 2>&1 || true
  rm -rf "$workdir"
}
trap cleanup EXIT

mkdir -m 700 "$GNUPGHOME"
passphrase_file="$workdir/passphrase"
(umask 077 && printf '%s' "${HELM_SIGNING_KEY_PASSPHRASE:-}" >"$passphrase_file")

printf '%s\n' "$HELM_SIGNING_KEY" | gpg --batch --quiet --import

fingerprint="$(gpg --batch --with-colons --list-secret-keys | awk -F: '$1 == "fpr" && !n++ { print $10 }')"
key_uid="$(gpg --batch --with-colons --list-secret-keys | awk -F: '$1 == "uid" && !n++ { print $10 }')"
if [[ -z "$fingerprint" || -z "$key_uid" ]]; then
  echo "HELM_SIGNING_KEY did not contain a usable secret key" >&2
  exit 1
fi
echo "Signing with key $fingerprint ($key_uid)"

# Check the annotation before packaging so a mismatch fails fast.
if [[ "$has_sign_key_annotation" == false ]]; then
  warn "Chart.yaml has no artifacthub.io/signKey annotation; Artifact Hub won't show which key signed this chart."
elif [[ "$annotated_fingerprint" != "$fingerprint" ]]; then
  fail "Chart.yaml signKey fingerprint (${annotated_fingerprint:-none}) doesn't match the signing key ($fingerprint)."
fi

# Helm reads the legacy (pre-GnuPG 2.1) keyring format, so export binary keyrings.
# The exported secret key keeps its passphrase protection.
if ! gpg --batch --pinentry-mode loopback --passphrase-file "$passphrase_file" \
  --export-secret-keys "$fingerprint" >"$workdir/secring.gpg" ||
  [[ ! -s "$workdir/secring.gpg" ]]; then
  echo "Couldn't export the signing key; check HELM_SIGNING_KEY_PASSPHRASE" >&2
  exit 1
fi

# Verify against the public key users download, so a signing key that isn't
# published fails the release. Before that file exists (first-time setup),
# fall back to a self-check against the signing key itself.
if [[ -f "$SIGNING_PUBKEY_FILE" ]]; then
  verify_source="$SIGNING_PUBKEY_FILE"
  gpg --batch --yes --dearmor --output "$workdir/pubring.gpg" "$SIGNING_PUBKEY_FILE"
else
  verify_source="the signing key itself"
  warn "$SIGNING_PUBKEY_FILE doesn't exist; verifying against the signing key itself."
  gpg --batch --export "$fingerprint" >"$workdir/pubring.gpg"
fi

helm package "${package_args[@]}" \
  --sign \
  --key "$key_uid" \
  --keyring "$workdir/secring.gpg" \
  --passphrase-file "$passphrase_file"

for archive in "$PACKAGE_DIR"/*.tgz; do
  echo "Verifying $archive against $verify_source"
  helm verify "$archive" --keyring "$workdir/pubring.gpg" ||
    fail "helm verify failed against $verify_source; is the signing key ($fingerprint) in it?"
done

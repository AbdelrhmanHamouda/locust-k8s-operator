#!/usr/bin/env bash
# Packages the Helm chart for release and, when a signing key is available,
# signs it so a provenance file (<chart>.tgz.prov) is written next to the
# archive. chart-releaser uploads that .prov to the GitHub release, which is
# what Artifact Hub and `helm pull --verify` look for.
#
# Used by .github/workflows/release.yaml. See docs/helm-chart-signing.md.
#
# Environment:
#   VERSION                      chart version and appVersion (required)
#   HELM_SIGNING_KEY             ASCII-armored private key (optional; signing is skipped without it)
#   HELM_SIGNING_KEY_PASSPHRASE  passphrase for that key (optional for an unprotected key)
#   CHART_DIR                    default: charts/locust-k8s-operator
#   PACKAGE_DIR                  default: .cr-release-packages
set -euo pipefail

: "${VERSION:?VERSION must be set}"
CHART_DIR="${CHART_DIR:-charts/locust-k8s-operator}"
PACKAGE_DIR="${PACKAGE_DIR:-.cr-release-packages}"

warn() {
  if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
    echo "::warning title=Helm chart signing::$1"
  else
    echo "WARNING: $1" >&2
  fi
}

rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

package_args=("$CHART_DIR" --app-version "$VERSION" --version "$VERSION" --destination "$PACKAGE_DIR")

if [[ -z "${HELM_SIGNING_KEY:-}" ]]; then
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

# Helm reads the legacy (pre-GnuPG 2.1) keyring format, so export binary keyrings.
# The exported secret key keeps its passphrase protection.
gpg --batch --pinentry-mode loopback --passphrase-file "$passphrase_file" \
  --export-secret-keys "$fingerprint" >"$workdir/secring.gpg"
if [[ ! -s "$workdir/secring.gpg" ]]; then
  echo "Couldn't export the signing key; check HELM_SIGNING_KEY_PASSPHRASE" >&2
  exit 1
fi
gpg --batch --export "$fingerprint" >"$workdir/pubring.gpg"

helm package "${package_args[@]}" \
  --sign \
  --key "$key_uid" \
  --keyring "$workdir/secring.gpg" \
  --passphrase-file "$passphrase_file"

for archive in "$PACKAGE_DIR"/*.tgz; do
  helm verify "$archive" --keyring "$workdir/pubring.gpg"
done

# The artifacthub.io/signKey annotation is what Artifact Hub shows users as the
# key to verify against. Flag it if it's missing or points at a different key.
annotated_fingerprint="$(awk '
  /artifacthub\.io\/signKey:/ { in_key = 1; next }
  in_key && /fingerprint:/ { print $2; exit }
  in_key && /^  [^ ]/ { exit }
' "$CHART_DIR/Chart.yaml")"
if [[ -z "$annotated_fingerprint" ]]; then
  warn "Chart.yaml has no artifacthub.io/signKey annotation; Artifact Hub won't show which key signed this chart."
elif [[ "$(tr '[:lower:]' '[:upper:]' <<<"$annotated_fingerprint")" != "$fingerprint" ]]; then
  warn "Chart.yaml signKey fingerprint ($annotated_fingerprint) doesn't match the signing key ($fingerprint)."
fi

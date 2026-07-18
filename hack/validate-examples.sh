#!/usr/bin/env bash
# Validate every example manifest against the provider's own generated CRDs.
#
# The examples are the provider's front door: an adopter copy-pastes them into a
# real cluster. kubectl has validated fields strictly since Kubernetes 1.25, so a
# manifest that does not match the generated schema is a hard error for them. This
# gate catches that offline — it needs a throwaway API server, but no UniFi
# controller and no credentials, so it can run on an ordinary pull request.
#
# Usage: hack/validate-examples.sh
# Requires: kubectl, and a KUBECONFIG pointing at a disposable cluster.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CRD_DIR="${REPO_ROOT}/package/crds"
EXAMPLES_DIR="${REPO_ROOT}/examples"

# Manifests deliberately not validated here, each with the reason it cannot be.
# Anything under examples/ that is neither validated nor listed here fails the
# run, so a new example can never slip through unchecked.
#
# Kept as a plain function rather than an associative array so the script also
# runs on the bash 3.2 that ships with macOS.
exclusion_reason() {
  case "$1" in
  examples/install.yaml)
    printf '%s' "pkg.crossplane.io/v1 Provider is owned by Crossplane core, not by this provider's CRDs"
    ;;
  *) printf '' ;;
  esac
}

log() { printf '%s\n' "$*"; }
fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

[[ -d "${CRD_DIR}" ]] || fail "CRD directory not found: ${CRD_DIR}"
[[ -d "${EXAMPLES_DIR}" ]] || fail "examples directory not found: ${EXAMPLES_DIR}"

command -v kubectl >/dev/null 2>&1 || fail "kubectl is required but not installed"
kubectl cluster-info >/dev/null 2>&1 || fail "no reachable cluster; point KUBECONFIG at a disposable one"

# The namespaced examples place their ProviderConfig in crossplane-system, which
# exists on any cluster that has Crossplane installed. A throwaway API server has
# no Crossplane, so create the namespace to match what an adopter actually has.
# This only satisfies a namespace precondition — a schema-invalid manifest is
# still rejected.
log "==> Preparing namespaces"
kubectl create namespace crossplane-system --dry-run=client -o yaml | kubectl apply -f - >/dev/null

log "==> Installing generated CRDs from package/crds"
kubectl apply --server-side -f "${CRD_DIR}" >/dev/null
kubectl wait --for=condition=Established --timeout=120s -f "${CRD_DIR}" >/dev/null
log "    $(find "${CRD_DIR}" -name '*.yaml' | wc -l | tr -d ' ') CRDs established"

# Collect the manifests up front so an enumeration failure surfaces as a failure
# rather than as an empty loop that exits 0.
manifests=()
while IFS= read -r file; do
  manifests+=("${file}")
done < <(cd "${REPO_ROOT}" && find examples -type f -name '*.yaml' | sort)

(( ${#manifests[@]} > 0 )) || fail "no example manifests found under examples/ — enumeration failed"

log "==> Validating ${#manifests[@]} example manifests with --validate=strict"

validated=0
skipped=0
failures=()

for manifest in "${manifests[@]}"; do
  reason="$(exclusion_reason "${manifest}")"
  if [[ -n "${reason}" ]]; then
    log "  SKIP ${manifest} (${reason})"
    skipped=$(( skipped + 1 ))
    continue
  fi

  if output="$(kubectl apply --dry-run=server --validate=strict -f "${REPO_ROOT}/${manifest}" 2>&1)"; then
    log "  PASS ${manifest}"
    validated=$(( validated + 1 ))
  else
    log "  FAIL ${manifest}"
    printf '%s\n' "${output}" | sed 's/^/         /'
    failures+=("${manifest}")
  fi
done

# A run that validated nothing means the gate silently did no work — treat that
# as a failure, never as success.
(( validated > 0 )) || fail "no manifests were validated (${skipped} skipped) — the gate did no work"

log ""
log "==> ${validated} validated, ${skipped} skipped, ${#failures[@]} failed"

if (( ${#failures[@]} > 0 )); then
  printf 'ERROR: %d example manifest(s) rejected by the API server:\n' "${#failures[@]}" >&2
  printf '  - %s\n' "${failures[@]}" >&2
  exit 1
fi

log "All example manifests are accepted by the API server."

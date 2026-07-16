#!/usr/bin/env bash
# Regression tests for version_diff.sh.
#
# These pin the exact stdout the Makefile schema-bump path reports, because the output
# is the deliverable: a maintainer reading a bump PR judges it by these lines. The
# expectations below were captured from the version_diff.py this script replaced, so a
# drift here is a behaviour change, not a formatting nit.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
version_diff="${script_dir}/version_diff.sh"
testdata="${script_dir}/testdata"

failures=0

# assert_output <name> <expected-exit> <expected-stdout> <base-schema>
assert_output() {
  local name="$1" want_exit="$2" want_out="$3" base="$4"
  local got_out got_exit=0

  got_out="$("${version_diff}" \
    "${testdata}/generated.lst" \
    "${base}" \
    "${testdata}/schema.bumped.json" 2>&1)" || got_exit=$?

  if [[ "${got_exit}" != "${want_exit}" ]]; then
    echo "FAIL ${name}: exit ${got_exit}, want ${want_exit}"
    failures=$((failures + 1))
    return
  fi

  if [[ "${got_out}" != "${want_out}" ]]; then
    echo "FAIL ${name}: stdout differs"
    diff <(printf '%s\n' "${want_out}") <(printf '%s\n' "${got_out}") || true
    failures=$((failures + 1))
    return
  fi

  echo "ok   ${name}"
}

# Covers all four cases in one pass, as the real invocation does:
#   unifi_network   version changed 0 -> 1        → reported
#   unifi_site      version changed 2 -> 3        → reported
#   unifi_user      version unchanged             → silent
#   unifi_gone      absent from both schemas      → named as the missing key
#   unifi_noversion present, but has no version   → 'version' named as the missing key
assert_output "reports changed versions, skips unchanged, names missing keys" 0 \
  "Reporting schema changes between \"${testdata}/schema.base.json\" as base version and \"${testdata}/schema.bumped.json\" as bumped version
unifi_network:0-1
unifi_site:2-3
unifi_gone is not found in schema: 'unifi_gone'
unifi_noversion is not found in schema: 'version'" \
  "${testdata}/schema.base.json"

# A base schema with no provider is unusable: fail loudly rather than report an empty
# diff, which would read as "nothing changed" on a bump PR.
assert_output "fails when the base schema has no provider" 255 \
  "Reporting schema changes between \"${testdata}/schema.no-provider.json\" as base version and \"${testdata}/schema.bumped.json\" as bumped version
Cannot extract the provider name from the base schema: ${testdata}/schema.no-provider.json" \
  "${testdata}/schema.no-provider.json"

if [[ "${failures}" -ne 0 ]]; then
  echo "${failures} test(s) failed"
  exit 1
fi

echo "all version_diff tests passed"

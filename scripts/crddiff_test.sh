#!/usr/bin/env bash
# Regression tests for crddiff.sh, the breaking CRD schema gate.
#
# The gate exists to fail a pull request, so most cases here assert that it
# FAILS: on a real breaking change, on a removed CRD, on a tool error, on a base
# it cannot resolve. The breaking and additive cases run the real pinned crddiff
# binary named by $CRDDIFF (`make crddiff.test` builds it), so a pin that stops
# resolving or a tool that stops detecting a removed field fails these tests.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
gate="${script_dir}/crddiff.sh"
testdata="${script_dir}/testdata/crddiff"

if [[ -z "${CRDDIFF:-}" || ! -x "${CRDDIFF}" ]]; then
	echo "FAIL: set CRDDIFF to the pinned crddiff binary (run 'make crddiff.test')"
	exit 1
fi

repo="$(mktemp -d)"
trap 'rm -rf "${repo}"' EXIT

git -C "${repo}" init --quiet
mkdir -p "${repo}/crds"
cp "${testdata}/base.yaml" "${repo}/crds/records.yaml"
cp "${testdata}/base.yaml" "${repo}/crds/removed.yaml"
git -C "${repo}" add crds
git -C "${repo}" -c user.name=test -c user.email=test@example.invalid -c commit.gpgsign=false \
	commit --quiet -m base
git -C "${repo}" tag base

# The two revisions are derived from the one base fixture, so each differs from
# it by exactly the change under test.
# breaking: the base without its optional `ttl` field.
breaking="${repo}/breaking.yaml"
awk '/^ +ttl:$/ {skip=3} skip>0 {skip--; next} {print}' "${testdata}/base.yaml" >"${breaking}"
# additive: the base with a new optional `weight` field after `ttl`.
additive="${repo}/additive.yaml"
awk '{print} /^ +type: number$/ && !done {
	print "                    weight:"
	print "                      description: An optional weight, added by this revision."
	print "                      type: number"
	done=1
}' "${testdata}/base.yaml" >"${additive}"
if grep -q 'ttl:' "${breaking}" || ! grep -q 'weight:' "${additive}"; then
	echo "FAIL: could not derive the revision fixtures from ${testdata}/base.yaml"
	exit 1
fi

# A stand-in for a crddiff that cannot run, such as a pin that no longer resolves.
broken_tool="${repo}/broken-crddiff"
printf '#!/bin/sh\necho "crddiff: failed to load CRDs" >&2\nexit 3\n' >"${broken_tool}"
chmod +x "${broken_tool}"

failures=0

# assert_gate <name> <want-exit> <want-output-substring> <tool> <base> [crd...]
assert_gate() {
	local name="$1" want_exit="$2" want_text="$3" tool="$4" base="$5"
	shift 5
	local got_out got_exit=0

	got_out="$(cd "${repo}" && "${gate}" "${base}" "${tool}" "$@" 2>&1)" || got_exit=$?

	if [[ "${got_exit}" != "${want_exit}" ]]; then
		echo "FAIL ${name}: exit ${got_exit}, want ${want_exit}"
		printf '%s\n' "${got_out}"
		failures=$((failures + 1))
		return
	fi
	if [[ "${got_out}" != *"${want_text}"* ]]; then
		echo "FAIL ${name}: output does not contain '${want_text}'"
		printf '%s\n' "${got_out}"
		failures=$((failures + 1))
		return
	fi
	echo "ok   ${name}"
}

# reset_tree restores every fixture CRD to its base content.
reset_tree() {
	git -C "${repo}" checkout --quiet base -- crds
}

assert_gate "passes when no CRD changed" 0 "No modified CRDs to check." \
	"${CRDDIFF}" base

reset_tree
cp "${additive}" "${repo}/crds/records.yaml"
assert_gate "passes an added optional field" 0 "No breaking CRD schema changes." \
	"${CRDDIFF}" base crds/records.yaml

reset_tree
cp "${breaking}" "${repo}/crds/records.yaml"
assert_gate "fails a removed field and names it" 1 "ttl" \
	"${CRDDIFF}" base crds/records.yaml

reset_tree
rm "${repo}/crds/removed.yaml"
assert_gate "fails a removed CRD" 1 "crds/removed.yaml exists at the base but has been removed" \
	"${CRDDIFF}" base crds/removed.yaml

reset_tree
cp "${testdata}/base.yaml" "${repo}/crds/new.yaml"
assert_gate "passes a CRD the base does not have" 0 "New CRD crds/new.yaml" \
	"${CRDDIFF}" base crds/new.yaml
rm "${repo}/crds/new.yaml"

reset_tree
cp "${additive}" "${repo}/crds/records.yaml"
assert_gate "fails when the tool cannot complete a comparison" 1 "failed to load CRDs" \
	"${broken_tool}" base crds/records.yaml

reset_tree
assert_gate "stops when the base cannot be resolved" 2 "Cannot resolve the base revision 'no-such-base'" \
	"${CRDDIFF}" no-such-base crds/records.yaml

assert_gate "stops when the base is empty" 2 "Cannot resolve the base revision ''" \
	"${CRDDIFF}" "" crds/records.yaml

assert_gate "stops when the tool is missing" 2 "crddiff is not an executable file" \
	"${repo}/no-such-crddiff" base crds/records.yaml

# A path in neither the tree nor the base means the gate is looking at the wrong
# repository or base; reading it as a new CRD would pass every change unchecked.
assert_gate "stops on a CRD in neither the tree nor the base" 2 "is in neither the working tree nor the base revision" \
	"${CRDDIFF}" base crds/missing.yaml

if [[ "${failures}" -ne 0 ]]; then
	echo "${failures} test(s) failed"
	exit 1
fi

echo "all crddiff tests passed"

#!/usr/bin/env bash
# Fail when a change makes a breaking change to a published CRD schema.
#
# Usage: crddiff.sh <base-revision> <crddiff-binary> [<crd-path>...]
#
# Each CRD path is compared against its version at <base-revision> with
# crddiff's revision check. The gate fails closed:
#   - a breaking change fails it;
#   - a CRD that exists at the base but is gone from the working tree fails it,
#     because removing a CRD removes that API;
#   - a comparison the tool could not complete, a base revision that cannot be
#     resolved, or a tool that cannot be run stops it, instead of passing.
# A CRD the base does not have yet is new, so there is nothing for it to break.
#
# CRDDIFF_ACCEPT_BREAKING=true accepts the breaking changes found, after
# printing them, for a change that intends them. It never accepts a check that
# could not run: only a report crddiff recognisably produced counts as a
# breaking change.
#
# Exit codes: 0 no breaking change (or accepted ones), 1 a breaking change,
# 2 the check could not run.

set -euo pipefail

if [[ "$#" -lt 2 ]]; then
	echo "usage: crddiff.sh <base-revision> <crddiff-binary> [<crd-path>...]" >&2
	exit 2
fi

base_revision="$1"
crddiff="$2"
shift 2

if [[ "$#" -eq 0 ]]; then
	echo "No modified CRDs to check."
	exit 0
fi

if [[ -z "${base_revision}" ]] || ! base_commit="$(git rev-parse --verify --quiet "${base_revision}^{commit}")"; then
	echo "Cannot resolve the base revision '${base_revision}'; fetch it before checking CRDs for breaking changes." >&2
	exit 2
fi

if [[ ! -f "${crddiff}" || ! -x "${crddiff}" ]]; then
	echo "crddiff is not an executable file: '${crddiff}'." >&2
	exit 2
fi

work_dir="$(mktemp -d)"
trap 'rm -rf "${work_dir}"' EXIT

failed=0
for crd in "$@"; do
	in_base=1
	git cat-file -e "${base_commit}:${crd}" 2>/dev/null || in_base=0
	if [[ ! -f "${crd}" ]]; then
		if [[ "${in_base}" -eq 0 ]]; then
			# Neither side has it: the path, the directory or the base is wrong,
			# so nothing here can be called safe.
			echo "${crd} is in neither the working tree nor the base revision; run from the repository root with the right base." >&2
			exit 2
		fi
		echo "Breaking change: ${crd} exists at the base but has been removed."
		failed=1
		continue
	fi
	if [[ "${in_base}" -eq 0 ]]; then
		echo "New CRD ${crd}: not present at the base, nothing to compare."
		continue
	fi

	git cat-file -p "${base_commit}:${crd}" >"${work_dir}/base.yaml"
	echo "Checking ${crd} for breaking API changes..."
	if output="$("${crddiff}" revision --enable-upjet-extensions "${work_dir}/base.yaml" "${crd}" 2>&1)"; then
		continue
	fi
	# crddiff reports breaking changes per CRD version, each under a
	# 'Version "<name>":' heading; any other failure is the tool not running.
	if ! grep -q '^Version "' <<<"${output}"; then
		echo "crddiff could not compare ${crd}:" >&2
		printf '%s\n' "${output}" >&2
		exit 2
	fi
	echo "Breaking change detected in ${crd}:"
	printf '%s\n\n' "${output}"
	failed=1
done

if [[ "${failed}" -ne 0 ]]; then
	if [[ "${CRDDIFF_ACCEPT_BREAKING:-}" == "true" ]]; then
		echo "Breaking CRD schema changes found and accepted (CRDDIFF_ACCEPT_BREAKING=true); see the report above."
		exit 0
	fi
	echo "Breaking CRD schema changes found; see the report above." >&2
	exit 1
fi
echo "No breaking CRD schema changes."

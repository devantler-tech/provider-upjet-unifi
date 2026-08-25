#!/bin/sh
set -eu

fail() {
	echo "release-version: $*" >&2
	exit 1
}

[ "$#" -eq 2 ] || fail "expected ref type and ref name"

ref_type=$1
ref_name=$2

[ "$ref_type" = "tag" ] || fail "releases must run from a tag"

numeric='(0|[1-9][0-9]*)'
prerelease="(${numeric}|[0-9]*[A-Za-z-][0-9A-Za-z-]*)"
oci_semver="^v${numeric}\\.${numeric}\\.${numeric}(-${prerelease}(\\.${prerelease})*)?$"

printf '%s\n' "$ref_name" | grep -Eq "$oci_semver" ||
	fail "tag must be an OCI-compatible semantic version prefixed with v"
[ "${#ref_name}" -le 128 ] || fail "tag must fit the OCI 128-character limit"

printf 'version=%s\n' "$ref_name"

#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
resolver="$root/scripts/release-version.sh"

fail() {
	echo "release-version.test: $*" >&2
	exit 1
}

expect_version() {
	ref_type=$1
	ref_name=$2
	want=$3

	got=$($resolver "$ref_type" "$ref_name") ||
		fail "expected $ref_type $ref_name to resolve"
	[ "$got" = "version=$want" ] ||
		fail "expected version=$want for $ref_name, got $got"
}

expect_rejected() {
	ref_type=$1
	ref_name=$2

	if $resolver "$ref_type" "$ref_name" >/dev/null 2>&1; then
		fail "expected $ref_type $ref_name to be rejected"
	fi
}

expect_version tag v0.1.0 v0.1.0
expect_version tag v12.34.56-rc.1 v12.34.56-rc.1

prerelease_121=$(printf '%0121d' 0 | tr '0' a)
prerelease_122=$(printf '%0122d' 0 | tr '0' a)
expect_version tag "v1.2.3-$prerelease_121" "v1.2.3-$prerelease_121"

expect_rejected branch v1.2.3
expect_rejected tag 1.2.3
expect_rejected tag v01.2.3
expect_rejected tag v1.2
expect_rejected tag v1.2.3/other
expect_rejected tag v1.2.3-01
expect_rejected tag v1.2.3-rc.01
expect_rejected tag v12.34.56+build.7
expect_rejected tag v12.34.56-rc.1+build.7
expect_rejected tag "v1.2.3-$prerelease_122"

echo "release-version.test: OCI-compatible tag versions resolve"

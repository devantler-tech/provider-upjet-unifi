#!/bin/sh
# Keep every Go toolchain entry point on the same advisory-free release. This
# catches the configuration split that otherwise fails only after CI starts.
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
minimum=1.26.6

fail() {
	echo "go-floor.test: $*" >&2
	exit 1
}

read_yaml_value() {
	awk -v key="$2" '$1 == key ":" { print $2; exit }' "$1" | tr -d '"'\'' '
}

module_version=$(awk '$1 == "go" { print $2; exit }' "$root/go.mod")
[ -n "$module_version" ] || fail "go.mod has no go directive"

if ! awk -v actual="$module_version" -v minimum="$minimum" 'BEGIN {
	split(actual, a, ".")
	split(minimum, m, ".")
	for (i = 1; i <= 3; i++) {
		if (a[i] + 0 > m[i] + 0) exit 0
		if (a[i] + 0 < m[i] + 0) exit 1
	}
	exit 0
}'; then
	fail "go.mod requires $module_version; minimum safe release is $minimum"
fi

ci_version=$(read_yaml_value "$root/.github/workflows/ci.yml" GO_VERSION)
e2e_version=$(read_yaml_value "$root/.github/workflows/e2e.yaml" GO_VERSION)
publish_version=$(awk '
	$1 == "go-version:" { found = 1; next }
	found && $1 == "default:" { print $2; exit }
' "$root/.github/workflows/publish-provider-package.yml" | tr -d '"'\'' ')

[ "$ci_version" = "$module_version" ] ||
	fail "ci.yml installs $ci_version, but go.mod requires $module_version"
[ "$e2e_version" = "$module_version" ] ||
	fail "e2e.yaml installs $e2e_version, but go.mod requires $module_version"
[ "$publish_version" = "$module_version" ] ||
	fail "publish-provider-package.yml defaults to $publish_version, but go.mod requires $module_version"

echo "go-floor.test: all Go entry points use advisory-free $module_version"

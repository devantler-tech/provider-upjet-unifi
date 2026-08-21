#!/bin/sh
# Hermetic check that go.mod's language floor is not the advisory-laden
# 1.25.12 pin. CI's setup-go + go test remain the compile proof.
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
mod="$root/go.mod"
[ -f "$mod" ] || {
	echo "go-floor.test: missing $mod" >&2
	exit 1
}
line=$(awk '/^go / { print; exit }' "$mod")
[ -n "$line" ] || {
	echo "go-floor.test: no go directive in $mod" >&2
	exit 1
}
case "$line" in
'go 1.25.12')
	echo "go-floor.test: floor still 1.25.12 (OSV GO-2026-5026 family)" >&2
	exit 1
	;;
'go 1.25.15')
	echo "go-floor.test: floor is 1.25.15"
	exit 0
	;;
*)
	echo "go-floor.test: unexpected go directive: $line" >&2
	exit 1
	;;
esac

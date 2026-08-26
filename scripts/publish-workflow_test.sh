#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
workflow="$root/.github/workflows/publish-provider-package.yml"
reference_workflow="$root/.github/workflows/ci.yml"

guard_script=$(mktemp)
trap 'rm -f "$guard_script"' EXIT INT TERM

# GitHub expressions are intentionally literal inside the Ruby program.
# shellcheck disable=SC2016
ruby -r yaml -e '
  workflow = YAML.safe_load(File.read(ARGV.fetch(0)))
  triggers = workflow.fetch("on")
  jobs = workflow.fetch("jobs")

  raise "push must select release tags" unless triggers.dig("push", "tags") == ["v*"]
  raise "manual dispatch must not accept a free-text version" unless triggers.fetch("workflow_dispatch") == {}

  resolve = jobs.fetch("resolve-version")
  raise "resolver must not skip a non-tag ref - a skipped job reports success" if resolve.key?("if")
  raise "resolver must publish a version output" unless resolve.dig("outputs", "version") == "${{ steps.release-version.outputs.version }}"

  steps = resolve.fetch("steps")
  guard_index = steps.index { |step| step["id"] == "reject-non-tag-ref" }
  raise "resolver must explicitly reject a non-tag ref" unless guard_index
  guard = steps.fetch(guard_index)
  raise "rejection must receive github.ref_type" unless guard.dig("env", "REF_TYPE") == "${{ github.ref_type }}"
  raise "rejection must receive github.ref_name" unless guard.dig("env", "REF_NAME") == "${{ github.ref_name }}"

  # The ref must reach the shell only through env, never inline ${{ }}, or ref
  # contents could inject shell syntax (zizmor template-injection). Asserting it
  # is also what makes the extracted guard safe to execute below.
  guard_run = guard.fetch("run")
  raise "rejection must not interpolate ${{ }} into the shell" if guard_run.include?("${{")
  File.write(ARGV.fetch(2), guard_run)

  checkout_index = steps.index { |step| step.fetch("uses", "").start_with?("actions/checkout@") }
  raise "resolver checkout is missing" unless checkout_index
  raise "non-tag rejection must run before checkout" unless guard_index < checkout_index
  checkout = steps.fetch(checkout_index)
  known_checkout_pins = File.read(ARGV.fetch(1)).scan(%r{actions/checkout@[0-9a-f]{40}}).uniq
  raise "resolver checkout pin must match a repository-validated pin" unless known_checkout_pins.include?(checkout.fetch("uses"))

  resolve_step = resolve.fetch("steps").find { |step| step["id"] == "release-version" }
  raise "resolver step is missing" unless resolve_step
  raise "resolver must receive github.ref_type" unless resolve_step.dig("env", "REF_TYPE") == "${{ github.ref_type }}"
  raise "resolver must receive github.ref_name" unless resolve_step.dig("env", "REF_NAME") == "${{ github.ref_name }}"
  raise "resolver must execute the tested helper" unless resolve_step.fetch("run").include?("./scripts/release-version.sh")

  publish = jobs.fetch("publish-provider-package")
  raise "publish must wait for validated version" unless publish.fetch("needs") == "resolve-version"
  raise "publish must consume the validated version" unless publish.dig("with", "version") == "${{ needs.resolve-version.outputs.version }}"

  sign = jobs.fetch("sign-provider-package")
  raise "sign must wait for validation and publication" unless sign.fetch("needs") == ["resolve-version", "publish-provider-package"]
  version = sign.fetch("steps").find { |step| step["name"] == "Sign the package image (keyless)" }.dig("env", "VERSION")
  raise "sign must consume the validated version" unless version == "${{ needs.resolve-version.outputs.version }}"
' "$workflow" "$reference_workflow" "$guard_script"

# Run the guard the workflow actually ships, rather than matching text in it: a
# guard that always exits, tests the wrong value, or drops its guidance reads
# identically to a correct one under a substring assertion.

# A tag ref is what a release runs from, so the guard must let it through.
if ! REF_TYPE=tag REF_NAME=v1.2.3 sh "$guard_script" >/dev/null 2>&1; then
	echo "publish-workflow.test: the guard rejected a tag ref, so no release can publish" >&2
	exit 1
fi

# A branch ref is the ref the Actions dispatch UI offers by default, and it must
# fail loudly rather than skip.
if guard_output=$(REF_TYPE=branch REF_NAME=main sh "$guard_script" 2>&1); then
	echo "publish-workflow.test: the guard admitted a non-tag ref instead of failing" >&2
	exit 1
fi

# A rejection an operator cannot act on sends them back to the workflow source.
case "$guard_output" in
*"v* tag"*) ;;
*)
	echo "publish-workflow.test: the rejection omits the v* rerun/push guidance" >&2
	exit 1
	;;
esac
case "$guard_output" in
*main*) ;;
*)
	echo "publish-workflow.test: the rejection does not name the ref it refused" >&2
	exit 1
	;;
esac

echo "publish-workflow.test: tag-derived version gates publish and sign"

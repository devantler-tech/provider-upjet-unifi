#!/bin/sh
set -eu

root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
workflow="$root/.github/workflows/publish-provider-package.yml"

# GitHub expressions are intentionally literal inside the Ruby program.
# shellcheck disable=SC2016
ruby -ryaml -e '
  workflow = YAML.safe_load(File.read(ARGV.fetch(0)))
  triggers = workflow.fetch("on")
  jobs = workflow.fetch("jobs")

  raise "push must select release tags" unless triggers.dig("push", "tags") == ["v*"]
  raise "manual dispatch must not accept a free-text version" unless triggers.fetch("workflow_dispatch") == {}

  resolve = jobs.fetch("resolve-version")
  raise "resolver must publish a version output" unless resolve.dig("outputs", "version") == "${{ steps.release-version.outputs.version }}"

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
' "$workflow"

echo "publish-workflow.test: tag-derived version gates publish and sign"

#!/usr/bin/env bash
# Reports Terraform provider native state schema `version` changes between a base
# and a bumped schema, for the resources this provider generates.
#
# usage: version_diff.sh <generated resource list> <base JSON schema path> <bumped JSON schema path>
# example: version_diff.sh config/generated.lst .work/schema.json.3.38.0 config/schema.json
#
# Replaces the former version_diff.py: the portfolio scripting stack is bash or Go,
# never Python. Output is unchanged from that script, including how missing keys are
# reported, so the Makefile schema-bump path reads exactly as before.

set -euo pipefail

resources_path="${1:?usage: version_diff.sh <generated resource list> <base schema> <bumped schema>}"
base_path="${2:?usage: version_diff.sh <generated resource list> <base schema> <bumped schema>}"
bumped_path="${3:?usage: version_diff.sh <generated resource list> <base schema> <bumped schema>}"

echo "Reporting schema changes between \"${base_path}\" as base version and \"${bumped_path}\" as bumped version"

# The provider name is the first key of .provider_schemas — keys_unsorted keeps the
# document order the Python dict iteration relied on, where `keys` would sort it.
provider_name="$(jq -r '.provider_schemas | keys_unsorted[0] // empty' "${base_path}")"
if [[ -z "${provider_name}" ]]; then
  echo "Cannot extract the provider name from the base schema: ${base_path}"
  # 255 mirrors the exit status of the replaced script's sys.exit(-1).
  exit 255
fi

# One jq pass over the generated resource list. ' is a literal single quote —
# written as an escape so it cannot terminate this single-quoted jq program. Python
# printed a caught KeyError, which renders as 'the-missing-key', and the Makefile
# output is expected to read the same. Each branch mirrors one lookup the Python did,
# in the same order, so the same missing key is named.
jq -r \
  --slurpfile base "${base_path}" \
  --slurpfile bumped "${bumped_path}" \
  --arg provider "${provider_name}" \
  '
    ($base[0].provider_schemas[$provider].resource_schemas // {}) as $b
    | ($bumped[0].provider_schemas[$provider].resource_schemas // {}) as $u
    | .[]
    | . as $name
    | if ($b | has($name) | not) then
        "\($name) is not found in schema: \u0027\($name)\u0027"
      elif ($b[$name] | has("version") | not) then
        "\($name) is not found in schema: \u0027version\u0027"
      elif ($u | has($name) | not) then
        "\($name) is not found in schema: \u0027\($name)\u0027"
      elif ($u[$name] | has("version") | not) then
        "\($name) is not found in schema: \u0027version\u0027"
      elif ($b[$name].version != $u[$name].version) then
        "\($name):\($b[$name].version)-\($u[$name].version)"
      else
        empty
      end
  ' "${resources_path}"

// Package config also rewrites the descriptions of write-only attributes, which
// this provider cannot set, before upjet generates the API from them.
package config

import (
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// writeOnlySuffix is the upstream provider's naming convention for a
// write-only attribute: <name>_wo is the write-only form of <name>.
const writeOnlySuffix = "_wo"

// WriteOnlyDescriptions rewrites the description of every write-only attribute.
//
// Upstream documents each one as "Write-only equivalent of `<name>` (Terraform
// 1.11+). ...". Upjet removes every sentence that mentions Terraform from a
// generated description and splits sentences at each '.', so the version number
// cut that first sentence and the description started mid-sentence. The
// provider runs a Terraform CLI older than 1.11, which rejects any value for a
// write-only attribute, so the description says that and names the attribute
// to set instead, in words that survive upjet's filter.
func WriteOnlyDescriptions() config.ResourceOption {
	return func(r *config.Resource) {
		if r.TerraformResource != nil {
			describeWriteOnly(r.TerraformResource.Schema)
		}
	}
}

func describeWriteOnly(attributes map[string]*schema.Schema) {
	for name, attribute := range attributes {
		if nested, ok := attribute.Elem.(*schema.Resource); ok {
			describeWriteOnly(nested.Schema)
		}
		base, ok := writeOnlyBase(name, attributes)
		if !ok {
			continue
		}
		attribute.Description = writeOnlyDescription(base)
	}
}

// writeOnlyBase returns the attribute a write-only attribute stands in for, when
// name follows the <name>_wo convention and its <name> sibling exists.
func writeOnlyBase(name string, siblings map[string]*schema.Schema) (string, bool) {
	base, ok := strings.CutSuffix(name, writeOnlySuffix)
	if !ok || base == "" {
		return "", false
	}
	if _, ok := siblings[base]; !ok {
		return "", false
	}
	return base, true
}

func writeOnlyDescription(base string) string {
	return "Write-only equivalent of `" + base + "`. " +
		"This provider cannot set write-only attributes: the engine it bundles predates them " +
		"and rejects any value, so setting this field fails validation. " +
		"Set `" + base + "` instead."
}

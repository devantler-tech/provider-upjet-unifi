// Package config also rewrites the descriptions of write-only attributes, which
// this provider cannot set, before upjet generates the API from them.
package config

import (
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/crossplane/upjet/v2/pkg/types/name"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// writeOnlySuffix is the upstream provider's naming convention for a
// write-only attribute: <name>_wo is the write-only form of <name>.
const writeOnlySuffix = "_wo"

// writeOnlyVersionSuffix names the version counter upstream pairs with some
// write-only attributes: <name>_wo_version tells it when <name>_wo changed.
const writeOnlyVersionSuffix = writeOnlySuffix + "_version"

// secretRefSuffix is what upjet appends to the API field of a sensitive
// attribute, which it reads from a Kubernetes Secret instead of the manifest.
const secretRefSuffix = "SecretRef"

// writeOnlyMinimumEngine is the first engine version that accepts write-only
// attributes. Upstream names it as "Terraform 1.11+", a sentence upjet drops.
const writeOnlyMinimumEngine = "1.11"

// WriteOnlyDescriptions rewrites the description of every write-only attribute
// and of the attribute it stands in for.
//
// Upstream documents each one as "Write-only equivalent of `<name>` (Terraform
// 1.11+). ...". Upjet removes every sentence that mentions Terraform from a
// generated description and splits sentences at each '.', so the version number
// cut that first sentence and the description started mid-sentence. The
// provider runs a Terraform CLI older than 1.11, which rejects any value for a
// write-only attribute, so the description says that and names the API field to
// set instead, in words that survive upjet's filter. Upstream also tells users
// of <name> to use <name>_wo instead; that advice leads to the same validation
// failure, so it is removed from the description of <name>.
//
// A <name>_wo_version counter only tells the engine that <name>_wo changed, so
// its description says it has no effect here and names the same API field.
func WriteOnlyDescriptions() config.ResourceOption {
	return func(r *config.Resource) {
		if r.TerraformResource != nil {
			describeWriteOnly(r.TerraformResource.Schema)
		}
	}
}

func describeWriteOnly(attributes map[string]*schema.Schema) {
	for attributeName, attribute := range attributes {
		if nested, ok := attribute.Elem.(*schema.Resource); ok {
			describeWriteOnly(nested.Schema)
		}
		if base, ok := writeOnlyVersionBase(attributeName, attributes); ok {
			attribute.Description = writeOnlyVersionDescription(base, apiFieldName(base, attributes[base]))
			continue
		}
		base, ok := writeOnlyBase(attributeName, attributes)
		if !ok {
			continue
		}
		attribute.Description = writeOnlyDescription(base, apiFieldName(base, attributes[base]))
		attributes[base].Description = withoutSentencesNaming(attributes[base].Description, attributeName)
	}
}

// writeOnlyBase returns the attribute a write-only attribute stands in for, when
// name follows the <name>_wo convention and its <name> sibling exists.
func writeOnlyBase(attributeName string, siblings map[string]*schema.Schema) (string, bool) {
	base, ok := strings.CutSuffix(attributeName, writeOnlySuffix)
	if !ok || base == "" {
		return "", false
	}
	if _, ok := siblings[base]; !ok {
		return "", false
	}
	return base, true
}

// writeOnlyVersionBase returns the attribute whose write-only form a version
// counter belongs to, when name follows the <name>_wo_version convention and
// both its <name>_wo and <name> siblings exist.
func writeOnlyVersionBase(attributeName string, siblings map[string]*schema.Schema) (string, bool) {
	base, ok := strings.CutSuffix(attributeName, writeOnlyVersionSuffix)
	if !ok || base == "" {
		return "", false
	}
	if _, ok := siblings[base+writeOnlySuffix]; !ok {
		return "", false
	}
	if _, ok := siblings[base]; !ok {
		return "", false
	}
	return base, true
}

// apiFieldName returns the field a managed-resource manifest sets for the
// attribute: its lower camel case name, with upjet's SecretRef suffix when the
// attribute is sensitive.
func apiFieldName(attributeName string, attribute *schema.Schema) string {
	field := name.NewFromSnake(attributeName).LowerCamelComputed
	if attribute.Sensitive {
		field += secretRefSuffix
	}
	return field
}

func writeOnlyDescription(base, field string) string {
	return "Write-only equivalent of `" + base + "`. " +
		"Write-only attributes need engine version " + writeOnlyMinimumEngine + " or later. " +
		"This provider cannot set them: the engine it bundles predates that version " +
		"and rejects any value, so setting this field fails validation. " +
		"Set `" + field + "` instead."
}

// writeOnlyVersionDescription replaces upstream's "Increment this value to
// trigger an update": the counter only tells the engine that the write-only
// value changed, and this provider cannot set that value.
func writeOnlyVersionDescription(base, field string) string {
	return "Version counter for `" + base + writeOnlySuffix + "`, which this provider cannot set, " +
		"so changing this field has no effect. " +
		"Set `" + field + "` instead."
}

// withoutSentencesNaming drops every sentence of description that names
// attributeName in backticks. A clause before an em dash that does not name
// the attribute, such as a state-persistence warning, is kept as its own
// sentence.
func withoutSentencesNaming(description, attributeName string) string {
	quoted := "`" + attributeName + "`"
	if !strings.Contains(description, quoted) {
		return description
	}
	var kept []string
	for _, sentence := range strings.SplitAfter(description, ". ") {
		if !strings.Contains(sentence, quoted) {
			kept = append(kept, sentence)
			continue
		}
		if clause, _, found := strings.Cut(sentence, " — "); found && !strings.Contains(clause, quoted) {
			kept = append(kept, strings.TrimSpace(clause)+". ")
		}
	}
	return strings.TrimSpace(strings.Join(kept, ""))
}

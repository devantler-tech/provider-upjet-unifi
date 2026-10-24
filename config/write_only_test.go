package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/crossplane/upjet/v2/pkg"
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/mod/semver"
)

// walkAttributes calls visit for every attribute of every configured resource,
// nested blocks included, with its sibling attributes.
func walkAttributes(t *testing.T, visit func(path, name string, attribute *schema.Schema, siblings map[string]*schema.Schema)) {
	t.Helper()
	providers := map[string]*ujconfig.Provider{
		"cluster":    GetProvider(),
		"namespaced": GetProviderNamespaced(),
	}
	for scope, p := range providers {
		if len(p.Resources) == 0 {
			t.Fatalf("%s provider configures no resources", scope)
		}
		for resourceName, r := range p.Resources {
			walkSchema(scope+"/"+resourceName, r.TerraformResource.Schema, visit)
		}
	}
}

func walkSchema(path string, attributes map[string]*schema.Schema, visit func(path, name string, attribute *schema.Schema, siblings map[string]*schema.Schema)) {
	for name, attribute := range attributes {
		visit(path, name, attribute, attributes)
		if nested, ok := attribute.Elem.(*schema.Resource); ok {
			walkSchema(path+"."+name, nested.Schema, visit)
		}
	}
}

// Upjet's description filter deletes every sentence that mentions Terraform,
// and its sentence split also breaks at version numbers, so any description it
// would change loses text in the generated types and CRDs.
func TestNoGeneratedDescriptionLosesText(t *testing.T) {
	walkAttributes(t, func(path, name string, attribute *schema.Schema, _ map[string]*schema.Schema) {
		description := attribute.Description
		if description == "" {
			return
		}
		if filtered := pkg.FilterDescription(description, pkg.TerraformKeyword); filtered != description {
			t.Errorf("%s.%s: upjet would cut this description to %q; rewrite it in config so no sentence mentions Terraform:\n%s",
				path, name, filtered, description)
		}
	})
}

func TestWriteOnlyAttributesKeepTheirFirstSentence(t *testing.T) {
	found := 0
	walkAttributes(t, func(path, name string, attribute *schema.Schema, siblings map[string]*schema.Schema) {
		base, ok := writeOnlyBase(name, siblings)
		if !ok {
			return
		}
		found++
		want := "Write-only equivalent of `" + base + "`. "
		if !strings.HasPrefix(attribute.Description, want) {
			t.Errorf("%s.%s: description does not start with %q: %q", path, name, want, attribute.Description)
		}
		if note := "need engine version " + writeOnlyMinimumEngine + " or later."; !strings.Contains(attribute.Description, note) {
			t.Errorf("%s.%s: description lost the version note %q: %q", path, name, note, attribute.Description)
		}
		field := apiFieldName(base, siblings[base])
		if !strings.HasSuffix(attribute.Description, "Set `"+field+"` instead.") {
			t.Errorf("%s.%s: description does not name %q as the field to set: %q", path, name, field, attribute.Description)
		}
		if strings.Contains(siblings[base].Description, "`"+name+"`") {
			t.Errorf("%s.%s: description still points users at %q, which fails validation: %q", path, base, name, siblings[base].Description)
		}
	})
	if found == 0 {
		t.Fatal("found no write-only attributes; the <name>_wo convention or the schema changed, so this test checks nothing")
	}
}

// engineVersionAssignment matches the Makefile line that sets TERRAFORM_VERSION,
// the engine version the image build passes to the provider's Dockerfile.
var engineVersionAssignment = regexp.MustCompile(`(?m)^[ \t]*(?:export[ \t]+)?TERRAFORM_VERSION[ \t]*[?:]?=[ \t]*(\S+)[ \t]*$`)

// bundledEngineVersion returns the engine version the provider image bundles,
// read from the repository Makefile.
func bundledEngineVersion(t *testing.T) string {
	t.Helper()
	makefile, err := os.ReadFile(filepath.Join("..", "Makefile"))
	if err != nil {
		t.Fatalf("read the Makefile: %v", err)
	}
	assignments := engineVersionAssignment.FindAllSubmatch(makefile, -1)
	if len(assignments) != 1 {
		t.Fatalf("the Makefile sets TERRAFORM_VERSION %d times, want exactly once; point this test at the bundled engine version", len(assignments))
	}
	dockerfile, err := os.ReadFile(filepath.Join("..", "cluster", "images", "provider-upjet-unifi", "Dockerfile"))
	if err != nil {
		t.Fatalf("read the provider Dockerfile: %v", err)
	}
	if engine := "releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/"; !strings.Contains(string(dockerfile), engine) {
		t.Fatalf("the provider Dockerfile no longer bundles the engine from %q, so TERRAFORM_VERSION may not be the bundled engine version", engine)
	}
	return string(assignments[0][1])
}

// The write-only descriptions say the bundled engine predates the first version
// that accepts write-only attributes. Once the image bundles an engine that
// accepts them, that sentence is false in every generated description.
func TestBundledEnginePredatesWriteOnlySupport(t *testing.T) {
	if claim := "the engine it bundles predates that version"; !strings.Contains(writeOnlyDescription("passphrase", "passphraseSecretRef"), claim) {
		t.Fatalf("the write-only description no longer says %q; update this test to check what it now claims", claim)
	}
	bundled := "v" + bundledEngineVersion(t)
	minimum := "v" + writeOnlyMinimumEngine
	for _, version := range []string{bundled, minimum} {
		if !semver.IsValid(version) {
			t.Fatalf("%q is not a version this test can compare", version)
		}
	}
	if semver.Compare(bundled, minimum) >= 0 {
		t.Errorf("the image bundles engine %s, which accepts write-only attributes (%s or later), but the write-only descriptions say it predates that version; rewrite writeOnlyDescription and run make generate",
			bundled, minimum)
	}
}

func TestWriteOnlyBase(t *testing.T) {
	siblings := map[string]*schema.Schema{
		"passphrase":    {},
		"passphrase_wo": {},
		"orphan_wo":     {},
		"_wo":           {},
	}
	cases := map[string]struct {
		base string
		ok   bool
	}{
		"passphrase_wo": {base: "passphrase", ok: true},
		"passphrase":    {},
		"orphan_wo":     {},
		"_wo":           {},
	}
	for name, want := range cases {
		base, ok := writeOnlyBase(name, siblings)
		if base != want.base || ok != want.ok {
			t.Errorf("writeOnlyBase(%q) = (%q, %v), want (%q, %v)", name, base, ok, want.base, want.ok)
		}
	}
}

func TestAPIFieldName(t *testing.T) {
	cases := map[string]struct {
		attribute *schema.Schema
		want      string
	}{
		"pre_shared_key": {attribute: &schema.Schema{Sensitive: true}, want: "preSharedKeySecretRef"},
		"passphrase":     {attribute: &schema.Schema{Sensitive: true}, want: "passphraseSecretRef"},
		"network_id":     {attribute: &schema.Schema{}, want: "networkId"},
	}
	for attributeName, c := range cases {
		if got := apiFieldName(attributeName, c.attribute); got != c.want {
			t.Errorf("apiFieldName(%q) = %q, want %q", attributeName, got, c.want)
		}
	}
}

func TestWithoutSentencesNaming(t *testing.T) {
	cases := map[string]string{
		"The passphrase for the network, only required if `security` is not `open`. Stored in state — use `passphrase_wo` to avoid persisting the secret.": "The passphrase for the network, only required if `security` is not `open`. Stored in state.",
		"The IPsec pre-shared key. Stored in state — use `passphrase_wo` to avoid persisting the secret.":                                                  "The IPsec pre-shared key. Stored in state.",
		"Use `passphrase_wo` instead.":                    "",
		"No mention of the write-only attribute here.":    "No mention of the write-only attribute here.",
		"Mentions passphrase_wo without backticks. Kept.": "Mentions passphrase_wo without backticks. Kept.",
	}
	for description, want := range cases {
		if got := withoutSentencesNaming(description, "passphrase_wo"); got != want {
			t.Errorf("withoutSentencesNaming(%q) = %q, want %q", description, got, want)
		}
	}
}

// Upstream tells users to increment a write-only attribute's version counter
// to push a new value. This provider cannot set that value, so the counter's
// description says it has no effect and names the field to set instead.
func TestWriteOnlyVersionCountersSayTheyHaveNoEffect(t *testing.T) {
	found := 0
	walkAttributes(t, func(path, name string, attribute *schema.Schema, siblings map[string]*schema.Schema) {
		base, ok := writeOnlyVersionBase(name, siblings)
		if !ok {
			return
		}
		found++
		if !strings.Contains(attribute.Description, "has no effect") {
			t.Errorf("%s.%s: description does not say the counter has no effect: %q", path, name, attribute.Description)
		}
		field := apiFieldName(base, siblings[base])
		if !strings.HasSuffix(attribute.Description, "Set `"+field+"` instead.") {
			t.Errorf("%s.%s: description does not name %q as the field to set: %q", path, name, field, attribute.Description)
		}
	})
	if found == 0 {
		t.Fatal("found no write-only version counters; the <name>_wo_version convention or the schema changed, so this test checks nothing")
	}
}

// No description may send a user to a write-only attribute or its counter,
// except the rewritten descriptions of those attributes themselves.
func TestNoDescriptionInvitesSettingAWriteOnlyAttribute(t *testing.T) {
	walkAttributes(t, func(path, name string, attribute *schema.Schema, siblings map[string]*schema.Schema) {
		if _, ok := writeOnlyBase(name, siblings); ok {
			return
		}
		if _, ok := writeOnlyVersionBase(name, siblings); ok {
			return
		}
		if strings.Contains(attribute.Description, writeOnlySuffix+"`") || strings.Contains(attribute.Description, writeOnlyVersionSuffix+"`") {
			t.Errorf("%s.%s: description points users at a write-only attribute, which this provider cannot set: %q", path, name, attribute.Description)
		}
		if strings.Contains(attribute.Description, "Increment this value") {
			t.Errorf("%s.%s: description asks users to increment a counter that has no effect here: %q", path, name, attribute.Description)
		}
	})
}

// An attribute upstream made optional because its write-only form can stand in
// stays required, in both API scopes. The entry must still match the schema:
// the attribute exists and still has a write-only sibling.
func TestAttributesStayRequiredWithoutWriteOnly(t *testing.T) {
	providers := map[string]*ujconfig.Provider{
		"cluster":    GetProvider(),
		"namespaced": GetProviderNamespaced(),
	}
	for scope, p := range providers {
		for resourceName, paths := range requiredWithoutWriteOnly {
			r, ok := p.Resources[resourceName]
			if !ok {
				t.Errorf("%s: %s is not a generated resource", scope, resourceName)
				continue
			}
			for _, path := range paths {
				attribute := attributeAt(r.TerraformResource.Schema, path)
				if attribute == nil {
					t.Errorf("%s: %s has no attribute %q; remove or correct its requiredWithoutWriteOnly entry", scope, resourceName, path)
					continue
				}
				if !attribute.Required || attribute.Optional {
					t.Errorf("%s: %s.%s is not required (required=%v, optional=%v)", scope, resourceName, path, attribute.Required, attribute.Optional)
				}
				if attributeAt(r.TerraformResource.Schema, path+writeOnlySuffix) == nil {
					t.Errorf("%s: %s.%s has no write-only sibling any more; its requiredWithoutWriteOnly entry may be stale", scope, resourceName, path)
				}
			}
		}
	}
}

func TestWriteOnlyVersionBase(t *testing.T) {
	siblings := map[string]*schema.Schema{
		"private_key":            {},
		"private_key_wo":         {},
		"private_key_wo_version": {},
		"orphan_wo_version":      {},
		"lonely_wo":              {},
		"lonely_wo_version":      {},
		"_wo_version":            {},
	}
	cases := map[string]struct {
		base string
		ok   bool
	}{
		"private_key_wo_version": {base: "private_key", ok: true},
		"private_key_wo":         {},
		"orphan_wo_version":      {},
		"lonely_wo_version":      {},
		"_wo_version":            {},
	}
	for name, want := range cases {
		base, ok := writeOnlyVersionBase(name, siblings)
		if base != want.base || ok != want.ok {
			t.Errorf("writeOnlyVersionBase(%q) = (%q, %v), want (%q, %v)", name, base, ok, want.base, want.ok)
		}
	}
}

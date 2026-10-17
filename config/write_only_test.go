package config

import (
	"strings"
	"testing"

	"github.com/crossplane/upjet/v2/pkg"
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		if !strings.HasSuffix(attribute.Description, "Set `"+base+"` instead.") {
			t.Errorf("%s.%s: description does not name %q as the attribute to set: %q", path, name, base, attribute.Description)
		}
	})
	if found == 0 {
		t.Fatal("found no write-only attributes; the <name>_wo convention or the schema changed, so this test checks nothing")
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

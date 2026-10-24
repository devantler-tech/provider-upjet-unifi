package config

import (
	"strings"
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Reference targets and the field name several resources share.
const (
	targetNetwork  = "unifi_network"
	fieldNetworkID = "network_id"
)

// wantReferences lists every cross-resource reference the provider generates:
// resource, then field path (nested fields joined with "."), then the resource
// the field references. Removing a relationship from config/unifi fails
// TestReferencesAreGenerated, so it cannot disappear from the API unnoticed.
var wantReferences = map[string]map[string]string{
	"unifi_account": {
		fieldNetworkID: targetNetwork,
	},
	"unifi_client": {
		fieldNetworkID: targetNetwork,
	},
	"unifi_device": {
		"mgmt_network_id": targetNetwork,
	},
	"unifi_firewall_policy": {
		"source.network_ids":      targetNetwork,
		"source.zone_id":          "unifi_firewall_zone",
		"destination.network_ids": targetNetwork,
		"destination.zone_id":     "unifi_firewall_zone",
	},
	"unifi_firewall_rule": {
		"src_firewall_group_ids": "unifi_firewall_group",
		"dst_firewall_group_ids": "unifi_firewall_group",
		"src_network_id":         targetNetwork,
		"dst_network_id":         targetNetwork,
	},
	"unifi_firewall_zone": {
		"network_ids": targetNetwork,
	},
	"unifi_port_profile": {
		"native_networkconf_id":            targetNetwork,
		"voice_networkconf_id":             targetNetwork,
		"tagged_networkconf_ids":           targetNetwork,
		"excluded_networkconf_ids":         targetNetwork,
		"multicast_router_networkconf_ids": targetNetwork,
	},
	"unifi_radius_user": {
		fieldNetworkID: targetNetwork,
	},
	"unifi_setting": {
		"igmp_snooping.network_ids": targetNetwork,
		"ips.honeypot.network_id":   targetNetwork,
	},
	"unifi_traffic_route": {
		"network_id": "unifi_vpn_client",
	},
	"unifi_vpn_server": {
		"radiusprofile_id": "unifi_radius_profile",
	},
	"unifi_wlan": {
		"ap_group_ids":                      "unifi_ap_group",
		fieldNetworkID:                      targetNetwork,
		"radius_profile_id":                 "unifi_radius_profile",
		"private_preshared_keys.network_id": targetNetwork,
	},
}

// TestReferencesAreGenerated checks, in both API scopes, that every expected
// reference is configured, that its field path exists in the resource's
// schema (a mistyped path would generate nothing), that the field and its
// target are generated resources, and that no reference exists that this test
// does not list.
func TestReferencesAreGenerated(t *testing.T) {
	providers := map[string]*ujconfig.Provider{
		"cluster":    GetProvider(),
		"namespaced": GetProviderNamespaced(),
	}
	for scope, p := range providers {
		for resourceName, fields := range wantReferences {
			r, ok := p.Resources[resourceName]
			if !ok {
				t.Errorf("%s: %s is not a generated resource", scope, resourceName)
				continue
			}
			for field, target := range fields {
				ref, ok := r.References[field]
				if !ok {
					t.Errorf("%s: %s.%s has no reference to %s", scope, resourceName, field, target)
					continue
				}
				if ref.TerraformName != target {
					t.Errorf("%s: %s.%s references %s, want %s", scope, resourceName, field, ref.TerraformName, target)
				}
				if _, ok := p.Resources[target]; !ok {
					t.Errorf("%s: %s.%s references %s, which is not a generated resource", scope, resourceName, field, target)
				}
				if attribute := schemaAttribute(r.TerraformResource.Schema, field); attribute == nil {
					t.Errorf("%s: %s has no attribute at %q, so the reference generates nothing", scope, resourceName, field)
				} else if attribute.Type != schema.TypeString && !isStringCollection(attribute) {
					t.Errorf("%s: %s.%s is not a string id or a collection of them", scope, resourceName, field)
				}
			}
		}
		for resourceName, r := range p.Resources {
			for field, ref := range r.References {
				if _, ok := wantReferences[resourceName][field]; !ok {
					t.Errorf("%s: %s.%s references %s, but wantReferences does not list it; add it there", scope, resourceName, field, ref.TerraformName)
				}
			}
		}
	}
}

// schemaAttribute returns the attribute at a "."-joined path, descending into
// nested blocks, or nil when the path does not exist.
func schemaAttribute(attributes map[string]*schema.Schema, path string) *schema.Schema {
	head, rest, nested := strings.Cut(path, ".")
	attribute, ok := attributes[head]
	if !ok {
		return nil
	}
	if !nested {
		return attribute
	}
	block, ok := attribute.Elem.(*schema.Resource)
	if !ok {
		return nil
	}
	return schemaAttribute(block.Schema, rest)
}

func isStringCollection(attribute *schema.Schema) bool {
	if attribute.Type != schema.TypeList && attribute.Type != schema.TypeSet {
		return false
	}
	element, ok := attribute.Elem.(*schema.Schema)
	return ok && element.Type == schema.TypeString
}

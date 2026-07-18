// Package unifi holds resource configuration overrides (group, kind) for the
// ubiquiti-community/unifi Terraform provider resources.
package unifi

import (
	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

// shortGroups maps each Terraform resource name to the ShortGroup (API group
// prefix) it should be exposed under. Resources are bucketed into coherent API
// groups so the generated CRDs land under <group>.unifi.crossplane.io.
var shortGroups = map[string]string{
	// account
	"unifi_account": "account",

	// dns
	"unifi_dns_record": "dns",

	// device
	"unifi_device":           "device",
	"unifi_client":           "device",
	"unifi_client_qos_rate":  "device",
	"unifi_power_supervisor": "device",

	// firewall
	resFirewallGroup:        "firewall",
	"unifi_firewall_policy": "firewall",
	"unifi_firewall_rule":   "firewall",
	resFirewallZone:         "firewall",

	// network
	resNetwork:  "network",
	"unifi_wan": "network",
	"unifi_bgp": "network",

	// port
	"unifi_port_forward": "port",
	"unifi_port_profile": "port",

	// radius
	"unifi_radius_profile": "radius",
	"unifi_radius_user":    "radius",

	// route
	"unifi_static_route":  "route",
	"unifi_traffic_route": "route",

	// setting
	"unifi_setting":     "setting",
	"unifi_dynamic_dns": "setting",

	// site
	"unifi_site": "site",

	// vpn
	"unifi_site_to_site_vpn": "vpn",
	resVPNClient:             "vpn",
	"unifi_vpn_server":       "vpn",
	"unifi_wireguard_peer":   "vpn",

	// wlan
	"unifi_wlan":     "wlan",
	"unifi_ap_group": "wlan",
}

// kindOverrides pins the Kind for resources whose upjet-derived Kind would
// otherwise collide within the same group. unifi_static_route and
// unifi_traffic_route both derive to Kind "Route" in group "route", which would
// silently drop one of them, so they are disambiguated here.
var kindOverrides = map[string]string{
	"unifi_static_route":  "StaticRoute",
	"unifi_traffic_route": "TrafficRoute",
	// unifi_ap_group would otherwise derive to the meaninglessly generic Kind
	// "Group" in the wlan group.
	"unifi_ap_group": "ApGroup",
}

// Terraform resource names used as cross-resource reference targets, pulled out
// as constants because several are referenced from multiple fields (and also
// appear as shortGroups keys), which the goconst linter flags as repeated
// string literals.
const (
	resVPNClient     = "unifi_vpn_client"
	resFirewallGroup = "unifi_firewall_group"
	resFirewallZone  = "unifi_firewall_zone"
	resNetwork       = "unifi_network"
)

// references declares Upjet cross-resource references: for each Terraform
// resource it maps a Terraform field name to the resource that field should be
// able to reference. The generator then emits <field>Ref/<field>Selector
// alongside the raw field, and (by default) resolves the value from the
// referenced managed resource's external name.
//
// unifi_traffic_route.network_id can be hard to wire declaratively because a VPN
// client's UniFi network id is only known after the unifi_vpn_client managed
// resource first reconciles. Referencing the unifi_vpn_client lets a consumer
// point a TrafficRoute at a Client by name (networkIdRef/networkIdSelector)
// instead of hard-coding a post-create id. A plain unifi_network id can still be
// supplied directly via the raw networkId field.
var references = map[string]ujconfig.References{
	"unifi_traffic_route": {
		"network_id": {
			TerraformName: resVPNClient,
		},
	},

	// A firewall rule points at firewall groups and networks by their UniFi ids,
	// which are only known after those objects reconcile. Referencing them lets a
	// consumer wire a rule to a FirewallGroup/Network by name via the generated
	// *Ref/*Selector companions; the raw id fields stay settable directly. The
	// *_firewall_group_ids/*network* list fields resolve as slice references.
	"unifi_firewall_rule": {
		"src_firewall_group_ids": {
			TerraformName: resFirewallGroup,
		},
		"dst_firewall_group_ids": {
			TerraformName: resFirewallGroup,
		},
		"src_network_id": {
			TerraformName: resNetwork,
		},
		"dst_network_id": {
			TerraformName: resNetwork,
		},
	},

	// A WLAN broadcasts on the access points selected by ap_group_ids (UniFi ids
	// only known after the ApGroup reconciles — unifi_ap_group is new in the
	// wrapped provider v0.55.0), sits on a network (VLAN) via network_id, and —
	// when security is wpaeap — authenticates against a RADIUS profile via
	// radius_profile_id. All three are post-reconcile UniFi ids, so referencing
	// ApGroup/Network/RadiusProfile by name lets a consumer wire a Wlan through
	// the generated *Ref/*Selector companions, mirroring the firewall wiring
	// above; the raw ids stay settable directly.
	"unifi_wlan": {
		"ap_group_ids": {
			TerraformName: "unifi_ap_group",
		},
		"network_id": {
			TerraformName: resNetwork,
		},
		"radius_profile_id": {
			TerraformName: "unifi_radius_profile",
		},
	},

	// A RADIUS user can be pinned to a network, whose UniFi id is likewise only
	// known once that Network reconciles; referencing it mirrors the wlan wiring
	// above and leaves the raw networkId settable.
	"unifi_radius_user": {
		"network_id": {
			TerraformName: resNetwork,
		},
	},

	// A firewall policy references networks and a firewall zone on each side of
	// the match; the network_ids/zone_id fields live inside the single-nested
	// source and destination blocks, so the references are keyed by their nested
	// paths. Both are post-reconcile UniFi ids, so referencing Network and
	// FirewallZone by name mirrors the firewall_rule wiring above.
	"unifi_firewall_policy": {
		"source.network_ids": {
			TerraformName: resNetwork,
		},
		"source.zone_id": {
			TerraformName: resFirewallZone,
		},
		"destination.network_ids": {
			TerraformName: resNetwork,
		},
		"destination.zone_id": {
			TerraformName: resFirewallZone,
		},
	},
}

// Configure assigns each UniFi resource to its API ShortGroup and pins Kinds
// where the default derivation would collide. Kinds not overridden are left to
// upjet's default derivation (CamelCase of the resource suffix), which gives
// sensible names such as Rule, Record, Wlan, etc.
func Configure(p *ujconfig.Provider) {
	for name, group := range shortGroups {
		group := group
		kind := kindOverrides[name]
		refs := references[name]
		p.AddResourceConfigurator(name, func(r *ujconfig.Resource) {
			r.ShortGroup = group
			if kind != "" {
				r.Kind = kind
			}
			for field, ref := range refs {
				r.References[field] = ref
			}
		})
	}
}

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
	resPortProfile:       "port",

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

// Terraform resource and field names used across reference declarations, pulled
// out as constants because the goconst linter flags their repeated literals.
const (
	resVPNClient     = "unifi_vpn_client"
	resFirewallGroup = "unifi_firewall_group"
	resFirewallZone  = "unifi_firewall_zone"
	resNetwork       = "unifi_network"
	resPortProfile   = "unifi_port_profile"
	fieldNetworkID   = "network_id"
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
	// Accounts and clients belong to a network whose UniFi id may only be
	// available after the Network reconciles. The generated reference fields let
	// consumers express that relationship by managed-resource name.
	"unifi_account": {
		fieldNetworkID: {
			TerraformName: resNetwork,
		},
	},
	"unifi_client": {
		fieldNetworkID: {
			TerraformName: resNetwork,
		},
	},

	// Device management and individual port overrides can point at networks and
	// a reusable port profile. All raw id fields remain available for externally
	// managed dependencies.
	"unifi_device": {
		"mgmt_network_id": {
			TerraformName: resNetwork,
		},
		"port_override.excluded_networkconf_ids": {
			TerraformName: resNetwork,
		},
		"port_override.multicast_router_networkconf_ids": {
			TerraformName: resNetwork,
		},
		"port_override.native_networkconf_id": {
			TerraformName: resNetwork,
		},
		"port_override.port_profile_id": {
			TerraformName: resPortProfile,
		},
		"port_override.tagged_networkconf_ids": {
			TerraformName: resNetwork,
		},
		"port_override.voice_networkconf_id": {
			TerraformName: resNetwork,
		},
	},

	"unifi_traffic_route": {
		fieldNetworkID: {
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
	// under enterprise security — authenticates against a RADIUS profile via
	// radius_profile_id. Private pre-shared keys can independently target a
	// network as well. These are post-reconcile UniFi ids, so referencing the
	// managed resources by name lets a consumer wire a Wlan through the generated
	// *Ref/*Selector companions; the raw ids stay settable directly.
	"unifi_wlan": {
		"ap_group_ids": {
			TerraformName: "unifi_ap_group",
		},
		fieldNetworkID: {
			TerraformName: resNetwork,
		},
		"radius_profile_id": {
			TerraformName: "unifi_radius_profile",
		},
		"private_preshared_keys.network_id": {
			TerraformName: resNetwork,
		},
	},

	// A RADIUS user can be pinned to a network, whose UniFi id is likewise only
	// known once that Network reconciles; referencing it mirrors the wlan wiring
	// above and leaves the raw networkId settable.
	"unifi_radius_user": {
		fieldNetworkID: {
			TerraformName: resNetwork,
		},
	},

	// An L2TP/OpenVPN server authenticates against a RADIUS profile via
	// radiusprofile_id (no underscore after "radius" — the wrapped provider's
	// own field name). That id only exists once the RadiusProfile reconciles,
	// so referencing it lets a consumer wire a VpnServer to a RadiusProfile by
	// name through the generated *Ref/*Selector companions, mirroring the wlan
	// wiring above; the raw radiusprofileId stays settable directly.
	"unifi_vpn_server": {
		"radiusprofile_id": {
			TerraformName: "unifi_radius_profile",
		},
	},

	// A firewall policy references firewall groups, networks and a firewall zone
	// on each side of the match. The fields live inside the single-nested source
	// and destination blocks, so the references are keyed by their nested paths.
	// All are post-reconcile UniFi ids, so managed-resource references avoid
	// copying those ids into the policy.
	"unifi_firewall_policy": {
		"source.ip_group_id": {
			TerraformName: resFirewallGroup,
		},
		"source.port_group_id": {
			TerraformName: resFirewallGroup,
		},
		"source.network_ids": {
			TerraformName: resNetwork,
		},
		"source.zone_id": {
			TerraformName: resFirewallZone,
		},
		"destination.ip_group_id": {
			TerraformName: resFirewallGroup,
		},
		"destination.port_group_id": {
			TerraformName: resFirewallGroup,
		},
		"destination.network_ids": {
			TerraformName: resNetwork,
		},
		"destination.zone_id": {
			TerraformName: resFirewallZone,
		},
	},

	// A firewall zone groups managed networks by their reconciled UniFi ids.
	"unifi_firewall_zone": {
		"network_ids": {
			TerraformName: resNetwork,
		},
	},

	// Source limiting for a port-forward rule can reuse a managed firewall group.
	"unifi_port_forward": {
		"source_limiting.firewall_group_id": {
			TerraformName: resFirewallGroup,
		},
	},

	// A port profile can classify untagged, tagged, voice, multicast-router and
	// excluded traffic using managed Networks.
	resPortProfile: {
		"excluded_networkconf_ids": {
			TerraformName: resNetwork,
		},
		"multicast_router_networkconf_ids": {
			TerraformName: resNetwork,
		},
		"native_networkconf_id": {
			TerraformName: resNetwork,
		},
		"tagged_networkconf_ids": {
			TerraformName: resNetwork,
		},
		"voice_networkconf_id": {
			TerraformName: resNetwork,
		},
	},

	// Controller-wide IGMP snooping and honeypot settings can target managed
	// Networks without consumers copying reconciled UniFi ids into settings.
	"unifi_setting": {
		"igmp_snooping.network_ids": {
			TerraformName: resNetwork,
		},
		"ips.honeypot.network_id": {
			TerraformName: resNetwork,
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

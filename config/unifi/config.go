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
	"unifi_firewall_group":  "firewall",
	"unifi_firewall_policy": "firewall",
	"unifi_firewall_rule":   "firewall",
	"unifi_firewall_zone":   "firewall",

	// network
	"unifi_network": "network",
	"unifi_wan":     "network",
	"unifi_bgp":     "network",

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
	"unifi_vpn_client":       "vpn",
	"unifi_vpn_server":       "vpn",
	"unifi_wireguard_peer":   "vpn",

	// wlan
	"unifi_wlan": "wlan",
}

// kindOverrides pins the Kind for resources whose upjet-derived Kind would
// otherwise collide within the same group. unifi_static_route and
// unifi_traffic_route both derive to Kind "Route" in group "route", which would
// silently drop one of them, so they are disambiguated here.
var kindOverrides = map[string]string{
	"unifi_static_route":  "StaticRoute",
	"unifi_traffic_route": "TrafficRoute",
}

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
			TerraformName: "unifi_vpn_client",
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

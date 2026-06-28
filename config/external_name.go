// Package config holds the hand-maintained Upjet provider configuration:
// per-resource external-name configs, resource grouping and Kind overrides, and
// the wiring of the cluster- and namespaced-scoped providers.
package config

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// ExternalNameConfigs contains all external name configurations for this
// provider.
//
// The ubiquiti-community/unifi Terraform provider assigns opaque, server-side
// identifiers to every resource on create (UniFi controller object IDs), so the
// safe default for all of them is config.IdentifierFromProvider. None of these
// resources have a stable, user-supplied name that can serve as the import ID,
// so no NameAsIdentifier / TemplatedStringAsIdentifier refinements apply.
var ExternalNameConfigs = map[string]config.ExternalName{
	"unifi_account":          config.IdentifierFromProvider,
	"unifi_bgp":              config.IdentifierFromProvider,
	"unifi_client":           config.IdentifierFromProvider,
	"unifi_client_qos_rate":  config.IdentifierFromProvider,
	"unifi_device":           config.IdentifierFromProvider,
	"unifi_dns_record":       config.IdentifierFromProvider,
	"unifi_dynamic_dns":      config.IdentifierFromProvider,
	"unifi_firewall_group":   config.IdentifierFromProvider,
	"unifi_firewall_policy":  config.IdentifierFromProvider,
	"unifi_firewall_rule":    config.IdentifierFromProvider,
	"unifi_firewall_zone":    config.IdentifierFromProvider,
	"unifi_network":          config.IdentifierFromProvider,
	"unifi_port_forward":     config.IdentifierFromProvider,
	"unifi_port_profile":     config.IdentifierFromProvider,
	"unifi_power_supervisor": config.IdentifierFromProvider,
	"unifi_radius_profile":   config.IdentifierFromProvider,
	"unifi_radius_user":      config.IdentifierFromProvider,
	"unifi_setting":          config.IdentifierFromProvider,
	"unifi_site":             config.IdentifierFromProvider,
	"unifi_site_to_site_vpn": config.IdentifierFromProvider,
	"unifi_static_route":     config.IdentifierFromProvider,
	"unifi_traffic_route":    config.IdentifierFromProvider,
	"unifi_vpn_client":       config.IdentifierFromProvider,
	"unifi_vpn_server":       config.IdentifierFromProvider,
	"unifi_wan":              config.IdentifierFromProvider,
	"unifi_wireguard_peer":   config.IdentifierFromProvider,
	"unifi_wlan":             config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}

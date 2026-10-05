package config

import (
	"testing"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

const (
	testFirewallGroup = "unifi_firewall_group"
	testNetwork       = "unifi_network"
)

func TestProviderExposesResourceReferences(t *testing.T) {
	t.Parallel()

	expected := map[string]map[string]string{
		"unifi_account": {
			"network_id": testNetwork,
		},
		"unifi_client": {
			"network_id": testNetwork,
		},
		"unifi_device": {
			"mgmt_network_id":                                testNetwork,
			"port_override.excluded_networkconf_ids":         testNetwork,
			"port_override.multicast_router_networkconf_ids": testNetwork,
			"port_override.native_networkconf_id":            testNetwork,
			"port_override.port_profile_id":                  "unifi_port_profile",
			"port_override.tagged_networkconf_ids":           testNetwork,
			"port_override.voice_networkconf_id":             testNetwork,
		},
		"unifi_firewall_policy": {
			"source.ip_group_id":        testFirewallGroup,
			"source.port_group_id":      testFirewallGroup,
			"destination.ip_group_id":   testFirewallGroup,
			"destination.port_group_id": testFirewallGroup,
		},
		"unifi_firewall_zone": {
			"network_ids": testNetwork,
		},
		"unifi_port_forward": {
			"source_limiting.firewall_group_id": testFirewallGroup,
		},
		"unifi_port_profile": {
			"excluded_networkconf_ids":         testNetwork,
			"multicast_router_networkconf_ids": testNetwork,
			"native_networkconf_id":            testNetwork,
			"tagged_networkconf_ids":           testNetwork,
			"voice_networkconf_id":             testNetwork,
		},
		"unifi_setting": {
			"igmp_snooping.network_ids": testNetwork,
			"ips.honeypot.network_id":   testNetwork,
		},
		"unifi_wlan": {
			"private_preshared_keys.network_id": testNetwork,
		},
	}

	for _, tc := range []struct {
		name     string
		provider *ujconfig.Provider
	}{
		{name: "cluster scoped", provider: GetProvider()},
		{name: "namespaced", provider: GetProviderNamespaced()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for resourceName, fields := range expected {
				resource, ok := tc.provider.Resources[resourceName]
				if !ok {
					t.Fatalf("resource %q is not configured", resourceName)
				}

				for field, want := range fields {
					got, ok := resource.References[field]
					if !ok {
						t.Errorf("%s.%s has no cross-resource reference", resourceName, field)
						continue
					}
					if got.TerraformName != want {
						t.Errorf("%s.%s references %q, want %q", resourceName, field, got.TerraformName, want)
					}
				}
			}
		})
	}
}

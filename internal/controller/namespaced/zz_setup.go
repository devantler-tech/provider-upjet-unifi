// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	account "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/account/account"
	client "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/device/client"
	device "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/device/device"
	qosrate "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/device/qosrate"
	supervisor "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/device/supervisor"
	record "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/dns/record"
	group "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/firewall/group"
	policy "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/firewall/policy"
	rule "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/firewall/rule"
	zone "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/firewall/zone"
	bgp "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/network/bgp"
	network "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/network/network"
	wan "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/network/wan"
	forward "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/port/forward"
	profile "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/port/profile"
	providerconfig "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/providerconfig"
	profileradius "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/radius/profile"
	user "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/radius/user"
	staticroute "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/route/staticroute"
	trafficroute "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/route/trafficroute"
	dns "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/setting/dns"
	setting "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/setting/setting"
	site "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/site/site"
	clientvpn "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/vpn/client"
	peer "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/vpn/peer"
	server "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/vpn/server"
	tositevpn "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/vpn/tositevpn"
	wlan "github.com/devantler-tech/provider-upjet-unifi/internal/controller/namespaced/wlan/wlan"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		account.Setup,
		client.Setup,
		device.Setup,
		qosrate.Setup,
		supervisor.Setup,
		record.Setup,
		group.Setup,
		policy.Setup,
		rule.Setup,
		zone.Setup,
		bgp.Setup,
		network.Setup,
		wan.Setup,
		forward.Setup,
		profile.Setup,
		providerconfig.Setup,
		profileradius.Setup,
		user.Setup,
		staticroute.Setup,
		trafficroute.Setup,
		dns.Setup,
		setting.Setup,
		site.Setup,
		clientvpn.Setup,
		peer.Setup,
		server.Setup,
		tositevpn.Setup,
		wlan.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		account.SetupGated,
		client.SetupGated,
		device.SetupGated,
		qosrate.SetupGated,
		supervisor.SetupGated,
		record.SetupGated,
		group.SetupGated,
		policy.SetupGated,
		rule.SetupGated,
		zone.SetupGated,
		bgp.SetupGated,
		network.SetupGated,
		wan.SetupGated,
		forward.SetupGated,
		profile.SetupGated,
		providerconfig.SetupGated,
		profileradius.SetupGated,
		user.SetupGated,
		staticroute.SetupGated,
		trafficroute.SetupGated,
		dns.SetupGated,
		setting.SetupGated,
		site.SetupGated,
		clientvpn.SetupGated,
		peer.SetupGated,
		server.SetupGated,
		tositevpn.SetupGated,
		wlan.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

package config

import (
	"context"
	"slices"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/devantler-tech/provider-upjet-unifi/apis/cluster"
	accountv1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/account/v1alpha1"
	devicev1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/device/v1alpha1"
	firewallv1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/firewall/v1alpha1"
	networkv1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/network/v1alpha1"
	portv1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/port/v1alpha1"
	settingv1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/setting/v1alpha1"
	wlanv1alpha1 "github.com/devantler-tech/provider-upjet-unifi/apis/cluster/wlan/v1alpha1"
	"github.com/devantler-tech/provider-upjet-unifi/apis/namespaced"
	accountv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/account/v1alpha1"
	devicev1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/device/v1alpha1"
	firewallv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/firewall/v1alpha1"
	networkv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/network/v1alpha1"
	portv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/port/v1alpha1"
	settingv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/setting/v1alpha1"
	wlanv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/wlan/v1alpha1"
)

// The generated resolvers fill a raw id field from the external name of the
// resource a *Ref names. These tests run them against a fake API server that
// holds Networks, FirewallGroups and a PortProfile with known external names, which is what Crossplane does
// before each reconcile, so they show the references resolve to the right ids.

// resolvable is a managed resource with generated reference resolvers.
type resolvable interface {
	client.Object
	ResolveReferences(ctx context.Context, c client.Reader) error
}

// networkIDs maps each fixture Network's name to its controller id.
var networkIDs = map[string]string{"lan": "net-lan", "iot": "net-iot", "guest": "net-guest"}

func clusterNetworks() []client.Object {
	objects := make([]client.Object, 0, len(networkIDs))
	for name, id := range networkIDs {
		n := &networkv1alpha1.Network{}
		n.SetName(name)
		meta.SetExternalName(n, id)
		objects = append(objects, n)
	}
	return objects
}

func resolve(t *testing.T, addToScheme func(*runtime.Scheme) error, networks []client.Object, mg resolvable) {
	t.Helper()
	s := runtime.NewScheme()
	if err := addToScheme(s); err != nil {
		t.Fatalf("build the scheme: %v", err)
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(networks...).Build()
	if err := mg.ResolveReferences(context.Background(), c); err != nil {
		t.Fatalf("resolve references: %v", err)
	}
}

func ref(name string) *xpv1.Reference { return &xpv1.Reference{Name: name} }

func refs(names ...string) []xpv1.Reference {
	out := make([]xpv1.Reference, 0, len(names))
	for _, name := range names {
		out = append(out, xpv1.Reference{Name: name})
	}
	return out
}

func wantID(t *testing.T, field string, got *string, network string) {
	t.Helper()
	if got == nil || *got != networkIDs[network] {
		t.Errorf("%s = %v, want %q (the external name of Network %q)", field, deref(got), networkIDs[network], network)
	}
}

func wantIDs(t *testing.T, field string, got []*string, networks ...string) {
	t.Helper()
	values := make([]string, 0, len(got))
	for _, v := range got {
		values = append(values, deref(v))
	}
	want := make([]string, 0, len(networks))
	for _, n := range networks {
		want = append(want, networkIDs[n])
	}
	slices.Sort(values)
	slices.Sort(want)
	if !slices.Equal(values, want) {
		t.Errorf("%s = %v, want %v", field, values, want)
	}
}

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func TestAccountNetworkReferenceResolves(t *testing.T) {
	mg := &accountv1alpha1.Account{}
	mg.Spec.ForProvider.NetworkIDRef = ref("lan")
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantID(t, "account networkId", mg.Spec.ForProvider.NetworkID, "lan")
}

func TestClientNetworkReferenceResolves(t *testing.T) {
	mg := &devicev1alpha1.Client{}
	mg.Spec.ForProvider.NetworkIDRef = ref("iot")
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantID(t, "client networkId", mg.Spec.ForProvider.NetworkID, "iot")
}

func TestDeviceManagementNetworkReferenceResolves(t *testing.T) {
	mg := &devicev1alpha1.Device{}
	mg.Spec.ForProvider.MgmtNetworkIDRef = ref("lan")
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantID(t, "device mgmtNetworkId", mg.Spec.ForProvider.MgmtNetworkID, "lan")
}

func TestFirewallZoneNetworkReferencesResolve(t *testing.T) {
	mg := &firewallv1alpha1.Zone{}
	mg.Spec.ForProvider.NetworkIdsRefs = refs("lan", "iot")
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantIDs(t, "zone networkIds", mg.Spec.ForProvider.NetworkIds, "lan", "iot")
}

func TestPortProfileNetworkReferencesResolve(t *testing.T) {
	mg := &portv1alpha1.Profile{}
	p := &mg.Spec.ForProvider
	p.NativeNetworkconfIDRef = ref("lan")
	p.VoiceNetworkconfIDRef = ref("iot")
	p.TaggedNetworkconfIdsRefs = refs("iot", "guest")
	p.ExcludedNetworkconfIdsRefs = refs("guest")
	p.MulticastRouterNetworkconfIdsRefs = refs("lan", "guest")
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantID(t, "port profile nativeNetworkconfId", p.NativeNetworkconfID, "lan")
	wantID(t, "port profile voiceNetworkconfId", p.VoiceNetworkconfID, "iot")
	wantIDs(t, "port profile taggedNetworkconfIds", p.TaggedNetworkconfIds, "iot", "guest")
	wantIDs(t, "port profile excludedNetworkconfIds", p.ExcludedNetworkconfIds, "guest")
	wantIDs(t, "port profile multicastRouterNetworkconfIds", p.MulticastRouterNetworkconfIds, "lan", "guest")
}

func TestSettingNetworkReferencesResolve(t *testing.T) {
	mg := &settingv1alpha1.Setting{}
	p := &mg.Spec.ForProvider
	p.IgmpSnooping = &settingv1alpha1.IgmpSnoopingParameters{NetworkIdsRefs: refs("lan", "iot")}
	p.Ips = &settingv1alpha1.IpsParameters{Honeypot: []settingv1alpha1.HoneypotParameters{
		{NetworkIDRef: ref("guest")},
		{NetworkIDRef: ref("lan")},
	}}
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantIDs(t, "setting igmpSnooping.networkIds", p.IgmpSnooping.NetworkIds, "lan", "iot")
	wantID(t, "setting ips.honeypot[0].networkId", p.Ips.Honeypot[0].NetworkID, "guest")
	wantID(t, "setting ips.honeypot[1].networkId", p.Ips.Honeypot[1].NetworkID, "lan")
}

func TestWlanPrivateKeyNetworkReferenceResolves(t *testing.T) {
	mg := &wlanv1alpha1.Wlan{}
	mg.Spec.ForProvider.PrivatePresharedKeys = []wlanv1alpha1.PrivatePresharedKeysParameters{
		{NetworkIDRef: ref("iot")},
	}
	resolve(t, cluster.AddToScheme, clusterNetworks(), mg)
	wantID(t, "wlan privatePresharedKeys[0].networkId", mg.Spec.ForProvider.PrivatePresharedKeys[0].NetworkID, "iot")
}

// The namespaced API scope has its own generated resolvers, so each relationship is resolved
// there too, against Networks in the resource's own namespace.
const homeNamespace = "home"

func namespacedNetworks() []client.Object {
	objects := make([]client.Object, 0, len(networkIDs))
	for name, id := range networkIDs {
		n := &networkv1alpha1m.Network{}
		n.SetName(name)
		n.SetNamespace(homeNamespace)
		meta.SetExternalName(n, id)
		objects = append(objects, n)
	}
	return objects
}

func namespacedRef(name string) *xpv1.NamespacedReference {
	return &xpv1.NamespacedReference{Name: name}
}

func namespacedRefs(names ...string) []xpv1.NamespacedReference {
	out := make([]xpv1.NamespacedReference, 0, len(names))
	for _, name := range names {
		out = append(out, xpv1.NamespacedReference{Name: name})
	}
	return out
}

func TestNamespacedAccountNetworkReferenceResolves(t *testing.T) {
	mg := &accountv1alpha1m.Account{}
	mg.SetNamespace(homeNamespace)
	mg.Spec.ForProvider.NetworkIDRef = namespacedRef("lan")
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), mg)
	wantID(t, "namespaced account networkId", mg.Spec.ForProvider.NetworkID, "lan")
}

func TestNamespacedClientAndDeviceNetworkReferencesResolve(t *testing.T) {
	c := &devicev1alpha1m.Client{}
	c.SetNamespace(homeNamespace)
	c.Spec.ForProvider.NetworkIDRef = namespacedRef("iot")
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), c)
	wantID(t, "namespaced client networkId", c.Spec.ForProvider.NetworkID, "iot")

	d := &devicev1alpha1m.Device{}
	d.SetNamespace(homeNamespace)
	d.Spec.ForProvider.MgmtNetworkIDRef = namespacedRef("lan")
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), d)
	wantID(t, "namespaced device mgmtNetworkId", d.Spec.ForProvider.MgmtNetworkID, "lan")
}

func TestNamespacedZoneAndPortProfileNetworkReferencesResolve(t *testing.T) {
	z := &firewallv1alpha1m.Zone{}
	z.SetNamespace(homeNamespace)
	z.Spec.ForProvider.NetworkIdsRefs = namespacedRefs("lan", "iot")
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), z)
	wantIDs(t, "namespaced zone networkIds", z.Spec.ForProvider.NetworkIds, "lan", "iot")

	pr := &portv1alpha1m.Profile{}
	pr.SetNamespace(homeNamespace)
	p := &pr.Spec.ForProvider
	p.NativeNetworkconfIDRef = namespacedRef("lan")
	p.VoiceNetworkconfIDRef = namespacedRef("iot")
	p.TaggedNetworkconfIdsRefs = namespacedRefs("iot", "guest")
	p.ExcludedNetworkconfIdsRefs = namespacedRefs("guest")
	p.MulticastRouterNetworkconfIdsRefs = namespacedRefs("lan")
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), pr)
	wantID(t, "namespaced port profile nativeNetworkconfId", p.NativeNetworkconfID, "lan")
	wantID(t, "namespaced port profile voiceNetworkconfId", p.VoiceNetworkconfID, "iot")
	wantIDs(t, "namespaced port profile taggedNetworkconfIds", p.TaggedNetworkconfIds, "iot", "guest")
	wantIDs(t, "namespaced port profile excludedNetworkconfIds", p.ExcludedNetworkconfIds, "guest")
	wantIDs(t, "namespaced port profile multicastRouterNetworkconfIds", p.MulticastRouterNetworkconfIds, "lan")
}

func TestNamespacedSettingAndWlanNetworkReferencesResolve(t *testing.T) {
	s := &settingv1alpha1m.Setting{}
	s.SetNamespace(homeNamespace)
	p := &s.Spec.ForProvider
	p.IgmpSnooping = &settingv1alpha1m.IgmpSnoopingParameters{NetworkIdsRefs: namespacedRefs("lan", "iot")}
	p.Ips = &settingv1alpha1m.IpsParameters{Honeypot: []settingv1alpha1m.HoneypotParameters{{NetworkIDRef: namespacedRef("guest")}}}
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), s)
	wantIDs(t, "namespaced setting igmpSnooping.networkIds", p.IgmpSnooping.NetworkIds, "lan", "iot")
	wantID(t, "namespaced setting ips.honeypot[0].networkId", p.Ips.Honeypot[0].NetworkID, "guest")

	w := &wlanv1alpha1m.Wlan{}
	w.SetNamespace(homeNamespace)
	w.Spec.ForProvider.PrivatePresharedKeys = []wlanv1alpha1m.PrivatePresharedKeysParameters{{NetworkIDRef: namespacedRef("iot")}}
	resolve(t, namespaced.AddToScheme, namespacedNetworks(), w)
	wantID(t, "namespaced wlan privatePresharedKeys[0].networkId", w.Spec.ForProvider.PrivatePresharedKeys[0].NetworkID, "iot")
}

// FirewallGroups and a PortProfile with known controller ids, beside the Networks.
const (
	groupWeb      = "fg-web"
	groupPorts    = "fg-ports"
	profileUplink = "pp-uplink"
)

func clusterFixtures() []client.Object {
	objects := clusterNetworks()
	for name, id := range map[string]string{"web": groupWeb, "ports": groupPorts} {
		g := &firewallv1alpha1.Group{}
		g.SetName(name)
		meta.SetExternalName(g, id)
		objects = append(objects, g)
	}
	p := &portv1alpha1.Profile{}
	p.SetName("uplink")
	meta.SetExternalName(p, profileUplink)
	return append(objects, p)
}

func wantExternalName(t *testing.T, field string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %q", field, deref(got), want)
	}
}

func TestFirewallPolicyGroupReferencesResolve(t *testing.T) {
	mg := &firewallv1alpha1.Policy{}
	p := &mg.Spec.ForProvider
	p.Source = &firewallv1alpha1.SourceParameters{IPGroupIDRef: ref("web"), PortGroupIDRef: ref("ports")}
	p.Destination = &firewallv1alpha1.DestinationParameters{IPGroupIDRef: ref("web"), PortGroupIDRef: ref("ports")}
	resolve(t, cluster.AddToScheme, clusterFixtures(), mg)
	wantExternalName(t, "policy source.ipGroupId", p.Source.IPGroupID, groupWeb)
	wantExternalName(t, "policy source.portGroupId", p.Source.PortGroupID, groupPorts)
	wantExternalName(t, "policy destination.ipGroupId", p.Destination.IPGroupID, groupWeb)
	wantExternalName(t, "policy destination.portGroupId", p.Destination.PortGroupID, groupPorts)
}

func TestPortForwardSourceLimitReferenceResolves(t *testing.T) {
	mg := &portv1alpha1.Forward{}
	mg.Spec.ForProvider.SourceLimiting = &portv1alpha1.SourceLimitingParameters{FirewallGroupIDRef: ref("web")}
	resolve(t, cluster.AddToScheme, clusterFixtures(), mg)
	wantExternalName(t, "port forward sourceLimiting.firewallGroupId", mg.Spec.ForProvider.SourceLimiting.FirewallGroupID, groupWeb)
}

func TestDevicePortOverrideReferencesResolve(t *testing.T) {
	mg := &devicev1alpha1.Device{}
	mg.Spec.ForProvider.PortOverride = []devicev1alpha1.PortOverrideParameters{{
		PortProfileIDRef:                  ref("uplink"),
		NativeNetworkconfIDRef:            ref("lan"),
		VoiceNetworkconfIDRef:             ref("iot"),
		TaggedNetworkconfIdsRefs:          refs("iot", "guest"),
		ExcludedNetworkconfIdsRefs:        refs("guest"),
		MulticastRouterNetworkconfIdsRefs: refs("lan"),
	}}
	resolve(t, cluster.AddToScheme, clusterFixtures(), mg)
	o := mg.Spec.ForProvider.PortOverride[0]
	wantExternalName(t, "device portOverride[0].portProfileId", o.PortProfileID, profileUplink)
	wantID(t, "device portOverride[0].nativeNetworkconfId", o.NativeNetworkconfID, "lan")
	wantID(t, "device portOverride[0].voiceNetworkconfId", o.VoiceNetworkconfID, "iot")
	wantIDs(t, "device portOverride[0].taggedNetworkconfIds", o.TaggedNetworkconfIds, "iot", "guest")
	wantIDs(t, "device portOverride[0].excludedNetworkconfIds", o.ExcludedNetworkconfIds, "guest")
	wantIDs(t, "device portOverride[0].multicastRouterNetworkconfIds", o.MulticastRouterNetworkconfIds, "lan")
}

// The same FirewallGroups and PortProfile in the namespaced scope, beside the namespaced Networks.
func namespacedFixtures() []client.Object {
	objects := namespacedNetworks()
	for name, id := range map[string]string{"web": groupWeb, "ports": groupPorts} {
		g := &firewallv1alpha1m.Group{}
		g.SetName(name)
		g.SetNamespace(homeNamespace)
		meta.SetExternalName(g, id)
		objects = append(objects, g)
	}
	p := &portv1alpha1m.Profile{}
	p.SetName("uplink")
	p.SetNamespace(homeNamespace)
	meta.SetExternalName(p, profileUplink)
	return append(objects, p)
}

func TestNamespacedGroupAndPortReferencesResolve(t *testing.T) {
	pol := &firewallv1alpha1m.Policy{}
	pol.SetNamespace(homeNamespace)
	pol.Spec.ForProvider.Source = &firewallv1alpha1m.SourceParameters{IPGroupIDRef: namespacedRef("web"), PortGroupIDRef: namespacedRef("ports")}
	pol.Spec.ForProvider.Destination = &firewallv1alpha1m.DestinationParameters{IPGroupIDRef: namespacedRef("web"), PortGroupIDRef: namespacedRef("ports")}
	resolve(t, namespaced.AddToScheme, namespacedFixtures(), pol)
	wantExternalName(t, "namespaced policy source.ipGroupId", pol.Spec.ForProvider.Source.IPGroupID, groupWeb)
	wantExternalName(t, "namespaced policy source.portGroupId", pol.Spec.ForProvider.Source.PortGroupID, groupPorts)
	wantExternalName(t, "namespaced policy destination.ipGroupId", pol.Spec.ForProvider.Destination.IPGroupID, groupWeb)
	wantExternalName(t, "namespaced policy destination.portGroupId", pol.Spec.ForProvider.Destination.PortGroupID, groupPorts)

	fwd := &portv1alpha1m.Forward{}
	fwd.SetNamespace(homeNamespace)
	fwd.Spec.ForProvider.SourceLimiting = &portv1alpha1m.SourceLimitingParameters{FirewallGroupIDRef: namespacedRef("web")}
	resolve(t, namespaced.AddToScheme, namespacedFixtures(), fwd)
	wantExternalName(t, "namespaced port forward sourceLimiting.firewallGroupId", fwd.Spec.ForProvider.SourceLimiting.FirewallGroupID, groupWeb)

	dev := &devicev1alpha1m.Device{}
	dev.SetNamespace(homeNamespace)
	dev.Spec.ForProvider.PortOverride = []devicev1alpha1m.PortOverrideParameters{{
		PortProfileIDRef:                  namespacedRef("uplink"),
		NativeNetworkconfIDRef:            namespacedRef("lan"),
		VoiceNetworkconfIDRef:             namespacedRef("iot"),
		TaggedNetworkconfIdsRefs:          namespacedRefs("iot", "guest"),
		ExcludedNetworkconfIdsRefs:        namespacedRefs("guest"),
		MulticastRouterNetworkconfIdsRefs: namespacedRefs("lan"),
	}}
	resolve(t, namespaced.AddToScheme, namespacedFixtures(), dev)
	o := dev.Spec.ForProvider.PortOverride[0]
	wantExternalName(t, "namespaced device portOverride[0].portProfileId", o.PortProfileID, profileUplink)
	wantID(t, "namespaced device portOverride[0].nativeNetworkconfId", o.NativeNetworkconfID, "lan")
	wantID(t, "namespaced device portOverride[0].voiceNetworkconfId", o.VoiceNetworkconfID, "iot")
	wantIDs(t, "namespaced device portOverride[0].taggedNetworkconfIds", o.TaggedNetworkconfIds, "iot", "guest")
	wantIDs(t, "namespaced device portOverride[0].excludedNetworkconfIds", o.ExcludedNetworkconfIds, "guest")
	wantIDs(t, "namespaced device portOverride[0].multicastRouterNetworkconfIds", o.MulticastRouterNetworkconfIds, "lan")
}

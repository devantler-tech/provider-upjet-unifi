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
	networkv1alpha1m "github.com/devantler-tech/provider-upjet-unifi/apis/namespaced/network/v1alpha1"
)

// The generated resolvers fill a raw id field from the external name of the
// Network a *Ref names. These tests run them against a fake API server that
// holds Networks with known external names, which is what Crossplane does
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

// A namespaced resource resolves a Network in its own namespace.
func TestNamespacedAccountNetworkReferenceResolves(t *testing.T) {
	n := &networkv1alpha1m.Network{}
	n.SetName("lan")
	n.SetNamespace("home")
	meta.SetExternalName(n, networkIDs["lan"])

	mg := &accountv1alpha1m.Account{}
	mg.SetNamespace("home")
	mg.Spec.ForProvider.NetworkIDRef = &xpv1.NamespacedReference{Name: "lan"}
	resolve(t, namespaced.AddToScheme, []client.Object{n}, mg)
	wantID(t, "namespaced account networkId", mg.Spec.ForProvider.NetworkID, "lan")
}

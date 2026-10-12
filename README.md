# provider-upjet-unifi

`provider-upjet-unifi` lets you manage a **Ubiquiti UniFi** network (networks, Wi-Fi, firewall, VPN,
DNS and more) as Kubernetes resources. It is for people who already run
[Crossplane](https://crossplane.io/), a Kubernetes add-on that keeps external systems in line with
resources you declare, and who want their UniFi controller configured from Git instead of by hand.

It is a Crossplane *provider*: the plug-in that teaches Crossplane one external system. The resource
types are generated with [Upjet](https://github.com/crossplane/upjet) from the community
[`ubiquiti-community/terraform-provider-unifi`](https://github.com/ubiquiti-community/terraform-provider-unifi)
Terraform provider.

## Install

You need a Kubernetes cluster with Crossplane installed. Apply
[`examples/install.yaml`](examples/install.yaml):

```sh
kubectl apply -f https://raw.githubusercontent.com/devantler-tech/provider-upjet-unifi/main/examples/install.yaml
kubectl get providers.pkg.crossplane.io provider-upjet-unifi
```

The provider is ready when the second command shows `INSTALLED` and `HEALTHY` as `True`. Installing
it only adds resource types to the cluster; nothing reaches a UniFi controller until you configure
credentials and create a resource.

The example pins one released version of the package
`ghcr.io/devantler-tech/provider-upjet-unifi`. Newer versions are listed under
[Releases](https://github.com/devantler-tech/provider-upjet-unifi/releases).

## Configure

The provider reads your controller's address and credentials from a Kubernetes `Secret`, which a
`ProviderConfig` points at. Start from the templates in
[`examples/cluster/providerconfig/`](examples/cluster/providerconfig/):

1. Copy `secret.yaml.tmpl`, fill in `api_url` and either an `api_key` or a `username` and
   `password`, and apply it. Keep the filled-in file out of Git.
2. Apply `providerconfig.yaml`.

That configures the cluster-scoped examples. The namespaced examples reference a
`ClusterProviderConfig` instead: to run those, use the same secret and apply
`clusterproviderconfig.yaml` from
[`examples/namespaced/providerconfig/`](examples/namespaced/providerconfig/). A namespaced resource
can also use a `ProviderConfig` in its own namespace; that directory has a template for it too.

Supported credential keys are `api_url`, `username`, `password`, `api_key`, `site`,
`allow_insecure`, `hardware_id` and `cloud_connector`. Set `allow_insecure` only for a controller
with a self-signed certificate.

## Create a resource

Every resource you create changes the live controller, so begin with something harmless on a
network you can afford to experiment on. Most examples come in two flavours: cluster-scoped
(the resource belongs to no namespace) and namespaced (it lives in a Kubernetes namespace). You
can create as many resources of either flavour as you need.

| Resource         | Cluster                                          | Namespaced                                          |
|------------------|--------------------------------------------------|-----------------------------------------------------|
| `Network`        | [network](examples/cluster/network/network.yaml) | [network](examples/namespaced/network/network.yaml) |
| `Wlan`           | [wlan](examples/cluster/wlan/wlan.yaml)          | [wlan](examples/namespaced/wlan/wlan.yaml)          |
| `firewall.Rule`  | [rule](examples/cluster/firewall/rule.yaml)      | —                                                   |
| `radius.Profile` | [profile](examples/cluster/radius/profile.yaml)  | [profile](examples/namespaced/radius/profile.yaml)  |
| `radius.User`    | [user](examples/cluster/radius/user.yaml)        | [user](examples/namespaced/radius/user.yaml)        |
| `vpn.Server`     | [server](examples/cluster/vpn/server.yaml)       | [server](examples/namespaced/vpn/server.yaml)       |

The RADIUS and VPN examples are meant to be applied together: the VPN server points at the RADIUS
profile through `radiusprofileIdRef`, and the RADIUS user at its network through `networkIdRef`, so
neither needs a controller-assigned id pasted in after the fact.

To manage something that already exists on the controller, tell Crossplane which object it is
before the first reconcile, so it adopts the object instead of creating a duplicate. Crossplane's
[import guide](https://docs.crossplane.io/latest/guides/import-existing-resources/) describes how.

## Contribute

[CONTRIBUTING.md](CONTRIBUTING.md) covers a safe first change, the local checks and how to submit.
Most of this repository is generated, and the guide says which parts are and how to change them.
Conventions for AI agents are in [`AGENTS.md`](AGENTS.md).

## Report a bug

Open an [issue](https://github.com/devantler-tech/provider-upjet-unifi/issues). Report a security
vulnerability privately through
[Report a vulnerability](https://github.com/devantler-tech/provider-upjet-unifi/security/advisories/new)
instead.

## License

The provider code is Apache-2.0 (see [`LICENSE`](LICENSE)). The upstream
`ubiquiti-community/terraform-provider-unifi` it is generated from is MPL-2.0; it is fetched at build
time, not vendored.

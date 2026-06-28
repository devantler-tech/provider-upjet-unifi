# provider-upjet-unifi

`provider-upjet-unifi` is a [Crossplane](https://crossplane.io/) provider for **Ubiquiti UniFi**,
built with [Upjet](https://github.com/crossplane/upjet) from the
[`ubiquiti-community/terraform-provider-unifi`](https://github.com/ubiquiti-community/terraform-provider-unifi)
Terraform provider (registry source `ubiquiti-community/unifi`). It exposes XRM-conformant managed
resources so UniFi objects — networks, WLANs, firewall zones/policies/rules, VPN & WireGuard, devices,
clients, RADIUS, DNS, sites and settings — can be managed as Kubernetes custom resources and reconciled
by GitOps.

## Install

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-upjet-unifi
spec:
  package: ghcr.io/devantler-tech/provider-upjet-unifi:v0.1.0
```

## Configure

Create a `ProviderConfig` referencing a `Secret` with the UniFi controller credentials. Supported keys:
`api_url`, `username`, `password`, `api_key`, `site`, `allow_insecure`, `hardware_id`, `cloud_connector`
(use either username/password **or** an `api_key`). See [`examples/`](examples/) for cluster- and
namespaced-scoped `ProviderConfig` and managed-resource manifests.

```yaml
apiVersion: unifi.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: unifi-creds
      key: credentials
```

## Develop

```console
make generate    # regenerate APIs/CRDs/controllers from the Terraform schema
make build       # build the provider + Crossplane package (xpkg)
make run         # run against the current kube context
make test        # unit tests
make lint        # golangci-lint
```

The provider's API surface is **generated** — edit `config/**` (resource coverage, groups) or
`internal/clients/unifi.go` (auth) and re-run `make generate`; never hand-edit `apis/**`,
`internal/controller/**`, or `package/crds/**`. See [`AGENTS.md`](AGENTS.md) for the full conventions.

## Report a bug

Open an [issue](https://github.com/devantler-tech/provider-upjet-unifi/issues).

## License

The provider code is Apache-2.0 (see [`LICENSE`](LICENSE)). The upstream
`ubiquiti-community/terraform-provider-unifi` it is generated from is MPL-2.0; it is fetched at build
time, not vendored.

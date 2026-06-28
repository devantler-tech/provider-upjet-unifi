# AGENTS.md — provider-upjet-unifi

Conventions for AI agents working in this repo. This is the **canonical** instructions file (plain
Markdown, read natively by GitHub Copilot, Cursor, Codex, … and by Claude Code via the `CLAUDE.md` →
`@AGENTS.md` shim). The monorepo's root [`AGENTS.md`](https://github.com/devantler-tech/monorepo/blob/main/AGENTS.md)
holds the **shared engineering contract** (PR/commit conventions, trust gate, guardrails, draft-PR
discipline); the rules below are what's **specific to this repo**.

## What this repo is

A [Crossplane](https://crossplane.io) provider for **Ubiquiti UniFi**, generated with
[Upjet](https://github.com/crossplane/upjet) from the
[`ubiquiti-community/terraform-provider-unifi`](https://github.com/ubiquiti-community/terraform-provider-unifi)
Terraform provider (registry source `ubiquiti-community/unifi`). It lets UniFi objects — networks, WLANs,
firewall zones/policies/rules, VPN & WireGuard, devices, clients, RADIUS, DNS, sites and settings — be
managed as Kubernetes custom resources and reconciled by GitOps, the same way the `platform` cluster
consumes `provider-upjet-github`. It is a **shared library** in the devantler-tech portfolio.

The provider's API surface (Go types, CRDs, controllers) is **generated** from the Terraform provider's
schema; at runtime the generated controllers drive the Terraform provider's resource logic while exposing
a pure Kubernetes API.

## Repository layout

| Path                           | Contents                                                                                    |
|--------------------------------|---------------------------------------------------------------------------------------------|
| `config/external_name.go`      | one entry per Terraform resource — **drives coverage** (only configured resources generate) |
| `config/unifi/config.go`       | resource grouping (ShortGroups) + Kind disambiguation                                       |
| `config/provider.go`           | wires the group configurators for both cluster- and namespaced-scoped providers             |
| `internal/clients/unifi.go`    | `ProviderConfig` → Terraform provider auth wiring (hand-maintained)                         |
| `apis/{cluster,namespaced}/**` | **generated** API types                                                                     |
| `internal/controller/**`       | **generated** controllers                                                                   |
| `package/crds/**`              | **generated** CRDs                                                                          |
| `examples/**`                  | example `ProviderConfig` + managed-resource manifests                                       |
| `Makefile`                     | `TERRAFORM_PROVIDER_*` pins (source/version) + Upjet build wiring                           |

## Authentication

`internal/clients/unifi.go` reads the `ProviderConfig` credentials secret (JSON) into the Terraform
provider configuration. Keys match the upstream provider schema: `api_url`, `username`, `password`,
`api_key`, `site`, `allow_insecure`, `hardware_id`, `cloud_connector` (only present, non-empty keys are
forwarded). Self-signed controllers are common — `allow_insecure` is supported.

## Validate before every PR

```sh
make generate            # regenerate from the Terraform schema; never hand-edit generated output
git diff --exit-code     # generated tree must be clean (no codegen drift)
go build ./...           # compiles
go vet ./...
go test ./...
make lint                # golangci-lint, 0 issues
```

`make build` additionally assembles the Crossplane package image (xpkg) — it needs Docker and is what CI
publishes on release.

## Generated artifacts — never hand-edit

`apis/**`, `internal/controller/**`, and `package/crds/**` are produced by `make generate`. Change behaviour
by editing `config/**` (external-name configs, groups, Kind overrides) or `internal/clients/unifi.go`, then
**re-run the generator** — never edit the generated files directly.

## Maintenance

The **roadmap of record** is this repo's GitHub Issues (`roadmap`-labelled epics + milestones). Triage
incoming issues into that structure; implementing PRs use `Fixes #N`. Shared cross-repo rules and the
draft-PR discipline live in the monorepo root `AGENTS.md`.

Repo-specific watch-list for the daily engineer:

- **Track upstream provider releases.** When
  [`ubiquiti-community/terraform-provider-unifi`](https://github.com/ubiquiti-community/terraform-provider-unifi)
  ships a release, bump `TERRAFORM_PROVIDER_VERSION` (+ `TERRAFORM_NATIVE_PROVIDER_BINARY`) in the
  `Makefile`, `make generate`, and open a draft PR with the regenerated tree. New upstream resources →
  add `config/external_name.go` entries so they generate.
- **Resource coverage.** All resources the upstream provider exposes should have an external-name config.
  A new opaque-ID resource defaults to `config.IdentifierFromProvider`; refine to
  `NameAsIdentifier`/`TemplatedStringAsIdentifier` only when the upstream identity warrants it.
- **CI/CD health.** CI runs lint + `make generate` drift + build/test; the release workflow publishes the
  xpkg to `ghcr.io/devantler-tech/provider-upjet-unifi`. Keep both green; root-cause failures.
- **Agent-file freshness.** Keep this `AGENTS.md` in sync with the actual layout, the validate command,
  and the generated-artifact list. Copilot reads `AGENTS.md` directly — don't let it drift.

# Contributing to provider-upjet-unifi

You can help by improving the documentation, fixing an example, or extending which UniFi resources
the provider covers. Discuss substantial changes in an
[issue](https://github.com/devantler-tech/provider-upjet-unifi/issues) first. Report sensitive
vulnerabilities through this repository's private
[Report a vulnerability](https://github.com/devantler-tech/provider-upjet-unifi/security/advisories/new)
form, not a public issue or PR.

## Make a first documentation change

You need Git. This first path needs no Go toolchain, cluster, UniFi controller or credentials.

Fork `devantler-tech/provider-upjet-unifi` on GitHub. Replace `YOUR-LOGIN` with your login:

```sh
git clone https://github.com/YOUR-LOGIN/provider-upjet-unifi.git
cd provider-upjet-unifi
git switch -c docs/first-contribution
```

Make a small wording improvement in `README.md`, then check the change for whitespace errors:

```sh
git diff --check
```

The command must exit successfully and print nothing. [CI](.github/workflows/ci.yml) also lints and
spell-checks the documentation.

## Change the code

For anything beyond documentation you also need Go (the version in [`go.mod`](go.mod) or newer).
A first check that needs no cluster or credentials is to build the code generator:

```sh
go build ./cmd/generator
```

It exits successfully and prints nothing; the first run downloads dependencies and can take a few
minutes. This proves the hand-written code compiles. It does not run the generator, build the
package or contact a controller.

Know which files you may edit:

- **Hand-written:** `config/**` (which resources exist, how they are grouped and named),
  `internal/clients/unifi.go` (how credentials reach the provider), `examples/**`, `hack/**`,
  `scripts/**` and the documentation.
- **Generated, never edited by hand:** `apis/**`, `internal/controller/**` and `package/crds/**`.
  Change the hand-written inputs and regenerate instead.

The everyday targets are:

```sh
make generate    # regenerate the generated files from the Terraform provider's schema
make test        # unit tests
make lint        # golangci-lint
make build       # build the provider and its Crossplane package (needs Docker)
make run         # run the provider against your current kubectl context
```

`make run` acts on whichever cluster your `kubectl` currently points at, so check that first. The
full validation sequence is in [AGENTS.md](AGENTS.md#validate-before-every-pr). Regeneration
downloads the upstream Terraform provider. CI runs the same checks on every pull request, so a
small change can rely on CI for the heavy steps.

Example manifests are checked against the generated resource types by
[`hack/validate-examples.sh`](hack/validate-examples.sh). It needs `kubectl` and a **disposable**
cluster, never one you care about, and no UniFi controller.

## Keep development and operations separate

Publishing a release, signing the package and installing it on a real cluster are maintainer
operations, not onboarding steps. Do not point a locally built provider at a controller you do not
own, and never commit controller addresses, credentials or filled-in secret templates.

## Submit your change

Review the diff, configure commit signing and stage only your changed paths. Use a Conventional
Commit subject describing the outcome; the release workflow uses it to decide whether a merge
releases and which version it produces. For the README-only first change above:

```sh
git add README.md
git commit -S -m "docs: clarify the install steps"
git push -u origin docs/first-contribution
```

Open a **draft** PR to `devantler-tech/provider-upjet-unifi:main` using the inherited
[PR template](https://github.com/devantler-tech/.github/blob/main/.github/PULL_REQUEST_TEMPLATE.md).
Keep Why/What focused on the reader-facing problem and outcome, link the issue for substantial work,
and put detailed validation evidence in a PR comment or check summary. Disclose AI assistance when
used. Address findings; required CI and a successful substantive review of the exact current head
must pass before merge. Follow the [code of conduct](CODE_OF_CONDUCT.md) and preserve the
[Apache-2.0 license](LICENSE).

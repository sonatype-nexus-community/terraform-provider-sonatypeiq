# Contributing

Pull requests are welcome - we thank you in advance. All we ask is that your PR has a clear description of its purpose or intent and that the purpose of a PR is a singluar intent - we'd rather accept multiple smaller PRs than fewer big ones.

## Setup

This provider uses the Custom Provider Framework from HashiCorp. A great reference is available from HashiCorp [here](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider).

Development is conducted using Go version specified in `go.mod`.

We utilise [tfplugindocs](github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs) for documentation generation, and this should be included in the `go.mod` and `go.sum`.

Documentation and examples should be regenerated and included in Pull Requests - run `cd tools && go generate ./...` before finalising your PR.

## Linting

Run `golangci-lint run`.

## Testing

### Unit Tests

These can be run locally by running `go test -v -cover ./internal/provider/`.

## Acceptance/Integration Tests

All acceptance tests require an active and licensed Sonatype IQ Server, so they do not run for Pull Requests until a maintainer approves them (see [Continuous Integration](#continuous-integration)). Without `TF_ACC=1` they skip cleanly.

To run Integration Tests locally, set the following 3 environment variables and then run `TF_ACC=1 go test -v -cover ./internal/provider/`:
- `IQ_SERVER_URL`: Full URL to your Sonatype IQ Server
- `IQ_SERVER_USERNAME`: Username to authenticate with
- `IQ_SERVER_PASSWORD`: Password to authenticate with

Some tests only apply from a given Sonatype IQ Server release and skip themselves on older servers. They find the release from `NXIQ_VERSION` (for example `1.201.0-02`), which CI sets, and otherwise ask the server at `IQ_SERVER_URL` - so you normally don't need to set it. A test that is skipped because the release could not be determined says so.

It is helpful when submitting Pull Requests to confirm whether you have been able to execute the Integration Tests locally, but not mandatory.

## Continuous Integration

CI follows the Sonatype Community [Pull Request CI & Secrets standard](https://contribute.sonatype.com/docs/standards/pull-request-ci/): code from a Pull Request never runs where secrets exist, and only data crosses that boundary.

### Workflows

| Workflow | Triggers | Secrets | Purpose |
|---|---|---|---|
| `test-untrusted.yml` (Test (Untrusted)) | Every Pull Request (including forks), push to `main` | None | **Tier 1.** Build, lint, docs check, workflow lint (`actionlint`, `zizmor`), and unit tests |
| `prepare-lifecycle-scan.yml` | Every Pull Request (including forks) | None | **Tier 1.** Produces the `lifecycle-scan-input` artifact (a dependency list). Small and quick, and separate from the tests so the Lifecycle check does not wait for them |
| `lifecycle-evaluate.yml` | After Prepare Lifecycle Scan Input completes for a Pull Request | Lifecycle | **Tier 2.** Evaluates the scan input in Sonatype Lifecycle (stage `develop`, advisory) and reports a check run. Never checks out the Pull Request |
| `approve-acceptance-tests.yml` | `/ok-to-test` comment; Pull Requests by maintainers | None | **Tier 3 gate.** Records commit-bound approval and starts the acceptance tests |
| `test-trusted.yml` (Test (Trusted)) | Push to `main`, or dispatched by the gate | IQ Server licence, Lifecycle | **Tier 3.** Acceptance tests on a licensed Sonatype IQ Server across the supported IQ Server and Terraform versions. On `main` it also runs the authoritative Lifecycle evaluation (stage `build`) |
| `release.yml` | Tag `v*` | Signing key, Lifecycle | Lifecycle evaluation (stage `release`), then GoReleaser |

### Approving acceptance tests (maintainers)

Pull Requests from forks cannot use repository secrets, so the acceptance tests wait for a maintainer. After reviewing the code (including any changes to `.github/`, `go.mod`, `tools/` and the test helpers), comment on the Pull Request:

```text
/ok-to-test <sha>
```

`<sha>` is the commit you reviewed (at least 7 characters). The approval belongs to that exact commit: it is rejected if the Pull Request head has moved, and a later push needs a new approval. Only people with the Maintain role or higher (the `terraform-provider-sonatypeiq` maintainers team) can approve. Pull Requests opened by maintainers are approved automatically.

The result appears on the Pull Request as the `Acceptance Tests (Licensed IQ)` commit status. Pull Requests that only change `CHANGELOG.md` or `README.md` report that status as skipped. A newer commit on a Pull Request cancels the in-progress runs for the older one.

### Secrets and Environments

- `NXIQ_LICENSE` is an Environment secret of `acceptance-tests`, which only deploys from `main`.
- `GPG_PRIVATE_KEY` and `PASSPHRASE` are Environment secrets of `release`, which only deploys from `v*` tags.
- Sonatype Lifecycle credentials are repository secrets used by `lifecycle-evaluate.yml`, `test-trusted.yml` and `release.yml`.

When changing workflows, run `actionlint` and `zizmor .github/workflows` - both are enforced by Test (Untrusted).

## Sign off your commits

Please sign off your commits, to show that you agree to publish your changes under the current terms and licenses of the project, and to indicate agreement with [Developer Certificate of Origin (DCO)](https://developercertificate.org/).

```shell
git commit --signed-off ...
```
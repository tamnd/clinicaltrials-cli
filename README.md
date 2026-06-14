# clinicaltrials

Search ClinicalTrials.gov clinical studies from the command line.

`clinicaltrials` is a single pure-Go binary. It speaks to clinicaltrials.gov
over plain HTTPS via the public REST API v2, shapes the responses into clean
records, and pipes into the rest of your tools. No API key required.

## Install

```bash
go install github.com/tamnd/clinicaltrials-cli/cmd/clinicaltrials@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/clinicaltrials-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/clinicaltrials:latest --help
```

## Usage

```bash
# Search by condition
clinicaltrials search --condition cancer --status RECRUITING

# Search by intervention
clinicaltrials search --intervention insulin --limit 5

# Search by free-text term
clinicaltrials search --term "phase 3 diabetes"

# Get full study details by NCT ID
clinicaltrials study NCT05608876

# Output as JSON
clinicaltrials search --condition cancer -o json

# Output as CSV
clinicaltrials search --condition cancer -o csv
```

## Output formats

`-o table` (default on TTY), `-o jsonl` (default piped), `-o json`, `-o csv`, `-o tsv`, `-o url`

## Development

```
cmd/clinicaltrials/   thin main, wires cli.NewApp into kit.Main
cli/                  kit app wiring: identity, defaults, domain registration
clinicaltrials/       the library: HTTP client, data models, kit Domain
docs/                 tago documentation site
```

```bash
make build      # ./bin/clinicaltrials
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

## License

Apache-2.0. See [LICENSE](LICENSE).

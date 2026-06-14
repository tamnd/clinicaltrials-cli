// Package cli builds the clinicaltrials command tree on top of the
// clinicaltrials library and the any-cli/kit framework. Every command is a kit
// operation: declared once and exposed as a CLI subcommand, an HTTP route, and
// an MCP tool, with --limit, the --db store tee, and the output formats handled
// by the framework.
package cli

import (
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/clinicaltrials-cli/clinicaltrials"
)

// Build metadata, injected via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp assembles the kit application: identity, defaults, client factory,
// and the clinicaltrials operations.
func NewApp() *kit.App {
	app := kit.New(kit.Identity{
		Binary:  "clinicaltrials",
		Version: Version,
		Short:   "Search ClinicalTrials.gov clinical studies",
		Long: `clinicaltrials turns clinicaltrials.gov into a fast, scriptable command line.

Search clinical studies by condition, intervention, or free-text term. Fetch
full study details by NCT ID. No API key required — the ClinicalTrials.gov
REST API v2 is fully public.

Quick start:
  clinicaltrials search --condition cancer --status RECRUITING
  clinicaltrials search --intervention insulin --limit 5
  clinicaltrials search --term "phase 3 diabetes"
  clinicaltrials study NCT05608876`,
		Site: "clinicaltrials.gov",
		Repo: "https://github.com/tamnd/clinicaltrials-cli",
	}, kit.WithDefaults(func(c *kit.Config) {
		c.Rate = 500 * time.Millisecond
		c.Retries = 3
		c.Timeout = 30 * time.Second
		c.UserAgent = clinicaltrials.DefaultUserAgent
	}))

	// Register the clinicaltrials domain's operations and client factory onto app.
	clinicaltrials.Domain{}.Register(app)

	return app
}

package clinicaltrials

import (
	"context"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go registers the clinicaltrials kit Domain so a blank import in a
// multi-domain host (ant) enables the driver:
//
//	import _ "github.com/tamnd/clinicaltrials-cli/clinicaltrials"
//
// The Domain also builds the standalone clinicaltrials binary via NewApp.
func init() { kit.Register(Domain{}) }

// Domain is the ClinicalTrials.gov driver. It carries no state; the per-run
// client is built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme and the identity the single-site binary inherits.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "clinicaltrials",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "clinicaltrials",
			Short:  "Search ClinicalTrials.gov clinical studies",
			Site:   Host,
			Repo:   "https://github.com/tamnd/clinicaltrials-cli",
		},
	}
}

// Register installs the client factory and operations onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "studies",
		Summary: "Search clinical trials by condition, intervention, or term",
	}, searchStudies)

	kit.Handle(app, kit.OpMeta{
		Name:     "study",
		Group:    "studies",
		Single:   true,
		Resolver: true,
		URIType:  "nctid",
		Summary:  "Get full study details by NCT ID",
		Args:     []kit.Arg{{Name: "nct_id", Help: "NCT ID (e.g. NCT05608876)"}},
	}, getStudy)
}

// newClient builds a Client from the resolved kit Config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	return NewClient(c), nil
}

// ─── input structs ───────────────────────────────────────────────────────────

type searchInput struct {
	Condition    string  `kit:"flag" help:"condition/disease (e.g. diabetes, cancer)"`
	Intervention string  `kit:"flag" help:"intervention/treatment (e.g. insulin)"`
	Term         string  `kit:"flag" help:"general search term"`
	Status       string  `kit:"flag" help:"filter by status (RECRUITING, COMPLETED, etc.)"`
	Limit        int     `kit:"flag,inherit" help:"max results" default:"10"`
	Client       *Client `kit:"inject"`
}

type studyInput struct {
	NCTID  string  `kit:"arg" help:"NCT ID (e.g. NCT05608876)"`
	Client *Client `kit:"inject"`
}

// ─── handlers ────────────────────────────────────────────────────────────────

func searchStudies(ctx context.Context, in searchInput, emit func(Study) error) error {
	studies, err := in.Client.Search(ctx, in.Condition, in.Intervention, in.Term, in.Status, in.Limit)
	if err != nil {
		return err
	}
	for _, s := range studies {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func getStudy(ctx context.Context, in studyInput, emit func(*Study) error) error {
	s, err := in.Client.GetStudy(ctx, in.NCTID)
	if err != nil {
		return mapErr(err)
	}
	return emit(s)
}

// ─── Resolver ────────────────────────────────────────────────────────────────

// Classify turns any accepted input into the canonical (uriType, id).
// NCT IDs (8+ chars starting with NCT) → ("nctid", id); otherwise → ("query", input).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("clinicaltrials: empty input")
	}
	upper := strings.ToUpper(strings.TrimSpace(input))
	if strings.HasPrefix(upper, "NCT") && len(upper) >= 8 {
		return "nctid", upper, nil
	}
	return "query", input, nil
}

// Locate returns the canonical ClinicalTrials.gov URL for a (uriType, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "nctid":
		return "https://" + Host + "/study/" + id, nil
	case "query":
		return "https://" + Host + "/search?query.term=" + id, nil
	default:
		return "", errs.Usage("clinicaltrials has no resource type %q", uriType)
	}
}

// mapErr translates library errors into kit error kinds with appropriate exit codes.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if err == ErrNotFound {
		return errs.NotFound("%s", err.Error())
	}
	return err
}

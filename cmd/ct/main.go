// Command ct is the short-name binary for ClinicalTrials.gov.
// It is identical to the clinicaltrials binary but installed as ct.
package main

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/clinicaltrials-cli/cli"
)

func main() {
	kit.Main(cli.NewApp())
}

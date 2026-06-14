// Command clinicaltrials is a single-binary CLI for ClinicalTrials.gov.
package main

import (
	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/clinicaltrials-cli/cli"
)

func main() {
	kit.Main(cli.NewApp())
}

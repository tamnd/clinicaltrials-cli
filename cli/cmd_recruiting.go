package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) recruitingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recruiting",
		Short: "List currently recruiting trials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			a.progressf("fetching %d recruiting trials...", n)
			trials, err := a.client.Recruiting(cmd.Context(), n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(trials, len(trials))
		},
	}
}

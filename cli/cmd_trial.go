package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) trialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trial <nct-id>",
		Short: "Show a single trial by NCT ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.progressf("fetching trial %s...", args[0])
			detail, err := a.client.Trial(cmd.Context(), args[0])
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(detail)
		},
	}
}

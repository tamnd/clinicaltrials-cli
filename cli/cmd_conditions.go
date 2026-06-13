package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) conditionsCmd() *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:   "conditions <condition>",
		Short: "Trials for a medical condition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(20)
			a.progressf("searching trials for condition %q...", args[0])
			trials, err := a.client.Conditions(cmd.Context(), args[0], status, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(trials, len(trials))
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter by overall status (RECRUITING, COMPLETED, NOT_YET_RECRUITING, ...)")
	return cmd
}

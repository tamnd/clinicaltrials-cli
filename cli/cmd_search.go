package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	var status string
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search trials by keyword",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(20)
			a.progressf("searching for %q...", args[0])
			trials, err := a.client.Search(cmd.Context(), args[0], status, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(trials, len(trials))
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "filter by overall status (RECRUITING, COMPLETED, NOT_YET_RECRUITING, ...)")
	return cmd
}

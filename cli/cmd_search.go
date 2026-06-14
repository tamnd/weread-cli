package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search books on WeRead",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(10)
			a.progressf("searching for %q...", args[0])
			books, err := a.client.Search(cmd.Context(), args[0], n)
			if err != nil {
				return codeError(exitError, err)
			}
			return a.renderOrEmpty(books, len(books))
		},
	}
}

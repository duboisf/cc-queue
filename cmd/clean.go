package cmd

import (
	"fmt"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newCleanCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Remove stale entries (dead processes)",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			removed, err := queue.CleanStale()
			if err != nil {
				return err
			}
			fmt.Fprintf(opts.Stdout, "Removed %d stale entries\n", removed)
			return nil
		},
	}
}

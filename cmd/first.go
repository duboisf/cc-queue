package cmd

import (
	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newFirstCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "first",
		Short: "Jump to the most recent queue entry",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := queue.List()
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				return nil
			}
			sortByNewest(entries)
			return jumpToEntry(entries[0])
		},
	}
}

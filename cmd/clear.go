package cmd

import (
	"fmt"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newClearCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Remove all entries",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := queue.RemoveAll(); err != nil {
				return err
			}
			fmt.Fprintln(opts.Stdout, "Queue cleared")
			return nil
		},
	}
}

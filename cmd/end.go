package cmd

import (
	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newEndCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:    "end",
		Hidden: true,
		Short:  "Remove session on exit (called by SessionEnd hook, reads stdin)",
		Args:   cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := queue.ParseHookInput(opts.Stdin)
			if err != nil {
				return err
			}

			queue.Debugf("END session=%s", input.SessionID)
			queue.Remove(input.SessionID)
			return nil
		},
	}
}

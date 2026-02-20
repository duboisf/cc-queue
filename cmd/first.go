package cmd

import (
	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newFirstCmd(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "first",
		Short: "Jump to the most recent session needing attention",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if fullTab, _ := cmd.Flags().GetBool("full-tab"); fullTab {
				restore, err := opts.FullTabber.EnterFullTab()
				if err != nil {
					return err
				}
				defer restore()
			}

			entries, err := queue.List()
			if err != nil {
				return err
			}

			// Filter to entries needing attention (PERM, ASK, IDLE).
			var pending []*queue.Entry
			for _, e := range entries {
				if queue.NeedsAttention(e.Event) {
					pending = append(pending, e)
				}
			}
			if len(pending) == 0 {
				return nil
			}
			sortForPicker(pending)
			return jumpToEntry(pending[0])
		},
	}
	cmd.Flags().Bool("full-tab", false, "Use stack layout to cover the entire tab, restore on exit")
	_ = cmd.RegisterFlagCompletionFunc("full-tab", cobra.NoFileCompletions)
	return cmd
}

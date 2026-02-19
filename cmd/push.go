package cmd

import (
	"os"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newPushCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Add entry (called by CC hooks, reads stdin)",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			kittyWinID := os.Getenv("KITTY_WINDOW_ID")
			if kittyWinID == "" {
				return nil
			}

			input, err := queue.ParseHookInput(opts.Stdin)
			if err != nil {
				return err
			}

			entry := &queue.Entry{
				Timestamp:     opts.TimeNow(),
				SessionID:     input.SessionID,
				KittyWindowID: kittyWinID,
				PID:           os.Getppid(),
				CWD:           input.CWD,
				Event:         input.EventType(),
			}

			queue.CleanStale()
			return queue.Write(entry)
		},
	}
}

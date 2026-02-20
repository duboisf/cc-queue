package cmd

import (
	"os"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newPopCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "pop",
		Short: "Mark session as working (called by CC hooks, reads stdin)",
		Args:  cobra.NoArgs,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := queue.ParseHookInput(opts.Stdin)
			if err != nil {
				return err
			}

			kittyWinID := os.Getenv("KITTY_WINDOW_ID")
			if kittyWinID == "" {
				// Outside kitty — fall back to removing the entry.
				queue.Debugf("POP session=%s (no kitty, removing)", input.SessionID)
				queue.Remove(input.SessionID)
				return nil
			}

			// In kitty — mark session as working.
			entry := &queue.Entry{
				Timestamp:     opts.TimeNow(),
				SessionID:     input.SessionID,
				KittyWindowID: kittyWinID,
				KittyListenOn: os.Getenv("KITTY_LISTEN_ON"),
				PID:           queue.AncestorPID(),
				CWD:           input.CWD,
				Event:         "working",
			}

			queue.Debugf("POP session=%s -> working", input.SessionID)
			return queue.Write(entry)
		},
	}
}

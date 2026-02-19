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
				queue.Debugf("PUSH skip: KITTY_WINDOW_ID not set")
				return nil
			}

			input, err := queue.ParseHookInput(opts.Stdin)
			if err != nil {
				return err
			}

			message, _ := input.Raw["message"].(string)

			entry := &queue.Entry{
				Timestamp:     opts.TimeNow(),
				SessionID:     input.SessionID,
				KittyWindowID: kittyWinID,
				KittyListenOn: os.Getenv("KITTY_LISTEN_ON"),
				PID:           os.Getppid(),
				CWD:           input.CWD,
				Event:         input.EventType(),
				Message:       message,
			}

			queue.Debugf("PUSH session=%s event=%s ppid=%d", input.SessionID, input.EventType(), os.Getppid())
			return queue.Write(entry)
		},
	}
}

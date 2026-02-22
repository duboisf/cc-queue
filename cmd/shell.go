package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

// KittyLaunchArgs builds the kitty CLI arguments to launch a new window
// in the same tab as the given entry, using its CWD.
func KittyLaunchArgs(entry *queue.Entry) []string {
	args := []string{"@"}
	if entry.KittyListenOn != "" {
		args = append(args, "--to", entry.KittyListenOn)
	}
	args = append(args, "launch", "--type=window",
		"--cwd="+entry.CWD,
		"--match", "id:"+entry.KittyWindowID)
	return args
}

func newShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_shell [session_id]",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return sessionIDCompletions(toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]
			entries, err := queue.List()
			if err != nil {
				return err
			}
			var target *queue.Entry
			for _, e := range entries {
				if e.SessionID == sessionID {
					target = e
					break
				}
			}
			if target == nil || target.KittyWindowID == "" {
				return nil
			}

			launchArgs := KittyLaunchArgs(target)
			launchCmd := exec.Command("kitty", launchArgs...)
			out, err := launchCmd.Output()
			if err != nil {
				// Do NOT remove the entry — the session is still valid.
				return fmt.Errorf("kitty launch failed: %w", err)
			}

			// kitty @ launch prints the new window ID — focus it.
			newWID := strings.TrimSpace(string(out))
			if newWID != "" {
				focusArgs := []string{"@"}
				if target.KittyListenOn != "" {
					focusArgs = append(focusArgs, "--to", target.KittyListenOn)
				}
				focusArgs = append(focusArgs, "focus-window", "--match", "id:"+newWID)
				focusCmd := exec.Command("kitty", focusArgs...)
				_ = focusCmd.Run()
			}
			return nil
		},
	}
}

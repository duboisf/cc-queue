package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/duboisf/cc-queue/internal/queue"
	"github.com/spf13/cobra"
)

func newListCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all pending items (plain text)",
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
				fmt.Fprintln(opts.Stdout, "No pending items")
				return nil
			}

			sortByNewest(entries)
			for _, e := range entries {
				fmt.Fprintf(opts.Stdout, "%-5s %-4s  %s\n",
					queue.FormatAge(e.Timestamp),
					queue.EventLabel(e.Event),
					queue.ShortenPath(e.CWD),
				)
			}
			return nil
		},
	}
}

// fzfLines outputs fzf-formatted lines for all queue entries.
// Format: session_id\tage label  path
func fzfLines() string {
	entries, err := queue.List()
	if err != nil || len(entries) == 0 {
		return ""
	}
	sortByNewest(entries)
	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%s\t%-5s %-4s  %s\n",
			e.SessionID,
			queue.FormatAge(e.Timestamp),
			queue.EventLabel(e.Event),
			queue.ShortenPath(e.CWD),
		)
	}
	return b.String()
}

func newListFzfCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_list-fzf",
		Hidden: true,
		Args:   cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStdout(), fzfLines())
		},
	}
}

// jumpRunE returns the RunE function for the root command (live fzf picker).
func jumpRunE(opts Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		self, err := os.Executable()
		if err != nil {
			self = "cc-queue"
		}
		reloadCmd := self + " _list-fzf"

		fzf := exec.Command("fzf",
			"--height=~50%",
			"--with-nth=2..",
			"--delimiter=\t",
			"--no-multi",
			"--header=cc-queue: select to jump (live)",
			"--prompt=cc-queue> ",
			"--bind=start:reload:"+reloadCmd,
			"--bind=load:reload(sleep 1; "+reloadCmd+")",
		)
		fzf.Stdin = nil
		fzf.Stderr = opts.Stderr

		// fzf writes the selected line to stdout, but since we need to capture
		// it separately from opts.Stdout, use a pipe.
		out, err := fzf.Output()
		if err != nil {
			return nil // user cancelled or fzf error
		}

		selected := strings.TrimSpace(string(out))
		if selected == "" {
			return nil
		}

		sessionID, _, _ := strings.Cut(selected, "\t")

		// Re-read entries since the list may have changed during fzf.
		entries, _ := queue.List()
		var target *queue.Entry
		for _, e := range entries {
			if e.SessionID == sessionID {
				target = e
				break
			}
		}
		if target == nil {
			return nil // entry was removed while fzf was open
		}

		if target.KittyWindowID != "" {
			kittyCmd := exec.Command("kitty", "@", "focus-window",
				"--match", "id:"+target.KittyWindowID)
			if err := kittyCmd.Run(); err != nil {
				return fmt.Errorf("kitty focus-window failed: %w", err)
			}
		}

		queue.Remove(sessionID)
		return nil
	}
}

func sortByNewest(entries []*queue.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
}

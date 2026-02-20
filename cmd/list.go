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

const defaultHeader = "cc-queue — Claude Code session picker (auto-refreshes)"

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

func newListFzfCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:    "_list-fzf",
		Hidden: true,
		Args:   cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if opts.CleanStaleWindowsFn != nil {
				opts.CleanStaleWindowsFn()
			}
			fmt.Fprint(cmd.OutOrStdout(), fzfLines())
		},
	}
}

func newPreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_preview [session_id]",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sessionID := args[0]
			entries, err := queue.List()
			if err != nil {
				return
			}
			for _, e := range entries {
				if e.SessionID == sessionID {
					fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n\n",
						queue.EventLabel(e.Event),
						queue.FormatAge(e.Timestamp),
						queue.ShortenPath(e.CWD))
					if e.Message != "" {
						fmt.Fprintln(cmd.OutOrStdout(), e.Message)
					}
					return
				}
			}
		},
	}
}

// jumpToEntry focuses the kitty window for the given entry and removes it from the queue.
func jumpToEntry(entry *queue.Entry) error {
	if entry.KittyWindowID != "" {
		kittyArgs := []string{"@"}
		if entry.KittyListenOn != "" {
			kittyArgs = append(kittyArgs, "--to", entry.KittyListenOn)
		}
		kittyArgs = append(kittyArgs, "focus-window", "--match", "id:"+entry.KittyWindowID)
		cmd := exec.Command("kitty", kittyArgs...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("kitty focus-window failed: %w\n%s", err, strings.TrimSpace(string(out)))
		}
	}
	queue.Remove(entry.SessionID)
	return nil
}

func newJumpInternalCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "_jump [session_id]",
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
			if target == nil {
				return nil
			}

			// Always remove the entry regardless of jump result.
			defer queue.Remove(target.SessionID)

			if target.KittyWindowID == "" {
				return nil
			}

			kittyArgs := []string{"@"}
			if target.KittyListenOn != "" {
				kittyArgs = append(kittyArgs, "--to", target.KittyListenOn)
			}
			kittyArgs = append(kittyArgs, "focus-window", "--match", "id:"+target.KittyWindowID)
			kittyCmd := exec.Command("kitty", kittyArgs...)
			if out, err := kittyCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("window no longer exists: %s", strings.TrimSpace(string(out)))
			}
			return nil
		},
	}
}

func sessionIDCompletions(toComplete string) []string {
	entries, err := queue.List()
	if err != nil {
		return nil
	}
	var ids []string
	for _, e := range entries {
		if strings.HasPrefix(e.SessionID, toComplete) {
			ids = append(ids, e.SessionID)
		}
	}
	return ids
}

// jumpRunE returns the RunE function for the root command (live fzf picker).
func jumpRunE(opts Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		queue.CleanStale()
		if opts.CleanStaleWindowsFn != nil {
			opts.CleanStaleWindowsFn()
		}

		if fullTab, _ := cmd.Flags().GetBool("full-tab"); fullTab {
			restore, err := opts.FullTabber.EnterFullTab()
			if err != nil {
				return err
			}
			defer restore()
		}

		self, err := os.Executable()
		if err != nil {
			self = "cc-queue"
		}
		reloadCmd := self + " _list-fzf"
		previewCmd := self + " _preview {1}"
		jumpCmd := self + " _jump {1}"

		fzf := exec.Command("fzf",
			"--height=50%",
			"--layout=reverse",
			"--with-nth=2..",
			"--delimiter=\t",
			"--no-multi",
			"--header-first",
			"--header="+defaultHeader,
			"--prompt=Pick a session to jump to it> ",
			"--preview="+previewCmd,
			"--preview-window=down,wrap,40%",
			"--bind=load:change-header("+defaultHeader+")+reload(sleep 2; "+reloadCmd+")",
			"--bind=enter:transform("+jumpCmd+" >/dev/null 2>&1 && echo abort || echo \"change-header(⚠ Kitty window closed — entry removed)\")",
		)
		fzf.Stdin = strings.NewReader(fzfLines())
		fzf.Stderr = opts.Stderr

		fzf.Run()
		return nil
	}
}

func sortByNewest(entries []*queue.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
}

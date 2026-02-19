package cmd

import (
	"fmt"
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

// jumpRunE returns the RunE function for the root command (fzf-based jump).
func jumpRunE(opts Options) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		entries, err := queue.List()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Fprintln(opts.Stdout, "No pending items")
			return nil
		}

		sortByNewest(entries)

		lines := make([]string, len(entries))
		for i, e := range entries {
			lines[i] = fmt.Sprintf("%s\t%-5s %-4s  %s",
				e.SessionID,
				queue.FormatAge(e.Timestamp),
				queue.EventLabel(e.Event),
				queue.ShortenPath(e.CWD),
			)
		}

		selected, err := pickWithFzf(lines, opts)
		if err != nil || selected == "" {
			return nil
		}

		sessionID, _, _ := strings.Cut(selected, "\t")
		var target *queue.Entry
		for _, e := range entries {
			if e.SessionID == sessionID {
				target = e
				break
			}
		}
		if target == nil {
			return fmt.Errorf("entry not found: %s", sessionID)
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

func pickWithFzf(lines []string, opts Options) (string, error) {
	if len(lines) == 1 {
		return lines[0], nil
	}

	fzf := exec.Command("fzf",
		"--height=~50%",
		"--with-nth=2..",
		"--delimiter=\t",
		"--no-multi",
		"--header=Select a Claude Code session",
		"--prompt=cc-queue> ",
	)
	fzf.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	fzf.Stderr = opts.Stderr

	out, err := fzf.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func sortByNewest(entries []*queue.Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
}

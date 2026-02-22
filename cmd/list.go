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

const defaultHeader = "cc-queue — Active Claude Code sessions (auto-refreshes)"

// entryRow holds precomputed display values for a queue entry.
type entryRow struct {
	sessionID string
	age       string
	event     string
	path      string
	branch    string
}

// buildRows precomputes display values and the max path width for alignment.
func buildRows(entries []*queue.Entry) ([]entryRow, int) {
	rows := make([]entryRow, len(entries))
	maxPath := len("PATH") // minimum width = header label
	for i, e := range entries {
		p := queue.ShortenPath(e.CWD)
		if len(p) > maxPath {
			maxPath = len(p)
		}
		branch := queue.GitBranch(e.CWD)
		if branch == "" {
			branch = "-"
		}
		rows[i] = entryRow{
			sessionID: e.SessionID,
			age:       queue.FormatAge(e.Timestamp),
			event:     queue.EventLabel(e.Event),
			path:      p,
			branch:    branch,
		}
	}
	return rows, maxPath
}

func newListCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all active sessions (plain text)",
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
				fmt.Fprintln(opts.Stdout, "No active sessions")
				return nil
			}

			sortForPicker(entries)
			rows, maxPath := buildRows(entries)
			fmt.Fprintf(opts.Stdout, "%-5s %-5s  %-*s  %s\n", "AGE", "EVENT", maxPath, "PATH", "BRANCH")
			for _, r := range rows {
				fmt.Fprintf(opts.Stdout, "%-5s %-5s  %-*s  %s\n",
					r.age, r.event, maxPath, r.path, r.branch)
			}
			return nil
		},
	}
}

// fzfLines outputs fzf-formatted lines for all queue entries.
// The first line is a column header (pinned via --header-lines=1).
// Format: _\theader / session_id\tage event  path  branch
func fzfLines() string {
	entries, err := queue.List()
	if err != nil || len(entries) == 0 {
		return ""
	}
	sortForPicker(entries)
	rows, maxPath := buildRows(entries)
	var b strings.Builder
	fmt.Fprintf(&b, "_\t%-5s %-5s  %-*s  %s\n", "AGE", "EVENT", maxPath, "PATH", "BRANCH")
	for _, r := range rows {
		fmt.Fprintf(&b, "%s\t%-5s %-5s  %-*s  %s\n",
			r.sessionID, r.age, r.event, maxPath, r.path, r.branch)
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
			sf, err := queue.ReadSessionByID(args[0])
			if err != nil || sf.Current == nil {
				return
			}

			w := cmd.OutOrStdout()
			e := sf.Current
			branch := queue.GitBranch(e.CWD)
			if branch != "" {
				fmt.Fprintf(w, "%s  %s  %s  (%s)\n\n",
					queue.EventLabel(e.Event),
					queue.FormatAge(e.Timestamp),
					queue.ShortenPath(e.CWD),
					branch)
			} else {
				fmt.Fprintf(w, "%s  %s  %s\n\n",
					queue.EventLabel(e.Event),
					queue.FormatAge(e.Timestamp),
					queue.ShortenPath(e.CWD))
			}
			if e.Message != "" {
				fmt.Fprintln(w, e.Message)
			}

			if len(sf.History) > 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "\u2500\u2500 Recent activity \u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")
				for _, h := range sf.History {
					msg := h.Message
					if len(msg) > 60 {
						msg = msg[:57] + "..."
					}
					fmt.Fprintf(w, "%5s  %-5s  %s\n",
						queue.FormatAge(h.Timestamp),
						queue.EventLabel(h.Event),
						msg)
				}
			}
		},
	}
}

// jumpToEntry focuses the kitty window for the given entry.
// Sessions persist — only stale entries (failed focus) are removed.
func jumpToEntry(entry *queue.Entry) error {
	if entry.KittyWindowID == "" {
		return nil
	}
	kittyArgs := []string{"@"}
	if entry.KittyListenOn != "" {
		kittyArgs = append(kittyArgs, "--to", entry.KittyListenOn)
	}
	kittyArgs = append(kittyArgs, "focus-window", "--match", "id:"+entry.KittyWindowID)
	cmd := exec.Command("kitty", kittyArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		queue.Remove(entry.SessionID)
		return fmt.Errorf("kitty focus-window failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
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
			if target == nil || target.KittyWindowID == "" {
				return nil
			}

			kittyArgs := []string{"@"}
			if target.KittyListenOn != "" {
				kittyArgs = append(kittyArgs, "--to", target.KittyListenOn)
			}
			kittyArgs = append(kittyArgs, "focus-window", "--match", "id:"+target.KittyWindowID)
			kittyCmd := exec.Command("kitty", kittyArgs...)
			if out, err := kittyCmd.CombinedOutput(); err != nil {
				// Window is stale — remove the entry.
				queue.Remove(target.SessionID)
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
			"--header-lines=1",
			"--prompt=Jump to session> ",
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

// sortForPicker sorts entries with attention-needed sessions first, then by newest.
func sortForPicker(entries []*queue.Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		ai := queue.NeedsAttention(entries[i].Event)
		aj := queue.NeedsAttention(entries[j].Event)
		if ai != aj {
			return ai
		}
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
}

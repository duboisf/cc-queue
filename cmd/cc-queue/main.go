package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/duboisf/cc-queue/internal/queue"
)

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	switch cmd {
	case "push":
		err = runPush()
	case "pop":
		err = runPop()
	case "list":
		err = runList()
	case "clear":
		err = runClear()
	case "clean":
		err = runClean()
	case "install":
		err = runInstall()
	case "help", "-h", "--help":
		printUsage()
	case "":
		err = runJump()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`cc-queue - Claude Code input queue for kitty terminal

Usage:
  cc-queue              Select pending item via fzf and jump to its kitty window
  cc-queue list         List all pending items (plain text)
  cc-queue push         Add entry (called by CC hooks, reads stdin)
  cc-queue pop          Remove entry (called by CC hooks, reads stdin)
  cc-queue clear        Remove all entries
  cc-queue clean        Remove stale entries (dead processes)
  cc-queue install      Install CC hooks (--user or --project)
  cc-queue help         Show this help
`)
}

func runPush() error {
	kittyWinID := os.Getenv("KITTY_WINDOW_ID")
	if kittyWinID == "" {
		return nil
	}

	input, err := queue.ParseHookInput(os.Stdin)
	if err != nil {
		return err
	}

	entry := &queue.Entry{
		Timestamp:     timeNow(),
		SessionID:     input.SessionID,
		KittyWindowID: kittyWinID,
		PID:           os.Getppid(),
		CWD:           input.CWD,
		Event:         input.EventType(),
	}

	queue.CleanStale()
	return queue.Write(entry)
}

func runPop() error {
	input, err := queue.ParseHookInput(os.Stdin)
	if err != nil {
		return err
	}
	queue.Remove(input.SessionID)
	return nil
}

func runList() error {
	entries, err := queue.List()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("No pending items")
		return nil
	}

	sortByNewest(entries)
	for _, e := range entries {
		fmt.Printf("%-5s %-4s  %s\n",
			queue.FormatAge(e.Timestamp),
			queue.EventLabel(e.Event),
			queue.ShortenPath(e.CWD),
		)
	}
	return nil
}

func runJump() error {
	entries, err := queue.List()
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("No pending items")
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

	selected, err := pickWithFzf(lines)
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
		cmd := exec.Command("kitty", "@", "focus-window",
			"--match", "id:"+target.KittyWindowID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("kitty focus-window failed: %w", err)
		}
	}

	queue.Remove(sessionID)
	return nil
}

func runClear() error {
	if err := queue.RemoveAll(); err != nil {
		return err
	}
	fmt.Println("Queue cleared")
	return nil
}

func runClean() error {
	removed, err := queue.CleanStale()
	if err != nil {
		return err
	}
	fmt.Printf("Removed %d stale entries\n", removed)
	return nil
}

func runInstall() error {
	target := queue.TargetUser
	for _, arg := range os.Args[2:] {
		switch arg {
		case "--user", "-u":
			target = queue.TargetUser
		case "--project", "-p":
			target = queue.TargetProject
		default:
			return fmt.Errorf("unknown flag: %s (use --user or --project)", arg)
		}
	}

	if err := queue.InstallHooks(target); err != nil {
		return err
	}

	path, _ := queue.SettingsPath(target)
	fmt.Printf("Hooks installed in %s\n", path)
	return nil
}

func pickWithFzf(lines []string) (string, error) {
	if len(lines) == 1 {
		return lines[0], nil
	}

	cmd := exec.Command("fzf",
		"--height=~50%",
		"--with-nth=2..",
		"--delimiter=\t",
		"--no-multi",
		"--header=Select a Claude Code session",
		"--prompt=cc-queue> ",
	)
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\n"))
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
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

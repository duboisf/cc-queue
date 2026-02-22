package queue

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// FormatAge returns a human-readable age string like "3s", "5m", "2h".
func FormatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
}

// EventLabel maps a notification event type to a short display label.
func EventLabel(event string) string {
	switch event {
	case "permission_prompt":
		return "PERM"
	case "elicitation_dialog":
		return "ASK"
	case "idle_prompt":
		return "IDLE"
	case "working":
		return "WORK"
	case "SessionStart":
		return "START"
	case "SessionEnd":
		return "END"
	default:
		return strings.ToUpper(event)
	}
}

// NeedsAttention returns true if the event represents a state that needs user input.
func NeedsAttention(event string) bool {
	switch event {
	case "", "working", "SessionStart", "SessionEnd":
		return false
	default:
		return true
	}
}

// GitBranch returns the current git branch for a directory, or "" if not a git repo.
func GitBranch(cwd string) string {
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ShortenPath replaces $HOME prefix with ~.
func ShortenPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

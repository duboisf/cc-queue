package queue

import (
	"fmt"
	"os"
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
	default:
		return strings.ToUpper(event)
	}
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

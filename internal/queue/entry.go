package queue

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Entry represents a pending Claude Code input request.
type Entry struct {
	Timestamp     time.Time `json:"timestamp"`
	SessionID     string    `json:"session_id"`
	KittyWindowID string    `json:"kitty_window_id"`
	PID           int       `json:"pid"`
	CWD           string    `json:"cwd"`
	Event         string    `json:"event"`
}

// Dir returns the queue storage directory.
func Dir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "cc-queue")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "cc-queue")
}

// EnsureDir creates the queue directory if it doesn't exist.
func EnsureDir() error {
	return os.MkdirAll(Dir(), 0755)
}

// entryPath returns the file path for a given session ID.
func entryPath(sessionID string) string {
	safe := strings.NewReplacer("/", "_", "\\", "_").Replace(sessionID)
	return filepath.Join(Dir(), safe+".json")
}

// Write persists an entry to disk, keyed by session ID.
// Newer events for the same session overwrite the previous entry.
func Write(e *Entry) error {
	if err := EnsureDir(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(entryPath(e.SessionID), data, 0644)
}

// Read loads an entry from a file path.
func Read(fpath string) (*Entry, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// List returns all entries in the queue directory.
func List() ([]*Entry, error) {
	files, err := filepath.Glob(filepath.Join(Dir(), "*.json"))
	if err != nil {
		return nil, err
	}
	var entries []*Entry
	for _, f := range files {
		e, err := Read(f)
		if err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Remove deletes the entry for a given session ID.
func Remove(sessionID string) error {
	return os.Remove(entryPath(sessionID))
}

// RemoveAll deletes every entry in the queue.
func RemoveAll() error {
	files, err := filepath.Glob(filepath.Join(Dir(), "*.json"))
	if err != nil {
		return err
	}
	for _, f := range files {
		os.Remove(f)
	}
	return nil
}

// IsProcessAlive checks whether a PID still exists.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// CleanStale removes entries whose PID is no longer running.
// Returns the number of entries removed.
func CleanStale() (int, error) {
	entries, err := List()
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if !IsProcessAlive(e.PID) {
			if err := Remove(e.SessionID); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

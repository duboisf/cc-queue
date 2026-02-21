package queue

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Entry represents a pending Claude Code input request.
type Entry struct {
	Timestamp     time.Time `json:"timestamp"`
	SessionID     string    `json:"session_id"`
	KittyWindowID string    `json:"kitty_window_id"`
	KittyListenOn string    `json:"kitty_listen_on,omitempty"`
	PID           int       `json:"pid"`
	CWD           string    `json:"cwd"`
	Event         string    `json:"event"`
	Message       string    `json:"message,omitempty"`
}

// SessionFile wraps the current entry with a capped history of previous entries.
type SessionFile struct {
	Current *Entry   `json:"current"`
	History []*Entry `json:"history,omitempty"`
}

// MaxHistory is the maximum number of historical entries to retain.
const MaxHistory = 10

// Locker abstracts file locking for testability.
type Locker interface {
	Lock(fd int) error
	Unlock(fd int) error
}

type flockLocker struct{}

func (flockLocker) Lock(fd int) error   { return syscall.Flock(fd, syscall.LOCK_EX) }
func (flockLocker) Unlock(fd int) error { return syscall.Flock(fd, syscall.LOCK_UN) }

// DefaultLocker is the file locker used by Write. Replace in tests.
var DefaultLocker Locker = flockLocker{}

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
// It locks the file, reads the existing session, pushes the old current entry
// into history (deduplicating consecutive same-event entries), and writes back.
func Write(e *Entry) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	path := entryPath(e.SessionID)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := DefaultLocker.Lock(int(f.Fd())); err != nil {
		return err
	}
	defer DefaultLocker.Unlock(int(f.Fd()))

	var sf SessionFile
	if data, err := io.ReadAll(f); err == nil && len(data) > 0 {
		sf, _ = parseSessionFile(data)
	}

	// Push current to history, skipping if both event and message are identical.
	if sf.Current != nil && (sf.Current.Event != e.Event || sf.Current.Message != e.Message) {
		sf.History = append([]*Entry{sf.Current}, sf.History...)
		if len(sf.History) > MaxHistory {
			sf.History = sf.History[:MaxHistory]
		}
	}
	sf.Current = e

	out, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Write(out); err != nil {
		return err
	}

	Debugf("WRITE session=%s event=%s cwd=%s pid=%d wid=%s", e.SessionID, e.Event, e.CWD, e.PID, e.KittyWindowID)
	return nil
}

// Read loads the current entry from a session file.
// Handles both new SessionFile format and legacy bare-entry format.
func Read(fpath string) (*Entry, error) {
	sf, err := ReadSession(fpath)
	if err != nil {
		return nil, err
	}
	return sf.Current, nil
}

// ReadSession loads a full SessionFile (current + history) from a file path.
// Handles both new SessionFile format and legacy bare-entry format.
func ReadSession(fpath string) (*SessionFile, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	sf, err := parseSessionFile(data)
	if err != nil {
		return nil, err
	}
	return &sf, nil
}

// ReadSessionByID loads a full SessionFile by session ID.
func ReadSessionByID(sessionID string) (*SessionFile, error) {
	return ReadSession(entryPath(sessionID))
}

// parseSessionFile unmarshals data into a SessionFile, handling backward
// compatibility with legacy bare-entry JSON files.
func parseSessionFile(data []byte) (SessionFile, error) {
	var sf SessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return SessionFile{}, err
	}
	if sf.Current != nil {
		return sf, nil
	}
	// Legacy format: bare Entry at top level.
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return SessionFile{}, err
	}
	if e.SessionID != "" {
		return SessionFile{Current: &e}, nil
	}
	return sf, nil
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
	Debugf("REMOVE session=%s", sessionID)
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

// AncestorPID returns the grandparent PID of the current process.
// Hook commands are typically run via a shell (CC → sh → cc-queue),
// so the grandparent is the Claude Code process, which stays alive
// for the duration of the session. Falls back to os.Getppid() if
// the grandparent cannot be determined.
func AncestorPID() int {
	ppid := os.Getppid()
	grandparent, err := readPPID(ppid)
	if err != nil {
		return ppid
	}
	return grandparent
}

// readPPID reads the parent PID of a given PID from /proc/<pid>/stat.
func readPPID(pid int) (int, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}
	// Format: pid (comm) state ppid ...
	// comm may contain spaces/parens, so find the last ")".
	s := string(data)
	idx := strings.LastIndex(s, ")")
	if idx < 0 || idx+2 >= len(s) {
		return 0, fmt.Errorf("malformed /proc/%d/stat", pid)
	}
	fields := strings.Fields(s[idx+2:])
	if len(fields) < 2 {
		return 0, fmt.Errorf("not enough fields in /proc/%d/stat", pid)
	}
	// fields[0] = state, fields[1] = ppid
	return strconv.Atoi(fields[1])
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
	Debugf("CLEAN_STALE found %d entries", len(entries))
	removed := 0
	for _, e := range entries {
		alive := IsProcessAlive(e.PID)
		Debugf("CLEAN_STALE session=%s pid=%d alive=%v", e.SessionID, e.PID, alive)
		if !alive {
			if err := Remove(e.SessionID); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

// CleanStaleWindows removes entries whose KittyWindowID is not in the valid set.
// Entries with an empty KittyWindowID are skipped.
// Returns the number of entries removed.
func CleanStaleWindows(validWindowIDs map[string]bool) (int, error) {
	entries, err := List()
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if e.KittyWindowID == "" {
			continue
		}
		if !validWindowIDs[e.KittyWindowID] {
			Debugf("CLEAN_STALE_WINDOW session=%s wid=%s", e.SessionID, e.KittyWindowID)
			if err := Remove(e.SessionID); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

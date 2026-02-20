package queue

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndRead(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	entry := &Entry{
		Timestamp:     time.Date(2026, 2, 18, 14, 30, 0, 0, time.UTC),
		SessionID:     "test-session-1",
		KittyWindowID: "42",
		PID:           os.Getpid(),
		CWD:           "/home/fred/git/project",
		Event:         "permission_prompt",
	}

	if err := Write(entry); err != nil {
		t.Fatalf("Write: %v", err)
	}

	fpath := filepath.Join(Dir(), "test-session-1.json")
	got, err := Read(fpath)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if got.SessionID != entry.SessionID {
		t.Errorf("SessionID = %q, want %q", got.SessionID, entry.SessionID)
	}
	if got.KittyWindowID != entry.KittyWindowID {
		t.Errorf("KittyWindowID = %q, want %q", got.KittyWindowID, entry.KittyWindowID)
	}
	if got.Event != entry.Event {
		t.Errorf("Event = %q, want %q", got.Event, entry.Event)
	}
}

func TestWriteOverwritesSameSession(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	e1 := &Entry{
		Timestamp: time.Now(),
		SessionID: "sess-1",
		Event:     "permission_prompt",
	}
	e2 := &Entry{
		Timestamp: time.Now(),
		SessionID: "sess-1",
		Event:     "idle_prompt",
	}

	Write(e1)
	Write(e2)

	entries, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Event != "idle_prompt" {
		t.Errorf("Event = %q, want %q", entries[0].Event, "idle_prompt")
	}
}

func TestListEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	entries, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestRemove(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	Write(&Entry{SessionID: "sess-rm", Timestamp: time.Now()})

	if err := Remove("sess-rm"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	entries, _ := List()
	if len(entries) != 0 {
		t.Errorf("got %d entries after remove, want 0", len(entries))
	}
}

func TestRemoveAll(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	Write(&Entry{SessionID: "a", Timestamp: time.Now()})
	Write(&Entry{SessionID: "b", Timestamp: time.Now()})
	Write(&Entry{SessionID: "c", Timestamp: time.Now()})

	RemoveAll()

	entries, _ := List()
	if len(entries) != 0 {
		t.Errorf("got %d entries after RemoveAll, want 0", len(entries))
	}
}

func TestIsProcessAlive(t *testing.T) {
	if !IsProcessAlive(os.Getpid()) {
		t.Error("current process should be alive")
	}
	// PID 0 is never a valid user process.
	if IsProcessAlive(0) {
		t.Error("PID 0 should not be alive")
	}
	if IsProcessAlive(-1) {
		t.Error("PID -1 should not be alive")
	}
}

func TestCleanStale(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	// Entry with our own PID (alive).
	Write(&Entry{SessionID: "alive", PID: os.Getpid(), Timestamp: time.Now()})
	// Entry with a certainly-dead PID.
	Write(&Entry{SessionID: "dead", PID: 999999999, Timestamp: time.Now()})

	removed, err := CleanStale()
	if err != nil {
		t.Fatalf("CleanStale: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	entries, _ := List()
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].SessionID != "alive" {
		t.Errorf("remaining entry = %q, want %q", entries[0].SessionID, "alive")
	}
}

func TestReadPPID_CurrentProcess(t *testing.T) {
	// readPPID of our own PID should return our parent.
	ppid, err := readPPID(os.Getpid())
	if err != nil {
		t.Fatalf("readPPID: %v", err)
	}
	if ppid != os.Getppid() {
		t.Errorf("readPPID(self) = %d, want %d", ppid, os.Getppid())
	}
}

func TestReadPPID_InvalidPID(t *testing.T) {
	_, err := readPPID(999999999)
	if err == nil {
		t.Error("expected error for non-existent PID")
	}
}

func TestAncestorPID_ReturnsGrandparent(t *testing.T) {
	// AncestorPID walks up two levels: self → parent → grandparent.
	// We can't easily test the exact value, but it should be > 0.
	pid := AncestorPID()
	if pid <= 0 {
		t.Errorf("AncestorPID() = %d, want > 0", pid)
	}
}

func TestCleanStaleWindows(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	Write(&Entry{SessionID: "valid", KittyWindowID: "10", Timestamp: time.Now()})
	Write(&Entry{SessionID: "stale", KittyWindowID: "99", Timestamp: time.Now()})
	Write(&Entry{SessionID: "no-wid", KittyWindowID: "", Timestamp: time.Now()})

	validIDs := map[string]bool{"10": true, "20": true}
	removed, err := CleanStaleWindows(validIDs)
	if err != nil {
		t.Fatalf("CleanStaleWindows: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	entries, _ := List()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}

	ids := map[string]bool{}
	for _, e := range entries {
		ids[e.SessionID] = true
	}
	if !ids["valid"] {
		t.Error("expected 'valid' entry to remain")
	}
	if !ids["no-wid"] {
		t.Error("expected 'no-wid' entry to remain (empty window ID should be skipped)")
	}
	if ids["stale"] {
		t.Error("expected 'stale' entry to be removed")
	}
}

func TestCleanStaleWindows_EmptyValidSet(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	Write(&Entry{SessionID: "s1", KittyWindowID: "10", Timestamp: time.Now()})
	Write(&Entry{SessionID: "s2", KittyWindowID: "20", Timestamp: time.Now()})

	removed, err := CleanStaleWindows(map[string]bool{})
	if err != nil {
		t.Fatalf("CleanStaleWindows: %v", err)
	}
	if removed != 2 {
		t.Errorf("removed = %d, want 2", removed)
	}
}

func TestSessionIDSanitization(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	Write(&Entry{SessionID: "a/b\\c", Timestamp: time.Now()})

	// The file should exist with sanitized name.
	fpath := filepath.Join(Dir(), "a_b_c.json")
	if _, err := os.Stat(fpath); err != nil {
		t.Errorf("sanitized file not found: %v", err)
	}
}

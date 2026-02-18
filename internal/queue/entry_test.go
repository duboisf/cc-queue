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
